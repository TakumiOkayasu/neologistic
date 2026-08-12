package codec

import "testing"

type sample struct {
	A int `json:"a"`
}

func TestDecodeStrictRejectsDuplicateKeys(t *testing.T) {
	if _, err := DecodeStrict[sample]([]byte(`{"a":1,"a":2}`)); err == nil {
		t.Fatal("expected duplicate key rejection")
	}
}

func TestDecodeStrictRejectsNestedDuplicateKeys(t *testing.T) {
	var value map[string]any
	if _, err := DecodeStrict[map[string]any]([]byte(`{"outer":{"x":1,"x":2}}`)); err == nil {
		t.Fatal("expected nested duplicate key rejection")
	}
	_ = value
}

func TestDecodeStrictRejectsUnknownAndMultipleValues(t *testing.T) {
	if _, err := DecodeStrict[sample]([]byte(`{"a":1,"b":2}`)); err == nil {
		t.Fatal("expected unknown field rejection")
	}
	if _, err := DecodeStrict[sample]([]byte(`{"a":1} {"a":2}`)); err == nil {
		t.Fatal("expected multiple values rejection")
	}
}

func TestDecodeStrictAcceptsCanonicalValue(t *testing.T) {
	value, err := DecodeStrict[sample]([]byte(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if value.A != 1 {
		t.Fatalf("unexpected value: %+v", value)
	}
}
