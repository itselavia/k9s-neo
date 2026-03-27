#!/usr/bin/env bash

set -euo pipefail

source "$(cd "$(dirname "$0")" && pwd)/common.sh"

label="${K9S_NEO_BENCH_LABEL:-local-baseline-v1}"
note_path="${K9S_NEO_BASELINE_NOTE_PATH:-${K9S_NEO_REPO_ROOT}/docs/development/step-6-local-baseline-note.md}"

"${K9S_NEO_REPO_ROOT}/hack/local-lab/bench-required.sh" \
  --label "${label}" \
  --cold-runs "${K9S_NEO_COLD_RUNS:-10}" \
  --warm-runs "${K9S_NEO_WARM_RUNS:-10}" \
  --write-note "${note_path}" \
  "$@"
