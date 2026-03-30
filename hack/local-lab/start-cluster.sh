#!/usr/bin/env bash

set -euo pipefail

source "$(cd "$(dirname "$0")" && pwd)/common.sh"

need_local_tool kubectl
need_local_tool minikube
need_local_tool colima
need_local_tool docker

driver="${K9S_NEO_MINIKUBE_DRIVER:-docker}"
minikube_cpus="${K9S_NEO_MINIKUBE_CPUS:-2}"
minikube_memory_mb="${K9S_NEO_MINIKUBE_MEMORY_MB:-2048}"
minikube_disk_size="${K9S_NEO_MINIKUBE_DISK_SIZE:-20g}"
minikube_nodes="${K9S_NEO_MINIKUBE_NODES:-1}"
container_runtime="${K9S_NEO_MINIKUBE_CONTAINER_RUNTIME:-containerd}"
colima_cpus="${K9S_NEO_COLIMA_CPUS:-2}"
colima_memory_gb="${K9S_NEO_COLIMA_MEMORY_GB:-3}"
colima_disk_gb="${K9S_NEO_COLIMA_DISK_GB:-20}"
enable_metrics_server="${K9S_NEO_ENABLE_METRICS_SERVER:-0}"
metrics_timeout="${K9S_NEO_METRICS_READY_TIMEOUT:-180s}"

wait_for_metrics_server() {
  "${K9S_NEO_TOOLS_BIN}/kubectl" config use-context "${K9S_NEO_PROFILE}" >/dev/null
  "${K9S_NEO_TOOLS_BIN}/kubectl" rollout status \
    -n kube-system \
    deployment/metrics-server \
    --timeout="${metrics_timeout}"
  "${K9S_NEO_TOOLS_BIN}/kubectl" wait \
    --for=condition=Available \
    apiservice/v1beta1.metrics.k8s.io \
    --timeout="${metrics_timeout}"

  local attempt
  for attempt in $(seq 1 30); do
    if "${K9S_NEO_TOOLS_BIN}/kubectl" top nodes >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done

  echo "metrics-server addon enabled but metrics API did not become queryable in time" >&2
  return 1
}

if ! "${K9S_NEO_TOOLS_BIN}/colima" status --profile="${K9S_NEO_COLIMA_PROFILE}" >/dev/null 2>&1; then
  "${K9S_NEO_TOOLS_BIN}/colima" start \
    --profile="${K9S_NEO_COLIMA_PROFILE}" \
    --cpu="${colima_cpus}" \
    --memory="${colima_memory_gb}" \
    --disk="${colima_disk_gb}"
fi

"${K9S_NEO_TOOLS_BIN}/docker" version >/dev/null

"${K9S_NEO_TOOLS_BIN}/minikube" start \
  --profile="${K9S_NEO_PROFILE}" \
  --driver="${driver}" \
  --nodes="${minikube_nodes}" \
  --cpus="${minikube_cpus}" \
  --memory="${minikube_memory_mb}" \
  --disk-size="${minikube_disk_size}" \
  --container-runtime="${container_runtime}"

if [[ "${enable_metrics_server}" == "1" ]]; then
  "${K9S_NEO_TOOLS_BIN}/minikube" addons enable metrics-server --profile="${K9S_NEO_PROFILE}"
  wait_for_metrics_server
fi

"${K9S_NEO_TOOLS_BIN}/kubectl" config use-context "${K9S_NEO_PROFILE}" >/dev/null
"${K9S_NEO_TOOLS_BIN}/kubectl" cluster-info
"${K9S_NEO_TOOLS_BIN}/kubectl" get nodes -o wide
