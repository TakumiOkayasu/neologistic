# AGENTS.md

- Evaluate retrieval candidate discovery separately from answer generation.
- Preserve raw observations and immutable run artifacts.
- Do not convert unavailable indexes into observed answer failures.
- Treat search hits as candidates, not verified evidence.
- Correctness and gold-evidence recall are gates before cost selection.
- Do not invent trial counts; define coverage and uncertainty from the evaluation contract.
