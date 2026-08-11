package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateDefaultsToPipeV1(t *testing.T) {
	input := writeCLIInput(t, "R1|CHECK:unit|OV|unit tests pass|test:unit#42|fixture environment\n")
	var stdout bytes.Buffer

	if err := validateTo([]string{"-input", input}, &stdout); err != nil {
		t.Fatalf("validateTo() error = %v", err)
	}
	if got, want := stdout.String(), "valid records=1\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestValidateJSONV1(t *testing.T) {
	input := writeCLIInput(t, `{"records":[{"kind":"R1","target":"CHECK:unit","epistemic":"OV","content":"unit tests pass","evidence":"test:unit#42","bounds":"fixture environment"},{"kind":"Q1","target":"CONTRACT:legacy-api","content":"choose whether to retain the legacy public API","recommendation":"retain the legacy public API","evidence":"analysis:compat#9","bounds":"human-only irreversible public compatibility decision"}]}`)
	var stdout bytes.Buffer

	if err := validateTo([]string{"-format", "json-v1", "-input", input}, &stdout); err != nil {
		t.Fatalf("validateTo() error = %v", err)
	}
	if got, want := stdout.String(), "valid records=2\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestCanonicalizeDefaultsToPipeV1(t *testing.T) {
	input := writeCLIInput(t, "R1|CHECK:unit|OV|unit tests pass|test:unit#42|fixture environment\r\n")
	var stdout bytes.Buffer

	if err := canonicalizeTo([]string{"-input", input}, &stdout); err != nil {
		t.Fatalf("canonicalizeTo() error = %v", err)
	}
	if got, want := stdout.String(), "R1|CHECK:unit|OV|unit tests pass|test:unit#42|fixture environment\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestCanonicalizeJSONV1(t *testing.T) {
	raw := "{\n  \"records\": [\n    {\"kind\":\"D1\",\"target\":\"DECISION:rollout\",\"content\":\"use a staged rollout\",\"authority_reference\":\"HUMAN-DECISION:release#2026-08-10\",\"bounds\":\"fixture release only\"}\n  ]\n}\n"
	input := writeCLIInput(t, raw)
	want := `{"records":[{"authority_reference":"HUMAN-DECISION:release#2026-08-10","bounds":"fixture release only","content":"use a staged rollout","kind":"D1","target":"DECISION:rollout"}]}`
	var stdout bytes.Buffer

	if err := canonicalizeTo([]string{"-format", "json-v1", "-input", input}, &stdout); err != nil {
		t.Fatalf("canonicalizeTo() error = %v", err)
	}
	if got := stdout.String(); got != want+"\n" {
		t.Fatalf("stdout = %q, want %q", got, want+"\n")
	}
}

func TestCommandsRejectUnsupportedFormat(t *testing.T) {
	input := writeCLIInput(t, "R1|CHECK:unit|OV|unit tests pass|test:unit#42|fixture environment\n")
	tests := []struct {
		name string
		run  func([]string, io.Writer) error
	}{
		{name: "validate", run: validateTo},
		{name: "canonicalize", run: canonicalizeTo},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			err := test.run([]string{"-format", "yaml-v1", "-input", input}, &stdout)
			if err == nil || !strings.Contains(err.Error(), "yaml-v1") {
				t.Fatalf("error = %v, want unsupported format naming yaml-v1", err)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
		})
	}
}

func TestCommandsReturnOutputErrors(t *testing.T) {
	input := writeCLIInput(t, "R1|CHECK:unit|OV|unit tests pass|test:unit#42|fixture environment\n")
	tests := []struct {
		name string
		run  func([]string, io.Writer) error
	}{
		{name: "validate", run: validateTo},
		{name: "canonicalize", run: canonicalizeTo},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.run([]string{"-input", input}, failingWriter{})
			if !errors.Is(err, errOutput) {
				t.Fatalf("error = %v, want %v", err, errOutput)
			}
		})
	}
}

var errOutput = errors.New("output failed")

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errOutput
}

func writeCLIInput(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
