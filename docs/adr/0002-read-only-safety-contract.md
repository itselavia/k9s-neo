# 0002 Read-Only Safety Contract

- Status: Accepted
- Date: 2026-03-22

## Context

Upstream K9s already exposes a `--readonly` mode, but that is not strong enough for K9s Neo's product promise. The fork is meant for safe triage on sensitive clusters, and the end state must be defensibly read-only even when the operator's credentials are powerful.

This ADR turns the read-only promise into a concrete safety contract that later implementation and tests can enforce.

## Decision

### End-State Guarantee

K9s Neo v0 must not expose any reachable path that mutates cluster state or launches cluster-side behavior.

Hidden affordances are not sufficient. Forbidden behavior must be blocked in code, unreachable from normal UI flows, and covered by tests.

### Forbidden Capabilities

The final product must not support:

- edit
- apply
- replace
- patch
- delete
- scale
- restart
- rollback
- port-forward
- shell
- exec
- attach
- copy or transfer to or from containers
- helper pods
- debug containers
- node shell flows
- plugin loading or execution
- any command path that creates, updates, deletes, or launches workloads as part of normal product behavior

### Allowed Capabilities

The following are in scope:

- get, list, and watch style resource reads
- logs and previous logs
- events
- YAML inspection
- describe inspection
- local filtering, searching, sorting, and screen rendering
- local copies of names or metadata
- local benchmark and trace artifact generation

### Enforcement Rules

The read-only guarantee must not depend on a cooperative environment alone.

- hiding a hotkey is not enough
- hiding a menu item is not enough
- hiding a command alias is not enough
- disabling a path in the default config is not enough if the code path remains reachable in product mode
- RBAC is defense in depth, not the primary guarantee

Staged implementation is acceptable while the fork is under construction, but the v0 release bar requires the strict contract to be the shipped behavior.

### Plugin Policy

Plugins are out of scope for v0.

- bundled plugin execution is forbidden
- user-provided plugin execution is forbidden
- later reconsideration requires a new ADR because plugins expand both safety and maintenance surface

### Network And API Semantics

This safety contract is about cluster mutation and cluster-side launch behavior. It does not require every internal API call to be literally a Kubernetes read verb, because some non-mutating control-plane checks may still exist during transition. Those calls are performance and failure-mode concerns, not allowed end-user features.

### Verification Mapping

| Contract statement | Future proof method | Owning step |
| --- | --- | --- |
| Mutating actions are unavailable from the UI action registry | Unit tests for action maps and command dispatch | Strict read-only hardening |
| Shell, exec, attach, copy, and node-shell flows are unreachable | Unit tests around pod and node action wiring | Strict read-only hardening |
| Port-forward is unavailable | Command and action tests | Strict read-only hardening |
| Plugin execution is unavailable even if plugin files exist | Config and startup tests | Strict read-only hardening |
| No forbidden path relies on RBAC denial to stay safe | Fake-client tests with privileged credentials | Strict read-only hardening |
| Allowed read workflows still work | Smoke tests for logs, events, YAML, and describe flows | Read-only hardening plus benchmark harness |
| Help text and docs do not over-promise unsupported behaviors | Documentation review during release prep | Public artifact pass |

## Consequences

- The product promise becomes testable instead of aspirational.
- Some upstream features that are harmless for general K9s users are explicitly out of scope here.
- Safety simplification is treated as first-class value even when hot-path performance impact is small.
- Future changes that reopen plugins or mutation-capable flows must go through an explicit decision, not drift in accidentally.

## Deferred Questions

- Whether a limited local-only extension mechanism is ever worth reintroducing.
- Whether any current internal helper path needs redesign rather than simple removal to preserve the triage loop cleanly.
