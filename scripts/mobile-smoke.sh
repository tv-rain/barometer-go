#!/usr/bin/env bash
set -euo pipefail

target=${1:-}
case "$target" in
  android)
    : "${ANDROID_HOME:?Set ANDROID_HOME to the Android SDK directory}"
    : "${ANDROID_NDK_HOME:?Set ANDROID_NDK_HOME to the selected NDK directory}"
    platform=android/arm64; artifact=Barometer.aar
    flags=(-androidapi=21 -bootclasspath "$ANDROID_HOME/platforms/android-35/android.jar")
    ;;
  apple) platform=ios/arm64; artifact=Barometer.xcframework; flags=(-iosversion=13.0) ;;
  *) echo 'Usage: bash scripts/mobile-smoke.sh android|apple' >&2; exit 2 ;;
esac

root=$(cd "$(dirname "$0")/.." && pwd)
work=$(mktemp -d "${TMPDIR:-/tmp}/barometer-mobile.XXXXXX")
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/bin" "$work/tmp"
export TMPDIR="$work/tmp"
export GOWORK=off
cd "$root/tools/mobile"

# Use the same pinned revision for both executables; never an ambient gobind.
go build -mod=readonly -o "$work/bin/gomobile" golang.org/x/mobile/cmd/gomobile
go build -mod=readonly -o "$work/bin/gobind" golang.org/x/mobile/cmd/gobind
export PATH="$work/bin:$PATH"
go version
go list -m golang.org/x/mobile

# Keep generated sources until the skip check, then remove all temporary output.
gomobile bind -work -target="$platform" "${flags[@]}" -tags=barometer_mobile_smoke \
  -o "$work/$artifact" github.com/tv-rain/barometer-go

# gobind sometimes reports unsupported exports only as comments and exits zero.
status=0
grep -R -n --include='*.java' --include='*.h' --include='*.m' --include='*.go' \
  '// skipped ' "$work/tmp" || status=$?
if [ "$status" -eq 0 ]; then
  echo 'Mobile smoke failed: unsupported exports were skipped by gobind.' >&2
  exit 1
fi
if [ "$status" -ne 1 ]; then
  echo 'Mobile smoke failed: could not inspect generated bindings.' >&2
  exit "$status"
fi
test -e "$work/$artifact"
echo "Mobile smoke passed: $platform ($artifact; temporary, not a release artifact)."
