package protocol

import "testing"

func FuzzEncodeParseRoundTrip(f *testing.F) {
	seeds := []Record{
		{
			Kind:      KindResult,
			Target:    "AC:1",
			Epistemic: Epistemic{Origin: OriginObserved, State: StateUnverified},
			Content:   "value|with\\escapes\nand unicode 日本語",
		},
		{
			Kind:           KindQuestion,
			Target:         "CONTRACT",
			Content:        "human ruling required",
			Recommendation: "retain current scope",
			Bounds:         "two irreversible outcomes remain",
		},
	}
	for _, seed := range seeds {
		line, err := EncodeLine(seed)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(line)
	}

	f.Fuzz(func(t *testing.T, input string) {
		record, err := ParseLine(input)
		if err != nil {
			return
		}
		canonical, err := EncodeLine(record)
		if err != nil {
			t.Fatalf("valid parsed record failed to encode: %v", err)
		}
		reparsed, err := ParseLine(canonical)
		if err != nil {
			t.Fatalf("canonical record failed to parse: %v", err)
		}
		if reparsed != record {
			t.Fatalf("round trip mismatch: got %#v, want %#v", reparsed, record)
		}
	})
}
