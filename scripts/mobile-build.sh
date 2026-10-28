#!/usr/bin/env bash
set -euo pipefail

target=${1:-}
out=${2:-}
case "$target" in
  android)
    : "${ANDROID_HOME:?Set ANDROID_HOME to the Android SDK directory}"
    : "${ANDROID_NDK_HOME:?Set ANDROID_NDK_HOME to the selected NDK directory}"
    platform=android/arm64,android/amd64; artifact=Barometer.aar
    flags=(-androidapi=21 -bootclasspath "$ANDROID_HOME/platforms/android-35/android.jar")
    ;;
  apple) platform=ios,iossimulator; artifact=Barometer.xcframework; flags=(-iosversion=13.0) ;;
  *) echo 'Usage: bash scripts/mobile-build.sh android|apple [output-directory]' >&2; exit 2 ;;
esac

root=$(cd "$(dirname "$0")/.." && pwd)
if [ -n "$out" ]; then
  mkdir -p "$out"
  out=$(cd "$out" && pwd)
fi
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
# -trimpath and stripped symbol and debug tables keep build-host paths out of
# the compiled libraries.
gomobile bind -work -trimpath -ldflags="-s -w" -target="$platform" "${flags[@]}" \
  -o "$work/$artifact" github.com/tv-rain/barometer-go

# gobind sometimes reports unsupported exports only as comments and exits zero.
status=0
grep -R -n --include='*.java' --include='*.h' --include='*.m' --include='*.go' \
  '// skipped ' "$work/tmp" || status=$?
if [ "$status" -eq 0 ]; then
  echo 'Mobile build failed: unsupported exports were skipped by gobind.' >&2
  exit 1
fi
if [ "$status" -ne 1 ]; then
  echo 'Mobile build failed: could not inspect generated bindings.' >&2
  exit "$status"
fi
test -e "$work/$artifact"

# Every requested architecture must be present in the output.
if [ "$target" = android ]; then
  slices=$(unzip -Z1 "$work/$artifact" 'jni/*/libgojni.so' | sort)
  expected=$'jni/arm64-v8a/libgojni.so\njni/x86_64/libgojni.so'
else
  slices=$(find "$work/$artifact" -mindepth 1 -maxdepth 1 -type d -exec basename {} \; | sort)
  expected=$'ios-arm64\nios-arm64_x86_64-simulator'
fi
if [ "$slices" != "$expected" ]; then
  printf 'Mobile build failed: unexpected slices:\n%s\n' "$slices" >&2
  exit 1
fi

if [ -n "$out" ]; then
  rm -rf "${out:?}/$artifact"
  cp -R "$work/$artifact" "$out/$artifact"
  info="$out/$artifact.build-info.txt"
  {
    echo "artifact: $artifact"
    echo "targets: $platform"
    echo "commit: $(git -C "$root" rev-parse HEAD)"
    if [ -n "$(git -C "$root" status --porcelain)" ]; then echo "worktree: modified"; fi
    echo "go: $(go env GOVERSION)"
    echo "gomobile: $(go list -m -f '{{.Version}}' golang.org/x/mobile)"
    if [ "$target" = android ]; then
      echo "ndk: $(sed -n 's/^Pkg.Revision *= *//p' "$ANDROID_NDK_HOME/source.properties")"
      echo "android-api: 21 (native), 35 (Java bootclasspath)"
      echo "javac: $(javac -version 2>&1)"
    else
      echo "xcode: $(xcodebuild -version | paste -sd ' ' -)"
      echo "iphoneos-sdk: $(xcrun --sdk iphoneos --show-sdk-version)"
      echo "iphonesimulator-sdk: $(xcrun --sdk iphonesimulator --show-sdk-version)"
      echo "ios-deployment-target: 13.0"
    fi
    echo "slices:"
    sed 's/^/  /' <<<"$slices"
    echo "status: compiled only; not a release, not platform support"
  } >"$info"
  echo "Mobile build passed: $platform ($out/$artifact)."
else
  echo "Mobile build passed: $platform ($artifact; temporary, not kept)."
fi
