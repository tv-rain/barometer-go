# Mobile Native Builds

## Decision and Scope

The production root package, including the `Client` facade, is bound with gomobile into an Android AAR and an Apple XCFramework. The root remains the only intended public package, and the build binds it directly rather than a separate example package, so every root export and transitive dependency is covered. No build tag or build-only export is used.

The libraries are compiled only. With no production platform adapters, `Initialize` fails closed, so a host cannot run a measurement through them. Successful compilation establishes only binding and compiler compatibility for the recorded build inputs. It does not establish loading or execution on a device, emulator or simulator, lifecycle behavior, conformance, safety, supported operating-system versions, platform support or readiness for distribution. The generated Java package and Objective-C names are gomobile defaults and are not a reviewed binding surface. Barometer Draft `0.1.0` status is unchanged.

## Commands and Prerequisites

Run from the repository root with Go 1.26.0 or later compatible Go, Make, and Bash:

| Command | Output | Additional requirements |
|---|---|---|
| `make mobile-smoke` | None; pure-Go Android ARM64 and iOS ARM64 cross-compilation | None |
| `make mobile-android` | `build/mobile/Barometer.aar` with `arm64-v8a` and `x86_64` libraries | JDK 17, Android SDK platform 35, NDK; `ANDROID_HOME`, `ANDROID_NDK_HOME`, and working Java tools |
| `make mobile-apple` | `build/mobile/Barometer.xcframework` with iOS device `arm64` and iOS simulator `arm64` and `x86_64` slices | macOS, full Xcode installation, selected developer directory, iPhoneOS and iPhoneSimulator SDKs |
| `python3 scripts/test-mobile-build.py android` | None; positive and negative Android build fixtures | Python 3 and Android prerequisites above |
| `python3 scripts/test-mobile-build.py apple` | None; positive and negative Apple build fixtures | Python 3 and Apple prerequisites above |

Set `MOBILE_OUT` to write elsewhere, for example `make mobile-android MOBILE_OUT=/path/to/dir`. `build/` is ignored by Git. Each library is written with a `<artifact>.build-info.txt` file that records the commit, whether the worktree was modified, the Go, gomobile and NDK or Xcode and SDK versions, the compiler targets and the slices. `bash scripts/mobile-build.sh android|apple` without an output directory builds and checks the library and keeps nothing.

The Android build uses SDK platform 35 for Java compilation and API 21 for native compilation. CI selects NDK `28.2.13676358` and JDK 17 on `ubuntu-24.04`. Set `ANDROID_HOME` to your SDK and `ANDROID_NDK_HOME` to its `ndk/28.2.13676358` directory. Install the SDK components with the SDK manager:

```sh
"$ANDROID_HOME/cmdline-tools/latest/bin/sdkmanager" "platforms;android-35" "ndk;28.2.13676358"
```

The Apple build uses deployment target 13.0. CI uses the selected Xcode on `macos-15` and records its version and SDK versions. These API levels and deployment targets are compiler inputs, not promises of minimum supported OS versions. Runner images and Xcode patch versions can change; the build-info file preserves the versions used.

Native builds are separate from `make check`, so the ordinary core validation command still requires no mobile SDK. Both native jobs run on pull requests and pushes to `main` and upload the library and its build-info file as a workflow artifact named `barometer-android` or `barometer-apple`, kept for 30 days. Workflow artifacts are not releases: they are unsigned, unversioned and not published.

## Tooling

The separate `tools/mobile` module pins `golang.org/x/mobile` to `v0.0.0-20260821190718-4776eadac327` and records its dependency checksums. Both `gomobile` and `gobind` are built from that graph into a temporary directory. No globally installed executable is used, and `gomobile init` is unnecessary. The core `go.mod` acquires no mobile tooling dependencies.

The tooling module replaces the core module with this checkout. Its module graph supplies the binding runtime as well as build tools. Future changes to core dependencies must check for version selection interactions with this graph; a passing build does not prove every downstream application's dependency graph works. Update the tooling module and checksums deliberately when those dependencies change.

The build passes `-trimpath` and strips the symbol and debug tables (`-ldflags="-s -w"`), so source, SDK and NDK paths of the build host are not embedded in the compiled libraries; Go stack traces keep function names and line numbers. Go module build information still records the checkout directory, because the tooling module replaces the core module with it. A local build therefore embeds the local checkout path; the CI libraries embed only the runner's checkout path. Do not share locally built libraries outside the team. The commands may download Go modules into the normal Go cache. They disable workspace-file discovery and remove their temporary build directory on exit, including failures. No signing identity, device, participant data, deployment endpoint, or measurement traffic is required.

## Failure Detection

A native compiler or binding-generation failure fails the build. Since gobind can omit unsupported exported functions, methods, or fields while succeeding, the script also rejects generated `// skipped` comments. This guard intentionally depends on the pinned generator's output convention and must be revalidated on a tool upgrade. The script then fails unless the library contains exactly the expected architecture slices.

The regression script creates a temporary synthetic module and runs the same native build command three times: a compatible export must build; an exported map-returning function must fail the skipped-export guard; an imported package excluded on Android and iOS must fail dependency compilation. It requires the expected diagnostic, so an unrelated setup failure cannot satisfy a negative case. It does not modify the real source checkout.

Validate changes with:

```sh
make check
make mobile-android
make mobile-apple
python3 scripts/test-mobile-build.py android
python3 scripts/test-mobile-build.py apple
cd tools/mobile
go mod tidy -diff
go mod verify
```

Run the platform-specific commands on hosts with the corresponding prerequisites; report missing validation explicitly. Before handoff, also run `git diff --check` from the repository root.

## References

- [gomobile command and native build requirements](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile).
- [gobind and supported boundary types](https://pkg.go.dev/golang.org/x/mobile/cmd/gobind).
- [Package architecture](package-architecture.md).
