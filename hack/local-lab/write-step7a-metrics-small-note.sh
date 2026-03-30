#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
source "${script_dir}/metrics-small-env.sh"
source "${script_dir}/common.sh"

if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "usage: $(basename "$0") ARTIFACT_ROOT [OUTPUT_PATH]" >&2
  exit 1
fi

artifact_root="$1"
output_path="${2:-${K9S_NEO_REPO_ROOT}/docs/development/step-7a-metrics-small-control-note.md}"
commit_sha="$(git -C "${K9S_NEO_REPO_ROOT}" rev-parse HEAD)"
captured_at="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
vars_path="${K9S_NEO_VARS_PATH:-${K9S_NEO_REPO_ROOT}/hack/bench/vars.metrics-small.json}"
vars_path="$(cd "$(dirname "${vars_path}")" && pwd)/$(basename "${vars_path}")"

cat >"${output_path}" <<EOF
# Step 7A Metrics-Small Control Note

- Captured at: ${captured_at}
- Git commit: \`${commit_sha}\`
- Artifact root: \`${artifact_root}\`
- Environment: local disposable cluster
- Profile: \`metrics-small\`
- Cluster path: \`colima\` plus \`minikube --driver=docker\`
- Minikube profile: \`${K9S_NEO_PROFILE}\`
- Colima backend profile: \`${K9S_NEO_COLIMA_PROFILE}\`
- Namespace: \`${K9S_NEO_NAMESPACE}\`
- Kubeconfig: \`${K9S_NEO_KUBECONFIG}\`
- Vars file: \`${vars_path}\`
- Metrics-server: enabled

## Required Scenario Set

- \`pods_startup\`
- \`pods_filter_settle\`
- \`nodes_first_render\`
- \`pod_yaml\`
- \`pod_describe\`

## Control Flags Under Test

The control binary wrapper adds only the promoted Step 7A flags:

- \`--perf-skip-crd-augment\`
- \`--perf-skip-namespace-validation\`

Ambient metrics remain enabled in this control run.

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

- This artifact is the control for later ambient-metrics A/B work on \`metrics-small\`.
- This control does not replace the frozen \`control-small\` Step 6 and Step 7A artifacts.
- Validate the trace \`config_snapshot\` when comparing later candidates:
  - \`perf_skip_crd_augment=true\`
  - \`perf_skip_namespace_validation=true\`
  - \`perf_disable_ambient_metrics\` absent or \`false\`
EOF

echo "wrote ${output_path}"
