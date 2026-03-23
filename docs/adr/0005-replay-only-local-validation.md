# ADR 0005: Replay-Only Local Validation

- Status: Accepted
- Date: 2026-03-22

## Context

Step 4 established a live PTY benchmark harness and benchmark-grade lifecycle markers, but this personal development machine does not have cluster access. We still need a local way to validate the harness, artifact generation, and summary math before moving to real-cluster baselines.

The original Step 5 space included both replay fixtures and a possible synthetic API server. A synthetic server would add more code surface and more fake-Kubernetes maintenance burden than we need right now.

## Decision

Step 5 uses replay fixtures only.

- checked-in replay fixtures are the only required local validation mechanism
- replay artifacts validate harness correctness, not Kubernetes behavior or performance
- synthetic server work is explicitly deferred
- only `live` results from the work machine may support benchmark claims
- replay fixtures must be synthetic or sanitized and small enough to review comfortably

Replay validation must exercise:

- startup timing derivation
- multi-step `view_seq` anchor scoping
- detail timing
- `no_data` handling and aggregate exclusion of non-`ok` runs

Replay validation is CI-enforced so harness/report regressions cannot silently drift.

## Consequences

Positive:

- Step 5 stays narrow, boring, and maintainable.
- We validate the real report path without faking Kubernetes behavior.
- Public-safe validation artifacts are easy to review and keep deterministic.

Negative:

- Replay does not prove the live binary can talk to a Kubernetes-shaped API.
- Full end-to-end local smoke remains deferred until later need justifies it.

## Deferred Questions

- Whether a tiny synthetic API smoke path is worth adding later.
- Whether replay validation should eventually move into `make test` instead of its own CI step.
- How much sanitized live-cluster raw data should be checked in after Step 6.
