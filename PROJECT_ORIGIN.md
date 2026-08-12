# Project origin and invariant objective

## Original question

> 現代のモデルにおいて、`genshijin`のようなツールは効果があるのか。

この問いから開始した。目的は独自protocol、BFV、JSON、Bridge、Retrievalの完成ではない。

## Invariant objective

LLM間workflowについて、次を最大化する。

$$
\operatorname{objective} = \frac{\text{accepted outcome value}}{\text{total interaction cost}}
$$

`accepted outcome value`には正確性、意思決定価値、証拠、反証耐性、再利用性を含む。

`total interaction cost`には以下を含む。

- input/output/reasoning token;
- latencyとtool call;
- retryと失敗run;
- 人間の認知負荷と訂正回数;
- 誤伝達による手戻り;
- converter、Skill、protocol、runtimeの実装・保守費用。

短い出力や厳密なserializationは、この目的への寄与が実測された場合だけ価値を持つ。

## Research axes

次の軸を混同しない。

| Axis | Question |
| --- | --- |
| Carrier | どの表現で意味を伝えるか |
| Control policy | scope、検証、収束をどう制御するか |
| Transport | どの実行環境・threadへ届けるか |
| Retrieval | 何を候補として発見するか |
| Acceptance | 誰が何を根拠に完了を確定するか |

一つの軸の結果から、別の軸の恒久実装を導かない。

## Current bounded conclusions

- `json-v1`は、実施済みの限定されたmatched comparisonにおけるoperational baselineである。
- BFVはcontrol policy候補であり、project全体の目的ではない。
- `chatgpt-codex-bridge`は実験と実作業のtransport/harnessである。
- Retrievalのglobal rankingは大規模corpusで有効になり得るが、手元のworkflowへ恒久導入する判断は未承認である。

## Authority rule

人間が明示した要求・pivot・裁定だけがscopeとside effectを承認できる。

Assistantの推論、外部論文、repository内の既存記述は、単独では人間の承認に昇格しない。
