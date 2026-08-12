# Carrier: json-carrier-v1

Return exactly one compact JSON object with no extra keys or surrounding prose:

```json
{"summary":"<short self-contained handoff>","facts":[{"key":"<every declared assertion ID exactly once>","origin":"observed|inferred|synthetic","state":"open|verified|refuted","evidence_ids":["<source id>"]}],"recommended_option":"<declared decision option id>","actions":[{"id":"<every declared action ID exactly once>","disposition":"include|exclude"}],"concepts":["<selected declared concept id>"],"human_boundary":null,"constraints":["<decision-relevant constraint>"],"intent_receipt":null}
```

When human input is genuinely required, `human_boundary` is:

```json
{"question_id":"<declared human-question option id>","recommendation_option":"<declared decision option id>","blocking_reason":"<short exact reason>"}
```

When an IntentEnvelope exists, `intent_receipt` is required and must match `intent-v1`. `added_assumptions` must be `[]`; unauthorized proposed effects belong in `planned_side_effects` and `conflicts`.

This is a neutral carrier schema for Phase 3. It is intentionally distinct from the earlier BFV-coupled R1/Q1/D1 protocol so carrier and control policy remain independently testable.
