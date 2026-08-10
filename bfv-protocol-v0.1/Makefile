.PHONY: check test vet sample-eval

check: test vet sample-eval

test:
	go test ./...

vet:
	go vet ./...

sample-eval:
	go run ./cmd/bfvctl eval -cases fixtures/cases.jsonl -responses fixtures/responses.sample.jsonl >/dev/null
