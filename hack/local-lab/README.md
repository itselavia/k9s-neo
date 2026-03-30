# Local Lab

This directory contains the repo-owned scripts for the disposable local benchmark lab.

The intended Step 6 execution order is:

1. install tools
2. start the local cluster
3. seed the benchmark namespace
4. run the required Step 6 smoke set
5. capture the first stable local baseline

The Step 7 lab-amplification profiles are:

- separate minikube profile on the proven `k9s-neo` colima backend
- `metrics-small`: same single-node topology with `metrics-server` enabled
- `nodes-small`: two-node topology with a deterministic hot worker node

## Scripts

- `install-tools.sh`: install `kubectl`, `minikube`, `colima`, `lima`, and Docker CLI into user-scoped local paths
- `start-cluster.sh`: start the `k9s-neo` colima and minikube profile
- `start-metrics-small.sh`: start the `k9s-neo-metrics-small` profile and enable `metrics-server`
- `start-nodes-small.sh`: start the `k9s-neo-nodes-small` profile with two minikube nodes
- `seed-bench.sh`: apply the `neo-bench` manifest and wait for rollouts
- `seed-metrics-small.sh`: seed the benchmark namespace on the `metrics-small` profile
- `seed-nodes-small.sh`: seed the benchmark namespace on the `nodes-small` profile and verify hot-node skew
- `k9s-step7a.sh`: run the built binary with only the promoted Step 7A flags
- `write-vars.sh`: write `hack/bench/vars.local.json` for the local profile
- `write-vars-metrics-small.sh`: write `hack/bench/vars.metrics-small.json` for the `metrics-small` profile
- `write-vars-nodes-small.sh`: write `hack/bench/vars.nodes-small.json` for the `nodes-small` profile
- `bench-nodes-small.sh`: run the narrow Step 7A node control on the `nodes-small` profile
- `bench-node-path-characterization-nodes-small.sh`: run the paired Step 7E node-path characterization scenarios on the `nodes-small` profile
- `smoke-required.sh`: run the required Step 6 scenario set once cold and once warm
- `smoke-step7a-metrics-small.sh`: run the Step 7A control smoke on `metrics-small`
- `smoke-step7a-nodes-small.sh`: run the Step 7A control smoke on `nodes-small`
- `smoke-step7e-node-path-characterization-nodes-small.sh`: run the Step 7E node-path characterization smoke on `nodes-small`
- `capture-baseline.sh`: run the required Step 6 scenario set with baseline run counts and write a baseline note
- `capture-step7a-metrics-small.sh`: run the Step 7A control capture on `metrics-small` with `10 + 10` counts and write the control note
- `capture-step7a-nodes-small.sh`: run the Step 7A control capture on `nodes-small`
- `capture-step7e-node-path-characterization-nodes-small.sh`: run the Step 7E node-path characterization capture on `nodes-small`
- `write-baseline-note.sh`: write a repo-tracked baseline note for an existing artifact root
- `write-step7a-metrics-small-note.sh`: write the repo-tracked Step 7A control note for `metrics-small`
- `write-step7a-nodes-small-note.sh`: write the repo-tracked Step 7A control note for `nodes-small`
- `write-step7e-node-path-characterization-note.sh`: write the repo-tracked Step 7E characterization note for `nodes-small`
- `delete-cluster.sh`: delete the local colima and minikube profile
- `delete-metrics-small.sh`: delete the `metrics-small` profile
- `delete-nodes-small.sh`: delete the `nodes-small` profile

## Typical Usage

Install tools:

```bash
./hack/local-lab/install-tools.sh
```

Start and seed the lab:

```bash
./hack/local-lab/start-cluster.sh
./hack/local-lab/seed-bench.sh
./hack/local-lab/write-vars.sh
```

Run the required Step 6 smoke set:

```bash
./hack/local-lab/smoke-required.sh
```

Capture the first stable baseline and write the note:

```bash
./hack/local-lab/capture-baseline.sh
```

Start and smoke the `metrics-small` profile:

```bash
./hack/local-lab/start-metrics-small.sh
./hack/local-lab/seed-metrics-small.sh
./hack/local-lab/write-vars-metrics-small.sh
./hack/local-lab/smoke-step7a-metrics-small.sh
./hack/local-lab/capture-step7a-metrics-small.sh
```

Start and capture the `nodes-small` control:

```bash
./hack/local-lab/start-nodes-small.sh
./hack/local-lab/seed-nodes-small.sh
./hack/local-lab/write-vars-nodes-small.sh
./hack/local-lab/smoke-step7a-nodes-small.sh
./hack/local-lab/capture-step7a-nodes-small.sh
```

The initial `nodes-small` control intentionally measured only `nodes_first_render`.
That control is historically paired with a measured Step 7D drilldown artifact,
but the maintained node-path workflow is now the paired Step 7E characterization.

Run the paired `nodes-small` characterization step:

```bash
./hack/local-lab/smoke-step7e-node-path-characterization-nodes-small.sh
./hack/local-lab/capture-step7e-node-path-characterization-nodes-small.sh
```

Current Step 7E result on this machine:

- the hotter path is concentrated in the drilldown-only tail
- the only stable drilldown-only request family before terminal Pod useful row is a cluster-scope `pods:watch`
- the next node-path step is to inspect or repair the drilldown hydration path, not node pod counting

On this laptop, keep only one minikube profile active at a time on the shared
`k9s-neo` Colima backend when using `nodes-small`. The repo-owned
`start-nodes-small.sh` wrapper now fails fast if other `k9s-neo*` profiles are
still running.

The `metrics-small` cluster is intentionally separate from the original
single-node control, but it reuses the proven `k9s-neo` colima backend by
default. Do not overwrite the control cluster itself with metrics-server.

## Required Scenario Set

These scripts intentionally limit Step 6 to the required baseline scenarios:

- `pods_startup`
- `pods_filter_settle`
- `nodes_first_render`
- `pod_yaml`
- `pod_describe`

Optional scenarios such as `pod_logs`, `node_pod_drilldown`, and `pod_events`
should remain separate until the first required baseline is frozen.
