# Task Design Rules

- Each fixture isolates one failure mechanism.
- The human-authorized intent is immutable and separately stored.
- Assistant inferences are explicit and may be wrong.
- The receiver must not be shown hidden gold labels.
- The acceptance oracle evaluates the resulting decision or action, not merely serialization.
- Surface punctuation and harmless wording differences are not hard failures.
- A structurally valid response with an incorrect meaning is a hard failure.
- Rejected work is not rewarded merely for being thorough.
