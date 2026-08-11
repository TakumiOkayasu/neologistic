# BFV LLM wire protocol v0.1

## Purpose

Carry only decision-relevant conclusions, human-only blockers, and authoritative human decisions between LLM-facing components.

## Serializations

The semantic record model below has two supported serializations. `json-v1` is the operational default for new integrations. `pipe-v1` remains supported for compatibility, and the CLI intentionally continues to select it when `-format` is omitted. This compatibility default does not override the Phase 2 product decision; new callers select `-format json-v1` explicitly.

### `json-v1`

A message is exactly one JSON object whose only key is `records`. `records` is a non-empty array. Each element must contain exactly the keys for its kind:

| Kind | Exact keys |
|---|---|
| `R1` | `kind`, `target`, `epistemic`, `content`, `evidence`, `bounds` |
| `Q1` | `kind`, `target`, `content`, `recommendation`, `evidence`, `bounds` |
| `D1` | `kind`, `target`, `content`, `authority_reference`, `bounds` |

All fields are non-empty JSON strings except that `evidence` and `bounds` may be JSON `null` where the semantic rules permit absence. A `Q1` always requires a non-empty string `bounds` value for its exact blocking boundary. An `R1` with verified or refuted state always requires a non-empty string `evidence` value. A `D1` always requires a non-empty string `authority_reference` value.

```json
{"records":[{"kind":"R1","target":"build","epistemic":"OV","content":"tests passed","evidence":"run-42","bounds":"go test ./..."}]}
```

Unknown, omitted, duplicated, or kind-inappropriate keys are invalid. An empty `records` array, more than one top-level JSON value, invalid record semantics, and prose or Markdown outside the object are also invalid. Canonical encoding is one compact JSON value followed by the CLI's output newline.

### `pipe-v1` grammar

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

## `pipe-v1` null and escaping

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

## CLI validation and canonicalization

```sh
go run ./cmd/bfvctl validate -format json-v1 -input message.json
go run ./cmd/bfvctl canonicalize -format json-v1 -input message.json
```

Both commands accept `pipe-v1` or `json-v1`. Omitting `-format` preserves the historical `pipe-v1` behavior. `canonicalize` emits the selected input format and does not convert a message to the other serialization.

`validate` checks serialization and the local record invariants described above. Its `valid records=N` output does not authenticate a `D1` authority reference. Before acting on a `D1`, the transport or downstream consumer must resolve that reference against an authenticated authority source or explicit allowlist and reject unresolved references.

## Excluded from the wire

Task completion, sender, receiver, run ID, contract version, source revision, and timestamp belong to transport/enforcement metadata when available.

Rejected candidate work is not transmitted by default.

## Acceptance boundary

An LLM may submit `R1` evidence and may issue a valid `Q1`. It does not declare `COMPLETED`. Completion is an external acceptance transition.
