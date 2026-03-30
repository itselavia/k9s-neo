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

## Status Reset After Step 7A

The original hypothesis order from the Step 6 handoff was:

1. lazy metrics
2. discovery and CRD breadth
3. node-path work
4. RBAC preflight fan-out

The first measured reset on this machine was correct:

- Step 7A small discovery cuts earned a real win
- read-only RBAC preflight stayed measurable but too modest to deserve the first
  shallow-win slot

The second reset is now also required:

- the Step 7B static-core smoke did not earn promotion on the current control
  lab
- the ambient-metrics diagnostic probe found the real source of the remaining
  startup `/api` and `/apis` requests, but it also did not earn promotion on the
  current control lab

That means the current single-node control lab has started to plateau for shallow
startup probes. The next serious Step 7 move is lab amplification, not another
startup micro-cut on the same control environment.

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
- later probes showed that the frozen control lab needed amplification before
  the next Step 7 target could be chosen confidently

### Step 7B: Split The Local Lab Before More Shallow Cuts

#### Why now

- the frozen single-node control lab was good enough to earn Step 7A
- it is no longer discriminating enough to justify more startup micro-probes
- the next best move is to amplify the lab without destroying comparability

#### Goal

Create targeted local profiles that expose the next hidden cost centers without
overwriting the Step 6 and Step 7A control environment.

#### Primary plan

Use `docs/development/step-7-lab-amplification-plan.md` as the execution guide.

Immediate profile order:

1. keep the current `control-small` profile frozen
2. add `metrics-small`:
   - same topology
   - metrics-server enabled
3. add `nodes-small`:
   - two-node topology
   - deliberate pod skew

#### Success criteria

- the original control remains comparable to Step 6 and Step 7A
- the new profiles are reproducible from repo-owned scripts
- the next implementation target is chosen from amplified evidence rather than
  from the stale pre-plateau plan

### Measured Candidate: Ambient Metrics On A Metrics-Enabled Lab

#### Result

The `metrics-small` A/B is now complete:

- control note: `docs/development/step-7a-metrics-small-control-note.md`
- control artifact root: `artifacts/bench/20260328-200943/step7a-metrics-small-control-v1`
- candidate note: `docs/development/step-7a-metrics-small-ambient-off-note.md`
- candidate artifact root: `artifacts/bench/20260329-124834/step7a-metrics-small-ambient-off-v1`

#### What it proved

- the candidate removed the expected request families before first useful row:
  - `/api`
  - `/apis`
  - real `metrics.k8s.io` pod and node requests
- startup request count and pre-cutoff response bytes dropped sharply

#### Why it is not promoted

- `pods_startup` warm first useful row improved only `-2.43%`
- `nodes_first_render` warm first useful row improved only `-4.87%`
- those are real gains, but still below the project `>=5%` primary promotion bar

#### Recommendation

- do not promote `--perf-disable-ambient-metrics` as the next headline Step 7 win
- keep it as a credible cleanup candidate with strong request-shape evidence
- keep it parked as cleanup-only on this machine

### Measured Control: Nodes-Small Step 7A

#### Result

The `nodes-small` Step 7A control is now complete:

- control note: `docs/development/step-7a-nodes-small-control-note.md`
- control artifact root: `artifacts/bench/20260329-153637/step7a-nodes-small-control-v1`

#### What it proved

- the two-node `nodes-small` profile is reproducible from repo-owned scripts
- the benchmark pods land deterministically on the worker node
- the narrow `nodes_first_render` control stays stable on that profile

#### Why it is not enough yet

- `nodes_first_render` stayed almost flat versus the earlier Step 7A control on
  this machine:
  - previous warm first useful row: about `1151.17 ms`
  - `nodes-small` warm first useful row: about `1153.36 ms`
- that means node view render alone still does not expose enough extra cost to
  justify a node optimization patch

#### Recommendation

- do not start a node pod counting code change yet
- use `node_pod_drilldown` on `nodes-small` next
- keep the initial `nodes-small` control as the node-path reference artifact

### Measured Candidate: Nodes-Small Node-Pod Drilldown

#### Result

The `nodes-small` drilldown evidence step is now complete:

- note: `docs/development/step-7d-node-pod-drilldown-nodes-small-note.md`
- candidate artifact root: `artifacts/bench/20260329-162922/step7d-node-pod-drilldown-nodes-small-v1`
- comparison basis:
  - `docs/development/step-7a-nodes-small-control-note.md`
  - `artifacts/bench/20260329-153637/step7a-nodes-small-control-v1`

#### What it proved

- `node_pod_drilldown` is materially hotter than the `nodes-small`
  `nodes_first_render` control on this machine:
  - warm first useful row: about `+17.35%`
  - warm stable interactive: about `+16.40%`
- the matched terminal view is the intended `Pod` view on the hot worker node

#### Why it is still not enough

- pre-cutoff request shape stayed effectively flat versus the control:
  - API requests: flat
  - bytes before first useful row: flat
  - objects before first useful row: flat
- the representative warm run still showed no pre-cutoff `pods:list`
- that means the hotter drilldown path is real, but it is not yet attributable
  enough to justify a node pod counting patch

#### Recommendation

- do not start a node pod counting code change yet
- treat Step 7D as evidence of a hotter node drilldown path, not yet as evidence
  of the specific fix
- Step 7E now provides that characterization:
  - shared node-entry phases are nearly flat
  - the extra cost sits in the drilldown-only tail
  - the only stable drilldown-only request family before terminal Pod useful
    row is a cluster-scope `pods:watch`
- only start a node optimization patch if that attribution points back to
  node-to-pod data work clearly enough

### Measured Characterization: Nodes-Small Node-Path Pair

#### Result

The paired node-path characterization is now complete:

- note: `docs/development/step-7e-node-path-characterization-note.md`
- artifact root: `artifacts/bench/20260329-204049/step7e-node-path-characterization-nodes-small-v1`

#### What it proved

- `node_pod_drilldown` stays materially hotter than `nodes_first_render` on
  this machine:
  - warm first useful row: about `+11.76%`
  - warm stable interactive: about `+11.57%`
- the shared node-entry phases are nearly flat between the two scenarios
- the stable extra cost is concentrated in the drilldown-only tail:
  - Node useful row -> filter settle: about `123 ms`
  - filter settle -> final Pod activate: about `50 ms`
  - final Pod activate -> final Pod useful row: only a few milliseconds

#### Why this changes the next target

- the hotter path is now attributable enough to reject the old node pod counting
  hypothesis on this lab
- the only stable drilldown-only request family before the terminal Pod useful
  row is a cluster-scope `pods:watch`
- pod bytes and pod objects before the terminal Pod useful row stayed at `0`

#### Recommendation

- do not start a node pod counting patch on this machine
- make the next implementation target the node-to-pod drilldown hydration path
- specifically inspect why node drilldown opens a cluster-scope `pods:watch`
  before the terminal Pod useful row

### Future Candidate: Node Pod Counting Reduction

#### Why not yet

- node-path cost remains plausible
- Step 7D proved the drilldown path is hotter, but it still did not show
  decision-grade attribution to the node pod counting hypothesis

#### Recommendation

Keep this plausible but parked, and no longer treat it as the next node-path
patch on this machine.

The next step is not the optimization patch itself. The next step is a
node-to-pod drilldown hydration-path repair or instrumentation pass, using the
Step 7E characterization artifact as the starting point.

### Future Candidate: Static Core Registry Plus Agones Allowlist

#### Why parked

- the static-core smoke was useful diagnostically
- it did not earn promotion on the current control lab
- the traces showed that the remaining startup `/api` and `/apis` requests were
  coming from the earlier metrics probe seam, not from the later registry path

#### Recommendation

Park this line until the amplified lab says discovery is still the best next
performance target.

### Future Candidate: Read-Only RBAC Preflight Reduction

#### Why still demoted

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

- the amplified lab says it still matters
- or the discovery and metrics work are already done and we still want to reduce
  request fan-out
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
continuing measurement-led work. Later Step 7 probes did trigger a direction
reset: the next node-path step must be attribution work, not a speculative
optimization patch.
- safety hardening starts to conflict with fair baseline comparisons

## Step 7 Recommendation

The next implementation order on this machine should be:

1. PR 7A: small discovery cuts
2. keep the `metrics-small` ambient-off result as cleanup-only, not promoted
3. keep the `nodes-small` Step 7A and Step 7D artifacts as the node-path evidence base
4. use the Step 7E characterization result to inspect the node-to-pod drilldown path
5. only if that path still points back to pod counting later, test node pod counting reduction
6. keep RBAC preflight reduction only as a secondary or cleanup change
7. keep strict read-only hardening as a separate safety track

That is the highest-signal, lowest-regret Step 7 path for the current local lab.
