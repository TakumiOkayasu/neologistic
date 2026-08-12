package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/artifact"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/codec"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/eval"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/freeze"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/model"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/pathguard"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/render"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/runplan"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/validate"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var err error
	switch os.Args[1] {
	case "validate":
		err = runValidate(os.Args[2:])
	case "render":
		err = runRender(os.Args[2:])
	case "score":
		err = runScore(os.Args[2:])
	case "score-cell":
		err = runScoreCell(os.Args[2:])
	case "freeze":
		err = runFreeze(os.Args[2:])
	case "verify-freeze":
		err = runVerifyFreeze(os.Args[2:])
	case "manifest":
		err = runManifest(os.Args[2:])
	case "preflight":
		err = runPreflight(os.Args[2:])
	case "artifact-freeze":
		err = runArtifactFreeze(os.Args[2:])
	case "artifact-verify":
		err = runArtifactVerify(os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "phase3ctl:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: phase3ctl <validate|render|score|score-cell|freeze|verify-freeze|manifest|preflight|artifact-freeze|artifact-verify> [flags]")
	os.Exit(2)
}

func defaultRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

func loadInputs(root, armsPath, fixturesPath string) ([]model.Arm, []model.Fixture, error) {
	armsFile, err := pathguard.ResolveRegular(root, armsPath)
	if err != nil {
		return nil, nil, err
	}
	fixturesFile, err := pathguard.ResolveRegular(root, fixturesPath)
	if err != nil {
		return nil, nil, err
	}
	arms, err := codec.LoadJSON[[]model.Arm](armsFile)
	if err != nil {
		return nil, nil, err
	}
	fixtures, err := codec.LoadJSONL[model.Fixture](fixturesFile)
	if err != nil {
		return nil, nil, err
	}
	return arms, fixtures, nil
}

func commonInputFlags(name string, args []string) (*flag.FlagSet, *string, *string, *string, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	root := fs.String("root", defaultRoot(), "Phase 3 root")
	arms := fs.String("arms", "fixtures/arms.json", "arms JSON path relative to root")
	fixtures := fs.String("fixtures", "fixtures/cases.jsonl", "fixtures JSONL path relative to root")
	if err := fs.Parse(args); err != nil {
		return nil, nil, nil, nil, err
	}
	return fs, root, arms, fixtures, nil
}

func validateAll(root, armsPath, fixturesPath string) ([]model.Arm, []model.Fixture, error) {
	arms, fixtures, err := loadInputs(root, armsPath, fixturesPath)
	if err != nil {
		return nil, nil, err
	}
	if err := validate.ValidateArms(root, arms); err != nil {
		return nil, nil, err
	}
	if err := validate.ValidateFixtures(fixtures); err != nil {
		return nil, nil, err
	}
	for _, arm := range arms {
		for _, fixture := range fixtures {
			sender, err := render.Sender(root, arm, fixture)
			if err != nil {
				return nil, nil, fmt.Errorf("render sender %s/%s: %w", fixture.ID, arm.ID, err)
			}
			if _, err := render.Receiver(root, arm, fixture, sender); err != nil {
				return nil, nil, fmt.Errorf("render receiver %s/%s: %w", fixture.ID, arm.ID, err)
			}
		}
	}
	if err := validateSchemaFiles(root); err != nil {
		return nil, nil, err
	}
	return arms, fixtures, nil
}

func runValidate(args []string) error {
	_, root, armsPath, fixturesPath, err := commonInputFlags("validate", args)
	if err != nil {
		return err
	}
	arms, fixtures, err := validateAll(*root, *armsPath, *fixturesPath)
	if err != nil {
		return err
	}
	fmt.Printf("valid arms=%d fixtures=%d cells=%d\n", len(arms), len(fixtures), len(arms)*len(fixtures))
	return nil
}

func validateSchemaFiles(root string) error {
	directory, err := pathguard.ResolveDirectory(root, "schemas")
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path, err := pathguard.ResolveRegular(root, filepath.ToSlash(filepath.Join("schemas", entry.Name())))
		if err != nil {
			return err
		}
		value, err := codec.LoadJSON[map[string]any](path)
		if err != nil {
			return fmt.Errorf("schema %s: %w", entry.Name(), err)
		}
		if value["$schema"] == nil || value["$id"] == nil {
			return fmt.Errorf("schema %s lacks $schema or $id", entry.Name())
		}
	}
	return nil
}

func runRender(args []string) error {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	root := fs.String("root", defaultRoot(), "Phase 3 root")
	armsPath := fs.String("arms", "fixtures/arms.json", "arms JSON path")
	fixturesPath := fs.String("fixtures", "fixtures/cases.jsonl", "fixtures JSONL path")
	caseID := fs.String("case", "", "fixture ID")
	armID := fs.String("arm", "", "arm ID")
	stage := fs.String("stage", "sender", "sender or receiver")
	handoffPath := fs.String("handoff", "", "handoff file for receiver stage")
	if err := fs.Parse(args); err != nil {
		return err
	}
	arms, fixtures, err := validateAll(*root, *armsPath, *fixturesPath)
	if err != nil {
		return err
	}
	var arm *model.Arm
	for i := range arms {
		if arms[i].ID == *armID {
			arm = &arms[i]
			break
		}
	}
	var fixture *model.Fixture
	for i := range fixtures {
		if fixtures[i].ID == *caseID {
			fixture = &fixtures[i]
			break
		}
	}
	if arm == nil || fixture == nil {
		return fmt.Errorf("unknown arm or case")
	}
	var prompt string
	switch *stage {
	case "sender":
		prompt, err = render.Sender(*root, *arm, *fixture)
	case "receiver":
		if *handoffPath == "" {
			return fmt.Errorf("receiver stage requires -handoff")
		}
		handoffFile, resolveErr := pathguard.ResolveRegular(*root, *handoffPath)
		if resolveErr != nil {
			return resolveErr
		}
		data, readErr := os.ReadFile(handoffFile)
		if readErr != nil {
			return readErr
		}
		prompt, err = render.Receiver(*root, *arm, *fixture, string(data))
	default:
		return fmt.Errorf("unsupported stage %q", *stage)
	}
	if err != nil {
		return err
	}
	fmt.Print(prompt)
	return nil
}

func runScore(args []string) error {
	fs := flag.NewFlagSet("score", flag.ContinueOnError)
	root := fs.String("root", defaultRoot(), "Phase 3 root")
	armsPath := fs.String("arms", "fixtures/arms.json", "arms JSON path")
	fixturesPath := fs.String("fixtures", "fixtures/cases.jsonl", "fixtures JSONL path")
	manifestPath := fs.String("manifest", "", "run manifest JSON path")
	observationsPath := fs.String("observations", "", "observations JSONL path")
	pricingPath := fs.String("pricing", "", "optional pricing JSON path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *manifestPath == "" || *observationsPath == "" {
		return fmt.Errorf("-manifest and -observations are required")
	}
	arms, fixtures, err := validateAll(*root, *armsPath, *fixturesPath)
	if err != nil {
		return err
	}
	manifestFile, err := pathguard.ResolveRegular(*root, *manifestPath)
	if err != nil {
		return err
	}
	manifest, err := codec.LoadJSON[model.RunManifest](manifestFile)
	if err != nil {
		return err
	}
	if err := validate.ValidateRunManifest(manifest, fixtures, arms); err != nil {
		return err
	}
	freezeHash, err := sha256File(filepath.Join(*root, "freeze.lock.json"))
	if err != nil {
		return err
	}
	if manifest.FreezeLockSHA256 != freezeHash {
		return fmt.Errorf("run manifest freeze hash does not match current freeze.lock.json")
	}
	observationsFile, err := pathguard.ResolveRegular(*root, *observationsPath)
	if err != nil {
		return err
	}
	observations, err := codec.LoadJSONL[model.Observation](observationsFile)
	if err != nil {
		return err
	}
	var pricing *model.PricingPolicy
	if *pricingPath != "" {
		pricingFile, resolveErr := pathguard.ResolveRegular(*root, *pricingPath)
		if resolveErr != nil {
			return resolveErr
		}
		value, loadErr := codec.LoadJSON[model.PricingPolicy](pricingFile)
		if loadErr != nil {
			return loadErr
		}
		if err := validate.ValidatePricingForManifest(value, manifest); err != nil {
			return err
		}
		pricing = &value
	}
	report := eval.Score(fixtures, arms, observations, pricing, manifest)
	data, err := codec.EncodeCanonical(report)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(data)
	return err
}

func runScoreCell(args []string) error {
	fs := flag.NewFlagSet("score-cell", flag.ContinueOnError)
	root := fs.String("root", defaultRoot(), "Phase 3 root")
	manifestPath := fs.String("manifest", "", "run manifest JSON path")
	observationPath := fs.String("observation", "", "single observation JSON path")
	pricingPath := fs.String("pricing", "", "optional pricing JSON path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *manifestPath == "" || *observationPath == "" {
		return fmt.Errorf("-manifest and -observation are required")
	}
	arms, fixtures, err := validateAll(*root, "fixtures/arms.json", "fixtures/cases.jsonl")
	if err != nil {
		return err
	}
	manifestFile, err := pathguard.ResolveRegular(*root, *manifestPath)
	if err != nil {
		return err
	}
	manifest, err := codec.LoadJSON[model.RunManifest](manifestFile)
	if err != nil {
		return err
	}
	if err := validate.ValidateRunManifest(manifest, fixtures, arms); err != nil {
		return err
	}
	freezeHash, err := sha256File(filepath.Join(*root, "freeze.lock.json"))
	if err != nil {
		return err
	}
	if manifest.FreezeLockSHA256 != freezeHash {
		return fmt.Errorf("run manifest freeze hash does not match current freeze.lock.json")
	}
	observationFile, err := pathguard.ResolveRegular(*root, *observationPath)
	if err != nil {
		return err
	}
	observation, err := codec.LoadJSON[model.Observation](observationFile)
	if err != nil {
		return err
	}
	var fixture *model.Fixture
	for i := range fixtures {
		if fixtures[i].ID == observation.CaseID {
			fixture = &fixtures[i]
			break
		}
	}
	var arm *model.Arm
	for i := range arms {
		if arms[i].ID == observation.ArmID {
			arm = &arms[i]
			break
		}
	}
	if fixture == nil || arm == nil {
		return fmt.Errorf("observation references unknown fixture or arm")
	}
	var pricing *model.PricingPolicy
	if *pricingPath != "" {
		pricingFile, resolveErr := pathguard.ResolveRegular(*root, *pricingPath)
		if resolveErr != nil {
			return resolveErr
		}
		value, loadErr := codec.LoadJSON[model.PricingPolicy](pricingFile)
		if loadErr != nil {
			return loadErr
		}
		if err := validate.ValidatePricingForManifest(value, manifest); err != nil {
			return err
		}
		pricing = &value
	}
	trial, err := eval.ScoreCell(*fixture, *arm, observation, pricing, manifest)
	if err != nil {
		return err
	}
	data, err := codec.EncodeCanonical(trial)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(data)
	return err
}

func runFreeze(args []string) error {
	fs := flag.NewFlagSet("freeze", flag.ContinueOnError)
	root := fs.String("root", defaultRoot(), "Phase 3 root")
	out := fs.String("out", "freeze.lock.json", "output path relative to root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	manifest, err := freeze.Build(*root)
	if err != nil {
		return err
	}
	data, err := freeze.Encode(manifest)
	if err != nil {
		return err
	}
	path, err := pathguard.JoinForCreate(*root, *out)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func verifyFreeze(root, manifestPath string) (freeze.Manifest, error) {
	path, err := pathguard.ResolveRegular(root, manifestPath)
	if err != nil {
		return freeze.Manifest{}, err
	}
	manifest, err := freeze.Load(path)
	if err != nil {
		return freeze.Manifest{}, err
	}
	if err := freeze.Verify(root, manifest); err != nil {
		return freeze.Manifest{}, err
	}
	return manifest, nil
}

func runVerifyFreeze(args []string) error {
	fs := flag.NewFlagSet("verify-freeze", flag.ContinueOnError)
	root := fs.String("root", defaultRoot(), "Phase 3 root")
	manifestPath := fs.String("manifest", "freeze.lock.json", "manifest path relative to root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	manifest, err := verifyFreeze(*root, *manifestPath)
	if err != nil {
		return err
	}
	fmt.Printf("freeze verified files=%d\n", len(manifest.Files))
	return nil
}

func runManifest(args []string) error {
	fs := flag.NewFlagSet("manifest", flag.ContinueOnError)
	root := fs.String("root", defaultRoot(), "Phase 3 root")
	runID := fs.String("run-id", "", "stable run ID")
	senderProvider := fs.String("sender-provider", "", "sender provider")
	senderModel := fs.String("sender-model", "", "sender model")
	senderEffort := fs.String("sender-effort", "", "sender reasoning effort")
	senderTemperature := fs.Float64("sender-temperature", -1, "sender temperature; omit when unavailable")
	senderMaxOutput := fs.Int64("sender-max-output-tokens", 0, "sender max output tokens; omit when unavailable")
	receiverProvider := fs.String("receiver-provider", "", "receiver provider")
	receiverModel := fs.String("receiver-model", "", "receiver model")
	receiverEffort := fs.String("receiver-effort", "", "receiver reasoning effort")
	receiverTemperature := fs.Float64("receiver-temperature", -1, "receiver temperature; omit when unavailable")
	receiverMaxOutput := fs.Int64("receiver-max-output-tokens", 0, "receiver max output tokens; omit when unavailable")
	transport := fs.String("transport", "", "fixed transport/harness identifier")
	retries := fs.Int("runtime-retries", 0, "fixed runtime retry count per stage")
	if err := fs.Parse(args); err != nil {
		return err
	}
	arms, fixtures, err := validateAll(*root, "fixtures/arms.json", "fixtures/cases.jsonl")
	if err != nil {
		return err
	}
	freezeHash, err := sha256File(filepath.Join(*root, "freeze.lock.json"))
	if err != nil {
		return err
	}
	senderTemp, err := optionalTemperature(*senderTemperature)
	if err != nil {
		return fmt.Errorf("sender temperature: %w", err)
	}
	receiverTemp, err := optionalTemperature(*receiverTemperature)
	if err != nil {
		return fmt.Errorf("receiver temperature: %w", err)
	}
	senderMax, err := optionalPositiveInt(*senderMaxOutput)
	if err != nil {
		return fmt.Errorf("sender max output tokens: %w", err)
	}
	receiverMax, err := optionalPositiveInt(*receiverMaxOutput)
	if err != nil {
		return fmt.Errorf("receiver max output tokens: %w", err)
	}
	manifest := model.RunManifest{
		Version: model.RunManifestVersion, Scope: model.ExperimentScope, RunID: *runID, FreezeLockSHA256: freezeHash,
		Sender: model.ModelConfig{Provider: *senderProvider, Model: *senderModel, ReasoningEffort: *senderEffort,
			Temperature: senderTemp, MaxOutputTokens: senderMax},
		Receiver: model.ModelConfig{Provider: *receiverProvider, Model: *receiverModel, ReasoningEffort: *receiverEffort,
			Temperature: receiverTemp, MaxOutputTokens: receiverMax},
		Transport: *transport,
		Isolation: model.IsolationConfig{
			Workspace: "empty-temporary", RepositoryAccess: false, Tools: []string{},
			PluginsEnabled: false, MemoryEnabled: false, Network: "provider-only",
		},
		EliminationPolicy:   model.EliminationPolicy,
		RuntimeRetriesFixed: *retries,
	}
	manifest.ExecutionOrder, err = runplan.ExecutionOrder(fixtures, arms)
	if err != nil {
		return err
	}
	if err := validate.ValidateRunManifest(manifest, fixtures, arms); err != nil {
		return err
	}
	data, err := codec.EncodeCanonical(manifest)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(data)
	return err
}

func runPreflight(args []string) error {
	fs := flag.NewFlagSet("preflight", flag.ContinueOnError)
	root := fs.String("root", defaultRoot(), "Phase 3 root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	arms, fixtures, err := validateAll(*root, "fixtures/arms.json", "fixtures/cases.jsonl")
	if err != nil {
		return err
	}
	manifest, err := verifyFreeze(*root, "freeze.lock.json")
	if err != nil {
		return err
	}
	freezeHash, err := sha256File(filepath.Join(*root, "freeze.lock.json"))
	if err != nil {
		return err
	}
	result := map[string]any{
		"ready": true, "model_calls": 0, "arms": len(arms), "fixtures": len(fixtures),
		"cells": len(arms) * len(fixtures), "frozen_files": len(manifest.Files), "freeze_lock_sha256": freezeHash,
	}
	data, err := codec.EncodeCanonical(result)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(data)
	return err
}

func runArtifactFreeze(args []string) error {
	fs := flag.NewFlagSet("artifact-freeze", flag.ContinueOnError)
	root := fs.String("root", defaultRoot(), "Phase 3 root")
	directoryPath := fs.String("directory", "", "artifact directory relative to root")
	out := fs.String("out", "artifact-manifest.json", "manifest path relative to artifact directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *directoryPath == "" {
		return fmt.Errorf("-directory is required")
	}
	directory, err := pathguard.ResolveDirectory(*root, *directoryPath)
	if err != nil {
		return err
	}
	manifest, err := artifact.Build(directory, *out)
	if err != nil {
		return err
	}
	data, err := artifact.Encode(manifest)
	if err != nil {
		return err
	}
	outPath, err := pathguard.JoinForCreate(directory, *out)
	if err != nil {
		return err
	}
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		return err
	}
	fmt.Printf("artifact manifest written files=%d path=%s\n", len(manifest.Files), filepath.ToSlash(filepath.Join(*directoryPath, *out)))
	return nil
}

func runArtifactVerify(args []string) error {
	fs := flag.NewFlagSet("artifact-verify", flag.ContinueOnError)
	root := fs.String("root", defaultRoot(), "Phase 3 root")
	directoryPath := fs.String("directory", "", "artifact directory relative to root")
	manifestPath := fs.String("manifest", "artifact-manifest.json", "manifest path relative to artifact directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *directoryPath == "" {
		return fmt.Errorf("-directory is required")
	}
	directory, err := pathguard.ResolveDirectory(*root, *directoryPath)
	if err != nil {
		return err
	}
	path, err := pathguard.ResolveRegular(directory, *manifestPath)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	manifest, err := artifact.Decode(data)
	if err != nil {
		return err
	}
	if err := artifact.Verify(directory, *manifestPath, manifest); err != nil {
		return err
	}
	fmt.Printf("artifact manifest verified files=%d\n", len(manifest.Files))
	return nil
}

func sha256File(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func optionalTemperature(value float64) (*float64, error) {
	if value == -1 {
		return nil, nil
	}
	if value < 0 || value > 2 {
		return nil, fmt.Errorf("must be between 0 and 2, or omitted")
	}
	return &value, nil
}

func optionalPositiveInt(value int64) (*int64, error) {
	if value == 0 {
		return nil, nil
	}
	if value < 0 {
		return nil, fmt.Errorf("must be positive, or omitted")
	}
	return &value, nil
}
