# Deterministic baseline

## First invocation

Command: `make check`

Exit status: 2.

```text
go test ./...
FAIL ./... [setup failed]
# ./...
pattern ./...: open /Users/yamazaki/Library/Caches/go-build/45/453803274dfc0a101a53c5b1a463a040378dc3b2c261047cd89c5d3285bb2fdd-d: operation not permitted
FAIL
make: *** [test] Error 1
```

Classification: environment setup failure. The sandbox denied writes to the default Go build cache before package tests ran.

## Reproducible sandbox-compatible invocation

Command: `env GOCACHE=/tmp/bfv-go-build-cache make check`

Exit status: 0.

```text
go test ./...
?   bfvprotocol/cmd/bfvctl [no test files]
ok  bfvprotocol/internal/eval 0.352s
ok  bfvprotocol/internal/protocol 0.716s
go vet ./...
go run ./cmd/bfvctl eval -cases fixtures/cases.jsonl -responses fixtures/responses.sample.jsonl >/dev/null
```
