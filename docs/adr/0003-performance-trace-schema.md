# 0003 Performance Trace Schema

- Status: Accepted
- Date: 2026-03-22

## Context

K9s Neo needs benchmark-grade evidence before any high-impact cuts are justified. Step 3 is the first instrumentation step: establish a stable request-level trace format and runtime plumbing without yet changing the UI lifecycle or benchmark harness behavior.

The trace needs to be:

- explicit and machine-readable
- disabled by default
- runtime-only, not persisted in user config
- extensible enough for later lifecycle markers

## Decision

K9s Neo uses a JSONL perf trace format, one JSON object per line.

Tracing is enabled only when `--perf-trace-file` is set. The runtime also accepts hidden metadata flags `--perf-trace-scenario` and `--perf-trace-run-id`.

Step 3 emits only these event types:

- `session_start`
- `config_snapshot`
- `kube_request_complete`
- `kube_stream_open`
- `kube_stream_close`
- `session_end`

The schema uses a common event envelope with:

- `schema_version`
- `seq`
- `ts`
- `type`
- `run_id`
- `scenario`

Request events also capture:

- client role
- request classification
- resource, namespace, and subresource attribution
- status code
- duration
- response bytes
- best-effort response kind and list item count

`response_bytes` means bytes read by the client from `resp.Body` after transport decoding. It is not guaranteed to be exact wire bytes.

`item_count` is best-effort only. It is emitted for safely inspected non-streaming JSON bodies and omitted when inspection is skipped or unsafe.

This ADR does not add lifecycle markers such as first render or first useful row. Those belong to Step 4.

## Consequences

- Bench traces become stable enough for harness and analysis work.
- Step 3 stays boring and low-risk by instrumenting the transport seam instead of many higher-level call sites.
- Large or streaming responses still produce useful request timing and byte data even when body inspection is skipped.
- Raw trace files may contain sensitive cluster metadata and are not public-safe artifacts until later sanitization.

## Deferred Questions

- Exact lifecycle event types for startup, model build, render, and detail hydration.
- Whether later steps need higher-level hydration counters beyond transport-level request data.
- How synthetic or replay fixtures should represent sanitized trace data.
