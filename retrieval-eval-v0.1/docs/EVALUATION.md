# Evaluation contract

## Stages

```text
candidate discovery
  -> original-source verification
  -> downstream reasoning
```

This package scores the first stage and accepts an optional downstream result without conflating them.

## Fixture

Each fixture declares a query, answerability, and gold source paths. Gold construction must be independent of the candidate being evaluated.

## Observation

Each observation binds one system to one fixture and records availability, source revision, stale status, ranked candidates, absence claims, latency, candidate bytes, tool calls, and token counters.

## Selection

A candidate is eligible only when:

- required coverage is complete;
- no stale or unavailable observation is hidden;
- unsupported absence claims are zero;
- recall and downstream correctness meet the declared acceptance gate.

Cost chooses among eligible candidates. Missing monetary cost remains unavailable rather than zero.
