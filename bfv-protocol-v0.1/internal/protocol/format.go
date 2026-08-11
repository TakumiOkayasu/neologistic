package protocol

import "fmt"

// Format identifies a supported BFV transfer serialization.
type Format string

const (
	FormatPipeV1 Format = "pipe-v1"
	FormatJSONV1 Format = "json-v1"
)

// Parse decodes a BFV message in the requested transfer format.
func Parse(format Format, raw string) ([]Record, error) {
	switch format {
	case FormatPipeV1:
		return ParseMessage(raw)
	case FormatJSONV1:
		return ParseJSONMessage(raw)
	default:
		return nil, fmt.Errorf("unsupported protocol format %q", format)
	}
}

// Encode validates and encodes BFV records in the requested transfer format.
func Encode(format Format, records []Record) (string, error) {
	switch format {
	case FormatPipeV1:
		return EncodeMessage(records)
	case FormatJSONV1:
		return EncodeJSONMessage(records)
	default:
		return "", fmt.Errorf("unsupported protocol format %q", format)
	}
}
