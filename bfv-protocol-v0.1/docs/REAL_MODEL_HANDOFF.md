# Real-model measurement handoff

## Historical base-runner capability

A runner that can provide:

1. a system/developer instruction;
2. one fixture prompt per trial;
3. the raw final model output without renderer modification;
4. usage counters when available;
5. model, reasoning effort, runtime version, and configuration identity.

The runner writes one JSON object per line:

```json
{
  "case_id": "observed-verified",
  "candidate": "pipe-v1",
  "trial": 1,
  "raw": "R1|...",
  "usage": {
    "input_tokens": 0,
    "cached_tokens": 0,
    "output_tokens": 0,
    "reasoning_tokens": 0
  }
}
```

## Evaluation command

```bash
go run ./cmd/bfvctl eval \
  -cases fixtures/cases.jsonl \
  -responses path/to/responses.jsonl \
  -details path/to/details.json
```

## Stop boundary

The historical single-candidate handoff remains reproducible under `artifacts/real-model/`. The authenticated Phase 2 runner now lives under `phase2/` and records producer and downstream-consumer attempts, raw JSONL, exact extracted response text, nullable usage, exit status, stderr, model/config identity, scorer results, defect buckets, and checksums. The completed result is documented in `docs/PHASE2_RESULTS.md`.
