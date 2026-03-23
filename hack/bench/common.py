from __future__ import annotations

import csv
import json
import statistics
from collections import defaultdict
from pathlib import Path
from typing import Any

SCHEMA_VERSION = 2
FIXTURE_SCHEMA_VERSION = 1

SOURCE_KIND_LIVE = "live"
SOURCE_KIND_REPLAY = "replay"

REQUEST_START_TYPES = {"kube_request_complete", "kube_stream_open"}
REQUEST_END_TYPES = {"kube_request_complete", "kube_stream_close"}


def load_json(path: Path) -> Any:
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def first_scoped_marker(
    markers: dict[str, list[dict[str, Any]]],
    name: str,
    *,
    view_seq: int | None = None,
    detail_kind: str | None = None,
) -> dict[str, Any] | None:
    for event in markers.get(name, []):
        if view_seq is not None and event.get("view_seq") != view_seq:
            continue
        if detail_kind is not None and event.get("detail_kind") != detail_kind:
            continue
        return event
    return None


def derive_metrics(
    events: list[dict[str, Any]],
    samples: list[dict[str, Any]],
    *,
    terminal_wait_event: dict[str, Any] | None = None,
    view_anchor_event: dict[str, Any] | None = None,
) -> dict[str, Any]:
    markers: dict[str, list[dict[str, Any]]] = defaultdict(list)
    for event in events:
        if event.get("type") == "lifecycle_mark":
            markers[event.get("marker", "")].append(event)

    view_seq = view_anchor_event.get("view_seq") if view_anchor_event else None
    detail_kind = terminal_wait_event.get("detail_kind") if terminal_wait_event else None

    first_paint = first_scoped_marker(markers, "first_render_committed", view_seq=view_seq) if view_seq is not None else None
    first_useful = first_scoped_marker(markers, "first_useful_row", view_seq=view_seq) if view_seq is not None else None
    first_key = first_scoped_marker(markers, "first_key_after_render", view_seq=view_seq) if view_seq is not None else None
    filter_start = first_scoped_marker(markers, "filter_start", view_seq=view_seq) if view_seq is not None else None
    filter_settle = first_scoped_marker(markers, "filter_settle", view_seq=view_seq) if view_seq is not None else None
    detail_open = first_scoped_marker(markers, "detail_open_start", detail_kind=detail_kind) if detail_kind else None
    detail_ready = first_scoped_marker(markers, "detail_content_ready", detail_kind=detail_kind) if detail_kind else None

    cutoff = first_useful.get("since_process_start_ms") if first_useful else None
    request_start_events = [event for event in events if event.get("type") in REQUEST_START_TYPES]
    request_end_events = [event for event in events if event.get("type") in REQUEST_END_TYPES]

    counts_by_resource_verb: dict[str, int] = defaultdict(int)
    for event in request_start_events:
        resource = event.get("resource") or event.get("path") or "unknown"
        verb = event.get("kube_verb") or event.get("http_method") or "unknown"
        counts_by_resource_verb[f"{resource}:{verb}"] += 1

    bytes_before_cutoff = 0
    objects_before_cutoff = 0
    if cutoff is not None:
        for event in request_end_events:
            if event.get("since_process_start_ms", 0) <= cutoff:
                bytes_before_cutoff += int(event.get("response_bytes", 0) or 0)
                objects_before_cutoff += int(event.get("object_count", 0) or 0)

    watch_count = sum(1 for event in request_start_events if event.get("watch"))
    watches_before_cutoff = 0
    if cutoff is not None:
        watches_before_cutoff = sum(
            1
            for event in request_start_events
            if event.get("watch") and event.get("since_process_start_ms", 0) <= cutoff
        )

    peak_rss = max((sample.get("rss_bytes", 0) for sample in samples), default=0)
    peak_cpu = max((sample.get("cpu_percent", 0.0) for sample in samples), default=0.0)

    return {
        "time_to_first_paint_ms": first_paint.get("since_process_start_ms") if first_paint else None,
        "time_to_first_useful_row_ms": first_useful.get("since_process_start_ms") if first_useful else None,
        "time_to_stable_interactive_ms": first_key.get("since_process_start_ms") if first_key else None,
        "filter_settle_ms": (
            filter_settle.get("since_process_start_ms", 0) - filter_start.get("since_process_start_ms", 0)
            if filter_start and filter_settle
            else None
        ),
        "detail_ready_ms": (
            detail_ready.get("since_process_start_ms", 0) - detail_open.get("since_process_start_ms", 0)
            if detail_open and detail_ready
            else None
        ),
        "peak_rss_bytes_30s": peak_rss,
        "peak_cpu_percent_30s": peak_cpu,
        "api_request_count": len(request_start_events),
        "api_request_count_by_resource_verb": dict(sorted(counts_by_resource_verb.items())),
        "total_response_bytes_before_first_useful_row": bytes_before_cutoff,
        "total_object_count_before_first_useful_row": objects_before_cutoff,
        "watch_count": watch_count,
        "watches_before_first_useful_row": watches_before_cutoff,
    }


def aggregate_numeric(values: list[float]) -> dict[str, float]:
    return {
        "count": float(len(values)),
        "min": min(values),
        "median": statistics.median(values),
        "mean": statistics.mean(values),
        "max": max(values),
    }


def aggregate_runs(results: list[dict[str, Any]]) -> dict[str, Any]:
    grouped: dict[tuple[str, str], list[dict[str, Any]]] = defaultdict(list)
    for result in results:
        grouped[(result["scenario"], result["mode"])].append(result)

    numeric_fields = [
        "time_to_first_paint_ms",
        "time_to_first_useful_row_ms",
        "time_to_stable_interactive_ms",
        "filter_settle_ms",
        "detail_ready_ms",
        "peak_rss_bytes_30s",
        "peak_cpu_percent_30s",
        "api_request_count",
        "total_response_bytes_before_first_useful_row",
        "total_object_count_before_first_useful_row",
        "watch_count",
        "watches_before_first_useful_row",
    ]

    aggregates: dict[str, Any] = {}
    for (scenario, mode), items in grouped.items():
        status_counts: dict[str, int] = defaultdict(int)
        source_kinds = sorted({item.get("source_kind", SOURCE_KIND_LIVE) for item in items})
        for item in items:
            status_counts[item["status"]] += 1
        ok_items = [item for item in items if item["status"] == "ok"]

        entry: dict[str, Any] = {
            "scenario": scenario,
            "mode": mode,
            "statuses": [item["status"] for item in items],
            "status_counts": dict(sorted(status_counts.items())),
            "source_kinds": source_kinds,
            "ok_run_count": len(ok_items),
            "total_run_count": len(items),
        }
        for field in numeric_fields:
            values = [item["metrics"].get(field) for item in ok_items if item["metrics"].get(field) is not None]
            if values:
                entry[field] = aggregate_numeric([float(value) for value in values])
        aggregates[f"{scenario}:{mode}"] = entry

    return aggregates


def write_csv(path: Path, results: list[dict[str, Any]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", newline="", encoding="utf-8") as handle:
        writer = csv.writer(handle)
        writer.writerow(
            [
                "source_kind",
                "scenario",
                "mode",
                "run_index",
                "status",
                "time_to_first_paint_ms",
                "time_to_first_useful_row_ms",
                "time_to_stable_interactive_ms",
                "filter_settle_ms",
                "detail_ready_ms",
                "peak_rss_bytes_30s",
                "peak_cpu_percent_30s",
                "api_request_count",
                "watch_count",
                "watches_before_first_useful_row",
                "total_response_bytes_before_first_useful_row",
                "total_object_count_before_first_useful_row",
                "api_request_count_by_resource_verb",
            ]
        )
        for result in results:
            metrics = result["metrics"]
            writer.writerow(
                [
                    result.get("source_kind", SOURCE_KIND_LIVE),
                    result["scenario"],
                    result["mode"],
                    result["run_index"],
                    result["status"],
                    metrics.get("time_to_first_paint_ms"),
                    metrics.get("time_to_first_useful_row_ms"),
                    metrics.get("time_to_stable_interactive_ms"),
                    metrics.get("filter_settle_ms"),
                    metrics.get("detail_ready_ms"),
                    metrics.get("peak_rss_bytes_30s"),
                    metrics.get("peak_cpu_percent_30s"),
                    metrics.get("api_request_count"),
                    metrics.get("watch_count"),
                    metrics.get("watches_before_first_useful_row"),
                    metrics.get("total_response_bytes_before_first_useful_row"),
                    metrics.get("total_object_count_before_first_useful_row"),
                    json.dumps(metrics.get("api_request_count_by_resource_verb", {}), sort_keys=True),
                ]
            )


def write_report(path: Path, aggregates: dict[str, Any]) -> None:
    all_source_kinds = sorted({kind for item in aggregates.values() for kind in item.get("source_kinds", [SOURCE_KIND_LIVE])})
    lines = [
        "# Benchmark Summary",
        "",
        f"Artifact source_kind(s): {', '.join(all_source_kinds)}",
        "",
        "Aggregates include only runs with status `ok`.",
        "",
    ]
    for key in sorted(aggregates):
        item = aggregates[key]
        lines.append(f"## {item['scenario']} ({item['mode']})")
        lines.append("")
        lines.append(f"- source_kinds: {', '.join(item['source_kinds'])}")
        status_counts = ", ".join(f"{status}={count}" for status, count in item["status_counts"].items())
        lines.append(f"- status_counts: {status_counts}")
        lines.append(f"- ok_runs: {item['ok_run_count']}/{item['total_run_count']}")
        for metric, stats in sorted(item.items()):
            if metric in {"scenario", "mode", "statuses", "status_counts", "source_kinds", "ok_run_count", "total_run_count"}:
                continue
            lines.append(
                f"- {metric}: min={stats['min']:.2f}, median={stats['median']:.2f}, mean={stats['mean']:.2f}, max={stats['max']:.2f}"
            )
        lines.append("")
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text("\n".join(lines), encoding="utf-8")


def write_run_artifacts(run_dir: Path, result: dict[str, Any], transcript: bytes = b"") -> dict[str, Any]:
    run_dir.mkdir(parents=True, exist_ok=True)
    trace_path = run_dir / "trace.jsonl"
    transcript_path = run_dir / "pty.txt"
    result_path = run_dir / "run.json"

    source_kind = result.get("source_kind", SOURCE_KIND_LIVE)
    if source_kind != SOURCE_KIND_LIVE or not trace_path.exists():
        with trace_path.open("w", encoding="utf-8") as handle:
            for event in result.get("events", []):
                handle.write(json.dumps(event, sort_keys=True))
                handle.write("\n")

    transcript_path.write_bytes(transcript)

    result = dict(result)
    result["trace_file"] = str(trace_path)
    result["transcript_file"] = str(transcript_path)
    result_path.write_text(json.dumps(result, indent=2, sort_keys=True), encoding="utf-8")

    return result
