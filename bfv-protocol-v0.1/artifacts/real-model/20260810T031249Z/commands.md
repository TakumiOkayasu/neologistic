# Commands

All paths below are relative to the package root unless absolute.

## Baseline and runtime inspection

```bash
make check
env GOCACHE=/tmp/bfv-go-build-cache make check
command -v codex
codex --version
codex --help
codex exec --help
codex login status
```

The first `make check` was blocked by the managed filesystem's denial of the default Go build cache. The second invocation supplied a writable cache and passed.

## Model invocation configuration

Every successful call used this command shape. `PROMPT`, `EVENTS`, and `STDERR` name per-case artifact files.

```bash
codex exec \
  --ephemeral \
  --ignore-user-config \
  --ignore-rules \
  --model gpt-5.6-sol \
  -c 'model_reasoning_effort="high"' \
  -c 'approval_policy="never"' \
  --sandbox read-only \
  --json \
  --cd /Users/yamazaki/prog/murata-lab/neologistic/bfv-protocol-v0.1 \
  - <"${PROMPT}" >"${EVENTS}" 2>"${STDERR}"
```

The exact batch wrapper is `run-pilot.sh`. It derives one prompt per JSONL fixture and saves the prompt, raw stdout events, stderr, and exit status separately.

```bash
bash artifacts/real-model/20260810T031249Z/run-pilot.sh iteration-2
bash artifacts/real-model/20260810T031249Z/run-pilot.sh \
  iteration-3 \
  artifacts/real-model/20260810T031249Z/iteration-3/affected-cases.jsonl
```

## Event adaptation and scoring

For each successful case, the following command extracts the final agent message and provider usage without modifying the message text.

```bash
jq -cs \
  --arg case_id "${CASE_ID}" \
  --argjson trial "${TRIAL}" \
  -f artifacts/real-model/20260810T031249Z/codex-events-to-response.jq \
  "${EVENTS}"
```

Every response set was scored with this command shape.

```bash
env GOCACHE=/tmp/bfv-go-build-cache go run ./cmd/bfvctl eval \
  -cases "${CASES}" \
  -responses "${RESPONSES}" \
  -details "${DETAILS}" >"${SUMMARY}"
```

## Final verification

```bash
env GOCACHE=/tmp/bfv-go-build-cache make check
```
