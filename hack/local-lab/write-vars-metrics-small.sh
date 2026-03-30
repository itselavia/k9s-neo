#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
source "${script_dir}/metrics-small-env.sh"

default_output="${script_dir}/../bench/vars.metrics-small.json"

if [[ $# -eq 0 ]]; then
  exec "${script_dir}/write-vars.sh" "${default_output}"
fi

exec "${script_dir}/write-vars.sh" "$@"
