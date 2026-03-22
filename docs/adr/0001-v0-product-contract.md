# 0001 V0 Product Contract

- Status: Accepted
- Date: 2026-03-22

## Context

K9s Neo is intended to be a narrower fork of K9s for high-cardinality SRE triage, not a general-purpose Kubernetes terminal. The project needs a stable product contract before instrumentation and performance work begin, otherwise later benchmarks and cuts will be judged against moving goals.

The contract also needs to reflect two constraints that differ from upstream:

- the final product must be defensibly read-only
- Agones is important enough to support explicitly, but broad CRD discovery is not acceptable on the startup hot path

This ADR defines the v0 target behavior. Early implementation may temporarily lag the contract while the fork is being hardened, but benchmarks, cuts, and later ADRs must align with this document.

## Decision

### Target User

K9s Neo v0 is for SREs and operators triaging very large Kubernetes clusters where pod and node cardinality are high enough that broad discovery, ambient metrics, and cluster-wide scans become costly or fragile.

### Preserved Workflows

V0 must preserve this triage loop:

- enter a scoped view fast
- find a pod or workload fast
- filter and search fast
- inspect logs
- inspect events
- inspect YAML
- inspect describe output
- drill into the selected object only when needed

### Non-Goals

V0 does not attempt to:

- replace all of K9s for all users
- preserve mutation workflows
- optimize cosmetics ahead of hot-path performance and safety
- support generic CRDs broadly on the startup hot path
- justify a rewrite before measured evidence says surgical changes are insufficient

### Default Scope And Navigation

V0 is optimized for scoped triage, not broad ambient visibility.

- namespace-scoped startup is the default posture
- all-namespaces is an explicit operator choice, not the default benchmark target
- the active view takes priority over background breadth

### Supported Resources In V0

The supported v0 resource set is intentionally narrow:

- pods
- nodes
- namespaces
- events
- deployments
- replicasets
- daemonsets
- statefulsets
- jobs
- cronjobs
- services

For these resources, v0 supports the preserved workflows above, including selected-object YAML and describe. Logs remain a pod-first workflow.

### CRD Policy

Generic CRDs are not part of the default startup contract.

- generic CRD discovery is off the startup hot path
- explicit allowlist CRDs are supported
- broad CRD discovery is opt-in or deferred work, not ambient startup behavior

Agones is first-class in v0 through an explicit allowlist:

- `gameservers.agones.dev`
- `fleets.agones.dev`
- `gameserverallocations.allocation.agones.dev`
- `fleetautoscalers.autoscaling.agones.dev`

Agones support should use curated aliases and explicit resource handling rather than broad CRD-first behavior.

### Metrics Policy

Metrics are optional to triage, not a prerequisite for first useful render.

- metrics must not be required for startup or first useful row
- ambient pod and node metrics are not part of the default list-view hot path
- metrics may be exposed in a dedicated mode or through selected-row detail after the primary triage view is usable

### Watch Policy

Watch behavior should serve the active triage loop and avoid ambient breadth.

- only watches that directly support the active view or selected-object detail are in scope for the hot path
- optional feature watches are not part of the default startup contract
- background activity should not be justified by convenience alone

### Read-Only Guarantee

The detailed safety contract lives in ADR 0002. V0 is not complete until the read-only guarantee is enforced in code and backed by tests, not just hidden behind UI affordances.

### Upstream Rebase Strategy

The default maintenance model is a thin fork or downstream patch stack.

- prefer small, isolated patches
- prefer flags and kill switches before hard deletions
- only accept deeper divergence when benchmark data shows the shallow path is insufficient

### Public Build-In-The-Open Hygiene

Public artifacts must remain safe to share.

- separate real-cluster measurements from synthetic or replay fixtures
- sanitize cluster names, namespaces, workload names, logs, and screenshots
- publish code, ADRs, and sanitized benchmark artifacts only

## Consequences

- The product is narrower than upstream by design.
- Agones support is explicit and supported, but generic CRD breadth is not promised.
- Later performance work is judged against a stable triage contract rather than upstream feature parity.
- Benchmark scope stays focused on workflows that matter to the target user.

## Deferred Questions

- Whether additional secondary resources such as Ingresses or HPAs belong in a later v0.x release.
- The exact implementation of metadata-first rendering, server-side Table support, chunking, or improved list-watch behavior.
- Packaging and distribution details beyond normal development builds.
