# Step 7 Plan: Measured Shallow Wins On The Local Baseline

## Purpose

This document defines the Step 7 execution order after the first stable local
baseline was captured in Step 6.

It is intentionally evidence-led:

- use the captured local baseline as the control
- change one thing at a time
- keep kill switches until a change earns permanence
- optimize for the highest-likelihood visible win first
- do not spend the first shallow-win slot on a change whose measured effect is
  already modest on the control environment

## Inputs

Primary local control artifacts:

- `docs/development/step-6-local-baseline-note.md`
- `docs/development/step-6-local-decision-matrix.md`
- `docs/development/step-6-closeout-step-7-entry-checklist.md`
- `artifacts/bench/20260326-222409/local-baseline-v1`

## What Changed After The Baseline

The original hypothesis order was:

1. lazy metrics
2. discovery and CRD breadth
3. node-path work
4. RBAC preflight fan-out

The local baseline changes the execution order on this machine:

- discovery and RBAC preflight fan-out are directly visible right now
- ambient metrics are not directly measurable right now because the local cluster
  has no metrics API
- node pod counting is still real code debt, but this single-node lab does not
  stress its worst-case behavior yet

That means Step 7 should optimize for the best measurable next PR, not the oldest
hypothesis ordering.

## Step 7 Execution Order

### PR 7A: Trim Hot-Path Discovery Breadth With Small Cuts First

#### Why first

- discovery is clearly on the hot path
- but the cleanest rebase-friendly version is likely a sequence of small cuts, not
  an immediate static registry rewrite
- this is the strongest currently measured candidate for a visible startup win on
  this machine

#### Goal

Remove obvious startup discovery work that is not required for the current local
baseline scenarios.

#### First small cuts to try

1. skip CRD augmentation on the startup hot path behind a runtime flag
2. avoid namespace list and its RBAC preflight on explicit namespace startup paths

#### Scope

- `internal/dao/registry.go`
- `internal/client/client.go`
- likely `cmd/root.go` for runtime-only flags
- tests in `internal/dao/` and `internal/client/`

#### Suggested flags

- `--perf-skip-crd-augment`
- `--perf-skip-namespace-validation`

#### Benchmark set

- `pods_startup`
- `pods_filter_settle`
- `nodes_first_render`
- `pod_yaml`
- `pod_describe`

#### Success criteria

- request count drops materially on startup paths
- `customresourcedefinitions:watch` is removed when the CRD cut is enabled
- namespace list and its related SAR disappear on explicit namespace runs when
  the namespace cut is enabled
- no correctness regressions on the required Step 6 scenario set

#### Rollback criteria

- namespace switching or scoped startup correctness breaks
- the cuts only move noise around without measurable benefit

#### Status

Step 7A is now implemented and measured locally behind runtime-only flags:

- `--perf-skip-crd-augment`
- `--perf-skip-namespace-validation`

Measured result note:

- `docs/development/step-7a-small-discovery-cuts-note.md`
- artifact root: `artifacts/bench/20260326-235134/step7a-discovery-smallcuts-v1`

Warm-run medians improved by a visible but bounded amount:

- `pods_startup` first useful row: `1172.81 ms -> 1069.86 ms` (`-8.8%`)
- `pods_filter_settle` first useful row: `1166.05 ms -> 1051.60 ms` (`-9.8%`)
- `nodes_first_render` first useful row: `1247.31 ms -> 1151.17 ms` (`-7.7%`)

Request shape on `pods_startup` before first useful row also narrowed:

- removed `customresourcedefinitions:watch`
- removed `namespaces:list`
- reduced watches from `3 -> 2`

Staff read:

- keep these cuts as measured runtime-only probes for now
- discovery is still clearly a top startup cost after Step 7A
- Step 7B is now the correct next move

### PR 7B: Escalate To Static Core Registry Plus Agones Allowlist Only If Needed

#### Why second

- this is still a strong candidate
- but it has more divergence cost than the smaller discovery cuts
- we should only take it if the smaller discovery cuts are not enough

#### Goal

Replace broad startup discovery for the hot path with a curated static set of core
resource metas plus explicit Agones allowlist handling.

#### Scope

- `internal/dao/registry.go`
- `internal/client/gvr.go`
- any command or alias lookup wiring needed to preserve the supported triage loop
- tests around meta lookup and supported resources

#### Entry condition

The entry condition is now met.

Step 7A yielded a real local win, but it did not exhaust discovery as the top
startup cost.

#### Benchmark set

- same as PR 7B

#### Success criteria

- startup request shape is materially narrower
- the supported v0 resource set still works
- no generic CRD hot-path dependency remains by default

### PR 7C: Node Pod Counting Reduction

#### Why third

- still likely important in real clusters
- but this local lab does not stress it enough to justify making it the first move

#### Goal

Disable node pod counting by default behind a measurable switch or config change.

#### Scope

- `internal/view/node.go`
- `internal/dao/node.go`
- `internal/config/k9s.go`
- related tests

#### Benchmark set

- `nodes_first_render`
- optional `node_pod_drilldown` after we decide whether to widen the local lab

#### Success criteria

- node view request and watch behavior stays correct
- local timings do not regress
- code path becomes simpler and safer for larger future environments

### PR 7D: Metrics Branch Decision

#### Why fourth

- the current local cluster has no metrics API
- the baseline cannot prove the value of lazy metrics yet

#### Decision point

After discovery work has settled:

- if the measured wins are already strong enough, keep moving without local metrics
- if we still need to validate the metrics hypothesis, explicitly decide whether to
  enable metrics-server locally

#### Recommendation

Do not enable local metrics-server before the discovery work has been measured.

### PR 7E: Read-Only RBAC Preflight Reduction

#### Why demoted

- it is real, measurable, and low risk
- but on this local lab the direct measured effect is modest:
  - `pods_startup`: about `5.5 ms` median total SAR duration before first useful row
  - `nodes_first_render`: about `10.6 ms` median total SAR duration before first useful row
- that does not meet the bar for the first shallow-win slot in this project

#### Goal

Remove `SelfSubjectAccessReview` preflight for read-only `get/list/watch` informer
access behind a runtime-only kill switch.

#### Scope

- `internal/watch/factory.go`
- `internal/client/config.go`
- likely `cmd/root.go` for runtime-only flag wiring
- tests in `internal/watch/` and `internal/client/`

#### Intended shape

- keep current behavior by default
- add a runtime-only flag for measurement, for example:
  - `--perf-skip-read-preflight`
- when enabled:
  - `CanForResource` and `CanForInstance` bypass `CanI` for read-only verbs only
  - mutation-capable paths stay unchanged

#### Recommendation

Do not spend the first shallow-win slot on this.
Only pick it up if:

- discovery work is already done and we still want to reduce request fan-out
- or we decide the failure-surface reduction is worth it as a secondary cleanup

## Separate Safety Track

Strict read-only hardening remains required, but it should not be the first
performance PR.

Keep it as a separate workstream:

- block mutation-capable actions by construction
- remove plugin loading and execution
- remove shell, exec, attach, port-forward, helper/debug pod paths
- justify it on safety and maintenance reduction, not on speed

## Benchmark Protocol For Every Step 7 PR

For every PR:

1. keep the Step 6 local baseline as the control
2. enable only the new change under test
3. rerun:
   - `pods_startup`
   - `pods_filter_settle`
   - `nodes_first_render`
   - plus any scenario directly touched by the change
4. compare:
   - first useful row
   - stable interactive
   - request count
   - response bytes before first useful row
   - watch count
   - status stability
5. update the decision matrix immediately

## Stop Rules

Stop and reassess if any of these happen:

- the first two shallow PRs both produce negligible wins
- a smaller discovery cut already solves most of the visible startup breadth
- the local lab becomes too unrepresentative for the remaining hypothesis

Step 7A did not trigger a stop rule. It validated the direction and justified
continuing discovery work at the larger static-core layer.
- safety hardening starts to conflict with fair baseline comparisons

## Step 7 Recommendation

The next implementation order on this machine should be:

1. PR 7A: small discovery cuts
2. PR 7B only if PR 7A is not enough
3. PR 7C node pod counting after that
4. PR 7D metrics decision later, not first
5. PR 7E RBAC preflight reduction only as a secondary or cleanup change

That is the highest-signal, lowest-regret Step 7 path for the current local lab.
