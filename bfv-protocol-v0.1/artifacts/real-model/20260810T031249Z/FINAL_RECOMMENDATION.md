# Final recommendation

Proceed with `pipe-v1` to the controlled candidate-comparison stage, but do not select it as superior yet.

## Evidence

- Deterministic parser, validation, evaluator, vet, and sample-eval checks pass.
- Codex CLI 0.147.0 with `gpt-5.6-sol`, high reasoning effort, read-only sandbox, and no approvals exercised every semantic fixture class.
- The smoke observation was parseable but semantically inexact and is retained as one hard failure.
- The first complete run had 8 hard failures in 8 observations. Every one was classified as a fixture defect because strict expected fields were absent or ambiguous in the model input.
- After the fixture supplied all decision-relevant field values, the second run had 4 hard failures in 8 observations. Each was a single appended Japanese full stop at an unquoted field boundary and was classified as a fixture defect.
- After JSON literals made those four boundaries unambiguous, the affected-fixture rerun had 0 hard failures in 4 observations.
- The final corrected-fixture coverage combines the four unaffected second-run observations and four affected-fixture reruns: 8 parse successes, 8 canonical exact matches, 8 semantic exact matches, no routing, epistemic, or evidence failures, and 0 hard failures.

## Usage reported by the provider

| Observation set | Calls | Input | Cached input | Output | Reasoning output | Hard failures |
|---|---:|---:|---:|---:|---:|---:|
| Smoke | 1 | 47,010 | 37,120 | 649 | 392 | 1 |
| Initial complete run | 8 | 123,182 | 59,392 | 1,419 | 882 | 8 |
| Fixture-explicit run | 8 | 123,703 | 88,064 | 553 | 184 | 4 |
| Boundary-clarified affected rerun | 4 | 61,856 | 35,840 | 231 | 36 | 0 |
| Final corrected-fixture coverage | 8 | 123,726 | 96,256 | 462 | 99 | 0 |

The final row is a selected coverage set, not additional provider calls and not a reliability estimate.

## Decision boundary

This pilot provides one successful corrected observation per semantic class. It does not estimate a population failure rate. It also does not measure concise natural language or equivalent JSON candidates, downstream task success, or comparative total cost. The evaluation contract therefore supports advancing `pipe-v1` to a matched comparison, not declaring it the winning format.

No protocol or shared prompt change was supported by the evidence. Only fixture inputs were corrected so the strict scorer's expected fields were actually specified.
