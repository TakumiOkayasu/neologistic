# Scoring

## Hard gates

Each receiver outcome is compared with hidden gold. The scorer records runtime, carrier parse, decision, semantic, epistemic, evidence, authority, scope, completion, reopen, Intent Gate, and avoidable human-correction failures. Any such failure makes that arm ineligible.

## Coverage and early elimination

The full matrix is 8 fixtures x 6 arms = 48 cells. An arm without a hard failure must produce all eight observations. A missing cell is accepted only when:

1. the same arm has an earlier executed hard-failure cell in manifest order; and
2. the missing cell occurs after that failure.

Such cells appear in `skipped_after_elimination`. Missing pre-failure cells, duplicate cells, unexpected cells, out-of-order observations, or execution after elimination block selection.

## Cost comparison

When a verified pricing policy and all billed counters are present, the scorer prices the sender and receiver separately against exact `provider/model` entries, then sums their nano-USD costs. A policy that omits either model, exceeds an entry's applicability limit, or cannot decompose input into uncached, cached, and cache-write counters is invalid. Otherwise the scorer compares mutually exclusive uncached input, cached input, cache-write input, and output counters. Reasoning-output tokens remain diagnostic and are not added again when included in output.

Calls, retries, tool calls, human interventions, correction turns, and converter calls must not worsen when a custom arm claims to dominate Arm A. One-observation latency remains diagnostic because isolated latency is noisy.

## Deletion default

If Arm A passes every hard gate, a custom arm wins the semantic-transfer screen only by being no worse on every matched task and strictly better on at least one. No arbitrary percentage threshold is invented. A cost-vector crossing is insufficient. A custom winner is marked `advance_to_stateful_validation`, not retained as an operational default, because maintenance and stateful execution costs are not established here.

## Defect attribution

Hard-failure mechanism and causal attribution are separate. The scorer reports failures from raw outputs. A later review may attribute them to fixture, carrier, policy, model, runtime, scorer, human Contract, or `unclassified`. Attribution never alters raw bytes or silently removes a hard failure.

## Decision scope

The score selects a bounded semantic-transfer winner and deletion candidates for the next experimental stage. It does not establish a population failure rate or authorize production-wide adoption without stateful validation.
