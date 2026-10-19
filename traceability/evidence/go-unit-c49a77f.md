# Go Unit-Test Evidence at c49a77f

DEV-597 records executed synthetic Go unit tests that partially evidence requirements in the cross-cutting manifest after adopting the GO-SPEC-ISSUE-008 disposition. Each section below is the conformance evidence for one cross-cutting manifest case. The measurement manifest keeps its earlier baseline and its [evidence at c5f4e92](go-unit-c5f4e92.md).

## Execution Record

- Implementation: `https://github.com/tv-rain/barometer-go` commit `c49a77fc35d0190992c272ba7447467ab8ff675c` (DEV-597 commit A), exported with `git archive` into an empty directory outside any checkout.
- Specification baseline: `https://github.com/tv-rain/barometer-spec` commit `d93aff93cbc17e7c065fb40d8f6e2e01a697bf35` (Draft `0.1.0`, public PR #14 merge). Between the previous baseline `5a37e82865a08922c4821c7318656286933421e5` and this commit, the only normative text change is the addition of REQ-TRUST-0036, REQ-TRUST-0037 and REQ-TRUST-0038; the text of every other requirement listed here is unchanged.
- Environment: Go 1.26.0 (`GOTOOLCHAIN=auto`, `GOFLAGS` empty, `CGO_ENABLED=1`), `darwin/arm64`, default test timeout, no race detector.
- Command, from the export root: `go test -count=1 -json ./...`
- Outcome: exit status 0; all 123 top-level tests passed, none failed or skipped. Each case below names one passing top-level test; a top-level pass includes all of its subtests.
- Executed on October 19, 2026. Inputs are the synthetic, in-repository fixtures at the implementation commit; no participant, measurement or private data is involved.

## Carried-Forward Cases

The 61 cases carried from the [4575079 report](go-unit-4575079.md) were re-executed at this baseline and keep their recorded aspects. Between `4575079` and `c49a77f` the only Go change is that `task.AdmissionCode` gains `Unsupported` (`task_unsupported`) and `TestAdmissionErrorVocabulary` in `internal/model/task` covers it; every other test body is unchanged. That case now also records an aspect of REQ-TRUST-0036. No case is added.

## Limits

Every case is partial evidence. A passing case shows only the stated aspects of each linked requirement; the notes in the manifest record the remaining obligations. The tests exercise pure-Go model, decoding, validation, result and bundle code on one host. They do not execute the planned conformance families, admission under local policy, a measurement runner, a runtime producer, native bindings, a device or a field deployment, and they establish no implemented requirement, released API, platform support or field validation. In particular, no admission implementation returns `task_unsupported` or applies the REQ-TRUST-0037 condition-to-class table.

## Cases

<a id="UT.internal.bundle.TestBundleValidationPrecedence"></a>

### UT.internal.bundle.TestBundleValidationPrecedence

- Case: `UT.internal.bundle.TestBundleValidationPrecedence`
- Result: passed
- Test: `TestBundleValidationPrecedence` in `internal/bundle/validation_test.go` (package `internal/bundle`)
- Reproduce one case: `go test -count=1 -run '^TestBundleValidationPrecedence$' ./internal/bundle`
- Scope: Existing synthetic Go unit test TestBundleValidationPrecedence in package internal/bundle, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-BUNDLE-0002: The 'invalid member rejects all' subtest asserts that a bundle containing a zero (unvalidated) result.Value is rejected with ErrValue and no partial bundle escapes.
  - REQ-BUNDLE-0022: The 'invalid member rejects all' subtest asserts that a zero result.Value alongside a valid result causes ErrValue, with no partial bundle.
  - REQ-LIMIT-0021: Asserts a 17-result bundle is rejected with ErrBounds.

<a id="UT.internal.bundle.TestContractMembershipAndIdentity"></a>

### UT.internal.bundle.TestContractMembershipAndIdentity

- Case: `UT.internal.bundle.TestContractMembershipAndIdentity`
- Result: passed
- Test: `TestContractMembershipAndIdentity` in `internal/bundle/bundle_test.go` (package `internal/bundle`)
- Reproduce one case: `go test -count=1 -run '^TestContractMembershipAndIdentity$' ./internal/bundle`
- Scope: Existing synthetic Go unit test TestContractMembershipAndIdentity in package internal/bundle, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-BUNDLE-0001: Builds 64 single-result bundles and asserts no bundle_id repeats, and asserts each ID strictly decodes to 16 non-zero bytes.
  - REQ-BUNDLE-0002: Asserts that Build rejects a results slice containing a zero result.Value together with a valid one.
  - REQ-BUNDLE-0003: Asserts that an unlinked bundle's wire encoding has exactly three top-level members, with schema_version "0.1.0".
  - REQ-BUNDLE-0004: Asserts that New rejects versions "", "0.1", and "0.2.0"; Build rejects 0 and 17 results and accepts 1 and 16; the output has exactly three top-level members with schema_version "0.1.0"; and bundle_id strictly decodes to 16 non-zero bytes.
  - REQ-BUNDLE-0006: Asserts that 64 consecutive Builds produce distinct bundle_id values.
  - REQ-BUNDLE-0017: Asserts that bundle_id strictly base64url-decodes to 16 bytes, is not all zeroes, and re-encodes identically.
  - REQ-BUNDLE-0021: Asserts that Build rejects 0 and 17 results and accepts 1 and 16 with membership preserved.
  - REQ-BUNDLE-0022: Asserts that Build rejects a results slice containing a zero result.Value.
  - REQ-BUNDLE-0029: Asserts that an unlinked bundle's wire encoding has exactly three top-level members (schema_version, bundle_id, results), so no extra bundle-level fields are present.

<a id="UT.internal.bundle.TestExactEncodedBoundaries"></a>

### UT.internal.bundle.TestExactEncodedBoundaries

- Case: `UT.internal.bundle.TestExactEncodedBoundaries`
- Result: passed
- Test: `TestExactEncodedBoundaries` in `internal/bundle/bundle_test.go` (package `internal/bundle`)
- Reproduce one case: `go test -count=1 -run '^TestExactEncodedBoundaries$' ./internal/bundle`
- Scope: Existing synthetic Go unit test TestExactEncodedBoundaries in package internal/bundle, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-BUNDLE-0004: Asserts that bundles of exactly 262,143 and 262,144 encoded bytes are accepted and that 262,145 bytes is rejected with ErrBounds, with and without linkage.
  - REQ-BUNDLE-0021: With and without linkage, asserts that encoded bundles of 262,143 and 262,144 bytes are produced exactly and that 262,145 bytes is rejected with ErrBounds.
  - REQ-LIMIT-0021: Asserts bundles encoding to 262143 and 262144 bytes are built and a 262145-byte bundle is rejected with ErrBounds, with and without linkage.
  - REQ-LIMIT-0022: Asserts a 262145-byte bundle is rejected with ErrBounds and exposes no partial bundle, while accepted bundles have exactly the target size and valid JSON.

<a id="UT.internal.bundle.TestImmutableRetryBytesAndViews"></a>

### UT.internal.bundle.TestImmutableRetryBytesAndViews

- Case: `UT.internal.bundle.TestImmutableRetryBytesAndViews`
- Result: passed
- Test: `TestImmutableRetryBytesAndViews` in `internal/bundle/bundle_test.go` (package `internal/bundle`)
- Reproduce one case: `go test -count=1 -run '^TestImmutableRetryBytesAndViews$' ./internal/bundle`
- Scope: Existing synthetic Go unit test TestImmutableRetryBytesAndViews in package internal/bundle, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-BUNDLE-0007: Asserts that MarshalJSON bytes stay identical, including under concurrent reads, after the source linkage, result slice, View, and returned buffer are mutated, and that rebuilding with changed or removed linkage yields a different ID.
  - REQ-BUNDLE-0013: Asserts that the bundled result's execution.state remains not_run after the input slice and view are mutated.
  - REQ-BUNDLE-0016: Asserts that rebuilding with a linkage whose epoch and token changed, or with nil linkage, yields a bundle ID different from the original.

<a id="UT.internal.bundle.TestMixedResultMembershipIsPreserved"></a>

### UT.internal.bundle.TestMixedResultMembershipIsPreserved

- Case: `UT.internal.bundle.TestMixedResultMembershipIsPreserved`
- Result: passed
- Test: `TestMixedResultMembershipIsPreserved` in `internal/bundle/bundle_test.go` (package `internal/bundle`)
- Reproduce one case: `go test -count=1 -run '^TestMixedResultMembershipIsPreserved$' ./internal/bundle`
- Scope: Existing synthetic Go unit test TestMixedResultMembershipIsPreserved in package internal/bundle, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-BUNDLE-0002: Bundles negative-completed, incomplete, and not-run results produced by result.New/Build and asserts that each results[i] is byte-identical to that result's MarshalJSON output.
  - REQ-BUNDLE-0013: Bundles a completed timeout result, an incomplete cancelled_by_host result, and a not_run result, then asserts that each bundled entry equals its original encoding and that view mutation does not alter the stored outcome.
  - REQ-BUNDLE-0014: Bundles two results that each carry a decoded network_context and asserts that every bundled entry is byte-identical to the result's own encoding, so the contexts are not merged or deduplicated.

<a id="UT.internal.bundle.TestTrustedOptionalLinkage"></a>

### UT.internal.bundle.TestTrustedOptionalLinkage

- Case: `UT.internal.bundle.TestTrustedOptionalLinkage`
- Result: passed
- Test: `TestTrustedOptionalLinkage` in `internal/bundle/bundle_test.go` (package `internal/bundle`)
- Reproduce one case: `go test -count=1 -run '^TestTrustedOptionalLinkage$' ./internal/bundle`
- Scope: Existing synthetic Go unit test TestTrustedOptionalLinkage in package internal/bundle, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-BUNDLE-0003: Asserts that bundles built with nil linkage on builders with linkage disabled and enabled contain no 'linkage' bytes.
  - REQ-BUNDLE-0004: Asserts that linkage is absent unless supplied and, when present, has mode, scope, epoch, and token fields.
  - REQ-BUNDLE-0006: Asserts that json.Unmarshal of wire input containing a bundle_id into bundle.Value fails, so serialized input cannot supply the identifier.
  - REQ-BUNDLE-0010: For epochs 0, 2948, 2^53+1, and MaxInt64, asserts that the encoded linkage has mode rotating_installation, the configured scope, the given epoch, and a token that strictly decodes to the supplied 32 bytes, and that epoch -1 is rejected.
  - REQ-BUNDLE-0011: Asserts that json.Unmarshal into bundle.Value fails for linkage inputs with mode 'stable' plus user_id, and with previous_token, and that the produced linkage mode is rotating_installation.
  - REQ-BUNDLE-0018: For a builder initialized from configuration '{}', asserts that the bundle contains no 'linkage' bytes and that Build with a linkage projection is rejected.
  - REQ-BUNDLE-0023: Asserts that Build with a linkage projection is rejected, with no partial bundle, for a builder with linkage disabled and for one configured for rotating_installation but lacking LinkageKeys, and that nil linkage yields no 'linkage' bytes.
  - REQ-BUNDLE-0025: Asserts that linkage scopes 'other', empty, and 129 'a' characters are rejected; that the emitted scope equals the configured 'org.example.production'; and that serialized wire input cannot be unmarshalled into a bundle Value.
  - REQ-BUNDLE-0026: Asserts that the emitted linkage.token strictly base64url-decodes to bytes equal to the supplied 32-byte token.
  - REQ-LINKAGE-0006: Asserts that a builder without local linkage configuration, or without a key store, rejects supplied linkage; that a mismatched scope is rejected; and that wire JSON cannot be unmarshalled into a bundle Value.

<a id="UT.internal.config.TestAtomicRejectionCorpus"></a>

### UT.internal.config.TestAtomicRejectionCorpus

- Case: `UT.internal.config.TestAtomicRejectionCorpus`
- Result: passed
- Test: `TestAtomicRejectionCorpus` in `internal/config/config_test.go` (package `internal/config`)
- Reproduce one case: `go test -count=1 -run '^TestAtomicRejectionCorpus$' ./internal/config`
- Scope: Existing synthetic Go unit test TestAtomicRejectionCorpus in package internal/config, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0004: Rejects host configuration with a Proxy key, an unknown top-level field, and other invalid sections, publishing no snapshot.
  - REQ-JSON-0005: Rejects null proxy, network_context and result_linkage sections, an empty proxy object, and a quoted port.
  - REQ-JSON-0006: Rejects configuration proxy ports 1.0, 1e0 and -0.
  - REQ-JSON-0010: Asserts each malformed or invalid configuration returns an error and a zero, invalid Snapshot.

<a id="UT.internal.config.TestClosedSchemasAndTrustSeparation"></a>

### UT.internal.config.TestClosedSchemasAndTrustSeparation

- Case: `UT.internal.config.TestClosedSchemasAndTrustSeparation`
- Result: passed
- Test: `TestClosedSchemasAndTrustSeparation` in `internal/config/config_test.go` (package `internal/config`)
- Reproduce one case: `go test -count=1 -run '^TestClosedSchemasAndTrustSeparation$' ./internal/config`
- Scope: Existing synthetic Go unit test TestClosedSchemasAndTrustSeparation in package internal/config, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0004: Rejects configuration containing trust or version field names at top level and inside proxy, network_context and result_linkage sections.

<a id="UT.internal.config.TestConfigurationValidationPrecedence"></a>

### UT.internal.config.TestConfigurationValidationPrecedence

- Case: `UT.internal.config.TestConfigurationValidationPrecedence`
- Result: passed
- Test: `TestConfigurationValidationPrecedence` in `internal/config/validation_test.go` (package `internal/config`)
- Reproduce one case: `go test -count=1 -run '^TestConfigurationValidationPrecedence$' ./internal/config`
- Scope: Existing synthetic Go unit test TestConfigurationValidationPrecedence in package internal/config, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0010: Asserts oversize, wrong-shape and unapproved configurations return no valid snapshot with a stable error class.
  - REQ-JSON-0011: Asserts the same expected error class, at most 128 bytes, is returned over 20 runs for each invalid configuration case.

<a id="UT.internal.config.TestVersionsApprovalsAndBounds"></a>

### UT.internal.config.TestVersionsApprovalsAndBounds

- Case: `UT.internal.config.TestVersionsApprovalsAndBounds`
- Result: passed
- Test: `TestVersionsApprovalsAndBounds` in `internal/config/config_test.go` (package `internal/config`)
- Reproduce one case: `go test -count=1 -run '^TestVersionsApprovalsAndBounds$' ./internal/config`
- Scope: Existing synthetic Go unit test TestVersionsApprovalsAndBounds in package internal/config, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0008: Asserts config.NewDecoder rejects contracts empty, latest, 0.1, 0.2.0 and 0.1.0 with trailing space with ErrVersion and returns no decoder.
  - REQ-JSON-0009: Asserts configuration input at MaxBytes is accepted and one byte over is rejected with ErrBounds.

<a id="UT.internal.execution.TestAllTerminationReasons"></a>

### UT.internal.execution.TestAllTerminationReasons

- Case: `UT.internal.execution.TestAllTerminationReasons`
- Result: passed
- Test: `TestAllTerminationReasons` in `internal/execution/attempt_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestAllTerminationReasons$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestAllTerminationReasons in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0010: Stops pending and started attempts with each proxy_* execution reason (proxy_not_configured, proxy_not_permitted, proxy_path_unsupported, proxy_unavailable, proxy_connect_failed, proxy_reconfigured) and asserts the resulting not_run/incomplete phase retains that explicit reason.
  - REQ-RESULT-0020: Stops pending and running attempts with each of the 15 registered reasons and asserts not_run/incomplete phase with the reason retained.
  - REQ-RESULT-0021: Asserts a second Stop after the first termination returns ErrTransition and leaves the first reason unchanged.
  - REQ-SCHEDULE-0027: Asserts only platform_suspended, network_context_unavailable, proxy_unavailable and internal_error yield a retry candidate, while cancelled_by_host, resource_limit, constraints_unsatisfied, local_policy_rejected, proxy_reconfigured and other reasons do not.

<a id="UT.internal.execution.TestProtocolOutcomesAndInvalidReasons"></a>

### UT.internal.execution.TestProtocolOutcomesAndInvalidReasons

- Case: `UT.internal.execution.TestProtocolOutcomesAndInvalidReasons`
- Result: passed
- Test: `TestProtocolOutcomesAndInvalidReasons` in `internal/execution/attempt_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestProtocolOutcomesAndInvalidReasons$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestProtocolOutcomesAndInvalidReasons in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0001: Asserts protocol-outcome strings (nxdomain, tcp_timeout, connection_refused, tls_validation_failed, unexpected_status, etc.) are rejected as execution reasons and that a started attempt still Completes with no reason and no retry candidate.
  - REQ-RESULT-0020: Asserts Stop rejects unregistered strings including protocol outcomes, an empty string, and 'internal_error ' with ErrReason.
  - REQ-RESULT-0021: Asserts protocol-outcome strings cannot be used as stop reasons and a running attempt still completes.
  - REQ-SCHEDULE-0027: Asserts measurement outcome strings are rejected as reasons and completed attempts never become retryable.

<a id="UT.internal.jsondecode.TestBounds"></a>

### UT.internal.jsondecode.TestBounds

- Case: `UT.internal.jsondecode.TestBounds`
- Result: passed
- Test: `TestBounds` in `internal/jsondecode/decode_test.go` (package `internal/jsondecode`)
- Reproduce one case: `go test -count=1 -run '^TestBounds$' ./internal/jsondecode`
- Scope: Existing synthetic Go unit test TestBounds in package internal/jsondecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0009: Asserts 64-byte input is accepted and 65-byte input rejected with ErrBounds even when it is invalid syntax, nested object MaxBytes counts original whitespace, and MaxCount, depth and node limits reject one-over values.
  - REQ-JSON-0010: Asserts depth 1024 and 128 unclosed arrays are rejected with ErrBounds, and over-limit byte, count, depth and node inputs fail.

<a id="UT.internal.jsondecode.TestConfigurationAndNoFallback"></a>

### UT.internal.jsondecode.TestConfigurationAndNoFallback

- Case: `UT.internal.jsondecode.TestConfigurationAndNoFallback`
- Result: passed
- Test: `TestConfigurationAndNoFallback` in `internal/jsondecode/decode_test.go` (package `internal/jsondecode`)
- Reproduce one case: `go test -count=1 -run '^TestConfigurationAndNoFallback$' ./internal/jsondecode`
- Scope: Existing synthetic Go unit test TestConfigurationAndNoFallback in package internal/jsondecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0008: Asserts a 0.1.0 input missing a field required only by the 0.1.0 schema fails with ErrRequired instead of matching a registered 0.2.0 schema.

<a id="UT.internal.jsondecode.TestErrorsAndDeterminism"></a>

### UT.internal.jsondecode.TestErrorsAndDeterminism

- Case: `UT.internal.jsondecode.TestErrorsAndDeterminism`
- Result: passed
- Test: `TestErrorsAndDeterminism` in `internal/jsondecode/decode_test.go` (package `internal/jsondecode`)
- Reproduce one case: `go test -count=1 -run '^TestErrorsAndDeterminism$' ./internal/jsondecode`
- Scope: Existing synthetic Go unit test TestErrorsAndDeterminism in package internal/jsondecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0011: Asserts errors for inputs containing a synthetic-password name and SYNTHETIC-TOKEN value are at most 32 bytes, contain neither string, and are identical over 50 runs, and required-field errors are stable.

<a id="UT.internal.jsondecode.TestPositiveAndPresence"></a>

### UT.internal.jsondecode.TestPositiveAndPresence

- Case: `UT.internal.jsondecode.TestPositiveAndPresence`
- Result: passed
- Test: `TestPositiveAndPresence` in `internal/jsondecode/decode_test.go` (package `internal/jsondecode`)
- Reproduce one case: `go test -count=1 -run '^TestPositiveAndPresence$' ./internal/jsondecode`
- Scope: Existing synthetic Go unit test TestPositiveAndPresence in package internal/jsondecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0001: Accepts an object surrounded by space, tab, carriage return and line feed.
  - REQ-JSON-0002: Accepts label names é and e plus combining acute as distinct keys, showing no Unicode normalization is applied.
  - REQ-JSON-0003: Accepts canonically equivalent but distinct label keys é and e plus combining acute in the same object.
  - REQ-JSON-0005: Asserts a missing optional member is absent from Lookup, explicit null is retained as Null, an empty string is retained as empty, and zero is retained as 0.
  - REQ-JSON-0007: Asserts 9007199254740993, int64 extremes and a 100-digit value within its declared bound decode with their exact token text, and values just outside int64 bounds are rejected with ErrRange.

<a id="UT.internal.jsondecode.TestRejectionCorpus"></a>

### UT.internal.jsondecode.TestRejectionCorpus

- Case: `UT.internal.jsondecode.TestRejectionCorpus`
- Result: passed
- Test: `TestRejectionCorpus` in `internal/jsondecode/decode_test.go` (package `internal/jsondecode`)
- Reproduce one case: `go test -count=1 -run '^TestRejectionCorpus$' ./internal/jsondecode`
- Scope: Existing synthetic Go unit test TestRejectionCorpus in package internal/jsondecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0001: Rejects empty input, trailing value, trailing content, comment, trailing comma, byte-order mark, truncated input, invalid number syntax, and top-level array or scalar against an object schema.
  - REQ-JSON-0002: Rejects invalid UTF-8, a malformed \u escape, unpaired high and low surrogate escapes, and unpaired surrogates in a member name and a nested label value with ErrText.
  - REQ-JSON-0003: Rejects duplicate count members with equal and conflicting values, count versus co\u0075nt, duplicate label keys, and duplicate name/n\u0061me in an array element object with ErrDuplicate.
  - REQ-JSON-0004: Rejects an unknown top-level field, a case-variant Version field, and an unknown field inside an array element object with ErrField.
  - REQ-JSON-0005: Rejects top-level null, a missing required count, count null, and count given as a string, boolean, array or object.
  - REQ-JSON-0006: Rejects integer field values -0, 1.0, 1e0, 1E+0 and 1.5 with ErrInteger and +1, 01, 1., 1e and --1 with ErrSyntax.
  - REQ-JSON-0007: Rejects count values -1 and 18446744073709551616 outside the declared 0..18446744073709551615 range with ErrRange.
  - REQ-JSON-0008: Asserts DecodeSelected rejects inputs with missing version fields, null, integer, empty or 9.0 versions, and an unregistered test_id with ErrVersion.
  - REQ-LABEL-0010: The 'duplicate-label' case asserts that the JSON parser rejects {"labels":{"x":"a","x":"a"}} with the duplicate error class.

<a id="UT.internal.jsondecode.TestValueContainers"></a>

### UT.internal.jsondecode.TestValueContainers

- Case: `UT.internal.jsondecode.TestValueContainers`
- Result: passed
- Test: `TestValueContainers` in `internal/jsondecode/decode_test.go` (package `internal/jsondecode`)
- Reproduce one case: `go test -count=1 -run '^TestValueContainers$' ./internal/jsondecode`
- Scope: Existing synthetic Go unit test TestValueContainers in package internal/jsondecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0002: Asserts a valid surrogate-pair escape decodes to the single scalar U+1F600 and a literal U+FFFD is preserved.

<a id="UT.internal.model.observation.TestAcceptedProducerMappings"></a>

### UT.internal.model.observation.TestAcceptedProducerMappings

- Case: `UT.internal.model.observation.TestAcceptedProducerMappings`
- Result: passed
- Test: `TestAcceptedProducerMappings` in `internal/model/observation/dns_tls_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestAcceptedProducerMappings$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestAcceptedProducerMappings in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0044: Asserts ValidateALPNOffer with contract 0.2.0 is rejected.
  - REQ-OBS-0045: Asserts DNSResponseCode, NegotiatedTLSVersion, NegotiatedALPN and PeerAlertClass return no value when exposed is false while mapping exposed out-of-subset values to OTHER/unknown/other.
  - REQ-OBS-0046: Asserts DNSResponseCode maps 0..5 to NOERROR..REFUSED and every other uint16 to OTHER when exposed, and returns no value for exposed=false.
  - REQ-OBS-0047: Asserts single-response observations with FORMERR, NOTIMP or OTHER are accepted only with failure inconclusive (SERVFAIL/NXDOMAIN/REFUSED with their specific failures) and are rejected with no_data.
  - REQ-OBS-0048: Asserts NegotiatedTLSVersion returns TLS1.2 for 0x0303, TLS1.3 for 0x0304, unknown for 0, 0x0301, 0x0305 and 65535, and no value when exposed is false.
  - REQ-OBS-0049: Asserts NegotiatedALPN returns http/1.1, h2, dot unchanged and unknown for h3/private, no value when not exposed or empty, and ValidateALPNOffer accepts the three identifiers but rejects unknown, h2c, h3, H2, private, duplicate h2 and nine entries.
  - REQ-OBS-0050: Asserts PeerAlertClass maps all 256 codes to the specified classes with other as default, and returns no value for exposed=false.

<a id="UT.internal.model.observation.TestAccessOnlySnapshot"></a>

### UT.internal.model.observation.TestAccessOnlySnapshot

- Case: `UT.internal.model.observation.TestAccessOnlySnapshot`
- Result: passed
- Test: `TestAccessOnlySnapshot` in `internal/model/observation/context_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestAccessOnlySnapshot$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestAccessOnlySnapshot in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-CONTEXT-FORMAT-0004: Asserts the all-unknown six-field access object decodes and re-serializes byte-identically, and that a null network_changed, an added device_id field, or trailing data is rejected.
  - REQ-CONTEXT-FORMAT-0005: Asserts DecodeAccess rejects an access object with an added device_id field.
  - REQ-OBS-0006: Asserts the access-only object decodes via DecodeAccess, serializes byte-identically, binds to its generation, and can be projected from a full NetworkContext.

<a id="UT.internal.model.observation.TestContextNormalizationStaysWithinASCIICodes"></a>

### UT.internal.model.observation.TestContextNormalizationStaysWithinASCIICodes

- Case: `UT.internal.model.observation.TestContextNormalizationStaysWithinASCIICodes`
- Result: passed
- Test: `TestContextNormalizationStaysWithinASCIICodes` in `internal/model/observation/validation_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestContextNormalizationStaysWithinASCIICodes$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestContextNormalizationStaysWithinASCIICodes in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-CONTEXT-FORMAT-0006: Asserts US/us/uS with matching region decode to US and US-SC and round-trip idempotently, while codes containing U+017F or leading whitespace are rejected with ErrValue.
  - REQ-CONTEXT-FORMAT-0018: Asserts us/us-sc and uS/Us-sC decode to US/US-SC, normalization is idempotent on round trip, and U+017F or whitespace-containing codes are rejected.
  - REQ-CONTEXT-FORMAT-0023: Asserts contexts rejected for non-ASCII codes cannot be marshalled.
  - REQ-JSON-0002: Asserts country/region codes containing U+017F or leading whitespace are rejected rather than case-folded, while ASCII mixed case is normalized.

<a id="UT.internal.model.observation.TestContextPathGenerationAndPrivacy"></a>

### UT.internal.model.observation.TestContextPathGenerationAndPrivacy

- Case: `UT.internal.model.observation.TestContextPathGenerationAndPrivacy`
- Result: passed
- Test: `TestContextPathGenerationAndPrivacy` in `internal/model/observation/context_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestContextPathGenerationAndPrivacy$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestContextPathGenerationAndPrivacy in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-CONTEXT-0017: Asserts a decoded direct context fails Matches for another network generation and for a proxy path, a proxy context fails Matches for a new proxy generation, a changed-network context fails Matches, and direct-path input is rejected when decoded against a proxy Path.
  - REQ-CONTEXT-FORMAT-0001: Asserts injected ip, network_generation, proxy_generation, device_id, ssid and context_id fields are rejected at top level, in access and in egress, that serialized context contains none of them, and that a Path cannot be JSON-serialized.
  - REQ-CONTEXT-FORMAT-0005: Asserts context decoding rejects ip, ssid, resolver_address, operator, provider and device_id fields injected into the access object.
  - REQ-CONTEXT-FORMAT-0016: Asserts context decoding rejects ip, endpoint_url, proxy_host, network_generation, proxy_generation, provider, operator and context_id at every object level, that serialized context omits them, and that json.Marshal of a Path fails.
  - REQ-CONTEXT-FORMAT-0017: Asserts unknown fields such as ip, ssid and context_id are rejected at top level, in access and in egress.

<a id="UT.internal.model.observation.TestContextRejectionAndNormalization"></a>

### UT.internal.model.observation.TestContextRejectionAndNormalization

- Case: `UT.internal.model.observation.TestContextRejectionAndNormalization`
- Result: passed
- Test: `TestContextRejectionAndNormalization` in `internal/model/observation/context_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestContextRejectionAndNormalization$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestContextRejectionAndNormalization in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-CONTEXT-FORMAT-0004: Asserts context decoding rejects a missing network_changed, a null network_changed, and an unregistered network_type value "private".
  - REQ-CONTEXT-FORMAT-0006: Asserts rejection of an unknown lookup_source, absent or unapproved classification_version, unapproved country or region, region inconsistent with country, connection_code on direct, and ASN values 0, 4294967296 and 1.0.
  - REQ-CONTEXT-FORMAT-0007: Asserts an unavailable egress object that still carries country, region, ASN and classification_version is rejected.
  - REQ-CONTEXT-FORMAT-0008: Asserts fresh with lookup_source cache and cached with lookup_source primary are rejected.
  - REQ-CONTEXT-FORMAT-0017: Asserts rejection of network_type "private" and lookup_source "unknown" enum values plus duplicate keys and oversize input.
  - REQ-CONTEXT-FORMAT-0018: Asserts a context with nl/nl-NH codes decodes to CountryCode NL and RegionCode NL-NH.
  - REQ-CONTEXT-FORMAT-0019: Asserts rejection of a region_code NL-NH with country_code US and of a region_code NL-NH with country_code removed.
  - REQ-CONTEXT-FORMAT-0020: Asserts a direct network_path egress object with connection_code reviewed_proxy is rejected.
  - REQ-CONTEXT-FORMAT-0022: Asserts the case "unavailable classification" (status unavailable, lookup_source none, with country/region/ASN/classification_version present) is rejected.
  - REQ-CONTEXT-FORMAT-0023: Asserts, via assertContextRejected for each invalid case, that the returned context fails to serialize, and that a zero NetworkContext cannot be marshalled.

<a id="UT.internal.model.observation.TestContextUnknownAndFreshnessStates"></a>

### UT.internal.model.observation.TestContextUnknownAndFreshnessStates

- Case: `UT.internal.model.observation.TestContextUnknownAndFreshnessStates`
- Result: passed
- Test: `TestContextUnknownAndFreshnessStates` in `internal/model/observation/context_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestContextUnknownAndFreshnessStates$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestContextUnknownAndFreshnessStates in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-CONTEXT-FORMAT-0006: Asserts fresh/primary, cached/cache, fresh/fallback without classification, and unavailable/none egress objects on a direct path decode and round-trip.
  - REQ-CONTEXT-FORMAT-0007: Asserts an egress object with status unavailable, lookup_source none, network_path direct and no classification fields decodes and round-trips.
  - REQ-CONTEXT-FORMAT-0008: Asserts fresh/primary, fresh/fallback and cached/cache egress combinations decode successfully.
  - REQ-CONTEXT-FORMAT-0009: Asserts a fresh/primary egress with only classification_version decodes with empty CountryCode and nil ASN, and that fresh/fallback without classification fields is accepted.

<a id="UT.internal.model.observation.TestDNSBoundsAndPrivacy"></a>

### UT.internal.model.observation.TestDNSBoundsAndPrivacy

- Case: `UT.internal.model.observation.TestDNSBoundsAndPrivacy`
- Result: passed
- Test: `TestDNSBoundsAndPrivacy` in `internal/model/observation/dns_tls_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestDNSBoundsAndPrivacy$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestDNSBoundsAndPrivacy in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0009: Asserts udp/ipv4/A and system/AAAA observations round-trip and rejects an unknown engine, null responses, response_limit_reached false, and response_limit_reached true before the bound.
  - REQ-OBS-0012: Asserts rejection of TXT type, an AAAA answer for an A query, Alias/café/trailing-dot/64-char-label CNAMEs, and answers 10.0.0.1, 127.0.0.1, 192.0.2.1, 001.1.1.1, ::ffff:1.1.1.1, fe80::1%en0 and other non-public addresses.
  - REQ-OBS-0013: Asserts an observation with two identical A answers in one response and an answer carrying ttl are rejected.
  - REQ-OBS-0014: Asserts a DNS observation padded to exactly 16384 bytes is accepted and one byte more is rejected, answers are rejected when RetainDNSAnswers is false, and resolver_address/raw_message fields are rejected.
  - REQ-OBS-0040: Asserts a DNS observation with an injected resolver_address field is rejected.
  - REQ-OBS-0042: Asserts rejection of arrival_index 2, 0 and 1.0, arrival_offset_ms 750 beyond duration and 251, missing answers, and rcode values noerror, BADVERS and 16.
  - REQ-OBS-0045: Asserts rcode noerror (incorrect case), BADVERS (raw label) and numeric 16 are rejected.
  - REQ-OBS-0047: Asserts a succeeded observation whose only response has rcode OTHER is rejected.

<a id="UT.internal.model.observation.TestDNSCollectionAndAnswerBoundaries"></a>

### UT.internal.model.observation.TestDNSCollectionAndAnswerBoundaries

- Case: `UT.internal.model.observation.TestDNSCollectionAndAnswerBoundaries`
- Result: passed
- Test: `TestDNSCollectionAndAnswerBoundaries` in `internal/model/observation/dns_tls_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestDNSCollectionAndAnswerBoundaries$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestDNSCollectionAndAnswerBoundaries in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0009: Asserts three UDP responses with response_limit_reached are accepted, multiple responses for tcp engine are rejected, and a configured MaxDNSResponses of 2 is enforced.
  - REQ-OBS-0013: Asserts NewDNS removes an appended exact duplicate answer leaving two answers with the original CNAME first.
  - REQ-OBS-0014: Asserts NewDNS accepts 16 answers and rejects 17 answer inputs in one response.
  - REQ-OBS-0042: Asserts nine CNAMEs are rejected while eight are accepted, and decreasing arrival offsets across responses are rejected.

<a id="UT.internal.model.observation.TestHTTPBoundsAndConsistency"></a>

### UT.internal.model.observation.TestHTTPBoundsAndConsistency

- Case: `UT.internal.model.observation.TestHTTPBoundsAndConsistency`
- Result: passed
- Test: `TestHTTPBoundsAndConsistency` in `internal/model/observation/observation_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestHTTPBoundsAndConsistency$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestHTTPBoundsAndConsistency in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0025: Asserts rejection of protocol h3, status_code 99 and 600, empty checks, null body and an unregistered redirect terminal class, and acceptance of a redirect with followed_count 2.
  - REQ-OBS-0028: Asserts rejection when truncated is missing, bytes_processed 513 exceeds a 512-byte configured bound, truncated is true with 511 bytes, and when the limit itself exceeds 65536.
  - REQ-OBS-0030: Asserts rejection of status check index 1, index -1, body check index 8, a duplicate status check, and a body check when BodyChecks is configured as 0.
  - REQ-OBS-0037: Asserts an http observation with failure raw_error is rejected.
  - REQ-OBS-0038: Asserts a succeeded http observation with an unmatched check and a failed body_check_failed without an unmatched body check are rejected, while body_check_failed with an unmatched body check is accepted.

<a id="UT.internal.model.observation.TestIncompleteFailureAbsentForEveryType"></a>

### UT.internal.model.observation.TestIncompleteFailureAbsentForEveryType

- Case: `UT.internal.model.observation.TestIncompleteFailureAbsentForEveryType`
- Result: passed
- Test: `TestIncompleteFailureAbsentForEveryType` in `internal/model/observation/dns_tls_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestIncompleteFailureAbsentForEveryType$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestIncompleteFailureAbsentForEveryType in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0051: Asserts incomplete tcp, http, dns and tls observations are accepted without failure and rejected with failure null, "", timeout or inconclusive.

<a id="UT.internal.model.observation.TestProtocolPresenceAndVocabulary"></a>

### UT.internal.model.observation.TestProtocolPresenceAndVocabulary

- Case: `UT.internal.model.observation.TestProtocolPresenceAndVocabulary`
- Result: passed
- Test: `TestProtocolPresenceAndVocabulary` in `internal/model/observation/observation_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestProtocolPresenceAndVocabulary$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestProtocolPresenceAndVocabulary in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0003: Asserts valid tcp/http/incomplete/failed observations round-trip and rejects missing/negative/251/float/null duration_ms, unknown status, failed without failure, failure on succeeded or incomplete, and DNS/TLS type with TCP details.
  - REQ-OBS-0036: Asserts a tcp observation with status succeeded and failure timeout is rejected.
  - REQ-OBS-0037: Asserts a failed tcp observation without failure and one with failure network_changed are rejected, and failed/inconclusive is accepted.
  - REQ-OBS-0044: Asserts DecodeProtocol with contract 0.2.0 returns ErrVersion.
  - REQ-OBS-0051: Asserts an incomplete tcp observation with failure timeout is rejected.

<a id="UT.internal.model.observation.TestProtocolPrivacyAndImmutableViews"></a>

### UT.internal.model.observation.TestProtocolPrivacyAndImmutableViews

- Case: `UT.internal.model.observation.TestProtocolPrivacyAndImmutableViews`
- Result: passed
- Test: `TestProtocolPrivacyAndImmutableViews` in `internal/model/observation/observation_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestProtocolPrivacyAndImmutableViews$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestProtocolPrivacyAndImmutableViews in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0004: Asserts tcp and http observations with injected unknown common or detail fields (raw_packet, headers, destination, etc.) are rejected.
  - REQ-OBS-0018: Asserts tcp observations with an injected destination field (at top level) are rejected by the decoder.
  - REQ-OBS-0031: Asserts http observations with injected headers, body_bytes or body_digest detail fields are rejected and that detached Check/Body/HTTPDetails values cannot be JSON-serialized.
  - REQ-OBS-0040: Asserts tcp and http observations with injected resolver_address, destination or proxy_host fields are rejected.

<a id="UT.internal.model.observation.TestProtocolValidationPrecedenceAndStableEncoding"></a>

### UT.internal.model.observation.TestProtocolValidationPrecedenceAndStableEncoding

- Case: `UT.internal.model.observation.TestProtocolValidationPrecedenceAndStableEncoding`
- Result: passed
- Test: `TestProtocolValidationPrecedenceAndStableEncoding` in `internal/model/observation/validation_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestProtocolValidationPrecedenceAndStableEncoding$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestProtocolValidationPrecedenceAndStableEncoding in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0008: Asserts DecodeProtocol with contract 9.0.0 returns ErrVersion for each protocol observation type.
  - REQ-JSON-0011: Asserts an observation containing a private DO_NOT_RETAIN field returns the same ErrField or ErrVersion class, at most 128 bytes, over 20 runs.
  - REQ-OBS-0004: Asserts an injected unknown common field "private" is rejected with jsondecode.ErrField for all four observation types.
  - REQ-OBS-0044: Asserts contract 9.0.0 returns ErrVersion before field validation for all four observation types.

<a id="UT.internal.model.observation.TestTLSVocabularyAndConsistency"></a>

### UT.internal.model.observation.TestTLSVocabularyAndConsistency

- Case: `UT.internal.model.observation.TestTLSVocabularyAndConsistency`
- Result: passed
- Test: `TestTLSVocabularyAndConsistency` in `internal/model/observation/dns_tls_test.go` (package `internal/model/observation`)
- Reproduce one case: `go test -count=1 -run '^TestTLSVocabularyAndConsistency$' ./internal/model/observation`
- Scope: Existing synthetic Go unit test TestTLSVocabularyAndConsistency in package internal/model/observation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0038: Asserts succeeded TLS with certificate_validation not_completed and alert on a succeeded handshake are rejected, and certificate_untrusted with trusted validation is rejected.
  - REQ-OBS-0043: Asserts TLS1.2/TLS1.3/unknown/absent versions are accepted, TLS1.0, 771 and tls1.3 versions, alpn private and empty, alert on succeeded and an extra certificate field are rejected, and each alert class is accepted only with failure alert.
  - REQ-OBS-0045: Asserts TLS1.0, numeric 771, lowercase tls1.3, alpn private and numeric alert 42 are rejected rather than repaired.
  - REQ-OBS-0049: Asserts a TLS observation with alpn private or empty string is rejected.
  - REQ-OBS-0050: Asserts each alert class plus absent alert is accepted with failed/alert, alert on a succeeded handshake and numeric alert 42 are rejected, and certificate_error with trusted local validation is accepted.

<a id="UT.internal.model.task.TestAdmissionErrorVocabulary"></a>

### UT.internal.model.task.TestAdmissionErrorVocabulary

- Case: `UT.internal.model.task.TestAdmissionErrorVocabulary`
- Result: passed
- Test: `TestAdmissionErrorVocabulary` in `internal/model/task/admission_error_test.go` (package `internal/model/task`)
- Reproduce one case: `go test -count=1 -run '^TestAdmissionErrorVocabulary$' ./internal/model/task`
- Scope: Existing synthetic Go unit test TestAdmissionErrorVocabulary in package internal/model/task, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0034: Asserts each of the ten codes maps to its registered string, is not an execution reason, and that all other uint8 values outside the eleven defined codes are rejected.
  - REQ-TRUST-0035: Asserts admission errors refuse JSON marshal/unmarshal and that a canary containing secret, credentials, participant, target and raw platform text never appears in formatted output.
  - REQ-TRUST-0036: Asserts the Unsupported code maps to the fixed string task_unsupported, is not an execution reason, carries no payload under the same encoding and canary checks, and that values 12 to 255 and 0 are rejected.

<a id="UT.internal.model.task.TestMalformedSourceRejectedAtomically"></a>

### UT.internal.model.task.TestMalformedSourceRejectedAtomically

- Case: `UT.internal.model.task.TestMalformedSourceRejectedAtomically`
- Result: passed
- Test: `TestMalformedSourceRejectedAtomically` in `internal/model/task/trust_test.go` (package `internal/model/task`)
- Reproduce one case: `go test -count=1 -run '^TestMalformedSourceRejectedAtomically$' ./internal/model/task`
- Scope: Existing synthetic Go unit test TestMalformedSourceRejectedAtomically in package internal/model/task, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0001: Asserts NewSourceContext returns ErrTrustModel and an invalid context for a remote source with Authenticated false, nil or zero IssuedAt, invalid origin, zero policy generation, or invalid issuer.

<a id="UT.internal.result.TestBindingBoundsAndProducerContext"></a>

### UT.internal.result.TestBindingBoundsAndProducerContext

- Case: `UT.internal.result.TestBindingBoundsAndProducerContext`
- Result: passed
- Test: `TestBindingBoundsAndProducerContext` in `internal/result/payload_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestBindingBoundsAndProducerContext$' ./internal/result`
- Scope: Existing synthetic Go unit test TestBindingBoundsAndProducerContext in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0011: Asserts New rejects a binding whose Test.Version or Test.ResultVersion alone is 9.0.0.
  - REQ-RESULT-0038: Asserts New rejects a binding with unsupported Test.Version or Test.ResultVersion.

<a id="UT.internal.result.TestCompositeCompletionAndPartialBoundaries"></a>

### UT.internal.result.TestCompositeCompletionAndPartialBoundaries

- Case: `UT.internal.result.TestCompositeCompletionAndPartialBoundaries`
- Result: passed
- Test: `TestCompositeCompletionAndPartialBoundaries` in `internal/result/payload_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestCompositeCompletionAndPartialBoundaries$' ./internal/result`
- Scope: Existing synthetic Go unit test TestCompositeCompletionAndPartialBoundaries in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0040: Accepts a NET-05 result with profile, network_validation and two ordered step measurements, including a negative primitive outcome, and rejects the steps in reverse order.
  - REQ-NET-05-0041: Accepts partial results with stopped_before_step naming the next profile step, rejects missing or wrong stopped_before_step, and rejects a profile_version different from the bound profile.
  - REQ-NET-05-0044: Rejects a step whose test result_version is 9.0.0, rejects an incomplete step that includes a measurement, and accepts completed steps with primitive measurements.
  - REQ-NET-05-0045: Accepts incomplete results with zero, one completed, one incomplete, or completed-then-incomplete steps and matching stopped_before_step, and rejects an incomplete first step followed by a completed step.
  - REQ-NET-05-0046: Accepts completed with all steps completed under completed state, rejects completed for partial step sets, accepts network_not_validated with no steps only when the profile requires validation, and rejects steps run under not_validated validation when required.
  - REQ-RESULT-0021: Accepts NET-05 incomplete results with partial step measurements and rejects inconsistent partial structures.
  - REQ-RESULT-0027: Accepts NET-05 measurements carrying the named profile, network_validation, steps, and stopped_before_step fields.
  - REQ-RESULT-0029: Accepts NET-05 incomplete results retaining zero, one completed, or a trailing incomplete step with stopped_before_step, and rejects inconsistent partial structures.

<a id="UT.internal.result.TestCompositeVersionPrecisionAndMarkerPresence"></a>

### UT.internal.result.TestCompositeVersionPrecisionAndMarkerPresence

- Case: `UT.internal.result.TestCompositeVersionPrecisionAndMarkerPresence`
- Result: passed
- Test: `TestCompositeVersionPrecisionAndMarkerPresence` in `internal/result/payload_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestCompositeVersionPrecisionAndMarkerPresence$' ./internal/result`
- Scope: Existing synthetic Go unit test TestCompositeVersionPrecisionAndMarkerPresence in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0007: Asserts a NET-05 profile_version of 9007199254740993 round-trips without rounding.
  - REQ-NET-05-0041: Rejects a completed NET-05 result containing an empty stopped_before_step.

<a id="UT.internal.result.TestEncodedAndObservationBoundaries"></a>

### UT.internal.result.TestEncodedAndObservationBoundaries

- Case: `UT.internal.result.TestEncodedAndObservationBoundaries`
- Result: passed
- Test: `TestEncodedAndObservationBoundaries` in `internal/result/result_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestEncodedAndObservationBoundaries$' ./internal/result`
- Scope: Existing synthetic Go unit test TestEncodedAndObservationBoundaries in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0009: Asserts a 65536-byte result is accepted and a 65537-byte result is rejected with ErrBounds.
  - REQ-LIMIT-0021: Asserts a 65536-byte TCP-01 result is accepted, 65537 bytes is rejected with ErrBounds, and a result with 17 observations is rejected.
  - REQ-RESULT-0004: For TCP-01, accepts a whitespace-padded input of exactly 65,536 bytes, rejects 65,537 bytes with jsondecode.ErrBounds, and rejects a measurement with 17 observations.
  - REQ-RESULT-0026: Rejects a TCP-01 measurement with 17 observations.

<a id="UT.internal.result.TestEnvelopePresenceAndBounds"></a>

### UT.internal.result.TestEnvelopePresenceAndBounds

- Case: `UT.internal.result.TestEnvelopePresenceAndBounds`
- Result: passed
- Test: `TestEnvelopePresenceAndBounds` in `internal/result/result_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestEnvelopePresenceAndBounds$' ./internal/result`
- Scope: Existing synthetic Go unit test TestEnvelopePresenceAndBounds in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0002: Rejects deletion of each required top-level field, mismatched task_id, zero/padded/noncanonical run_id, attempt 0/3/1.5 and non-canonical numeric spellings, invalid platform and library_version, and malformed time/execution values for a TCP-01 binding.
  - REQ-RESULT-0016: Rejects started_at with seconds, fractional seconds, or +00:00 offset, precision other than minute, duration_ms 999, duration_capped false, and duration_capped true below 60,000, and accepts duration_ms 60000 with duration_capped true.
  - REQ-RESULT-0019: Rejects execution.state 'failed' and completed results carrying reason cancelled_by_host or an empty reason.
  - REQ-RESULT-0021: Rejects a completed TCP-01 result with the measurement removed.
  - REQ-RESULT-0023: Rejects a completed TCP-01 result with network_context missing or null.
  - REQ-RESULT-0034: Rejects deletion of each required top-level field, a completed result without measurement or with reason, and missing network_context for a completed TCP-01 result.
  - REQ-RESULT-0036: Rejects an all-zero run_id, a padded run_id, and a noncanonical run_id encoding.
  - REQ-RESULT-0041: Rejects a completed TCP-01 result whose measurement is removed.

<a id="UT.internal.result.TestExecutionContextAndAttemptBinding"></a>

### UT.internal.result.TestExecutionContextAndAttemptBinding

- Case: `UT.internal.result.TestExecutionContextAndAttemptBinding`
- Result: passed
- Test: `TestExecutionContextAndAttemptBinding` in `internal/result/result_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestExecutionContextAndAttemptBinding$' ./internal/result`
- Scope: Existing synthetic Go unit test TestExecutionContextAndAttemptBinding in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0019: Accepts incomplete and not_run results with each registered reason and rejects them when the reason is removed.
  - REQ-RESULT-0020: Decodes incomplete and not_run TCP-01 results with each of the 15 registered reasons and rejects reason 'timeout'.
  - REQ-RESULT-0021: For TCP-01, rejects incomplete and not_run results that include a measurement or omit the reason, and accepts them with a registered reason and no measurement.
  - REQ-RESULT-0023: Rejects an incomplete (network_changed) TCP-01 result whose network_context was deleted, while accepting results whose egress.status is unavailable.
  - REQ-RESULT-0024: Accepts a not_run TCP-01 result without network_context, rejects it when Binding.CapturedContext is true, and rejects an incomplete result without context.
  - REQ-RESULT-0029: Rejects TCP-01 incomplete results that include a measurement.
  - REQ-RESULT-0034: Rejects incomplete/not_run results without reason or with measurement, and not_run without context when context was captured.
  - REQ-RESULT-0042: For incomplete and not_run TCP-01 results, accepts each of 15 registered reasons and rejects a missing reason and reason 'timeout'.

<a id="UT.internal.result.TestImmutableViewsAndProducerBuild"></a>

### UT.internal.result.TestImmutableViewsAndProducerBuild

- Case: `UT.internal.result.TestImmutableViewsAndProducerBuild`
- Result: passed
- Test: `TestImmutableViewsAndProducerBuild` in `internal/result/result_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestImmutableViewsAndProducerBuild$' ./internal/result`
- Scope: Existing synthetic Go unit test TestImmutableViewsAndProducerBuild in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0004: Asserts Build rejects a producer result with 17 observations and a recursive NET-05 step structure before encoding.

<a id="UT.internal.result.TestIndependentVersionsAndNegativeCompletion"></a>

### UT.internal.result.TestIndependentVersionsAndNegativeCompletion

- Case: `UT.internal.result.TestIndependentVersionsAndNegativeCompletion`
- Result: passed
- Test: `TestIndependentVersionsAndNegativeCompletion` in `internal/result/result_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestIndependentVersionsAndNegativeCompletion$' ./internal/result`
- Scope: Existing synthetic Go unit test TestIndependentVersionsAndNegativeCompletion in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0001: Decodes TCP-01 results whose single tcp_connect observation failed (connection_refused, timeout, connection_reset, network_unreachable, address_unreachable, permission_denied, socket_error, inconclusive) as execution.state completed with no reason.
  - REQ-RESULT-0011: Mutating schema_version, test.id, test.version, or test.result_version to an unsupported value each yields jsondecode.ErrVersion and no usable Value.
  - REQ-RESULT-0032: Mutating schema_version, test.id, test.version, or test.result_version to an unsupported value each yields jsondecode.ErrVersion and no usable Value.
  - REQ-RESULT-0038: Mutating schema_version, test.id, test.version, or test.result_version to an unsupported value each yields jsondecode.ErrVersion and no usable Value.

<a id="UT.internal.result.TestInitialPayloadSchemasAndEvidence"></a>

### UT.internal.result.TestInitialPayloadSchemasAndEvidence

- Case: `UT.internal.result.TestInitialPayloadSchemasAndEvidence`
- Result: passed
- Test: `TestInitialPayloadSchemasAndEvidence` in `internal/result/payload_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestInitialPayloadSchemasAndEvidence$' ./internal/result`
- Scope: Existing synthetic Go unit test TestInitialPayloadSchemasAndEvidence in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-01-0037: Accepts DNS-01 results whose outcome resolved has a succeeded system dns_query and whose outcome nxdomain has a dns_query failed with nxdomain retaining rcode NXDOMAIN, and rejects the same results when the outcome is replaced by the unregistered value censored.
  - REQ-DNS-03-0027: Accepts DNS-03 results with outcome response (tcp, tls, http, succeeded doh dns_query with NOERROR) and tcp_failure (failed tcp_connect), and rejects them when the outcome is replaced by censored.
  - REQ-DNS-04-0027: Accepts DNS-04 results with outcome dns_refused (dot dns_query failed with refused retaining rcode REFUSED after tcp and tls) and tcp_timeout (tcp_connect failed with timeout), and rejects them when the outcome is replaced by censored.
  - REQ-HTTP-01-0036: Accepts HTTP-01 results with outcome expected_response (terminal succeeded http_exchange) and tls_failure (tls_handshake failed with certificate_untrusted), and rejects them when the outcome is replaced by censored.
  - REQ-NET-01-0023: Accepts a NET-01 snapshot result with an access object and empty observations, rejects snapshot without access, rejects a top-level network_context member, and rejects a snapshot with network_changed true.
  - REQ-NET-03-0027: Accepts NET-03 results validated (expected response class, platform unknown) and connection_failure (failed tcp_connect), and rejects them when the outcome is replaced by censored.
  - REQ-RESULT-0024: Asserts a NET-01 result that adds a network_context object is rejected.
  - REQ-RESULT-0026: Accepts registered outcomes with ordered observations for TCP-01, TLS-01, DNS-01, DNS-03, DNS-04, HTTP-01, NET-01, and NET-03, rejects outcome 'censored', and rejects out-of-order or mismatched observation sequences.
  - REQ-RESULT-0027: For each initial test, asserts a measurement containing an extra raw_body field is rejected.
  - REQ-RESULT-0041: Rejects outcome 'censored' for each initial test, an incomplete tcp_connect observation in a completed TCP-01 result, and TLS-01 outcomes inconsistent with observations.
  - REQ-RESULT-0043: Rejects a TCP-01 result carrying a tls_handshake observation and TLS-01 results with wrong observation order or missing handshake.
  - REQ-RESULT-0046: Rejects a measurement with an added raw_body field for each initial test.
  - REQ-TCP-01-0024: Accepts a TCP-01 connected result with one succeeded tcp_connect observation, rejects a TCP-01 result whose observation is tls_handshake, rejects an incomplete tcp_connect, and rejects unknown measurement fields.
  - REQ-TLS-01-0025: Accepts TLS-01 results with tcp_connect followed by tls_handshake and a tcp_refused result with only a failed tcp_connect, and rejects tls_handshake before tcp_connect, handshake_succeeded without a tls_handshake, and a tls_handshake after a failed tcp_connect.

<a id="UT.internal.result.TestProducerTimingAndRandomIdentity"></a>

### UT.internal.result.TestProducerTimingAndRandomIdentity

- Case: `UT.internal.result.TestProducerTimingAndRandomIdentity`
- Result: passed
- Test: `TestProducerTimingAndRandomIdentity` in `internal/result/result_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestProducerTimingAndRandomIdentity$' ./internal/result`
- Scope: Existing synthetic Go unit test TestProducerTimingAndRandomIdentity in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0006: Generates 100 run IDs with NewRunID and asserts each is 22 characters, not all-zero, and unique within the sample.
  - REQ-RESULT-0016: Asserts Timing maps elapsed 0, 1ns, 250ms, 250ms+1ns, 60s, and 60s+1ns to 0, 250, 250, 500, 60000 (uncapped), and 60000 (capped).
  - REQ-RESULT-0018: Asserts Timing rejects start.Round(0) (stripped monotonic reading) and an end before start, and caps elapsed time above 60 s at 60000 with duration_capped.
  - REQ-RESULT-0036: Asserts NewRunID output is 22 characters and never the all-zero encoding over 100 samples.

<a id="UT.internal.result.TestResultValidationPrecedence"></a>

### UT.internal.result.TestResultValidationPrecedence

- Case: `UT.internal.result.TestResultValidationPrecedence`
- Result: passed
- Test: `TestResultValidationPrecedence` in `internal/result/validation_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestResultValidationPrecedence$' ./internal/result`
- Scope: Existing synthetic Go unit test TestResultValidationPrecedence in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0008: Asserts a result with schema_version 9.0.0 or test result_version 9.0.0 is rejected with ErrVersion.
  - REQ-JSON-0011: Asserts each invalid result case, including one with a private DO_NOT_RETAIN field, returns the expected error class, at most 128 bytes, over 30 runs.
  - REQ-RESULT-0032: Asserts an unknown schema_version is rejected with ErrVersion before unknown fields are evaluated and a test.result_version mismatch is rejected with ErrVersion.

<a id="UT.internal.result.TestUninitializedDecoderRejectsBeforeInput"></a>

### UT.internal.result.TestUninitializedDecoderRejectsBeforeInput

- Case: `UT.internal.result.TestUninitializedDecoderRejectsBeforeInput`
- Result: passed
- Test: `TestUninitializedDecoderRejectsBeforeInput` in `internal/result/validation_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestUninitializedDecoderRejectsBeforeInput$' ./internal/result`
- Scope: Existing synthetic Go unit test TestUninitializedDecoderRejectsBeforeInput in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0010: Asserts nil and zero result decoders return ErrBinding for any input and the returned value cannot be serialized.

<a id="UT.internal.result.TestUnknownPrivateAndMalformedFields"></a>

### UT.internal.result.TestUnknownPrivateAndMalformedFields

- Case: `UT.internal.result.TestUnknownPrivateAndMalformedFields`
- Result: passed
- Test: `TestUnknownPrivateAndMalformedFields` in `internal/result/result_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestUnknownPrivateAndMalformedFields$' ./internal/result`
- Scope: Existing synthetic Go unit test TestUnknownPrivateAndMalformedFields in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0003: Adds an unknown private_endpoint field at the root (and keys such as input, options, labels, user_id, raw_error, annotations) and asserts the result is rejected with a sanitized error.
  - REQ-RESULT-0005: Adds top-level input, options, schedule, safety, labels, and input_digest keys to a valid TCP-01 result and asserts each is rejected without echoing the value in the error.
  - REQ-RESULT-0014: Adds a private_endpoint field inside the implementation object, and an annotations key at the root, and asserts both results are rejected.
  - REQ-RESULT-0025: Adds private_endpoint inside network_context.access, network_context.egress, and other objects, and ip_address at the root, and asserts each result is rejected.
  - REQ-RESULT-0027: Asserts an unknown private_endpoint field inside measurement is rejected.
  - REQ-RESULT-0032: Asserts unknown fields at root and nested objects are rejected rather than discarded.
  - REQ-RESULT-0043: Rejects unknown fields added to an observation and to its details object.
  - REQ-RESULT-0044: Rejects unknown fields at every object level, keys such as raw_error, user_id, ip_address, and input_digest, trailing data, duplicate keys, invalid escapes, and a BOM, with errors not containing the injected value.
  - REQ-RESULT-0046: Rejects results carrying raw_error or an injected private_endpoint value in root, measurement, observation, and details objects.

<a id="UT.internal.taskdecode.TestByteAndLabelBounds"></a>

### UT.internal.taskdecode.TestByteAndLabelBounds

- Case: `UT.internal.taskdecode.TestByteAndLabelBounds`
- Result: passed
- Test: `TestByteAndLabelBounds` in `internal/taskdecode/decode_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestByteAndLabelBounds$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestByteAndLabelBounds in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0009: Asserts task inputs of 65536 bytes are accepted and 65537 rejected with ErrBounds, and input, options and labels nested-byte limits count original whitespace bytes.
  - REQ-LABEL-0005: Asserts that encoded labels objects of 2047 and 2048 bytes are accepted and 2049 is rejected, and that 17 labels are rejected with ErrBounds while 16 are accepted.

<a id="UT.internal.taskdecode.TestDefaultsAndDeterminism"></a>

### UT.internal.taskdecode.TestDefaultsAndDeterminism

- Case: `UT.internal.taskdecode.TestDefaultsAndDeterminism`
- Result: passed
- Test: `TestDefaultsAndDeterminism` in `internal/taskdecode/decode_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestDefaultsAndDeterminism$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestDefaultsAndDeterminism in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0013: Asserts a task omitting execution limits decodes with defaults DurationMS 30000, BytesSent 32768, BytesReceived 131072 and Connections 8.
  - REQ-LABEL-0004: Asserts that decoding a task with "labels":{} yields a candidate deeply equal to the candidate decoded with labels omitted.

<a id="UT.internal.taskdecode.TestExactSchemaSelection"></a>

### UT.internal.taskdecode.TestExactSchemaSelection

- Case: `UT.internal.taskdecode.TestExactSchemaSelection`
- Result: passed
- Test: `TestExactSchemaSelection` in `internal/taskdecode/decode_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestExactSchemaSelection$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestExactSchemaSelection in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0008: Asserts a 0.2.0 task selects only the 0.2.0 codec, invalid 0.2.0 input calls no codec, and unsupported or case-variant test_id/test_version tuples are rejected with ErrVersion.

<a id="UT.internal.taskdecode.TestHostConfigurationCannotComeFromTask"></a>

### UT.internal.taskdecode.TestHostConfigurationCannotComeFromTask

- Case: `UT.internal.taskdecode.TestHostConfigurationCannotComeFromTask`
- Result: passed
- Test: `TestHostConfigurationCannotComeFromTask` in `internal/taskdecode/config_boundary_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestHostConfigurationCannotComeFromTask$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestHostConfigurationCannotComeFromTask in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LABEL-0001: Asserts that a label whose value is host result_linkage JSON is retained as opaque text and does not yield a valid deployment configuration.
  - REQ-LINKAGE-0006: Asserts that a task containing a top-level result_linkage field is rejected, and that a label containing result_linkage JSON does not produce a valid configuration.
  - REQ-TASK-0001: Asserts tasks with top-level proxy, network_context, result_linkage, local_policy, policy_generation, authenticated, clock or linkage_keys members are rejected with no candidate.

<a id="UT.internal.taskdecode.TestModelValidationPrecedenceAndExactCodec"></a>

### UT.internal.taskdecode.TestModelValidationPrecedenceAndExactCodec

- Case: `UT.internal.taskdecode.TestModelValidationPrecedenceAndExactCodec`
- Result: passed
- Test: `TestModelValidationPrecedenceAndExactCodec` in `internal/taskdecode/validation_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestModelValidationPrecedenceAndExactCodec$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestModelValidationPrecedenceAndExactCodec in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0008: Asserts an unsupported test_version is rejected with ErrVersion and a 0.2.0 task with 0.1.0-shaped input fails with ErrField without invoking the 0.1.0 codec.
  - REQ-JSON-0010: Asserts invalid task inputs return a nil candidate and do not invoke the test codec before earlier validation succeeds.
  - REQ-JSON-0011: Asserts each invalid task case returns the expected error class, at most 128 bytes, over 20 runs.

<a id="UT.internal.taskdecode.TestRegisteredNumericalBoundaries"></a>

### UT.internal.taskdecode.TestRegisteredNumericalBoundaries

- Case: `UT.internal.taskdecode.TestRegisteredNumericalBoundaries`
- Result: passed
- Test: `TestRegisteredNumericalBoundaries` in `internal/taskdecode/boundary_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestRegisteredNumericalBoundaries$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestRegisteredNumericalBoundaries in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0013: Asserts the lowest and highest registered values of max_duration_ms, max_bytes_sent, max_bytes_received and max_connections are accepted.
  - REQ-LABEL-0005: Asserts via Decode that 64-byte label keys are accepted and 65-byte keys rejected, and that 128-byte UTF-8 values (64 'я') are accepted and 130-byte values rejected.

<a id="UT.internal.taskdecode.TestRejectionCorpus"></a>

### UT.internal.taskdecode.TestRejectionCorpus

- Case: `UT.internal.taskdecode.TestRejectionCorpus`
- Result: passed
- Test: `TestRejectionCorpus` in `internal/taskdecode/decode_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestRejectionCorpus$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestRejectionCorpus in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0013: Asserts max_duration_ms 60001, max_bytes_sent 65537, max_bytes_received 262145 and max_connections 13 are rejected with ErrRange and no candidate.
  - REQ-LABEL-0014: The 'label-control' case asserts that a decoded label value 'a\nb' is rejected with validation.ErrTask.
  - REQ-TASK-0001: Asserts a task with an unknown top-level extra member is rejected with ErrField and no candidate.

<a id="UT.internal.taskdecode.TestRepresentationAndSanitizedErrors"></a>

### UT.internal.taskdecode.TestRepresentationAndSanitizedErrors

- Case: `UT.internal.taskdecode.TestRepresentationAndSanitizedErrors`
- Result: passed
- Test: `TestRepresentationAndSanitizedErrors` in `internal/taskdecode/decode_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestRepresentationAndSanitizedErrors$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestRepresentationAndSanitizedErrors in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0003: Rejects task inputs with task_id and task_\u0069d, and with repeated subject or enabled members, returning no candidate.
  - REQ-JSON-0006: Rejects task max_connections values -0, 1.0 and 1e0 with no candidate.
  - REQ-JSON-0011: Asserts a failing test codec's error is surfaced as ErrParameters without the codec's sensitive text.

<a id="UT.internal.taskdecode.TestRiskFactorAndTimezoneDomains"></a>

### UT.internal.taskdecode.TestRiskFactorAndTimezoneDomains

- Case: `UT.internal.taskdecode.TestRiskFactorAndTimezoneDomains`
- Result: passed
- Test: `TestRiskFactorAndTimezoneDomains` in `internal/taskdecode/domain_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestRiskFactorAndTimezoneDomains$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestRiskFactorAndTimezoneDomains in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0012: Decodes synthetic SYN-01 tasks and asserts that `medium`, `high` and `unknown` with `risk_factors` omitted return ErrTask and a nil candidate, while `low` with it omitted decodes.
  - REQ-SAFETY-0013: Asserts that one `network_observability` factor at the task's level decodes for each risk level and is preserved in the candidate.
  - REQ-SAFETY-0031: Asserts that `medium`, `high` and `unknown` with `risk_factors: []` return ErrTask and a nil candidate on every repetition, that one factor decodes, and that `low` with `[]` decodes with an empty factor list.
  - REQ-SCHEDULE-0013: Asserts that a `time_window` with `timezone` omitted decodes with timezone `local`, and with `local` decodes unchanged.
  - REQ-SCHEDULE-0014: Asserts that `local` is the decoded and accepted timezone value; device-zone evaluation is not exercised.
  - REQ-SCHEDULE-0043: Asserts that `timezone` values `UTC`, `Europe/Berlin`, `""`, `Local`, `LOCAL` and `local ` return ErrTask and a nil candidate.

<a id="UT.internal.taskdecode.TestSerializedInputCannotSupplyTrustOrPolicy"></a>

### UT.internal.taskdecode.TestSerializedInputCannotSupplyTrustOrPolicy

- Case: `UT.internal.taskdecode.TestSerializedInputCannotSupplyTrustOrPolicy`
- Result: passed
- Test: `TestSerializedInputCannotSupplyTrustOrPolicy` in `internal/taskdecode/trust_boundary_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestSerializedInputCannotSupplyTrustOrPolicy$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestSerializedInputCannotSupplyTrustOrPolicy in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0004: Rejects tasks carrying trust/policy field names at top level and under input, options, safety, execution and schedule, rejects the same names in configuration, and keeps a label value resembling a source record as opaque text.
  - REQ-LABEL-0001: Asserts that a label whose value is a serialized trust/policy assertion is retained as unchanged opaque text.
  - REQ-TASK-0001: Asserts tasks with top-level trust/policy-named members such as source_context, origin and issuer_id are rejected with no candidate.

<a id="UT.internal.testkit.conformance.TestConformanceFixtures"></a>

### UT.internal.testkit.conformance.TestConformanceFixtures

- Case: `UT.internal.testkit.conformance.TestConformanceFixtures`
- Result: passed
- Test: `TestConformanceFixtures` in `internal/testkit/conformance/runner_test.go` (package `internal/testkit/conformance`)
- Reproduce one case: `go test -count=1 -run '^TestConformanceFixtures$' ./internal/testkit/conformance`
- Scope: Existing synthetic Go unit test TestConformanceFixtures in package internal/testkit/conformance, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0012: Fixtures `risk-factors-*-one` and `risk-factors-low-empty` decode through adapter `common-task-v1` to the expected normalized projection; this projection does not include safety fields.
  - REQ-SAFETY-0013: Fixtures `risk-factors-medium-one`, `risk-factors-high-one` and `risk-factors-unknown-one`, each with one valid factor, decode to the expected projection.
  - REQ-SAFETY-0031: Fixtures `risk-factors-medium-empty`, `risk-factors-high-empty` and `risk-factors-unknown-empty` reject with exact class `task` and a nil candidate, repeatedly and concurrently; `risk-factors-low-empty` is accepted.
  - REQ-SCHEDULE-0013: Fixtures `timezone-omitted` and `timezone-local` decode to the expected projection.
  - REQ-SCHEDULE-0014: Fixture `timezone-local` is accepted; device-zone evaluation is not exercised.
  - REQ-SCHEDULE-0043: Fixtures `timezone-utc`, `timezone-iana`, `timezone-empty` and `timezone-local-case` reject with exact class `task` and a nil candidate.

<a id="UT.internal.validation.TestCommonLabelInvariants"></a>

### UT.internal.validation.TestCommonLabelInvariants

- Case: `UT.internal.validation.TestCommonLabelInvariants`
- Result: passed
- Test: `TestCommonLabelInvariants` in `internal/validation/task_test.go` (package `internal/validation`)
- Reproduce one case: `go test -count=1 -run '^TestCommonLabelInvariants$' ./internal/validation`
- Scope: Existing synthetic Go unit test TestCommonLabelInvariants in package internal/validation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LABEL-0004: The 'omitted' (nil) and 'empty' ([]) subtests assert that Common accepts both and returns nil Labels.
  - REQ-LABEL-0005: Asserts acceptance of a 64-byte key with a 128-byte UTF-8 value and of 16 labels, and rejection of a 65-byte key, a 130-byte value, and 17 labels.
  - REQ-LABEL-0010: The 'duplicate' subtest asserts that Common rejects two labels with key 'campaign' and returns a zero candidate.
  - REQ-LABEL-0014: Asserts that values containing 0xff (invalid UTF-8) and '\n' are rejected, and that 64x'я' and '${literal}' are accepted.

<a id="UT.internal.validation.TestCommonNormalizationIsIdempotentAndConservative"></a>

### UT.internal.validation.TestCommonNormalizationIsIdempotentAndConservative

- Case: `UT.internal.validation.TestCommonNormalizationIsIdempotentAndConservative`
- Result: passed
- Test: `TestCommonNormalizationIsIdempotentAndConservative` in `internal/validation/task_test.go` (package `internal/validation`)
- Reproduce one case: `go test -count=1 -run '^TestCommonNormalizationIsIdempotentAndConservative$' ./internal/validation`
- Scope: Existing synthetic Go unit test TestCommonNormalizationIsIdempotentAndConservative in package internal/validation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0013: Asserts validation.Common rejects attempt limits with DurationMS 60001 or BytesSent 0 with ErrTask and returns a zero candidate instead of a clamped one.
  - REQ-LABEL-0001: Asserts that Common keeps label values like '${literal}' verbatim, only sorts labels, leaves the rest of the candidate equal to the expected normalized form, and does not alias the input.
  - REQ-LABEL-0014: Asserts that the label values '${literal}' and 'Я' are preserved verbatim through normalization.

<a id="UT.internal.validation.TestCommonRiskFactorAndTimezoneDomains"></a>

### UT.internal.validation.TestCommonRiskFactorAndTimezoneDomains

- Case: `UT.internal.validation.TestCommonRiskFactorAndTimezoneDomains`
- Result: passed
- Test: `TestCommonRiskFactorAndTimezoneDomains` in `internal/validation/task_test.go` (package `internal/validation`)
- Reproduce one case: `go test -count=1 -run '^TestCommonRiskFactorAndTimezoneDomains$' ./internal/validation`
- Scope: Existing synthetic Go unit test TestCommonRiskFactorAndTimezoneDomains in package internal/validation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0012: Asserts that validation.Common rejects `medium`, `high` and `unknown` candidates with no risk factor list with ErrTask and a zero candidate, and accepts a `low` candidate with the list omitted.
  - REQ-SAFETY-0013: Asserts that one factor of type `network_observability` at the task's own level is accepted for each of `low`, `medium`, `high` and `unknown`.
  - REQ-SAFETY-0031: Asserts that `medium`, `high` and `unknown` candidates with an empty, non-nil factor list are rejected with ErrTask and a zero candidate, that one factor makes each valid, and that a `low` candidate with an empty list remains valid.
  - REQ-SCHEDULE-0013: Asserts that a time window with timezone `local` validates and keeps `local`.
  - REQ-SCHEDULE-0014: Asserts that `local` is the timezone value accepted for a time window; the device-reported zone itself is not evaluated by this test.
  - REQ-SCHEDULE-0043: Asserts that time-window timezones `UTC`, `Europe/Berlin`, `""`, `Local` and `LOCAL` are rejected with ErrTask and a zero candidate.
