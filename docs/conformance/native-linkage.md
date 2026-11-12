# Native Linkage Validation

## Purpose and Status

This document is the planned case location of the four result-linkage case families that the Go core cannot run: `CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE`, `CC-LINKAGE-CLEAR-DATA-CONTRACT-HOST`, `CC-LINKAGE-ANDROID-STORAGE` and `CC-LINKAGE-APPLE-PLATFORM-STORAGE` ([case plans](../cross-cutting-requirements.md#cc-linkage-go-core-and-mobile-binding-architecture)). For each case-plan row it names the existing validation evidence of the Android and Apple bindings in `barometer-flutter`, the Go test that covers the Go-owned part of a row, or the reason no evidence exists.

All four families stay `planned`. The evidence below ran on Android emulators, iOS simulators and one project iPhone 13 with debug builds of the plugin's test and example applications; it is evidence for those recorded runs only, not release-configuration, field or platform-support evidence. Under coordination decision 0015, Android validation runs on emulators: a row that only a physical Android device can answer points to the deferred device check AD-08 of the coordination repository's [deferred device checks](https://github.com/tv-rain/barometer/blob/main/docs/device-checks.md#android-checks) and is not claimed. Apple rows that no run covered are marked not run. The two review cases below are open. The manifest records these families at their planned locations until the next milestone update.

The evidence was recorded with the native libraries built from Go `7f5d7de7ada4709c1b2a4fdebbb4fb9b013ef918` and is read at `barometer-flutter` `8e2f4a2f6efcdbf8e4e012cc3d74a651d8b9e27a`:

- Android: [Signals, Lifecycle and Linkage Keys](https://github.com/tv-rain/barometer-flutter/blob/8e2f4a2f6efcdbf8e4e012cc3d74a651d8b9e27a/docs/android-validation.md#signals-lifecycle-and-linkage-keys) (API 21, 24 and 35 emulators), [Reinstall and Backup](https://github.com/tv-rain/barometer-flutter/blob/8e2f4a2f6efcdbf8e4e012cc3d74a651d8b9e27a/docs/android-validation.md#reinstall-and-backup) (API 24 and 35 emulators) and [Durability, Full Disk and the Nine Measurements](https://github.com/tv-rain/barometer-flutter/blob/8e2f4a2f6efcdbf8e4e012cc3d74a651d8b9e27a/docs/android-validation.md#durability-full-disk-and-the-nine-measurements); the adapter design is [Android Linkage Keys](https://github.com/tv-rain/barometer-flutter/blob/8e2f4a2f6efcdbf8e4e012cc3d74a651d8b9e27a/README.md#android-linkage-keys).
- Apple: [Signals, Lifecycle and Linkage Keys](https://github.com/tv-rain/barometer-flutter/blob/8e2f4a2f6efcdbf8e4e012cc3d74a651d8b9e27a/docs/apple-validation.md#signals-lifecycle-and-linkage-keys) (iOS 27.0 and 18.4 simulators), [Durability, Storage Full and the Nine Measurements](https://github.com/tv-rain/barometer-flutter/blob/8e2f4a2f6efcdbf8e4e012cc3d74a651d8b9e27a/docs/apple-validation.md#durability-storage-full-and-the-nine-measurements) (the full XCTest suite on the iPhone 13, so `LinkageKeysTests` and the clear-data rows of `LifecycleTests` ran against the device's Keychain) and [Observed Behavior and Gaps](https://github.com/tv-rain/barometer-flutter/blob/8e2f4a2f6efcdbf8e4e012cc3d74a651d8b9e27a/docs/apple-validation.md#observed-behavior-and-gaps); the adapter design is [Apple Linkage Keys](https://github.com/tv-rain/barometer-flutter/blob/8e2f4a2f6efcdbf8e4e012cc3d74a651d8b9e27a/README.md#apple-linkage-keys).

The native runs check the protected key through the HMAC of a fixed synthetic message, which stands in for the token `T1` of the case plans; no token or derived value was printed or recorded.

## CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE

Requirements: REQ-LINKAGE-0018 to REQ-LINKAGE-0022.

| Case-plan row | Evidence | Status |
|---|---|---|
| Automated native suites on Android and Apple release-configuration builds and through the Flutter package | The instrumented tests and XCTests ran on debug builds; the example's Dart integration test ran on emulators, simulators and the iPhone 13 without linkage, because its trusted document approves no linkage scope. | Release configuration not run; Dart with linkage not run |
| Initialize through Flutter with only the plain `result_linkage` map, create a bundle, call `clearMeasurementData`, with equivalent semantics on Android and Apple | Dart `BarometerClient.initialize` sends the configuration to the channel as one JSON string, and the plugin's `flutter test` checks that `initialize` carries only that string. With rotating linkage through the facade and the production adapters, `ClearDataDeviceTest` (Android, API 24 and 35 emulators) and the clear-data rows of `LifecycleTests` (Apple, simulators and the iPhone 13) create bundles and clear them with the same semantics. | Emulator, simulator and iPhone evidence below the channel; not run from Dart with linkage |
| A callback, native interface, key handle, byte array or platform storage object inside or alongside `result_linkage` is rejected | The channel takes only the JSON string, so no object reaches native code. The Go decoder rejects every member the schema does not define: `TestCcLinkageInstallationSecret` and `TestCcLinkageGoalsAndNonGoals` reject secret, key, seed, token, account and device members inside and beside `result_linkage` before any key operation. | Go evidence; channel shape by Dart unit test |
| A process restart within one epoch keeps the token, so the persistent adapter is connected | The control runs of `scripts/android-reinstall-backup-check.sh` (Android, API 24 and 35) and `scripts/ios-reinstall-check.sh` (Apple, iOS 27.0 and 18.4 simulators) seed in one process and verify in another without a reinstall; the verification fails on the reused key, so the key persists across processes. `LifecycleDeviceTest` reopens a scope after a process kill on Android. | Emulator and simulator evidence through the key's HMAC, not a bundle token |
| Protected-storage unavailability makes bundles omit linkage instead of falling back to an in-memory store | Below API 23 Android Keystore holds no HMAC key, so the adapter is not installed and linkage stays disabled (`LinkageKeysDeviceTest.noKeyStoreBeforeApi23`, API 21). On Apple a key that cannot be stored with its attributes fails derivation with no other store (Apple Linkage Keys). The Go rows are `TestCcLinkageStorageFailure`. A device whose key store cannot protect the key is the deferred device check AD-08. | Partial; device row deferred (AD-08) |
| A Go desktop test build marks its ephemeral in-memory implementation and loses linkage on process exit | Go ships no desktop or ephemeral implementation. The only in-memory key store is the test-scoped `internal/testkit/capability.LinkageKeys`, documented as not protected storage and never a mobile fallback. Without a bound key store linkage is disabled (`TestCapabilityPreparation`), and `barometermobile` takes the key store only from the native binding. | Recorded disposition: no such implementation exists |

### RV-CC-LINKAGE-GO-CORE-AND-MOBILE-BINDING-ARCHITECTURE

Human review case; not automated coverage. The reviewer role, reviewed artifact and checklist are in the [case plan](../cross-cutting-requirements.md#cc-linkage-go-core-and-mobile-binding-architecture). Open: no review evidence record exists.

## CC-LINKAGE-CLEAR-DATA-CONTRACT-HOST

Requirements: REQ-LINKAGE-0023 to REQ-LINKAGE-0026. The Go-owned rows are `CC-LINKAGE-CLEAR-DATA-CONTRACT`.

| Case-plan row | Evidence | Status |
|---|---|---|
| With linkage on, an active attempt and pending and acknowledged bundles, `clearMeasurementData` ends the attempt and deletes the protected key, the epoch state, the reinstall sentinel and every Barometer-owned file | `ClearDataDeviceTest` (Android, API 24 and 35): returns within 10 seconds, ends the attempt, deletes every record, the key and the sentinel, and is repeatable. `LifecycleTests` (Apple, simulators and the iPhone 13): the same, and the run handle is invalidated. `StorageFullDeviceTest` clears on a full data partition (Android). | Emulator, simulator and iPhone evidence |
| No replacement secret exists until linkage is next required | Both tests initialize again after clearing and find no key; `LinkageKeysDeviceTest` and `LinkageKeysTests` show deletion creates nothing. | Emulator, simulator and iPhone evidence |
| Network Context cache deleted or invalidated | The cache is Go-owned and in memory; clearing erases it in `persistence.Clear` (`CC-LINKAGE-CLEAR-DATA-CONTRACT`). | Go evidence |
| No upload or network request during clearing | Clearing has no transport in Go (`CC-LINKAGE-CLEAR-DATA-CONTRACT`); the native tests do not record traffic during clearing. | Go evidence only |
| Interrupted deletion reports incomplete clearing and recovers | A directory in place of the sentinel makes clearing return `clearing_incomplete`, blocks initialization and is completed by a later clear (`ClearDataDeviceTest` on Android, `LifecycleTests` on Apple). | Emulator, simulator and iPhone evidence |
| Operating-system data clearing without calling Barometer | `scripts/android-clear-storage-check.sh` runs `pm clear` as Settings "Clear storage" does: no Barometer file or key survives and initializing creates no key (API 21, 24 and 35; the control fails as expected). iOS has no application data clearing apart from uninstalling; see `CC-LINKAGE-APPLE-PLATFORM-STORAGE`. | Emulator evidence (Android) |
| Every host action that promises to remove Barometer data calls `clearMeasurementData`, and no host-owned copy remains | A host-application obligation (REQ-LINKAGE-0025): no host application is in scope, and the example application has no such action. | Host integration; not run |
| Clearing still removes every record class and the key on a physical device | The iPhone 13 ran the Apple clear-data rows. An Android physical device is the deferred device checks AD-07 and AD-08. | Apple device evidence; Android deferred (AD-07, AD-08) |

## CC-LINKAGE-ANDROID-STORAGE

Requirements: REQ-LINKAGE-0027 to REQ-LINKAGE-0031. Every row ran on API 24 and 35 emulators unless stated; emulator keys are software-backed.

| Case-plan row | Evidence | Status |
|---|---|---|
| Application-scoped Android Keystore HMAC-SHA-256 key, non-exportable where supported | `LinkageKeysDeviceTest`: a non-exportable key created on first use, derivation inside Keystore. Hardware backing on a device is AD-08. | Emulator evidence; hardware-backed key deferred (AD-08) |
| Associated non-secret state in the no-backup directory | The sentinel and the record store are under `Context.getNoBackupFilesDir()`; the reinstall and backup check finds no Barometer file there after a reinstall or restore. | Emulator evidence |
| No Barometer state in shared preferences, Auto Backup or device-to-device-eligible files, external storage or host analytics storage | The `adb backup` archive and the local transport's copy hold only the control file, no record, sentinel, reserve, lock or other Barometer file. The adapters write no shared preferences or external storage by design. | Emulator evidence for backup paths; other stores by design only |
| Backup rules explicitly exclude all Barometer state, with no cloud or cross-device synchronization | Nothing enables synchronization, and the no-backup directory is outside both backup paths, so no exclusion rule was added. REQ-LINKAGE-0028 recommends an explicit rule as defense in depth; none exists. | Gap: no explicit backup rule |
| Explicit `clearMeasurementData` | `ClearDataDeviceTest`. | Emulator evidence |
| System Settings Clear storage or Clear data | `scripts/android-clear-storage-check.sh` (`pm clear`), API 21, 24 and 35. | Emulator evidence |
| Uninstall then reinstall | `scripts/android-reinstall-backup-check.sh reinstall`: no key, sentinel or record survives, and the first derivation gives another HMAC than the seed's key. | Emulator evidence |
| Backup and restore to the same device | `adb backup` restored into a fresh and into the existing installation, and `bmgr` through the local transport (API 35): no key or sentinel returns and the first derivation uses a new key. A vendor backup transport is AD-08. | Emulator evidence; vendor transport deferred (AD-08) |
| Backup and restore to a different device, and device-to-device migration | Not run on emulators. | Deferred device check (AD-08) |
| Interrupted deletion | `ClearDataDeviceTest` (incomplete clearing completed later) and `LinkageKeysDeviceTest` (a deletion that fails and completes later). | Emulator evidence |
| `T1` is never emitted again in the same epoch after each case, and no case depends on token continuity | Each reinstall and restore verification requires another HMAC than the seed's key, and its control run fails on the reused key; the clear checks find no key after clearing. | Emulator evidence through the key's HMAC |
| Orphan recovery: only the key or only the no-backup state deleted | `LinkageKeysDeviceTest`: a key without its sentinel is deleted and replaced only when a derivation needs one, and a missing key under an existing sentinel is created only when needed. | Emulator evidence |

## CC-LINKAGE-APPLE-PLATFORM-STORAGE

Requirements: REQ-LINKAGE-0032 to REQ-LINKAGE-0036. `LinkageKeysTests` and the clear-data rows of `LifecycleTests` ran on iOS 27.0 and 18.4 simulators and on the iPhone 13 (iOS 27.2); the reinstall check ran on the simulators only, because the project does not uninstall the application on its iPhone.

| Case-plan row | Evidence | Status |
|---|---|---|
| Keychain item with `kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly` for background derivation and `kSecAttrSynchronizable` false | `LinkageKeysTests`: the key is a 32-byte item stored with these attributes, and its HMAC matches CryptoKit's. | Simulator and iPhone evidence |
| Application-private access group, unreadable by a separate test application or extension | The item uses the application's default access group, which is private unless the host shares its first keychain access group. No separate application or extension was run. | Not run |
| Non-secret reinstall sentinel in the application sandbox, not in Keychain, iCloud key-value storage, CloudKit, a shared app group or `NSUbiquitousKeyValueStore` | `LinkageKeysTests`: the sentinel is a file under Application Support excluded from backup; the adapter uses no other store by design. | Simulator and iPhone evidence for location and backup exclusion |
| Sentinel truth table: both present, sentinel absent with a surviving item, sentinel present with the item absent or invalid | `LinkageKeysTests`: a key without the sentinel, a missing key and an invalid key are each replaced on the next derivation, never before. | Simulator and iPhone evidence |
| Explicit `clearMeasurementData` | `LifecycleTests` clear-data rows. | Simulator and iPhone evidence |
| Uninstall then reinstall | `scripts/ios-reinstall-check.sh`: the Keychain item survived the uninstall and was deleted before use because the sentinel was gone; initializing created no key and the first derivation used a new key; the control failed on the reused key. | Simulator evidence; not run on the iPhone |
| Missing sentinel, missing Keychain item and interrupted deletion | `LinkageKeysTests` (truth table; a deletion that fails and completes later) and `LifecycleTests` (incomplete clearing completed later). | Simulator and iPhone evidence |
| Encrypted and unencrypted backup and restore, restored-container detection | Not run. | Not run |
| Migration to another device | Not run. | Not run |
| iCloud Keychain enabled and disabled | Not run; the item is not synchronizable. | Not run |
| `T1` is never accepted or emitted again after reinstall, restore or migration | Reinstall on the simulators, as above; restore and migration not run. | Simulator evidence for reinstall only |

### RV-CC-LINKAGE-APPLE-PLATFORM-STORAGE

Human review case; not automated coverage. The reviewer role, reviewed artifact and checklist are in the [case plan](../cross-cutting-requirements.md#cc-linkage-apple-platform-storage). Open: no review evidence record exists. The Apple validation record documents the attribute choice and the sentinel; whether a restored container can be distinguished from the original installation is not yet documented for any OS version.
