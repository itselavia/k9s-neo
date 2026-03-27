#!/usr/bin/env bash

set -euo pipefail

source "$(cd "$(dirname "$0")" && pwd)/common.sh"

need_cmd curl
need_cmd python3
need_cmd shasum
need_cmd install
need_cmd tar

tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT

download() {
  local url="$1"
  local output="$2"

  curl -fsSL -o "${output}" "${url}"
}

latest_github_tag() {
  local repo="$1"

  python3 - "$repo" <<'PY'
import json
import sys
import urllib.request

repo = sys.argv[1]
url = f"https://api.github.com/repos/{repo}/releases/latest"
with urllib.request.urlopen(url) as response:
    payload = json.load(response)
print(payload["tag_name"])
PY
}

verify_sha256() {
  local checksum_file="$1"
  local target_file="$2"

  echo "$(cat "${checksum_file}")  ${target_file}" | shasum -a 256 --check >/dev/null
}

echo "Installing local lab tools into ${K9S_NEO_TOOLS_BIN}"

kubectl_version="$(curl -fsSL https://dl.k8s.io/release/stable.txt)"
download "https://dl.k8s.io/release/${kubectl_version}/bin/darwin/arm64/kubectl" "${tmp_dir}/kubectl"
download "https://dl.k8s.io/release/${kubectl_version}/bin/darwin/arm64/kubectl.sha256" "${tmp_dir}/kubectl.sha256"
verify_sha256 "${tmp_dir}/kubectl.sha256" "${tmp_dir}/kubectl"
install -m 0755 "${tmp_dir}/kubectl" "${K9S_NEO_TOOLS_BIN}/kubectl"

minikube_tag="$(latest_github_tag kubernetes/minikube)"
download "https://github.com/kubernetes/minikube/releases/download/${minikube_tag}/minikube-darwin-arm64" "${tmp_dir}/minikube"
download "https://github.com/kubernetes/minikube/releases/download/${minikube_tag}/minikube-darwin-arm64.sha256" "${tmp_dir}/minikube.sha256"
verify_sha256 "${tmp_dir}/minikube.sha256" "${tmp_dir}/minikube"
install -m 0755 "${tmp_dir}/minikube" "${K9S_NEO_TOOLS_BIN}/minikube"

colima_tag="$(latest_github_tag abiosoft/colima)"
download "https://github.com/abiosoft/colima/releases/download/${colima_tag}/colima-Darwin-arm64" "${tmp_dir}/colima"
chmod +x "${tmp_dir}/colima"
xattr -d com.apple.quarantine "${tmp_dir}/colima" >/dev/null 2>&1 || true
install -m 0755 "${tmp_dir}/colima" "${K9S_NEO_TOOLS_BIN}/colima"

lima_tag="$(latest_github_tag lima-vm/lima)"
download "https://github.com/lima-vm/lima/releases/download/${lima_tag}/lima-${lima_tag#v}-Darwin-arm64.tar.gz" "${tmp_dir}/lima.tar.gz"
rm -rf "${K9S_NEO_LIMA_HOME}"
mkdir -p "${K9S_NEO_LIMA_HOME}"
tar -xzf "${tmp_dir}/lima.tar.gz" -C "${K9S_NEO_LIMA_HOME}" --strip-components=1

docker_version="${K9S_NEO_DOCKER_CLI_VERSION:-29.3.1}"
download "https://download.docker.com/mac/static/stable/aarch64/docker-${docker_version}.tgz" "${tmp_dir}/docker.tgz"
mkdir -p "${tmp_dir}/docker"
tar -xzf "${tmp_dir}/docker.tgz" -C "${tmp_dir}/docker"
install -m 0755 "${tmp_dir}/docker/docker/docker" "${K9S_NEO_TOOLS_BIN}/docker"

cat <<EOF
Installed:
  kubectl ${kubectl_version}
  minikube ${minikube_tag}
  colima ${colima_tag}
  lima ${lima_tag}
  docker ${docker_version}

Environment:
$(print_env_summary)
EOF
