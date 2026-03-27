# Local Lab

This directory contains the repo-owned scripts for the disposable local benchmark lab.

The intended Step 6 execution order is:

1. install tools
2. start the local cluster
3. seed the benchmark namespace
4. run the required Step 6 smoke set
5. capture the first stable local baseline

## Scripts

- `install-tools.sh`: install `kubectl`, `minikube`, `colima`, `lima`, and Docker CLI into user-scoped local paths
- `start-cluster.sh`: start the `k9s-neo` colima and minikube profile
- `seed-bench.sh`: apply the `neo-bench` manifest and wait for rollouts
- `write-vars.sh`: write `hack/bench/vars.local.json` for the local profile
- `smoke-required.sh`: run the required Step 6 scenario set once cold and once warm
- `capture-baseline.sh`: run the required Step 6 scenario set with baseline run counts and write a baseline note
- `write-baseline-note.sh`: write a repo-tracked baseline note for an existing artifact root
- `delete-cluster.sh`: delete the local colima and minikube profile

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

## Required Scenario Set

These scripts intentionally limit Step 6 to the required baseline scenarios:

- `pods_startup`
- `pods_filter_settle`
- `nodes_first_render`
- `pod_yaml`
- `pod_describe`

Optional scenarios such as `pod_logs`, `node_pod_drilldown`, and `pod_events`
should remain separate until the first required baseline is frozen.
