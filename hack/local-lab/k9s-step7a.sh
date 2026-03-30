#!/usr/bin/env bash

set -euo pipefail

source "$(cd "$(dirname "$0")" && pwd)/common.sh"

bin_path="${K9S_NEO_STEP7A_BASE_BIN:-${K9S_NEO_REPO_ROOT}/execs/k9s}"

if [[ ! -x "${bin_path}" ]]; then
  echo "missing executable K9s Neo binary: ${bin_path}" >&2
  echo "build it first with ./hack/with-go.sh make build" >&2
  exit 1
fi

exec "${bin_path}" \
  --perf-skip-crd-augment \
  --perf-skip-namespace-validation \
  "$@"
