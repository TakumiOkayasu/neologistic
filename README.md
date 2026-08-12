# neologistic

LLM同士の会話と変換を実験し、受理可能な成果品質を保ちながら総interaction costを下げられるかを検証する。

## 原点

最初の問いは、[`genshijin`](https://github.com/InterfaceX-co-jp/genshijin)のような簡略化・変換手法が現代のLLMでも有効か、から。

$$
\operatorname{objective} = \frac{\text{accepted outcome value}}{\text{total interaction cost}}
$$

総costにはmodel tokenだけでなく、retry、人間の訂正、手戻り、converter/Skillの保守、誤伝達による損失を含める。

## 現在の成果物

- `bfv-protocol-v0.1/`: BFV control policyとLLM間carrierのPhase 1/2実験package
- BFV `json-v1`: Phase 2の限定されたmodel/configuration/workloadにおけるoperational baseline
- `pipe-v1`: Phase 2 challenger
- `experiments/phase3/`: native、genshijin-like、neutral JSONとBFV on/offを直交比較するdeterministic harness

Phase 3は8 fixture × 6 armの48-cell matched runを対象にする。現時点ではfixture、hidden gold、prompt renderer、Intent Gate、scorer、cost ledger、freeze manifestまで実装済み。paid model/Codex runは未実施。

## 境界

**このrepositoryはruntimeや複数projectを集約するmonorepoではない。**

- ChatGPTとCodexの接続は`chatgpt-codex-bridge`が担当する。
- Retrievalは入力候補発見の独立した研究軸であり、通信方式の一部として自動採用しない。
- 新規repository、MCP、Plugin、runtime、architectureは、元の研究質問からの追跡可能な根拠と人間の明示承認なしに追加しない。
- 図はMermaidを使用し、ASCII artは使用しない。

最初に[PROJECT_ORIGIN.md](PROJECT_ORIGIN.md)と[DECISION_LINEAGE.md](DECISION_LINEAGE.md)を読み、次の実験は[EXPERIMENT_PLAN.md](EXPERIMENT_PLAN.md)に従う。

Phase 3の実装とpreflightは[experiments/phase3/README.md](experiments/phase3/README.md)を参照。
