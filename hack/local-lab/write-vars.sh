#!/usr/bin/env bash

set -euo pipefail

source "$(cd "$(dirname "$0")" && pwd)/common.sh"

output_path="${1:-${K9S_NEO_REPO_ROOT}/hack/bench/vars.local.json}"

cat >"${output_path}" <<EOF
{
  "kubeconfig": "${K9S_NEO_KUBECONFIG}",
  "context": "${K9S_NEO_PROFILE}",
  "big_namespace": "${K9S_NEO_NAMESPACE}",
  "pod_filter_regex": "api|worker",
  "node_filter": "",
  "allow_all_namespaces": false,
  "terminal_cols": 220,
  "terminal_rows": 60
}
EOF

echo "wrote ${output_path}"
