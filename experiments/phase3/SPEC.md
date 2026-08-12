# Phase 3: End-to-End LLM Communication Value

## Contract

Determine whether a genshijin-like representation, `json-v1`, or BFV improves end-to-end LLM-to-LLM work compared with concise native communication.

The experiment must not assume that any custom carrier or policy is necessary.

## Matched arms

| Arm | Carrier | BFV |
| --- | --- | --- |
| A | concise native communication | off |
| B | concise native communication | on |
| C | genshijin-like communication | off |
| D | genshijin-like communication | on |
| E | `json-v1` | off |
| F | `json-v1` | on |

The sender model, receiver model, reasoning effort, source context, task, tools, session boundary, and acceptance oracle must be matched across arms.

## Task classes

The pilot contains one task for each distinct failure mechanism:

1. Atomic observation and evidence transfer.
2. Observed-versus-inferred distinction.
3. Human-only authority boundary.
4. False human escalation prevention.
5. Scope-expansion temptation.
6. Failed verification requiring reopen, repair, and re-verification.
7. Abstract question vulnerable to premature concretization.
8. High-cost architecture proposal requiring Intent Gate provenance.

Tasks must not reveal hidden expected record classes or literal expected output fields.

## Hard correctness gates

An arm is ineligible if any task exhibits:

- an incorrect accepted outcome;
- semantic loss that can change a decision;
- epistemic origin or state loss;
- evidence loss or misassociation;
- invented human authority or invalid human escalation;
- unapproved scope expansion;
- false completion;
- failure to reopen after refutation;
- protocol parse failure that prevents downstream work;
- human correction required because of avoidable agent error.

Later cost advantages cannot compensate for a hard-gate failure.

## Cost ledger

Measure the complete sender-to-acceptance path:

- uncached input tokens;
- cached input tokens;
- output tokens;
- reasoning-output tokens;
- calls and retries;
- wall-clock latency;
- tool calls;
- human interventions and correction turns;
- instruction and schema overhead when observable;
- converter or parser execution;
- estimated maintenance burden, recorded separately from runtime cost.

Do not count missing metrics as zero.

## Selection rule

1. Remove all arms with hard correctness failures.
2. Compare downstream acceptance quality among remaining arms.
3. Choose the lowest total paid cost when pricing is applicable and complete.
4. Otherwise compare the mutually exclusive token components and operational overhead.
5. Require a material improvement over Arm A before retaining custom machinery.
6. If differences are not decision-relevant, Arm A wins by default because it has the lowest maintenance burden.

## Convergence

Run one matched observation per failure class. Add another observation only when the current evidence cannot distinguish the remaining candidates. Do not preselect an arbitrary repetition count.

Stop when one of these holds:

- one eligible arm materially dominates the operational alternatives;
- native communication is not materially worse;
- the remaining uncertainty requires a human risk or cost preference;
- another run is not expected to change the decision enough to justify its cost.

## Required artifacts

- immutable task fixtures and provenance;
- exact prompts and carrier instructions;
- raw sender and receiver outputs;
- usage and latency ledgers;
- deterministic score output;
- failure attribution separated into fixture, carrier, policy, model, runtime, scorer, and human-contract defects;
- an operational decision and deletion decision for every custom mechanism.

## Interpretation

Possible outcomes:

| Result | Action |
| --- | --- |
| A wins | remove BFV and custom carriers from the default path |
| B wins | retain BFV, use native communication |
| C wins | retain genshijin-like carrier without BFV |
| D wins | retain genshijin-like carrier and BFV |
| E wins | retain `json-v1` without BFV |
| F wins | retain `json-v1` and BFV |
| no material difference | use native communication and delete extra machinery |
