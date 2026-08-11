package protocol

import (
	"reflect"
	"strings"
	"testing"
)

func TestJSONRoundTripAllRecordKinds(t *testing.T) {
	t.Parallel()

	records := []Record{
		{
			Kind:      KindResult,
			Target:    "CHECK:unit",
			Epistemic: Epistemic{Origin: OriginObserved, State: StateVerified},
			Content:   "unit tests pass",
			Evidence:  "test:unit#42",
			Bounds:    "fixture environment",
		},
		{
			Kind:           KindQuestion,
			Target:         "CONTRACT:api",
			Content:        "choose compatibility",
			Recommendation: "retain compatibility",
			Evidence:       "analysis:compat#9",
			Bounds:         "human-only decision",
		},
		{
			Kind:      KindDecision,
			Target:    "DECISION:rollout",
			Content:   "use staged rollout",
			Reference: "approval:1",
		},
	}

	encoded, err := EncodeJSONMessage(records)
	if err != nil {
		t.Fatalf("EncodeJSONMessage() error = %v", err)
	}
	const want = `{"records":[{"bounds":"fixture environment","content":"unit tests pass","epistemic":"OV","evidence":"test:unit#42","kind":"R1","target":"CHECK:unit"},{"bounds":"human-only decision","content":"choose compatibility","evidence":"analysis:compat#9","kind":"Q1","recommendation":"retain compatibility","target":"CONTRACT:api"},{"authority_reference":"approval:1","bounds":null,"content":"use staged rollout","kind":"D1","target":"DECISION:rollout"}]}`
	if encoded != want {
		t.Fatalf("EncodeJSONMessage() = %s\nwant = %s", encoded, want)
	}

	parsed, err := ParseJSONMessage(encoded)
	if err != nil {
		t.Fatalf("ParseJSONMessage() error = %v", err)
	}
	if !reflect.DeepEqual(parsed, records) {
		t.Fatalf("round trip mismatch\nparsed: %#v\nwant: %#v", parsed, records)
	}
}

func TestJSONNullOptionalFieldsRoundTrip(t *testing.T) {
	t.Parallel()

	records := []Record{{
		Kind:      KindResult,
		Target:    "ASSERT:unverified",
		Epistemic: Epistemic{Origin: OriginInferred, State: StateUnverified},
		Content:   "runtime state is unknown",
	}}
	encoded, err := EncodeJSONMessage(records)
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"records":[{"bounds":null,"content":"runtime state is unknown","epistemic":"IU","evidence":null,"kind":"R1","target":"ASSERT:unverified"}]}`
	if encoded != want {
		t.Fatalf("EncodeJSONMessage() = %s\nwant = %s", encoded, want)
	}
	parsed, err := ParseJSONMessage(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(parsed, records) {
		t.Fatalf("null round trip mismatch: %#v", parsed)
	}
}

func TestJSONStrictSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{name: "top-level extra key", raw: `{"records":[{"kind":"R1","target":"T","epistemic":"OU","content":"c","evidence":null,"bounds":null}],"extra":true}`},
		{name: "duplicate top-level key", raw: `{"records":[],"records":[{"kind":"R1","target":"T","epistemic":"OU","content":"c","evidence":null,"bounds":null}]}`},
		{name: "record extra key", raw: `{"records":[{"kind":"R1","target":"T","epistemic":"OU","content":"c","evidence":null,"bounds":null,"extra":true}]}`},
		{name: "duplicate record key", raw: `{"records":[{"kind":"R1","kind":"R1","target":"T","epistemic":"OU","content":"c","evidence":null,"bounds":null}]}`},
		{name: "record missing key", raw: `{"records":[{"kind":"R1","target":"T","epistemic":"OU","content":"c","evidence":null}]}`},
		{name: "wrong nullable type", raw: `{"records":[{"kind":"R1","target":"T","epistemic":"OU","content":"c","evidence":7,"bounds":null}]}`},
		{name: "empty optional evidence", raw: `{"records":[{"kind":"R1","target":"T","epistemic":"OU","content":"c","evidence":"","bounds":null}]}`},
		{name: "empty optional bounds", raw: `{"records":[{"kind":"D1","target":"T","content":"c","authority_reference":"approval:1","bounds":""}]}`},
		{name: "unsupported kind", raw: `{"records":[{"kind":"X1","target":"T","content":"c"}]}`},
		{name: "empty records", raw: `{"records":[]}`},
		{name: "null records", raw: `{"records":null}`},
		{name: "trailing value", raw: `{"records":[{"kind":"R1","target":"T","epistemic":"OU","content":"c","evidence":null,"bounds":null}]} {}`},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := ParseJSONMessage(test.raw); err == nil {
				t.Fatalf("ParseJSONMessage(%s) succeeded, want failure", test.raw)
			}
		})
	}
}

func TestJSONRejectsInvalidUTF8(t *testing.T) {
	t.Parallel()

	raw := string([]byte{'{', '"', 'r', 'e', 'c', 'o', 'r', 'd', 's', '"', ':', '[', 0xff, ']', '}'})
	if _, err := ParseJSONMessage(raw); err == nil || !strings.Contains(err.Error(), "UTF-8") {
		t.Fatalf("error = %v, want invalid UTF-8 failure", err)
	}
}

func TestFormatDispatchAndPipeCompatibility(t *testing.T) {
	t.Parallel()

	records := []Record{{
		Kind:      KindResult,
		Target:    "T",
		Epistemic: Epistemic{Origin: OriginInferred, State: StateUnverified},
		Content:   "candidate",
	}}
	pipe, err := Encode(FormatPipeV1, records)
	if err != nil {
		t.Fatal(err)
	}
	legacyPipe, err := EncodeMessage(records)
	if err != nil {
		t.Fatal(err)
	}
	if pipe != legacyPipe {
		t.Fatalf("pipe dispatch = %q, legacy = %q", pipe, legacyPipe)
	}
	parsed, err := Parse(FormatPipeV1, pipe)
	if err != nil || !reflect.DeepEqual(parsed, records) {
		t.Fatalf("pipe dispatch parse = %#v, %v", parsed, err)
	}
	jsonMessage, err := Encode(FormatJSONV1, records)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err = Parse(FormatJSONV1, jsonMessage)
	if err != nil || !reflect.DeepEqual(parsed, records) {
		t.Fatalf("JSON dispatch parse = %#v, %v", parsed, err)
	}
	if _, err := Parse(Format("unknown"), pipe); err == nil {
		t.Fatal("Parse() accepted unknown format")
	}
	if _, err := Encode(Format("unknown"), records); err == nil {
		t.Fatal("Encode() accepted unknown format")
	}
}
