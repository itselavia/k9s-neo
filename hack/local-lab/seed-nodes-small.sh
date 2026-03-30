#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
source "${script_dir}/nodes-small-env.sh"
source "${script_dir}/common.sh"

need_local_tool kubectl

"${script_dir}/seed-bench.sh" "$@"

control_plane_nodes="$(
  "${K9S_NEO_TOOLS_BIN}/kubectl" get nodes \
    -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.labels.node-role\.kubernetes\.io/control-plane}{"\t"}{.metadata.labels.node-role\.kubernetes\.io/master}{"\n"}{end}' |
    awk '$2 != "" || $3 != "" {print $1}'
)"

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

hot_count="$(
  "${K9S_NEO_TOOLS_BIN}/kubectl" get pods \
    -n "${K9S_NEO_NAMESPACE}" \
    -l bench.k9s.io/profile=nodes-small \
    -o jsonpath='{range .items[*]}{.spec.nodeName}{"\n"}{end}' |
    sort |
    uniq -c |
    sort -nr |
    awk 'NR==1 {print $1}'
)"

unique_nodes="$(
  "${K9S_NEO_TOOLS_BIN}/kubectl" get pods \
    -n "${K9S_NEO_NAMESPACE}" \
    -l bench.k9s.io/profile=nodes-small \
    -o jsonpath='{range .items[*]}{.spec.nodeName}{"\n"}{end}' |
    sort -u |
    wc -l |
    tr -d ' '
)"

if [[ -z "${hot_node}" ]]; then
  echo "nodes-small seed failed: no benchmark pods found" >&2
  exit 1
fi

if [[ "${hot_count}" != "14" ]]; then
  echo "nodes-small seed failed: expected 14 benchmark pods on the hot node, got ${hot_count}" >&2
  exit 1
fi

if [[ "${unique_nodes}" != "1" ]]; then
  echo "nodes-small seed failed: expected all benchmark pods on one node, got ${unique_nodes} nodes" >&2
  exit 1
fi

if printf '%s\n' "${control_plane_nodes}" | grep -Fx "${hot_node}" >/dev/null 2>&1; then
  echo "nodes-small seed failed: hot node ${hot_node} is a control-plane node" >&2
  exit 1
fi

echo
echo "nodes-small hot node: ${hot_node}"
echo "nodes-small benchmark pod count on hot node: ${hot_count}"
"${K9S_NEO_TOOLS_BIN}/kubectl" get pods \
  -n "${K9S_NEO_NAMESPACE}" \
  -l bench.k9s.io/profile=nodes-small \
  -o wide \
  --sort-by=.spec.nodeName
echo
"${K9S_NEO_TOOLS_BIN}/kubectl" get pods \
  -n "${K9S_NEO_NAMESPACE}" \
  -l bench.k9s.io/profile=nodes-small \
  -o jsonpath='{range .items[*]}{.spec.nodeName}{"\n"}{end}' |
  sort |
  uniq -c
