#!/usr/bin/env bash

set -euo pipefail

source "$(cd "$(dirname "$0")" && pwd)/common.sh"

if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "usage: $(basename "$0") ARTIFACT_ROOT [OUTPUT_PATH]" >&2
  exit 1
fi

artifact_root="$1"
output_path="${2:-${K9S_NEO_REPO_ROOT}/docs/development/step-6-local-baseline-note.md}"
commit_sha="$(git -C "${K9S_NEO_REPO_ROOT}" rev-parse HEAD)"
captured_at="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

cat >"${output_path}" <<EOF
# Step 6 Local Baseline Note

- Captured at: ${captured_at}
- Git commit: \`${commit_sha}\`
- Artifact root: \`${artifact_root}\`
- Environment: local disposable cluster
- Cluster path: \`colima\` plus \`minikube --driver=docker\`
- Minikube profile: \`${K9S_NEO_PROFILE}\`
- Namespace: \`${K9S_NEO_NAMESPACE}\`
- Kubeconfig: \`${K9S_NEO_KUBECONFIG}\`

## Required Scenario Set

- \`pods_startup\`
- \`pods_filter_settle\`
- \`nodes_first_render\`
- \`pod_yaml\`
- \`pod_describe\`

## Environment Defaults

- minikube CPUs: \`${K9S_NEO_MINIKUBE_CPUS:-2}\`
- minikube memory: \`${K9S_NEO_MINIKUBE_MEMORY_MB:-2048} MiB\`
- minikube disk: \`${K9S_NEO_MINIKUBE_DISK_SIZE:-20g}\`
- container runtime: \`${K9S_NEO_MINIKUBE_CONTAINER_RUNTIME:-containerd}\`
- colima CPUs: \`${K9S_NEO_COLIMA_CPUS:-2}\`
- colima memory: \`${K9S_NEO_COLIMA_MEMORY_GB:-3} GiB\`
- colima disk: \`${K9S_NEO_COLIMA_DISK_GB:-20} GiB\`
- seed manifest: \`hack/local-lab/manifests/neo-bench.yaml\`

## Notes

- This baseline is valid for before-and-after engineering comparisons on this machine.
- This baseline is not production-cluster evidence.
- Recreate the local lab from repo-owned scripts before comparing later shallow changes.
EOF

echo "wrote ${output_path}"
