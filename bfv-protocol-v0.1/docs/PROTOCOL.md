# BFV LLM wire protocol v0.1

## Purpose

Carry only decision-relevant conclusions, human-only blockers, and authoritative human decisions between LLM-facing components.

## Grammar

```text
message  = record *(LF record) [LF]
record   = result / question / decision
result   = "R1" "|" target "|" epistemic "|" conclusion "|" evidence "|" bounds
question = "Q1" "|" target "|" request "|" recommendation "|" evidence "|" bounds
decision = "D1" "|" target "|" decision-text "|" authority-reference "|" bounds
```

`R1`, `Q1`, and `D1` combine record kind and protocol version.

## Semantics

### R1

An atomic proposition.

`epistemic` is `origin + state`:

| Origin | Meaning |
|---|---|
| `O` | observed |
| `I` | inferred |
| `S` | synthetic hypothesis or counterexample candidate |

| State | Meaning |
|---|---|
| `U` | unverified/open |
| `V` | verified within `bounds` |
| `R` | refuted within `bounds` |

`V` and `R` require traceable `evidence`.

### Q1

The only wire record that may suspend autonomous progress. It is valid only when progress requires human-only authority, inaccessible information, or an irreducible subjective/irreversible choice.

A `Q1` must include:

- one exact request;
- the agent's best-supported recommendation;
- evidence when available;
- the exact blocking boundary.

### D1

Transmits a human or pre-authorized policy decision. An LLM may forward a `D1` but may not invent one.

## Null and escaping

`-` is the null field marker. A literal single hyphen is `\-`.

| Value | Encoding |
|---|---|
| `\` | `\\` |
| `|` | `\|` |
| LF | `\n` |
| CR | `\r` |
| TAB | `\t` |
| literal single `-` | `\-` |

Unknown escape sequences are invalid. Field count is fixed by record kind. A malformed record is rejected rather than guessed.

## Excluded from the wire

Task completion, sender, receiver, run ID, contract version, source revision, and timestamp belong to transport/enforcement metadata when available.

Rejected candidate work is not transmitted by default.

## Acceptance boundary

An LLM may submit `R1` evidence and may issue a valid `Q1`. It does not declare `COMPLETED`. Completion is an external acceptance transition.
