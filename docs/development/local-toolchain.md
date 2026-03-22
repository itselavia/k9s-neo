# Local Toolchain Bootstrap

This repo expects the Go version declared in `go.mod`.

For local development on machines without Homebrew or a preinstalled Go toolchain, use the repo-owned bootstrap script:

```bash
./hack/bootstrap-go.sh
./hack/with-go.sh go version
```

The installer:

- reads the required Go version from `go.mod`
- downloads the official Go archive from `go.dev`
- verifies the archive checksum against the official release metadata
- installs Go under `~/.k9s-neo-tools/go/<version>`
- avoids mutating your shell profile

For builds in fresh environments, keep caches out of the repo and out of system-global locations:

```bash
GOCACHE=/tmp/k9s-neo-gocache \
GOMODCACHE=/tmp/k9s-neo-gomodcache \
GOPATH=/tmp/k9s-neo-gopath \
./hack/with-go.sh go build -o ./execs/k9s ./main.go
```

You can use the same pattern on the work machine later.

`kubectl` is not required for local code work on this personal machine. It only becomes necessary when running live-cluster benchmarks or real-cluster validation.
