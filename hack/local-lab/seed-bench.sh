#!/usr/bin/env bash

set -euo pipefail

source "$(cd "$(dirname "$0")" && pwd)/common.sh"

need_local_tool kubectl

"${K9S_NEO_TOOLS_BIN}/kubectl" config use-context "${K9S_NEO_PROFILE}" >/dev/null
"${K9S_NEO_TOOLS_BIN}/kubectl" apply -f "${K9S_NEO_MANIFEST}"

"${K9S_NEO_TOOLS_BIN}/kubectl" rollout status -n "${K9S_NEO_NAMESPACE}" deployment/bench-api --timeout=180s
"${K9S_NEO_TOOLS_BIN}/kubectl" rollout status -n "${K9S_NEO_NAMESPACE}" deployment/bench-worker --timeout=180s
"${K9S_NEO_TOOLS_BIN}/kubectl" rollout status -n "${K9S_NEO_NAMESPACE}" deployment/bench-indexer --timeout=180s
"${K9S_NEO_TOOLS_BIN}/kubectl" get pods -n "${K9S_NEO_NAMESPACE}" -o wide
