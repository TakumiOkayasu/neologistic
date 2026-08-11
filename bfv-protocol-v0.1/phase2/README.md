# Phase 2 matched transfer pilot

This harness compares `direct-v1`, `nl-v1`, `pipe-v1`, and `json-v1` with one frozen model and reasoning configuration. `direct-v1` makes no producer call: the source bytes are the identity handoff to the same consumer used by every other arm. The other arms make one producer call and one consumer call.

The pilot is coverage-driven, not a reliability sample. Its five preregistered cases are split into one `copy` control and four `source-derived` cases. They cover R1 epistemic state and evidence/bounds association, Q1 false-positive and false-negative routing, authentic D1 forwarding, rejection of invented D1 authority, terminal-punctuation tolerance, and a downstream task oracle. One pass is 20 matched candidate cells and 35 initial Codex CLI invocations, plus one preflight smoke call. Candidate order rotates by fixture to reduce fixed order/cache bias. Arbitrary repetitions are not allowed.

## One run command

From this package directory:

```sh
./phase2/run-pilot.sh
```

Do not run that command merely to test the harness: it invokes the installed Codex runtime. The canonical corrected run is `artifacts/phase2/20260811T035522.867570000Z/`; its reviewed conclusion is no overall winner. The earlier `20260811T032926.172378000Z` run is retained as defect-discovery evidence.

The command first saves `codex --version`, `codex exec --help`, `codex login status`, `jq --version`, and a JSONL smoke invocation. It fixes `gpt-5.6-sol`, high reasoning, read-only sandboxing, ignored user/rule configuration, and no approvals through [`config.json`](config.json). It also records SHA-256 hashes of the config, sole fixture authority, all Phase 2 prompts, the existing pipe protocol instruction, and the jq adapter before any candidate cell runs. A failed preflight stops the pilot and retains its evidence.

The preregistered first pass fixes automatic runtime retries to zero, so every attempted call remains visible and no initial coverage result is replaced. An exit failure, malformed JSONL, missing completed agent message, or missing completed turn is classified as a runtime hard failure. A parse, semantic, routing, or downstream task failure is a completed trial and is never retried. Any later run is selected only as the report's single discriminating test and remains separate from first-pass coverage.

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

`response.txt` is the final `agent_message` text selected from JSONL without trimming, normalization, fence removal, or repair. Missing usage counters remain missing rather than becoming zero. `observations.jsonl`, `execution-order.json`, full scores, coverage, candidate summaries, selection, separated defect outputs, and a run-wide `SHA256SUMS.txt` are also saved.

After adding a reviewed score or report without changing the immutable model attempts, refresh the run-wide manifest with `go run ./cmd/phase2ctl checksum -root <run-directory>`.

Defect causality is separate from failure mechanism. Automatic attribution is limited to observed runtime failures; other hard failures remain `unclassified`. An optional evidence-backed JSONL passed to `phase2ctl score -defects` can place findings separately in `fixture`, `protocol`, `model`, `scorer`, or `runtime` output files.

## Scoring and decision rule

The scorer reports producer serialization, source-derived producer derivation, consumer receipt versus producer handoff, receipt versus hidden semantic gold, downstream task correctness, provider token categories, CLI call/retry counts, Q1 FP/FN, epistemic promotion, evidence/bounds association, and D1 transmission/authority by candidate. Copy cases never enter the producer-derivation denominator. For direct, consumption means source-to-receipt fidelity; for the other arms it means parsed producer-handoff-to-receipt fidelity.

Terminal sentence punctuation is ignored only for semantic prose comparison. Targets, evidence identifiers, authority references, JSON/pipe syntax, and the raw response remain exact.

Cost is a vector, not a synthetic scalar. Because provider `input_tokens` includes cached input, dominance uses mutually exclusive uncached input (`input - cached`) and cached input plus output, reasoning, calls, and retries. Instruction versus case-input tokens, internal provider request count, and monetary cost are explicitly unavailable because the event stream does not report them. A winner is emitted only for complete coverage when one no-hard-failure candidate weakly dominates every other candidate on correctness and all measured cost components, with at least one strict improvement. Otherwise `winner` is `null` and exactly one discriminating test is emitted from the observed decision uncertainty.

## Synthetic verification

These commands do not call a model:

```sh
env GOCACHE=/tmp/bfv-phase2-go-cache go test ./internal/phase2 ./cmd/phase2ctl
env GOCACHE=/tmp/bfv-phase2-go-cache go test ./...
env GOCACHE=/tmp/bfv-phase2-go-cache go vet ./...
go run ./cmd/phase2ctl validate
```

The unit suite explicitly tests Q1 false positive and false negative, epistemic promotion, evidence misassociation, D1 authority invention, punctuation-only tolerance, missing coverage cells, strict JSON kind keys/trailing values, literal-hyphen pipe escaping, raw JSONL preservation, and that ordinary semantic failures are not retried.
