#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

detect_go_version() {
  python3 - "$ROOT_DIR/go.mod" <<'PY'
import pathlib
import re
import sys

mod = pathlib.Path(sys.argv[1]).read_text()
match = re.search(r"^go\s+([0-9]+\.[0-9]+(?:\.[0-9]+)?)\s*$", mod, re.MULTILINE)
if not match:
    raise SystemExit("unable to locate Go version in go.mod")
print(match.group(1))
PY
}

if [[ $# -eq 0 ]]; then
  echo "usage: ./hack/with-go.sh <command> [args...]" >&2
  exit 1
fi

GO_VERSION="${GO_VERSION:-$(detect_go_version)}"
TOOLS_DIR="${K9S_NEO_TOOLS_DIR:-$HOME/.k9s-neo-tools}"
GOROOT="$TOOLS_DIR/go/$GO_VERSION"

if [[ ! -x "$GOROOT/bin/go" ]]; then
  echo "Go $GO_VERSION is not installed at $GOROOT" >&2
  echo "Run ./hack/bootstrap-go.sh first." >&2
  exit 1
fi

export GOROOT
export PATH="$GOROOT/bin:$PATH"

exec "$@"
