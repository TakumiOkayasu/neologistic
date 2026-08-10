# Real-model measurement handoff

## Required external capability

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

The vendor-neutral package is complete when deterministic checks pass. Real-model measurement requires a selected model/runtime with usable credentials or an already connected CLI/API.
