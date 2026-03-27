# Step 6 Local Decision Matrix Seed

## Purpose

This document seeds the first decision matrix from the captured local Step 6
baseline.

Baseline source:

- note: `docs/development/step-6-local-baseline-note.md`
- artifact root: `artifacts/bench/20260326-222409/local-baseline-v1`

This is local disposable-cluster evidence only. It is valid for before-and-after
engineering comparisons on this machine. It is not production-cluster evidence.

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
- auth preflight fan-out is directly visible on the hot path
- nodes view does additional work beyond the pod startup path even on a tiny cluster
- YAML and describe detail hydration are not the dominant local bottleneck

### Important local limitation

This local cluster does not currently expose a metrics API.

Evidence:

- `kubectl api-resources | rg metrics` returned no results
- traced runs show no metrics API requests

That means the local baseline cannot yet measure the effect of ambient metrics
removal, even though that remains a plausible production-relevant hypothesis.

## Candidate Matrix

| Candidate Change | Expected Impact Band | Local Baseline Evidence | Primary Scenario | Implementation Complexity | Divergence Cost | Recommendation |
| --- | --- | --- | --- | --- | --- | --- |
| Narrow discovery with static core aliases plus Agones allowlist | substantial | strong: discovery endpoints, namespace list, and CRD watch are on every startup path | `pods_startup`, `nodes_first_render` | medium | medium | measure first on this local baseline |
| Reduce read preflight auth fan-out | modest to substantial | strong: 4 SAR creates on pod startup and 6 on nodes view even in the tiny lab | `pods_startup`, `nodes_first_render` | medium | medium | measure early after discovery narrowing or as a separate distinct change |
| Lazy metrics by default | substantial in metrics-enabled clusters | weak locally: current lab has no metrics API, so the baseline does not expose this cost | startup and list views | low | low | keep high priority globally, but do not use this lab as proof until metrics are enabled |
| Disable node pod counting by default | substantial at real scale or broken-to-works | weak locally: single-node lab does not stress the catastrophic pod-count path yet | `nodes_first_render`, later `node_pod_drilldown` | low | low | keep queued, but not the first local measurement |
| Strict read-only hardening | negligible for speed, major for safety | performance is not the reason to do it; safety contract still requires it | whole product | medium | medium | keep as a separate safety workstream, not as the first performance claim |

## Step 7 Recommendation For This Machine

On this local lab, the most evidence-led first change is:

1. narrow discovery with static core aliases plus Agones allowlist
2. reduce read preflight auth fan-out
3. choose between:
   - enabling local metrics only if we specifically want to measure lazy metrics here
   - or continuing to node-path work without local metrics evidence
4. keep strict read-only hardening separate from speed claims

## Open Questions

- whether to enable metrics-server locally before measuring the metrics hypothesis
- whether to add a small multi-node local profile before trying to measure node pod counting
- whether discovery narrowing alone removes enough startup breadth that auth fan-out becomes the clearer next target
