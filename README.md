# neologistic

LLM同士の会話と変換を実験し、受理可能な成果品質を保ちながら総interaction costを下げられるかを検証するrepositoryです。

## 原点

最初の問いは、[`genshijin`](https://github.com/interfacex-co-jp/genshijin)のような簡略化・変換手法が現代のLLMでも有効か、です。

```text
accepted outcome value
----------------------
total interaction cost
```

総costにはmodel tokenだけでなく、retry、人間の訂正、手戻り、converter/Skillの保守、誤伝達による損失を含めます。

## 現在の成果物

- `bfv-protocol-v0.1/`: BFV control policyとLLM間carrierの実験package。
- `json-v1`: 現在の限定されたmodel/configuration/workloadでのoperational baseline。
- `pipe-v1`: challenger。

これらは研究目的そのものではなく、比較対象です。

## 境界

このrepositoryはruntimeや複数projectを集約するmonorepoではありません。

- ChatGPTとCodexの接続は`chatgpt-codex-bridge`が担当します。
- Retrievalは入力候補発見の独立した研究軸です。通信方式の一部として自動採用しません。
- 新規repository、MCP、Plugin、runtime、architectureは、元の研究質問からの追跡可能な根拠と人間の明示承認なしに追加しません。

最初に[PROJECT_ORIGIN.md](PROJECT_ORIGIN.md)と[DECISION_LINEAGE.md](DECISION_LINEAGE.md)を読み、次の実験は[EXPERIMENT_PLAN.md](EXPERIMENT_PLAN.md)に従います。
