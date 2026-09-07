# Versioning and Releases

## Purpose

This document defines versioning, compatibility, changelog, tagging, artifact-naming, and release conventions for the `github.com/tv-rain/barometer-go` module. It applies only to this implementation repository and its release history.

Barometer Go follows [Semantic Versioning 2.0.0](https://semver.org/), the [Go module version-numbering rules](https://go.dev/doc/modules/version-numbers), and the [Go module publishing process](https://go.dev/doc/modules/publishing). A release version communicates the maturity and compatibility of this Go module; it does not by itself establish conformance, mobile-platform support, field validation, or support for any Draft specification or registered measurement.

## Independent Version Domains

Keep the following version and status domains explicit and independent:

| Domain | Meaning | Authority |
|---|---|---|
| Go module release | Version of the source in this repository, expressed by a Git tag such as `v0.1.0-alpha.1` | This repository |
| Public specification baseline | Version of the public task, result, observation, trust, proxy, context, linkage, limit, and measurement contracts used as implementation inputs | `barometer-spec` |
| Registered measurement behavior | Identifier and version of an individual measurement contract | Public specification registry and measurement specification |
| Implementation support | Behavior backed by implementation, conformance, and applicable platform evidence for a declared release | Release notes and repository evidence |
| Platform or artifact support | Validated Go, Android, Apple, architecture, lifecycle, and packaging scope | Release notes and platform-validation evidence |

Matching numbers in different domains do not make them equivalent. For example, a Barometer Go `v0.1.0` release would not mean that every Draft `0.1.0` specification is implemented or supported. Every release note must identify the specification baseline used and separately list implemented, conformant, supported, experimental, and unsupported scope without inferring one status from another.

## Semantic Versioning Policy

Release tags use `vMAJOR.MINOR.PATCH` with an optional SemVer prerelease suffix. The leading `v` belongs to the Git tag and Go module version; the underlying semantic version is `MAJOR.MINOR.PATCH`.

### Initial Development

All `v0.x.y` versions are experimental initial-development releases. They make no general stability or backward-compatibility guarantee.

For `v0.MINOR.PATCH` releases, apply the following project convention when safety and correctness allow:

- increment PATCH for backward-compatible fixes, documentation corrections, and internal changes that do not intentionally break the declared public surface;
- increment MINOR for new public functionality or an intentional compatibility break, and document migration requirements;
- use a prerelease when the release is intended for bounded evaluation rather than ordinary consumption.

Safety, privacy, or correctness requirements take priority over compatibility. If an urgent fix must break compatibility, publish a new version permitted by this policy, describe the break prominently, and never replace or retag the affected release.

### Prereleases

Use only these ordered prerelease stages:

- `alpha.N` for incomplete, rapidly changing, or foundation-stage work;
- `beta.N` for feature-complete evaluation that still lacks release-readiness evidence;
- `rc.N` for a release candidate believed to meet the intended scope pending final validation;
- no suffix for a completed release at the declared maturity level.

`N` starts at `1` and increases for each release at the same base version and stage. Do not reuse a prerelease identifier. The first planned experimental version is `v0.1.0-alpha.1`; this convention does not itself create or authorize that release.

### Stable Releases

`v1.0.0` is the first version that may declare a stable public Go API. It requires a reviewed public API, documented compatibility commitments, appropriate conformance evidence, and an explicit release-readiness decision.

After `v1.0.0`:

- increment PATCH for backward-compatible fixes;
- increment MINOR for backward-compatible functionality and deprecations;
- increment MAJOR for incompatible public-API changes;
- provide the required `/vN` Go module-path suffix for major versions `v2` and later.

Raising the minimum Go version, materially changing dependencies, or altering serialized behavior must be called out in the changelog and release notes. Choose a version increment proportionate to the compatibility impact; do not treat a technically compiling change as automatically compatible.

## Compatibility Surface

Compatibility evaluation includes more than exported Go identifiers. Review each release for changes to:

- exported packages, types, functions, methods, fields, constants, errors, and documented semantics;
- accepted task and configuration representations;
- result and observation representations;
- registered identifiers and versions;
- cancellation, resource, retention, clearing, and failure behavior;
- Go toolchain and dependency requirements;
- any separately reviewed native or mobile artifact contract.

The public specification remains authoritative for observable behavior. A release must not use versioning language to silently redefine or weaken a specification. Specification changes follow the cross-repository review and ordering rules before the implementation release that depends on them.

## Changelog

Maintain the root [`CHANGELOG.md`](../CHANGELOG.md) in a format based on [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/).

- Keep pending notable changes under `Unreleased`.
- Use `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, and `Security` headings as applicable; omit empty headings.
- Write entries for users and integrators rather than copying commit messages.
- Identify compatibility breaks, migrations, minimum-Go-version changes, and material dependency changes explicitly.
- Use a heading such as `## [0.1.0-alpha.1] - YYYY-MM-DD` when cutting a release.
- Move released entries out of `Unreleased`; do not edit a historical release except to correct a link or clearly marked factual error.
- Add comparison links when the first release is prepared and maintain them for later releases.

The changelog records what changed. Release notes additionally record evidence and status boundaries, including the specification baseline, implemented measurements, supported platforms, validation performed, and known limitations.

## Tags and Release Records

- Create an annotated Git tag named exactly `vMAJOR.MINOR.PATCH[-PRERELEASE]` on the intended commit on `main`.
- Treat every pushed tag as immutable. Correct a released problem with a new version; never move, delete and recreate, or overwrite a published tag.
- Use the release title `Barometer Go <tag>`, for example `Barometer Go v0.1.0-alpha.1`.
- Mark alpha, beta, and release-candidate GitHub releases as prereleases.
- Do not use a mutable tag such as `latest`, and do not tag a pull-request head or an unmerged task branch.
- Signing is optional until the project adopts and documents a signing mechanism. Do not imply provenance or verification that was not performed.

## Artifact Naming

The Go module is published from its immutable repository tag and does not require a compiled binary archive. If a reviewed release produces attached assets, name them with this pattern:

```text
barometer-go_<tag>_<kind>[_<platform>_<architecture>].<extension>
```

Examples:

```text
barometer-go_v0.1.0-alpha.1_source.tar.gz
barometer-go_v0.1.0-alpha.1_checksums.txt
```

Use lowercase stable identifiers, the complete tag including its leading `v`, and an explicit platform and architecture only when the artifact is platform-specific. A future compiled Android, Apple, gomobile, or other platform asset requires its own reviewed build and validation scope before its name appears in a release. Provider-generated source archives may retain provider-generated names and are identified by the immutable release tag.

Do not publish credentials, private configuration, participant or measurement data, personal paths, unrestricted logs, or an asset whose build inputs and validation cannot be described.

## Experimental Release Process

The release owner performs the following process after the release-preparation change is reviewed and merged. The example commands show the first planned experimental version; substitute the approved version for later releases.

1. Confirm that the intended release scope is bounded, the specification baseline is identified, and every implementation or platform support statement has corresponding evidence.
2. In a release-preparation pull request, move the applicable changelog entries from `Unreleased` to `## [0.1.0-alpha.1] - YYYY-MM-DD`, update comparison links, and prepare release notes.
3. Merge the approved release-preparation pull request to `main`. Do not release directly from its feature branch.
4. Refresh a clean local `main`, verify that it matches `origin/main`, and run the canonical validation:

   ```sh
   git fetch origin
   git switch main
   git pull --ff-only origin main
   make check
   git diff --check
   git status --short
   ```

5. Review the exact commit and confirm that the working tree is clean, then create and push the annotated immutable tag:

   ```sh
   git tag -a v0.1.0-alpha.1 -m "Barometer Go v0.1.0-alpha.1 (experimental)"
   git push origin v0.1.0-alpha.1
   ```

6. Create the GitHub release `Barometer Go v0.1.0-alpha.1`, mark it as a prerelease, copy the matching changelog entries, and include:

   - status: Experimental;
   - Go module version and commit;
   - public specification baseline and its Draft, Experimental, Current, Deprecated, or unsupported status;
   - implemented and conformant measurements, or an explicit statement that there are none;
   - validated Go and platform scope, or an explicit statement that no mobile platform is supported;
   - validation performed and links to durable evidence;
   - compatibility notes, known limitations, and required migration or release ordering;
   - checksums and build provenance for any attached assets.

7. If public Go module discovery is intended and authorized, request the exact immutable version through the Go proxy and verify the returned version:

   ```sh
   GOPROXY=proxy.golang.org go list -m github.com/tv-rain/barometer-go@v0.1.0-alpha.1
   ```

8. Record the release URL, tag, commit, validation evidence, and any limitations in the release task. A successful tag or module lookup is publication evidence, not conformance, mobile validation, field validation, or proof of support beyond the release notes.

If any command, CI result, tag, asset, or module lookup differs from the intended release, stop and investigate. Do not repair a pushed release by mutating its tag or assets in place; prepare a new version and preserve the audit trail.
