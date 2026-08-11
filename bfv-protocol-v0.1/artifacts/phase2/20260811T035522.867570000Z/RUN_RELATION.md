# Run relation and defect boundary

This run is the canonical Phase 2 correction pass. It replaces the first-pass conclusion, not its evidence.

The earlier run at `../20260811T032926.172378000Z/` demonstrated one hidden-oracle defect, one NL codec/protocol mismatch, two scorer defects, and source-derived fixture cue leakage. Its raw events, prompts, responses, usage, initial scores, and reviewed defect analysis remain immutable.

Before this correction pass:

- source-derived fixture text was rewritten to contain domain facts without naming `R1`, `Q1`, or `D1` or instructing the producer which protocol class to emit;
- the NL codec was aligned with the visible labeled-sentence grammar;
- D1 support without an evidence field and sentence-initial capitalization were scored semantically;
- total input was split into mutually exclusive uncached and cached input components for cost dominance.

All four candidates were then rerun across the complete five-fixture authority under the same model, reasoning effort, sandbox, approval, adapter, and execution policy. No first-run candidate response was substituted into this run.
