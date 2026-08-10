# Final deterministic check

Command: `env GOCACHE=/tmp/bfv-go-build-cache make check`

Exit status: 0.

```text
go test ./...
?   bfvprotocol/cmd/bfvctl [no test files]
ok  bfvprotocol/internal/eval (cached)
ok  bfvprotocol/internal/protocol (cached)
go vet ./...
go run ./cmd/bfvctl eval -cases fixtures/cases.jsonl -responses fixtures/responses.sample.jsonl >/dev/null
```
