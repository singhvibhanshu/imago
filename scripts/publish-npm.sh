#!/usr/bin/env bash
# Publish all imago npm packages. Run `npm login` first.
# Platform packages must be published BEFORE the main package, because the main
# package lists them as (optional) dependencies.
set -euo pipefail

cd "$(dirname "$0")/.."

if [ ! -d npm/platforms ]; then
  echo "No built binaries found. Run scripts/build-npm.sh first." >&2
  exit 1
fi

echo "Publishing platform packages..."
for d in npm/platforms/*/; do
  echo "  → ${d}"
  (cd "$d" && npm publish --access public)
done

echo "Publishing main package..."
(cd npm/imago && npm publish --access public)

echo "All done. Try: npm install -g @singhvibhanshu/imago"
