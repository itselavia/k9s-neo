# K9s Neo Project Brief

This file captures the current shared project context for Codex threads working in this repo.
Update it when major decisions change.

Canonical decision records live in `docs/adr/`.
Current accepted product contracts:

- `docs/adr/0001-v0-product-contract.md`
- `docs/adr/0002-read-only-safety-contract.md`
- `docs/adr/0003-performance-trace-schema.md`
- `docs/adr/0004-lifecycle-markers-and-benchmark-harness.md`
- `docs/adr/0005-replay-only-local-validation.md`

## Mission

Fork K9s into a read-only, scale-first Kubernetes triage TUI for very large clusters.

Primary use is internal SRE triage, but development is public and should be credible as a real example of LLM-assisted work on a complex codebase.

## Product Thesis

Build a narrower tool that is obviously faster and safer for high-cardinality SRE triage.

Preserve this core loop:

- enter a scoped view fast
- find a workload or pod fast
- filter and search fast
- inspect logs
- inspect events
- inspect YAML and describe
- drill into the selected object only when needed

## Non-Goals

- do not replace all of K9s for all users
- do not optimize cosmetics first
- do not rewrite from scratch unless measurements prove surgical changes are insufficient
- do not preserve mutation-capable behavior in the final product

## Read-Only Contract

Final product must not support:

- edit
- delete
- apply
- replace
- patch
- scale
- restart
- rollback
- port-forward
- shell
- exec
- attach
- transfer
- helper pods
- debug containers
- node shell flows
- plugin loading or execution

Do not settle for "hidden in the UI". The end state must be defensibly read-only.

## Engineering Rules

- Evidence over vibes.
- No speed claims without measurements.
- No large refactor before baseline instrumentation and benchmarks exist.
- Add instrumentation and kill switches before hard deletions.
- Prefer a thin downstream patch stack unless data justifies deeper divergence.
- Every cut must do at least one of:
  - materially improve a hot path
  - eliminate a real failure mode
  - materially reduce safety risk
  - materially reduce maintenance surface
- If a cut has less than 5% hot-path impact and little safety benefit, mark it low priority.
- Keep the work boring, legible, and rebaseable.

## Success Criteria

1. Beat stock K9s by at least 2x on at least one benchmark that matters, or turn one broken path into a working one.
2. Produce a ranked decision matrix with estimated and actual impact bands.
3. Preserve the basic SRE triage loop.
4. End state is truly read-only.
5. Produce public-safe artifacts: ADRs, sanitized benchmark reports, synthetic or replay fixtures if needed, and code.

## CRD Policy

Do not support generic CRDs broadly on the startup hot path.

Policy for v0:

- generic CRDs are off by default
- explicit allowlist CRDs are supported
- broad CRD discovery is opt-in or deferred

Agones is first-class and in scope from the start.

Initial Agones allowlist:

- `gameservers.agones.dev`
- `fleets.agones.dev`
- `gameserverallocations.allocation.agones.dev`
- `fleetautoscalers.autoscaling.agones.dev`

Support Agones through curated aliases and explicit resource handling, not broad CRD startup discovery.

## Benchmarking Rules

Baseline and fork must use the same instrumentation patch when compared.

Primary scenarios:

- startup to first useful row in big-namespace pod view
- startup to stable interactive state in that view
- filter-settle latency after regex search
- nodes first render
- node to pod drill-down on a large node
- all-namespaces pod view when relevant and permitted
- open logs, events, YAML, and describe for a selected pod

For each scenario, capture:

- time to first paint
- time to first useful row
- time to stable interactive state
- peak RSS in first 30 seconds
- peak CPU in first 30 seconds
- API request count by resource and verb
- total response bytes before first useful row
- total object count hydrated before first useful row
- watch count
- whether watches start before or after first useful render

Use explicit lifecycle markers in code as the primary measurement source.
TTY scraping is secondary validation only.

## Development Context

Current intended workflow:

- personal machine: local development, instrumentation, unit tests, fake-client tests, replay fixtures, harness work, docs
- work machine: real-cluster benchmarking and final proof

Important:

- do not block local development on live-cluster access
- do not make public performance claims from synthetic-only results

## Current Bottleneck Hypotheses

Most likely early wins:

1. ambient metrics are on startup and list-view hot paths
2. startup discovery and CRD loading are broader than needed
3. node views do cluster-wide pod work that may be catastrophic at scale
4. read preflight auth fan-out adds avoidable latency and failure modes

## Serialized Roadmap

1. Local toolchain bootstrap and baseline build.
2. Freeze the v0 product contract in ADRs.
3. Add performance trace schema and request-level instrumentation.
4. Add lifecycle markers and the rerunnable benchmark harness.
5. Validate the harness locally with replay fixtures.
6. Capture real baseline on the work machine.
7. Land shallow, high-confidence wins one by one:
   - lazy metrics by default
   - disable node pod counting by default
   - static core aliases plus Agones allowlist
   - strict read-only hardening
8. Re-benchmark after each change and update the decision matrix.
9. Only then consider deeper changes:
   - node drill-down path repair
   - server-side Table responses for hot views
   - metadata-first rendering
   - chunking
   - improved list and watch behavior

## Immediate Next Tasks

The next tasks should be:

1. capture the real baseline on the work machine
2. begin measuring shallow, high-confidence changes one by one
3. keep deeper data-path work blocked on benchmark evidence
4. keep strict read-only hardening separate from benchmark-baseline capture

Step 4 is complete in this branch:

- ADR-backed lifecycle trace schema v2
- startup, discovery, auth, render, filter, and detail lifecycle markers
- external benchmark harness in `hack/bench/`
- checked-in scenario manifest and local vars template
- corrected marker semantics for post-draw useful-row timing and action-boundary detail timing
- scenario-anchor-aware harness metric derivation for multi-step flows
- `ok`-only aggregate summaries with raw failed and `no_data` artifacts preserved

Step 5 scope is now fixed:

- replay-only local validation
- CI-enforced replay artifact regeneration
- explicit `live` vs `replay` artifact provenance
- synthetic API smoke deferred unless replay proves insufficient

Step 5 is complete in this branch:

- checked-in replay fixtures regenerate the standard artifact layout
- replay validation is CI-enforced
- raw live traces are preserved as captured bytes when present
- local methodology validation is complete; real benchmark evidence still requires the work machine

Do not jump ahead to broad feature removals before Step 5 local validation and the first real baseline run exist.

## Final Recommendation Bias

Default bias is:

- maintain as a thin fork or patch stack

Only justify deeper divergence if benchmark data shows the shallow path cannot hit the target.
