# ADR 0004: Lifecycle Markers And Benchmark Harness

- Status: Accepted
- Date: 2026-03-22

## Context

ADR 0003 established request-level transport tracing, but request traces alone are not enough to benchmark the hot paths that matter for this fork. We need deterministic lifecycle markers inside the application so startup, first useful row, filter settle, and detail readiness can be measured without depending on fragile screen scraping.

We also need a rerunnable external harness that can drive scenarios, collect raw traces, sample local resource usage, and write artifact formats that are easy to inspect and later sanitize.

## Decision

We extend the perf trace schema to version `2` and add lifecycle markers as structured JSONL events.

Step 4 lifecycle markers are:

- `config_init_start`
- `config_init_end`
- `connection_init_start`
- `connection_init_end`
- `metrics_probe_start`
- `metrics_probe_end`
- `discovery_start`
- `discovery_end`
- `auth_preflight_start`
- `auth_preflight_end`
- `command_start`
- `view_activate`
- `first_model_built`
- `first_render_committed`
- `first_useful_row`
- `first_key_after_render`
- `filter_start`
- `filter_settle`
- `detail_open_start`
- `detail_content_ready`

All events include `since_process_start_ms`. View-scoped lifecycle events also carry a `view_seq` plus view metadata so the harness can distinguish a newly activated view from a resumed or background one.

Marker semantics are intentionally strict:

- `first_render_committed` is emitted from the post-draw hook for the active main view.
- `first_useful_row` is emitted only after draw commit, when at least one non-header row is actually visible for that view.
- `detail_open_start` is emitted at the user action boundary that opens logs, YAML, or describe, before the expensive work for that detail flow begins.

The benchmark harness lives under `hack/bench/` and is intentionally external to the binary:

- `run.py` is the only executable entrypoint.
- `scenarios.json` is checked in and defines scenario steps.
- `vars.example.json` is checked in as a template.
- `vars.local.json` is untracked and holds machine-specific inputs.

The harness uses internal lifecycle markers as the primary timing source. PTY transcript capture is only secondary validation.
Per-scenario timing is derived from the marker sequence that actually satisfied the scenario, not from the first matching marker seen anywhere in the trace. Aggregate summary statistics are computed from `ok` runs only; non-`ok` runs remain visible in raw artifacts for debugging and failure analysis.

## Consequences

Positive:

- Benchmark timing is tied to explicit application lifecycle events instead of terminal heuristics.
- The harness remains simple, scriptable, and portable to the work machine later.
- Raw per-run artifacts are suitable for later sanitization and publication.

Negative:

- Trace schema consumers must now understand schema version `2`.
- Trace files can contain sensitive context, namespace, and request metadata; sanitization is still required before publication.
- The harness adds some repo surface area in Python, JSON, and local artifact directories.

## Deferred Questions

- Synthetic and replay fixtures remain a separate step.
- Real-cluster benchmark baselines remain a work-machine step.
- Any public benchmark claims remain blocked on sanitized live-cluster results.
