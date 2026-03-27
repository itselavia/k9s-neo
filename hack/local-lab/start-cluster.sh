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
container_runtime="${K9S_NEO_MINIKUBE_CONTAINER_RUNTIME:-containerd}"
colima_cpus="${K9S_NEO_COLIMA_CPUS:-2}"
colima_memory_gb="${K9S_NEO_COLIMA_MEMORY_GB:-3}"
colima_disk_gb="${K9S_NEO_COLIMA_DISK_GB:-20}"

if ! "${K9S_NEO_TOOLS_BIN}/colima" status --profile="${K9S_NEO_PROFILE}" >/dev/null 2>&1; then
  "${K9S_NEO_TOOLS_BIN}/colima" start \
    --profile="${K9S_NEO_PROFILE}" \
    --cpu="${colima_cpus}" \
    --memory="${colima_memory_gb}" \
    --disk="${colima_disk_gb}"
fi

"${K9S_NEO_TOOLS_BIN}/docker" version >/dev/null

"${K9S_NEO_TOOLS_BIN}/minikube" start \
  --profile="${K9S_NEO_PROFILE}" \
  --driver="${driver}" \
  --cpus="${minikube_cpus}" \
  --memory="${minikube_memory_mb}" \
  --disk-size="${minikube_disk_size}" \
  --container-runtime="${container_runtime}"

"${K9S_NEO_TOOLS_BIN}/kubectl" config use-context "${K9S_NEO_PROFILE}" >/dev/null
"${K9S_NEO_TOOLS_BIN}/kubectl" cluster-info
"${K9S_NEO_TOOLS_BIN}/kubectl" get nodes -o wide
