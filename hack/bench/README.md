# Benchmark Harness

This directory contains the rerunnable external benchmark harness for Step 4.

## Files

- `common.py`: shared metric derivation, aggregation, and artifact writing
- `run.py`: benchmark entrypoint
- `replay.py`: replay-fixture validation entrypoint
- `fixtures/replay/`: checked-in replay fixtures for Step 5 validation
- `scenarios.json`: checked-in scenario manifest
- `vars.example.json`: template for local machine or later optional validation-machine inputs
- `tests/`: Python unit tests for live and replay harness behavior

## Local Setup

1. Copy `vars.example.json` to `vars.local.json`.
2. Fill in local values such as `kubeconfig`, `context`, and `big_namespace`.
3. Keep `vars.local.json` untracked.

Replay validation does not require kubeconfig or cluster access.
The preferred local live path for this repo is a disposable `minikube` cluster on
macOS using the `vfkit` driver.
The current execution plan for that path lives in
`docs/development/step-6-local-baseline-plan.md`.

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

Run replay-only validation:

```bash
python3 hack/bench/replay.py \
  --fixtures hack/bench/fixtures/replay \
  --label replay-validation
```

Or use the convenience target:

```bash
make bench-replay-validate
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

Artifacts now carry `source_kind` so replay outputs cannot be confused with live benchmark results.

## Evidence Boundaries

- `run.py` produces `source_kind=live` artifacts and is the only path intended for real benchmark evidence.
- `replay.py` produces `source_kind=replay` artifacts and is validation-only.
- Replay results are never performance claims.
- The default next-step live environment is a disposable local cluster, not a real work cluster.
- Local live results are valid for before-and-after engineering comparisons on the same machine.
- Local live results should be labeled with their actual environment and should not be presented as production-cluster results by default.

## Cache Modes

- `isolated` (default):
  - cold runs use a fresh temp `HOME` and `K9S_CONFIG_DIR`
  - warm runs reuse a temp `HOME` and `K9S_CONFIG_DIR` per scenario
- `user-home`:
  - preserves the current `HOME`
  - still uses a scenario-scoped temp `K9S_CONFIG_DIR`

Use `user-home` on machines where auth helpers require the real home directory.
For local minikube benchmarking, `isolated` remains the preferred default unless an
auth helper forces `user-home`.
