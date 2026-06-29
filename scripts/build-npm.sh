#!/usr/bin/env bash
# Cross-compile the imago binary for every supported platform and lay out the
# per-platform npm packages under npm/platforms/. Run from anywhere.
set -euo pipefail

cd "$(dirname "$0")/.."
ROOT="$(pwd)"
SCOPE="@singhvibhanshu"
VERSION="$(node -p "require('./npm/imago/package.json').version")"

echo "Building imago v${VERSION} for all platforms..."

# Format: GOOS GOARCH npmPlatform npmCpu ext
targets=(
  "darwin  arm64 darwin arm64 "
  "darwin  amd64 darwin x64   "
  "linux   amd64 linux  x64   "
  "linux   arm64 linux  arm64 "
  "windows amd64 win32  x64   .exe"
)

for t in "${targets[@]}"; do
  read -r goos goarch nplat ncpu ext <<<"$t"
  pkg="imago-${nplat}-${ncpu}"
  outdir="npm/platforms/${pkg}"
  mkdir -p "${outdir}/bin"

  echo "  → ${goos}/${goarch}  →  ${SCOPE}/${pkg}"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
    go build -trimpath -ldflags "-s -w" -o "${outdir}/bin/imago${ext}" .

  cat >"${outdir}/package.json" <<EOF
{
  "name": "${SCOPE}/${pkg}",
  "version": "${VERSION}",
  "description": "imago prebuilt binary for ${nplat}-${ncpu}",
  "os": ["${nplat}"],
  "cpu": ["${ncpu}"],
  "files": ["bin/"],
  "license": "MIT"
}
EOF
done

echo "Done. Platform packages are in npm/platforms/"
