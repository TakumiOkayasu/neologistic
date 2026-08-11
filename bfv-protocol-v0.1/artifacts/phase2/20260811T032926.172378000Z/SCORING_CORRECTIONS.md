# Phase 2 scoring corrections

> Superseded by the full corrected run at `../20260811T035522.867570000Z/`, after independent review found additional fixture cue leakage and overlapping total/cached input cost accounting.

The initial model calls and raw artifacts are immutable. No candidate response was retried, edited, normalized in place, or selectively replaced.

Manual review demonstrated three evaluation defects before candidate selection:

1. The `nl-v1` template visibly labels lines with `R1:`, `Q1:`, and `D1:`, while the initial decoder omitted those labels. All five producers followed the visible template. The codec was aligned to the frozen prompt.
2. The authorized-decision task tells the consumer to follow the D1, but the hidden oracle selected the unrelated canary R1 as task support. All four consumers selected the D1 and correctly cited its authority. The hidden oracle now selects that D1 and permits null evidence, because D1 has no evidence field.
3. The initial semantic scorer treated sentence-initial capitalization plus terminal punctuation as a hard semantic change. The corrected comparison tolerates only a one-rune sentence-case difference and terminal sentence punctuation; internal punctuation, identifiers, evidence, authority references, targets, and structural syntax remain exact.

The consumer decoder additionally maps only complete natural-language epistemic pairs such as `inferred and unverified` to `IU`. A lone state such as `verified` is not repaired, because it loses the origin dimension. Those remaining failures are model defects.

The initial root-level `score.json`, `candidate-summaries.json`, `selection.json`, and `defects/` remain the original first scoring pass. Corrected results are written under `reviewed/`, and `reviewed-defects.jsonl` records the evidence-backed causal separation.
