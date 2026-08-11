package phase2

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var (
	// ErrMalformedJSONL identifies input that is not one complete JSON object per line.
	ErrMalformedJSONL = errors.New("malformed Codex JSONL")
	// ErrInvalidCodexEvent identifies a well-formed JSON event with an incompatible schema.
	ErrInvalidCodexEvent = errors.New("invalid Codex event")
	// ErrNoAgentMessage means no completed agent message was present in the event stream.
	ErrNoAgentMessage = errors.New("no completed agent_message event")
	// ErrNoTurnCompleted means no completed turn was present in the event stream.
	ErrNoTurnCompleted = errors.New("no turn.completed event")
)

// CodexUsage is the usage reported by the final turn.completed event. Pointer
// counters preserve the difference between an unreported value and a reported zero.
type CodexUsage struct {
	InputTokens           *int64 `json:"input_tokens"`
	CachedInputTokens     *int64 `json:"cached_input_tokens"`
	CacheWriteInputTokens *int64 `json:"cache_write_input_tokens"`
	OutputTokens          *int64 `json:"output_tokens"`
	ReasoningOutputTokens *int64 `json:"reasoning_output_tokens"`
}

// CodexEventResult contains the final completed agent text and final turn usage.
type CodexEventResult struct {
	FinalText string     `json:"final_text"`
	Usage     CodexUsage `json:"usage"`
}

type codexEventEnvelope struct {
	Type  string          `json:"type"`
	Item  json.RawMessage `json:"item"`
	Usage json.RawMessage `json:"usage"`
}

// ExtractCodexEvents reads Codex CLI JSONL without altering the decoded agent
// text. If more than one matching event exists, the last event of each kind wins.
func ExtractCodexEvents(r io.Reader) (CodexEventResult, error) {
	var result CodexEventResult
	var foundAgentMessage bool
	var foundTurnCompleted bool

	reader := bufio.NewReader(r)
	for lineNumber := 1; ; lineNumber++ {
		line, readErr := reader.ReadBytes('\n')
		if len(line) > 0 {
			if err := consumeCodexEventLine(line, lineNumber, &result, &foundAgentMessage, &foundTurnCompleted); err != nil {
				return CodexEventResult{}, err
			}
		}

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return CodexEventResult{}, fmt.Errorf("read Codex JSONL after line %d: %w", lineNumber, readErr)
		}
	}

	if !foundAgentMessage {
		return CodexEventResult{}, ErrNoAgentMessage
	}
	if !foundTurnCompleted {
		return CodexEventResult{}, ErrNoTurnCompleted
	}
	return result, nil
}

func consumeCodexEventLine(line []byte, lineNumber int, result *CodexEventResult, foundAgentMessage, foundTurnCompleted *bool) error {
	trimmed := bytes.TrimSpace(line)
	if len(trimmed) == 0 {
		return fmt.Errorf("%w at line %d: empty line", ErrMalformedJSONL, lineNumber)
	}
	if trimmed[0] != '{' {
		return fmt.Errorf("%w at line %d: event must be a JSON object", ErrMalformedJSONL, lineNumber)
	}

	var event codexEventEnvelope
	if err := json.Unmarshal(line, &event); err != nil {
		return fmt.Errorf("%w at line %d: %v", ErrMalformedJSONL, lineNumber, err)
	}

	switch event.Type {
	case "item.completed":
		return consumeCompletedItem(event.Item, lineNumber, result, foundAgentMessage)
	case "turn.completed":
		return consumeCompletedTurn(event.Usage, lineNumber, result, foundTurnCompleted)
	default:
		return nil
	}
}

func consumeCompletedItem(rawItem json.RawMessage, lineNumber int, result *CodexEventResult, found *bool) error {
	if len(rawItem) == 0 || bytes.Equal(bytes.TrimSpace(rawItem), []byte("null")) {
		return nil
	}

	var header struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(rawItem, &header); err != nil {
		return fmt.Errorf("%w at line %d: item: %v", ErrInvalidCodexEvent, lineNumber, err)
	}
	if header.Type != "agent_message" {
		return nil
	}

	var message struct {
		Text *string `json:"text"`
	}
	if err := json.Unmarshal(rawItem, &message); err != nil {
		return fmt.Errorf("%w at line %d: agent_message: %v", ErrInvalidCodexEvent, lineNumber, err)
	}
	if message.Text == nil {
		return fmt.Errorf("%w at line %d: agent_message.text is required", ErrInvalidCodexEvent, lineNumber)
	}

	result.FinalText = *message.Text
	*found = true
	return nil
}

func consumeCompletedTurn(rawUsage json.RawMessage, lineNumber int, result *CodexEventResult, found *bool) error {
	var usage CodexUsage
	if len(rawUsage) > 0 && !bytes.Equal(bytes.TrimSpace(rawUsage), []byte("null")) {
		if err := json.Unmarshal(rawUsage, &usage); err != nil {
			return fmt.Errorf("%w at line %d: usage: %v", ErrInvalidCodexEvent, lineNumber, err)
		}
	}

	result.Usage = usage
	*found = true
	return nil
}
