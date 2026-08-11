# Phase 2 final report

## Result

No candidate wins the full correctness-and-total-cost dominance test.

`pipe-v1` and `json-v1` both achieved 5/5 semantic consumption, 5/5 downstream task correctness, and zero hard failures. Their measured cost vectors cross: JSON used fewer uncached input tokens, while pipe used fewer cached input, output, and reasoning-output tokens. Provider monetary cost was not reported, so neither structured candidate dominates the other.

| Candidate | Serialization | Source derivation | Semantic consumption | Task correctness | Hard failures | Input / uncached / cached / output / reasoning | Calls / retries |
|---|---:|---:|---:|---:|---:|---:|---:|
| `direct-v1` | N/A | N/A | 3/5 | 5/5 | 2 | 77,609 / 30,761 / 46,848 / 1,017 / 372 | 5 / 0 |
| `nl-v1` | 5/5 | 4/4 | 3/5 | 5/5 | 2 | 154,871 / 78,583 / 76,288 / 1,348 / 308 | 10 / 0 |
| `pipe-v1` | 5/5 | 4/4 | 5/5 | 5/5 | 0 | 156,069 / 64,421 / 91,648 / 1,148 / 269 | 10 / 0 |
| `json-v1` | 5/5 | 4/4 | 5/5 | 5/5 | 0 | 154,162 / 41,010 / 113,152 / 1,356 / 326 | 10 / 0 |

The `input` counter includes cached input. Dominance therefore compares `uncached = input - cached` and cached input as mutually exclusive components.

Q1 false positives/negatives, evidence/bounds failures, D1 transmission/authority failures, invented D1, runtime failures, and retries were all zero. Every downstream closed-choice task was correct. The four remaining semantic failures all lost observed origin by emitting `verified` instead of `OV`: two in direct and two in NL.

Provider monetary cost, provider-internal request count, and separate instruction/task-input token counts were not reported. Total cost remains a call/retry/token vector, not an invented scalar or dollar amount.

## Matched evidence

- Frozen runtime: Codex CLI 0.147.0, `gpt-5.6-sol`, high reasoning, read-only, approval `never`.
- Coverage: 20/20 matched cells, 35 candidate calls, plus one preflight smoke call excluded from candidate cost.
- Fixture cohorts: one expected-field-copy control and four source-derived cases with no protocol-class names or expected routing directives in their source facts.
- Raw preservation: all 36 model attempts have byte-identical extracted response text and matching usage data.
- Reviewed scoring: [`reviewed/score.json`](reviewed/score.json).
- Defect evidence: [`reviewed-defects.jsonl`](reviewed-defects.jsonl), with 4 model and 0 fixture/protocol/scorer/runtime/unclassified findings.
- Correction relation: [`RUN_RELATION.md`](RUN_RELATION.md).
- Superseded defect-discovery run: [`../20260811T032926.172378000Z`](../20260811T032926.172378000Z/).

No response from the first run was substituted into this corrected run.

## Decision

- Adopt: none.
- Hold: `pipe-v1` and `json-v1`.
- Reject for this selection: `direct-v1` and `nl-v1`.

The sole next discriminating test is one matched, billing-enabled `derive-autonomous` observation for `pipe-v1` and `json-v1`. Provider-reported total cost must resolve the crossing cost vector; no arbitrary repetition count is added.
