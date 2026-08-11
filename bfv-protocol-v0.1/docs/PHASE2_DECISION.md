# Phase 2 operational decision

## Ratified decision

As of 2026-08-11, the operational transfer default is `json-v1`.

| Role | Candidate | Reason |
|---|---|---|
| Winner and default | `json-v1` | Passed every correctness gate and had the lower paid-token cost under the pinned applicable rates. |
| Challenger | `pipe-v1` | Passed every correctness gate, but cost $0.100063 more on the canonical coverage pass. |
| Rejected | `direct-v1`, `nl-v1` | Each lost epistemic origin in two semantic receipts, which is a hard failure. |

The canonical evidence is the frozen corrected run at [`artifacts/phase2/20260811T035522.867570000Z`](../artifacts/phase2/20260811T035522.867570000Z/). This decision is deterministic post-processing of its reviewed correctness results and reported usage. No model was rerun, no response was substituted, and no byte or checksum in the canonical artifact was changed.

## Decision gates

The operational decision applies these gates in order:

1. complete preregistered coverage;
2. zero hard failures;
3. semantic transfer and downstream task success;
4. paid-token cost;
5. total `input_tokens + output_tokens`, only if paid-token cost ties exactly.

Later gates never compensate for an earlier failure. Direct and NL therefore remain rejected even though their downstream closed choices were 5/5. JSON and pipe both had complete coverage, zero hard failures, 5/5 semantic consumption, and 5/5 downstream task success, so they reached the cost gate.

## Pinned pricing policy

The policy identifier is `gpt-5.6-sol-2026-08-11`. Its model, verification date, source, integer nano-dollar coefficients, applicability limit, and reasoning-output treatment are pinned in [`phase2/config.json`](../phase2/config.json). The official source recorded by that policy is the [GPT-5.6-sol model page](https://developers.openai.com/api/docs/models/gpt-5.6-sol).

| Counter | Standard rate |
|---|---:|
| Uncached input | $5.00 / 1M tokens |
| Cached input | $0.50 / 1M tokens |
| Cache-write input | $6.25 / 1M tokens |
| Output | $30.00 / 1M tokens |

Let `I` be total input, `C` cached input, `W` cache-write input, and `O` output. Uncached input is calculated as `U = I - C - W`, and:

```text
cost_usd = U * $5.00 / 1,000,000
         + C * $0.50 / 1,000,000
         + W * $6.25 / 1,000,000
         + O * $30.00 / 1,000,000
```

`reasoning_output_tokens` is diagnostic and already included in `output_tokens`; it is not billed again. Integer nano-dollar arithmetic is used before rendering the decimal amount.

The standard-rate policy is valid only when every call has at most 272,000 input tokens. The largest canonical call had 15,753, so it is in range. Missing counters, negative counters, an invalid `I = U + C + W` decomposition, arithmetic overflow, or an over-limit call blocks operational selection. The scorer does not guess a rate or silently fall back.

## Canonical calculation

Both safe candidates reported zero cache-write tokens.

| Candidate | U | C | W | O | Cost | Input + output |
|---|---:|---:|---:|---:|---:|---:|
| `json-v1` | 41,010 | 113,152 | 0 | 1,356 | $0.302306 | 155,518 |
| `pipe-v1` | 64,421 | 91,648 | 0 | 1,148 | $0.402369 | 157,217 |

JSON saves $0.100063, approximately 24.9% relative to pipe. It also uses 1,699 fewer input-plus-output tokens. The dollar amounts do not tie, so the token-count tie-break is not invoked.

## Pareto compatibility boundary

The Pareto view answers a different question: whether one safe candidate weakly dominates the others in every measured cost-vector component. It remains a useful diagnostic and still reports `no_winner`, because JSON used fewer uncached-input tokens while pipe used fewer cached-input and output tokens.

The frozen artifact keeps that result byte-for-byte. For future generated reports, the compatibility contract is:

| Surface | Meaning |
|---|---|
| `selection.json` | Legacy Pareto alias. |
| `score.json.selection` | Legacy Pareto alias. |
| CLI `winner` with `winner_semantics: "legacy_pareto"` | Legacy Pareto alias. |
| `pareto-selection.json`, `score.json.pareto_selection`, CLI `pareto_status` / `pareto_winner` | Explicit Pareto diagnostic. |
| `operational-selection.json`, `score.json.operational_selection`, CLI `operational_status` / `operational_winner` / `operational_challenger` / `operational_rejected` | Explicit operational decision. |

Consumers that need the runtime default must read the operational output, not reinterpret the legacy `winner` field.

The frozen Pareto object's `next_discriminating_test` is retained as historical diagnostic data so the artifact stays unchanged. It is not the active operational plan and does not authorize a model rerun.

## Evidence boundary and re-evaluation

This run contains one corrected observation for each preregistered semantic class. It supports the operational decision for this model/configuration, workload, and pricing policy; it is not a population reliability estimate and does not establish universal JSON superiority.

No additional model run is required to enact this decision. Re-evaluate the stored usage or run a new matched experiment if the model/configuration, pricing or applicability limit, task distribution, correctness evidence, or transfer formats materially change. Until then, use `json-v1` by default and keep `pipe-v1` as the challenger.
