# 0008. Code-quality tooling and where it runs

Date: 2026-07-01
Status: Accepted

## Context

MyBudget has two codebases with different toolchains — an Angular/TypeScript web client and
a Go HTTP API — and is worked on both by hand and through coding agents. Until now,
formatting and linting were applied ad hoc: Prettier and `gofmt` were installed or
documented but never run automatically, and there was no linter at all. We want consistent,
low-noise quality gates that apply regardless of who — or what — is editing.

## Decision

Adopt a small, layered setup with a single source of truth for the commands.

- **Formatting** — Prettier for the web (`.ts/.html/.scss/.json/.md`), `gofmt` for Go.
- **Linting** — `angular-eslint` (flat config, run via `ng lint`) for the web; `golangci-lint`
  for Go with a deliberately minimal starter set (`govet`, `staticcheck`, `errcheck`,
  `ineffassign`, `unused`). `go vet` is also runnable standalone.
- **Commands** live in `web/package.json` scripts and a top-level `Makefile` (`make fmt`,
  `make lint`, `make vet`, `make test`, `make check`). Every enforcement layer calls these
  rather than duplicating command logic.
- **Enforcement layers**, fast → comprehensive:
  - Editor (`.vscode/`) — format-on-save + inline diagnostics.
  - Agent hook (`.claude/settings.json`) — auto-formats files an agent edits.
  - Pre-commit (**lefthook**) — formats and lints staged files; works with no agent involved.
  - CI is intentionally deferred (see Consequences).

## Alternatives considered

- **husky + lint-staged** for the commit hook — the JS-ecosystem default, but Node-centric
  and awkward for the Go half of a polyglot repo. lefthook is a single binary that handles
  both languages in parallel from one config.
- **`pre-commit`** (the Python framework) — large ecosystem, but pulls in a Python toolchain
  we otherwise don't need.
- **staticcheck alone** instead of golangci-lint — lighter, but golangci-lint bundles
  staticcheck plus errcheck/ineffassign/govet behind one binary and one config.
- **Agent instructions only** — advisory, not deterministic, and misses hand edits entirely.

## Consequences

- (+) Consistent formatting/linting for every contributor and every agent, defined once.
- (+) Fast local feedback; the commit hook catches issues with no agent in the loop.
- (−) New developer-machine dependencies to install: `golangci-lint`, `lefthook`, and the
  eslint toolchain (added via `ng add angular-eslint`).
- (−) Pre-commit hooks can be bypassed with `--no-verify` and must be installed per clone
  (`lefthook install`). Without CI there is no unbypassable backstop — accepted for now
  given a small, trusted set of contributors; revisit with a new ADR (and a CI workflow)
  when the team or the risk grows.
