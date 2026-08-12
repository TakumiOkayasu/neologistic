# Task Design Rules

- Each fixture isolates one failure mechanism.
- Human-authorized intent is immutable and separately represented.
- Assistant inferences are explicit and may be wrong.
- The sender sees source evidence and opaque IDs paired with option text.
- The receiver sees no source or option text; it receives only opaque ID allowlists, the exact handoff, and immutable authority projection when required.
- Fixture IDs and all public catalog IDs are opaque sequential symbols, not semantic hints.
- The common receiver outcome schema is identical across arms.
- The oracle evaluates decision, epistemic state, evidence association, actions, scope, authority, readiness, and IntentReceipt rather than serialization alone.
- Harmless punctuation and wording differences are not hard failures.
- Structurally valid but semantically incorrect output is a hard failure.
- Rejected work is not rewarded for being thorough.
- No fixture depends on a personal absolute path, one operating system's path alias, external network availability, or mutable third-party state.
- High-cost fixtures include an IntentEnvelope whose allowed effects trace directly to `USER_AUTHORIZED` IDs.
- A fixture class/title and hidden gold never enter sender or receiver prompts.
