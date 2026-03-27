# Step 6 Local Baseline Note

- Captured at: 2026-03-27T05:27:19Z
- Git commit: `7b9a7c08d7b2024cee1ca9180e0ea57e814d7b78`
- Artifact root: `artifacts/bench/20260326-222409/local-baseline-v1`
- Environment: local disposable cluster
- Cluster path: `colima` plus `minikube --driver=docker`
- Minikube profile: `k9s-neo`
- Namespace: `neo-bench`
- Kubeconfig: `/Users/akshay/.k9s-neo-state/kubeconfig`

## Required Scenario Set

- `pods_startup`
- `pods_filter_settle`
- `nodes_first_render`
- `pod_yaml`
- `pod_describe`

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

- This baseline is valid for before-and-after engineering comparisons on this machine.
- This baseline is not production-cluster evidence.
- Recreate the local lab from repo-owned scripts before comparing later shallow changes.
