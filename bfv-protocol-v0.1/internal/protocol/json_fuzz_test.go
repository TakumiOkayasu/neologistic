package protocol

import (
	"reflect"
	"testing"
)

func FuzzParseJSONMessageStableRoundTrip(f *testing.F) {
	seeds := []string{
		`{"records":[{"kind":"R1","target":"CHECK:unit","epistemic":"OV","content":"unit tests pass","evidence":"test:unit#42","bounds":"fixture environment"}]}`,
		`{"records":[{"kind":"Q1","target":"CONTRACT:api","content":"choose compatibility","recommendation":"retain compatibility","evidence":null,"bounds":"human-only decision"}]}`,
		`{"records":[{"kind":"D1","target":"DECISION:rollout","content":"use staged rollout","authority_reference":"approval:1","bounds":null}]}`,
		`{"records":[{"kind":"R1","target":"ASSERT:unverified","epistemic":"IU","content":"runtime state is unknown","evidence":null,"bounds":null}]}`,
		`{"records":[],"records":[{"kind":"R1","target":"T","epistemic":"OU","content":"c","evidence":null,"bounds":null}]}`,
		`{"records":[{"kind":"R1","kind":"R1","target":"T","epistemic":"OU","content":"c","evidence":null,"bounds":null}]}`,
		`{"records":[{"kind":"R1","target":"T","epistemic":"OU","content":"c","evidence":null,"bounds":null}]} {}`,
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		records, err := ParseJSONMessage(input)
		if err != nil {
			return
		}

		canonical, err := EncodeJSONMessage(records)
		if err != nil {
			t.Fatalf("successfully parsed records failed to encode: %v", err)
		}
		reparsed, err := ParseJSONMessage(canonical)
		if err != nil {
			t.Fatalf("canonical JSON failed to parse: %v", err)
		}
		if !reflect.DeepEqual(reparsed, records) {
			t.Fatalf("round trip mismatch: got %#v, want %#v", reparsed, records)
		}
		reencoded, err := EncodeJSONMessage(reparsed)
		if err != nil {
			t.Fatalf("reparsed records failed to encode: %v", err)
		}
		if reencoded != canonical {
			t.Fatalf("canonical encoding changed: got %q, want %q", reencoded, canonical)
		}
	})
}
