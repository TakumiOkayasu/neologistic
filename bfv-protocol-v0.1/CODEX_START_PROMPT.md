# Codex task: run the BFV protocol v0.1 real-model pilot

Work autonomously until a decision reserved to the human is genuinely required.

## Contract

Outcome:
- Produce a reproducible real-model pilot for the BFV `pipe-v1` record in this package.

Acceptance criteria:
- Existing deterministic checks still pass.
- The actual runtime/model/configuration is recorded.
- Raw model output and provider usage are captured without renderer modification.
- Every fixture class in `fixtures/cases.jsonl` is exercised.
- Results are scored with `bfvctl eval`.
- Hard failures are not hidden or averaged away.
- The final recommendation is evidence-backed and does not claim a failure rate unsupported by the observations.

## Required sequence

1. Read `README.md`, `docs/PROTOCOL.md`, `docs/EVALUATION.md`, and `prompts/protocol-instruction.md`.
2. Run `make check` and record the result.
3. Inspect the installed Codex CLI/runtime instead of assuming its current event schema or flags.
4. Perform one smoke invocation and save its unmodified structured output.
5. Implement only the thinnest adapter needed to write `Response` JSON objects accepted by `bfvctl eval`.
6. Run one trial for every fixture to cover all semantic classes.
7. Score the run. Classify every failure as one of:
   - protocol ambiguity;
   - prompt non-compliance;
   - runtime extraction error;
   - fixture defect;
   - model semantic error.
8. Modify the protocol or prompt only when evidence identifies that layer as the cause. Preserve the semantic distinctions and rerun affected fixtures.
9. Continue only while an iteration adds decision-relevant evidence or closes a failure. Stop at the bounded fixed point.
10. Save raw events, normalized responses, scorer details, environment identity, commands, and the final recommendation under `artifacts/real-model/<run-id>/`.

## Hard constraints

- Do not let the model declare `COMPLETED`.
- Do not treat a synthetic counterexample as observed evidence.
- Do not convert an open assertion to verified without traceable evidence.
- Do not issue `Q1` when the agent can resolve the matter through available tools, reversible action, or further verification.
- Do not redesign unrelated layers.
- Do not add arbitrary reliability claims or fixed trial counts.

## Human stop condition

Stop for the human only when:
- credentials or access unavailable to Codex are required;
- an irreversible or materially different policy decision remains;
- accepting a measured cost/reliability tradeoff requires the human's risk ruling.

When stopped, return one exact question with a best-supported recommendation and the evidence path.
