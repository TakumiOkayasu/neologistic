#!/usr/bin/env bash
set -euo pipefail

go test ./...
go vet ./...
go run ./cmd/bfvctl eval \
  -cases fixtures/cases.jsonl \
  -responses fixtures/responses.sample.jsonl \
  -details /tmp/bfv-protocol-eval-details.json >/tmp/bfv-protocol-eval-summary.json
