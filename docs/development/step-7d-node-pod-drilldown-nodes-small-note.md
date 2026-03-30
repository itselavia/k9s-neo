# Step 7D Nodes-Small Node-Pod-Drilldown Note

- Captured at: 2026-03-29T23:32:46Z
- Git commit: `c279288f3025de4c455fb35ed92cb9140e7864f7`
- Candidate artifact root: `artifacts/bench/20260329-162922/step7d-node-pod-drilldown-nodes-small-v1`
- Control note: `docs/development/step-7a-nodes-small-control-note.md`
- Control artifact root: `artifacts/bench/20260329-153637/step7a-nodes-small-control-v1`
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

## Scenario Under Test

- `node_pod_drilldown`

Scenario intent:

- start in scoped Pod view
- switch to `Node`
- filter to the hot worker node
- enter the selected node's Pod view
- complete on the matched `Pod` view `first_useful_row`

## Perf Flags Under Test

The candidate wrapper adds only the promoted Step 7A flags:

- `--perf-skip-crd-augment`
- `--perf-skip-namespace-validation`

## Warm-Median Comparison

Comparison basis:

- control: `nodes_first_render` on `nodes-small`
- candidate: `node_pod_drilldown` on `nodes-small`

| Metric | Nodes-Small Control | Drilldown Candidate | Delta |
| --- | ---: | ---: | ---: |
| first useful row | 1153.3554585000002 ms | 1353.444667 ms | 200.09 (+17.35%) |
| stable interactive | 1428.949771 ms | 1663.339417 ms | 234.39 (+16.40%) |
| API requests | 17.0 | 17.0 | 0.00 (+0.00%) |
| bytes before first useful row | 28206.0 | 28206.0 | 0.00 (+0.00%) |
| objects before first useful row | 22.0 | 22.0 | 0.00 (+0.00%) |
| peak CPU percent | 22.549999999999997 | 23.8 | 1.25 (+5.54%) |
| peak RSS bytes | 96526336.0 | 97386496.0 | 860160.00 (+0.89%) |

## Request-Shape Result

- terminal matched view: `Pod`
- representative warm-run `pods:list` before first useful row: `0`
- representative warm-run `api_request_count_by_resource_verb`:

```json
{"/api:get": 1, "/apis:get": 1, "/version:get": 4, "nodes:list": 1, "nodes:watch": 2, "pods:watch": 2, "selfsubjectaccessreviews:create": 6}
```

## Staff Read

The drilldown artifact exists, but the promotion decision should stay strict until the warm median clearly separates from the nodes-small nodes_first_render control.

The decision bar for a follow-on node optimization patch remains:

- the drilldown path should be clearly hotter than the nodes-small control
- and the trace should show pod-path work that matches the node-to-pod hypothesis

## Promotion Decision

- reject for now

## Next Step

- compare this artifact against the existing nodes-small control
- if the drilldown path is clearly hotter and trace-attributable, start the node-path optimization probe
- otherwise, park node pod counting on this lab
