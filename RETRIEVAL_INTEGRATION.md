# Retrieval integration

`retrieval-eval-v0.1/` evaluates candidate discovery independently from LLM communication-format evaluation.

- Runtime implementation: `TakumiOkayasu/agent-retrieval-gateway`
- Delegation integration: `TakumiOkayasu/chatgpt-codex-bridge`
- Evaluation and immutable artifacts: this repository

Source code remains in sibling repositories. Integration occurs through versioned schemas, MCP tools, and evaluation artifacts rather than a source merge.
