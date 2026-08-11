package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"bfvprotocol/internal/phase2"
)

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	var err error
	var code int
	switch args[0] {
	case "validate":
		code, err = validateCommand(args[1:])
	case "score":
		code, err = scoreCommand(args[1:])
	case "run":
		code, err = runCommand(args[1:])
	case "checksum":
		code, err = checksumCommand(args[1:])
	default:
		usage()
		return 2
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if code == 0 {
			return 1
		}
	}
	return code
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: phase2ctl validate|score|run|checksum [options]")
}

func checksumCommand(args []string) (int, error) {
	flags := flag.NewFlagSet("checksum", flag.ContinueOnError)
	root := flags.String("root", "", "artifact directory")
	if err := flags.Parse(args); err != nil {
		return 2, err
	}
	if *root == "" {
		return 2, fmt.Errorf("checksum requires -root")
	}
	if err := phase2.WriteSHA256SUMS(*root); err != nil {
		return 1, err
	}
	return 0, writeStdout(map[string]any{"root": *root, "checksum": filepath.Join(*root, "SHA256SUMS.txt")})
}

func validateCommand(args []string) (int, error) {
	flags := flag.NewFlagSet("validate", flag.ContinueOnError)
	fixturesPath := flags.String("fixtures", "phase2/fixtures/cases.jsonl", "fixture JSONL")
	configPath := flags.String("config", "phase2/config.json", "fixed run config")
	if err := flags.Parse(args); err != nil {
		return 2, err
	}
	fixtures, err := readFixtures(*fixturesPath)
	if err != nil {
		return 1, err
	}
	config, err := phase2.LoadPilotConfig(*configPath)
	if err != nil {
		return 1, err
	}
	return 0, writeStdout(map[string]any{
		"fixtures": len(fixtures), "copy": countCohort(fixtures, "copy"), "source_derived": countCohort(fixtures, "source-derived"),
		"candidates": config.Candidates, "planned_cells": len(fixtures) * len(config.Candidates),
		"planned_initial_cli_invocations": len(fixtures) * 7,
		"automatic_runtime_retries":       config.AutomaticRuntimeRetries,
		"claim_boundary":                  "semantic-class coverage, not a population reliability estimate",
	})
}

func scoreCommand(args []string) (int, error) {
	flags := flag.NewFlagSet("score", flag.ContinueOnError)
	fixturesPath := flags.String("fixtures", "phase2/fixtures/cases.jsonl", "fixture JSONL")
	observationsPath := flags.String("observations", "", "observation JSONL")
	outputPath := flags.String("out", "", "output directory")
	defectsPath := flags.String("defects", "", "optional evidence-backed defect JSONL")
	if err := flags.Parse(args); err != nil {
		return 2, err
	}
	if *observationsPath == "" || *outputPath == "" {
		return 2, fmt.Errorf("score requires -observations and -out")
	}
	fixtures, err := readFixtures(*fixturesPath)
	if err != nil {
		return 1, err
	}
	file, err := os.Open(*observationsPath)
	if err != nil {
		return 1, err
	}
	observations, err := phase2.LoadObservations(file)
	closeErr := file.Close()
	if err != nil {
		return 1, err
	}
	if closeErr != nil {
		return 1, closeErr
	}
	report := phase2.Score(fixtures, observations)
	defects, err := readDefects(*defectsPath)
	if err != nil {
		return 1, err
	}
	if err := phase2.WriteReportArtifacts(*outputPath, report, defects); err != nil {
		return 1, err
	}
	if err := phase2.WriteSHA256SUMS(*outputPath); err != nil {
		return 1, err
	}
	if err := writeStdout(map[string]any{"output": *outputPath, "coverage_complete": report.Coverage.Complete, "winner": report.Selection.Winner}); err != nil {
		return 1, err
	}
	return report.ExitCode(), nil
}

func runCommand(args []string) (int, error) {
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	root := flags.String("root", ".", "package root")
	configPath := flags.String("config", "phase2/config.json", "fixed run config")
	fixturesPath := flags.String("fixtures", "phase2/fixtures/cases.jsonl", "fixture JSONL")
	artifactRoot := flags.String("artifacts", "artifacts/phase2", "artifact root")
	if err := flags.Parse(args); err != nil {
		return 2, err
	}
	absoluteRoot, err := filepath.Abs(*root)
	if err != nil {
		return 1, err
	}
	resolve := func(path string) string {
		if filepath.IsAbs(path) {
			return path
		}
		return filepath.Join(absoluteRoot, path)
	}
	runDir, report, err := phase2.RunPilot(context.Background(), phase2.RunOptions{
		PackageRoot: absoluteRoot, ConfigPath: resolve(*configPath), FixturesPath: resolve(*fixturesPath), ArtifactRoot: resolve(*artifactRoot),
	})
	if err != nil {
		return 1, fmt.Errorf("run artifacts retained at %s: %w", runDir, err)
	}
	if err := writeStdout(map[string]any{"run_dir": runDir, "coverage_complete": report.Coverage.Complete, "winner": report.Selection.Winner}); err != nil {
		return 1, err
	}
	return report.ExitCode(), nil
}

func readFixtures(path string) ([]phase2.Fixture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fixtures, err := phase2.LoadFixtures(file)
	closeErr := file.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return fixtures, nil
}

func readDefects(path string) ([]phase2.DefectRecord, error) {
	if path == "" {
		return nil, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var records []phase2.DefectRecord
	scanner := bufio.NewScanner(file)
	for line := 1; scanner.Scan(); line++ {
		var record phase2.DefectRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, fmt.Errorf("defect line %d: %w", line, err)
		}
		records = append(records, record)
	}
	return records, scanner.Err()
}

func countCohort(fixtures []phase2.Fixture, cohort string) int {
	count := 0
	for _, fixture := range fixtures {
		if fixture.Cohort == cohort {
			count++
		}
	}
	return count
}
func writeStdout(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
