package protocol

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const nullField = "-"

// ParseMessage parses one or more LF-delimited BFV records.
// A single trailing LF and CRLF transport line endings are accepted.
func ParseMessage(raw string) ([]Record, error) {
	if !utf8.ValidString(raw) {
		return nil, errors.New("message is not valid UTF-8")
	}
	if raw == "" {
		return nil, errors.New("message is empty")
	}

	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	if strings.ContainsRune(raw, '\r') {
		return nil, errors.New("bare carriage return is not allowed")
	}
	if strings.HasSuffix(raw, "\n") {
		raw = strings.TrimSuffix(raw, "\n")
	}
	if raw == "" {
		return nil, errors.New("message contains no record")
	}

	lines := strings.Split(raw, "\n")
	records := make([]Record, 0, len(lines))
	for i, line := range lines {
		if line == "" {
			return nil, fmt.Errorf("line %d: blank record is not allowed", i+1)
		}
		record, err := ParseLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		records = append(records, record)
	}
	return records, nil
}

// ParseLine parses one canonical BFV record.
func ParseLine(line string) (Record, error) {
	rawFields, err := splitRawFields(line)
	if err != nil {
		return Record{}, err
	}
	if len(rawFields) == 0 {
		return Record{}, errors.New("record has no fields")
	}

	fields := make([]string, len(rawFields))
	for i, raw := range rawFields {
		value, err := decodeField(raw)
		if err != nil {
			return Record{}, fmt.Errorf("field %d: %w", i+1, err)
		}
		fields[i] = value
	}

	kind := Kind(fields[0])
	var record Record
	switch kind {
	case KindResult:
		if len(fields) != 6 {
			return Record{}, fmt.Errorf("R1 requires 6 fields, got %d", len(fields))
		}
		epistemic, err := ParseEpistemic(fields[2])
		if err != nil {
			return Record{}, fmt.Errorf("epistemic: %w", err)
		}
		record = Record{
			Kind:      kind,
			Target:    fields[1],
			Epistemic: epistemic,
			Content:   fields[3],
			Evidence:  fields[4],
			Bounds:    fields[5],
		}
	case KindQuestion:
		if len(fields) != 6 {
			return Record{}, fmt.Errorf("Q1 requires 6 fields, got %d", len(fields))
		}
		record = Record{
			Kind:           kind,
			Target:         fields[1],
			Content:        fields[2],
			Recommendation: fields[3],
			Evidence:       fields[4],
			Bounds:         fields[5],
		}
	case KindDecision:
		if len(fields) != 5 {
			return Record{}, fmt.Errorf("D1 requires 5 fields, got %d", len(fields))
		}
		record = Record{
			Kind:      kind,
			Target:    fields[1],
			Content:   fields[2],
			Reference: fields[3],
			Bounds:    fields[4],
		}
	default:
		return Record{}, fmt.Errorf("unsupported record kind %q", fields[0])
	}

	if err := record.Validate(); err != nil {
		return Record{}, err
	}
	return record, nil
}

// EncodeMessage validates and emits canonical LF-delimited records.
func EncodeMessage(records []Record) (string, error) {
	if len(records) == 0 {
		return "", errors.New("at least one record is required")
	}
	lines := make([]string, len(records))
	for i, record := range records {
		line, err := EncodeLine(record)
		if err != nil {
			return "", fmt.Errorf("record %d: %w", i+1, err)
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n"), nil
}

// EncodeLine validates and emits one canonical record.
func EncodeLine(record Record) (string, error) {
	if err := record.Validate(); err != nil {
		return "", err
	}

	var fields []string
	switch record.Kind {
	case KindResult:
		fields = []string{
			string(record.Kind),
			record.Target,
			record.Epistemic.String(),
			record.Content,
			record.Evidence,
			record.Bounds,
		}
	case KindQuestion:
		fields = []string{
			string(record.Kind),
			record.Target,
			record.Content,
			record.Recommendation,
			record.Evidence,
			record.Bounds,
		}
	case KindDecision:
		fields = []string{
			string(record.Kind),
			record.Target,
			record.Content,
			record.Reference,
			record.Bounds,
		}
	default:
		return "", fmt.Errorf("unsupported record kind %q", record.Kind)
	}

	encoded := make([]string, len(fields))
	for i, field := range fields {
		encoded[i] = encodeField(field)
	}
	return strings.Join(encoded, "|"), nil
}

func splitRawFields(line string) ([]string, error) {
	fields := make([]string, 0, 6)
	var current strings.Builder
	escaped := false
	for _, r := range line {
		if escaped {
			current.WriteRune('\\')
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if r == '|' {
			fields = append(fields, current.String())
			current.Reset()
			continue
		}
		current.WriteRune(r)
	}
	if escaped {
		return nil, errors.New("trailing escape character")
	}
	fields = append(fields, current.String())
	return fields, nil
}

func decodeField(raw string) (string, error) {
	if raw == "" {
		return "", errors.New("empty field is not allowed; use - for null")
	}
	if raw == nullField {
		return "", nil
	}

	var out strings.Builder
	escaped := false
	for _, r := range raw {
		if !escaped {
			if r == '\\' {
				escaped = true
				continue
			}
			out.WriteRune(r)
			continue
		}

		switch r {
		case '\\':
			out.WriteRune('\\')
		case '|':
			out.WriteRune('|')
		case 'n':
			out.WriteRune('\n')
		case 'r':
			out.WriteRune('\r')
		case 't':
			out.WriteRune('\t')
		case '-':
			out.WriteRune('-')
		default:
			return "", fmt.Errorf("unsupported escape sequence \\%c", r)
		}
		escaped = false
	}
	if escaped {
		return "", errors.New("trailing escape character")
	}
	return out.String(), nil
}

func encodeField(value string) string {
	if value == "" {
		return nullField
	}
	if value == nullField {
		return `\-`
	}

	var out strings.Builder
	for _, r := range value {
		switch r {
		case '\\':
			out.WriteString(`\\`)
		case '|':
			out.WriteString(`\|`)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		default:
			out.WriteRune(r)
		}
	}
	return out.String()
}
