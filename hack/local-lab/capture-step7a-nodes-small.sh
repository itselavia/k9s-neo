#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
source "${script_dir}/nodes-small-env.sh"

label="${K9S_NEO_BENCH_LABEL:-step7a-nodes-small-control-v1}"
note_path="${K9S_NEO_NODES_SMALL_NOTE_PATH:-${script_dir}/../../docs/development/step-7a-nodes-small-control-note.md}"

"${script_dir}/bench-nodes-small.sh" \
  --label "${label}" \
  --cold-runs "${K9S_NEO_COLD_RUNS:-10}" \
  --warm-runs "${K9S_NEO_WARM_RUNS:-10}" \
  --bin "${K9S_NEO_BIN:-${script_dir}/k9s-step7a.sh}" \
  --vars "${K9S_NEO_VARS_PATH:-${script_dir}/../bench/vars.nodes-small.json}" \
  --write-note "${note_path}" \
  "$@"
