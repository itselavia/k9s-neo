#!/usr/bin/env bash

set -euo pipefail

K9S_NEO_REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
K9S_NEO_TOOLS_HOME="${K9S_NEO_TOOLS_HOME:-$HOME/.k9s-neo-tools}"
K9S_NEO_STATE_HOME="${K9S_NEO_STATE_HOME:-$HOME/.k9s-neo-state}"
K9S_NEO_TOOLS_BIN="${K9S_NEO_TOOLS_HOME}/bin"
K9S_NEO_DOWNLOADS_DIR="${K9S_NEO_TOOLS_HOME}/downloads"
K9S_NEO_LIMA_HOME="${K9S_NEO_TOOLS_HOME}/lima"
K9S_NEO_MINIKUBE_HOME="${K9S_NEO_STATE_HOME}/minikube"
K9S_NEO_COLIMA_HOME="${K9S_NEO_STATE_HOME}/colima"
K9S_NEO_KUBECONFIG="${K9S_NEO_STATE_HOME}/kubeconfig"
K9S_NEO_PROFILE="${K9S_NEO_PROFILE:-k9s-neo}"
K9S_NEO_COLIMA_PROFILE="${K9S_NEO_COLIMA_PROFILE:-$K9S_NEO_PROFILE}"
K9S_NEO_NAMESPACE="${K9S_NEO_NAMESPACE:-neo-bench}"
K9S_NEO_MINIKUBE_NODES="${K9S_NEO_MINIKUBE_NODES:-1}"
K9S_NEO_MANIFEST="${K9S_NEO_MANIFEST:-${K9S_NEO_REPO_ROOT}/hack/local-lab/manifests/neo-bench.yaml}"
K9S_NEO_DOCKER_SOCK="${K9S_NEO_COLIMA_HOME}/${K9S_NEO_COLIMA_PROFILE}/docker.sock"

mkdir -p \
  "${K9S_NEO_TOOLS_BIN}" \
  "${K9S_NEO_DOWNLOADS_DIR}" \
  "${K9S_NEO_MINIKUBE_HOME}" \
  "${K9S_NEO_COLIMA_HOME}" \
  "${K9S_NEO_STATE_HOME}"

export PATH="${K9S_NEO_LIMA_HOME}/bin:${K9S_NEO_TOOLS_BIN}:${PATH}"
export MINIKUBE_HOME="${K9S_NEO_MINIKUBE_HOME}"
export KUBECONFIG="${K9S_NEO_KUBECONFIG}"
export COLIMA_HOME="${K9S_NEO_COLIMA_HOME}"
export DOCKER_HOST="unix://${K9S_NEO_DOCKER_SOCK}"

need_cmd() {
  local name="$1"

  if command -v "${name}" >/dev/null 2>&1; then
    return 0
  fi

  echo "missing required command: ${name}" >&2
  exit 1
}

need_local_tool() {
  local name="$1"

  if [[ -x "${K9S_NEO_TOOLS_BIN}/${name}" ]]; then
    return 0
  fi

  echo "missing local tool: ${K9S_NEO_TOOLS_BIN}/${name}" >&2
  echo "run hack/local-lab/install-tools.sh first" >&2
  exit 1
}

print_env_summary() {
  cat <<EOF
repo_root=${K9S_NEO_REPO_ROOT}
tools_bin=${K9S_NEO_TOOLS_BIN}
lima_home=${K9S_NEO_LIMA_HOME}
state_home=${K9S_NEO_STATE_HOME}
minikube_home=${K9S_NEO_MINIKUBE_HOME}
colima_home=${K9S_NEO_COLIMA_HOME}
kubeconfig=${K9S_NEO_KUBECONFIG}
docker_host=${DOCKER_HOST}
profile=${K9S_NEO_PROFILE}
colima_profile=${K9S_NEO_COLIMA_PROFILE}
minikube_nodes=${K9S_NEO_MINIKUBE_NODES}
namespace=${K9S_NEO_NAMESPACE}
manifest=${K9S_NEO_MANIFEST}
EOF
}
