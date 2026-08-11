# Evaluation specification

## Decision tested

Whether a constrained delimiter record produces a better accepted-result-to-total-cost ratio than less structured output for LLM-to-LLM transfer.

The matched Phase 2 coverage pass supports `json-v1` as the operational default and `pipe-v1` as its challenger under the pinned pricing policy below. This operational choice is separate from the frozen Pareto-vector diagnostic, which remains `no_winner`.

## Mandatory measures

- parse success;
- semantic record equality;
- false and missed `Q1` routing;
- epistemic state preservation;
- evidence preservation;
- instruction, input, output, reasoning, and retry tokens when reported;
- task success after the receiving model consumes the record;
- model/version/config identity.

## Hard failures

- malformed output that cannot be parsed;
- a missed or false `Q1`;
- an unverified result promoted to verified/refuted;
- verified/refuted output without evidence;
- loss of a decision-relevant bound;
- downstream action that differs because required semantics were lost.

## Execution sequence

1. Run deterministic parser, validation, round-trip, and fuzz tests.
2. Run a small real-model pilot across all semantic fixture classes.
3. Inspect failure classes and token distribution; do not infer reliability from average token count alone.
4. Correct the protocol or prompt only when the failure is protocol-caused.
5. Run the comparison with the same tasks, model, reasoning effort, tool access, and source revision.
6. Apply the operational gates in order: complete coverage, zero hard failures, semantic transfer and downstream task success, paid-token cost, then total input-plus-output tokens only when monetary cost ties exactly.

## Candidate comparison

The Phase 2 matched run compares:

- raw-source identity transfer (`direct-v1`);
- concise natural language;
- `pipe-v1` from this package;
- an equivalent prompt-only JSON structured record (`json-v1`).

The canonical semantic fixture remains the same across candidates. Rendering differences are not scored as semantic improvements.

The corrected 20-cell coverage pilot put `json-v1` and `pipe-v1` through every correctness gate with no hard failures. Direct and NL had epistemic-origin hard failures and are rejected before cost is considered. Under the pinned standard rates, JSON is the operational winner and default because its $0.302306 canonical cost is below pipe's $0.402369. See `docs/PHASE2_RESULTS.md` and `docs/PHASE2_DECISION.md`.

## Operational decision rule

Candidates are filtered and ranked in this exact order:

1. require complete preregistered coverage;
2. require zero hard failures;
3. compare semantic transfer and downstream task success, without trading either away for cost;
4. among candidates tied on the preceding gates, choose the lower paid-token cost;
5. only when paid-token cost ties exactly, choose the lower total `input_tokens + output_tokens` count.

A cheaper candidate that failed an earlier gate cannot win. If the final tie remains, there is no operational winner.

For the `gpt-5.6-sol-2026-08-11` policy, define mutually exclusive uncached input as `U = I - C - W`, where `I` is total input, `C` is cached input, and `W` is cache-write input. Cost is:

```text
cost_usd = U * $5.00 / 1,000,000
         + C * $0.50 / 1,000,000
         + W * $6.25 / 1,000,000
         + O * $30.00 / 1,000,000
```

`O` is `output_tokens`. Reported reasoning-output tokens are diagnostic and already included in `output_tokens`, so they are not billed a second time. The standard-rate policy is applicable only when every call is at or below 272,000 input tokens. Missing counters, negative counters, invalid `I = U + C + W` decomposition, arithmetic overflow, or an over-limit call blocks operational selection rather than falling back to an assumed cost.

The canonical maximum was 15,753 input tokens in one call, so the policy applies. The rate source and verification date are pinned in `phase2/config.json`; changing either requires a new operational evaluation of the stored usage.

## Pareto and operational outputs

Pareto selection remains a useful cost-vector diagnostic and a compatibility contract, but it is not the operational default decision. For future generated reports:

- `selection.json`, `score.json.selection`, and CLI `winner` remain the legacy Pareto aliases;
- `pareto-selection.json`, `score.json.pareto_selection`, and CLI `pareto_status` / `pareto_winner` expose the same diagnostic explicitly;
- `operational-selection.json`, `score.json.operational_selection`, and CLI `operational_status` / `operational_winner` / `operational_challenger` / `operational_rejected` expose the rate-card decision separately.

The frozen canonical artifact predates these explicit outputs. Its selection remains `no_winner` and all of its bytes remain unchanged.

Its legacy `next_discriminating_test` is preserved as historical Pareto diagnostic data, not as an active instruction to rerun the model.

## Sample size

Do not invent a universal trial count. The pilot establishes observed variance and failure classes. A later reliability claim must state its target error bound and confidence, then derive the required sample size from that claim.

## Artifacts

Every run stores:

- raw model output;
- parsed record or parse error;
- usage data;
- model/runtime/config identity;
- expected fixture;
- scorer result;
- downstream outcome when measured.

Future scored reports also store separate Pareto and operational selections. Applying the pinned policy to the frozen usage is deterministic post-processing; it does not require or authorize a model rerun.
