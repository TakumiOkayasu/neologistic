#!/usr/bin/env bash
set -euo pipefail

phase2_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export GOCACHE="${GOCACHE:-/tmp/bfv-phase2-go-cache}"

cd "${phase2_root}"
exec go run ./cmd/phase2ctl run \
  -root "${phase2_root}" \
  -config phase2/config.json \
  -fixtures phase2/fixtures/cases.jsonl \
  -artifacts artifacts/phase2
