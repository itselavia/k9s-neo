# Benchmark Harness

This directory contains the rerunnable external benchmark harness for Step 4.

## Files

- `run.py`: benchmark entrypoint
- `scenarios.json`: checked-in scenario manifest
- `vars.example.json`: template for local machine or work machine inputs
- `tests/`: Python unit tests for manifest parsing and metric derivation

## Local Setup

1. Copy `vars.example.json` to `vars.local.json`.
2. Fill in local values such as `kubeconfig`, `context`, and `big_namespace`.
3. Keep `vars.local.json` untracked.

## Usage

Run all scenarios once cold and once warm:

```bash
python3 hack/bench/run.py \
  --bin ./execs/k9s \
  --label local-smoke \
  --vars hack/bench/vars.local.json
```

Run a single scenario:

```bash
python3 hack/bench/run.py \
  --bin ./execs/k9s \
  --label local-smoke \
  --vars hack/bench/vars.local.json \
  --scenario pods_startup
```

Override run counts:

```bash
python3 hack/bench/run.py \
  --bin ./execs/k9s \
  --label baseline \
  --vars hack/bench/vars.local.json \
  --cold-runs 10 \
  --warm-runs 10
```

## Artifact Layout

Artifacts are written under:

`artifacts/bench/<timestamp>/<label>/`

Each run writes:

- a raw trace copy
- a PTY transcript
- a JSON summary

Each invocation also writes:

- `summary/runs.csv`
- `summary/report.md`

Scenario metrics are anchored to the marker sequence that actually satisfied the scenario. For multi-step flows such as `:nodes`, node drill-down, and filtered events, the harness reports the target view or action timing rather than the first startup marker in the trace.

Aggregate summaries in `summary/report.md` use only runs with status `ok`. Raw JSON and CSV artifacts still retain `no_data`, `skipped`, and `failed` runs for debugging.

## Cache Modes

- `isolated` (default):
  - cold runs use a fresh temp `HOME` and `K9S_CONFIG_DIR`
  - warm runs reuse a temp `HOME` and `K9S_CONFIG_DIR` per scenario
- `user-home`:
  - preserves the current `HOME`
  - still uses a scenario-scoped temp `K9S_CONFIG_DIR`

Use `user-home` on machines where auth helpers require the real home directory.
