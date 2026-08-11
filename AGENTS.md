# neologistic guardrails

Before changing this repository, read `PROJECT_ORIGIN.md` and `DECISION_LINEAGE.md`.

- Preserve the original research question: whether genshijin-like LLM communication improves accepted outcome value per total interaction cost.
- Treat carrier, control policy, transport, retrieval, and acceptance as separate axes.
- Do not redefine BFV, `json-v1`, the Bridge, or Retrieval as the project purpose.
- Distinguish `USER_AUTHORIZED`, `ASSISTANT_INFERENCE`, `EXTERNAL_EVIDENCE`, and withdrawn decisions.
- Do not create repositories, add runtimes/dependencies/MCP/plugins/hooks, change repository roles, or publish releases from an assistant inference alone.
- Retrieval integration `X1` is withdrawn. Do not restore it without a new human ruling and workload-specific evidence.
- Freeze deterministic fixtures and scorers before invoking paid model/Codex runs.
- Use repository-relative paths or `~/path/to/...` in human-facing instructions; never embed a user's absolute home path.
