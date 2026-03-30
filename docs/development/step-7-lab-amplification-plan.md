# Step 7 Lab Amplification Plan

## Purpose

Step 7A earned a real local win on the frozen single-node control lab.

The next two follow-on probes did not earn promotion on that same control:

- Step 7B static-core registry smoke stayed functional but did not remove the
  remaining startup `GET /api` and `GET /apis` requests before first useful row
- the ambient-metrics diagnostic probe did remove those requests, but it did not
  produce a visible timing win on the current control lab

That means the next serious Step 7 move is not another shallow startup cut on
the same tiny lab. The next move is to amplify the local lab so the remaining
hypotheses are meaningfully measurable.

## Current State

### Decision-grade evidence

- Step 6 local baseline exists:
  - `docs/development/step-6-local-baseline-note.md`
  - `artifacts/bench/20260326-222409/local-baseline-v1`
- Step 7A is a real promoted win:
  - `docs/development/step-7a-small-discovery-cuts-note.md`
  - `artifacts/bench/20260326-235134/step7a-discovery-smallcuts-v1`

### Diagnostic but non-promoted probes

- Step 7B static-core smoke:
  - `artifacts/bench/20260327-201915/step7b2-static-core-smoke`
  - result: did not earn promotion on the current control lab
- ambient-metrics diagnostic smoke:
  - `artifacts/bench/20260327-204553/step7c-ambient-metrics-smoke`
  - `artifacts/bench/20260327-204655/step7c-ambient-plus-7a-smoke`
  - result: identified the real source of the remaining startup `/api` and
    `/apis` requests, but did not earn promotion on the current control lab

### New metrics-enabled control

- `metrics-small` Step 7A control:
  - `docs/development/step-7a-metrics-small-control-note.md`
  - `artifacts/bench/20260328-200943/step7a-metrics-small-control-v1`
  - result: the amplified profile is reproducible and shows real
    `metrics.k8s.io` traffic on the hot path before first useful row

### Metrics-enabled A/B result

- ambient-off candidate:
  - `docs/development/step-7a-metrics-small-ambient-off-note.md`
  - `artifacts/bench/20260329-124834/step7a-metrics-small-ambient-off-v1`
  - result: request cleanup is real and large, but the primary warm-median wins
    stayed below the `>=5%` promotion bar on this machine

## What This Means

- keep the current Step 6 control lab intact
- keep Step 7A as the current promoted performance change
- do not promote Step 7B or the ambient-metrics probe from smoke artifacts
- increase the discriminating power of the local lab before making more shallow
  performance claims

## Fixed Control Profile

Preserve the current control profile exactly as-is:

- profile name: `control-small`
- implementation path: `colima` plus `minikube --driver=docker`
- topology: single-node
- benchmark namespace: `neo-bench`
- current seed: `hack/local-lab/manifests/neo-bench.yaml`

Use this profile for:

- smoke checks
- regression checks
- Step 6 and Step 7A comparability

Do not overwrite this control with metrics-server, extra nodes, or larger seed
data. Add new profiles instead.

## Next Profiles

### Profile 1: `metrics-small`

Goal:
make the metrics hypothesis measurable on this machine without changing topology
at the same time

Shape:

- separate minikube profile on the proven `control-small` colima backend
- same single-node topology as `control-small`
- same namespace-scoped required scenario set
- metrics-server enabled

Primary questions:

- is lazy or ambient metrics work a real visible win when a metrics API is
  actually present?
- do startup and list-view hot paths pay a measurable metrics cost?

Primary scenarios:

- `pods_startup`
- `pods_filter_settle`
- `nodes_first_render`
- `pod_yaml`
- `pod_describe`

Acceptance gate:

- the required scenario set stays green
- traced runs show real metrics-related request families
- the lab is reproducible from repo-owned scripts

Initial repo-owned entrypoints:

- `hack/local-lab/start-metrics-small.sh`
- `hack/local-lab/seed-metrics-small.sh`
- `hack/local-lab/write-vars-metrics-small.sh`
- `hack/local-lab/smoke-step7a-metrics-small.sh`
- `hack/local-lab/capture-step7a-metrics-small.sh`
- `hack/local-lab/write-step7a-metrics-small-note.sh`

### Profile 2: `nodes-small`

Goal:
make node-path cost measurable enough to judge node pod counting and later node
drill-down work

Shape:

- separate minikube profile
- two nodes to start, not more
- deliberate pod skew so one node is much hotter than the other

Primary questions:

- does node pod counting become meaningfully visible?
- does `nodes_first_render` or later `node_pod_drilldown` show a clear target?

Primary scenarios:

- `nodes_first_render`
- optional `node_pod_drilldown` once the seed is ready

Acceptance gate:

- node view remains functionally correct
- traces and request counts clearly expose node-path costs
- the added topology is still boring to recreate

### Optional Later Profile: `pods-large`

Goal:
increase list hydration and filter discrimination only if later work needs it

Shape:

- same topology as the control unless a node-focused profile is already active
- larger pod count in one namespace

Use only when:

- pod list hydration or filter-settle becomes the active target
- the smaller profiles no longer separate candidate changes well enough

## Explicit Not-Now List

Do not do these before the profile split exists:

- no more startup micro-probes on the frozen control lab
- no promotion of Step 7B or the ambient-metrics probe
- no real work-cluster dependency
- no large multi-node local cluster
- no synthetic API work
- no network-latency injection
- no full Agones install just to create performance signal

## Recommended Execution Order

1. preserve `control-small` as the frozen reference profile
2. add `metrics-small`
3. rerun the promoted Step 7A control on `metrics-small`
4. keep the `metrics-small` ambient-off result as a non-promoted cleanup data point
5. add `nodes-small`
6. capture the narrow `nodes_first_render` control on `nodes-small`
7. capture `node_pod_drilldown` on `nodes-small`
8. only if that drilldown path is both hotter and attributable, start a node
   optimization patch
9. only revisit broader discovery or static registry work if the amplified lab
   still shows it as a top bottleneck

## Current Recommendation On The Ambient-Metrics Probe

The ambient-metrics patch has now been measured against the `metrics-small`
Step 7A control.

It earned a good technical result:

- real `metrics.k8s.io` hot-path traffic disappeared before first useful row
- request count and bytes dropped sharply

But it still did not earn promotion as a primary user-visible win on this
machine.

## Definition Of Done

This lab-amplification step is complete when:

- `control-small` remains comparable to the Step 6 and Step 7A artifacts
- `metrics-small` exists and is reproducible
- `nodes-small` exists and is reproducible
- the initial `nodes-small` Step 7A control exists and is documented
- the `nodes-small` drilldown evidence step exists and is documented
- the paired `nodes-small` characterization artifact exists and is documented
- the required scenario set has been rerun on the relevant amplified profiles
- the next Step 7 implementation target is chosen from the amplified evidence,
  not from the stale pre-plateau plan

## Current Nodes-Small Read

The `nodes-small` control now exists:

- note: `docs/development/step-7a-nodes-small-control-note.md`
- artifact root: `artifacts/bench/20260329-153637/step7a-nodes-small-control-v1`

The two-node skewed seed is real and reproducible, but `nodes_first_render`
stayed almost flat versus the earlier Step 7A control on this machine.

The follow-on drilldown artifact now also exists:

- note: `docs/development/step-7d-node-pod-drilldown-nodes-small-note.md`
- artifact root: `artifacts/bench/20260329-162922/step7d-node-pod-drilldown-nodes-small-v1`

What it proved:

- `node_pod_drilldown` is materially hotter than `nodes_first_render` on this
  machine
- but the pre-cutoff request, byte, and object shape stayed effectively flat
- and the representative warm run did not show pre-cutoff `pods:list`

The paired characterization artifact now also exists:

- note: `docs/development/step-7e-node-path-characterization-note.md`
- artifact root: `artifacts/bench/20260329-204049/step7e-node-path-characterization-nodes-small-v1`

That tighter characterization proved:

- the shared node-entry phases are almost flat
- the stable extra cost lives in the drilldown-only tail
- the only stable drilldown-only request family before the terminal Pod useful
  row is a cluster-scope `pods:watch`

That means the next meaningful node-path step is:

- inspect and repair the node-to-pod drilldown hydration path, not a node
  pod counting patch yet

On this laptop, keep only one minikube profile active at a time on the shared
`k9s-neo` Colima backend when using `nodes-small`.
