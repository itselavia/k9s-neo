#!/usr/bin/env python3

from __future__ import annotations

import argparse
import csv
import json
import os
import pty
import select
import shlex
import shutil
import signal
import statistics
import struct
import subprocess
import tempfile
import termios
import threading
import time
from collections import defaultdict
from dataclasses import dataclass
from fcntl import ioctl
from pathlib import Path
from typing import Any

SCHEMA_VERSION = 2
DEFAULT_COLS = 220
DEFAULT_ROWS = 60
DEFAULT_STEP_TIMEOUT_MS = 30_000
DEFAULT_SAMPLE_INTERVAL_S = 0.25
DEFAULT_SAMPLE_WINDOW_S = 30.0
REQUEST_START_TYPES = {"kube_request_complete", "kube_stream_open"}
REQUEST_END_TYPES = {"kube_request_complete", "kube_stream_close"}

KEY_MAP = {
    "ENTER": "\r",
    "ESC": "\x1b",
    "CTRL_C": "\x03",
    "CTRL_R": "\x12",
    "TAB": "\t",
    "BACKSPACE": "\x7f",
}


@dataclass
class CacheLayout:
    home: Path
    config_dir: Path


class TraceReader:
    def __init__(self, path: Path) -> None:
        self.path = path
        self.offset = 0
        self.pending = b""
        self.events: list[dict[str, Any]] = []

    def poll(self) -> list[dict[str, Any]]:
        if not self.path.exists():
            return []
        data = self.path.read_bytes()
        if self.offset >= len(data):
            return []
        chunk = self.pending + data[self.offset :]
        self.offset = len(data)

        new_events: list[dict[str, Any]] = []
        self.pending = b""
        for line in chunk.splitlines(keepends=True):
            if not line.endswith(b"\n"):
                self.pending = line
                continue
            line = line.strip()
            if not line:
                continue
            event = json.loads(line)
            self.events.append(event)
            new_events.append(event)

        return new_events


class TranscriptRecorder(threading.Thread):
    def __init__(self, fd: int) -> None:
        super().__init__(daemon=True)
        self.fd = fd
        self.stop_event = threading.Event()
        self.buffer = bytearray()

    def run(self) -> None:
        while not self.stop_event.is_set():
            ready, _, _ = select.select([self.fd], [], [], 0.1)
            if not ready:
                continue
            try:
                chunk = os.read(self.fd, 4096)
            except OSError:
                return
            if not chunk:
                return
            self.buffer.extend(chunk)

    def stop(self) -> bytes:
        self.stop_event.set()
        self.join(timeout=1.0)
        return bytes(self.buffer)


class ResourceSampler(threading.Thread):
    def __init__(self, pid: int, interval_s: float, window_s: float) -> None:
        super().__init__(daemon=True)
        self.pid = pid
        self.interval_s = interval_s
        self.window_s = window_s
        self.samples: list[dict[str, Any]] = []
        self.stop_event = threading.Event()
        self.start_ts = time.monotonic()

    def run(self) -> None:
        while not self.stop_event.is_set():
            elapsed = time.monotonic() - self.start_ts
            if elapsed > self.window_s:
                return
            sample = sample_process(self.pid)
            sample["elapsed_s"] = elapsed
            self.samples.append(sample)
            time.sleep(self.interval_s)

    def stop(self) -> list[dict[str, Any]]:
        self.stop_event.set()
        self.join(timeout=1.0)
        return self.samples


def load_json(path: Path) -> Any:
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def render_string(template: str, values: dict[str, Any]) -> str:
    return template.format(**values)


def render_value(value: Any, values: dict[str, Any]) -> Any:
    if isinstance(value, str):
        return render_string(value, values)
    if isinstance(value, list):
        return [render_value(item, values) for item in value]
    if isinstance(value, dict):
        return {key: render_value(item, values) for key, item in value.items()}
    return value


def encode_key(name: str) -> str:
    key = KEY_MAP.get(name.upper())
    if key is None:
        raise ValueError(f"unsupported key {name!r}")
    return key


def load_manifest(path: Path) -> dict[str, Any]:
    manifest = load_json(path)
    defaults = manifest.setdefault("defaults", {})
    defaults.setdefault("terminal_cols", DEFAULT_COLS)
    defaults.setdefault("terminal_rows", DEFAULT_ROWS)
    defaults.setdefault("step_timeout_ms", DEFAULT_STEP_TIMEOUT_MS)
    scenarios = manifest.get("scenarios", [])
    if not isinstance(scenarios, list):
        raise ValueError("manifest scenarios must be a list")
    for scenario in scenarios:
        if "name" not in scenario:
            raise ValueError("each scenario must define name")
        scenario.setdefault("steps", [])
        scenario.setdefault("optional", False)
    return manifest


def set_pty_size(fd: int, rows: int, cols: int) -> None:
    ioctl(fd, termios.TIOCSWINSZ, struct.pack("HHHH", rows, cols, 0, 0))


def parse_selected_path(path: str) -> tuple[str, str]:
    if not path:
        return "", ""
    if "/" not in path:
        return "", path
    namespace, name = path.split("/", 1)
    return namespace, name


def update_derived_vars(state: dict[str, Any], event: dict[str, Any]) -> None:
    view_name = event.get("view_name")
    if view_name:
        state["active_view"] = view_name

    selected_path = event.get("selected_path")
    if selected_path:
        state["selected_path"] = selected_path
        selected_namespace, selected_name = parse_selected_path(selected_path)
        if selected_namespace:
            state["selected_namespace"] = selected_namespace
            state.setdefault("origin_selected_namespace", selected_namespace)
        if selected_name:
            state["selected_name"] = selected_name
            state.setdefault("origin_selected_name", selected_name)
        state.setdefault("origin_selected_path", selected_path)


def marker_matches(event: dict[str, Any], step: dict[str, Any]) -> bool:
    if event.get("type") != "lifecycle_mark":
        return False
    if event.get("marker") != step["marker"]:
        return False
    for field in ("view_name", "detail_kind"):
        expected = step.get(field)
        if expected and event.get(field) != expected:
            return False
    return True


def wait_for_marker(
    trace_reader: TraceReader,
    scenario_state: dict[str, Any],
    step: dict[str, Any],
    timeout_ms: int,
    proc: subprocess.Popen[str],
) -> tuple[str, dict[str, Any] | None]:
    deadline = time.monotonic() + (timeout_ms / 1000.0)
    while time.monotonic() < deadline:
        for event in trace_reader.poll():
            update_derived_vars(scenario_state, event)
            if marker_matches(event, step):
                if step.get("allow_no_data") and event.get("rows_visible", 0) == 0:
                    return "no_data", event
                return "ok", event
        if proc.poll() is not None:
            return "process_exited", None
        time.sleep(0.1)
    return "timeout", None


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


def sample_process(pid: int) -> dict[str, Any]:
    proc = subprocess.run(
        ["ps", "-o", "rss=,%cpu=", "-p", str(pid)],
        check=False,
        capture_output=True,
        text=True,
    )
    if proc.returncode != 0:
        return {"rss_bytes": 0, "cpu_percent": 0.0}
    output = proc.stdout.strip()
    if not output:
        return {"rss_bytes": 0, "cpu_percent": 0.0}
    rss_kib, cpu_percent = output.split()
    return {
        "rss_bytes": int(float(rss_kib) * 1024),
        "cpu_percent": float(cpu_percent),
    }


def build_command(binary: Path, scenario: dict[str, Any], values: dict[str, Any]) -> list[str]:
    argv = [str(binary)] + render_value(scenario["argv"], values)
    kubeconfig = values.get("kubeconfig")
    context = values.get("context")
    if kubeconfig:
        argv.extend(["--kubeconfig", str(kubeconfig)])
    if context:
        argv.extend(["--context", str(context)])
    return argv


def create_cache_layout(base_dir: Path, mode: str) -> CacheLayout:
    home = base_dir / "home"
    config_dir = base_dir / "k9s-config"
    home.mkdir(parents=True, exist_ok=True)
    config_dir.mkdir(parents=True, exist_ok=True)
    return CacheLayout(home=home if mode == "isolated" else Path(os.environ.get("HOME", "")), config_dir=config_dir)


def should_run_scenario(scenario: dict[str, Any], values: dict[str, Any]) -> bool:
    enabled_var = scenario.get("enabled_var")
    if not enabled_var:
        return True
    return bool(values.get(enabled_var))


def run_process(
    argv: list[str],
    env: dict[str, str],
    trace_path: Path,
    scenario_name: str,
    run_id: str,
    rows: int,
    cols: int,
) -> tuple[subprocess.Popen[str], int]:
    master_fd, slave_fd = pty.openpty()
    set_pty_size(slave_fd, rows, cols)
    proc = subprocess.Popen(
        argv
        + [
            "--perf-trace-file",
            str(trace_path),
            "--perf-trace-scenario",
            scenario_name,
            "--perf-trace-run-id",
            run_id,
        ],
        stdin=slave_fd,
        stdout=slave_fd,
        stderr=slave_fd,
        env=env,
        text=False,
        preexec_fn=os.setsid,
        close_fds=True,
    )
    os.close(slave_fd)
    return proc, master_fd


def terminate_process(proc: subprocess.Popen[str], master_fd: int) -> None:
    if proc.poll() is None:
        os.write(master_fd, encode_key("CTRL_C").encode("utf-8"))
        try:
            proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            os.killpg(proc.pid, signal.SIGTERM)
            try:
                proc.wait(timeout=5)
            except subprocess.TimeoutExpired:
                os.killpg(proc.pid, signal.SIGKILL)
                proc.wait(timeout=5)
    try:
        os.close(master_fd)
    except OSError:
        pass


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
        for item in items:
            status_counts[item["status"]] += 1
        ok_items = [item for item in items if item["status"] == "ok"]

        entry: dict[str, Any] = {
            "scenario": scenario,
            "mode": mode,
            "statuses": [item["status"] for item in items],
            "status_counts": dict(sorted(status_counts.items())),
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
    lines = ["# Benchmark Summary", "", "Aggregates include only runs with status `ok`.", ""]
    for key in sorted(aggregates):
        item = aggregates[key]
        lines.append(f"## {item['scenario']} ({item['mode']})")
        lines.append("")
        status_counts = ", ".join(f"{status}={count}" for status, count in item["status_counts"].items())
        lines.append(f"- status_counts: {status_counts}")
        lines.append(f"- ok_runs: {item['ok_run_count']}/{item['total_run_count']}")
        for metric, stats in sorted(item.items()):
            if metric in {"scenario", "mode", "statuses", "status_counts", "ok_run_count", "total_run_count"}:
                continue
            lines.append(
                f"- {metric}: min={stats['min']:.2f}, median={stats['median']:.2f}, mean={stats['mean']:.2f}, max={stats['max']:.2f}"
            )
        lines.append("")
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text("\n".join(lines), encoding="utf-8")


def run_scenario_once(
    binary: Path,
    scenario: dict[str, Any],
    base_values: dict[str, Any],
    root_dir: Path,
    mode: str,
    run_index: int,
    cache_layout: CacheLayout,
    terminal_cols: int,
    terminal_rows: int,
) -> dict[str, Any]:
    run_dir = root_dir / "runs" / scenario["name"] / f"{mode}-{run_index:02d}"
    run_dir.mkdir(parents=True, exist_ok=True)
    trace_path = run_dir / "trace.jsonl"
    transcript_path = run_dir / "pty.txt"
    result_path = run_dir / "run.json"
    run_id = f"{scenario['name']}-{mode}-{run_index:02d}"

    values = dict(base_values)
    values["scenario"] = scenario["name"]
    env = os.environ.copy()
    if cache_layout.home:
        env["HOME"] = str(cache_layout.home)
    env["K9S_CONFIG_DIR"] = str(cache_layout.config_dir)
    if base_values.get("kubeconfig"):
        env["KUBECONFIG"] = str(base_values["kubeconfig"])

    argv = build_command(binary, scenario, values)
    proc, master_fd = run_process(argv, env, trace_path, scenario["name"], run_id, terminal_rows, terminal_cols)
    recorder = TranscriptRecorder(master_fd)
    recorder.start()
    sampler = ResourceSampler(proc.pid, DEFAULT_SAMPLE_INTERVAL_S, DEFAULT_SAMPLE_WINDOW_S)
    sampler.start()
    trace_reader = TraceReader(trace_path)
    scenario_state = dict(values)
    matched_wait_events: list[dict[str, Any]] = []
    terminal_wait_event: dict[str, Any] | None = None
    view_anchor_event: dict[str, Any] | None = None

    status = "ok"
    error: str | None = None
    started_at = time.time()
    try:
        for raw_step in scenario["steps"]:
            step = render_value(raw_step, scenario_state)
            if step.get("if_var") and not scenario_state.get(step["if_var"]):
                continue
            step_type = step["type"]
            if step_type == "wait_marker":
                timeout_ms = int(step.get("timeout_ms", DEFAULT_STEP_TIMEOUT_MS))
                outcome, event = wait_for_marker(trace_reader, scenario_state, step, timeout_ms, proc)
                if outcome == "ok":
                    if event:
                        update_derived_vars(scenario_state, event)
                        matched_wait_events.append(event)
                        terminal_wait_event = event
                        if event.get("view_seq") is not None:
                            view_anchor_event = event
                    continue
                if outcome == "no_data":
                    if event:
                        update_derived_vars(scenario_state, event)
                        matched_wait_events.append(event)
                        terminal_wait_event = event
                        if event.get("view_seq") is not None:
                            view_anchor_event = event
                    status = "no_data"
                    break
                if scenario.get("optional"):
                    status = "skipped"
                    error = outcome
                    break
                raise RuntimeError(f"scenario {scenario['name']} failed waiting for marker {step['marker']}: {outcome}")
            if step_type == "send_text":
                os.write(master_fd, step["text"].encode("utf-8"))
                continue
            if step_type == "send_key":
                os.write(master_fd, encode_key(step["key"]).encode("utf-8"))
                continue
            if step_type == "sleep_ms":
                time.sleep(float(step["sleep_ms"]) / 1000.0)
                continue
            raise ValueError(f"unsupported step type {step_type!r}")
    except Exception as exc:  # pragma: no cover - exercised through CLI
        status = "failed"
        error = str(exc)
    finally:
        time.sleep(0.2)
        trace_reader.poll()
        terminate_process(proc, master_fd)
        transcript = recorder.stop()
        samples = sampler.stop()
        time.sleep(0.1)
        trace_reader.poll()

    transcript_path.write_bytes(transcript)
    metrics = derive_metrics(
        trace_reader.events,
        samples,
        terminal_wait_event=terminal_wait_event,
        view_anchor_event=view_anchor_event,
    )
    result = {
        "schema_version": SCHEMA_VERSION,
        "scenario": scenario["name"],
        "mode": mode,
        "run_index": run_index,
        "status": status,
        "error": error,
        "argv": argv,
        "argv_shell": " ".join(shlex.quote(arg) for arg in argv),
        "started_at": started_at,
        "duration_s": time.time() - started_at,
        "trace_file": str(trace_path),
        "transcript_file": str(transcript_path),
        "metrics": metrics,
        "events": trace_reader.events,
        "samples": samples,
        "state": scenario_state,
        "anchors": {
            "terminal_wait_event": terminal_wait_event,
            "view_anchor_event": view_anchor_event,
            "matched_wait_events": matched_wait_events,
        },
    }
    result_path.write_text(json.dumps(result, indent=2, sort_keys=True), encoding="utf-8")
    return result


def make_cache_layout(
    cache_root: Path,
    scenario_name: str,
    mode: str,
    cache_mode: str,
    run_index: int,
) -> CacheLayout:
    if mode == "cold":
        base = Path(tempfile.mkdtemp(prefix=f"k9s-bench-{scenario_name}-{mode}-{run_index:02d}-"))
        return create_cache_layout(base, cache_mode)

    base = cache_root / scenario_name / mode
    if base.exists():
        base.mkdir(parents=True, exist_ok=True)
    else:
        base.mkdir(parents=True, exist_ok=True)
    return create_cache_layout(base, cache_mode)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Run K9s lifecycle benchmarks")
    parser.add_argument("--bin", required=True, help="path to the K9s binary")
    parser.add_argument("--label", required=True, help="artifact label")
    parser.add_argument("--vars", required=True, help="path to vars JSON")
    parser.add_argument("--manifest", default="hack/bench/scenarios.json", help="scenario manifest path")
    parser.add_argument("--scenario", action="append", default=[], help="scenario name to run")
    parser.add_argument("--cold-runs", type=int, default=1, help="number of cold runs per scenario")
    parser.add_argument("--warm-runs", type=int, default=1, help="number of warm runs per scenario")
    parser.add_argument(
        "--cache-mode",
        choices=("isolated", "user-home"),
        default="isolated",
        help="cache directory mode",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    manifest = load_manifest(Path(args.manifest))
    values = load_json(Path(args.vars))
    scenarios = manifest["scenarios"]
    if args.scenario:
        wanted = set(args.scenario)
        scenarios = [scenario for scenario in scenarios if scenario["name"] in wanted]

    timestamp = time.strftime("%Y%m%d-%H%M%S")
    root_dir = Path("artifacts/bench") / timestamp / args.label
    cache_root = root_dir / "cache"
    cache_root.mkdir(parents=True, exist_ok=True)

    terminal_cols = int(values.get("terminal_cols", manifest["defaults"]["terminal_cols"]))
    terminal_rows = int(values.get("terminal_rows", manifest["defaults"]["terminal_rows"]))

    results: list[dict[str, Any]] = []
    for scenario in scenarios:
        if not should_run_scenario(scenario, values):
            continue
        for mode, run_total in (("cold", args.cold_runs), ("warm", args.warm_runs)):
            for run_index in range(1, run_total + 1):
                cache_layout = make_cache_layout(cache_root, scenario["name"], mode, args.cache_mode, run_index)
                if args.cache_mode == "user-home":
                    cache_layout.home = Path(os.environ.get("HOME", ""))
                result = run_scenario_once(
                    binary=Path(args.bin),
                    scenario=scenario,
                    base_values=values,
                    root_dir=root_dir,
                    mode=mode,
                    run_index=run_index,
                    cache_layout=cache_layout,
                    terminal_cols=terminal_cols,
                    terminal_rows=terminal_rows,
                )
                results.append(result)

    write_csv(root_dir / "summary" / "runs.csv", results)
    write_report(root_dir / "summary" / "report.md", aggregate_runs(results))
    print(json.dumps({"artifact_root": str(root_dir), "runs": len(results)}, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
