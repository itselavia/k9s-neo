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

detect_platform() {
  local os arch machine

  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  machine="$(uname -m)"

  case "$machine" in
    arm64|aarch64)
      arch="arm64"
      ;;
    x86_64|amd64)
      arch="amd64"
      ;;
    *)
      echo "unsupported architecture: $machine" >&2
      exit 1
      ;;
  esac

  case "$os" in
    darwin|linux)
      printf '%s %s\n' "$os" "$arch"
      ;;
    *)
      echo "unsupported operating system: $os" >&2
      exit 1
      ;;
  esac
}

lookup_release() {
  python3 - "$1" "$2" "$3" "$4" <<'PY'
import json
import sys

payload, version, os_name, arch = sys.argv[1:5]
data = json.load(open(payload))

for release in data:
    if release["version"] != version:
        continue
    for entry in release["files"]:
        if entry["os"] == os_name and entry["arch"] == arch and entry["kind"] == "archive":
            print(entry["filename"])
            print(entry["sha256"])
            raise SystemExit(0)

raise SystemExit(f"unable to find archive for {version} {os_name}/{arch}")
PY
}

GO_VERSION="${GO_VERSION:-$(detect_go_version)}"
read -r OS_NAME ARCH_NAME <<<"$(detect_platform)"
TOOLS_DIR="${K9S_NEO_TOOLS_DIR:-$HOME/.k9s-neo-tools}"
INSTALL_DIR="$TOOLS_DIR/go/$GO_VERSION"
CURRENT_LINK="$TOOLS_DIR/go/current"

if [[ -x "$INSTALL_DIR/bin/go" ]]; then
  mkdir -p "$(dirname "$CURRENT_LINK")"
  ln -sfn "$INSTALL_DIR" "$CURRENT_LINK"
  cat <<EOF
Go $GO_VERSION is already installed at:
  $INSTALL_DIR

Use it with:
  ./hack/with-go.sh go version
EOF
  exit 0
fi

TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/k9s-neo-go.XXXXXX")"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

RELEASE_JSON="$TMP_DIR/releases.json"
/usr/bin/curl -fsSL "https://go.dev/dl/?mode=json&include=all" -o "$RELEASE_JSON"
RELEASE_INFO="$(lookup_release "$RELEASE_JSON" "go${GO_VERSION}" "$OS_NAME" "$ARCH_NAME")"
ARCHIVE_NAME="$(printf '%s\n' "$RELEASE_INFO" | sed -n '1p')"
EXPECTED_SHA="$(printf '%s\n' "$RELEASE_INFO" | sed -n '2p')"
ARCHIVE_PATH="$TMP_DIR/$ARCHIVE_NAME"

/usr/bin/curl -fsSL "https://go.dev/dl/${ARCHIVE_NAME}" -o "$ARCHIVE_PATH"
ACTUAL_SHA="$(/usr/bin/shasum -a 256 "$ARCHIVE_PATH" | awk '{print $1}')"
if [[ "$ACTUAL_SHA" != "$EXPECTED_SHA" ]]; then
  echo "checksum mismatch for $ARCHIVE_NAME" >&2
  echo "expected: $EXPECTED_SHA" >&2
  echo "actual:   $ACTUAL_SHA" >&2
  exit 1
fi

EXTRACT_DIR="$TMP_DIR/extract"
mkdir -p "$EXTRACT_DIR"
/usr/bin/tar -C "$EXTRACT_DIR" -xzf "$ARCHIVE_PATH"
mkdir -p "$(dirname "$INSTALL_DIR")"
rm -rf "$INSTALL_DIR"
mv "$EXTRACT_DIR/go" "$INSTALL_DIR"
ln -sfn "$INSTALL_DIR" "$CURRENT_LINK"

cat <<EOF
Installed Go $GO_VERSION to:
  $INSTALL_DIR

Use it with:
  ./hack/with-go.sh go version
  ./hack/with-go.sh go build ./...
EOF
