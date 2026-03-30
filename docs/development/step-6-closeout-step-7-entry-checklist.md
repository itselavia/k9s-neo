# Step 6 Closeout And Step 7 Entry Checklist

## Purpose

This document is the execution checklist for closing Step 6 and entering Step 7.

Use it as the operational companion to:

- `docs/adr/0006-local-disposable-cluster-baseline.md`
- `docs/development/step-6-local-baseline-plan.md`

This checklist is intentionally narrow:

- finish and freeze the first comparable local live baseline
- do not blur baseline capture with product changes
- enter Step 7 only when the local benchmark loop is boring and repeatable

## Fixed Control Environment

Until the first shallow wins are measured, treat this environment as fixed:

- cluster path: `colima` plus `minikube --driver=docker`
- minikube profile: `k9s-neo`
- benchmark namespace: `neo-bench`
- kubeconfig path: `~/.k9s-neo-state/kubeconfig`
- minikube defaults from `hack/local-lab/start-cluster.sh`:
  - minikube CPUs: `2`
  - minikube memory: `2048 MiB`
  - minikube disk: `20g`
  - container runtime: `containerd`
- colima defaults from `hack/local-lab/start-cluster.sh`:
  - colima CPUs: `2`
  - colima memory: `3 GiB`
  - colima disk: `20 GiB`
- seeded manifest: `hack/local-lab/manifests/neo-bench.yaml`

Do not change any of the above between baseline runs and the first shallow-change
reruns unless the current setup is proven unusable. If it must change, restart the
baseline from scratch.

## Required Scenario Set

The first stable baseline must cover these scenarios:

- `pods_startup`
- `pods_filter_settle`
- `nodes_first_render`
- `pod_yaml`
- `pod_describe`

Optional follow-on scenarios:

- `pod_logs`
- `node_pod_drilldown`

Deferred unless clearly useful:

- `pod_events`
- `all_namespaces_pods`

## Step 6 Exit Criteria

Step 6 is complete only when all of these are true:

- the local lab can be recreated from repo-owned scripts
- the fixed control environment is recorded and unchanged
- the required scenario set passes reliably
- a stable local baseline artifact set exists
- the summary clearly states the environment is local minikube on this machine
- the first decision matrix is seeded from measured local numbers
- no shallow product changes were merged ahead of that baseline

## Phase 1: Freeze The Local Lab

### Checklist

- confirm the local tools install path is still repo-owned and user-scoped
- confirm the local bring-up path is still `colima` plus `minikube --driver=docker`
- confirm the benchmark namespace is still `neo-bench`
- confirm the seed manifest has not drifted unexpectedly
- confirm the K9s Neo binary to benchmark is the current repo build

### Commands

Preferred repo-owned commands:

```bash
./hack/local-lab/install-tools.sh
./hack/local-lab/start-cluster.sh
./hack/local-lab/seed-bench.sh
./hack/local-lab/write-vars.sh
```

### Acceptance

- `kubectl get nodes` works against the `k9s-neo` context
- `kubectl get pods -n neo-bench` shows the seeded workloads
- the repo binary starts under the live harness

## Phase 2: Prove Required Scenarios Are Repeatably Green

### Checklist

- run each required scenario once cold and once warm
- fix harness or seed-data issues before moving on
- do not start the full baseline set while any required scenario is flaky

### Commands

Preferred repo-owned smoke command:

```bash
./hack/local-lab/smoke-required.sh
```

This wrapper runs the required Step 6 scenario set once cold and once warm:

- `pods_startup`
- `pods_filter_settle`
- `nodes_first_render`
- `pod_yaml`
- `pod_describe`

### Acceptance

- every required scenario finishes with `status=ok`
- the trace contains the expected lifecycle markers for the scenario
- the report is readable without manual repair

## Phase 3: Capture The First Stable Local Baseline

### Checklist

- keep the seeded cluster running unchanged
- use one benchmark label for the full baseline run
- run the full required scenario set with `10 cold + 10 warm`
- preserve raw artifacts
- record the artifact root in the baseline note

### Command

```bash
./hack/local-lab/capture-baseline.sh
```

If a required scenario is still unstable, do not downgrade the run counts silently.
Fix the instability first or narrow the required set explicitly in docs.

### Acceptance

- one artifact root contains the full baseline set
- `summary/runs.csv` exists
- `summary/report.md` exists
- the report explicitly reflects a local live environment

## Phase 4: Seed The First Decision Matrix

Use the local baseline report and raw artifacts to create the first measured
control table.

### Required Metrics Per Scenario

- `time_to_first_paint_ms`
- `time_to_first_useful_row_ms`
- `time_to_stable_interactive_ms`
- `filter_settle_ms` when applicable
- `detail_ready_ms` when applicable
- `peak_rss_bytes_30s`
- `peak_cpu_percent_30s`
- `api_request_count`
- `api_request_count_by_resource_verb`
- `total_response_bytes_before_first_useful_row`
- `total_object_count_before_first_useful_row`
- `watch_count`
- `watches_before_first_useful_row`

### Minimum Decision Matrix Columns

- candidate change
- expected impact band
- measured baseline bottleneck it targets
- primary scenario affected
- implementation complexity
- divergence cost
- recommendation status

### Historical Candidate Order

1. small discovery cuts first
2. static core aliases plus Agones allowlist if discovery is still dominant after the small cuts
3. disable node pod counting by default
4. decide later whether local metrics are needed to measure lazy metrics on this machine
5. strict read-only hardening as a separate safety track

This was the correct order at the end of Step 6.

It is no longer the current execution order after the later Step 7 probes.
Use `docs/development/step-7-plan.md` and
`docs/development/step-7-lab-amplification-plan.md` for the current next-step
sequence.

## Step 7 Entry Criteria

Do not start Step 7 until all of these are true:

- Step 6 exit criteria are met
- the local baseline can be rerun on demand without re-debugging the lab
- the required scenario set is boringly repeatable
- request counts, bytes, and watch behavior are attributable enough to judge
  shallow changes
- there is a repo-tracked note pointing to the artifact root of the baseline set

## Step 7 Rules

These rules are mandatory for the first shallow wins:

- change one thing at a time
- keep the control environment fixed
- rerun the same required scenario set after every change
- update the decision matrix immediately after each measurement
- separate safety-driven changes from speed-driven claims
- revert or downgrade any change whose measured effect is negligible

## Explicit Not-Now List

Do not do these before the first local baseline is frozen:

- broad runtime feature deletion
- plugin-runtime removal
- XRay removal
- Pulses removal
- port-forward code removal
- deep CRD architecture changes
- metadata-first refactors
- server-side Table work
- chunking or list-watch redesign

Those may become correct later. They are not the right move before the baseline is
frozen.

## Historical Step 7 Execution Order

The remainder of this section is preserved as the original Step 6 handoff.
Do not treat it as the current Step 7 execution order.

### Change 1: Lazy Metrics By Default

Goal:
remove ambient metrics from startup and list-view hot paths where possible

Why first:

- highest-confidence shallow win
- low divergence
- easy to measure cleanly

### Change 2: Disable Node Pod Counting By Default

Goal:
remove broad pod-scan work from node-view startup paths

Why second:

- likely large effect on node flows
- low implementation complexity

### Change 3: Narrow Discovery With Static Core Aliases Plus Agones Allowlist

Goal:
reduce startup discovery breadth without giving up the v0 CRD policy

Why third:

- plausible high impact
- slightly higher divergence cost

### Change 4: Strict Read-Only Hardening

Goal:
enforce the product safety contract in code, not just intent

Why separate:

- must happen
- justified mainly by safety, not by speed

## Handoff Note

If a new thread picks this up, it should start here:

1. confirm the local lab still matches the fixed control environment
2. confirm the required scenario set is still green
3. capture or verify the stable Step 6 baseline artifact set
4. use `docs/development/step-7-plan.md` for the current Step 7 order
5. use `docs/development/step-7-lab-amplification-plan.md` for the next
   execution step after the current plateau
