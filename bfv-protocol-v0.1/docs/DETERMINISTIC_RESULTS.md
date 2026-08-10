# Deterministic validation result

## Environment

```text
OS: Linux
Go: go1.23.2 linux/amd64
Owner: root
```

## Completed checks

```text
go test ./...
go vet ./...
go test -race ./...
go test ./internal/protocol -run=^$ -fuzz=FuzzEncodeParseRoundTrip -fuzztime=5s
scripts/check.sh
```

Observed result:

- parser/encoder unit tests passed;
- evaluator tests passed;
- race detector passed;
- fuzz round-trip passed with more than 100,000 executions in the recorded run;
- all 8 sample semantic fixtures parsed, matched, and routed without a hard failure;
- protocol package statement coverage: 83.5%;
- evaluator package statement coverage: 85.5%.

The sample fixture result validates the implementation and scorer, not LLM reliability. No model reliability claim is made before real-model trials.
