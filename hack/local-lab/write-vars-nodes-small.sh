#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
source "${script_dir}/nodes-small-env.sh"
source "${script_dir}/common.sh"

need_local_tool kubectl

output_path="${1:-${script_dir}/../bench/vars.nodes-small.json}"

"${K9S_NEO_TOOLS_BIN}/kubectl" config use-context "${K9S_NEO_PROFILE}" >/dev/null

hot_node="$(
  "${K9S_NEO_TOOLS_BIN}/kubectl" get pods \
    -n "${K9S_NEO_NAMESPACE}" \
    -l bench.k9s.io/profile=nodes-small \
    -o jsonpath='{range .items[*]}{.spec.nodeName}{"\n"}{end}' |
    sort |
    uniq -c |
    sort -nr |
    awk 'NR==1 {print $2}'
)"

if [[ -z "${hot_node}" ]]; then
  echo "nodes-small vars failed: could not determine hot node" >&2
  exit 1
fi

cat >"${output_path}" <<EOF
{
  "kubeconfig": "${K9S_NEO_KUBECONFIG}",
  "context": "${K9S_NEO_PROFILE}",
  "big_namespace": "${K9S_NEO_NAMESPACE}",
  "pod_filter_regex": "api|worker",
  "node_filter": "${hot_node}",
  "allow_all_namespaces": false,
  "terminal_cols": 220,
  "terminal_rows": 60
}
EOF

echo "wrote ${output_path}"
