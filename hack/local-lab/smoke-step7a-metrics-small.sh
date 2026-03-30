#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
source "${script_dir}/metrics-small-env.sh"

export K9S_NEO_BIN="${K9S_NEO_BIN:-${script_dir}/k9s-step7a.sh}"
export K9S_NEO_VARS_PATH="${K9S_NEO_VARS_PATH:-${script_dir}/../bench/vars.metrics-small.json}"
export K9S_NEO_BENCH_LABEL="${K9S_NEO_BENCH_LABEL:-step7a-metrics-small-control-smoke-v1}"

exec "${script_dir}/smoke-required.sh" "$@"
