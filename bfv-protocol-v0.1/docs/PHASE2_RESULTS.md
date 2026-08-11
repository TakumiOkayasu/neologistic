# Phase 2 matched candidate results

## Conclusion

The corrected coverage pilot has no overall winner. `pipe-v1` and `json-v1` both preserved every scored semantic field and completed every downstream task, but their observed cost vectors cross. JSON used fewer uncached input tokens, while pipe used fewer cached input, output, and reasoning-output tokens. Without provider-reported monetary cost or an applicable billing coefficient, neither candidate dominates the other on total cost.

`direct-v1` and `nl-v1` completed every closed-choice downstream task but lost the observed-origin component in two semantic receipts each. Those hard failures prevent selection even though some cost components are lower.

This is one corrected observation for each preregistered semantic class, not a population reliability estimate.

## Frozen corrected run

- Canonical artifact: [`artifacts/phase2/20260811T035522.867570000Z`](../artifacts/phase2/20260811T035522.867570000Z/)
- Reviewed score: [`reviewed/score.json`](../artifacts/phase2/20260811T035522.867570000Z/reviewed/score.json)
- Superseded defect-discovery run: [`20260811T032926.172378000Z`](../artifacts/phase2/20260811T032926.172378000Z/)
- Runtime: Codex CLI 0.147.0, `gpt-5.6-sol`, high reasoning, read-only sandbox, no approvals.
- Coverage: 5 fixtures, 4 candidates, 20/20 cells, 35 candidate calls plus one excluded preflight smoke call.
- Retry count: 0 for every candidate.
- Every candidate call exited 0 and reported input, cached-input, output, and reasoning-output tokens.
- Provider monetary cost, internal request count, and separate instruction/task-input token counts were not exposed and remain unavailable rather than zero.

The pass contains one expected-field-copy control and four source-derived cases. The source-derived fixtures contain domain facts without naming the expected `R1`, `Q1`, or `D1` protocol class. They cover serialization, semantic derivation, downstream consumption and task correctness, Q1 false-positive/false-negative routing, epistemic preservation, evidence/bounds association, authentic D1 transfer, and rejection of invented D1 authority.

## Candidate metrics

`input` is the provider's total input counter and includes cached input. Dominance uses the mutually exclusive `uncached = input - cached` and `cached` components, not total input plus its cached subset.

| Candidate | Serialization | Source derivation | Semantic consumption | Downstream task | Hard failures | Input | Uncached | Cached | Output | Reasoning output | Calls / retries |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| `direct-v1` | N/A identity | N/A identity | 3/5 | 5/5 | 2 | 77,609 | 30,761 | 46,848 | 1,017 | 372 | 5 / 0 |
| `nl-v1` | 5/5 | 4/4 | 3/5 | 5/5 | 2 | 154,871 | 78,583 | 76,288 | 1,348 | 308 | 10 / 0 |
| `pipe-v1` | 5/5 | 4/4 | 5/5 | 5/5 | 0 | 156,069 | 64,421 | 91,648 | 1,148 | 269 | 10 / 0 |
| `json-v1` | 5/5 | 4/4 | 5/5 | 5/5 | 0 | 154,162 | 41,010 | 113,152 | 1,356 | 326 | 10 / 0 |

Across all candidates, Q1 false positives and false negatives, evidence failures, bounds failures, D1 transmission failures, D1 authority failures, invented D1 records, runtime failures, and retries were all zero. The four hard-failure cells were epistemic origin loss: `verified` replaced `OV`, preserving state but dropping observed origin. They occurred twice in `direct-v1` and twice in `nl-v1`. The closed downstream choices still happened to be correct in all 20 cells, but that does not erase the semantic hard failures.

## Defect separation

The first pass is retained because it exposed fixture, protocol/codec, scorer, and cost-accounting defects. Review then removed protocol-class cues from source-derived fixtures, aligned the NL codec to its visible grammar, corrected D1 support and surface-prose scoring, and split input into uncached and cached cost components. All four candidates were rerun across the full corrected fixture authority; no first-pass response was substituted.

The canonical reviewed defect buckets contain 0 fixture, 0 protocol, 0 scorer, 4 model, 0 runtime, and 0 unclassified findings. Full causal evidence is in [`reviewed-defects.jsonl`](../artifacts/phase2/20260811T035522.867570000Z/reviewed-defects.jsonl) and [`RUN_RELATION.md`](../artifacts/phase2/20260811T035522.867570000Z/RUN_RELATION.md).

## Phase decision

- Adopt: none as the overall communication standard.
- Hold: `pipe-v1` and `json-v1`; both are on the safe cost frontier.
- Reject for this selection: `direct-v1` and `nl-v1`, because they had semantic hard failures.

Exactly one next discriminating test is defined: run one matched, billing-enabled `derive-autonomous` observation for `pipe-v1` and `json-v1`. Provider-reported total cost must resolve the observed crossing between uncached input and cached/output/reasoning components. No additional reliability count is preselected.
