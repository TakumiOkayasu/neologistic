package phase2

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPilotConfigRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	raw := `{"schema_version":1,"model":"m","reasoning_effort":"high","sandbox":"read-only","approval_policy":"never","automatic_runtime_retries":0,"candidates":["direct-v1","nl-v1","pipe-v1","json-v1"],"unknown":true}`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPilotConfig(path); err == nil {
		t.Fatal("unknown config field was accepted")
	}
}

func TestProducerPromptUsesExistingPipeAuthorityAndDoesNotLeakOracle(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	prompt, err := buildProducerPrompt(root, CandidatePipe, "SOURCE_MARKER")
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"# BFV wire output v0.1", "Do not inspect files", "SOURCE_MARKER"} {
		if !strings.Contains(prompt, required) {
			t.Fatalf("prompt missing %q", required)
		}
	}
	if strings.Contains(prompt, "task_oracle") {
		t.Fatal("hidden oracle leaked into producer prompt")
	}
}

func TestRuntimeSuccessWithBadSemanticTextIsNotRetried(t *testing.T) {
	temp := t.TempDir()
	executable := filepath.Join(temp, "fake-codex")
	script := "#!/bin/sh\nprintf '%s\\n' '{\"type\":\"item.completed\",\"item\":{\"type\":\"agent_message\",\"text\":\"ordinary-semantic-failure\"}}' '{\"type\":\"turn.completed\",\"usage\":{\"input_tokens\":1,\"cached_input_tokens\":0,\"cache_write_input_tokens\":0,\"output_tokens\":1,\"reasoning_output_tokens\":0}}'\n"
	if err := os.WriteFile(executable, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	config := PilotConfig{Model: "m", ReasoningEffort: "high", Sandbox: "read-only", ApprovalPolicy: "never", AutomaticRuntimeRetries: 3}
	stage, response, ok, err := runModelStage(context.Background(), temp, filepath.Join(temp, "stage"), "prompt", executable, config)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || response != "ordinary-semantic-failure" || len(stage.Attempts) != 1 {
		t.Fatalf("stage=%+v response=%q ok=%v", stage, response, ok)
	}
	stored, err := os.ReadFile(filepath.Join(temp, "stage", "attempt-001", "response.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(stored) != response {
		t.Fatal("raw response artifact changed")
	}
}

func TestCandidateExecutionOrderRotates(t *testing.T) {
	want := [][]string{
		{CandidateDirect, CandidateNL, CandidatePipe, CandidateJSON},
		{CandidateNL, CandidatePipe, CandidateJSON, CandidateDirect},
		{CandidatePipe, CandidateJSON, CandidateDirect, CandidateNL},
	}
	for index := range want {
		got := rotatedCandidates(index)
		for i := range got {
			if got[i] != want[index][i] {
				t.Fatalf("rotation %d = %v", index, got)
			}
		}
	}
}

func TestWriteSHA256SUMSExcludesItself(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteSHA256SUMS(root); err != nil {
		t.Fatal(err)
	}
	manifest, err := os.ReadFile(filepath.Join(root, "SHA256SUMS.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(manifest), "  a.txt\n") || strings.Contains(string(manifest), "SHA256SUMS.txt") {
		t.Fatalf("manifest=%q", manifest)
	}
}

func TestSnapshotFrozenInputsPreservesBytesAndPaths(t *testing.T) {
	root := t.TempDir()
	run := t.TempDir()
	for _, relative := range frozenInputPaths() {
		path := filepath.Join(root, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(relative+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := snapshotFrozenInputs(root, run); err != nil {
		t.Fatal(err)
	}
	for _, relative := range frozenInputPaths() {
		data, err := os.ReadFile(filepath.Join(run, "frozen-inputs", filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != relative+"\n" {
			t.Fatalf("snapshot changed %s: %q", relative, data)
		}
	}
}

func TestExplicitDefectAttributionSuppressesDuplicateUnclassified(t *testing.T) {
	root := t.TempDir()
	report := Report{
		Candidates: []CandidateSummary{{Candidate: CandidateDirect, Defects: map[string]int{"model": 0, "unclassified": 0}}},
		Trials:     []TrialScore{{CaseID: "case", Candidate: CandidateDirect, HardFailure: true}},
	}
	explicit := []DefectRecord{{CaseID: "case", Candidate: CandidateDirect, Stage: "consumer", Category: "model", Evidence: "origin was lost"}}
	if err := WriteReportArtifacts(root, report, explicit); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "defects", "unclassified.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Fatalf("duplicate unclassified attribution: %s", data)
	}
	var summaries []CandidateSummary
	summaryData, err := os.ReadFile(filepath.Join(root, "candidate-summaries.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(summaryData, &summaries); err != nil {
		t.Fatal(err)
	}
	if summaries[0].Defects["model"] != 1 {
		t.Fatalf("explicit defect was not aggregated: %+v", summaries[0].Defects)
	}
}
