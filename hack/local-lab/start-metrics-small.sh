#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
source "${script_dir}/metrics-small-env.sh"

exec "${script_dir}/start-cluster.sh" "$@"
