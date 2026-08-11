# BFV protocol v0.1

Vendor-neutral reference package for a bounded falsification and verification kernel projection and an LLM-to-LLM wire record.

## Current status

- Protocol semantics: specified.
- Go parser/encoder/validator: implemented.
- Semantic evaluator: implemented.
- Deterministic and fuzz checks: available.
- Real-model pipe pilot: complete.
- Phase 2 matched transfer pilot: complete.
- Operational transfer default: `json-v1`; challenger: `pipe-v1`; rejected: `direct-v1` and `nl-v1`.
- Operational `json-v1` parser, canonical encoder, model instruction, and CLI format selection: implemented.
- Frozen Pareto diagnostic: no winner; the canonical artifact and its legacy selection remain byte-unchanged.

## Wire formats

Use `json-v1` for new operational integrations. Its exact object shape and validation rules are specified in [`docs/PROTOCOL.md`](docs/PROTOCOL.md), and the standalone model instruction is [`prompts/json-v1-instruction.md`](prompts/json-v1-instruction.md).

`pipe-v1` remains a compatibility format with these canonical records:

```text
R1|target|epistemic|conclusion|evidence|bounds
Q1|target|request|recommendation|evidence|bounds
D1|target|decision|authority-reference|bounds
```

## Commands

```bash
go test ./...
go vet ./...
go run ./cmd/bfvctl validate -format json-v1 -input message.json
go run ./cmd/bfvctl canonicalize -format json-v1 -input message.json
go run ./cmd/bfvctl eval -cases fixtures/cases.jsonl -responses fixtures/responses.sample.jsonl
```

The `validate` and `canonicalize` commands still default to `pipe-v1` when `-format` is omitted, so existing callers do not change behavior. New callers should select `-format json-v1` explicitly. `canonicalize` parses and emits the selected format; it does not convert between formats.

`validate` enforces wire/schema semantics only. It does not authenticate `D1.authority_reference`; the transport or consumer must resolve that reference against an authenticated source or allowlist before acting.

## Files

- `docs/PROTOCOL.md`: wire semantics and grammar.
- `docs/KERNEL_PROJECTION.md`: compact BFV model projection.
- `docs/EVALUATION.md`: real-model evaluation contract.
- `docs/REAL_MODEL_HANDOFF.md`: external runner boundary.
- `docs/PHASE2_RESULTS.md`: matched candidate metrics, defect separation, and selection result.
- `docs/PHASE2_DECISION.md`: ratified operational default, pricing policy, and compatibility boundary.
- `phase2/`: reproducible matched runner, fixtures, prompts, and scorer.
- `prompts/json-v1-instruction.md`: operational model-facing JSON instruction.
- `prompts/protocol-instruction.md`: compatibility `pipe-v1` model instruction used by the historical evaluation.
- `fixtures/`: semantic cases and sample responses.

The Phase 1 and Phase 2 model artifacts and their captured prompts remain frozen evidence. Operational adoption adds a new instruction and runtime surface; it does not retrofit those historical inputs or imply a population reliability estimate.
