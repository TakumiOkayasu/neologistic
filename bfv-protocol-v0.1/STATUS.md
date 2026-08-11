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
- Phase 2 operational pricing decision
- Phase 3 operational `json-v1` codec and CLI format selection

## Phase 2 operational result

- Winner and default: `json-v1`.
- Challenger: `pipe-v1`.
- Reject: `direct-v1` and `nl-v1`, because each had two epistemic-origin hard failures.
- The two safe candidates tied on observed semantic transfer and downstream success. At the pinned official standard rates, `json-v1` cost $0.302306 and `pipe-v1` cost $0.402369 for the canonical coverage pass. The saving is $0.100063, approximately 24.9%, and 1,699 total input-plus-output tokens.
- Q1, evidence, bounds, D1 transmission/authority, runtime, and retry failures were zero.
- The standard-rate policy applies only at or below 272,000 input tokens per call. The canonical maximum was 15,753.
- No model rerun was used to make this decision, and the result is not a population reliability claim.

## Diagnostic compatibility

The frozen artifact's Pareto-vector result remains `no_winner`, and the canonical artifact is byte-unchanged. For future reports, legacy `selection.json`, `score.json.selection`, and CLI `winner` remain aliases for the Pareto diagnostic. The operational decision is emitted separately. See `docs/PHASE2_DECISION.md` and `docs/PHASE2_RESULTS.md`.

## Operational interface

- New integrations use `json-v1` through `bfvctl validate -format json-v1`, `bfvctl canonicalize -format json-v1`, and `prompts/json-v1-instruction.md`.
- Omitting `-format` continues to select `pipe-v1`, preserving the established CLI contract.
- `canonicalize` emits the same format it parses; this surface does not perform cross-format conversion.
- `validate` does not authenticate `D1` authority references; downstream enforcement must resolve them against an authenticated source or allowlist.
- Phase 1 and Phase 2 artifacts and captured prompts remain frozen. No model rerun was required for operationalization, and the observed coverage is not a population reliability claim.

## Human-only decision boundary

No human decision, credential, or additional model run is currently required. Re-evaluation is needed only if the model, pricing policy, workload, correctness evidence, or standard-rate applicability changes.
