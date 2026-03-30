#!/usr/bin/env bash

set -euo pipefail

source "$(cd "$(dirname "$0")" && pwd)/common.sh"

need_cmd python3

bin_path="${K9S_NEO_BIN:-${K9S_NEO_REPO_ROOT}/execs/k9s}"
vars_path="${K9S_NEO_VARS_PATH:-${K9S_NEO_REPO_ROOT}/hack/bench/vars.nodes-small.json}"
label="${K9S_NEO_BENCH_LABEL:-step7-nodes-small-$(date +%Y%m%d-%H%M%S)}"
cold_runs="${K9S_NEO_COLD_RUNS:-1}"
warm_runs="${K9S_NEO_WARM_RUNS:-1}"
ensure_cluster="${K9S_NEO_ENSURE_CLUSTER:-1}"
write_note_path="${K9S_NEO_WRITE_NODES_SMALL_NOTE:-}"

usage() {
  cat <<EOF
usage: $(basename "$0") [--label LABEL] [--cold-runs N] [--warm-runs N] [--bin PATH] [--vars PATH] [--write-note PATH] [--no-ensure-cluster]

Runs the nodes-small Step 7A control scenario set:
  - nodes_first_render
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --label)
      label="$2"
      shift 2
      ;;
    --cold-runs)
      cold_runs="$2"
      shift 2
      ;;
    --warm-runs)
      warm_runs="$2"
      shift 2
      ;;
    --bin)
      bin_path="$2"
      shift 2
      ;;
    --vars)
      vars_path="$2"
      shift 2
      ;;
    --write-note)
      write_note_path="$2"
      shift 2
      ;;
    --no-ensure-cluster)
      ensure_cluster="0"
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ ! -x "${bin_path}" ]]; then
  echo "missing executable K9s Neo binary: ${bin_path}" >&2
  echo "build it first with ./hack/with-go.sh make build" >&2
  exit 1
fi

if [[ "${ensure_cluster}" == "1" ]]; then
  "${K9S_NEO_REPO_ROOT}/hack/local-lab/start-nodes-small.sh"
  "${K9S_NEO_REPO_ROOT}/hack/local-lab/seed-nodes-small.sh"
fi

if [[ ! -f "${vars_path}" ]]; then
  "${K9S_NEO_REPO_ROOT}/hack/local-lab/write-vars-nodes-small.sh" "${vars_path}"
fi

run_output="$(
  python3 "${K9S_NEO_REPO_ROOT}/hack/bench/run.py" \
    --bin "${bin_path}" \
    --label "${label}" \
    --vars "${vars_path}" \
    --cold-runs "${cold_runs}" \
    --warm-runs "${warm_runs}" \
    --scenario nodes_first_render
)"

printf '%s\n' "${run_output}"

artifact_root="$(
  printf '%s\n' "${run_output}" | python3 -c 'import json, sys; print(json.load(sys.stdin)["artifact_root"])'
)"

if [[ -n "${write_note_path}" ]]; then
  "${K9S_NEO_REPO_ROOT}/hack/local-lab/write-step7a-nodes-small-note.sh" "${artifact_root}" "${write_note_path}"
fi
