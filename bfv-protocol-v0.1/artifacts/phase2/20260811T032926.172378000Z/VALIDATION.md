# Validation record

All checks below passed on 2026-08-11 in the final reviewed workspace:

- `make check` with repository-external Go cache;
- `scripts/check.sh`;
- `go test -race ./...`;
- `go test ./internal/protocol -run=^$ -fuzz=FuzzEncodeParseRoundTrip -fuzztime=5s`, 1,085,476 executions;
- `go vet ./...`;
- `go run ./cmd/phase2ctl validate`, reporting 5 fixtures, 20 cells, and 35 initial candidate calls;
- shell syntax for `phase2/run-pilot.sh`;
- JSON/JSONL parsing for every nonempty artifact data file;
- raw adapter comparison for all 36 model attempts, with 0 response or usage mismatches;
- 35/35 candidate attempt exit statuses equal to zero;
- every required usage counter reported for every candidate attempt;
- reviewed and run-wide SHA-256 manifests.

The preflight smoke call is included in the 36 raw-preservation checks but excluded from candidate costs. Empty JSONL files in unused defect buckets are intentional.
