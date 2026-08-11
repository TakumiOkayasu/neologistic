package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// ParseJSONMessage parses one strict json-v1 BFV transfer object.
func ParseJSONMessage(raw string) ([]Record, error) {
	if !utf8.ValidString(raw) {
		return nil, errors.New("message is not valid UTF-8")
	}

	decoder := json.NewDecoder(strings.NewReader(raw))
	object, err := decodeStrictJSONObject(decoder)
	if err != nil {
		return nil, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("more than one JSON value")
	}
	if len(object) != 1 || object["records"] == nil {
		return nil, fmt.Errorf("json-v1 requires exactly the records key")
	}

	var rawRecords []json.RawMessage
	if err := json.Unmarshal(object["records"], &rawRecords); err != nil {
		return nil, fmt.Errorf("records: %w", err)
	}
	if len(rawRecords) == 0 {
		return nil, errors.New("records must not be empty")
	}

	records := make([]Record, len(rawRecords))
	for i, rawRecord := range rawRecords {
		record, err := parseJSONRecord(rawRecord)
		if err != nil {
			return nil, fmt.Errorf("record %d: %w", i+1, err)
		}
		records[i] = record
	}
	return records, nil
}

// EncodeJSONMessage validates and emits one canonical json-v1 transfer object.
func EncodeJSONMessage(records []Record) (string, error) {
	if len(records) == 0 {
		return "", errors.New("at least one record is required")
	}

	objects := make([]map[string]any, len(records))
	for i, record := range records {
		if err := record.Validate(); err != nil {
			return "", fmt.Errorf("record %d: %w", i+1, err)
		}
		objects[i] = jsonRecordObject(record)
	}

	encoded, err := json.Marshal(map[string]any{"records": objects})
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func parseJSONRecord(raw json.RawMessage) (Record, error) {
	object, err := decodeStrictJSONObject(json.NewDecoder(strings.NewReader(string(raw))))
	if err != nil {
		return Record{}, err
	}

	var kind string
	if err := json.Unmarshal(object["kind"], &kind); err != nil {
		return Record{}, errors.New("kind is required")
	}

	var required []string
	switch Kind(kind) {
	case KindResult:
		required = []string{"kind", "target", "epistemic", "content", "evidence", "bounds"}
	case KindQuestion:
		required = []string{"kind", "target", "content", "recommendation", "evidence", "bounds"}
	case KindDecision:
		required = []string{"kind", "target", "content", "authority_reference", "bounds"}
	default:
		return Record{}, fmt.Errorf("unsupported record kind %q", kind)
	}
	if len(object) != len(required) {
		return Record{}, fmt.Errorf("%s requires exactly keys %v", kind, required)
	}
	for _, key := range required {
		if _, ok := object[key]; !ok {
			return Record{}, fmt.Errorf("%s requires key %q", kind, key)
		}
	}

	target, err := jsonString(object["target"], "target")
	if err != nil {
		return Record{}, err
	}
	content, err := jsonString(object["content"], "content")
	if err != nil {
		return Record{}, err
	}
	record := Record{Kind: Kind(kind), Target: target, Content: content}

	switch record.Kind {
	case KindResult:
		epistemic, err := jsonString(object["epistemic"], "epistemic")
		if err != nil {
			return Record{}, err
		}
		record.Epistemic, err = ParseEpistemic(epistemic)
		if err != nil {
			return Record{}, fmt.Errorf("epistemic: %w", err)
		}
		record.Evidence, err = nullableJSONString(object["evidence"], "evidence")
		if err != nil {
			return Record{}, err
		}
	case KindQuestion:
		record.Recommendation, err = jsonString(object["recommendation"], "recommendation")
		if err != nil {
			return Record{}, err
		}
		record.Evidence, err = nullableJSONString(object["evidence"], "evidence")
		if err != nil {
			return Record{}, err
		}
	case KindDecision:
		record.Reference, err = jsonString(object["authority_reference"], "authority_reference")
		if err != nil {
			return Record{}, err
		}
	}

	record.Bounds, err = nullableJSONString(object["bounds"], "bounds")
	if err != nil {
		return Record{}, err
	}
	if err := record.Validate(); err != nil {
		return Record{}, err
	}
	return record, nil
}

func decodeStrictJSONObject(decoder *json.Decoder) (map[string]json.RawMessage, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	delim, ok := token.(json.Delim)
	if !ok || delim != '{' {
		return nil, errors.New("JSON value must be an object")
	}

	object := make(map[string]json.RawMessage)
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyToken.(string)
		if !ok {
			return nil, errors.New("JSON object key must be a string")
		}
		if _, duplicate := object[key]; duplicate {
			return nil, fmt.Errorf("duplicate JSON key %q", key)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		object[key] = value
	}
	closing, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if closing != json.Delim('}') {
		return nil, errors.New("JSON object is not closed")
	}
	return object, nil
}

func jsonString(raw json.RawMessage, name string) (string, error) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("%s must be a string", name)
	}
	return value, nil
}

func nullableJSONString(raw json.RawMessage, name string) (string, error) {
	var value *string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("%s must be a string or null", name)
	}
	if value == nil {
		return "", nil
	}
	if *value == "" {
		return "", fmt.Errorf("%s must be nonempty or null", name)
	}
	return *value, nil
}

func jsonRecordObject(record Record) map[string]any {
	object := map[string]any{
		"kind":    record.Kind,
		"target":  record.Target,
		"content": record.Content,
		"bounds":  nullableJSONValue(record.Bounds),
	}
	switch record.Kind {
	case KindResult:
		object["epistemic"] = record.Epistemic.String()
		object["evidence"] = nullableJSONValue(record.Evidence)
	case KindQuestion:
		object["recommendation"] = record.Recommendation
		object["evidence"] = nullableJSONValue(record.Evidence)
	case KindDecision:
		object["authority_reference"] = record.Reference
	}
	return object
}

func nullableJSONValue(value string) any {
	if value == "" {
		return nil
	}
	return value
}
