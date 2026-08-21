# C1 audit record

## Verdict

**C1 incomplete / C2 not started**である.

2026-08-22(JST)時点の実行環境にはinstalled Ponytailと`codex` executableがなく,対象Macのlocal installation,current Codex app activation,live subagent propagationを確認できなかった.この環境から生成したsubagentにもPonytail instructionは観測されなかったが,pluginのinstall/load自体が未確認なのでnegative propagation evidenceにはしない.

以下のupstream source確認とisolated probeはdrift比較用であり,local/live evidenceへ昇格させない.

## Audit checklist

| Issue #3 C1 item | Current finding | Evidence | Status |
| --- | --- | --- | --- |
| version / commit / artifact hash | Target localは未採取.Upstream versionは`4.9.0`,sourceは`2ed6c52c9d7e5e56942508591085fd45dea277d3` | upstream reference | `unknown` |
| plugin manifest / skills / hooks | UpstreamはCodex manifest 1,skills 6,hook directory files 11,lifecycle event 3 | upstream reference | `unknown` |
| default / active mode / state | Source defaultは`full`.Codex stateは`<PLUGIN_DATA>/.ponytail-active` | upstream source | `unknown` |
| `SessionStart` | Isolated `startup`と`compact`はexit 0 | isolated probe | `unknown` |
| `SubagentStart` | Isolated contextは`SessionStart`と同一hash | isolated probe | `unknown` |
| `UserPromptSubmit` | Isolated `full -> ultra -> full(compact) -> off` transitionはexit 0 | isolated probe | `unknown` |
| current Codex app activation | Target hostを観測できず | none | `unknown` |
| live subagent propagation | Target hostを観測できず | none | `unknown` |
| emitted instruction bytes / SHA-256 | Upstream `full` isolated contextは5,252 bytes / `da4fb09cff2f6726691ce6591cebc38c95597d79da132e49c6fa2665c4e8a3ff` | isolated probe | `unknown` |
| Node / timeout / exit / fallback / disable | Upstream sourceとisolated probeのみ確認 | source + isolated probe | `unknown` |
| repository instruction overlap | Current repository rulesとの6分類を記録 | repository primary | partial |
| effective global instruction overlap | Effective installationを採取できず | none | `unknown` |

`Status`はlocal-primaryまたはlive evidenceが揃うまで`unknown`を維持する.

## Codex host contract

[Codex Hooks公式仕様](https://learn.chatgpt.com/docs/hooks)との照合結果である.

| Host contract | Audit consequence |
| --- | --- |
| 複数sourceのmatching hookは全てloadされ,same eventのcommandはconcurrentに動く.override/dedupeはない | `/hooks`でPonytail以外のsame-event hookも採取する |
| Pluginのinstall/enableだけではhookはtrustされず,exact definition hashが変わると再reviewまでskipされる | Versionだけでactivation済みと判断しない |
| `PLUGIN_ROOT`,`PLUGIN_DATA`とcompatibility用の`CLAUDE_PLUGIN_ROOT`,`CLAUDE_PLUGIN_DATA`が渡される | Ponytail command pathとstate pathを別々に確認する |
| `systemMessage`はUI/event warning,`hookSpecificOutput.additionalContext`はmodel-visible developer context | Instruction hashは`additionalContext`だけを対象にする |
| `SubagentStart` additional contextはsubagentへ入り,`SessionStart` contextはroot sessionへ入る | Parent markerをpropagation証拠にしない |
| Errorまたはtimeoutはhook failure,exit 0 + empty stdoutはsuccess | Node不在と`off`時のempty outputを区別する |
| `additionalContextLimit`省略時は約2,500 tokensでspillし,複数hook/pluginのcontextは加算される | Local live runでspillと重複を確認する |
| `/hooks`でindividual non-managed hookをdisableでき,`[features] hooks=false`で全hookを止められる | Hook absenceは実証できるが,bundled skill/plugin absenceとは別に確認する |

Codex pluginの削除command `codex plugin remove <plugin>`は[Developer commands公式仕様](https://learn.chatgpt.com/docs/developer-commands)に存在する.C1では現在のartifactとtrust stateを保った`off`も観測するが,negative controlとは断定しない.

## Upstream reference snapshot

Reference: [`DietrichGebert/ponytail@2ed6c52`](https://github.com/DietrichGebert/ponytail/commit/2ed6c52c9d7e5e56942508591085fd45dea277d3).Artifact fingerprintはignored local fileを含むworking checkoutでなく,同commitのclean `git archive` materializationから採取した.

Isolated hook値はstatic hash確認後に同じSHA-256を明示gateし,temporary snapshotのprobe前後hash一致を確認したrunから採取した.

| Field | Value |
| --- | --- |
| package / manifest version | `4.9.0` |
| commit | `2ed6c52c9d7e5e56942508591085fd45dea277d3` |
| `git describe` | `v4.9.0-3-g2ed6c52` |
| tree object | `93c5eaa97ec448b74055f6478898324b65ddb7f9` |
| `git archive` SHA-256 | `d5105dbac777431dc0041c6b8c2119df58c51737223794da825c41665368d866` |
| collector artifact SHA-256 | `083ebdf527a567044fdcdb03f75daf30c484e1e26e0dfb9e5a374d0c1fbe3fa2` |
| artifact count / bytes / complete | 159 / 1,636,232 / `true` |
| tag `v4.9.0` commit | `0a4dd63ad4541f4f655c4108a295916f3c1d8fda` |

Version `4.9.0`だけではsourceを一意に決められない.mainはtagから3 commits先である.さらにraw Codex hook-map file SHA-256はtagの`dd0837e870a8b81eb45ef4adebfc413a48c6daf84329befd897640f731aa0e39`からmainの`25f1af0c65db50fcc3b7c7e7535144a3ce2303d79c1a83feedf31bc4ce8673b1`へ変化している一方,manifest versionは同じである.Raw file SHA-256とCodex hostのtrust-definition hashは同一と仮定しない.Local commitまたはartifact hashに加え,hostのtrust statusと,表示される場合だけhost definition hashを別々に採取する.

### Manifest and hook map

`.codex-plugin/plugin.json`は1,280 bytes / `dfe092d6d9ead638c641ddf9f2daf44aded4aad788ad5be63fa2a66a6d3027dc`で,`./skills/`と`./hooks/claude-codex-hooks.json`を参照する.

| Event | Matcher | Script | Manifest timeout | Isolated exit |
| --- | --- | --- | ---: | ---: |
| `SessionStart` | `startup|resume|clear|compact` | `ponytail-activate.js` | 5 s | 0 |
| `SubagentStart` | omitted | `ponytail-subagent.js` | 5 s | 0 |
| `UserPromptSubmit` | omitted | `ponytail-mode-tracker.js` | 5 s | 0 |

Upstream READMEの「two tiny Node.js lifecycle hooks」およびCodex install時の「trust its two lifecycle hooks」は,manifestの3 event / 3 scriptと一致しない.分類は`conflict`であり,local hostがどのentryを表示・実行するかは`unknown`である.

## Activation,mode,state

- Default mode resolutionは`PONYTAIL_DEFAULT_MODE` -> config `defaultMode` -> `full`である.`review`はsession-only active modeでdefaultにはできない.
- Codex判定は`PLUGIN_DATA`の有無に依存し,stateはfixed filename `<PLUGIN_DATA>/.ponytail-active`へ保存される.
- Hook inputの`session_id`をstate keyに使わないため,同一`PLUGIN_DATA`を共有するconcurrent session isolationは`unknown`である.
- `SessionStart`は`startup`,`resume`,`clear`,`compact`の全matcherでdefaultをstateへ上書きする.

Isolated sequenceでは`SessionStart(full) -> @ponytail ultra -> SessionStart(source=compact)`が`full -> ultra -> full`となった.「modeはchangedまたはsession endまでpersistする」というPonytail instructionと一致せず,compactごとにfull contextも再注入された.分類は`conflict`である.Live hostで`resume`,`clear`,`compact`を再確認する.

State writeはbest-effortでerrorを握り潰すため,parent `SessionStart`がcontextを出してもstateが残らず,後続`SubagentStart`が無注入になる経路がある.Parent activation markerだけではpropagationを証明できない.

## Emitted context and overlap

`additionalContext`のUTF-8実測であり,stdout envelope hashやsource template hashとは分離する.

| Mode | Instruction bytes | Instruction SHA-256 |
| --- | ---: | --- |
| `lite` | 5,225 | `ea09a138c7aad46645e7ad1e60b4c552638314e689cdca1d27c2ba42fc2380eb` |
| `full` | 5,252 | `da4fb09cff2f6726691ce6591cebc38c95597d79da132e49c6fa2665c4e8a3ff` |
| `ultra` | 5,290 | `37d8be344d6e30c7ac8e1276fa0aff342344441d9ee5a70e3167d23277656d81` |
| `review` | 83 | `97ad5759253da70137afe32cfd34cef007446258909c74f6add6276a7ccf3ca6` |

`full`の`SessionStart` envelopeは5,479 bytes / `7173511e8827fc95d148252a6df3b2f365c17379ba24ef8e76a045b4a7602792`,`SubagentStart` envelopeは5,480 bytes / `9bbfa395fa3021aec7b0e4abb80df94f698bf3561deef17850e344224f9c639c`だった.両eventのinstruction hashは一致したが,これはscript直接実行の結果でありhost delivery evidenceではない.

Mode switch時の`UserPromptSubmit`は`ultra`で38 bytes,disableで17 bytesの`additionalContext`も返す.これは`systemMessage`と異なりmodel-visibleであるため,live cost ledgerへ算入する.

### Bundled skillとの重複

| Content | Bytes | Lines |
| --- | ---: | ---: |
| `skills/ponytail/SKILL.md` | 6,637 | - |
| Frontmatter除外skill body | 5,700 | 101 |
| `full` header除外hook body | 5,214 | 97 |

Terminal empty split elementを除くhook bodyの97 / 97 linesはskill bodyにverbatimかつ同じ順序で存在する.Manifestはbundled skillsを登録し,`SessionStart`は同じskill bodyのfiltered copyを注入する.Codex hostでskill bodyがloadされたturnではほぼ全rulesetが重複する可能性がある.公式hook contractにcross-source dedupeはないが,skillのload timingとhost側のcontext処理はlive未確認なので,実効重複は`unknown`とする.

## Runtime,fallback,disable

- Isolated runtime: Node `v24.19.0`,Linux `x64`.対象MacのNode/runtimeは`unknown`.
- PackageにNode `engines` constraintはない.Manifest commandは3件ともbare `node`,host timeoutは5 s.
- Matcher指定時の`SubagentStart`と`UserPromptSubmit`はstdin未closeに1,000 msのinternal fallbackを持つ.Open-stdin probeは1,052 ms,exit 0,full instruction hash一致だった.
- Invalid regex,malformed payload,missing `agent_type`,stdin error/timeoutはfail-openでsubagentへ注入する.Definite matcher mismatchだけが抑止する.
- Parse/I/O/EPIPE errorの多くは空stdout / exit 0として扱われ,durable hook logはない.
- Bare `node`をPATHから除いたmanifest command probeはexit 127,stdout 0 bytes,stderr 28 bytes / `771957a20dfdf9ef8cf7850a6bfadc884a40d331f5a314446a389bfa32a54a3f`だった.READMEの「always-on activation just stays quiet instead of erroring」と,公式host contractのhook failure表示は一致しない.分類は`conflict`.
- Disable pathはsession command `/ponytail off`,`stop ponytail`,`normal mode`,persistent default `off`,individual hook disable,all-hooks disable,またはplugin removeである.各方法のscopeを同一視しない.

## Instruction classification

Repository primary sources:

| Source | Bytes | SHA-256 |
| --- | ---: | --- |
| `AGENTS.md` | 1,186 | `f167374152605e2f6ddd036b75da6dc13a10cf3512ae4ef3ceaa9f9edb521ef1` |
| `PROJECT_ORIGIN.md` | 2,084 | `d252fa589d893821fd0ba0b5dbbabcf191a2be75dc516f668ccf216c15afc6fc` |
| `DECISION_LINEAGE.md` | 4,102 | `7f6697829657cb245cfe135000e96902e39fff1e8d2b1960511262f369e121b7` |

未deployまたはuncommittedな別worktree候補はeffective global evidenceではないため,hashと内容をreportへ採用しない.Target Macからlocal-effective instructionを採取するまでglobal比較は`unknown`を維持する.Issue #3で定義された6分類だけを使用する.

| Atomic topic | Classification | Evidence and consequence |
| --- | --- | --- |
| Filtered hook bodyとbundled skill body | `duplicate` | Staticにはhook 97 / 97 linesがskill bodyに同順で存在する.Live二重loadは別途`unknown` |
| Existing assetを再利用する | `compatible` | Repositoryのartifact preservationと同時適用できるが,同義ではない |
| Minimum correct implementationを選ぶ | `compatible` | PC1の実装最小化と整合し,CompanionでPonytail機能を再実装しない |
| explicit requestを理由なく削らない | `compatible` | Issue #3のC1/C2 gate,fixture,report要求を維持できる |
| one runnable check,no fixture unless asked | `compatible` | 本件ではC2 fixture/scorerが明示され,repository-defined checkも存在する.一般taskでのprecedenceはlive再確認する |
| session modeがcompact後もpersistする | `conflict` | Ponytail instructionのsession persistenceとhook実装のdefault resetが一致しない |
| Node不在時はquietにoffとなる | `conflict` | README説明とmanifest command/host failure contractが一致しない |
| known ceilingへ`ponytail:` commentを残す | `Ponytail-only` | Current repository ruleには同じcomment contractがない |
| intensity ladder,mode switch,caller scan | `Ponytail-only` | Ponytail固有のruntime/engineering contract |
| authority provenance,axis separation,artifact freeze,no custom wire | `existing-rule-only` | Ponytailの一般ruleで代替せず,repository guardrailを優先する |
| effective global instruction全体とprecedence | `unknown` | Target Macからlive bytes/hashを未採取 |
| skill + hookの実効重複 | `unknown` | Static bodyは重複するが,live skill load/dedupe traceがない |

## C2 gate and pilot boundary

C1が未完了なのでfixture,prompt,acceptance test,scorer,cost ledgerはfreezeしておらず,runも行っていない.C1完了後は別branch / PRでIssue #3の固定条件を使用する.

- 2 feature tasks + 2 bugfix tasks.
- True baselineと`ponytail-full`の2 arms,各cell 1 run,serial,seeded order.
- Task prompt,model,reasoning,verbosity,sandbox,repository revision,acceptance testはbyte-identical.
- Deterministic test/static predicateだけをscoreし,LLM judgeを使わない.
- Missing metricは`null`,quality gate通過前にLOC/file数等で優劣を決めない.
- 追加repeatはdecision-changingな曖昧さと対象cellをrun前に記録した場合だけ行う.

Ponytail `off`はmain hook ruleset注入を止めても,plugin/skill discovery,hook trust,3 hook declarationのapplicable Node invocation,latency/failure,state clearが残る.さらにenabled bundled skillはcoding taskでimplicit activationし得るため,skill absenceを観測しない`off`をnegative controlとは断定しない.したがって:

| Arm/control | Definition |
| --- | --- |
| true zero-plugin baseline | Ponytail pluginがabsentまたはplugin,skills,hooksの全てがdisableされ,new sessionでhook invocation/context/skillがないことを実証.`[features] hooks=false`だけでは不足 |
| installed-off observation | Plugin installed/trusted,mode `off`.Baseline/negative controlとは呼ばない |
| `ponytail-full` | Exact artifact,enabled,trusted,new session,SessionStart hash/stateと実subagent receiptを実証 |

## Remaining C1 evidence

1. Target Macのinstalled rootから`--subject local-installation` collector JSONを採取し,install kindとmarketplace/sourceを別記する.
2. Local version,commit status,artifact hash,manifest/hook hashをupstream referenceと照合する.
3. Effective global instruction fileをhash化し,上表をlocal emitted instructionに対して再判定する.
4. Current Codex appとCLIを分離し,3 eventのload,trust status,表示されるhost definition hash,exit,timeout,additional contextを観測する.
5. `startup/resume/clear/compact`,concurrent session,skill explicit invokeを観測する.
6. Host traceで`SubagentStart`のinvocation,exit,subagent向けcontext hashを確認する.Traceなしのbehavioral canaryはrepository instructionを持たない一時workspaceで補助観測し,単独ではrequired `unknown`を解消しない.
7. Local/live resultでrequired `unknown`を解消してから,C1完了を判断する.

Protected objectsは変更していない.

| Object | Base object id |
| --- | --- |
| `bfv-protocol-v0.1/` | `5792a80abbd3a5d9033534463906b91d2c146615` |
| `experiments/phase3/` | `1ce951ef1252d8fc039d6b7e8b6c62fa8ec8fb92` |
| `.github/workflows/phase3.yml` | `6e32341ba558a02fcb0fe09f4599963d60a854d5` |
