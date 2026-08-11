# Phase 2 matched candidate results

## Conclusion

The ratified operational winner and default is `json-v1`. `pipe-v1` remains the challenger. Both preserved every scored semantic field and completed every downstream task, so the pinned official standard-rate policy resolves their observed cost-vector crossing: JSON cost $0.302306 and pipe cost $0.402369 for the canonical coverage pass. JSON saved $0.100063, approximately 24.9% relative to pipe, and used 1,699 fewer total input-plus-output tokens.

`direct-v1` and `nl-v1` are rejected. They completed every closed-choice downstream task but lost the observed-origin component in two semantic receipts each. Those hard failures stop them before the cost gate even though some cost components are lower.

This operational decision does not rewrite the earlier Pareto diagnostic. The frozen canonical artifact's `selection.json` and `reviewed/selection.json` remain byte-unchanged at `no_winner`, because neither safe candidate dominates the other in every individual cost-vector component. Future legacy selection fields retain that Pareto meaning, while operational selection is reported separately.

This is one corrected observation for each preregistered semantic class, not a population reliability estimate.

## Frozen corrected run

- Canonical artifact: [`artifacts/phase2/20260811T035522.867570000Z`](../artifacts/phase2/20260811T035522.867570000Z/)
- Reviewed score: [`reviewed/score.json`](../artifacts/phase2/20260811T035522.867570000Z/reviewed/score.json)
- Superseded defect-discovery run: [`20260811T032926.172378000Z`](../artifacts/phase2/20260811T032926.172378000Z/)
- Runtime: Codex CLI 0.147.0, `gpt-5.6-sol`, high reasoning, read-only sandbox, no approvals.
- Coverage: 5 fixtures, 4 candidates, 20/20 cells, 35 candidate calls plus one excluded preflight smoke call.
- Retry count: 0 for every candidate.
- Every candidate call exited 0 and reported input, cached-input, output, and reasoning-output tokens.
- Provider-reported monetary cost, internal request count, and separate instruction/task-input token counts were not exposed and remain unavailable rather than zero. The operational dollar amounts below are deterministic calculations from the reported token counters and the pinned official rates, not provider-reported billing amounts.
- No model call was rerun for the operational decision, and no canonical artifact byte or checksum was changed.

The pass contains one expected-field-copy control and four source-derived cases. The source-derived fixtures contain domain facts without naming the expected `R1`, `Q1`, or `D1` protocol class. They cover serialization, semantic derivation, downstream consumption and task correctness, Q1 false-positive/false-negative routing, epistemic preservation, evidence/bounds association, authentic D1 transfer, and rejection of invented D1 authority.

## Candidate metrics

`input` is the provider's total input counter and includes cached and cache-write input. Mutually exclusive input is `uncached = input - cached - cache write`. Cache-write input was reported as zero in every canonical call.

| Candidate | Serialization | Source derivation | Semantic consumption | Downstream task | Hard failures | Input | Uncached | Cached | Cache write | Output | Reasoning output | Max input / call | Calls / retries |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| `direct-v1` | N/A identity | N/A identity | 3/5 | 5/5 | 2 | 77,609 | 30,761 | 46,848 | 0 | 1,017 | 372 | 15,545 | 5 / 0 |
| `nl-v1` | 5/5 | 4/4 | 3/5 | 5/5 | 2 | 154,871 | 78,583 | 76,288 | 0 | 1,348 | 308 | 15,526 | 10 / 0 |
| `pipe-v1` | 5/5 | 4/4 | 5/5 | 5/5 | 0 | 156,069 | 64,421 | 91,648 | 0 | 1,148 | 269 | 15,753 | 10 / 0 |
| `json-v1` | 5/5 | 4/4 | 5/5 | 5/5 | 0 | 154,162 | 41,010 | 113,152 | 0 | 1,356 | 326 | 15,528 | 10 / 0 |

Across all candidates, Q1 false positives and false negatives, evidence failures, bounds failures, D1 transmission failures, D1 authority failures, invented D1 records, runtime failures, and retries were all zero. The four hard-failure cells were epistemic origin loss: `verified` replaced `OV`, preserving state but dropping observed origin. They occurred twice in `direct-v1` and twice in `nl-v1`. The closed downstream choices still happened to be correct in all 20 cells, but that does not erase the semantic hard failures.

## Operational cost decision

The gate order is complete coverage, zero hard failures, semantic transfer and downstream success, paid-token cost, then total input-plus-output tokens only for an exact dollar-cost tie. Direct and NL fail before cost. JSON and pipe tie on all preceding correctness gates.

The pinned `gpt-5.6-sol-2026-08-11` standard rates are $5.00 per million uncached-input tokens, $0.50 per million cached-input tokens, $6.25 per million cache-write-input tokens, and $30.00 per million output tokens. With `U = I - C - W`:

| Candidate | Cost calculation | Operational cost | Input + output |
|---|---|---:|---:|
| `json-v1` | 41,010 x $5/M + 113,152 x $0.50/M + 0 x $6.25/M + 1,356 x $30/M | $0.302306 | 155,518 |
| `pipe-v1` | 64,421 x $5/M + 91,648 x $0.50/M + 0 x $6.25/M + 1,148 x $30/M | $0.402369 | 157,217 |

Reasoning-output tokens are diagnostic and already included in `output_tokens`, so they are not added again. JSON's $0.100063 advantage decides the operational gate; the 1,699-token advantage is supporting evidence, not the tie-break, because monetary cost did not tie.

These standard rates apply only when every call is at or below 272,000 input tokens. The canonical maximum was 15,753, so this run is in range. Missing, invalid, or over-limit usage blocks an operational winner instead of silently assuming this policy applies.

## Defect separation

The first pass is retained because it exposed fixture, protocol/codec, scorer, and cost-accounting defects. Review then removed protocol-class cues from source-derived fixtures, aligned the NL codec to its visible grammar, corrected D1 support and surface-prose scoring, and split input into uncached and cached cost components. All four candidates were rerun across the full corrected fixture authority; no first-pass response was substituted.

The canonical reviewed defect buckets contain 0 fixture, 0 protocol, 0 scorer, 4 model, 0 runtime, and 0 unclassified findings. Full causal evidence is in [`reviewed-defects.jsonl`](../artifacts/phase2/20260811T035522.867570000Z/reviewed-defects.jsonl) and [`RUN_RELATION.md`](../artifacts/phase2/20260811T035522.867570000Z/RUN_RELATION.md).

## Phase decision

- Operational adopt/default: `json-v1`.
- Challenger: `pipe-v1`.
- Reject: `direct-v1` and `nl-v1`, because they had semantic hard failures.
- Pareto diagnostic: `no_winner`; JSON and pipe remain on its safe frontier.

The pinned rates resolve the previous cost-vector uncertainty without another model run. This decision remains bounded to the observed canonical workload, model/configuration, and applicable rate policy. It does not establish a population failure rate or universal format superiority. See [`PHASE2_DECISION.md`](PHASE2_DECISION.md) for the operational contract and output compatibility boundary.
