#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
source "${script_dir}/nodes-small-env.sh"

label="${K9S_NEO_BENCH_LABEL:-step7a-nodes-small-control-smoke-v1}"

"${script_dir}/bench-nodes-small.sh" \
  --label "${label}" \
  --cold-runs 1 \
  --warm-runs 1 \
  --bin "${K9S_NEO_BIN:-${script_dir}/k9s-step7a.sh}" \
  --vars "${K9S_NEO_VARS_PATH:-${script_dir}/../bench/vars.nodes-small.json}" \
  "$@"
