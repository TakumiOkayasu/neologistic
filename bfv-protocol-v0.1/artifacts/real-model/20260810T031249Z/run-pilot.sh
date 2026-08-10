#!/usr/bin/env bash
set -uo pipefail

artifact_root="artifacts/real-model/20260810T031249Z"
iteration="${1:?usage: run-pilot.sh ITERATION_NAME}"
cases_path="${2:-fixtures/cases.jsonl}"
run_dir="${artifact_root}/${iteration}"
mkdir -p "${run_dir}/prompts" "${run_dir}/raw-events" "${run_dir}/stderr"
: >"${run_dir}/statuses.tsv"

while IFS= read -r fixture; do
  case_id="$(jq -r '.id' <<<"${fixture}")"
  input="$(jq -r '.input' <<<"${fixture}")"
  prompt_path="${run_dir}/prompts/${case_id}.txt"
  events_path="${run_dir}/raw-events/${case_id}.jsonl"
  stderr_path="${run_dir}/stderr/${case_id}.log"

  {
    sed -n '1,$p' prompts/protocol-instruction.md
    printf '\nFixture id: %s\n\n%s\n' "${case_id}" "${input}"
  } >"${prompt_path}"

  codex exec \
    --ephemeral \
    --ignore-user-config \
    --ignore-rules \
    --model gpt-5.6-sol \
    -c 'model_reasoning_effort="high"' \
    -c 'approval_policy="never"' \
    --sandbox read-only \
    --json \
    --cd /Users/yamazaki/prog/murata-lab/neologistic/bfv-protocol-v0.1 \
    - <"${prompt_path}" >"${events_path}" 2>"${stderr_path}"
  status=$?
  printf '%s\t%s\n' "${case_id}" "${status}" >>"${run_dir}/statuses.tsv"
done < "${cases_path}"
