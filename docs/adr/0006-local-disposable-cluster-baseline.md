# ADR 0006: Local Disposable Cluster Baseline

- Status: Accepted
- Date: 2026-03-26

## Context

Steps 1 through 5 established the local toolchain, product contracts, trace schema,
lifecycle markers, live benchmark harness, and replay-only harness validation.

The original roadmap assumed the first live baseline would run on a work machine
against real clusters. That assumption is no longer the preferred path.

Two facts now matter more than the original plan:

- the current branch is still the measurement baseline and is not yet defensibly
  read-only by construction
- the live harness only needs kubeconfig, context, and a benchmark namespace, so a
  disposable local cluster can exercise the real binary and the real benchmark path

We want to continue gathering live evidence without taking unnecessary risk on real
clusters while mutation-capable paths still exist in the codebase.

## Decision

Step 6 will use a disposable local cluster as the primary live benchmark environment.

For the current macOS Apple Silicon development machine, the preferred setup is:

- `minikube` as the local Kubernetes distribution
- `vfkit` as the preferred macOS VM driver
- single-node startup as the default baseline posture
- namespace-scoped benchmark runs as the default measurement posture

Multi-node local testing is allowed later, but only for targeted node-view and
node-drill-down scenarios after the single-node baseline is working.

Local live results are valid for engineering decisions in this repo, including:

- before-and-after comparisons on the same local environment
- proving that a path is broken or made to work
- validating benchmark harness behavior against a live Kubernetes API
- ranking shallow candidate changes on this machine

Local live results must be described honestly as local disposable-cluster results.
They are not automatically equivalent to production-cluster evidence.

Real-cluster validation becomes optional later work. It is only required if:

- local results are too ambiguous to drive product decisions
- a claimed win depends heavily on auth, RBAC, network, CRD, or cluster-version
  behavior that the local environment cannot represent
- we decide a public claim needs stronger external validity

Agones remains first-class in the product contract, but local Agones installation is
not a prerequisite for the first live baseline. The initial local baseline should
focus on the core hot paths first.

## Consequences

Positive:

- We can continue the project without touching real clusters.
- Live measurements now happen in a disposable environment that matches the current
  safety state of the codebase.
- We preserve the boring, local feedback loop for instrumented before-and-after work.

Negative:

- Local node scale and cluster behavior will not match large production clusters.
- Some later claims may still need stronger validation outside the local lab.
- We must be explicit in docs and benchmark reports about what local results do and
  do not prove.

## Deferred Questions

- Whether we need a second multi-node local profile after the single-node baseline.
- Whether Agones should be installed locally before or after the first shallow
  performance wins land.
- Whether live artifact metadata should explicitly distinguish local live from later
  remote live runs beyond the current `source_kind=live` marker.
