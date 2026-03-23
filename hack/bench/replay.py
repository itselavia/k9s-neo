#!/usr/bin/env python3

from __future__ import annotations

import argparse
import json
import sys
import time
from pathlib import Path
from typing import Any

SCRIPT_DIR = Path(__file__).resolve().parent
if str(SCRIPT_DIR) not in sys.path:
    sys.path.insert(0, str(SCRIPT_DIR))

from common import (  # noqa: E402
    FIXTURE_SCHEMA_VERSION,
    SCHEMA_VERSION,
    SOURCE_KIND_REPLAY,
    aggregate_runs,
    derive_metrics,
    load_json,
    write_csv,
    write_report,
    write_run_artifacts,
)

REQUIRED_FIXTURE_FIELDS = {
    "fixture_schema_version",
    "scenario",
    "mode",
    "run_index",
    "status",
    "error",
    "events",
    "samples",
    "anchors",
}

REQUIRED_ANCHOR_FIELDS = {
    "terminal_wait_event",
    "view_anchor_event",
    "matched_wait_events",
}


def fixture_paths(path: Path) -> list[Path]:
    if path.is_file():
        return [path]
    if not path.is_dir():
        raise ValueError(f"fixture path not found: {path}")
    return sorted(child for child in path.iterdir() if child.suffix == ".json" and child.is_file())


def load_fixture(path: Path) -> dict[str, Any]:
    fixture = load_json(path)
    missing = REQUIRED_FIXTURE_FIELDS - fixture.keys()
    if missing:
        raise ValueError(f"{path} missing required fields: {', '.join(sorted(missing))}")
    if fixture["fixture_schema_version"] != FIXTURE_SCHEMA_VERSION:
        raise ValueError(
            f"{path} uses unsupported fixture schema {fixture['fixture_schema_version']} (expected {FIXTURE_SCHEMA_VERSION})"
        )
    if not isinstance(fixture["events"], list):
        raise ValueError(f"{path} events must be a list")
    if not isinstance(fixture["samples"], list):
        raise ValueError(f"{path} samples must be a list")
    if not isinstance(fixture["anchors"], dict):
        raise ValueError(f"{path} anchors must be an object")
    anchor_missing = REQUIRED_ANCHOR_FIELDS - fixture["anchors"].keys()
    if anchor_missing:
        raise ValueError(f"{path} anchors missing required fields: {', '.join(sorted(anchor_missing))}")
    return fixture


def build_replay_result(fixture: dict[str, Any]) -> dict[str, Any]:
    anchors = fixture["anchors"]
    metrics = derive_metrics(
        fixture["events"],
        fixture["samples"],
        terminal_wait_event=anchors["terminal_wait_event"],
        view_anchor_event=anchors["view_anchor_event"],
    )
    return {
        "schema_version": SCHEMA_VERSION,
        "source_kind": SOURCE_KIND_REPLAY,
        "scenario": fixture["scenario"],
        "mode": fixture["mode"],
        "run_index": fixture["run_index"],
        "status": fixture["status"],
        "error": fixture["error"],
        "argv": [],
        "argv_shell": "",
        "started_at": fixture.get("started_at", 0.0),
        "duration_s": fixture.get("duration_s", 0.0),
        "metrics": metrics,
        "events": fixture["events"],
        "samples": fixture["samples"],
        "state": fixture.get("state", {}),
        "anchors": anchors,
    }


def run_replay(fixtures_path: Path, label: str, artifacts_base: Path = Path("artifacts/bench")) -> tuple[Path, list[dict[str, Any]]]:
    timestamp = time.strftime("%Y%m%d-%H%M%S")
    root_dir = artifacts_base / timestamp / label
    results: list[dict[str, Any]] = []

    for fixture_path in fixture_paths(fixtures_path):
        fixture = load_fixture(fixture_path)
        result = build_replay_result(fixture)
        result["fixture_file"] = str(fixture_path)
        run_dir = root_dir / "runs" / result["scenario"] / f"{result['mode']}-{int(result['run_index']):02d}"
        results.append(write_run_artifacts(run_dir, result))

    write_csv(root_dir / "summary" / "runs.csv", results)
    write_report(root_dir / "summary" / "report.md", aggregate_runs(results))
    return root_dir, results


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Replay benchmark fixtures into standard benchmark artifacts")
    parser.add_argument("--fixtures", required=True, help="fixture JSON file or directory")
    parser.add_argument("--label", required=True, help="artifact label")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    root_dir, results = run_replay(Path(args.fixtures), args.label)
    print(json.dumps({"artifact_root": str(root_dir), "runs": len(results)}, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
