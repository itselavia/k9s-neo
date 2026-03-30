# Step 7 MacBook Air Scale-Up Handoff

## Purpose

This document is the handoff for moving K9s Neo local benchmarking to a newer
Apple Silicon laptop with materially more memory, such as a `32 GiB` MacBook
Air.

Its job is not to restate the whole project.

Its job is to ensure that the move creates a better local lab rather than a
sloppier story.

## Current Accepted Evidence

These statements are already established on the original machine and should be
treated as priors:

- Step 7A small discovery cuts are the only promoted performance win so far.
  - note: `docs/development/step-7a-small-discovery-cuts-note.md`
- The static-core registry probe is diagnostic, not promoted.
  - see: `docs/development/step-7-plan.md`
- The `metrics-small` ambient-metrics A/B is real cleanup, not a promoted win on
  the original machine.
  - notes:
    - `docs/development/step-7a-metrics-small-control-note.md`
    - `docs/development/step-7a-metrics-small-ambient-off-note.md`
- The `nodes-small` path is now the most important local evidence track.
  - notes:
    - `docs/development/step-7a-nodes-small-control-note.md`
    - `docs/development/step-7d-node-pod-drilldown-nodes-small-note.md`
    - `docs/development/step-7e-node-path-characterization-note.md`
- The current best next code target is the node-to-pod drilldown hydration path,
  not node pod counting.
  - see:
    - `AGENTS.md`
    - `docs/development/step-7-plan.md`
    - `docs/development/step-7e-node-path-characterization-note.md`

## Evidence Boundary

The new laptop is a new lab generation.

That means:

- old-machine artifacts are valid priors for hypothesis selection
- new-machine artifacts are valid controls for new-machine A/B work
- old-machine and new-machine medians must not be merged into one series

Allowed cross-machine use:

- compare request families qualitatively
- compare whether a hotspot still reproduces
- compare whether a candidate remains promoted, parked, or rejected
- compare ordering of bottlenecks

Forbidden cross-machine use:

- claiming a product win from raw old-machine vs new-machine timing deltas
- averaging or ranking mixed-machine medians
- calling a candidate promoted on the new machine without a same-machine control
- treating the new machine as proof that the old machine was wrong

## Goal On The New Machine

Use the extra headroom to create a more discriminating local lab, not just lower
absolute times.

The scale-up should help answer these questions:

1. Does the hotter node drilldown path become more attributable on a larger
   local lab?
2. Does ambient metrics stay below the promotion bar, or does it become a real
   visible win when the lab is larger?
3. Does a denser pod lab make startup, filter, or list hydration separate more
   clearly?

The move is useful only if it changes the decision landscape, not if it merely
produces prettier local medians.

## Required Bootstrap Sequence

Do not skip this sequence.

### Phase 0: Rebuild The Current World

Clone the repo and rebuild using the existing repo-owned path:

```bash
./hack/bootstrap-go.sh
./hack/with-go.sh make build
./execs/k9s version
```

Install the repo-owned local tools:

```bash
./hack/local-lab/install-tools.sh
```

### Phase 1: Recreate The Small Controls Unchanged

Before any scale-up, re-establish the existing lab shapes on the new machine
without changing budgets.

Required:

1. `control-small`
2. `metrics-small`
3. `nodes-small`

This preserves portability and gives the new machine its own small-profile
controls.

Do not compare these raw medians to the old machine as product evidence.

### Phase 2: Re-run The Current Accepted Step 7 Evidence

On the new machine, re-run:

1. Step 7A small discovery control/candidate
2. `metrics-small` Step 7A control
3. `nodes-small` Step 7A control
4. `nodes-small` Step 7E characterization path

Only after these are stable should the new machine be treated as a trustworthy
next-step environment.

### Phase 3: Start The Larger Lab Generation

Once the unchanged profiles are green and stable, scale up one axis at a time:

1. larger single-node control
2. larger single-node metrics profile
3. larger node-stress profile
4. only then a denser pod profile if still needed

Do not change topology, seed density, and feature surface all at once.

For the larger-machine generation, do not overwrite the recreated small
controls. Use a new profile family, for example:

- `K9S_NEO_COLIMA_PROFILE=k9s-neo-air`
- `K9S_NEO_PROFILE=k9s-neo-air-control`
- `K9S_NEO_PROFILE=k9s-neo-air-metrics`
- `K9S_NEO_PROFILE=k9s-neo-air-nodes`
- `K9S_NEO_PROFILE=k9s-neo-air-pods-large`

Keep artifact labels equally explicit, for example:

- `m5air-control-small-v1`
- `m5air-metrics-small-v1`
- `m5air-nodes-small-v1`
- `m5air-control-medium-v1`

Do not let the larger-machine generation silently overwrite the recreated small
profiles.

## Starting Profile Budgets On A 32 GiB Fanless MacBook Air

These are starting points, not new defaults.

Because the machine is fanless, prefer:

- more memory headroom
- denser seeds
- moderate CPU counts

Do not chase maximum CPU allocation first.

### Shared Backend Baseline

For the normal amplified Step 7 profiles on the new machine, standardize the
shared Colima backend first:

```bash
export K9S_NEO_COLIMA_PROFILE=k9s-neo-air
export K9S_NEO_COLIMA_CPUS=4
export K9S_NEO_COLIMA_MEMORY_GB=12
export K9S_NEO_COLIMA_DISK_GB=60
```

That is the best default starting point for a `32 GiB` fanless machine:

- enough memory headroom to keep the host out of the way
- moderate CPU count for sustained repeated runs
- no need to resize the backend profile constantly

### Bootstrap Small Controls

Use a small but not tiny bootstrap profile on the new machine:

```bash
export K9S_NEO_MINIKUBE_CPUS=2
export K9S_NEO_MINIKUBE_MEMORY_MB=4096
export K9S_NEO_MINIKUBE_DISK_SIZE=25g
export K9S_NEO_MINIKUBE_NODES=1
```

This is still a bootstrap control, not yet the amplified single-node profile.

### Amplified Single-Node Control

First meaningful larger single-node starting point:

```bash
K9S_NEO_PROFILE=k9s-neo-air-control \
K9S_NEO_MINIKUBE_CPUS=3 \
K9S_NEO_MINIKUBE_MEMORY_MB=8192 \
K9S_NEO_MINIKUBE_DISK_SIZE=40g \
./hack/local-lab/start-cluster.sh
```

Why this shape:

- enough memory headroom to stop the host from being the first bottleneck
- still conservative for a fanless machine
- keeps the topology unchanged while increasing discriminating power

### Amplified `metrics-small`

Starting point:

```bash
K9S_NEO_PROFILE=k9s-neo-air-metrics \
K9S_NEO_MINIKUBE_CPUS=3 \
K9S_NEO_MINIKUBE_MEMORY_MB=8192 \
K9S_NEO_MINIKUBE_DISK_SIZE=40g \
./hack/local-lab/start-metrics-small.sh
```

Reason for a little more memory here:

- `metrics-server` adds noise on a tiny host
- the goal is to make metrics-related work measurable without the host becoming
  the first limiter

### Amplified `nodes-small`

Starting point:

```bash
K9S_NEO_PROFILE=k9s-neo-air-nodes \
K9S_NEO_MINIKUBE_NODES=2 \
K9S_NEO_MINIKUBE_CPUS=2 \
K9S_NEO_MINIKUBE_MEMORY_MB=4096 \
K9S_NEO_MINIKUBE_DISK_SIZE=35g \
./hack/local-lab/start-nodes-small.sh
```

Why this shape:

- give the shared backend more headroom first
- keep per-node minikube settings moderate
- preserve the existing `nodes-small` topology semantics

If this remains stable and under-stressed, only then consider:

- `K9S_NEO_MINIKUBE_NODES=3`
- denser hot-node seeds

### Possible `pods-large`

Do not add this until:

- the amplified single-node control is stable
- `metrics-small` and `nodes-small` have been rerun on the new machine
- pod startup/filter remains a live question

Starting point:

```bash
K9S_NEO_PROFILE=k9s-neo-air-pods-large \
K9S_NEO_COLIMA_PROFILE=k9s-neo-large \
K9S_NEO_COLIMA_CPUS=6 \
K9S_NEO_COLIMA_MEMORY_GB=16 \
K9S_NEO_COLIMA_DISK_GB=100 \
K9S_NEO_MINIKUBE_CPUS=4 \
K9S_NEO_MINIKUBE_MEMORY_MB=12288 \
K9S_NEO_MINIKUBE_DISK_SIZE=80g \
./hack/local-lab/start-cluster.sh
```

This profile should differ from `control-small` by seed density first, not by
changing the whole world.

Use a separate Colima backend for this profile so the larger pod-density work
does not contaminate the normal shared backend.

## Machine-Specific Discipline

On a fanless Air:

- if sustained runs become thermally unstable, reduce CPU before reducing memory
- if the backend becomes noisy, prefer fewer active profiles and more serialized
  runs
- prefer raising memory before raising CPU when a profile feels tight
- keep only one active minikube profile at a time on the shared Colima backend
  when using `nodes-small`

The extra memory is the main advantage. Use it.

Do not spend it on unnecessary concurrency or overlapping profiles.

## When Scale-Up Has Actually Changed The Decision Landscape

Treat the scale-up as decision-changing only if at least one of these happens on
the new machine:

1. Step 7A no longer reproduces as the clearest promoted shallow win.
2. Ambient metrics crosses the `>=5%` bar and becomes a promoted local win.
3. A larger node-stress profile makes the drilldown hydration path much more
   attributable.
4. A denser pod profile makes startup/filter/list hydration separate much more
   clearly.
5. A previously parked candidate now shows both:
   - visible timing change
   - trace-attributable request-shape change

If none of these happen, the decision landscape did not change. The new machine
just became a nicer development box.

## Stop Rules

Stop scaling up further if any of these happen:

- unchanged small controls are not reproducible on the new machine
- warm medians on an untouched control are too noisy to support decisions
- the host becomes the bottleneck instead of the workload
- the fanless thermal envelope makes long repeated runs unstable
- bigger profiles lower confidence instead of increasing discrimination

Practical stop rule:

- if repeated warm runs on an untouched control differ enough that the project
  `>=5%` promotion bar cannot be trusted, stop scaling further until the lab is
  stabilized
- if a larger profile feels memory-tight, raise minikube memory first
- do not push CPUs past `4` on the fanless machine until memory headroom and run
  stability are both clearly healthy

## Immediate Next Actions On The New Machine

1. clone the repo at this commit
2. rebuild and install tools
3. recreate `control-small` unchanged
4. recreate `metrics-small` unchanged
5. recreate `nodes-small` unchanged
6. only then start the amplified profile ladder
7. if hotspot ranking still points at the node drilldown tail, keep the next
   implementation target as the node-to-pod drilldown hydration path

## Current Recommendation

The new machine is worth using if the intent is to scale the existing lab
generation up with discipline.

It is not worth using merely to rerun the same tiny profiles and declare a new
story from raw timing deltas.
