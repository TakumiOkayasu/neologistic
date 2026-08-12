.PHONY: check fmt test vet race hygiene validate freeze verify-freeze preflight manifest-smoke artifact-smoke

check: fmt test vet race hygiene validate verify-freeze preflight manifest-smoke artifact-smoke

fmt:
	@test -z "$$(gofmt -l ./cmd ./internal)" || { gofmt -l ./cmd ./internal; exit 1; }

test:
	go test ./...

vet:
	go vet ./...

race:
	go test -race ./...

hygiene:
	python3 scripts/hygiene.py

validate:
	go run ./cmd/phase3ctl validate

freeze:
	go run ./cmd/phase3ctl freeze

verify-freeze:
	go run ./cmd/phase3ctl verify-freeze

preflight:
	go run ./cmd/phase3ctl preflight

manifest-smoke:
	go run ./cmd/phase3ctl manifest \
		-run-id deterministic-smoke \
		-sender-provider test -sender-model sender -sender-effort fixed \
		-receiver-provider test -receiver-model receiver -receiver-effort fixed \
		-transport isolated-text-harness >/dev/null

artifact-smoke:
	@dir="$$(mktemp -d ./artifact-smoke.XXXXXX)"; \
	trap 'rm -rf "$$dir"' EXIT HUP INT TERM; \
	printf '%s\n' sample > "$$dir/response.txt"; \
	go run ./cmd/phase3ctl artifact-freeze -directory "$$dir" >/dev/null; \
	go run ./cmd/phase3ctl artifact-verify -directory "$$dir" >/dev/null
