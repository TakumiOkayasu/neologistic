package phase2

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type PilotConfig struct {
	SchemaVersion           int                 `json:"schema_version"`
	Model                   string              `json:"model"`
	ReasoningEffort         string              `json:"reasoning_effort"`
	Sandbox                 string              `json:"sandbox"`
	ApprovalPolicy          string              `json:"approval_policy"`
	AutomaticRuntimeRetries int                 `json:"automatic_runtime_retries"`
	Candidates              []string            `json:"candidates"`
	OperationalPricing      *OperationalPricing `json:"operational_pricing,omitempty"`
}

type RunOptions struct {
	PackageRoot  string
	ConfigPath   string
	FixturesPath string
	ArtifactRoot string
}

type commandCapture struct {
	Stdout     []byte
	Stderr     []byte
	ExitStatus int
	StartError string
}

type attemptMetadata struct {
	Attempt      int    `json:"attempt"`
	CallExecuted bool   `json:"call_executed"`
	RuntimeOK    bool   `json:"runtime_ok"`
	RetryReason  string `json:"retry_reason,omitempty"`
	AdapterError string `json:"adapter_error,omitempty"`
	ExitStatus   *int   `json:"exit_status"`
}

type DefectRecord struct {
	CaseID    string `json:"case_id,omitempty"`
	Candidate string `json:"candidate,omitempty"`
	Stage     string `json:"stage"`
	Category  string `json:"category"`
	Evidence  string `json:"evidence"`
}

func LoadPilotConfig(path string) (PilotConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PilotConfig{}, err
	}
	var config PilotConfig
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return PilotConfig{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return PilotConfig{}, fmt.Errorf("config must contain exactly one JSON object")
	}
	if config.SchemaVersion != 1 || config.Model == "" || config.ReasoningEffort == "" || config.Sandbox != "read-only" || config.ApprovalPolicy != "never" {
		return PilotConfig{}, fmt.Errorf("config must fix schema_version=1, model, reasoning_effort, read-only sandbox, and never approval")
	}
	if config.AutomaticRuntimeRetries != 0 {
		return PilotConfig{}, fmt.Errorf("automatic_runtime_retries must be 0; later trials are chosen only from observed failure, variance, or decision uncertainty")
	}
	if len(config.Candidates) != len(Candidates) {
		return PilotConfig{}, fmt.Errorf("config candidates must be exactly %v", Candidates)
	}
	for i := range Candidates {
		if config.Candidates[i] != Candidates[i] {
			return PilotConfig{}, fmt.Errorf("config candidates must be exactly %v", Candidates)
		}
	}
	if config.OperationalPricing != nil {
		if err := validateOperationalPricing(*config.OperationalPricing, config.Model); err != nil {
			return PilotConfig{}, fmt.Errorf("operational_pricing: %w", err)
		}
	}
	return config, nil
}

// RunPilot performs preflight and one coverage-driven pass. It never retries a
// parse, semantic, routing, or task failure; only an unsuccessful Codex runtime
// attempt may consume the configured retry slot.
func RunPilot(ctx context.Context, options RunOptions) (string, Report, error) {
	root, err := filepath.Abs(options.PackageRoot)
	if err != nil {
		return "", Report{}, err
	}
	config, err := LoadPilotConfig(options.ConfigPath)
	if err != nil {
		return "", Report{}, fmt.Errorf("config: %w", err)
	}
	fixtureFile, err := os.Open(options.FixturesPath)
	if err != nil {
		return "", Report{}, err
	}
	fixtures, err := LoadFixtures(fixtureFile)
	closeErr := fixtureFile.Close()
	if err != nil {
		return "", Report{}, fmt.Errorf("fixture defect: %w", err)
	}
	if closeErr != nil {
		return "", Report{}, closeErr
	}

	runID := time.Now().UTC().Format("20060102T150405.000000000Z")
	runDir := filepath.Join(options.ArtifactRoot, runID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return "", Report{}, err
	}
	codexPath, err := exec.LookPath("codex")
	if err != nil {
		return runDir, Report{}, fmt.Errorf("preflight: codex executable: %w", err)
	}

	if err := runPreflight(ctx, root, runDir, codexPath, config); err != nil {
		return runDir, Report{}, err
	}

	observationPath := filepath.Join(runDir, "observations.jsonl")
	observationFile, err := os.OpenFile(observationPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return runDir, Report{}, err
	}
	encoder := json.NewEncoder(observationFile)
	var observations []Observation
	var executionOrder []map[string]any
	for fixtureIndex, fixture := range fixtures {
		for orderIndex, candidate := range rotatedCandidates(fixtureIndex) {
			executionOrder = append(executionOrder, map[string]any{"sequence": len(executionOrder) + 1, "fixture_index": fixtureIndex, "candidate_order_index": orderIndex, "case_id": fixture.ID, "cohort": fixture.Cohort, "candidate": candidate})
			cellRoot := filepath.Join(runDir, "cells", fixture.Cohort, fixture.ID, candidate)
			observation := Observation{CaseID: fixture.ID, Candidate: candidate, CoveragePass: 1}
			var handoff string
			var producerOK bool
			if candidate == CandidateDirect {
				observation.Producer, err = writeIdentityStage(filepath.Join(cellRoot, "producer"), fixture.Source)
				handoff, producerOK = fixture.Source, err == nil
			} else {
				var prompt string
				prompt, err = buildProducerPrompt(root, candidate, fixture.Source)
				if err == nil {
					observation.Producer, handoff, producerOK, err = runModelStage(ctx, root, filepath.Join(cellRoot, "producer"), prompt, codexPath, config)
				}
			}
			if err != nil {
				_ = observationFile.Close()
				return runDir, Report{}, err
			}
			observation.Handoff = handoff
			if err := os.MkdirAll(cellRoot, 0o755); err != nil {
				_ = observationFile.Close()
				return runDir, Report{}, err
			}
			if err := os.WriteFile(filepath.Join(cellRoot, "handoff.txt"), []byte(handoff), 0o644); err != nil {
				_ = observationFile.Close()
				return runDir, Report{}, err
			}
			if producerOK {
				consumerPrompt, promptErr := buildConsumerPrompt(root, handoff, fixture.DownstreamTask)
				if promptErr != nil {
					_ = observationFile.Close()
					return runDir, Report{}, promptErr
				}
				observation.Consumer, _, _, err = runModelStage(ctx, root, filepath.Join(cellRoot, "consumer"), consumerPrompt, codexPath, config)
			} else {
				observation.Consumer = StageObservation{Mode: "model", SkippedReason: "producer runtime exhausted its runtime-only retry budget"}
				err = writeJSON(filepath.Join(cellRoot, "consumer", "stage.json"), observation.Consumer)
			}
			if err != nil {
				_ = observationFile.Close()
				return runDir, Report{}, err
			}
			if err := encoder.Encode(observation); err != nil {
				_ = observationFile.Close()
				return runDir, Report{}, err
			}
			observations = append(observations, observation)
		}
	}
	if err := observationFile.Close(); err != nil {
		return runDir, Report{}, err
	}
	if err := writeJSON(filepath.Join(runDir, "execution-order.json"), executionOrder); err != nil {
		return runDir, Report{}, err
	}
	report := ScoreWithOperationalPolicy(fixtures, observations, config.OperationalPricing)
	if err := WriteReportArtifacts(runDir, report, nil); err != nil {
		return runDir, report, err
	}
	if err := WriteSHA256SUMS(runDir); err != nil {
		return runDir, report, err
	}
	return runDir, report, nil
}

func runPreflight(ctx context.Context, root, runDir, codexPath string, config PilotConfig) error {
	preflightRoot := filepath.Join(runDir, "preflight")
	if err := snapshotFrozenInputs(root, runDir); err != nil {
		return fmt.Errorf("snapshot frozen inputs: %w", err)
	}
	checks := []struct {
		name    string
		command string
		args    []string
	}{
		{"codex-version", codexPath, []string{"--version"}},
		{"codex-exec-help", codexPath, []string{"exec", "--help"}},
		{"codex-login-status", codexPath, []string{"login", "status"}},
		{"jq-version", "jq", []string{"--version"}},
	}
	results := map[string]any{}
	for _, check := range checks {
		capture := captureCommand(ctx, root, check.command, check.args, nil)
		checkDir := filepath.Join(preflightRoot, check.name)
		if err := writeCommandCapture(checkDir, capture); err != nil {
			return err
		}
		results[check.name] = map[string]any{"exit_status": capture.ExitStatus, "start_error": capture.StartError}
		if capture.ExitStatus != 0 {
			_ = writeJSON(filepath.Join(preflightRoot, "summary.json"), results)
			return fmt.Errorf("preflight %s failed with exit %d", check.name, capture.ExitStatus)
		}
	}
	smokePrompt, err := os.ReadFile(filepath.Join(root, "phase2", "prompts", "smoke.txt"))
	if err != nil {
		return err
	}
	smoke, response, ok, err := runModelStage(ctx, root, filepath.Join(preflightRoot, "smoke"), string(smokePrompt), codexPath, config)
	if err != nil {
		return err
	}
	results["smoke"] = map[string]any{"runtime_ok": ok, "stage": smoke, "raw_response": response, "cost_scope": "preflight only; excluded from every candidate cost vector"}
	if !ok {
		_ = writeJSON(filepath.Join(preflightRoot, "summary.json"), results)
		return fmt.Errorf("preflight smoke failed after runtime-only retries")
	}
	var smokeObject map[string]any
	if err := json.Unmarshal([]byte(response), &smokeObject); err != nil || smokeObject["phase2_smoke"] != true || len(smokeObject) != 1 {
		results["smoke_semantic_error"] = "response was not the requested exact smoke object"
		_ = writeJSON(filepath.Join(preflightRoot, "summary.json"), results)
		return fmt.Errorf("preflight smoke response mismatch")
	}
	selectedAttempt := 0
	for _, attempt := range smoke.Attempts {
		if attempt.RuntimeOK {
			selectedAttempt = attempt.Attempt
		}
	}
	eventsPath := filepath.Join(preflightRoot, "smoke", fmt.Sprintf("attempt-%03d", selectedAttempt), "stdout.jsonl")
	events, err := os.ReadFile(eventsPath)
	if err != nil {
		return err
	}
	adapterPath := filepath.Join(root, "phase2", "codex-events-to-response.jq")
	adapterCapture := captureCommand(ctx, root, "jq", []string{"-s", "-f", adapterPath}, events)
	if err := writeCommandCapture(filepath.Join(preflightRoot, "jq-event-adapter"), adapterCapture); err != nil {
		return err
	}
	var adapted struct {
		Raw   string     `json:"raw"`
		Usage CodexUsage `json:"usage"`
	}
	if adapterCapture.ExitStatus != 0 || json.Unmarshal(adapterCapture.Stdout, &adapted) != nil || adapted.Raw != response {
		return fmt.Errorf("preflight jq event adapter did not preserve the smoke response exactly")
	}

	metadata := map[string]any{
		"model": config.Model, "reasoning_effort": config.ReasoningEffort, "sandbox": config.Sandbox,
		"approval_policy": config.ApprovalPolicy, "automatic_runtime_retries": config.AutomaticRuntimeRetries,
		"codex_executable": codexPath, "codex_executable_sha256": fileSHA256(codexPath),
		"source_revision":  commandText(root, "git", "rev-parse", "HEAD"),
		"source_status":    commandText(root, "git", "status", "--short"),
		"observed_adapter": "last item.completed agent_message text plus last turn.completed usage; raw text is not trimmed, normalized, or repaired",
		"usage_fields":     []string{"input_tokens", "cached_input_tokens", "cache_write_input_tokens", "output_tokens", "reasoning_output_tokens"},
	}
	metadata["frozen_input_sha256"] = frozenInputHashes(root)
	results["environment"] = metadata
	return writeJSON(filepath.Join(preflightRoot, "summary.json"), results)
}

func runModelStage(ctx context.Context, root, stageRoot, prompt, codexPath string, config PilotConfig) (StageObservation, string, bool, error) {
	stage := StageObservation{Mode: "model"}
	var final string
	for index := 0; index <= config.AutomaticRuntimeRetries; index++ {
		attemptNumber := index + 1
		retryReason := ""
		if index > 0 {
			retryReason = "runtime_hard_failure"
		}
		args := codexExecArgs(root, config)
		capture := captureCommand(ctx, root, codexPath, args, []byte(prompt))
		result, adapterErr := ExtractCodexEvents(bytes.NewReader(capture.Stdout))
		runtimeOK := capture.ExitStatus == 0 && adapterErr == nil
		response := ""
		usage := CodexUsage{}
		if adapterErr == nil {
			response, usage = result.FinalText, result.Usage
		}
		callExecuted := capture.StartError == ""
		attempt := Attempt{Attempt: attemptNumber, CallExecuted: callExecuted, RuntimeOK: runtimeOK, RetryReason: retryReason, Prompt: prompt, RawEvents: string(capture.Stdout), Response: response, Usage: usage, ExitStatus: capture.ExitStatus, Stderr: string(capture.Stderr)}
		stage.Attempts = append(stage.Attempts, attempt)
		attemptRoot := filepath.Join(stageRoot, fmt.Sprintf("attempt-%03d", attemptNumber))
		if err := writeAttemptArtifacts(attemptRoot, prompt, capture, response, usage, attemptMetadata{Attempt: attemptNumber, CallExecuted: callExecuted, RuntimeOK: runtimeOK, RetryReason: retryReason, AdapterError: errorText(adapterErr), ExitStatus: &capture.ExitStatus}); err != nil {
			return stage, "", false, err
		}
		if runtimeOK {
			final = response
			break
		}
	}
	if err := writeJSON(filepath.Join(stageRoot, "stage.json"), stage); err != nil {
		return stage, "", false, err
	}
	return stage, final, finalStageSucceeded(stage), nil
}

func writeIdentityStage(stageRoot, source string) (StageObservation, error) {
	stage := StageObservation{
		Mode: "not_applicable", Attempts: []Attempt{},
		SkippedReason: "direct-v1 is the raw-source identity baseline and has no producer serialization call",
	}
	if err := os.MkdirAll(stageRoot, 0o755); err != nil {
		return stage, err
	}
	if err := os.WriteFile(filepath.Join(stageRoot, "raw-source.txt"), []byte(source), 0o644); err != nil {
		return stage, err
	}
	return stage, writeJSON(filepath.Join(stageRoot, "stage.json"), stage)
}

func buildProducerPrompt(root, candidate, source string) (string, error) {
	common, err := os.ReadFile(filepath.Join(root, "phase2", "prompts", "producer-common.txt"))
	if err != nil {
		return "", err
	}
	format, err := os.ReadFile(filepath.Join(root, "phase2", "prompts", "producer-"+candidate+".txt"))
	if err != nil {
		return "", err
	}
	parts := []string{string(common)}
	if candidate == CandidatePipe {
		protocolInstruction, readErr := os.ReadFile(filepath.Join(root, "prompts", "protocol-instruction.md"))
		if readErr != nil {
			return "", readErr
		}
		parts = append(parts, "Existing pipe-v1 protocol authority:\n"+string(protocolInstruction))
	}
	parts = append(parts, string(format), "SOURCE BEGIN\n"+source+"\nSOURCE END")
	return strings.Join(parts, "\n\n"), nil
}

func buildConsumerPrompt(root, handoff, downstreamTask string) (string, error) {
	instruction, err := os.ReadFile(filepath.Join(root, "phase2", "prompts", "consumer.txt"))
	if err != nil {
		return "", err
	}
	return string(instruction) + "\n\nPAYLOAD BEGIN\n" + handoff + "\nPAYLOAD END\n\nTASK BEGIN\n" + downstreamTask + "\nTASK END\n", nil
}

func codexExecArgs(root string, config PilotConfig) []string {
	return []string{"exec", "--ephemeral", "--ignore-user-config", "--ignore-rules", "--model", config.Model,
		"-c", `model_reasoning_effort="` + config.ReasoningEffort + `"`, "-c", `approval_policy="` + config.ApprovalPolicy + `"`,
		"--sandbox", config.Sandbox, "--json", "--cd", root, "-"}
}

func captureCommand(ctx context.Context, directory, command string, args []string, input []byte) commandCapture {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = directory
	if input != nil {
		cmd.Stdin = bytes.NewReader(input)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	capture := commandCapture{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitStatus: 0}
	if err == nil {
		return capture
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		capture.ExitStatus = exitError.ExitCode()
	} else {
		capture.ExitStatus = -1
		capture.StartError = err.Error()
	}
	return capture
}

func writeAttemptArtifacts(root, prompt string, capture commandCapture, response string, usage CodexUsage, metadata attemptMetadata) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	files := map[string][]byte{
		"prompt.txt": []byte(prompt), "response.txt": []byte(response), "stdout.jsonl": capture.Stdout,
		"stderr.txt": capture.Stderr, "exit-status.txt": []byte(fmt.Sprintf("%d\n", capture.ExitStatus)),
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(root, name), data, 0o644); err != nil {
			return err
		}
	}
	if err := writeJSON(filepath.Join(root, "usage.json"), usage); err != nil {
		return err
	}
	return writeJSON(filepath.Join(root, "attempt.json"), metadata)
}

func writeCommandCapture(root string, capture commandCapture) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(root, "stdout.txt"), capture.Stdout, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(root, "stderr.txt"), capture.Stderr, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, "exit-status.txt"), []byte(fmt.Sprintf("%d\n", capture.ExitStatus)), 0o644)
}

func finalStageSucceeded(stage StageObservation) bool {
	for _, attempt := range stage.Attempts {
		if attempt.RuntimeOK {
			return true
		}
	}
	return false
}

func WriteReportArtifacts(root string, report Report, explicit []DefectRecord) error {
	if report.ParetoSelection.Status == "" && report.ParetoSelection.Rule == "" {
		report.ParetoSelection = report.Selection
	}
	// selection.json and score.json.selection are the legacy Pareto aliases.
	// Normalize them here so callers cannot accidentally publish divergent values.
	report.Selection = report.ParetoSelection
	for _, defect := range explicit {
		if defect.Candidate == "" {
			continue
		}
		for index := range report.Candidates {
			if report.Candidates[index].Candidate == defect.Candidate {
				report.Candidates[index].Defects[defect.Category]++
			}
		}
	}
	if err := writeJSON(filepath.Join(root, "score.json"), report); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(root, "coverage.json"), report.Coverage); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(root, "candidate-summaries.json"), report.Candidates); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(root, "selection.json"), report.ParetoSelection); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(root, "operational-selection.json"), report.OperationalSelection); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(root, "pareto-selection.json"), report.ParetoSelection); err != nil {
		return err
	}
	defects := append([]DefectRecord(nil), explicit...)
	for _, trial := range report.Trials {
		if trial.RuntimeHardFailureAttempts > 0 {
			defects = append(defects, DefectRecord{CaseID: trial.CaseID, Candidate: trial.Candidate, Stage: "runtime", Category: "runtime", Evidence: fmt.Sprintf("%d Codex CLI attempts were runtime hard failures", trial.RuntimeHardFailureAttempts)})
		}
		if trial.HardFailure && trial.RuntimeHardFailureAttempts == 0 && !hasExplicitAttribution(explicit, trial.CaseID, trial.Candidate) {
			defects = append(defects, DefectRecord{CaseID: trial.CaseID, Candidate: trial.Candidate, Stage: "scoring", Category: "unclassified", Evidence: "observed hard failure; causal attribution requires separate evidence"})
		}
	}
	return writeDefects(root, defects)
}

func hasExplicitAttribution(defects []DefectRecord, caseID, candidate string) bool {
	for _, defect := range defects {
		if defect.CaseID == caseID && defect.Candidate == candidate && defect.Category != "unclassified" {
			return true
		}
	}
	return false
}

func writeDefects(root string, defects []DefectRecord) error {
	categories := []string{"fixture", "protocol", "model", "scorer", "runtime", "unclassified"}
	buckets := map[string][]DefectRecord{}
	for _, category := range categories {
		buckets[category] = nil
	}
	for _, defect := range defects {
		if _, ok := buckets[defect.Category]; !ok || defect.Stage == "" || defect.Evidence == "" {
			return fmt.Errorf("invalid defect attribution %+v", defect)
		}
		buckets[defect.Category] = append(buckets[defect.Category], defect)
	}
	defectRoot := filepath.Join(root, "defects")
	if err := os.MkdirAll(defectRoot, 0o755); err != nil {
		return err
	}
	counts := map[string]int{}
	for _, category := range categories {
		path := filepath.Join(defectRoot, category+".jsonl")
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		encoder := json.NewEncoder(file)
		for _, defect := range buckets[category] {
			if err := encoder.Encode(defect); err != nil {
				_ = file.Close()
				return err
			}
		}
		if err := file.Close(); err != nil {
			return err
		}
		counts[category] = len(buckets[category])
	}
	return writeJSON(filepath.Join(defectRoot, "summary.json"), counts)
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func fileSHA256(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return "unavailable: " + err.Error()
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "unavailable: " + err.Error()
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func commandText(directory, command string, args ...string) string {
	capture := captureCommand(context.Background(), directory, command, args, nil)
	if capture.ExitStatus != 0 {
		return fmt.Sprintf("unavailable (exit %d): %s", capture.ExitStatus, strings.TrimSpace(string(capture.Stderr)))
	}
	return strings.TrimSpace(string(capture.Stdout))
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func rotatedCandidates(fixtureIndex int) []string {
	rotated := make([]string, len(Candidates))
	for i := range Candidates {
		rotated[i] = Candidates[(fixtureIndex+i)%len(Candidates)]
	}
	return rotated
}

func frozenInputHashes(root string) map[string]string {
	hashes := map[string]string{}
	for _, path := range frozenInputPaths() {
		hashes[path] = fileSHA256(filepath.Join(root, path))
	}
	return hashes
}

func frozenInputPaths() []string {
	return []string{
		"phase2/config.json", "phase2/fixtures/cases.jsonl", "phase2/prompts/consumer.txt", "phase2/prompts/producer-common.txt",
		"phase2/prompts/producer-nl-v1.txt", "phase2/prompts/producer-pipe-v1.txt", "phase2/prompts/producer-json-v1.txt",
		"phase2/prompts/smoke.txt", "prompts/protocol-instruction.md", "phase2/codex-events-to-response.jq",
	}
}

func snapshotFrozenInputs(root, runDir string) error {
	for _, relative := range frozenInputPaths() {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			return err
		}
		destination := filepath.Join(runDir, "frozen-inputs", filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(destination, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func WriteSHA256SUMS(root string) error {
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() == "SHA256SUMS.txt" {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		return err
	}
	sort.Strings(paths)
	var output strings.Builder
	for _, relative := range paths {
		hash := fileSHA256(filepath.Join(root, filepath.FromSlash(relative)))
		if strings.HasPrefix(hash, "unavailable:") {
			return fmt.Errorf("checksum %s: %s", relative, hash)
		}
		fmt.Fprintf(&output, "%s  %s\n", hash, relative)
	}
	return os.WriteFile(filepath.Join(root, "SHA256SUMS.txt"), []byte(output.String()), 0o644)
}
