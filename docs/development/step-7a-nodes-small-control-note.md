# Step 7A Nodes-Small Control Note

- Captured at: 2026-03-29T22:37:12Z
- Git commit: `c279288f3025de4c455fb35ed92cb9140e7864f7`
- Artifact root: `artifacts/bench/20260329-153637/step7a-nodes-small-control-v1`
- Environment: local disposable cluster
- Profile: `nodes-small`
- Cluster path: `colima` plus `minikube --driver=docker`
- Minikube profile: `k9s-neo-nodes-small`
- Colima backend profile: `k9s-neo`
- Minikube nodes: `2`
- Namespace: `neo-bench`
- Kubeconfig: `/Users/akshay/.k9s-neo-state/kubeconfig`
- Vars file: `/Users/akshay/workspace/k9s-neo/hack/bench/vars.nodes-small.json`
- Hot node filter: `k9s-neo-nodes-small-m02`

## Scenario Set

- `nodes_first_render`

## Control Flags Under Test

The control binary wrapper adds only the promoted Step 7A flags:

- `--perf-skip-crd-augment`
- `--perf-skip-namespace-validation`

## Environment Defaults

- minikube CPUs: `2`
- minikube memory: `2048 MiB`
- minikube disk: `20g`
- container runtime: `containerd`
- colima CPUs: `2`
- colima memory: `3 GiB`
- colima disk: `20 GiB`
- seed manifest: `/Users/akshay/workspace/k9s-neo/hack/local-lab/manifests/neo-bench-nodes-small.yaml`

## Notes

- This artifact is the nodes-small control for later node-path A/B work.
- The initial control intentionally stays narrow and measures only `nodes_first_render`.
- `node_pod_drilldown` is deferred until the two-node seed proves stable enough to trust.
- Validate the trace `config_snapshot` when comparing later candidates:
  - `perf_skip_crd_augment=true`
  - `perf_skip_namespace_validation=true`
