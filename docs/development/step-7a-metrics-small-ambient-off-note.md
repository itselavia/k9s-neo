# Step 7A Metrics-Small Ambient-Off Candidate Note

- Captured at: 2026-03-29T12:48:34-07:00
- Git commit: `c279288f3025de4c455fb35ed92cb9140e7864f7`
- Candidate artifact root: `artifacts/bench/20260329-124834/step7a-metrics-small-ambient-off-v1`
- Control note: `docs/development/step-7a-metrics-small-control-note.md`
- Control artifact root: `artifacts/bench/20260328-200943/step7a-metrics-small-control-v1`
- Environment: local disposable cluster
- Profile: `metrics-small`
- Cluster path: `colima` plus `minikube --driver=docker`

## Candidate Flags Under Test

The candidate wrapper adds one flag on top of the promoted Step 7A control:

- `--perf-disable-ambient-metrics`

That means the A/B comparison is:

- control:
  - `--perf-skip-crd-augment`
  - `--perf-skip-namespace-validation`
- candidate:
  - `--perf-skip-crd-augment`
  - `--perf-skip-namespace-validation`
  - `--perf-disable-ambient-metrics`

## Warm-Median Comparison

| Scenario | Metric | Control | Candidate | Delta |
| --- | --- | ---: | ---: | ---: |
| `pods_startup` | first useful row | 1086.64 ms | 1060.28 ms | `-2.43%` |
| `pods_startup` | stable interactive | 1107.46 ms | 1098.73 ms | `-0.79%` |
| `pods_filter_settle` | first useful row | 1077.60 ms | 1067.66 ms | `-0.92%` |
| `pods_filter_settle` | filter settle | 58.72 ms | 55.30 ms | `-5.82%` |
| `nodes_first_render` | first useful row | 1179.12 ms | 1121.70 ms | `-4.87%` |
| `nodes_first_render` | stable interactive | 1450.78 ms | 1396.73 ms | `-3.73%` |
| `pod_yaml` | detail ready | 14.46 ms | 12.50 ms | `-13.55%` |
| `pod_describe` | detail ready | 26.45 ms | 20.68 ms | `-21.81%` |

## Request-Shape Comparison

### `pods_startup` warm median

- API requests before first useful row:
  - `18 -> 10` (`-44.44%`)
- response bytes before first useful row:
  - `31897 -> 2650` (`-91.69%`)

### `nodes_first_render` warm median

- API requests before first useful row:
  - `23 -> 15` (`-34.78%`)
- response bytes before first useful row:
  - `32735 -> 3575` (`-89.08%`)

## Trace Attribution

The candidate removed the expected request families before `first_useful_row`.

In the control traces, warm startup still includes:

- `GET /api`
- `GET /apis`
- `GET /apis/metrics.k8s.io/v1beta1/nodes`
- `GET /apis/metrics.k8s.io/v1beta1/namespaces/neo-bench/pods`

In the candidate traces, those metrics-related requests are gone before
`first_useful_row`, and the trace headers show:

- `perf_skip_crd_augment=true`
- `perf_skip_namespace_validation=true`
- `perf_disable_ambient_metrics=true`

## Verdict

This candidate is technically correct and meaningfully narrows the request
shape, but it does **not** earn promotion as a primary Step 7 performance win
on this machine.

Why:

- no primary startup scenario reached the project bar of `>=5%` warm-median
  improvement
- the closest primary result was `nodes_first_render` first useful row at
  `-4.87%`
- the strongest visible improvement was on the secondary `filter_settle` metric,
  not on the main startup gate

## Recommendation

- do not promote `--perf-disable-ambient-metrics` as a headline performance win
  from this artifact alone
- keep it as a real cleanup candidate with strong request-fan-out and byte
  reduction evidence
- move the next amplified-lab step to `nodes-small`
