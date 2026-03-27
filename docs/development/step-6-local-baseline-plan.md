# Step 6: Local Disposable Baseline Plan

## Objective

Establish the first live benchmark baseline for K9s Neo on this personal machine
using a disposable local Kubernetes cluster.

This step exists to turn the current benchmark infrastructure into a real
before-and-after engineering loop without taking risk on real clusters while the
fork is still mutation-capable.

## Environment Decision

The baseline environment for Step 6 is:

- `minikube` for the local cluster
- `vfkit` as the preferred macOS VM driver
- a disposable profile dedicated to K9s Neo benchmarking
- namespace-scoped benchmarks first

This is the primary next-step path.
Real-cluster validation is optional later work, not a prerequisite to keep moving.

## Reference Docs

Use these as the primary external references for the local lab:

- [Install kubectl on macOS](https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/)
- [minikube drivers](https://minikube.sigs.k8s.io/docs/drivers/)
- [minikube vfkit driver](https://minikube.sigs.k8s.io/docs/drivers/vfkit/)
- [Agones on minikube](https://agones.dev/site/docs/installation/creating-cluster/minikube/)

## Local Constraints

This machine is a constrained but workable local benchmark lab:

- macOS Apple Silicon
- hardware virtualization available
- 8 GiB RAM
- no existing Docker runtime requirement

Those constraints drive the execution plan:

- start with a single-node profile
- do not chase production-like node scale locally
- focus first on pod-heavy and startup-heavy scenarios
- add multi-node only if node-path scenarios justify the extra cost

## Step 6 Deliverables

The step is complete when we have all of the following:

- a documented local cluster setup for this machine
- a disposable local kube context dedicated to benchmarking
- a namespace with enough seeded objects to exercise the pod list, filter, YAML,
  describe, and node-view flows meaningfully
- raw live benchmark artifacts from `hack/bench/run.py`
- a baseline summary report that clearly states the environment is local minikube
- an updated decision matrix seeded with real local measurements

## Non-Goals

Step 6 does not attempt to:

- prove production-cluster equivalence
- harden the read-only boundary yet
- land performance cuts before the first local live baseline exists
- install Agones before the core baseline is captured
- justify deep data-path divergence before shallow changes are measured

## Execution Plan

### Phase 0: Tooling And Cluster Bootstrap

Install only the tooling required for the local lab:

- `kubectl`
- `minikube`
- `vfkit`
- `vmnet-helper` only if we later need a multi-node profile

Keep the install boring and reversible.
Do not add Docker Desktop unless the preferred path fails badly.

### Phase 1: Disposable Cluster Bring-Up

Create one dedicated local profile for the first baseline.

Guardrails:

- single-node first
- local-only kube context
- separate benchmark namespace
- no all-namespaces baseline by default

Acceptance gate:

- `kubectl` can reach the cluster
- the K9s Neo binary can start against the local context
- the harness can complete at least one `pods_startup` run

### Phase 2: Seed Benchmark Data

Seed enough local workload data to make the benchmark scenarios useful.

Initial focus:

- pod-heavy namespace
- multiple lightweight workloads
- predictable object names for filter scenarios
- enough YAML and describe content to exercise detail flows
- enough node metadata to exercise the nodes view, even if node scale stays small

Defer Agones installation here unless the core pod and node flows prove insufficient.

Acceptance gate:

- `pods_startup`, `pods_filter_settle`, `nodes_first_render`, `pod_yaml`, and
  `pod_describe` all complete locally
- `pod_events` is optional if the seeded namespace does not naturally produce useful events yet

### Phase 3: Capture The First Local Live Baseline

Run the benchmark harness against the disposable local cluster and collect:

- cold and warm runs
- raw JSON artifacts
- CSV summary
- markdown report

The report must say clearly that the environment is local minikube on this machine.

Acceptance gate:

- raw artifacts are complete and readable
- lifecycle markers and request traces line up with the expected scenarios
- results are stable enough to compare future code changes on the same setup

### Phase 4: Start Measuring Shallow Wins

Only after Phase 3 should we begin product changes.

Priority order remains:

1. lazy metrics by default
2. disable node pod counting by default
3. narrow discovery with static core aliases plus Agones allowlist
4. strict read-only hardening as a separate workstream

Each change should be benchmarked individually against the same local setup before
we accept it as real progress.

## Decision Gates

Gate A:
Local cluster is running and the harness completes at least one live scenario.

Gate B:
Seeded data is rich enough to make the primary scenarios meaningful.

Gate C:
The first local live baseline exists with raw artifacts and a readable summary.

Gate D:
Only measured shallow changes proceed into implementation.

## Risks And Mitigations

Risk:
The laptop cannot represent production-scale node counts.

Mitigation:
Use local results for engineering comparisons and broken-path detection, not as a
proxy for full production equivalence.

Risk:
The current branch is still mutation-capable.

Mitigation:
Use a disposable local cluster only.

Risk:
Agones local setup adds too much early complexity.

Mitigation:
Baseline core pod and node flows first, then add local Agones only if it changes the
decision matrix materially.

## Expected Output

At the end of Step 6 we should be able to say:

- the live benchmark harness works end to end against a real local Kubernetes API
- we have the first comparable live baseline for K9s Neo
- the next shallow product change can be measured immediately on the same machine
