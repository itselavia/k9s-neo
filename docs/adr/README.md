# Architecture Decision Records

This directory holds the durable project decisions for K9s Neo.

Use an ADR when a decision changes product scope, safety guarantees, benchmark methodology, or fork maintenance strategy.

Keep ADRs short and boring:

- write the decision that will actually guide implementation
- separate measured facts from hypotheses
- call out what is intentionally deferred
- prefer updating or superseding an ADR over burying the change in thread context

## Naming

- File name: `NNNN-short-kebab-case-title.md`
- Numbering is append-only
- Do not renumber older ADRs

## Status Values

- `Accepted`: the current project decision
- `Superseded`: replaced by a later ADR
- `Deprecated`: no longer the preferred path, but not formally replaced

## Template

```md
# NNNN Title

- Status: Accepted
- Date: YYYY-MM-DD

## Context

What problem or ambiguity forced this decision?

## Decision

What are we doing?

## Consequences

What becomes simpler, harder, narrower, or more explicit?

## Deferred Questions

What still requires measurement or later work?
```
