package protocol

import (
	"reflect"
	"strings"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	records := []Record{
		{
			Kind:      KindResult,
			Target:    "AC:2",
			Epistemic: Epistemic{Origin: OriginObserved, State: StateVerified},
			Content:   "TTL units differ between config|expiry",
			Evidence:  "file:config.ts:22,file:cache.ts:81",
			Bounds:    "commit:abc123\\worktree",
		},
		{
			Kind:           KindQuestion,
			Target:         "AC:3",
			Content:        "production DB credential is required",
			Recommendation: "grant read-only temporary credential",
			Evidence:       "tool:db-connect#17",
			Bounds:         "integration verification cannot continue",
		},
		{
			Kind:      KindDecision,
			Target:    "AC:3",
			Content:   "waive production integration verification",
			Reference: "human:ruling-42",
			Bounds:    "unit and local integration evidence remain required",
		},
	}

	encoded, err := EncodeMessage(records)
	if err != nil {
		t.Fatalf("EncodeMessage() error = %v", err)
	}
	parsed, err := ParseMessage(encoded + "\n")
	if err != nil {
		t.Fatalf("ParseMessage() error = %v", err)
	}
	if !reflect.DeepEqual(parsed, records) {
		t.Fatalf("round trip mismatch\nencoded: %s\nparsed: %#v\nwant: %#v", encoded, parsed, records)
	}
}

func TestLiteralDash(t *testing.T) {
	t.Parallel()

	record := Record{
		Kind:      KindResult,
		Target:    "AC:1",
		Epistemic: Epistemic{Origin: OriginObserved, State: StateUnverified},
		Content:   "-",
	}
	encoded, err := EncodeLine(record)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(encoded, `\-`) {
		t.Fatalf("literal dash was not escaped: %q", encoded)
	}
	parsed, err := ParseLine(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Content != "-" {
		t.Fatalf("content = %q, want literal dash", parsed.Content)
	}
}

func TestVerifiedRequiresEvidence(t *testing.T) {
	t.Parallel()

	_, err := ParseLine("R1|AC:1|IV|cause inferred|-|runtime not reproduced")
	if err == nil || !strings.Contains(err.Error(), "requires evidence") {
		t.Fatalf("error = %v, want evidence validation failure", err)
	}
}

func TestQuestionRequiresRecommendationAndBounds(t *testing.T) {
	t.Parallel()

	cases := []string{
		"Q1|AC:1|choose behavior|-|artifact:1|irreversible",
		"Q1|AC:1|choose behavior|retain current behavior|artifact:1|-",
	}
	for _, input := range cases {
		if _, err := ParseLine(input); err == nil {
			t.Fatalf("ParseLine(%q) succeeded, want failure", input)
		}
	}
}

func TestRejectsMalformedRecords(t *testing.T) {
	t.Parallel()

	cases := []string{
		"",
		"R1|AC:1|OV|x|evidence",
		"R1|AC:1|ZZ|x|evidence|-",
		"Q1|AC:1|x|recommendation|-|reason|extra",
		"D1|AC:1|accept|-|-",
		"X1|candidate",
		"R1|AC:1|OU|bad\\qescape|-|-",
		"R1|AC:1|OU|x||-",
		"R1|AC:1|OU|trailing\\|-|-",
	}
	for _, input := range cases {
		if _, err := ParseLine(input); err == nil {
			t.Fatalf("ParseLine(%q) succeeded, want failure", input)
		}
	}
}

func TestMessageRejectsBlankRecord(t *testing.T) {
	t.Parallel()

	_, err := ParseMessage("R1|AC:1|OU|x|-|-\n\nR1|AC:2|OU|y|-|-")
	if err == nil || !strings.Contains(err.Error(), "blank record") {
		t.Fatalf("error = %v, want blank record failure", err)
	}
}
