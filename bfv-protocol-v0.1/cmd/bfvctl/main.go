package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"bfvprotocol/internal/eval"
	"bfvprotocol/internal/protocol"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "validate":
		err = validate(os.Args[2:])
	case "canonicalize":
		err = canonicalize(os.Args[2:])
	case "eval":
		err = evaluate(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: bfvctl <validate|canonicalize|eval> [options]")
}

func validate(args []string) error {
	return validateTo(args, os.Stdout)
}

func validateTo(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	input := fs.String("input", "-", "wire message file, or - for stdin")
	format := fs.String("format", string(protocol.FormatPipeV1), "wire format: pipe-v1 or json-v1")
	if err := fs.Parse(args); err != nil {
		return err
	}
	r, closeFn, err := openInput(*input)
	if err != nil {
		return err
	}
	defer closeFn()
	raw, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	records, err := protocol.Parse(protocol.Format(*format), string(raw))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "valid records=%d\n", len(records))
	return err
}

func canonicalize(args []string) error {
	return canonicalizeTo(args, os.Stdout)
}

func canonicalizeTo(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("canonicalize", flag.ContinueOnError)
	input := fs.String("input", "-", "wire message file, or - for stdin")
	format := fs.String("format", string(protocol.FormatPipeV1), "wire format: pipe-v1 or json-v1")
	if err := fs.Parse(args); err != nil {
		return err
	}
	r, closeFn, err := openInput(*input)
	if err != nil {
		return err
	}
	defer closeFn()
	raw, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	selectedFormat := protocol.Format(*format)
	records, err := protocol.Parse(selectedFormat, string(raw))
	if err != nil {
		return err
	}
	canonical, err := protocol.Encode(selectedFormat, records)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, canonical)
	return err
}

func evaluate(args []string) error {
	fs := flag.NewFlagSet("eval", flag.ContinueOnError)
	casesPath := fs.String("cases", "", "JSONL case file")
	responsesPath := fs.String("responses", "", "JSONL response file")
	detailsPath := fs.String("details", "", "optional JSON detail output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *casesPath == "" || *responsesPath == "" {
		return fmt.Errorf("-cases and -responses are required")
	}

	casesFile, err := os.Open(*casesPath)
	if err != nil {
		return err
	}
	defer casesFile.Close()
	cases, err := eval.ReadCases(casesFile)
	if err != nil {
		return err
	}

	responsesFile, err := os.Open(*responsesPath)
	if err != nil {
		return err
	}
	defer responsesFile.Close()
	responses, err := eval.ReadResponses(responsesFile)
	if err != nil {
		return err
	}

	results, summaries, err := eval.Score(cases, responses)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(summaries); err != nil {
		return err
	}

	if *detailsPath != "" {
		f, err := os.Create(*detailsPath)
		if err != nil {
			return err
		}
		defer f.Close()
		detailEncoder := json.NewEncoder(f)
		detailEncoder.SetIndent("", "  ")
		if err := detailEncoder.Encode(results); err != nil {
			return err
		}
	}
	return nil
}

func openInput(path string) (io.Reader, func(), error) {
	if path == "-" {
		return os.Stdin, func() {}, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	return f, func() { _ = f.Close() }, nil
}
