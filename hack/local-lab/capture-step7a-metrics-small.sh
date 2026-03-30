#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
source "${script_dir}/metrics-small-env.sh"

export K9S_NEO_BIN="${K9S_NEO_BIN:-${script_dir}/k9s-step7a.sh}"
export K9S_NEO_VARS_PATH="${K9S_NEO_VARS_PATH:-${script_dir}/../bench/vars.metrics-small.json}"
export K9S_NEO_BENCH_LABEL="${K9S_NEO_BENCH_LABEL:-step7a-metrics-small-control-v1}"
export K9S_NEO_COLD_RUNS="${K9S_NEO_COLD_RUNS:-10}"
export K9S_NEO_WARM_RUNS="${K9S_NEO_WARM_RUNS:-10}"

run_output="$("${script_dir}/bench-required.sh" "$@")"

printf '%s\n' "${run_output}"

artifact_root="$(
  printf '%s\n' "${run_output}" | python3 -c 'import json, sys; print(json.load(sys.stdin)["artifact_root"])'
)"

"${script_dir}/write-step7a-metrics-small-note.sh" "${artifact_root}"
