# Phase 3: End-to-End LLM Communication Value

This directory contains the operational specification for the next matched experiment. The repository-root `PROJECT_ORIGIN.md`, `DECISION_LINEAGE.md`, and `EXPERIMENT_PLAN.md` remain authoritative and are not duplicated here.

## Status

- Project reset: complete.
- Specification integration: complete after this change is committed.
- Deterministic fixtures, hidden gold, scorer, and cost ledger: not implemented yet.
- Paid model and Codex runs: prohibited until the deterministic harness is frozen.
- Retrieval: out of scope.

## Experiment structure

```mermaid
flowchart TD
    O[Original question: does genshijin-like communication help modern LLMs?]
    O --> M[Matched end-to-end experiment]

    M --> A[Arm A: native, BFV off]
    M --> B[Arm B: native, BFV on]
    M --> C[Arm C: genshijin-like, BFV off]
    M --> D[Arm D: genshijin-like, BFV on]
    M --> E[Arm E: json-v1, BFV off]
    M --> F[Arm F: json-v1, BFV on]

    A --> G[Deterministic correctness gates]
    B --> G
    C --> G
    D --> G
    E --> G
    F --> G

    G --> H[Cost and human-intervention ledger]
    H --> I[Operational decision]
    I --> J[Retain or delete each custom mechanism]
```

## Documents

- [Experiment specification](SPEC.md)
- [High-cost delegation intent gate](INTENT_GATE.md)
- [IntentEnvelope schema](schemas/intent-envelope.schema.json)
- [IntentReceipt schema](schemas/intent-receipt.schema.json)
- [Fixture design rules](fixtures/TASK_DESIGN_RULES.md)
- [Three-pass review](REVIEW.md)

## Next implementation boundary

Implement only deterministic fixtures, hidden gold, scorer, and ledgers. Do not invoke Codex or another paid model while those artifacts are still mutable.
