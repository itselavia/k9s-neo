#!/usr/bin/env bash

set -euo pipefail

source "$(cd "$(dirname "$0")" && pwd)/common.sh"

label="${K9S_NEO_BENCH_LABEL:-step6-smoke-$(date +%Y%m%d-%H%M%S)}"

"${K9S_NEO_REPO_ROOT}/hack/local-lab/bench-required.sh" \
  --label "${label}" \
  --cold-runs 1 \
  --warm-runs 1 \
  "$@"
