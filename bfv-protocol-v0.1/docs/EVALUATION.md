# Evaluation specification

## Decision being tested

Whether a constrained delimiter record produces a better accepted-result-to-total-cost ratio than less structured output for LLM-to-LLM transfer.

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
6. Select a format only when it has no hard failure and improves total cost without reducing downstream task success.

## Candidate comparison

The first real-model run compares:

- concise natural language;
- `pipe-v1` from this package;
- an equivalent JSON structured record when the runtime supports constrained output.

The canonical semantic fixture remains the same across candidates. Rendering differences are not scored as semantic improvements.

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
