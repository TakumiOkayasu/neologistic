package phase2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"bfvprotocol/internal/protocol"
)

var jsonToken = `(?:"(?:\\.|[^"\\])*"|null)`

var (
	nlResult = regexp.MustCompile(`^R1: For target (` + jsonToken + `), the (observed|inferred|synthetic) and (unverified|verified|refuted) result is (` + jsonToken + `)\. Evidence: (` + jsonToken + `)\. Bounds: (` + jsonToken + `)\.$`)
	nlQ1     = regexp.MustCompile(`^Q1: For target (` + jsonToken + `), human input is required: (` + jsonToken + `)\. Recommendation: (` + jsonToken + `)\. Evidence: (` + jsonToken + `)\. Blocking boundary: (` + jsonToken + `)\.$`)
	nlD1     = regexp.MustCompile(`^D1: For target (` + jsonToken + `), the authorized decision is (` + jsonToken + `)\. Authority reference: (` + jsonToken + `)\. Bounds: (` + jsonToken + `)\.$`)
)

func ParseCandidate(candidate, raw string) ([]Record, string, error) {
	switch candidate {
	case CandidateNL:
		records, err := parseNL(raw)
		if err != nil {
			return nil, "", err
		}
		canonical, _ := EncodeCandidate(candidate, records)
		return records, canonical, nil
	case CandidatePipe:
		records, err := parsePipe(raw)
		if err != nil {
			return nil, "", err
		}
		canonical, _ := EncodeCandidate(candidate, records)
		return records, canonical, nil
	case CandidateJSON:
		records, err := parseJSONRecords(raw)
		if err != nil {
			return nil, "", err
		}
		canonical, _ := EncodeCandidate(candidate, records)
		return records, canonical, nil
	default:
		return nil, "", fmt.Errorf("candidate %q has no producer serialization parser", candidate)
	}
}

func EncodeCandidate(candidate string, records []Record) (string, error) {
	for i := range records {
		if err := ValidateRecord(records[i]); err != nil {
			return "", fmt.Errorf("record %d: %w", i+1, err)
		}
	}
	switch candidate {
	case CandidateNL:
		lines := make([]string, len(records))
		for i, record := range records {
			lines[i] = encodeNLRecord(record)
		}
		return strings.Join(lines, "\n"), nil
	case CandidatePipe:
		lines := make([]string, len(records))
		for i, record := range records {
			lines[i] = encodePipeRecord(record)
		}
		return strings.Join(lines, "\n"), nil
	case CandidateJSON:
		objects := make([]map[string]any, len(records))
		for i := range records {
			objects[i] = recordObject(records[i])
		}
		encoded, err := json.Marshal(map[string]any{"records": objects})
		return string(encoded), err
	default:
		return "", fmt.Errorf("unsupported candidate %q", candidate)
	}
}

func ParseConsumer(raw string) (ConsumerOutput, error) {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	var object map[string]json.RawMessage
	if err := decoder.Decode(&object); err != nil {
		return ConsumerOutput{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return ConsumerOutput{}, fmt.Errorf("more than one JSON value")
	}
	if len(object) != 2 || object["records"] == nil || object["task"] == nil {
		return ConsumerOutput{}, fmt.Errorf("consumer output requires exactly records and task")
	}
	var rawRecords []json.RawMessage
	if err := json.Unmarshal(object["records"], &rawRecords); err != nil {
		return ConsumerOutput{}, fmt.Errorf("records: %w", err)
	}
	output := ConsumerOutput{Records: make([]Record, len(rawRecords))}
	for i := range rawRecords {
		record, err := decodeRecord(rawRecords[i], false)
		if err != nil {
			return ConsumerOutput{}, fmt.Errorf("record %d: %w", i+1, err)
		}
		record.Epistemic = normalizeConsumerEpistemic(record.Epistemic)
		output.Records[i] = record
	}
	if err := decodeJSONMap(object["task"], &output.Task); err != nil {
		return ConsumerOutput{}, fmt.Errorf("task: %w", err)
	}
	if err := validateTaskObject(output.Task); err != nil {
		return ConsumerOutput{}, err
	}
	if epistemic, ok := output.Task["support_epistemic"].(string); ok {
		output.Task["support_epistemic"] = normalizeConsumerEpistemic(epistemic)
	}
	return output, nil
}

func EncodeConsumer(records []Record, task map[string]any) (string, error) {
	objects := make([]map[string]any, len(records))
	for i := range records {
		objects[i] = recordObject(records[i])
	}
	return marshalCanonical(map[string]any{"records": objects, "task": task})
}

func parseJSONRecords(raw string) ([]Record, error) {
	decoder := json.NewDecoder(strings.NewReader(raw))
	var object map[string]json.RawMessage
	if err := decoder.Decode(&object); err != nil {
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
		return nil, err
	}
	if len(rawRecords) == 0 {
		return nil, fmt.Errorf("records must not be empty")
	}
	records := make([]Record, len(rawRecords))
	for i := range rawRecords {
		record, err := decodeRecord(rawRecords[i], true)
		if err != nil {
			return nil, fmt.Errorf("record %d: %w", i+1, err)
		}
		records[i] = record
	}
	return records, nil
}

func parseNL(raw string) ([]Record, error) {
	if raw == "" || strings.HasSuffix(raw, "\n") {
		return nil, fmt.Errorf("nl-v1 must be nonempty and have no trailing LF")
	}
	lines := strings.Split(raw, "\n")
	records := make([]Record, 0, len(lines))
	for index, line := range lines {
		var record Record
		switch {
		case nlResult.MatchString(line):
			parts := nlResult.FindStringSubmatch(line)
			record.Kind = "R1"
			record.Target = mustJSONString(parts[1])
			record.Epistemic = map[string]string{"observed": "O", "inferred": "I", "synthetic": "S"}[parts[2]] + map[string]string{"unverified": "U", "verified": "V", "refuted": "R"}[parts[3]]
			record.Content = mustJSONString(parts[4])
			record.Evidence = nullableJSONString(parts[5])
			record.Bounds = nullableJSONString(parts[6])
		case nlQ1.MatchString(line):
			parts := nlQ1.FindStringSubmatch(line)
			record = Record{Kind: "Q1", Target: mustJSONString(parts[1]), Content: mustJSONString(parts[2]), Recommendation: mustJSONString(parts[3]), Evidence: nullableJSONString(parts[4]), Bounds: nullableJSONString(parts[5])}
		case nlD1.MatchString(line):
			parts := nlD1.FindStringSubmatch(line)
			record = Record{Kind: "D1", Target: mustJSONString(parts[1]), Content: mustJSONString(parts[2]), AuthorityReference: mustJSONString(parts[3]), Bounds: nullableJSONString(parts[4])}
		default:
			return nil, fmt.Errorf("line %d does not match nl-v1", index+1)
		}
		if err := ValidateRecord(record); err != nil {
			return nil, fmt.Errorf("line %d: %w", index+1, err)
		}
		records = append(records, record)
	}
	return records, nil
}

func encodeNLRecord(record Record) string {
	q := func(value string) string { encoded, _ := json.Marshal(value); return string(encoded) }
	n := func(value *string) string {
		if value == nil {
			return "null"
		}
		return q(*value)
	}
	switch record.Kind {
	case "R1":
		origin := map[byte]string{'O': "observed", 'I': "inferred", 'S': "synthetic"}[record.Epistemic[0]]
		state := map[byte]string{'U': "unverified", 'V': "verified", 'R': "refuted"}[record.Epistemic[1]]
		return fmt.Sprintf("R1: For target %s, the %s and %s result is %s. Evidence: %s. Bounds: %s.", q(record.Target), origin, state, q(record.Content), n(record.Evidence), n(record.Bounds))
	case "Q1":
		return fmt.Sprintf("Q1: For target %s, human input is required: %s. Recommendation: %s. Evidence: %s. Blocking boundary: %s.", q(record.Target), q(record.Content), q(record.Recommendation), n(record.Evidence), n(record.Bounds))
	default:
		return fmt.Sprintf("D1: For target %s, the authorized decision is %s. Authority reference: %s. Bounds: %s.", q(record.Target), q(record.Content), q(record.AuthorityReference), n(record.Bounds))
	}
}

func parsePipe(raw string) ([]Record, error) {
	parsed, err := protocol.ParseMessage(raw)
	if err != nil {
		return nil, err
	}
	records := make([]Record, len(parsed))
	for i := range parsed {
		records[i] = fromProtocolRecord(parsed[i])
	}
	return records, nil
}

func encodePipeRecord(record Record) string {
	line, _ := protocol.EncodeLine(toProtocolRecord(record))
	return line
}
func stringPointer(value string) *string { return &value }

func mustJSONString(raw string) string {
	var value string
	_ = json.Unmarshal([]byte(raw), &value)
	return value
}

func nullableJSONString(raw string) *string {
	if raw == "null" {
		return nil
	}
	value := mustJSONString(raw)
	return &value
}

func decodeJSONMap(raw []byte, target *map[string]any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	return decoder.Decode(target)
}

func marshalCanonical(value any) (string, error) {
	encoded, err := json.Marshal(value)
	return string(encoded), err
}

func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func recordObject(record Record) map[string]any {
	object := map[string]any{"kind": record.Kind, "target": record.Target, "content": record.Content, "bounds": record.Bounds}
	switch record.Kind {
	case "R1":
		object["epistemic"] = record.Epistemic
		object["evidence"] = record.Evidence
	case "Q1":
		object["recommendation"] = record.Recommendation
		object["evidence"] = record.Evidence
	case "D1":
		object["authority_reference"] = record.AuthorityReference
	}
	return object
}

func decodeRecord(raw json.RawMessage, validateSemantics bool) (Record, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return Record{}, err
	}
	var kind string
	if err := json.Unmarshal(object["kind"], &kind); err != nil {
		return Record{}, fmt.Errorf("kind is required")
	}
	var required []string
	switch kind {
	case "R1":
		required = []string{"kind", "target", "epistemic", "content", "evidence", "bounds"}
	case "Q1":
		required = []string{"kind", "target", "content", "recommendation", "evidence", "bounds"}
	case "D1":
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
	var record Record
	if err := json.Unmarshal(raw, &record); err != nil {
		return Record{}, err
	}
	if validateSemantics {
		if err := ValidateRecord(record); err != nil {
			return Record{}, err
		}
	}
	return record, nil
}

func validateTaskObject(task map[string]any) error {
	keys := []string{"choice", "route", "support_target", "support_epistemic", "support_evidence", "authority_reference"}
	if len(task) != len(keys) {
		return fmt.Errorf("task requires exactly keys %v", keys)
	}
	for _, key := range keys {
		if _, ok := task[key]; !ok {
			return fmt.Errorf("task requires key %q", key)
		}
	}
	for _, key := range []string{"choice", "route", "support_target"} {
		value, ok := task[key].(string)
		if !ok || value == "" {
			return fmt.Errorf("task.%s must be a nonempty string", key)
		}
	}
	if route := task["route"].(string); route != "continue" && route != "ask-human" {
		return fmt.Errorf("task.route must be continue or ask-human")
	}
	for _, key := range []string{"support_epistemic", "support_evidence", "authority_reference"} {
		if task[key] != nil {
			value, ok := task[key].(string)
			if !ok {
				return fmt.Errorf("task.%s must be a string or null", key)
			}
			if value == "" {
				return fmt.Errorf("task.%s must be nonempty when present", key)
			}
		}
	}
	return nil
}

// normalizeConsumerEpistemic decodes only complete natural-language pairs.
// A lone state such as "verified" deliberately remains unchanged so the
// scorer can report loss of the origin dimension rather than repairing it.
func normalizeConsumerEpistemic(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, ",", " and")
	normalized = strings.Join(strings.Fields(normalized), " ")
	parts := strings.Split(normalized, " and ")
	if len(parts) != 2 {
		return value
	}
	origin := map[string]string{"observed": "O", "inferred": "I", "synthetic": "S"}[parts[0]]
	state := map[string]string{"unverified": "U", "verified": "V", "refuted": "R"}[parts[1]]
	if origin == "" || state == "" {
		return value
	}
	return origin + state
}

func fromProtocolRecord(record protocol.Record) Record {
	result := Record{Kind: string(record.Kind), Target: record.Target, Content: record.Content}
	if record.Bounds != "" {
		result.Bounds = stringPointer(record.Bounds)
	}
	switch record.Kind {
	case protocol.KindResult:
		result.Epistemic = record.Epistemic.String()
		if record.Evidence != "" {
			result.Evidence = stringPointer(record.Evidence)
		}
	case protocol.KindQuestion:
		result.Recommendation = record.Recommendation
		if record.Evidence != "" {
			result.Evidence = stringPointer(record.Evidence)
		}
	case protocol.KindDecision:
		result.AuthorityReference = record.Reference
	}
	return result
}

func toProtocolRecord(record Record) protocol.Record {
	result := protocol.Record{Kind: protocol.Kind(record.Kind), Target: record.Target, Content: record.Content, Recommendation: record.Recommendation, Reference: record.AuthorityReference}
	if record.Evidence != nil {
		result.Evidence = *record.Evidence
	}
	if record.Bounds != nil {
		result.Bounds = *record.Bounds
	}
	if record.Kind == "R1" {
		result.Epistemic, _ = protocol.ParseEpistemic(record.Epistemic)
	}
	return result
}
