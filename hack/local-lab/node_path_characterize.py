#!/usr/bin/env python3

from __future__ import annotations

import argparse
import json
import statistics
import subprocess
from collections import Counter
from datetime import datetime, timezone
from pathlib import Path


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Write a repo-tracked Step 7E node-path characterization note."
    )
    parser.add_argument("--artifact-root", required=True)
    parser.add_argument("--control-note", required=True)
    parser.add_argument("--vars", required=True)
    parser.add_argument("--output", required=True)
    return parser.parse_args()


def load_note_artifact_root(note_path: Path) -> Path:
    for line in note_path.read_text().splitlines():
        if not line.startswith("- Artifact root: "):
            continue
        parts = line.split("`")
        if len(parts) >= 3:
            return Path(parts[1])
    raise SystemExit(f"could not determine artifact root from {note_path}")


def load_trace_events(path: Path) -> list[dict]:
    events: list[dict] = []
    with path.open() as handle:
        for raw_line in handle:
            raw_line = raw_line.strip()
            if not raw_line:
                continue
            try:
                events.append(json.loads(raw_line))
            except json.JSONDecodeError:
                continue
    return events


def load_run_json(path: Path) -> dict:
    with path.open() as handle:
        return json.load(handle)


def marker_event(
    events: list[dict],
    marker: str,
    *,
    view_seq: int | None = None,
    command_line: str | None = None,
) -> dict:
    for event in events:
        if event.get("type") != "lifecycle_mark":
            continue
        if event.get("marker") != marker:
            continue
        if view_seq is not None and event.get("view_seq") != view_seq:
            continue
        if command_line is not None and event.get("command_line") != command_line:
            continue
        return event
    raise SystemExit(f"missing marker {marker} for view_seq={view_seq}")


def delta_ms(start: dict, end: dict) -> float:
    return float(end["since_process_start_ms"]) - float(start["since_process_start_ms"])


def median(values: list[float]) -> float | None:
    if not values:
        return None
    return statistics.median(values)


def fmt_num(value: float | None) -> str:
    if value is None:
        return "n/a"
    return f"{value:.2f}"


def fmt_delta(control: float | None, candidate: float | None) -> str:
    if control is None or candidate is None:
        return "n/a"
    delta = candidate - control
    if control == 0:
        return f"{delta:.2f}"
    pct = (delta / control) * 100.0
    return f"{delta:.2f} ({pct:+.2f}%)"


def request_key(event: dict) -> str:
    verb = event.get("kube_verb", "").upper()
    path = event.get("path", "")
    status = event.get("status_code", "")
    return f"{verb} {path} [{status}]"


def request_counter_in_window(events: list[dict], start_ms: float, end_ms: float) -> Counter:
    counter: Counter = Counter()
    for event in events:
        event_type = event.get("type")
        if event_type not in {"kube_request_complete", "kube_stream_open"}:
            continue
        since = event.get("since_process_start_ms")
        if since is None or since < start_ms or since > end_ms:
            continue
        counter[request_key(event)] += 1
    return counter


def sum_resource_stats_in_window(
    events: list[dict], start_ms: float, end_ms: float, resource: str
) -> tuple[int, int]:
    total_bytes = 0
    total_objects = 0
    for event in events:
        if event.get("type") != "kube_request_complete":
            continue
        since = event.get("since_process_start_ms")
        if since is None or since < start_ms or since > end_ms:
            continue
        if event.get("resource") != resource:
            continue
        total_bytes += int(event.get("response_bytes", 0) or 0)
        total_objects += int(event.get("object_count", 0) or 0)
    return total_bytes, total_objects


def warm_run_paths(root: Path, scenario: str) -> list[Path]:
    return sorted((root / "runs" / scenario).glob("warm-*/run.json"))


def characterize_control_run(run_path: Path) -> dict:
    run_json = load_run_json(run_path)
    if run_json.get("status") != "ok":
        return {}

    events = load_trace_events(run_path.with_name("trace.jsonl"))
    matched = run_json["anchors"]["matched_wait_events"]
    initial_pod = matched[0]
    node_ready = run_json["anchors"]["terminal_wait_event"]
    command_start = marker_event(events, "command_start", command_line="nodes")
    node_view_seq = int(node_ready["view_seq"])
    node_activate = marker_event(events, "view_activate", view_seq=node_view_seq)
    node_model = marker_event(events, "first_model_built", view_seq=node_view_seq)
    node_render = marker_event(events, "first_render_committed", view_seq=node_view_seq)
    node_first_key = marker_event(events, "first_key_after_render", view_seq=node_view_seq)

    return {
        "total_first_useful_row_ms": float(node_ready["since_process_start_ms"]),
        "total_stable_interactive_ms": float(node_first_key["since_process_start_ms"]),
        "initial_pod_ready_ms": float(initial_pod["since_process_start_ms"]),
        "pod_to_nodes_command_ms": delta_ms(initial_pod, command_start),
        "command_to_node_activate_ms": delta_ms(command_start, node_activate),
        "node_activate_to_model_ms": delta_ms(node_activate, node_model),
        "node_model_to_render_ms": delta_ms(node_model, node_render),
        "node_render_to_useful_ms": delta_ms(node_render, node_ready),
        "shared_node_window_counts": request_counter_in_window(
            events,
            float(command_start["since_process_start_ms"]),
            float(node_ready["since_process_start_ms"]),
        ),
    }


def characterize_candidate_run(run_path: Path) -> dict:
    run_json = load_run_json(run_path)
    if run_json.get("status") != "ok":
        return {}

    events = load_trace_events(run_path.with_name("trace.jsonl"))
    matched = run_json["anchors"]["matched_wait_events"]
    initial_pod = matched[0]
    node_ready = matched[1]
    filter_settle = matched[2]
    final_pod = matched[3]
    command_start = marker_event(events, "command_start", command_line="nodes")

    node_view_seq = int(node_ready["view_seq"])
    node_activate = marker_event(events, "view_activate", view_seq=node_view_seq)
    node_model = marker_event(events, "first_model_built", view_seq=node_view_seq)
    node_render = marker_event(events, "first_render_committed", view_seq=node_view_seq)

    final_view_seq = int(final_pod["view_seq"])
    pod_activate = marker_event(events, "view_activate", view_seq=final_view_seq)
    pod_model = marker_event(events, "first_model_built", view_seq=final_view_seq)
    pod_render = marker_event(events, "first_render_committed", view_seq=final_view_seq)
    pod_first_key = marker_event(events, "first_key_after_render", view_seq=final_view_seq)

    drilldown_pod_bytes, drilldown_pod_objects = sum_resource_stats_in_window(
        events,
        float(node_ready["since_process_start_ms"]),
        float(final_pod["since_process_start_ms"]),
        "pods",
    )

    return {
        "total_first_useful_row_ms": float(final_pod["since_process_start_ms"]),
        "total_stable_interactive_ms": float(pod_first_key["since_process_start_ms"]),
        "initial_pod_ready_ms": float(initial_pod["since_process_start_ms"]),
        "pod_to_nodes_command_ms": delta_ms(initial_pod, command_start),
        "command_to_node_activate_ms": delta_ms(command_start, node_activate),
        "node_activate_to_model_ms": delta_ms(node_activate, node_model),
        "node_model_to_render_ms": delta_ms(node_model, node_render),
        "node_render_to_useful_ms": delta_ms(node_render, node_ready),
        "node_useful_to_filter_settle_ms": delta_ms(node_ready, filter_settle),
        "filter_settle_to_pod_activate_ms": delta_ms(filter_settle, pod_activate),
        "pod_activate_to_model_ms": delta_ms(pod_activate, pod_model),
        "pod_model_to_render_ms": delta_ms(pod_model, pod_render),
        "pod_render_to_useful_ms": delta_ms(pod_render, final_pod),
        "drilldown_tail_ms": delta_ms(node_ready, final_pod),
        "shared_node_window_counts": request_counter_in_window(
            events,
            float(command_start["since_process_start_ms"]),
            float(node_ready["since_process_start_ms"]),
        ),
        "drilldown_window_counts": request_counter_in_window(
            events,
            float(node_ready["since_process_start_ms"]),
            float(final_pod["since_process_start_ms"]),
        ),
        "drilldown_pod_bytes": drilldown_pod_bytes,
        "drilldown_pod_objects": drilldown_pod_objects,
        "hot_node_selected": filter_settle.get("selected_path", ""),
        "final_selected_path": final_pod.get("selected_path", ""),
        "final_rows_total": int(final_pod.get("rows_total", 0) or 0),
    }


def collect_stage_medians(rows: list[dict], keys: list[str]) -> dict[str, float | None]:
    return {key: median([row[key] for row in rows if key in row]) for key in keys}


def request_delta_rows(control_rows: list[dict], candidate_rows: list[dict]) -> list[dict]:
    control_samples = [row["shared_node_window_counts"] for row in control_rows]
    candidate_samples = [row["shared_node_window_counts"] for row in candidate_rows]
    keys = sorted({key for counter in control_samples + candidate_samples for key in counter})
    rows: list[dict] = []
    for key in keys:
        control_values = [counter.get(key, 0) for counter in control_samples]
        candidate_values = [counter.get(key, 0) for counter in candidate_samples]
        control_med = statistics.median(control_values)
        candidate_med = statistics.median(candidate_values)
        if control_med == candidate_med and max(control_values + candidate_values) == control_med:
            continue
        rows.append(
            {
                "key": key,
                "control_median": control_med,
                "candidate_median": candidate_med,
                "control_seen": sum(1 for value in control_values if value > 0),
                "candidate_seen": sum(1 for value in candidate_values if value > 0),
                "runs": len(candidate_values),
            }
        )
    rows.sort(key=lambda row: (-abs(row["candidate_median"] - row["control_median"]), row["key"]))
    return rows


def candidate_request_rows(candidate_rows: list[dict]) -> list[dict]:
    counters = [row["drilldown_window_counts"] for row in candidate_rows]
    keys = sorted({key for counter in counters for key in counter})
    rows: list[dict] = []
    for key in keys:
        values = [counter.get(key, 0) for counter in counters]
        seen = sum(1 for value in values if value > 0)
        med = statistics.median(values)
        if med == 0 and seen == 0:
            continue
        rows.append(
            {
                "key": key,
                "median": med,
                "seen": seen,
                "runs": len(values),
                "max": max(values),
            }
        )
    rows.sort(key=lambda row: (-row["median"], -row["seen"], row["key"]))
    return rows


def render_request_delta_table(rows: list[dict]) -> str:
    if not rows:
        return "- none\n"
    lines = [
        "| Request Path | Control Warm Median | Candidate Warm Median | Seen (control/candidate) |",
        "| --- | ---: | ---: | --- |",
    ]
    for row in rows:
        lines.append(
            f"| `{row['key']}` | {fmt_num(row['control_median'])} | {fmt_num(row['candidate_median'])} | "
            f"{row['control_seen']}/{row['runs']} -> {row['candidate_seen']}/{row['runs']} |"
        )
    return "\n".join(lines) + "\n"


def render_candidate_request_table(rows: list[dict]) -> str:
    if not rows:
        return "- none\n"
    lines = ["| Request Path | Warm Median Count | Seen | Max |", "| --- | ---: | --- | ---: |"]
    for row in rows:
        lines.append(
            f"| `{row['key']}` | {fmt_num(row['median'])} | {row['seen']}/{row['runs']} | {fmt_num(row['max'])} |"
        )
    return "\n".join(lines) + "\n"


def main() -> None:
    args = parse_args()
    artifact_root = Path(args.artifact_root)
    control_note_path = Path(args.control_note)
    vars_path = Path(args.vars)
    output_path = Path(args.output)

    repo_root = Path(__file__).resolve().parents[2]
    historical_control_root = load_note_artifact_root(control_note_path)
    vars_data = json.loads(vars_path.read_text())
    hot_node = vars_data.get("node_filter", "")

    control_rows = [
        row
        for path in warm_run_paths(artifact_root, "nodes_first_render")
        if (row := characterize_control_run(path))
    ]
    candidate_rows = [
        row
        for path in warm_run_paths(artifact_root, "node_pod_drilldown")
        if (row := characterize_candidate_run(path))
    ]

    if not control_rows or not candidate_rows:
        raise SystemExit("missing warm ok runs for characterization")

    control_keys = [
        "initial_pod_ready_ms",
        "pod_to_nodes_command_ms",
        "command_to_node_activate_ms",
        "node_activate_to_model_ms",
        "node_model_to_render_ms",
        "node_render_to_useful_ms",
        "total_first_useful_row_ms",
        "total_stable_interactive_ms",
    ]
    candidate_keys = [
        "initial_pod_ready_ms",
        "pod_to_nodes_command_ms",
        "command_to_node_activate_ms",
        "node_activate_to_model_ms",
        "node_model_to_render_ms",
        "node_render_to_useful_ms",
        "node_useful_to_filter_settle_ms",
        "filter_settle_to_pod_activate_ms",
        "pod_activate_to_model_ms",
        "pod_model_to_render_ms",
        "pod_render_to_useful_ms",
        "drilldown_tail_ms",
        "total_first_useful_row_ms",
        "total_stable_interactive_ms",
    ]

    control_medians = collect_stage_medians(control_rows, control_keys)
    candidate_medians = collect_stage_medians(candidate_rows, candidate_keys)

    shared_delta_rows = request_delta_rows(control_rows, candidate_rows)
    drilldown_request_rows = candidate_request_rows(candidate_rows)

    drilldown_pod_bytes = median([row["drilldown_pod_bytes"] for row in candidate_rows])
    drilldown_pod_objects = median([row["drilldown_pod_objects"] for row in candidate_rows])
    final_rows_total = median([row["final_rows_total"] for row in candidate_rows])
    selected_hot_node = Counter(row["hot_node_selected"] for row in candidate_rows).most_common(1)[0][0]
    final_selected_path = Counter(row["final_selected_path"] for row in candidate_rows).most_common(1)[0][0]

    candidate_total_median = candidate_medians["total_first_useful_row_ms"]
    candidate_total_max = max(row["total_first_useful_row_ms"] for row in candidate_rows)
    tail_median = candidate_medians["drilldown_tail_ms"]
    tail_max = max(row["drilldown_tail_ms"] for row in candidate_rows)
    control_total = control_medians["total_first_useful_row_ms"]

    commit_sha = subprocess.check_output(
        ["git", "-C", str(repo_root), "rev-parse", "HEAD"], text=True
    ).strip()
    captured_at = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")

    outlier_note = "- no severe warm outlier pattern detected beyond the normal drilldown path spread\n"
    if candidate_total_median and tail_median:
        if candidate_total_max > candidate_total_median * 1.5 and tail_max <= tail_median * 1.25:
            outlier_note = (
                f"- warm drilldown total has a severe outlier: median `{fmt_num(candidate_total_median)} ms`, "
                f"max `{fmt_num(candidate_total_max)} ms`\n"
                f"- the drilldown-only tail stays comparatively tight: median `{fmt_num(tail_median)} ms`, "
                f"max `{fmt_num(tail_max)} ms`\n"
                "- that means the worst warm outlier is already happening before the final Pod drilldown tail finishes\n"
            )

    note = f"""# Step 7E Node-Path Characterization Note

- Captured at: {captured_at}
- Git commit: `{commit_sha}`
- Characterization artifact root: `{artifact_root}`
- Historical control note: `docs/development/step-7a-nodes-small-control-note.md`
- Historical control artifact root: `{historical_control_root}`
- Environment: local disposable cluster
- Profile: `nodes-small`
- Scenario pair:
  - `nodes_first_render`
  - `node_pod_drilldown`
- Vars file: `{vars_path}`
- Hot node filter: `{hot_node}`

## Purpose

Characterize where the extra `node_pod_drilldown` time appears on `nodes-small`
before starting any node optimization patch.

## Warm-Median Scenario Comparison

| Metric | `nodes_first_render` | `node_pod_drilldown` | Delta |
| --- | ---: | ---: | ---: |
| first useful row | {fmt_num(control_total)} ms | {fmt_num(candidate_total_median)} ms | {fmt_delta(control_total, candidate_total_median)} |
| stable interactive | {fmt_num(control_medians["total_stable_interactive_ms"])} ms | {fmt_num(candidate_medians["total_stable_interactive_ms"])} ms | {fmt_delta(control_medians["total_stable_interactive_ms"], candidate_medians["total_stable_interactive_ms"])} |

## Shared Node-Entry Phases

These stages exist in both scenarios and end at the `Node` view `first_useful_row`.

| Stage | `nodes_first_render` Warm Median | `node_pod_drilldown` Warm Median | Delta |
| --- | ---: | ---: | ---: |
| initial Pod first useful row | {fmt_num(control_medians["initial_pod_ready_ms"])} ms | {fmt_num(candidate_medians["initial_pod_ready_ms"])} ms | {fmt_delta(control_medians["initial_pod_ready_ms"], candidate_medians["initial_pod_ready_ms"])} |
| initial Pod useful row -> `:nodes` command | {fmt_num(control_medians["pod_to_nodes_command_ms"])} ms | {fmt_num(candidate_medians["pod_to_nodes_command_ms"])} ms | {fmt_delta(control_medians["pod_to_nodes_command_ms"], candidate_medians["pod_to_nodes_command_ms"])} |
| `:nodes` command -> Node activate | {fmt_num(control_medians["command_to_node_activate_ms"])} ms | {fmt_num(candidate_medians["command_to_node_activate_ms"])} ms | {fmt_delta(control_medians["command_to_node_activate_ms"], candidate_medians["command_to_node_activate_ms"])} |
| Node activate -> first model built | {fmt_num(control_medians["node_activate_to_model_ms"])} ms | {fmt_num(candidate_medians["node_activate_to_model_ms"])} ms | {fmt_delta(control_medians["node_activate_to_model_ms"], candidate_medians["node_activate_to_model_ms"])} |
| Node first model built -> first render committed | {fmt_num(control_medians["node_model_to_render_ms"])} ms | {fmt_num(candidate_medians["node_model_to_render_ms"])} ms | {fmt_delta(control_medians["node_model_to_render_ms"], candidate_medians["node_model_to_render_ms"])} |
| Node first render committed -> first useful row | {fmt_num(control_medians["node_render_to_useful_ms"])} ms | {fmt_num(candidate_medians["node_render_to_useful_ms"])} ms | {fmt_delta(control_medians["node_render_to_useful_ms"], candidate_medians["node_render_to_useful_ms"])} |

## Drilldown-Only Tail

These stages exist only in `node_pod_drilldown`, starting after the `Node`
view is already useful.

| Stage | Warm Median |
| --- | ---: |
| Node first useful row -> filter settle | {fmt_num(candidate_medians["node_useful_to_filter_settle_ms"])} ms |
| filter settle -> final Pod activate | {fmt_num(candidate_medians["filter_settle_to_pod_activate_ms"])} ms |
| final Pod activate -> first model built | {fmt_num(candidate_medians["pod_activate_to_model_ms"])} ms |
| final Pod first model built -> first render committed | {fmt_num(candidate_medians["pod_model_to_render_ms"])} ms |
| final Pod first render committed -> first useful row | {fmt_num(candidate_medians["pod_render_to_useful_ms"])} ms |
| total Node useful row -> final Pod useful row | {fmt_num(tail_median)} ms |

## Shared Node-Entry Request Differences

Median request-count differences in the shared `:nodes` command window:

{render_request_delta_table(shared_delta_rows)}
## Drilldown-Only Request Shape

Median request families between `Node` `first_useful_row` and final `Pod`
`first_useful_row`:

{render_candidate_request_table(drilldown_request_rows)}
Additional drilldown-only sanity:

- hot node selected at `filter_settle`: `{selected_hot_node}`
- representative final selected path: `{final_selected_path}`
- final Pod rows total warm median: `{fmt_num(final_rows_total)}`
- pod response bytes in the drilldown-only window: `{fmt_num(drilldown_pod_bytes)}`
- pod object count in the drilldown-only window: `{fmt_num(drilldown_pod_objects)}`

## Outlier Read

{outlier_note}
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
"""

    output_path.write_text(note)


if __name__ == "__main__":
    main()
