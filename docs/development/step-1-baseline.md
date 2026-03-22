# Step 1 Baseline

This file records the local development baseline before any instrumentation or product changes.

## Environment

- Date: 2026-03-22
- Machine class: `Darwin arm64`
- Repo commit: `b67a756394eaac1b36d8d713afe1463728e8d249`
- Required Go version: `go1.25.8`
- Installed Go location: `~/.k9s-neo-tools/go/1.25.8`

## Verification

Toolchain bootstrap:

```bash
./hack/bootstrap-go.sh
./hack/with-go.sh go version
```

Observed result:

```text
go version go1.25.8 darwin/arm64
```

Baseline build:

```bash
GOCACHE=/tmp/k9s-neo-gocache \
GOMODCACHE=/tmp/k9s-neo-gomodcache \
GOPATH=/tmp/k9s-neo-gopath \
./hack/with-go.sh make build
```

Observed artifact:

```text
./execs/k9s
```

Binary verification:

```bash
./execs/k9s version
```

Observed version:

```text
Version: v0.50.18
Commit:  b67a7563
```

## Notes

- The bootstrap path is intentionally repo-owned and does not mutate shell startup files.
- `/tmp` caches keep the repo working tree clean and make local cleanup simple.
- In restricted environments, `./execs/k9s version` may warn if it cannot create its log directory under `~/Library/Application Support/k9s`. That does not affect build verification.
