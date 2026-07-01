#!/usr/bin/env bash
# Publish all imago npm packages. Run `npm login` first.
# Platform packages must be published BEFORE the main package, because the main
# package lists them as (optional) dependencies.
#
# Auth:
#   - If 2FA is on, this script prompts for a fresh one-time code (OTP) before
#     each package. Just open your authenticator app and type the 6 digits.
#   - If you publish with a token/automation (e.g. in CI), leave the OTP blank.
set -euo pipefail

cd "$(dirname "$0")/.."

if [ ! -d npm/platforms ]; then
  echo "No built binaries found. Run scripts/build-npm.sh first." >&2
  exit 1
fi

publish_one() {
  local dir="${1%/}" # strip any trailing slash
  # Skip if this exact name@version is already on npm (makes re-runs idempotent,
  # e.g. when a tag is re-pushed, avoids a spurious "cannot publish over
  # previously published version" failure).
  local name version
  name="$(node -p "require('./${dir}/package.json').name")"
  version="$(node -p "require('./${dir}/package.json').version")"
  if npm view "${name}@${version}" version >/dev/null 2>&1; then
    echo "    already published, skipping: ${name}@${version}"
    return 0
  fi
  local otp=""
  if [ -t 0 ]; then
    read -r -p "OTP for ${dir} (blank if using a token): " otp
  fi
  if [ -n "$otp" ]; then
    (cd "$dir" && npm publish --access public --otp="$otp")
  else
    (cd "$dir" && npm publish --access public)
  fi
}

echo "Publishing platform packages..."
for d in npm/platforms/*/; do
  echo "  → ${d}"
  publish_one "$d"
done

echo "Publishing main package..."
publish_one "npm/imago"

echo "All done. Try: npm install -g @singhvibhanshu/imago"
