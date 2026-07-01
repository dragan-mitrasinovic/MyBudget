#!/usr/bin/env bash
# PostToolUse hook: format the file an agent just edited. See ADR-0008.
# Best-effort and dependency-free (no jq) — it never blocks the edit (always exits 0).
set -uo pipefail

input=$(cat)
# Extract tool_input.file_path without jq. Paths here contain no escaped quotes.
file=$(printf '%s' "$input" | sed -n 's/.*"file_path"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)

[ -n "$file" ] || exit 0
[ -f "$file" ] || exit 0

root=$(git -C "$(dirname "$file")" rev-parse --show-toplevel 2>/dev/null) || exit 0

case "$file" in
  *.go)
    command -v gofmt >/dev/null 2>&1 && gofmt -w "$file"
    ;;
  "$root"/web/*.ts | "$root"/web/*.html | "$root"/web/*.scss | "$root"/web/*.json | "$root"/web/*.md)
    (cd "$root/web" && npx --no-install prettier --write "$file" >/dev/null 2>&1)
    ;;
esac

exit 0
