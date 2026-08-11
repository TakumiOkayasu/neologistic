# Retrieval evaluation v0.1

Deterministic scoring for ranked candidate discovery produced by `agent-retrieval-gateway` or a native agent-search baseline.

## Metrics

- availability and stale-index failures;
- recall@k over gold source paths;
- reciprocal rank;
- unsupported absence claims;
- candidate bytes, latency, tool calls, and reported model tokens;
- optional downstream correctness, reported separately.

## Run

```bash
go test ./...
go vet ./...
go run ./cmd/retrieval-evalctl score \
  -fixtures fixtures/cases.sample.jsonl \
  -observations fixtures/observations.sample.jsonl \
  -systems filesystem
```

The sample data is synthetic and validates the scorer only. It is not a retrieval-performance claim.
