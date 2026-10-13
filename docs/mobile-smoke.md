# Mobile Compilation Smoke Checks

## Decision and Scope

DEV-465 checks whether the current root Go package can generate and compile Android and Apple bindings. The root remains the only intended public package. Because it has no production exports yet, `mobile_smoke.go` adds one string-to-string `MobileSmoke` function only under the `barometer_mobile_smoke` build tag. This synthetic echo is a build fixture, not a production API, measurement, serialized protocol, or native adapter. Normal builds exclude it. Final API signatures remain subject to their own review.

The native checks bind the actual root package, including its other exports and transitive dependencies; they do not bind only a disconnected example package. When the production facade exists, replace the synthetic export through reviewed work while retaining coverage of that facade. Do not use the smoke tag for releases or application integration.

Successful compilation establishes only binding/compiler compatibility for the selected build inputs. It does not establish device execution, lifecycle behavior, conformance, safety, platform support, simulator support, or readiness for distribution. Barometer Draft `0.1.0` status is unchanged.

## Commands and Prerequisites

Run from the repository root with Go 1.26.0 or later compatible Go, Make, and Bash:

| Command | Check | Additional requirements |
|---|---|---|
| `make mobile-smoke` | Existing pure-Go Android ARM64 and iOS ARM64 cross-compilation | None |
| `make mobile-smoke-android` | Generate Java bindings and compile an Android ARM64 AAR | JDK 17, Android SDK platform 35, NDK; `ANDROID_HOME`, `ANDROID_NDK_HOME`, and working Java tools |
| `make mobile-smoke-apple` | Generate Objective-C bindings and compile an iOS device ARM64 XCFramework | macOS, full Xcode installation, selected developer directory, iPhoneOS SDK |
| `python3 scripts/test-mobile-smoke.py android` | Positive and negative Android build fixtures | Python 3 and Android prerequisites above |
| `python3 scripts/test-mobile-smoke.py apple` | Positive and negative Apple build fixtures | Python 3 and Apple prerequisites above |

The Android check uses SDK platform 35 for Java compilation and API 21 for native compilation. CI selects NDK `28.2.13676358` and JDK 17 on `ubuntu-24.04`. Set `ANDROID_HOME` to your SDK and `ANDROID_NDK_HOME` to its `ndk/28.2.13676358` directory. Install the SDK components with the SDK manager:

```sh
"$ANDROID_HOME/cmdline-tools/latest/bin/sdkmanager" "platforms;android-35" "ndk;28.2.13676358"
```

The Apple check uses `ios/arm64` with deployment target 13.0. CI uses the selected Xcode on `macos-15` and records its version and SDK version. These API/deployment targets are compiler inputs, not promises of minimum supported OS versions. Runner images and Xcode patch versions can change; preserve the reported versions with validation evidence.

Native checks are separate from `make check`, so the ordinary core validation command still requires no mobile SDK. Both native jobs run on pull requests and pushes to `main`. Branch protection configuration and a green hosted CI run are separate from adding the workflow.

## Tooling and Outputs

The separate `tools/mobile` module pins `golang.org/x/mobile` to `v0.0.0-20260821190718-4776eadac327` and records its dependency checksums. Both `gomobile` and `gobind` are built from that graph into a temporary directory. No globally installed executable is used, and `gomobile init` is unnecessary. The core `go.mod` acquires no mobile tooling dependencies.

The tooling module replaces the core module with this checkout. Its module graph supplies the binding runtime as well as build tools. Future changes to core dependencies must check for version selection interactions with this graph; smoke success does not prove every downstream application's dependency graph works. Update the tooling module and checksums deliberately when those dependencies change.

The commands may download Go modules into the normal Go cache. They disable workspace-file discovery, create temporary build outputs, and remove those outputs on exit, including failures. The AAR and XCFramework are not published or retained as release artifacts. No signing identity, device, participant data, deployment endpoint, or measurement traffic is required.

## Failure Detection

A native compiler or binding-generation failure fails the check. Since gobind can omit unsupported exported functions, methods, or fields while succeeding, the script also rejects generated `// skipped` comments. This guard intentionally depends on the pinned generator's output convention and must be revalidated on a tool upgrade.

The regression script creates a temporary synthetic module and runs the same native build command three times: a compatible export must build; an exported map-returning function must fail the skipped-export guard; an imported package excluded on Android and iOS must fail dependency compilation. It requires the expected diagnostic, so an unrelated setup failure cannot satisfy a negative case. It does not modify the real source checkout.

Validate changes with:

```sh
make check
make mobile-smoke-android
make mobile-smoke-apple
python3 scripts/test-mobile-smoke.py android
python3 scripts/test-mobile-smoke.py apple
cd tools/mobile
go mod tidy -diff
go mod verify
```

Run the platform-specific commands on hosts with the corresponding prerequisites; report missing validation explicitly. Before handoff, also run `git diff --check` from the repository root.

## References

- [gomobile command and native build requirements](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile).
- [gobind and supported boundary types](https://pkg.go.dev/golang.org/x/mobile/cmd/gobind).
- [Package architecture](package-architecture.md).
