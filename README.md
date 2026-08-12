# neologistic

LLM同士の会話と変換を実験し、受理可能な成果品質を保ちながら総 interaction cost を下げられるかを検証する。

## 原点

最初の問いは、[`genshijin`](https://github.com/interfacex-co-jp/genshijin)のような簡略化・変換手法が現代のLLMでも有効か、から。

$$
\operatorname{objective} = \frac{\text{accepted outcome value}}{\text{total interaction cost}}
$$

総 cost には model token だけでなく、 retry 、人間の訂正、手戻り、 converter/Skill の保守、誤伝達による損失を含める。

## 現在の成果物

- `bfv-protocol-v0.1/`: `BFV control policy` と `LLM` 間の carrier の実験 package
- `json-v1`: 現在の限定された `model/configuration/workload` での `operational baseline`
- `pipe-v1`: challenger
- `experiments/phase3/`: 元の問いへ戻した `end-to-end matched experiment` の運用仕様

これらは研究目的そのものではなく、比較対象。

## 境界

**この `repository` は `runtime` や `複数project` を集約する `monorepo` ではない。**

- `ChatGPT` と `Codex` の接続は `chatgpt-codex-bridge` が担当する。
- `Retrieval` は入力候補発見の独立した研究軸であり、通信方式の一部として自動採用しない。
- 新規 `repository`、`MCP`、`Plugin`、`runtime`、`architecture` は、元の研究質問からの追跡可能な根拠と人間の明示承認なしに追加しない。

最初に[PROJECT_ORIGIN.md](PROJECT_ORIGIN.md)と[DECISION_LINEAGE.md](DECISION_LINEAGE.md)を読み、次の実験は[EXPERIMENT_PLAN.md](EXPERIMENT_PLAN.md)に従う。

Phase 3の詳細仕様は[experiments/phase3/README.md](experiments/phase3/README.md)を参照。
