# Intent Gate for High-Cost Delegation

Apply this gate only when a task proposes one or more of:

- a new repository;
- changes across multiple repositories;
- a new runtime, dependency, MCP server, plugin, hook, or architecture;
- an irreversible operation or a change with material rollback cost.

## Required provenance classes

- `USER_AUTHORIZED`: explicitly requested or approved by the human.
- `ASSISTANT_INFERENCE`: derived by the sending agent and not yet human-authorized.
- `EXTERNAL_EVIDENCE`: source-grounded observation.
- `NON_GOAL`: explicitly excluded from the current contract.

## Rule

An `ASSISTANT_INFERENCE` must not be transmitted as `USER_AUTHORIZED`.

## Execution gate

Execution may begin only when:

```text
receipt.conflicts is empty
AND receipt.added_assumptions does not expand the contract
AND receipt.planned_side_effects is a subset of envelope.allowed_side_effects
AND every acceptance criterion traces to USER_AUTHORIZED intent
```

The gate allows one correction round. If the second receipt still conflicts, stop before implementation.
