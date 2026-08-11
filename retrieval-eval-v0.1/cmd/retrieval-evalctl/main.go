package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"retrievaleval/internal/eval"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 || args[0] != "score" {
		return fmt.Errorf("usage: retrieval-evalctl score -fixtures FILE -observations FILE")
	}
	fs := flag.NewFlagSet("score", flag.ContinueOnError)
	fixturesPath := fs.String("fixtures", "", "fixture JSONL")
	observationsPath := fs.String("observations", "", "observation JSONL")
	systemsRaw := fs.String("systems", "", "comma-separated expected systems")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if *fixturesPath == "" || *observationsPath == "" || strings.TrimSpace(*systemsRaw) == "" {
		return fmt.Errorf("-fixtures, -observations, and -systems are required")
	}
	fixtures, err := eval.ReadJSONL[eval.Fixture](*fixturesPath)
	if err != nil {
		return err
	}
	observations, err := eval.ReadJSONL[eval.Observation](*observationsPath)
	if err != nil {
		return err
	}
	var systems []string
	for _, value := range strings.Split(*systemsRaw, ",") {
		if value = strings.TrimSpace(value); value != "" {
			systems = append(systems, value)
		}
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(eval.ScoreWithSystems(fixtures, observations, systems))
}
