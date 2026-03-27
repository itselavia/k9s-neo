#!/usr/bin/env bash

set -euo pipefail

source "$(cd "$(dirname "$0")" && pwd)/common.sh"

need_local_tool minikube
need_local_tool colima

"${K9S_NEO_TOOLS_BIN}/minikube" delete --profile="${K9S_NEO_PROFILE}" || true
"${K9S_NEO_TOOLS_BIN}/colima" delete --profile="${K9S_NEO_PROFILE}" --force || true
