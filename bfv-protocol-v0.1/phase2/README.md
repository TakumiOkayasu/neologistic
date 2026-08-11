# Phase 2 matched transfer pilot

This harness compares `direct-v1`, `nl-v1`, `pipe-v1`, and `json-v1` with one frozen model and reasoning configuration. `direct-v1` makes no producer call: the source bytes are the identity handoff to the same consumer used by every other arm. The other arms make one producer call and one consumer call.

The pilot is coverage-driven, not a reliability sample. Its five preregistered cases are split into one `copy` control and four `source-derived` cases. They cover R1 epistemic state and evidence/bounds association, Q1 false-positive and false-negative routing, authentic D1 forwarding, rejection of invented D1 authority, terminal-punctuation tolerance, and a downstream task oracle. One pass is 20 matched candidate cells and 35 initial Codex CLI invocations, plus one preflight smoke call. Candidate order rotates by fixture to reduce fixed order/cache bias. Arbitrary repetitions are not allowed.

## One run command

From this package directory:

```sh
./phase2/run-pilot.sh
```

Do not run that command merely to test the harness: it invokes the installed Codex runtime. The canonical corrected run is `artifacts/phase2/20260811T035522.867570000Z/`. Its frozen Pareto diagnostic is `no_winner`, while the ratified operational decision is `json-v1` as default, `pipe-v1` as challenger, and direct/NL rejected. That decision was calculated from the stored usage without a model rerun, and the canonical artifact remains byte-unchanged. The earlier `20260811T032926.172378000Z` run is retained as defect-discovery evidence.

The command first saves `codex --version`, `codex exec --help`, `codex login status`, `jq --version`, and a JSONL smoke invocation. It fixes `gpt-5.6-sol`, high reasoning, read-only sandboxing, ignored user/rule configuration, no approvals, and the versioned operational pricing policy through [`config.json`](config.json). It also records SHA-256 hashes of the config, sole fixture authority, all Phase 2 prompts, the existing pipe protocol instruction, and the jq adapter before any candidate cell runs. A failed preflight stops the pilot and retains its evidence.

The preregistered first pass fixes automatic runtime retries to zero, so every attempted call remains visible and no initial coverage result is replaced. An exit failure, malformed JSONL, missing completed agent message, or missing completed turn is classified as a runtime hard failure. A parse, semantic, routing, or downstream task failure is a completed trial and is never retried. Any later run must be justified by a changed re-evaluation condition and remains separate from the canonical coverage evidence.

## Artifacts

Each run is written below `artifacts/phase2/<UTC-run-id>/`. Every model attempt lives at:

```text
cells/<cohort>/<case>/<candidate>/<stage>/attempt-NNN/
  prompt.txt
  response.txt
  stdout.jsonl
  stderr.txt
  exit-status.txt
  usage.json
  attempt.json
```

The direct identity producer stage is recorded as `not_applicable` with zero attempts and stores the unchanged bytes in `producer/raw-source.txt` and the cell-level `handoff.txt`. Its producer serialization and producer derivation are not applicable; its source-to-consumer receipt fidelity is still a downstream-consumption measure.

`response.txt` is the final `agent_message` text selected from JSONL without trimming, normalization, fence removal, or repair. Missing usage counters remain missing rather than becoming zero. `observations.jsonl`, `execution-order.json`, full scores, coverage, candidate summaries, Pareto and operational selections, separated defect outputs, and a run-wide `SHA256SUMS.txt` are also saved.

For future generated reports, `selection.json`, `score.json.selection`, and CLI `winner` retain their legacy Pareto meaning. `pareto-selection.json`, `score.json.pareto_selection`, and the CLI Pareto fields expose that diagnostic explicitly. `operational-selection.json`, `score.json.operational_selection`, and the CLI operational fields expose the default decision separately. The frozen canonical artifact predates those explicit files and is not rewritten. Its legacy `next_discriminating_test` is historical diagnostic data, not an active instruction to rerun the model.

After adding a reviewed score or report without changing the immutable model attempts, refresh the run-wide manifest with `go run ./cmd/phase2ctl checksum -root <run-directory>`.

Defect causality is separate from failure mechanism. Automatic attribution is limited to observed runtime failures; other hard failures remain `unclassified`. An optional evidence-backed JSONL passed to `phase2ctl score -defects` can place findings separately in `fixture`, `protocol`, `model`, `scorer`, or `runtime` output files.

## Scoring and decision rules

The scorer reports producer serialization, source-derived producer derivation, consumer receipt versus producer handoff, receipt versus hidden semantic gold, downstream task correctness, provider token categories, CLI call/retry counts, Q1 FP/FN, epistemic promotion, evidence/bounds association, and D1 transmission/authority by candidate. Copy cases never enter the producer-derivation denominator. For direct, consumption means source-to-receipt fidelity; for the other arms it means parsed producer-handoff-to-receipt fidelity.

Terminal sentence punctuation is ignored only for semantic prose comparison. Targets, evidence identifiers, authority references, JSON/pipe syntax, and the raw response remain exact.

The Pareto diagnostic keeps cost as a vector. Because provider `input_tokens` includes cached and cache-write input, it uses mutually exclusive uncached input (`input - cached - cache write`), cached input, cache-write input, output, reasoning, calls, and retries. A Pareto winner is emitted only for complete coverage when one no-hard-failure candidate weakly dominates every other candidate on correctness and every measured cost component, with at least one strict improvement. Otherwise its winner is `null`. Legacy selection outputs remain aliases for this diagnostic.

Operational selection is separate and applies gates in this exact order: complete coverage, zero hard failures, semantic transfer and downstream task success, paid-token cost, then total `input_tokens + output_tokens` only if dollar cost ties exactly. The pinned standard rates are $5.00/M uncached input, $0.50/M cached input, $6.25/M cache-write input, and $30.00/M output, with `U = I - C - W`. Reasoning output is already included in output and is not added twice. The rate policy applies only at or below 272,000 input tokens per call; missing, invalid, overflowing, or over-limit usage blocks operational selection.

On the canonical pass, the maximum call was 15,753 input tokens. JSON cost $0.302306 and pipe cost $0.402369. JSON therefore wins operationally by $0.100063, approximately 24.9%, and used 1,699 fewer input-plus-output tokens. Direct and NL are rejected at the hard-failure gate. See [`docs/PHASE2_DECISION.md`](../docs/PHASE2_DECISION.md).

## Synthetic verification

These commands do not call a model:

```sh
env GOCACHE=/tmp/bfv-phase2-go-cache go test ./internal/phase2 ./cmd/phase2ctl
env GOCACHE=/tmp/bfv-phase2-go-cache go test ./...
env GOCACHE=/tmp/bfv-phase2-go-cache go vet ./...
go run ./cmd/phase2ctl validate
```

The unit suite explicitly tests Q1 false positive and false negative, epistemic promotion, evidence misassociation, D1 authority invention, punctuation-only tolerance, missing coverage cells, strict JSON kind keys/trailing values, literal-hyphen pipe escaping, raw JSONL preservation, and that ordinary semantic failures are not retried.
