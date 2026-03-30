#!/usr/bin/env bash

set -euo pipefail

source "$(cd "$(dirname "$0")" && pwd)/common.sh"

need_local_tool minikube
need_local_tool colima

delete_colima_profile="${K9S_NEO_DELETE_COLIMA_PROFILE:-1}"

"${K9S_NEO_TOOLS_BIN}/minikube" delete --profile="${K9S_NEO_PROFILE}" || true

if [[ "${delete_colima_profile}" == "1" ]]; then
  "${K9S_NEO_TOOLS_BIN}/colima" delete --profile="${K9S_NEO_COLIMA_PROFILE}" --force || true
fi
