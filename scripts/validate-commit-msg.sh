#!/usr/bin/env bash
set -euo pipefail

file="${1:?commit message file required}"
first_line="$(sed -n '1p' "$file")"

if [[ "$first_line" =~ ^(feat|fix|docs|chore|refactor|test|ops|data|build|ci|perf)(\([a-zA-Z0-9._-]+\))?:\ .+ ]]; then
  exit 0
fi

echo "Commit message must use Conventional Commits, e.g. feat: add parser"
exit 1
