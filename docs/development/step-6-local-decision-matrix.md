# Step 6 Local Decision Matrix Seed

## Purpose

This document seeds the first decision matrix from the captured local Step 6
baseline.

Baseline source:

- note: `docs/development/step-6-local-baseline-note.md`
- artifact root: `artifacts/bench/20260326-222409/local-baseline-v1`

This is local disposable-cluster evidence only. It is valid for before-and-after
engineering comparisons on this machine. It is not production-cluster evidence.

Follow-on Step 7A measurement:

- note: `docs/development/step-7a-small-discovery-cuts-note.md`
- artifact root: `artifacts/bench/20260326-235134/step7a-discovery-smallcuts-v1`

## Baseline Scenario Medians

Warm-run medians from the Step 6 baseline:

| Scenario | First Useful Row (ms) | Stable Interactive (ms) | Detail Ready (ms) | Filter Settle (ms) | API Requests | Bytes Before First Useful Row | Objects Before First Useful Row | Watches |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `pods_startup` | 1172.81 | 1213.65 | n/a | n/a | 13 | 28606 | 21 | 3 |
| `pods_filter_settle` | 1166.05 | 1252.92 | n/a | 54.20 | 13 | 28606 | 21 | 3 |
| `nodes_first_render` | 1247.31 | 1528.95 | n/a | n/a | 18 | 29531 | 22 | 5 |
| `pod_yaml` | n/a | n/a | 7.97 | n/a | 14 | 0 | 0 | 3 |
| `pod_describe` | n/a | n/a | 16.95 | n/a | 13 | 0 | 0 | 3 |

## What The Local Baseline Shows

### Observed startup/request shape

Representative warm-run request breakdowns:

- `pods_startup`
  - `/version:get` x3
  - `/api:get` x1
  - `/apis:get` x1
  - `customresourcedefinitions:watch` x1
  - `namespaces:list` x1
  - `pods:watch` x1
  - `nodes:watch` x1
  - `selfsubjectaccessreviews:create` x4
- `nodes_first_render`
  - everything above, plus:
  - `nodes:list` x1
  - `nodes:watch` x2
  - `pods:watch` x2
  - `selfsubjectaccessreviews:create` x6

### Strong local evidence

- discovery and startup breadth are directly visible on the hot path
- read-only RBAC preflight fan-out is directly visible on the hot path
- nodes view does additional work beyond the pod startup path even on a tiny cluster
- YAML and describe detail hydration are not the dominant local bottleneck

### Important local limitation

This local cluster does not currently expose a metrics API.

Evidence:

- `kubectl api-resources | rg metrics` returned no results
- traced runs show no metrics API requests

That means the local baseline cannot measure the full metrics-enabled data path
yet, even though a smaller startup metrics capability probe is still visible in
trace data on this machine.

### Metrics-enabled follow-on control

The dedicated `metrics-small` profile now exists and has a Step 7A control
capture:

- note: `docs/development/step-7a-metrics-small-control-note.md`
- artifact root: `artifacts/bench/20260328-200943/step7a-metrics-small-control-v1`

This control is the correct A/B baseline for later ambient-metrics work on a
metrics-enabled local cluster. It does not replace the frozen `control-small`
Step 6 and Step 7A artifacts.

That A/B is now complete:

- candidate note: `docs/development/step-7a-metrics-small-ambient-off-note.md`
- candidate artifact root: `artifacts/bench/20260329-124834/step7a-metrics-small-ambient-off-v1`

Result:

- real `metrics.k8s.io` pod and node requests disappeared before first useful row
- warm startup request count and bytes dropped sharply
- but the primary warm-median wins stayed below the project `>=5%` bar:
  - `pods_startup` first useful row: `-2.43%`
  - `nodes_first_render` first useful row: `-4.87%`

## Post-Step 7A Probe Outcomes

Two follow-on probes were useful diagnostically but did not earn promotion on
the current control lab:

- static-core registry smoke:
  - `artifacts/bench/20260327-201915/step7b2-static-core-smoke`
  - result: did not produce a keep-worthy win on the frozen control
- ambient-metrics diagnostic smoke:
  - `artifacts/bench/20260327-204553/step7c-ambient-metrics-smoke`
  - `artifacts/bench/20260327-204655/step7c-ambient-plus-7a-smoke`
  - result: removed the remaining startup `/api` and `/apis` requests, but did
    not produce a keep-worthy visible timing win on the frozen control

That means the current single-node control lab is now plateauing for shallow
startup probes. The next Step 7 move should be lab amplification, not another
startup micro-cut on the same control.

## Candidate Matrix

| Candidate Change | Expected Impact Band | Local Baseline Evidence | Primary Scenario | Implementation Complexity | Divergence Cost | Recommendation |
| --- | --- | --- | --- | --- | --- | --- |
| Small discovery cuts behind runtime switches | modest-visible | confirmed in Step 7A: `customresourcedefinitions:watch` and `namespaces:list` were removed from `pods_startup`, with about `8.8%` warm median improvement to first useful row on `pods_startup` and about `7.7%` on `nodes_first_render` | `pods_startup`, `pods_filter_settle`, `nodes_first_render` | low | low | keep as a measured probe; not the final discovery design |
| Narrow discovery with static core aliases plus Agones allowlist | uncertain on current control | Step 7B static-core smoke was diagnostically useful, but it did not earn promotion on the frozen control lab; the remaining startup `/api` and `/apis` requests were traced to the earlier metrics probe seam instead | `pods_startup`, `nodes_first_render` | medium | medium | park on the current control; revisit only if an amplified lab still shows discovery dominance |
| Reduce read-only RBAC preflight fan-out | modest on this local lab; potentially larger on higher-latency clusters | strong: 4 SAR creates on pod startup and 6 on nodes view even in the tiny lab, but only about `5.5 ms` and `10.6 ms` median direct SAR duration respectively before first useful row | `pods_startup`, `nodes_first_render` | medium | medium | demote; do later only if it still helps request fan-out or failure surface |
| Lazy or ambient metrics on hot paths | measurable and useful for request cleanup, but below the primary promotion bar on this machine | the `metrics-small` A/B removed `/api`, `/apis`, and real `metrics.k8s.io` pod/node requests before first useful row, with large request/byte reduction but only `-2.43%` on `pods_startup` and `-4.87%` on `nodes_first_render` warm first useful row | startup and list views | low | low | park as cleanup-only for now; keep it available as a non-promoted cleanup candidate |
| Disable node pod counting by default | substantial at real scale or broken-to-works | Step 7E on `nodes-small` proved the hotter drilldown path is real, but the shared node-entry phases stayed flat and the only stable drilldown-only request family before terminal Pod useful row was a cluster-scope `pods:watch` | `nodes_first_render`, then `node_pod_drilldown` | low | low | reject for now on this machine; inspect the drilldown hydration path before revisiting pod counting |
| Strict read-only hardening | negligible for speed, major for safety | performance is not the reason to do it; safety contract still requires it | whole product | medium | medium | keep as a separate safety workstream, not as the first performance claim |

## Step 7 Recommendation For This Machine

On this local lab, the current evidence-led order is:

1. keep the Step 7A small discovery cuts as a benchmarked runtime-only probe
2. preserve the current single-node control profile unchanged
3. keep the `metrics-small` A/B result as a real but non-promoted cleanup data point
4. use the new `nodes-small` profile to inspect and repair the hotter
   drilldown hydration path before revisiting node pod counting
5. keep static-core registry work parked until the amplified lab says discovery
   is still the best next target
6. keep read-only RBAC preflight reduction as secondary work, not the first
   shallow-win slot
7. keep strict read-only hardening separate from speed claims

## Open Questions

- how to enable a reproducible `metrics-small` profile without contaminating the
  original Step 6 control
- whether narrowing the node-to-pod drilldown hydration path removes the stable
  `pods:watch` tail on the reproducible `nodes-small` profile
- whether discovery remains a meaningful next target once the amplified profiles
  exist
