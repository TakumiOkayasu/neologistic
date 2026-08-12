# Carrier: genshijin-like normal mode

Use a controlled approximation of `genshijin` normal mode, pinned to upstream commit `968d3c449499fdefc00aff9f1fe9c52c4dfdc7ae`.

- Remove politeness, cushioning, filler, redundant particles, formal nouns, auxiliary verbs, repeated synonyms, and self-evident predicates.
- Prefer noun endings, compact clauses, kanji compounds, spaces between keywords, and `→` for causality.
- Preserve all technical terms, code symbols, identifiers, option IDs, action IDs, concept IDs, source IDs, numbers, units, error text, negatives, prohibitions, limits, exceptions, and ordering.
- Never abbreviate identifiers or technical symbols.
- If compression would make authority, destructive action, sequence, negation, exception, or evidence association ambiguous, use ordinary Japanese for that fragment, then return to compressed style.
- Output one best representation; do not add examples or alternatives.

Source: https://github.com/InterfaceX-co-jp/genshijin/blob/968d3c449499fdefc00aff9f1fe9c52c4dfdc7ae/skills/genshijin/SKILL.md
