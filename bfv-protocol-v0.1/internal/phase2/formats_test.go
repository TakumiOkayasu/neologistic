package phase2

import (
	"strings"
	"testing"
)

func TestPipeLiteralHyphenIsNotNull(t *testing.T) {
	records, _, err := ParseCandidate(CandidatePipe, `R1|T|OU|\-|-|-`)
	if err != nil {
		t.Fatal(err)
	}
	if records[0].Content != "-" {
		t.Fatalf("literal hyphen decoded as %q", records[0].Content)
	}
}

func TestJSONCandidateRejectsUnknownKindKeyAndTrailingValue(t *testing.T) {
	tests := []string{
		`{"records":[{"kind":"D1","target":"T","content":"d","authority_reference":"approval:1","bounds":null,"evidence":null}]}`,
		`{"records":[{"kind":"R1","target":"T","epistemic":"OU","content":"c","evidence":null,"bounds":null,"extra":1}]}`,
		`{"records":[{"kind":"R1","target":"T","epistemic":"OU","content":"c","evidence":null,"bounds":null}]} {}`,
	}
	for _, raw := range tests {
		if _, _, err := ParseCandidate(CandidateJSON, raw); err == nil {
			t.Fatalf("accepted non-strict JSON: %s", raw)
		}
	}
}

func TestJSONAndNLFormatsRoundTripExactKindKeys(t *testing.T) {
	records := []Record{
		{Kind: "R1", Target: "T|1", Epistemic: "OU", Content: "line\nbreak", Evidence: nil, Bounds: stringPointer("scope")},
		{Kind: "D1", Target: "D", Content: "keep", AuthorityReference: "approval:1", Bounds: nil},
	}
	for _, candidate := range []string{CandidateNL, CandidatePipe, CandidateJSON} {
		raw, err := EncodeCandidate(candidate, records)
		if err != nil {
			t.Fatalf("%s encode: %v", candidate, err)
		}
		got, canonical, err := ParseCandidate(candidate, raw)
		if err != nil {
			t.Fatalf("%s parse: %v\n%s", candidate, err, raw)
		}
		if canonical != raw || len(got) != len(records) {
			t.Fatalf("%s did not round trip canonically", candidate)
		}
	}
}

func TestJSONCandidateKeepsCanonicalWireShape(t *testing.T) {
	records := []Record{
		{Kind: "R1", Target: "R", Epistemic: "OU", Content: "result", Evidence: nil, Bounds: stringPointer("scope")},
		{Kind: "Q1", Target: "Q", Content: "choose", Recommendation: "retain", Evidence: stringPointer("analysis:1"), Bounds: stringPointer("human-only")},
		{Kind: "D1", Target: "D", Content: "staged", AuthorityReference: "approval:1", Bounds: nil},
	}
	raw, err := EncodeCandidate(CandidateJSON, records)
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"records":[{"bounds":"scope","content":"result","epistemic":"OU","evidence":null,"kind":"R1","target":"R"},{"bounds":"human-only","content":"choose","evidence":"analysis:1","kind":"Q1","recommendation":"retain","target":"Q"},{"authority_reference":"approval:1","bounds":null,"content":"staged","kind":"D1","target":"D"}]}`
	if raw != want {
		t.Fatalf("EncodeCandidate(json-v1) = %s\nwant = %s", raw, want)
	}
	parsed, canonical, err := ParseCandidate(CandidateJSON, raw)
	if err != nil {
		t.Fatal(err)
	}
	if canonical != want || len(parsed) != len(records) {
		t.Fatalf("JSON candidate delegation changed the canonical form")
	}
}

func TestConsumerRejectsExtraKeysAndCodeFence(t *testing.T) {
	for _, raw := range []string{
		`{"records":[],"task":{"kind":"record_review"},"extra":true}`,
		"```json\n{\"records\":[],\"task\":{\"kind\":\"record_review\"}}\n```",
	} {
		if _, err := ParseConsumer(raw); err == nil {
			t.Fatalf("accepted invalid consumer output: %s", strings.Split(raw, "\n")[0])
		}
	}
}

func TestConsumerDecodesCompleteNaturalLanguageEpistemicPairOnly(t *testing.T) {
	raw := `{"records":[{"kind":"R1","target":"I","epistemic":"inferred, unverified","content":"c","evidence":"e","bounds":null},{"kind":"R1","target":"O","epistemic":"verified","content":"c","evidence":"e","bounds":null}],"task":{"choice":"continue","route":"continue","support_target":"I","support_epistemic":"inferred and unverified","support_evidence":"e","authority_reference":null}}`
	output, err := ParseConsumer(raw)
	if err != nil {
		t.Fatal(err)
	}
	if output.Records[0].Epistemic != "IU" || output.Task["support_epistemic"] != "IU" {
		t.Fatalf("complete epistemic pair was not decoded: %+v", output)
	}
	if output.Records[1].Epistemic != "verified" {
		t.Fatalf("origin loss was repaired instead of preserved: %+v", output.Records[1])
	}
}

func TestConsumerAllowsD1SupportWithoutEvidence(t *testing.T) {
	raw := `{"records":[{"kind":"D1","target":"D","content":"use staged rollout","authority_reference":"approval:1","bounds":null}],"task":{"choice":"staged","route":"continue","support_target":"D","support_epistemic":null,"support_evidence":null,"authority_reference":"approval:1"}}`
	output, err := ParseConsumer(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !taskSupportMatches(output.Task, output.Records) {
		t.Fatalf("D1 support was rejected: %+v", output)
	}
}
