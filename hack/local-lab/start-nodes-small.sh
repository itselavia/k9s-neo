#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
source "${script_dir}/nodes-small-env.sh"
source "${script_dir}/common.sh"

need_local_tool docker

if [[ "${K9S_NEO_ALLOW_SHARED_ACTIVE_PROFILES:-0}" != "1" ]]; then
  other_running_profiles="$(
    "${K9S_NEO_TOOLS_BIN}/docker" ps --format '{{.Names}}' |
      awk -v current="${K9S_NEO_PROFILE}" '
        $0 ~ /^k9s-neo/ && $0 != current && $0 !~ ("^" current "-") {print $0}
      '
  )"

  if [[ -n "${other_running_profiles}" ]]; then
    echo "nodes-small start blocked: stop other active minikube profiles on the shared Colima backend first" >&2
    printf '%s\n' "${other_running_profiles}" >&2
    echo "set K9S_NEO_ALLOW_SHARED_ACTIVE_PROFILES=1 only if you are deliberately testing overlap" >&2
    exit 1
  fi
fi

exec "${script_dir}/start-cluster.sh" "$@"
