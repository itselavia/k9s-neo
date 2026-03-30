# Step 7A Metrics-Small Control Note

- Captured at: 2026-03-29T03:15:14Z
- Git commit: `c279288f3025de4c455fb35ed92cb9140e7864f7`
- Artifact root: `artifacts/bench/20260328-200943/step7a-metrics-small-control-v1`
- Environment: local disposable cluster
- Profile: `metrics-small`
- Cluster path: `colima` plus `minikube --driver=docker`
- Minikube profile: `k9s-neo-metrics-small`
- Colima backend profile: `k9s-neo`
- Namespace: `neo-bench`
- Kubeconfig: `/Users/akshay/.k9s-neo-state/kubeconfig`
- Vars file: `/Users/akshay/workspace/k9s-neo/hack/bench/vars.metrics-small.json`
- Metrics-server: enabled

## Required Scenario Set

- `pods_startup`
- `pods_filter_settle`
- `nodes_first_render`
- `pod_yaml`
- `pod_describe`

## Control Flags Under Test

The control binary wrapper adds only the promoted Step 7A flags:

- `--perf-skip-crd-augment`
- `--perf-skip-namespace-validation`

Ambient metrics remain enabled in this control run.

## Warm-Run Medians

| Scenario | First Useful Row (ms) | Stable Interactive (ms) | Detail Ready (ms) | Filter Settle (ms) | API Requests | Bytes Before First Useful Row | Objects Before First Useful Row | Watches |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `pods_startup` | 1086.64 | 1107.46 | n/a | n/a | 18 | 31897 | 22 | 2 |
| `pods_filter_settle` | 1077.60 | 1149.01 | n/a | 58.72 | 18 | 31897 | 22 | 2 |
| `nodes_first_render` | 1179.12 | 1450.78 | n/a | n/a | 23 | 32735 | 23 | 4 |
| `pod_yaml` | n/a | n/a | 14.46 | n/a | 19 | 0 | 0 | 2 |
| `pod_describe` | n/a | n/a | 26.45 | n/a | 18 | 0 | 0 | 2 |

## Environment Defaults

- minikube CPUs: `2`
- minikube memory: `2048 MiB`
- minikube disk: `20g`
- container runtime: `containerd`
- colima CPUs: `2`
- colima memory: `3 GiB`
- colima disk: `20 GiB`
- seed manifest: `hack/local-lab/manifests/neo-bench.yaml`

## Notes

- This artifact is the control for later ambient-metrics A/B work on `metrics-small`.
- This control does not replace the frozen `control-small` Step 6 and Step 7A artifacts.
- Warm `pods_startup` traces now show real metrics hot-path traffic before first useful row:
  - `GET /api`
  - `GET /apis`
  - `GET /apis/metrics.k8s.io/v1beta1/nodes`
  - `GET /apis/metrics.k8s.io/v1beta1/namespaces/neo-bench/pods`
- Validate the trace `config_snapshot` when comparing later candidates:
  - `perf_skip_crd_augment=true`
  - `perf_skip_namespace_validation=true`
  - `perf_disable_ambient_metrics` absent or `false`
