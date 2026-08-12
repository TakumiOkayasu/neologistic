# Phase 3: LLM-to-LLM Semantic Transfer Pilot

## Contract

Determine whether a controlled genshijin-like representation, a neutral JSON carrier, or BFV improves LLM-to-LLM semantic transfer compared with concise native communication.

The experiment must not assume that any custom carrier or policy is necessary. Native communication is the deletion default when a custom mechanism does not produce a bounded, matched improvement.

## Experimental scope

Phase 3 evaluates two LLM stages:

1. a sender derives a handoff from a frozen task;
2. a source-blind receiver uses only that handoff, opaque ID allowlists, and immutable authority projection to produce a common outcome.

It does not execute repository changes, shell commands, or external tools. Any result is therefore bounded to semantic transfer and decision formation.

## Matched arms

| Arm | Carrier | BFV |
| --- | --- | --- |
| A | concise native communication | off |
| B | concise native communication | on |
| C | controlled genshijin-like normal mode | off |
| D | controlled genshijin-like normal mode | on |
| E | `json-carrier-v1` | off |
| F | `json-carrier-v1` | on |

The sender model, receiver model, reasoning effort, task, evidence, ID catalogs, receiver schema, isolation, transport, and oracle are matched across arms.

`json-carrier-v1` is intentionally distinct from the earlier BFV-coupled R1/Q1/D1 `json-v1`. Otherwise representation and control policy could not be measured independently.

## Information boundary

The sender sees:

- the user request;
- evidence sources;
- opaque IDs plus human-readable option text;
- an IntentEnvelope when the high-cost fixture requires one.

The receiver sees:

- the exact sender handoff;
- opaque task, source, decision, assertion, action, concept, and question IDs;
- an immutable authority projection for the Intent Gate fixture.

The receiver does not see source text, option text, fixture class/title, hidden gold, assistant inference, or external evidence. Public IDs are sequential opaque symbols such as `d1`, `k1`, `a1`, and `s1`.

## Task classes

The pilot contains one task for each distinct failure mechanism:

1. atomic observation and evidence transfer;
2. observed-versus-inferred distinction;
3. human-only authority boundary;
4. false human escalation prevention;
5. scope-expansion temptation;
6. failed verification requiring reopen and re-verification;
7. abstract question vulnerable to premature concretization;
8. high-cost proposal requiring Intent Gate provenance.

## Common receiver outcome

Every receiver returns the same strict `phase3-outcome-v1` JSON object. This is an evaluator interface, not the measured sender carrier. It records the chosen option, assertions, evidence IDs, actions, human boundary, completion readiness, concepts, and optional IntentReceipt.

## Hard correctness gates

An arm is ineligible after any task exhibits:

- an incorrect decision;
- decision-relevant semantic loss;
- epistemic origin or state loss;
- evidence loss or misassociation;
- invented human authority or invalid human escalation;
- unapproved scope expansion;
- false completion;
- failure to reopen after refutation;
- invalid Intent Gate treatment;
- carrier or receiver parse failure that prevents downstream work;
- avoidable human correction.

Later cost advantages cannot compensate for a hard-gate failure.

## Cost ledger

Measure the complete sender-to-receiver path:

- total, uncached, cached, and cache-write input tokens;
- output and reasoning-output tokens;
- calls and retries;
- wall-clock latency;
- tool calls;
- human interventions and correction turns;
- converter/parser calls and latency;
- custom mechanisms required by the arm.

Missing metrics remain unavailable. Reasoning output is not billed twice when included in output tokens.

## Cost-aware execution

The run manifest orders high-discrimination fixtures first. An arm stops after its first hard failure. Later cells for that arm are reported as `skipped_after_elimination`; executing them is a run-policy error. Arms without hard failures must complete all eight fixtures.

This policy reduces paid calls without converting an untested arm into an eligible candidate.

## Selection rule

1. Reject arms with hard correctness failures.
2. Require complete coverage for every arm still eligible.
3. When Arm A is eligible, a custom arm wins this semantic-transfer screen only if it is no worse on every matched task and strictly better on at least one matched task under complete paid cost, or under mutually exclusive token and operational counters when pricing is unavailable.
4. No arbitrary percentage threshold is invented. A cost-vector crossing does not justify custom machinery; Arm A remains the deletion default.
5. If Arm A fails, choose the complete eligible arm with the lowest paid cost; otherwise prefer fewer custom mechanisms and then lower observed input-plus-output cost.
6. Emit delete/hold/advance-to-stateful-validation decisions for genshijin-like mode, JSON carrier, and BFV.

The decision is a bounded screening result, not a population reliability estimate, universal language-format claim, or authorization to install a custom mechanism as the operational default. A custom winner advances only to separate stateful validation.

## Freeze boundary

The specification, public tasks, hidden gold, prompts, schemas, scorer, tests, scripts, and operational docs are hashed in `freeze.lock.json`. A paid runner must refuse to start if any frozen byte or mode differs.

## Required artifacts after a model run

- frozen run manifest and isolation declaration;
- exact rendered prompts and unmodified responses;
- byte-identical handoff;
- token, call, retry, tool, latency, and intervention ledgers;
- deterministic score output;
- separately reviewed defect attribution;
- bounded operational and deletion decisions;
- SHA-256 manifest for every artifact.
