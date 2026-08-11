# BFV protocol v0.1

Vendor-neutral reference package for a bounded falsification and verification kernel projection and an LLM-to-LLM wire record.

## Current status

- Protocol semantics: specified.
- Go parser/encoder/validator: implemented.
- Semantic evaluator: implemented.
- Deterministic and fuzz checks: available.
- Real-model pipe pilot: complete.
- Phase 2 matched transfer pilot: complete; no overall winner, with `pipe-v1` and `json-v1` held on the safe cost frontier.

## Canonical records

```text
R1|target|epistemic|conclusion|evidence|bounds
Q1|target|request|recommendation|evidence|bounds
D1|target|decision|authority-reference|bounds
```

## Commands

```bash
go test ./...
go vet ./...
go run ./cmd/bfvctl validate -input message.txt
go run ./cmd/bfvctl canonicalize -input message.txt
go run ./cmd/bfvctl eval -cases fixtures/cases.jsonl -responses fixtures/responses.sample.jsonl
```

## Files

- `docs/PROTOCOL.md`: wire semantics and grammar.
- `docs/KERNEL_PROJECTION.md`: compact BFV model projection.
- `docs/EVALUATION.md`: real-model evaluation contract.
- `docs/REAL_MODEL_HANDOFF.md`: external runner boundary.
- `docs/PHASE2_RESULTS.md`: matched candidate metrics, defect separation, and selection result.
- `phase2/`: reproducible matched runner, fixtures, prompts, and scorer.
- `prompts/protocol-instruction.md`: model-facing wire instruction.
- `fixtures/`: semantic cases and sample responses.
