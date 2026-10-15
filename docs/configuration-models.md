# Initialization and Configuration Models

## Authority and Scope

DEV-486 adopts Draft `0.1.0` at immutable public `tv-rain/barometer-spec` revision `05c1364af4f9a604b617fca45f060949b557d265`: [Proxy](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/library/proxy.md), [Network Context](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/library/network-context.md), [Result Linkage](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/library/result-linkage.md), [Task Trust](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/library/task-trust.md), and [Runtime JSON Decoding](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/data-formats/json-decoding.md). The [defined Go facade](public-api.md) supplies the closed initialization container: optional `proxy`, `network_context`, and `result_linkage` sections only.

This is internal model and preparation work. No root-package export, client lifecycle, measurement, network lookup, secret creation, native implementation, persistence backend or public serialized schema is introduced. Successful preparation is not activation or admission and does not establish full runtime conformance or platform support.

## Ownership and Call Order

`internal/config` owns the host configuration schema, pure validation, immutable snapshots, trusted deployment approval inputs and initialization preparation. It depends on the shared strict JSON decoder, narrow platform contracts and the standard library. It does not import task decoding, admission, execution or concrete adapters. `internal/platform` owns the consumed trusted clock and protected linkage-key interfaces. They have no public host/Flutter registration API. Domain models remain independent of these packages.

An internal reviewed adapter constructs `Approvals` with exact endpoint URLs, coarse proxy connection codes, linkage scopes and a local linkage enablement policy. These lists are not loaded from task JSON, arbitrary Flutter values or remote responses. Review must establish that entries are deployment-scoped and non-user-specific; syntax cannot prove this. Approval is configuration selection only and grants no issuer, target, replay or execution authority. A compromised trusted adapter remains a residual risk.

`NewDecoder` accepts only the explicit trusted containing contract `0.1.0`. Configuration JSON itself remains versionless. Missing, malformed or unsupported integration versions fail with no decoder and no fallback; JSON `version` and `schema_version` fields are unknown fields. The decoder copies all approval lists and exposes no mutation method. Callers must not mutate their inputs concurrently with construction; independent decoding calls may run concurrently after construction.

`Decode` first applies the strict closed schemas, then validates every section against its semantic constraints and trusted approvals. It returns a valid `Snapshot` only after all sections succeed. Every error returns the zero, invalid snapshot, with no state change, capability call or network activity. Snapshot fields are private; getters return independent string/integer values. Standard JSON unmarshalling cannot construct a valid snapshot.

`WithProxy` validates a complete standalone proxy object and returns a replacement snapshot. Any error returns the unchanged previous snapshot. It requires the decoder that created the snapshot, preventing replacement from mixing separately installed approval contexts. `WithoutProxy` returns a copy with the proxy removed. Neither operation mutates the original or other configuration sections. These are value transformations, not implementations of `SetProxy`/`ClearProxy`: the eventual runtime must serialize transitions, block and cancel old-generation attempts, perform bounded cleanup and invalidate caches before publishing a successful replacement. Invalid input must be checked before starting those effects. Policy or endpoint changes require a new initialization; these models provide no in-place update method for them.

## Defaults, Validation and Local Bounds

Omitted proxy means no proxy. Omitted Network Context means access context only, with no egress endpoint. Omitted result linkage means `none`. Explicit null never means omission. Linkage in `none` mode rejects scope and rotation fields. Rotating linkage requires both, with a 1–128 character lowercase ASCII deployment scope and 86,400–604,800 second period. An approved scope is still required when trusted local policy disables requested linkage; complete invalid input never becomes valid through policy defaulting. A disabled request normalizes to `none` without keeping its scope or rotation.

Proxy hosts must be loopback IP literals or exactly `localhost`, ports are integers from 1 through 65,535, and connection codes contain 1–64 lowercase ASCII letters, digits, `.`, `_` or `-` and must match trusted approval. IPv6 literals use address form without URL brackets or zone identifiers. Proxy authentication and extra properties reject.

Endpoint URLs require HTTPS, no credentials or fragment, valid port syntax and exact deployment approval. Literal IPs reject obvious local/private/non-unicast destinations. DNS hosts use fully qualified ASCII names (international names use their ASCII encoding); localhost and local multicast-DNS names reject. Endpoint paths and any queries must be approved as fixed, non-user-specific values. Parsing does not resolve names or prove a public route: the eventual transport still requires reviewed destination/DNS validation, path snapshots, TLS, redirect prohibition and request minimization. Approval does not authorize weakening these controls.

The public contracts do not register a complete configuration byte maximum. This implementation selects a finite local envelope: 16,384 original encoded UTF-8 bytes including whitespace, 8 value levels and 64 value nodes. A URL is at most 2,048 decoded bytes. Each trusted approval list contains at most 64 entries. Trusted lists are validated before copying and are finite in both count and entry length. These local choices require review as part of this change; they are not new normative maxima or measured device limits. The complete JSON byte bound applies before parsing, regardless of unknown fields or claimed versions.

All errors are fixed `config` or shared `jsondecode` sentinel classes. They never contain input, unknown field names, URLs, scope, secrets or raw adapter errors. A failure is not a Measurement Result or evidence of censorship.

## Capability Preparation and Remaining Gates

`Prepare` requires a valid snapshot and an internally supplied trusted `platform.Clock`; absent and typed-nil clock implementations fail closed. It captures capability references without invoking them. No ordinary system-clock fallback is installed. References are private copies, while concrete implementations retain responsibility for concurrency safety, trusted acquisition and honoring finite caller deadlines and cancellation.

`platform.LinkageKeyStore` exposes bounded internal HMAC derivation and deletion without a raw-key getter. Protected key creation is lazy during future derivation, never preparation. Missing or typed-nil protected storage makes linkage unavailable alone; measurement configuration may still be prepared. Disabled linkage returns no usable key capability. Future unavailable-clock or derivation errors must also omit linkage and must never trigger insecure storage fallback. Preparation neither derives nor deletes a key and cannot validate native protection merely from an interface value.

Delivery verification and exact-byte/generation binding, immutable authenticated local policy/target registrations, monotonic policy updates, replay persistence, retention, clearing, normalized device/network acquisition and their additional interfaces remain with their runtime tasks. Preparation supplies no successful substitute for these required capabilities. Native key protection, reinstall consistency and lifecycle behavior require concrete adapters and device evidence. No public API or cross-repository merge ordering changes are required; the shared decoder and public contract used here are already merged.

## Synthetic Evidence

| Area | Cases |
|---|---|
| Defaults, valid sections, boundary values | `internal/config/config_test.go`: `TestDefaultsAndComplete`, `TestVersionsApprovalsAndBounds` |
| Complete-model rejection, strict JSON, null/type/integer/Unicode/duplicate failures | `TestAtomicRejectionCorpus`, `TestClosedSchemasAndTrustSeparation` |
| Unsupported containing version, byte/depth bounds, trusted approvals and local disabling | `TestVersionsApprovalsAndBounds` |
| Copied approval lists, copied input/accessors, unchanged old snapshot on failed/successful replacement, clearing | `TestSnapshotOwnershipAndReplacement` |
| Shared immutable decoder and replacement reads | `TestConcurrentImmutableDecoding`, race detector |
| Required clock, typed-nil ports, unavailable linkage storage, no eager key operations | `TestCapabilityPreparation` |
| Task-field injection and opaque labels cannot supply reviewed host configuration | `internal/taskdecode/config_boundary_test.go`: `TestHostConfigurationCannotComeFromTask` |
| Deterministic bounded atomic outcomes under mutation | `FuzzConfiguration` |

Run `make check`, `git diff --check`, and `python3 -B ../barometer-spec/scripts/check_requirements.py --consumer .`. A finite fuzz check uses `go test ./internal/config -run='^$' -fuzz='^FuzzConfiguration$' -fuzztime=20s -parallel=2`. These synthetic checks do not establish runtime activation, native storage, mobile resource behavior, binding agreement, field validation or release support.

## Local Validation — October 15, 2026

The prepared change passed `make check`: formatting, host compilation, unit tests, race detection, static analysis, module consistency/verification, diff whitespace and pure-Go Android ARM64/iOS ARM64 compilation. Requirement-reference structure passed against the explicit public checkout; allocation history, coverage and semantic conformance were not assessed by that checker. Changed-file whitespace and local documentation links passed separately, including newly added files.

The final 20-second `FuzzConfiguration` run with two workers passed with 50,613 executions, checking deterministic snapshots/errors and atomic rejection. This finite mutation test is regression evidence, not exhaustive proof. No native artifact, device test, CI result, runtime activation, release or platform-support claim is established. The coordination and specification repositories remain unchanged.
