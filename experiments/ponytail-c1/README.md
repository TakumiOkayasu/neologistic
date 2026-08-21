# Ponytail C1 integration audit

Issue [#3](https://github.com/TakumiOkayasu/neologistic/issues/3) のC1を,対象hostに実在するPonytail installationから採取するためのboundedな監査packageである.Ponytailのrule engineやhookは再実装せず,installed artifactをhash化し,manifestが宣言する既存hookだけを隔離用directoryで呼び出す.

## Gate

- C1: **incomplete**.対象Macのinstallationとlive Codex app evidenceが未採取.
- C2: **not started**.C1完了前にfixture,prompt,scorer,cost ledgerをfreezeまたは実行しない.
- 既存の`bfv-protocol-v0.1/`,`experiments/phase3/`,`.github/workflows/phase3.yml`は変更しない.

静的source確認,isolated hook probe,current appでのlive observationを同一の証拠として扱わない.

| Evidence kind | 証明できること | 証明できないこと |
| --- | --- | --- |
| local installation | version,commitの有無,artifact hash,manifest,skill,hook | hostがloadまたは実行したこと |
| isolated hook probe | 対象artifactのhook code path,exit,output hash,state transition | current appでのactivation,subagentへのdelivery |
| live host observation | current appでのhook invocationとsubagent propagation | 別versionや別artifactの挙動 |
| upstream reference | source上の期待挙動とdrift比較 | local installationの同一性 |

## Collector

依存はNode stdlibだけで,network accessを行わない.`--plugin-root`の`.codex-plugin/plugin.json`からskills,hook map,3 eventのscriptを解決する.親directoryのGit repositoryをlocal plugin commitとして採用しない.

Static collectionがdefaultであり,installed scriptは実行しない.Isolated probeは利用者がstatic artifact SHA-256を確認し,同じhashを`--expect-artifact-sha256`へ明示した場合だけopt-inできる.Probeはroot `.git`を除くverified temporary snapshotから実行し,実行後hashが変われば全observationをinvalidとする.Script自体はsandboxではなくcollector user権限で動くため,未信頼artifactにはopt-inしない.子processのcwd,HOME,config,`PLUGIN_DATA`,environment,timeout,stdout/stderr sizeはboundedにする.

```sh
cd experiments/ponytail-c1

node --test collect-audit.test.mjs

node collect-audit.mjs \
  --subject local-installation \
  --plugin-root ~/path/to/installed/ponytail \
  --state-dir ~/path/to/codex/plugin-data \
  --config ~/.config/ponytail/config.json \
  --instruction repository=../../AGENTS.md \
  --instruction global=~/path/to/effective/global_AGENTS.md \
  --out ~/ponytail-c1-local-static.json

# Static reportのartifact SHA-256を人間が確認した後だけ実行する.
node collect-audit.mjs \
  --subject local-installation \
  --plugin-root ~/path/to/installed/ponytail \
  --probe-isolated \
  --expect-artifact-sha256 CONFIRMED_SHA256 \
  --out ~/ponytail-c1-local-probe.json
```

`--state-dir`,`--config`,`--instruction`は省略可能である.ただし未指定または採取不能の対象は`unknown`のままで,C1完了にはならない.Probe reportにも必要なoptional argsは繰り返す.`--out`は既存fileを上書きしない.Upstream snapshotを比較用に採取する場合は`--subject upstream-reference`とし,local primaryへ昇格させない.

Artifact hashはplugin root直下の`.git`だけを除くfileとsymlinkを`path,type,bytes,sha256`のcanonical JSONLへ変換し,UTF-8 path-byte順に並べたbytesをSHA-256でhash化する.上限は5,000 entries,1 file 32 MiB,total 256 MiBである.Symlinkはtarget文字列だけをhash化し,reportの`complete`を`false`にする.Socket/FIFO等のunsupported entryは黙って除外せずcollectorを停止する.

Reportはisolated probeのoverall pass/failを出さない.`current_codex_app_activation`,`live_subagent_propagation`,`current_codex_hook_runtime`は常にlive確認待ちとして残す.

## Live verification

[Codex Hooks公式仕様](https://learn.chatgpt.com/docs/hooks)では,pluginのinstall/enableだけでhookはtrustされず,exact definition hashごとにreviewが必要である.対象Codex app/CLIで次を観測し,host version,installed path,artifact SHA-256,hook definition hash,実行時刻を同じsnapshotへ結び付ける.

1. `codex plugin list --json`またはappのplugin browserでPonytailのversionとenabled stateを確認する.Installed pathはlist outputから推測せず,install時の`codex plugin add --json` outputまたはhost diagnosticsから別に取得し,collectorへ渡す.取得不能なら`unknown`を維持する.Install kind,marketplace/source,Codex app activation,CLI activationは別fieldとして記録する.
2. `/hooks`でmanifest由来の`SessionStart`,`SubagentStart`,`UserPromptSubmit`が各1 entryだけloadされ,現在のdefinitionがtrust済みであることを確認する.別sourceの同event hookも記録する.Raw hook-map file SHA-256とhost trust-definition hashは同一と仮定せず,別fieldにする.
3. 新規threadで`SessionStart(source=startup)`のexitと`<PLUGIN_DATA>/.ponytail-active`を確認する.`@ponytail ultra`後にcompactを発生させ,stateと再注入hashが維持またはresetされるかを記録する.`resume`と`clear`も同じ方法で確認する.
4. `@ponytail lite`,`@ponytail-review`,`stop ponytail`で`UserPromptSubmit`のexit,model-visible `additionalContext`,state transitionを分離して記録する.`systemMessage`はUI/event表示でありinstruction bytesへ算入しない.
5. Host execution/event traceで`SubagentStart`のinvocation,exit,subagent向け`additionalContext` hashを確認する.Traceがない場合,repository instructionを持たない一時workspaceでbehavioral canaryを補助観測してもよいが,それだけではpropagationをverifiedにしない.
6. 2 threadを並行に異なるmodeへ切り替え,fixed `<PLUGIN_DATA>/.ponytail-active`がsession間で干渉しないか確認する.
7. Bundled `ponytail` skillを明示invokeしたturnで,skill bodyと`SessionStart` additional contextの重複load/dedupeをhost traceから確認する.traceがなければ`unknown`を維持し,隠しinstruction本文の開示をagentへ要求しない.

Ponytail `off`はinstalled-off observationに留める.Main hook rulesetを止めてもplugin/skill discovery,hook trust,applicable Node invocation,state処理が残り,bundled skillのimplicit activationも未排除だからである.C2のzero-plugin baselineにはしない.

## Verification

```sh
node --test experiments/ponytail-c1/collect-audit.test.mjs
node experiments/ponytail-c1/collect-audit.mjs --help
git diff --check
git diff --exit-code d2fc63a -- \
  bfv-protocol-v0.1 \
  experiments/phase3 \
  .github/workflows/phase3.yml
```

Current partial resultとinstruction分類は[AUDIT.md](AUDIT.md)を参照する.
