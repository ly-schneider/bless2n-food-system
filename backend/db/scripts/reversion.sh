#!/usr/bin/env bash
# backend/db/scripts/reversion.sh
# Usage:
#   ./reversion.sh V012__insert_into_event_roles.sql [<migrations-dir>]

set -euo pipefail

# ── 1. arguments ────────────────────────────────────────────────────────────
if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "Usage: $0 <new-file> [migrations-dir]" >&2
  exit 1
fi

target_file=$1
migrations_dir=${2:-"$(cd "$(dirname "$0")/../migrations" && pwd)"}

[[ -f "$migrations_dir/$target_file" ]] || {
  echo "File not found: $migrations_dir/$target_file" >&2; exit 1;
}

cd "$migrations_dir"

# ── 2. extract anchor version ───────────────────────────────────────────────
if [[ $target_file =~ ^V([0-9]{3})__ ]]; then
  anchor=${BASH_REMATCH[1]}
else
  echo "Filename must start with V###__" >&2; exit 1
fi

# ── 3. find migrations to bump (descending order) ───────────────────────────
find . -maxdepth 1 -type f -name 'V[0-9][0-9][0-9]*__*.sql' \
      ! -samefile "$target_file" |
  sed -E 's|^\./||' |
  awk -F'__' -v a="$anchor" '
        { ver = substr($1,2); if (ver >= a) print ver ":" $0 }' |
  sort -t: -k1,1nr |
  while IFS=: read -r old_ver file; do
    new_ver=$(printf "%03d" $((10#$old_ver + 1)))
    new_name="V${new_ver}${file:4}"     # keep “__rest_of_name.sql”
    echo "mv '$file' '$new_name'"
    mv "$file" "$new_name"
  done

echo "✅ Re-versioning complete in $migrations_dir"
