# Step 7E Node-Path Characterization Note

- Captured at: 2026-03-30T03:44:38Z
- Git commit: `c279288f3025de4c455fb35ed92cb9140e7864f7`
- Characterization artifact root: `artifacts/bench/20260329-204049/step7e-node-path-characterization-nodes-small-v1`
- Historical control note: `docs/development/step-7a-nodes-small-control-note.md`
- Historical control artifact root: `artifacts/bench/20260329-153637/step7a-nodes-small-control-v1`
- Environment: local disposable cluster
- Profile: `nodes-small`
- Scenario pair:
  - `nodes_first_render`
  - `node_pod_drilldown`
- Vars file: `/Users/akshay/workspace/k9s-neo/hack/bench/vars.nodes-small.json`
- Hot node filter: `k9s-neo-nodes-small-m02`

## Purpose

Characterize where the extra `node_pod_drilldown` time appears on `nodes-small`
before starting any node optimization patch.

## Warm-Median Scenario Comparison

| Metric | `nodes_first_render` | `node_pod_drilldown` | Delta |
| --- | ---: | ---: | ---: |
| first useful row | 1216.93 ms | 1360.04 ms | 143.11 (+11.76%) |
| stable interactive | 1493.37 ms | 1666.20 ms | 172.83 (+11.57%) |

## Shared Node-Entry Phases

These stages exist in both scenarios and end at the `Node` view `first_useful_row`.

| Stage | `nodes_first_render` Warm Median | `node_pod_drilldown` Warm Median | Delta |
| --- | ---: | ---: | ---: |
| initial Pod first useful row | 1090.95 ms | 1095.22 ms | 4.27 (+0.39%) |
| initial Pod useful row -> `:nodes` command | 82.05 ms | 68.46 ms | -13.59 (-16.56%) |
| `:nodes` command -> Node activate | 6.76 ms | 6.18 ms | -0.58 (-8.57%) |
| Node activate -> first model built | 4.06 ms | 3.65 ms | -0.40 (-9.95%) |
| Node first model built -> first render committed | 1.17 ms | 1.18 ms | 0.01 (+0.63%) |
| Node first render committed -> first useful row | 2.20 ms | 2.15 ms | -0.05 (-2.24%) |

## Drilldown-Only Tail

These stages exist only in `node_pod_drilldown`, starting after the `Node`
view is already useful.

| Stage | Warm Median |
| --- | ---: |
| Node first useful row -> filter settle | 122.99 ms |
| filter settle -> final Pod activate | 49.60 ms |
| final Pod activate -> first model built | 1.36 ms |
| final Pod first model built -> first render committed | 0.88 ms |
| final Pod first render committed -> first useful row | 2.22 ms |
| total Node useful row -> final Pod useful row | 176.24 ms |

## Shared Node-Entry Request Differences

Median request-count differences in the shared `:nodes` command window:

| Request Path | Control Warm Median | Candidate Warm Median | Seen (control/candidate) |
| --- | ---: | ---: | --- |
| `WATCH /api/v1/pods [200]` | 0.00 | 0.00 | 0/10 -> 1/10 |

## Drilldown-Only Request Shape

Median request families between `Node` `first_useful_row` and final `Pod`
`first_useful_row`:

| Request Path | Warm Median Count | Seen | Max |
| --- | ---: | --- | ---: |
| `WATCH /api/v1/pods [200]` | 1.00 | 9/10 | 1.00 |
| `LIST /api/v1/namespaces/neo-bench/nodes [404]` | 0.00 | 2/10 | 1.00 |

Additional drilldown-only sanity:

- hot node selected at `filter_settle`: `k9s-neo-nodes-small-m02`
- representative final selected path: `kube-system/kindnet-bss9b`
- final Pod rows total warm median: `16.00`
- pod response bytes in the drilldown-only window: `0.00`
- pod object count in the drilldown-only window: `0.00`

## Outlier Read

- no severe warm outlier pattern detected beyond the normal drilldown path spread

## Staff Read

The hotter node drilldown path is real, but the extra time is not pointing at
node pod counting yet. The stable drilldown-only tail is dominated by hot-node
filter settle plus the final Pod activation path, and the only stable
drilldown-only request family before the terminal Pod useful row is a
cluster-scope `pods:watch` open.

## Decision

- reject node pod counting for now

## Next Step

- inspect and repair the node-to-pod drilldown hydration path, especially why
  the drilldown opens a cluster-scope `pods:watch` before the terminal Pod
  useful row, before testing any node pod counting patch
