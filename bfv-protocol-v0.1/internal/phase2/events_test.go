package phase2

import (
	"errors"
	"strings"
	"testing"
)

func TestExtractCodexEventsSmokeSchema(t *testing.T) {
	t.Parallel()

	const events = `{"type":"thread.started","thread_id":"019fe9aa-57e1-7c53-af80-9f9f288c63db"}
{"type":"turn.started"}
{"type":"item.started","item":{"id":"item_0","type":"command_execution","command":"/bin/zsh -lc sed","aggregated_output":"","exit_code":null,"status":"in_progress"}}
{"type":"item.completed","item":{"id":"item_0","type":"command_execution","command":"/bin/zsh -lc sed","aggregated_output":"not found\n","exit_code":1,"status":"failed"}}
{"type":"item.completed","item":{"id":"item_2","type":"agent_message","text":"R1|AC:2|OV|TTLは秒として定義されている|config.ts:22|実ファイル確認範囲\nR1|AC:2|OV|TTL値は直接加算されている|cache.ts:81|実ファイル確認範囲"}}
{"type":"turn.completed","usage":{"input_tokens":47010,"cached_input_tokens":37120,"cache_write_input_tokens":0,"output_tokens":649,"reasoning_output_tokens":392}}
`

	got, err := ExtractCodexEvents(strings.NewReader(events))
	if err != nil {
		t.Fatal(err)
	}
	const wantText = "R1|AC:2|OV|TTLは秒として定義されている|config.ts:22|実ファイル確認範囲\nR1|AC:2|OV|TTL値は直接加算されている|cache.ts:81|実ファイル確認範囲"
	if got.FinalText != wantText {
		t.Fatalf("final text changed:\n got: %q\nwant: %q", got.FinalText, wantText)
	}
	assertCounter(t, "input_tokens", got.Usage.InputTokens, 47010)
	assertCounter(t, "cached_input_tokens", got.Usage.CachedInputTokens, 37120)
	assertCounter(t, "cache_write_input_tokens", got.Usage.CacheWriteInputTokens, 0)
	assertCounter(t, "output_tokens", got.Usage.OutputTokens, 649)
	assertCounter(t, "reasoning_output_tokens", got.Usage.ReasoningOutputTokens, 392)
}

func TestExtractCodexEventsUsesLastEventsAndPreservesTextWhitespace(t *testing.T) {
	t.Parallel()

	const events = `{"type":"item.completed","item":{"type":"agent_message","text":"older"}}
{"type":"turn.completed","usage":{"input_tokens":1}}
{"type":"item.completed","item":{"type":"agent_message","text":" \nanswer\n\n "}}
{"type":"turn.completed","usage":{"input_tokens":2}}
`
	got, err := ExtractCodexEvents(strings.NewReader(events))
	if err != nil {
		t.Fatal(err)
	}
	if got.FinalText != " \nanswer\n\n " {
		t.Fatalf("whitespace was not preserved: %q", got.FinalText)
	}
	assertCounter(t, "input_tokens", got.Usage.InputTokens, 2)
}

func TestExtractCodexEventsDistinguishesMissingCounterFromZero(t *testing.T) {
	t.Parallel()

	const events = `{"type":"item.completed","item":{"type":"agent_message","text":"ok"}}
{"type":"turn.completed","usage":{"input_tokens":0,"cache_write_input_tokens":0,"output_tokens":0}}
`
	got, err := ExtractCodexEvents(strings.NewReader(events))
	if err != nil {
		t.Fatal(err)
	}
	assertCounter(t, "input_tokens", got.Usage.InputTokens, 0)
	assertCounter(t, "cache_write_input_tokens", got.Usage.CacheWriteInputTokens, 0)
	assertCounter(t, "output_tokens", got.Usage.OutputTokens, 0)
	if got.Usage.CachedInputTokens != nil {
		t.Fatalf("missing cached_input_tokens became reported: %v", *got.Usage.CachedInputTokens)
	}
	if got.Usage.ReasoningOutputTokens != nil {
		t.Fatalf("missing reasoning_output_tokens became reported: %v", *got.Usage.ReasoningOutputTokens)
	}
}

func TestExtractCodexEventsErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		events string
		want   error
	}{
		{
			name:   "malformed JSONL",
			events: "{\"type\":\"item.completed\"}\nnot-json\n",
			want:   ErrMalformedJSONL,
		},
		{
			name:   "no agent message",
			events: "{\"type\":\"turn.completed\",\"usage\":{}}\n",
			want:   ErrNoAgentMessage,
		},
		{
			name:   "no completed turn",
			events: "{\"type\":\"item.completed\",\"item\":{\"type\":\"agent_message\",\"text\":\"ok\"}}\n",
			want:   ErrNoTurnCompleted,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ExtractCodexEvents(strings.NewReader(tt.events))
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want errors.Is(_, %v)", err, tt.want)
			}
		})
	}
}

func assertCounter(t *testing.T, name string, got *int64, want int64) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s was not reported", name)
	}
	if *got != want {
		t.Fatalf("%s = %d, want %d", name, *got, want)
	}
}
