# neologistic

`neologistic`は、[Ponytail](https://github.com/DietrichGebert/ponytail)のboundedなCompanionである。Ponytailがprimary workflowとuser-facing interactionを担い、このrepositoryはPonytailで不十分、未検証、または適用外となる領域だけを計測・監査・補完する。

## 責務境界

| Owner | Responsibility |
| --- | --- |
| Ponytail | primary workflowとuser-facing interaction、YAGNI / KISS、stdlib・native platform・既存dependencyの優先、不要な抽象化の抑制、最小diffの選択、over-engineering review |
| neologistic / Ponytail Companion | activation・instruction伝播・既存rulesとの重複 / 競合の監査、実taskのcost / 変更量計測、quality gate、drift検出、evidence reportを担うboundedなinterop / evaluation artifactとevidence-producing check、Ponytailで具体的な△ / ×が観測された領域だけの限定的な補完 |

初期targetに対する現行判断は**`no custom wire`**である。独自wire protocolをPonytailの代替として開発せず、Ponytailの機能を再実装するCompanionにもならない。

## 原点

最初の問いは、[`genshijin`](https://github.com/InterfaceX-co-jp/genshijin)のような簡略化・変換手法が現代のLLMでも有効か、から。

$$
\operatorname{objective} = \frac{\text{accepted outcome value}}{\text{total interaction cost}}
$$

総costにはmodel tokenだけでなく、retry、人間の訂正、手戻り、converter/Skillの保守、誤伝達による損失を含める。

## 保存する既存成果物

このpivotでは、既存のBFV package / artifactをhistorical evidenceとして、frozen Phase 3 harnessをexperimental assetとして保持する。いずれも変更・削除・成功結果への置換を行わない。

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

最初に[PROJECT_ORIGIN.md](PROJECT_ORIGIN.md)と[DECISION_LINEAGE.md](DECISION_LINEAGE.md)を読む。保存済みのPhase 3実験計画は[EXPERIMENT_PLAN.md](EXPERIMENT_PLAN.md)、実装とpreflightは[experiments/phase3/README.md](experiments/phase3/README.md)を参照。これらはhistorical / experimental evidenceであり、Companionのcurrent roadmapではない。
