# 0016. Development-stage migration lifecycle

Date: 2026-07-11
Status: Accepted

## Context

MyBudget is in active development, has never been deployed, and has no real users or persistent
user data to keep compatible. The schema is expected to change frequently before prerelease, and
recording every removed or corrected design as a forward migration would make the eventual
history describe abandoned experiments rather than useful deployment transitions.

ADR-0013 established the embedded ordered SQLite migration runner and made every applied
migration immediately immutable. The runner remains useful, but that immutability boundary is
premature for the current lifecycle stage.

## Decision

Keep the embedded, contiguous, transactional migration runner from ADR-0013, but treat migration
history as mutable until the owner explicitly declares MyBudget prerelease.

During active development:

- Database additions, including new tables, columns, constraints, and indexes, receive dedicated
  numbered migrations.
- Destructive or corrective changes are folded into the migration that defines the affected
  schema. Obsolete columns and tables are removed from that definition instead of receiving drop
  migrations. Obsolete later migrations may be removed, with the remaining sequence kept
  contiguous.
- Local development databases are disposable and may need to be recreated after migration
  history is rewritten. Compatibility migrations are not added solely to preserve local data.

When the owner declares the project prerelease, all migration files then present become
immutable. From that point onward, every schema or data change uses a new forward migration,
including removals, renames, and corrections, and persistent databases must be backed up before
schema changes are applied.

## Alternatives considered

- **Immutable migrations from the first development database** — production-like, but preserves
  abandoned schema experiments and creates noisy history before any compatibility obligation
  exists.
- **No migrations until prerelease** — keeps the baseline compact, but gives up early testing of
  migration ordering, atomicity, and additive schema evolution.
- **Squash all migrations once at prerelease** — possible, but a separate one-time operation is
  unnecessary if destructive churn is already folded into its defining migrations as work
  proceeds.

## Consequences

- (+) The prerelease migration history will focus on intentional additions rather than temporary
  columns, tables, and corrective drops.
- (+) Additive migrations and the migration runner continue to be exercised during development.
- (+) The transition to immutable, forward-only production history has an explicit owner-controlled
  boundary.
- (-) Existing local databases can become incompatible with rewritten migration history and must
  be recreated.
- (-) Until prerelease, migration files do not provide a permanent audit trail of every schema
  experiment.
