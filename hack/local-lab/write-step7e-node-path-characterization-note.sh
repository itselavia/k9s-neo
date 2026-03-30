#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
source "${script_dir}/nodes-small-env.sh"
source "${script_dir}/common.sh"

if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "usage: $(basename "$0") ARTIFACT_ROOT [OUTPUT_PATH]" >&2
  exit 1
fi

artifact_root="$1"
output_path="${2:-${K9S_NEO_REPO_ROOT}/docs/development/step-7e-node-path-characterization-note.md}"
control_note_path="${K9S_NEO_REPO_ROOT}/docs/development/step-7a-nodes-small-control-note.md"
vars_path="${K9S_NEO_VARS_PATH:-${K9S_NEO_REPO_ROOT}/hack/bench/vars.nodes-small.json}"

python3 "${script_dir}/node_path_characterize.py" \
  --artifact-root "${artifact_root}" \
  --control-note "${control_note_path}" \
  --vars "${vars_path}" \
  --output "${output_path}"

