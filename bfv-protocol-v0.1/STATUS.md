# Phase status

## Closed

- BFV model-facing projection v0.1
- LLM wire semantics v0.1
- strict delimiter codec
- semantic/routing/epistemic evaluator
- deterministic, race, and fuzz validation
- real-model runner contract
- authenticated `pipe-v1` real-model coverage pilot
- Phase 2 matched 20-cell candidate coverage pilot

## Phase 2 result

- Overall winner: none.
- Hold: `pipe-v1` and `json-v1`; both had zero hard failures, while their uncached/cached/output/reasoning cost components crossed.
- Reject for this selection: `direct-v1` and `nl-v1` because of epistemic-origin hard failures.
- Q1, evidence, bounds, D1 transmission/authority, runtime, and retry failures: zero.
- Evidence: `docs/PHASE2_RESULTS.md` and canonical artifact `artifacts/phase2/20260811T035522.867570000Z/`.

## Next discriminating test

Run exactly one matched, billing-enabled `derive-autonomous` observation for `pipe-v1` and `json-v1` so provider-reported total cost can resolve the observed cost-vector crossing.

## Human-only decision boundary

No human decision or credential is currently required. The next test is already specified by the observed correctness/cost frontier.
