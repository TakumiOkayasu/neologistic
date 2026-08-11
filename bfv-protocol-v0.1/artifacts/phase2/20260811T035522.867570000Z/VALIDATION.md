# Validation record

The following checks passed for the canonical correction pass and final workspace:

- preflight version, help, login, jq, JSONL smoke, and raw adapter checks;
- 20/20 candidate coverage cells and 35/35 candidate calls with exit status zero;
- no model tool calls, runtime failures, or retries;
- all required token counters reported for all 35 candidate calls;
- 36/36 response and usage extractions identical to their raw JSONL events;
- every nonempty JSON or JSONL artifact parsed successfully;
- `make check` and `scripts/check.sh`;
- `go test -race ./...`;
- `go test ./internal/protocol -run=^$ -fuzz=FuzzEncodeParseRoundTrip -fuzztime=5s`;
- `go vet ./...`;
- `go run ./cmd/phase2ctl validate`;
- shell syntax, JSON syntax, repository diff, and artifact SHA-256 manifests.

The preflight smoke call is included in raw-preservation checks but excluded from every candidate cost. Empty JSONL files in unused defect buckets are intentional.
