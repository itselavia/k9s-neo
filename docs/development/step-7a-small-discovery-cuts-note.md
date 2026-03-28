# Step 7A Small Discovery Cuts Note

- Captured at: 2026-03-27T06:57:39Z
- Artifact root: `artifacts/bench/20260326-235134/step7a-discovery-smallcuts-v1`
- Environment: local disposable cluster
- Cluster path: `colima` plus `minikube --driver=docker`
- Minikube profile: `k9s-neo`
- Namespace: `neo-bench`
- Flags under test:
  - `--perf-skip-crd-augment`
  - `--perf-skip-namespace-validation`

## What Changed

Step 7A was intentionally kept narrow and runtime-only:

- skip CRD augmentation on the startup path
- skip startup namespace validation
- avoid namespace completion fetches when the prompt already contains the active namespace

Default behavior is unchanged. These cuts are still hidden perf switches, not the
new product default.

## Warm-Run Median Comparison

Compared against the Step 6 local baseline in
`artifacts/bench/20260326-222409/local-baseline-v1`.

| Scenario | Metric | Baseline | Step 7A | Delta |
| --- | --- | ---: | ---: | ---: |
| `pods_startup` | first useful row | 1172.81 ms | 1069.86 ms | -102.95 ms (-8.8%) |
| `pods_startup` | stable interactive | 1213.65 ms | 1130.68 ms | -82.97 ms (-6.8%) |
| `pods_startup` | API requests | 13 | 12 | -1 |
| `pods_startup` | bytes before first useful row | 28606 | 27281 | -1325 (-4.6%) |
| `pods_startup` | watches | 3 | 2 | -1 |
| `pods_filter_settle` | first useful row | 1166.05 ms | 1051.60 ms | -114.45 ms (-9.8%) |
| `pods_filter_settle` | filter settle | 54.20 ms | 50.52 ms | -3.69 ms (-6.8%) |
| `nodes_first_render` | first useful row | 1247.31 ms | 1151.17 ms | -96.14 ms (-7.7%) |
| `nodes_first_render` | stable interactive | 1528.95 ms | 1441.36 ms | -87.59 ms (-5.7%) |
| `nodes_first_render` | API requests | 18 | 17 | -1 |
| `nodes_first_render` | bytes before first useful row | 29531 | 28206 | -1325 (-4.5%) |
| `nodes_first_render` | watches | 5 | 4 | -1 |
| `pod_yaml` | first paint | 1207.27 ms | 1135.13 ms | -72.14 ms (-6.0%) |
| `pod_describe` | first paint | 1277.77 ms | 1149.27 ms | -128.50 ms (-10.1%) |

## Request-Shape Result

On `pods_startup`, Step 7A removes the two startup request families it targeted
before first useful row:

- `customresourcedefinitions:watch`
- `namespaces:list`

Representative before/after request counts before first useful row:

- baseline:
  - `customresourcedefinitions:watch` x1
  - `namespaces:list` x1
  - `selfsubjectaccessreviews:create` x4
  - `pods:watch` x1
  - `nodes:watch` x1
- Step 7A:
  - `selfsubjectaccessreviews:create` x3
  - `pods:watch` x1
  - `nodes:watch` x1

## Staff-Level Read

- This is a real local win, not a smoke-only fluke.
- The effect is visible and repeatable, but still modest-to-moderate rather than
  transformative.
- The cuts are valuable because they remove concrete startup breadth with low
  divergence cost.
- Discovery is still clearly on the hot path after these cuts, so the next move
  should be the larger static-core discovery step rather than smaller cleanup
  work.

## Recommendation

- keep Step 7A as a benchmarked runtime-switch probe
- do not claim this is the final discovery design
- start Step 7B next:
  - static core aliases
  - explicit Agones allowlist
  - generic CRDs off the hot path by default
