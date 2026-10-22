# Go Unit-Test Evidence at e372219

DEV-564 records executed synthetic Go tests that partially evidence requirements in the cross-cutting manifest after adding persistent measurement-data clearing behind a durable clearing marker. Each case section below is the conformance evidence for one cross-cutting manifest case: 28 conformance family cases and the 170 `UT.*` unit-test cases. The measurement manifest keeps its earlier baseline and its [evidence at c5f4e92](go-unit-c5f4e92.md).

## Execution Record

- Implementation: `https://github.com/tv-rain/barometer-go` commit `e3722195bdbeb9a25b9f9612eead36683be6e92b` (DEV-564 code commit), exported with `git archive` into an empty directory outside any checkout.
- Specification baseline: `https://github.com/tv-rain/barometer-spec` commit `5d01879cfd1d47396c780befa2567f7e90439948` (Draft `0.1.0`, public PR #24 merge), unchanged from the [e867c98 report](go-unit-e867c98.md). Later public revisions, which add specification-issue records only, are not adopted by this report.
- Environment: go1.26.0 darwin/arm64 (`GOTOOLCHAIN=local`, `GOFLAGS` empty, `CGO_ENABLED=1`), default test timeout, no race detector.
- Command, from the export root: `go test -count=1 -json ./...`
- Outcome: exit status 0; all 362 top-level tests passed, none failed or skipped. Each case below names one passing top-level test; a top-level pass includes all of its subtests. Fuzz targets ran their seed corpora only; the finite fuzzing campaigns are recorded in the [admission design](../../docs/admission.md#fuzz-validation-record) and the [configuration models](../../docs/configuration-models.md#synthetic-evidence).
- Executed on October 22, 2026. Inputs are the synthetic, in-repository fixtures at the implementation commit, including synthetic Ed25519 keys derived from fixed public test seeds, synthetic host configuration approvals, synthetic registry entries and runners, scripted consent, device and time-zone capabilities, a test run ledger, the test-only fault-injectable record store, the production pending store and the production `persistence.State` over that record store, a synthetic linkage key store that holds no key material, and test-only loopback echo, holding and exchange servers on `127.0.0.1`; no deployment key, endpoint, participant, measurement or private data is involved, and no host other than the loopback address is contacted.

## Carried-Forward Cases

Every case from the [e867c98 report](go-unit-e867c98.md) was re-executed at this baseline. Between `e867c98` and `e372219`, DEV-564 adds `persistence.Clear`, the clearing marker class, `State.Forget` and `pending.Store.Forget`, derives the root clearing capability from the record port at bind time and renames the adapter field that holds the port from `pending` to `records`. The root lifecycle tests therefore no longer run without a clearing capability: `UT.root.TestLifecycleTransitionTable` and `UT.root.TestClearInvalidatesHandlesAndDeliveryReferences` reach `clearing_incomplete` through an injected record-port deletion failure, and the transition table now records successful clears from the uninitialized, initialized, close-incomplete and closed states; their aspects below are updated accordingly. `UT.root.TestPanicsAreRecoveredToFixedCodes` still substitutes a panicking synthetic capability. Every other carried case keeps its recorded aspects. `TestAdmissionHasNoNetworkDependency` and `TestRootPackageImportsNoNetworkPackages` still pass.

## New Cases

DEV-564 executes the conformance family cases `CC-LINKAGE-CLEAR-DATA-CONTRACT`, for its Go-owned rows only, and `CC-TRUST-REPLAY-AFTER-CLEARING`, and adds three unit-test cases: `UT.root.TestClearDeletesEveryStateClass`, `UT.internal.persistence.TestClearOrder` and `UT.internal.persistence.TestStateForgetAfterClear`. The host, operating-system and native rows of the clear-data contract move to the planned `CC-LINKAGE-CLEAR-DATA-CONTRACT-HOST` case (DEV-455) and are not evidenced here. The other new tests (`TestRepeatedClearingReachesTheSameScope`, `TestClearFaultAtEachStep`, `TestClearLinkageKeyFailure`, `TestInitializeWithUnreadableMarker`, `TestClearRacesEntryPoints`, `TestClearingCoversEveryAdapter` and `TestClearingHasNoNetworkDependency` in the root package, `TestClearCoversEveryRecordClass`, `TestClearFaults` and `TestClearingTreatsAnyMarkerValueAsPresent` in `internal/persistence` and `TestForgetAfterClear` in `internal/persistence/pending`) passed in the same run but are not mapped as cases. The DEV-563 and DEV-641 unit-test cases are carried forward.

## Limits

Every case is partial evidence. A passing case shows only the stated aspects of each linked requirement; the notes in the manifest record the remaining obligations. Clearing, like the replay store, run ledger and pending store, is exercised only over the test-only, in-memory record store with a synthetic linkage key store: no platform backend, native keystore deletion, reinstall sentinel, operating-system data clearing, host user-interface wiring or host-owned copy exists or is exercised (DEV-455), no durability on any platform is shown, and nothing binds the stores in production. The absence of network activity during clearing is shown at import level, not by observed traffic through production adapters. In-memory clearing drops references and clears slices and maps only; Go does not guarantee that memory is overwritten. No durable policy-generation floor exists, so clearing removes the rollback signal held in replay records. The carried-forward limits of the [e867c98 report](go-unit-e867c98.md#limits) still apply to the carried cases. These reports establish no implemented requirement, released API, platform or field support.

## Conformance Family Cases

<a id="CC-TRUST-ACCEPTED-TASK-ORIGINS"></a>

### CC-TRUST-ACCEPTED-TASK-ORIGINS

- Case: `CC-TRUST-ACCEPTED-TASK-ORIGINS`
- Result: passed
- Test: `TestCcTrustAcceptedTaskOrigins` in `internal/admission/cross_cutting_conformance_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestCcTrustAcceptedTaskOrigins$' ./internal/admission`
- Scope: Synthetic core contract cases; Admit the same synthetic task bytes through delivery records of each origin: embedded with its issuer and no issued_at is accepted, and remote as a signed package from a trusted key of its issuer, issued at the fake clock, is accepted with a replay digest equal to the digest of the task bytes. For remote origin, remove one input at a time: a remote source with a host-supplied issuer cannot be constructed; the bare task presented as remote without a signed package, and a package with an added authenticated member, are rejected as task_source_unauthenticated; a record presented with package bytes differing by one byte is rejected as task_source_unauthenticated; an unavailable trusted clock is rejected as task_trust_unavailable; no delivery record at all is rejected as task_source_unauthenticated. For every rejection assert no replay or admitted-task record is written. Admission has no transport seam, so zero transport calls are established by checking that the dependency closure of admission contains no network package. Requirements: REQ-TRUST-0041, REQ-TRUST-0042, REQ-TRUST-0043, REQ-TRUST-0044.
- Requirement aspects exercised:

  - REQ-TRUST-0041: Asserts that the same task bytes are admitted through an embedded record, whose source carries its issuer and no `issued_at`, and as a signed package from a trusted key of the remote issuer issued at the trusted clock.
  - REQ-TRUST-0042: Asserts that the host cannot assert remote authentication: a remote source with a host-supplied issuer cannot be constructed, and presenting the bare task or a package claiming `authenticated` through a remote record is rejected as `task_source_unauthenticated`.
  - REQ-TRUST-0043: Asserts that remote admission is rejected without any of its prerequisites: the bare task presented as remote without a signed package and a package with an added `authenticated` member are `task_source_unauthenticated`; a record presented with package bytes differing by one byte is `task_source_unauthenticated`; an unavailable trusted clock is `task_trust_unavailable`; and an empty, unknown or malformed reference is `task_source_unauthenticated`. Each rejection writes no replay or admitted-task record.
  - REQ-TRUST-0044: Asserts that the remote replay identity is the issuer of the verified package and the digest of the exact task bytes, equal to the embedded admission of the same bytes, and that no host-supplied remote issuer can be represented.

<a id="CC-TRUST-KEY-ROTATION"></a>

### CC-TRUST-KEY-ROTATION

- Case: `CC-TRUST-KEY-ROTATION`
- Result: passed
- Test: `TestCcTrustKeyRotationByPolicyReplacement` in `policy_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcTrustKeyRotationByPolicyReplacement$' .`
- Scope: Synthetic core contract cases; Initialize a client with a generation-2 policy holding two keys of the remote issuer and admit one synthetic task signed with each. Replace it with generation 3 holding only the second key and assert every task handle and queued delivery reference is invalidated, that the task signed with the removed key and a new task signed with it are rejected as task_source_unauthenticated on redelivery, and that the task signed with the retained key resolves to its retained acceptance without a new replay record. Replace it with generation 4 without the remote issuer and assert its handles are invalidated and its packages rejected with no replay write. Assert that a policy whose key names an issuer it does not authorize is not constructed and the active policy is unchanged. With a key valid until one hour after acceptance, admit a task valid for three hours, advance the clock two hours, and assert the handle still resolves with its original expiry while a new presentation of the same package is rejected as task_source_unauthenticated and the handle still resolves after that rejection. With the same key validity, replace the policy after two hours with generation 3 keeping the expired key and assert the handle is invalidated and a redelivery of the same package is rejected as task_source_unauthenticated with no replay write. With an expiring key and a second valid key of the same issuer, advance two hours and assert the same task bytes re-signed with the valid key resolve to the held handle, then after a generation-3 replacement assert the original package is rejected as task_source_unauthenticated and the re-signed package resolves to the retained acceptance as a redelivery with its original expiry and no new replay record. Requirements: REQ-TRUST-0030, REQ-TRUST-0051, REQ-TRUST-0054, REQ-TRUST-0055.
- Requirement aspects exercised:

  - REQ-TRUST-0030: Asserts through `Client.replacePolicy` that a replacement removing a key and then an issuer invalidates every task handle and queued delivery reference, and that only tasks still permitted by the new policy are admitted again, resolving to their retained acceptance with no new replay record.
  - REQ-TRUST-0051: Asserts that a policy whose trusted key names an issuer the policy does not authorize is not constructed, leaving the active policy unchanged.
  - REQ-TRUST-0054: Asserts that keys and issuers change only through replacement with a higher generation: a task signed with a removed key, a new task signed with it and a task of a removed issuer are rejected as `task_source_unauthenticated`, a task signed with a retained key is admitted again, and a task admitted before its key expired keeps its handle and expiry while a later presentation of its package is rejected.
  - REQ-TRUST-0055: Asserts that every presentation of a package is verified at its own acceptance time: within one generation a later presentation after key expiry is rejected as `task_source_unauthenticated` and the held handle still resolves with its original expiry after that rejection; after a generation-3 replacement that keeps the expired key, the redelivered package is rejected as `task_source_unauthenticated` with no replay write although a replay record exists; and the same task bytes re-signed with a currently valid key of the same issuer resolve to the held handle within the generation and, after replacement, to the retained acceptance as a redelivery with its original expiry and no new replay record, while the original package is rejected. Redelivery after reinitialization of a new instance and after data clearing is not exercised.

<a id="CC-TRUST-LOCAL-POLICY-LIFECYCLE"></a>

### CC-TRUST-LOCAL-POLICY-LIFECYCLE

- Case: `CC-TRUST-LOCAL-POLICY-LIFECYCLE`
- Result: passed
- Test: `TestCcTrustLocalPolicyLifecycle` in `policy_test.go` (root package)
- Reproduce one case: `go test -count=1 -run '^TestCcTrustLocalPolicyLifecycle$' .`
- Scope: Synthetic core contract cases; Initialize with policy generation 2 and admit a task. Assert from the root package source that only the replacement seam assigns the active policy, adapters or admitter, that admission leaves the active generation unchanged and that Initialize on an initialized client returns already_initialized. Drive replacement by quiesce and reinitialization with fault injection at each implemented step. Validation refuses generation 2 (equal) and 1 (lower), an incomplete policy, retained replay acceptances above the new generation, unavailable replay generations and a panicking trusted clock with policy_rejected. After quiescing, a panic while cancelling the current generation and a delivery store that refuses the new generation leave the client uninitialized, and reserved work outliving the lifecycle deadline leaves it blocked with cleanup_incomplete until close. After each validation failure assert the complete generation-2 policy, its handles and its delivery references stay active and its tasks still admit; after each later failure assert no policy remains and that a reinitialized client rejects generation-2 and generation-3 deliveries with task_policy_unavailable, never mixing registrations. Assert in-flight admission is cancelled with no replay record or handle. On a successful generation-3 replacement that removes the task's registration, assert the previous generation's context is cancelled, its task handles and delivery records no longer resolve, a generation-3 redelivery is rejected as target_not_registered and a generation-2 delivery as policy_rollback; after a generation-4 replacement restoring the registration, a redelivery resolves to the retained acceptance under a new handle, and after clearing, initializing generation 2 or 3 fails closed while generation 4 initializes. Send policy-shaped content through permitted task labels and unknown task members and assert the active generation, issuers and registrations are unchanged. Run concurrent replacements and admissions and assert one highest winning generation. Host authentication of policy updates, run cancellation, Network Context cache invalidation and collector, target, proxy and context endpoint paths are planned separately in CC-TRUST-LOCAL-POLICY-LIFECYCLE-RUNTIME. Requirements: REQ-TRUST-0029, REQ-TRUST-0030, REQ-TRUST-0031, REQ-TRUST-0032.
- Requirement aspects exercised:

  - REQ-TRUST-0029: Asserts from the root package source that only `replacePolicy` assigns the active policy, only `bind`, `close` and `replacePolicy` rebind adapters and nothing reassigns an admitter; that admission, including tasks carrying policy-shaped labels, leaves the active generation, issuers and registrations unchanged; and that `Initialize` on an initialized client returns `already_initialized`.
  - REQ-TRUST-0030: Asserts that replacement refuses an equal or lower generation and an incomplete policy, and validates retained replay generations and builds the successor before quiescing; that it cancels the previous generation's context, including an in-flight admission that then leaves no replay record or handle; that task handles and delivery records of the previous generation stop resolving; that a task whose registration was removed is rejected as `target_not_registered` under the new generation; and that the new generation is active before any new-policy admission. Authentication of the update by the host, run cancellation and Network Context cache invalidation are not exercised.
  - REQ-TRUST-0031: Asserts that every validation failure, including a panicking trusted clock and unavailable replay generations, leaves the complete generation-2 policy, its handles and its delivery references active; that a panic while cancelling and a refused delivery-store generation leave the client uninitialized with no policy, so a reinitialized client rejects generation-2 and generation-3 deliveries as `task_policy_unavailable`; that unfinished reserved work leaves the client blocked until close; that a generation-2 delivery after replacement is `policy_rollback`; that retained acceptances above the new generation refuse it; and that initialization below the highest activated generation fails closed after clearing, while concurrent replacements end on one highest generation.
  - REQ-TRUST-0032: Asserts that policy-shaped content in permitted task labels is admitted as opaque text and that policy-shaped unknown task members are rejected as `task_schema_invalid`, with the active generation, issuers and registrations unchanged in both cases. Collector, target, proxy and context endpoint paths do not exist yet and are not exercised.

<a id="CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING"></a>

### CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING

- Case: `CC-TRUST-REPLAY-DUPLICATE-AND-ROLLBACK-HANDLING`
- Result: passed
- Test: `TestCcTrustReplayDuplicateAndRollbackHandling` in `internal/admission/cross_cutting_conformance_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestCcTrustReplayDuplicateAndRollbackHandling$' ./internal/admission`
- Scope: Synthetic core contract cases; Exercise every replay-table row with a deterministic clock and fake persistent storage: a new replay key continues admission; the same key with identical bytes is idempotent redelivery with no duplicate runs and unchanged expiry; the same key with one changed byte (including whitespace only) is rejected as task_id_conflict; a lower policy generation than the active one is rejected as policy_rollback; an expired task is rejected as task_expired and never reactivated on later redelivery; unavailable or corrupted replay state fails closed with task_trust_unavailable for a remote task and, with the same storage fault injected, for an embedded task, which is never accepted, assigned a handle or made executable without replay state. Assert the persisted record contains only the (issuer_id, task_id) key, a SHA-256 digest equal to an independently computed digest of the exact accepted bytes, the accepted policy generation and acceptance/expiry classes with retention ending 86,400 seconds after task expiry, and that synthetic canary input, target, label, participant and result strings are absent from stored bytes. Post-clearing redelivery rows are planned separately in CC-TRUST-REPLAY-AFTER-CLEARING. Requirements: REQ-TRUST-0021, REQ-TRUST-0022, REQ-TRUST-0023, REQ-TRUST-0024, REQ-TRUST-0038, REQ-LIMIT-0023.
- Requirement aspects exercised:

  - REQ-TRUST-0021: Asserts that acceptance of a new key commits exactly one record through the replay store before `Admit` returns an admitted task, that the record survives a simulated restart and that redelivery after restart resolves to it, against the synthetic replay store and against `persistence.State` over the fault-injectable record store, which is reopened on restart.
  - REQ-TRUST-0022: Asserts that the committed record equals the replay key (issuer_id, task_id), the SHA-256 of the exact accepted bytes computed independently, accepted policy generation 2, the origin and the expiry classes (not_after and not_after plus 86,400 seconds), for embedded and remote origin.
  - REQ-TRUST-0023: Asserts that synthetic canary input, option, label-key, label-value and participant strings, the registry target, test ID and field names are absent from the stored replay bytes of both replay-store backends.
  - REQ-TRUST-0024: Asserts every replay-table row: a new key is admitted; the same bytes redelivered 0, 1 and 23 hours after acceptance are idempotent with an unchanged record and no write; a trailing space, one changed byte, different input and re-encoded bytes are task_id_conflict without changing stored state; a source generation below the active one and a restart under a lower active generation than a retained record are policy_rollback, while equal and higher generations are accepted; redelivery at and after not_after, through and beyond the retention end, is task_expired and never reactivated; and 16 concurrent identical deliveries yield one creation and one commit.
  - REQ-TRUST-0038: Asserts that lookup and generation-read failures, failures before, during and after a write, corrupt and structurally invalid records and a full store each return task_trust_unavailable with no admitted task for embedded and remote origin, that no staged write survives a restart, and that after restart a failed write leaves the task admissible as new while an uncertain durable write resolves as redelivery.
  - REQ-LIMIT-0023: Asserts that the committed record's retention end is 86,400 seconds after the task's not_after.

<a id="CC-TRUST-SIGNED-PACKAGE-DISCLOSURE"></a>

### CC-TRUST-SIGNED-PACKAGE-DISCLOSURE

- Case: `CC-TRUST-SIGNED-PACKAGE-DISCLOSURE`
- Result: passed
- Test: `TestCcTrustSignedPackageErrorsDiscloseNothing` in `internal/admission/signed_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestCcTrustSignedPackageErrorsDiscloseNothing$' ./internal/admission`
- Scope: Synthetic core contract cases; Reject six synthetic signed packages that each fail a different verification check (a bad signature, an issuer moved after signing, an unknown key, malformed JSON, issuance one hour ahead and 98,305 bytes), with canaries in the task payload, key IDs, issuers, signature and public-key prefixes and the raw clock error. Assert that each error formats under %v, %+v, %#v and Error() without any canary or the words ed25519, base64, check, key or signature, and that the six errors produce only the two fixed class strings task_source_unauthenticated and task_limit_exceeded, so the failing check cannot be distinguished beyond its class. Requirements: REQ-TRUST-0035, REQ-TRUST-0052.
- Requirement aspects exercised:

  - REQ-TRUST-0035: Asserts that errors for six failing packages, formatted with `%v`, `%+v`, `%#v` and `Error()`, contain no task payload canary, key ID, issuer, signature or public-key prefix, raw clock error, or the words ed25519, base64, check, key or signature.
  - REQ-TRUST-0052: Asserts that the six failing packages, each failing a different check, yield only the fixed texts `task_source_unauthenticated` and `task_limit_exceeded`, so the failing check is not disclosed beyond its class.

<a id="CC-TRUST-SIGNED-PACKAGE-RETENTION"></a>

### CC-TRUST-SIGNED-PACKAGE-RETENTION

- Case: `CC-TRUST-SIGNED-PACKAGE-RETENTION`
- Result: passed
- Test: `TestCcTrustSignedPackageNotRetained` in `internal/admission/signed_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestCcTrustSignedPackageNotRetained$' ./internal/admission`
- Scope: Synthetic core contract cases; Admit a synthetic signed package through a fake delivery store and the test replay store, then assert the replay record holds the remote issuer, the task_id and the SHA-256 digest of the exact task bytes, not of the package, and that neither the record nor the dumped replay store contains the key ID, the encoded signature, the package bytes or the package digest in hex or base64. Assert the remote delivery record no longer resolves after admission, that a package re-signed later over the same task bytes and presented through a new reference resolves to the same admitted identity as a redelivery, and that a package with a corrupted signature leaves no replay record, staged write or resolvable delivery record. Requirements: REQ-TRUST-0053.
- Requirement aspects exercised:

  - REQ-TRUST-0053: Asserts that after acceptance the replay record holds the remote issuer, `task_id` and the digest of the exact task bytes, and that neither it nor the dumped replay store contains the key ID, encoded signature, package bytes or package digest; that the single-use remote delivery record no longer resolves after acceptance or rejection; that a re-signed package of the same task bytes resolves to the same admitted identity; and that a rejected package writes no replay state.

<a id="CC-TRUST-SIGNED-REMOTE-PACKAGES"></a>

### CC-TRUST-SIGNED-REMOTE-PACKAGES

- Case: `CC-TRUST-SIGNED-REMOTE-PACKAGES`
- Result: passed
- Test: `TestCcTrustSignedRemotePackages` in `internal/admission/signed_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestCcTrustSignedRemotePackages$' ./internal/admission`
- Scope: Synthetic core contract cases; Present signed remote task packages built from synthetic Ed25519 keys through a fake delivery store, a policy holding two keys of the remote issuer and one of another issuer, and a fake clock, changing one property of an otherwise valid package per case. Assert acceptance for a valid package, a second key of the issuer, a package padded with whitespace to exactly 98,304 bytes, acceptance exactly at not_after, issued_at exactly at not_before and issued_at 300 seconds ahead. Assert task_limit_exceeded for 98,305 bytes; task_policy_unavailable with no installed policy; task_source_unauthenticated for malformed packages (not JSON, a bare task, unknown, duplicate, escaped-duplicate or missing members, wrong member types, padded, standard-alphabet or newline-containing base64url, nonzero trailing bits, an empty task, a 63- or 65-byte signature, fractional, offset, lowercase, leap-second or invalid-date issued_at, invalid identifiers), an unknown key_id, a key of another issuer or a package re-wrapped under another issuer, acceptance one second after not_after, issued_at one second before not_before or after not_after, issued_at 301 seconds ahead, an altered task byte, issued_at, key_id or signature, and a signature over another context or over the bare task; and task_trust_unavailable when trusted time is unavailable at check 6. Assert exactly one class, from the first failing check, for packages failing several checks, and that no rejected package reaches the task codec. A validly signed package whose task is invalid or over 65,536 bytes fails its later step instead. Requirements: REQ-TRUST-0043, REQ-TRUST-0046, REQ-TRUST-0047, REQ-TRUST-0048, REQ-TRUST-0049, REQ-TRUST-0050, REQ-TRUST-0052, REQ-LIMIT-0032.
- Requirement aspects exercised:

  - REQ-LIMIT-0032: Asserts that a valid package padded with whitespace to exactly 98,304 bytes is accepted, and that 98,305 bytes, including oversized bytes that are not a package, are `task_limit_exceeded` before anything is decoded.
  - REQ-TRUST-0043: Asserts that a remote task is rejected when its package fails any verification check, and when verification cannot be performed: no installed policy is `task_policy_unavailable` and trusted time unavailable at check 6 is `task_trust_unavailable`.
  - REQ-TRUST-0046: Asserts that step 2 verification precedes parsing: no rejected package reaches the task codec, a validly signed invalid task fails at step 3 as `task_schema_invalid`, and a validly signed task over 65,536 bytes fails the step-3 byte limit as `task_limit_exceeded`.
  - REQ-TRUST-0047: Asserts exactly one class from the first failing check for packages failing several: an oversized package with no policy is `task_limit_exceeded`, no policy with a malformed package `task_policy_unavailable`, a malformed or unknown-key package with unavailable time `task_source_unauthenticated`, an expired key with unavailable time `task_trust_unavailable`, and a bad signature over an invalid task `task_source_unauthenticated`.
  - REQ-TRUST-0048: Asserts `task_source_unauthenticated` for packages that are not a JSON object, carry unknown, duplicate, escaped-duplicate or missing members, members of the wrong type or null, padded, standard-alphabet, newline-containing or nonzero-trailing-bit base64url, an empty task, a 63- or 65-byte signature, fractional, zero-fraction, offset, lowercase, leap-second or invalid-date `issued_at`, or invalid or over-long `key_id` and `issuer_id`.
  - REQ-TRUST-0049: Asserts that a valid package is accepted while an altered or appended task byte, `issued_at` one second later or earlier, `key_id` switched to another key of the same issuer, and signatures over another context string or over the bare task are rejected.
  - REQ-TRUST-0050: Asserts that a signature with one bit flipped and an all-zero signature are rejected; the canonical-encoding edge vectors are recorded under `UT.internal.admission.TestSignatureStrictnessEdgeVectors`.
  - REQ-TRUST-0052: Asserts each of the nine checks in order with its class: the package limit, policy availability, decoding, an unknown key, a key of another issuer or a package re-wrapped under another issuer, trusted time, validity with both ends inclusive and one second outside each end, `issued_at` 300 seconds ahead accepted and 301 rejected, and the signature.

<a id="CC-RESULT-EXECUTION-OBJECT"></a>

### CC-RESULT-EXECUTION-OBJECT

- Case: `CC-RESULT-EXECUTION-OBJECT`
- Result: passed
- Test: `TestCcResultExecutionObject` in `internal/execution/cross_cutting_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcResultExecutionObject$' ./internal/execution`
- Scope: Synthetic core contract cases; Exercise all states/reasons, first termination precedence, measurement presence/absence/partial rules and separation of malformed task errors from results. Interrupt an attempt before and after its test starts and assert not_run before it and incomplete after it (REQ-RESULT-0051), and interrupt it after the test records its terminal outcome and assert the result stays completed with an unchanged state, reason and measurement (REQ-RESULT-0052).
- Requirement aspects exercised:

  - REQ-RESULT-0019: Asserts on written results that `completed` attempts carry no reason and `incomplete` and `not_run` attempts carry one, for positive and negative outcomes, runner failure, host cancellation and the attempt deadline.
  - REQ-RESULT-0020: Asserts that the engine writes `internal_error`, `cancelled_by_host` and `attempt_deadline_exceeded` for a runner failure, a host cancellation and the outer attempt deadline; the other registered reasons are not produced by any implemented component.
  - REQ-RESULT-0021: Asserts that `completed` results carry a measurement, including a negative outcome, that `incomplete` and `not_run` results here carry none, and that when the deadline and a host cancellation both occur the first condition is the reason.
  - REQ-RESULT-0022: Asserts that a plan beyond the registered limits is refused by `Engine.Run` with an error and produces no attempt and no Measurement Result.
  - REQ-RESULT-0051: Asserts that host cancellation and the deadline before the test starts produce `not_run` without Network Context, and after it `incomplete`.
  - REQ-RESULT-0052: Asserts that a cancellation or deadline arriving after the runner returned its outcome leaves the attempt `completed` with its measurement.

<a id="CC-EXECUTION-FIELDS"></a>

### CC-EXECUTION-FIELDS

- Case: `CC-EXECUTION-FIELDS`
- Result: passed
- Test: `TestCcExecutionFields` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcExecutionFields$' .`
- Scope: Synthetic core contract cases; Decode synthetic execution objects through the strict task decoder. Reject an execution object without limits. Accept limits as an empty object (registered defaults applied: max_duration_ms 30000, max_bytes_sent 32768, max_bytes_received 131072, max_connections 8) and assert omitted network_path decodes as direct. Accept network_path direct and proxy and reject any other value. Accept network, device and concurrency when absent or present as typed objects of their defined schemas, and reject each when supplied as a non-object (string, array, null) or with an unknown field. Reject an unknown top-level execution field. Requirements: REQ-EXECUTION-0005.
- Requirement aspects exercised:

  - REQ-EXECUTION-0005: Asserts through `Client.AdmitTask` that an execution object without `limits` is `task_schema_invalid`, that empty `limits` admit the registered defaults 30,000 ms, 32,768 and 131,072 bytes and 8 connections, that an omitted `network_path` is `direct` and `proxy` is kept, that `vpn` and unknown execution members are rejected, and that `network`, `device` and `concurrency` are accepted as typed objects and rejected as a string, array, null or with an unknown member.

<a id="CC-EXECUTION-FIELDS-LIMITS"></a>

### CC-EXECUTION-FIELDS-LIMITS

- Case: `CC-EXECUTION-FIELDS-LIMITS`
- Result: passed
- Test: `TestCcExecutionFieldsLimits` in `internal/execution/cross_cutting_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcExecutionFieldsLimits$' ./internal/execution`
- Scope: Synthetic core contract cases; Exercise an accepted task with limits max_duration_ms 30000, max_bytes_sent 32768, max_bytes_received 131072 and max_connections 8 on a deterministic clock and scripted transport, with a retry bound permitting a second attempt of the same run. In attempt 1 consume exactly 8 connections, 32,768 application bytes sent and 131,072 received and terminate it with a cross-cutting reason that the public retry rules make retryable. Assert attempt 2 (same run_id, attempt 2) starts with full budgets: it can open 8 connections and send and receive the full byte limits, a 9th connection or one byte beyond either byte limit stops it with resource_limit, and its duration deadline is measured from its own start rather than from attempt 1. Requirements: REQ-EXECUTION-0011.
- Requirement aspects exercised:

  - REQ-EXECUTION-0011: Asserts on real loopback sockets that attempt 1 consumes exactly 8 connections, 32,768 bytes sent and 131,072 received and ends with retryable `internal_error`; that attempt 2 of the same run can use the same full budgets, while a ninth connection or one byte beyond either byte limit ends it `incomplete` with `resource_limit`; and that its deadline runs from its own start.

<a id="CC-EXECUTION-FIELDS-CONCURRENCY"></a>

### CC-EXECUTION-FIELDS-CONCURRENCY

- Case: `CC-EXECUTION-FIELDS-CONCURRENCY`
- Result: passed
- Test: `TestCcExecutionFieldsConcurrency` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcExecutionFieldsConcurrency$' .`
- Scope: Synthetic core contract cases; Schedule synthetic tasks on a deterministic clock with scripted transports that record start and end times: two shared tasks S1 and S2 and one exclusive task E, submitted in the orders S1,S2,E and E,S1,S2 and with E submitted while S1 is running. Assert S1 and S2 may overlap, that E's execution interval never overlaps any other measurement's interval, and that no shared task starts while E is active. Assert that when E cannot run alone it either waits or ends not_run per the public constraint-evaluation rules, without starting any network operation while another measurement is active. Requirements: REQ-EXECUTION-0027.
- Requirement aspects exercised:

  - REQ-EXECUTION-0027: Asserts through `Client.AdmitTask` and `Client.StartRun` on the injected clock that two shared runs are active at once, that an exclusive run never starts while another run is active and no run starts while it is active, in the orders S1, S2, E and E, S1, S2 and with E submitted while S1 runs, and that a waiting start returns `run_not_due` without reaching the path, the runner or the output.

<a id="CC-SAFETY-USER-CONSENT"></a>

### CC-SAFETY-USER-CONSENT

- Case: `CC-SAFETY-USER-CONSENT`
- Result: passed
- Test: `TestCcSafetyUserConsent` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcSafetyUserConsent$' .`
- Scope: Synthetic core contract cases; Combine task requires_user_consent true and false with a fake host consent provider reporting valid, absent, expired and unavailable consent, and with the local consent requirement off and on. Task true with valid consent: an attempt may start. Task true with absent, expired or unavailable consent: no attempt starts and no Measurement Result is produced. Task false with the local requirement on and absent consent: no attempt starts. Task false with the local requirement off: no consent requirement is added, and risk and target checks still apply. Admission records the consent requirement as required when the task, issuer permissions or local policy requires it, and a task value of false never clears an issuer or local requirement. Requirements: REQ-SAFETY-0020, REQ-SAFETY-0021, REQ-SAFETY-0022.
- Requirement aspects exercised:

  - REQ-SAFETY-0020: Asserts through `Client.AdmitTask` and `Client.StartRun` on the injected clock that a task requiring consent starts an attempt only with valid consent, and with absent, ended or unavailable consent or no consent capability starts no attempt and writes no Measurement Result, including before a retry.
  - REQ-SAFETY-0021: Asserts that the recorded consent requirement is set when the task, the issuer permissions or local policy requires consent, that a task value of `false` never clears an issuer or local requirement, and that consent is not queried when nothing requires it.
  - REQ-SAFETY-0022: Asserts that with no consent requirement a task runs without a consent query while risk review and target registration still reject a high-risk task and an unregistered target.

<a id="CC-SCHEDULE-FIELDS"></a>

### CC-SCHEDULE-FIELDS

- Case: `CC-SCHEDULE-FIELDS`
- Result: passed
- Test: `TestCcScheduleFields` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcScheduleFields$' .`
- Scope: Synthetic core contract cases; Exercise omitted/present and conditional schedule fields, defaults, wrong types and unknown properties.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0002: Asserts through `Client.AdmitTask` and `Client.StartRun` on the injected clock that a schedule with only `not_after` is admitted with one run and one attempt, that every optional field can be present, that omitting `not_after` is rejected for an embedded one-time, a remote and a repeated task, and that wrong types and unknown members are `task_schema_invalid` with no replay record.

<a id="CC-SCHEDULE-SCHEDULING-MODEL"></a>

### CC-SCHEDULE-SCHEDULING-MODEL

- Case: `CC-SCHEDULE-SCHEDULING-MODEL`
- Result: passed
- Test: `TestCcScheduleSchedulingModel` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcScheduleSchedulingModel$' .`
- Scope: Synthetic core contract cases; Use a deterministic clock and a synthetic task with repeat.mode interval, interval 21600, jitter 300, max_runs 3, a not_after after the third due run and retry.max_attempts 2 with retry_on internal_error and backoff mode fixed, initial_delay_ms 30000, max_delay_ms 30000. Script attempt 1 of run 1 to end with internal_error and every other attempt to complete. Assert three distinct run identities, run 1 holding attempts 1 and 2 and runs 2 and 3 holding attempt 1 only, the run counter reaching 3 and attempt numbering scoped to each run. Assert a retry attempt shares its run identity and never creates a new run, and that each interval-due run receives a new run identity. With repeat.mode none and the same retry policy, assert exactly one run whose attempt count is at most 2. Requirements: REQ-SCHEDULE-0003, REQ-SCHEDULE-0004, REQ-SCHEDULE-0005, REQ-SCHEDULE-0006.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0003: Asserts through `Client.AdmitTask` and `Client.StartRun` on the injected clock that an interval schedule with `max_runs` 3 creates exactly three runs with distinct run identities and that `repeat` `none` creates exactly one run however often the host starts it.
  - REQ-SCHEDULE-0004: Asserts that run 1 holds attempts 1 and 2 under one run identity, with attempt numbering scoped to each run.
  - REQ-SCHEDULE-0005: Asserts that each interval-due run receives a new run identity and a new run count.
  - REQ-SCHEDULE-0006: Asserts that the retry is attempt 2 of the existing run, shares its run identity and creates no run or run count.

<a id="CC-SCHEDULE-SCHEDULING-MODEL-START-TIME"></a>

### CC-SCHEDULE-SCHEDULING-MODEL-START-TIME

- Case: `CC-SCHEDULE-SCHEDULING-MODEL-START-TIME`
- Result: passed
- Test: `TestCcScheduleSchedulingModelStartTime` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcScheduleSchedulingModelStartTime$' .`
- Scope: Synthetic core contract cases; Test before/at/after starts_at plus permitted null/absence; do not run early or require exact timing.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0007: Asserts through `Client.AdmitTask` and `Client.StartRun` on the injected clock that a start one second before `starts_at` returns `run_not_due` without an attempt, and starts at and after it run.
  - REQ-SCHEDULE-0008: Asserts that runs five and 22 hours after `starts_at` still start, and that the result's `started_at` records the actual start to minute precision rather than `starts_at`.
  - REQ-SCHEDULE-0009: Asserts that with `starts_at` `null` or absent a run starts at acceptance.

<a id="CC-SCHEDULE-SCHEDULING-MODEL-TIME-LIMIT"></a>

### CC-SCHEDULE-SCHEDULING-MODEL-TIME-LIMIT

- Case: `CC-SCHEDULE-SCHEDULING-MODEL-TIME-LIMIT`
- Result: passed
- Test: `TestCcScheduleSchedulingModelTimeLimit` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcScheduleSchedulingModelTimeLimit$' .`
- Scope: Synthetic core contract cases; Test acceptance/issuance-relative expiry at boundaries and the reviewed embedded one-time exception.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0010: Asserts through `Client.AdmitTask` and `Client.StartRun` on the injected clock that a daily window inside the lifetime permits a run before `not_after` and none in the next window after it.
  - REQ-SCHEDULE-0011: Asserts for signed remote packages that `not_after` at acceptance is `task_expired`, that one week after the earlier of issuance and acceptance is admitted and one second more is `task_limit_exceeded`, and that an embedded one-time task without `not_after` fails closed at the facade, which takes no release-review input, while `validation.ScheduleAt` honors the trusted exception only for one run of an embedded task.

<a id="CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW"></a>

### CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW

- Case: `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW`
- Result: passed
- Test: `TestCcScheduleSchedulingModelTimeWindow` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcScheduleSchedulingModelTimeWindow$' .`
- Scope: Synthetic core contract cases; Drive the scheduler with a deterministic clock and fake platform timezone UTC, admitting NET-01 tasks whose runs are otherwise eligible. With `time_window` `{"timezone": "local", "ranges": [{"start": "12:00", "end": "18:00"}]}`, advance the clock to local 11:59, 12:00, 17:59 and 18:00 and assert a run starts only at 12:00 and 17:59 (start inclusive, end exclusive) and not at 11:59 or 18:00. With the midnight-crossing range `22:00`-`06:00`, assert no start at 21:59 and 06:00 and a start at 22:00 and 05:59. With two ranges `08:00`-`10:00` and `14:00`-`16:00`, assert no start at 12:00 and a start at 14:00. Assert a run that becomes due outside the window is delayed to the next allowed range start before `not_after` rather than started early. Requirements: REQ-SCHEDULE-0012.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0012: Asserts through `Client.AdmitTask` and `Client.StartRun` on the injected clock with the device zone UTC that 12:00–18:00 starts at 12:00 and 17:59 and not at 11:59 or 18:00, that 22:00–06:00 starts at 22:00 and 05:59 and not at 21:59 or 06:00, that two ranges start in the second and not between them, that a run due outside the window waits until the range opens, and that a range opening after `not_after` never starts.

<a id="CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-FIELDS"></a>

### CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-FIELDS

- Case: `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-FIELDS`
- Result: passed
- Test: `TestCcScheduleSchedulingModelTimeWindowFields` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcScheduleSchedulingModelTimeWindowFields$' .`
- Scope: Synthetic core contract cases; Validate synthetic `time_window` objects at admission and evaluate accepted ones with a deterministic clock and a fake platform timezone. Reject `ranges` omitted and `ranges: []` (0 ranges), accept 1 range and 4 non-overlapping ranges (`00:00`-`02:00`, `04:00`-`06:00`, `08:00`-`10:00`, `12:00`-`14:00`) and reject 5 ranges, with a bounded non-measurement admission error. Reject two identical ranges `08:00`-`12:00`, overlapping ranges `08:00`-`12:00` and `11:00`-`14:00`, and a midnight-crossing overlap `22:00`-`06:00` with `05:00`-`07:00`. Accept `timezone` omitted and assert it is treated as `local`. Reject `timezone` values `UTC`, `Europe/Berlin`, `""` and `Local` with a bounded non-measurement admission error and no run or attempt created. With range `12:00`-`18:00` and `timezone: local`, set the fake device timezone to UTC at fixed instant 10:30Z (local 10:30, no start), then switch it to UTC+03:00 at the same instant (local 13:30) and assert a run starts, demonstrating `local` uses the current device-reported timezone. Requirements: REQ-SCHEDULE-0013, REQ-SCHEDULE-0014, REQ-SCHEDULE-0043.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0013: Asserts through `Client.AdmitTask` and `Client.StartRun` on the injected clock that omitted and empty `ranges` are `task_schema_invalid`, one and four non-overlapping ranges are admitted, five are `task_limit_exceeded`, identical, overlapping and midnight-overlapping ranges are rejected, and an omitted `timezone` behaves as `local`.
  - REQ-SCHEDULE-0014: Asserts that `local` uses the current device-reported zone: at 10:30Z a 12:00–18:00 window does not start in UTC and starts once the device zone is UTC+03:00 at the same instant.
  - REQ-SCHEDULE-0043: Asserts that the `timezone` values `UTC`, `Europe/Berlin`, empty and `Local` are rejected with no run accounting.

<a id="CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-RANGE-FIELDS"></a>

### CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-RANGE-FIELDS

- Case: `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-RANGE-FIELDS`
- Result: passed
- Test: `TestCcScheduleSchedulingModelTimeWindowRangeFields` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcScheduleSchedulingModelTimeWindowRangeFields$' .`
- Scope: Synthetic core contract cases; Test HH:mm syntax, inclusive start/exclusive end, midnight expansion and overlapping ranges after expansion.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0015: Asserts through `Client.AdmitTask` and `Client.StartRun` on the injected clock that a missing `start` or `end` or an unknown range member is rejected, that a start at the range start runs and a start at its end does not.
  - REQ-SCHEDULE-0016: Asserts that 00:00, 23:59, 09:05 and 17:30 are accepted and 24:00, 7:00, 12:60, 12:00:00, a word, an empty string and a number are rejected.
  - REQ-SCHEDULE-0017: Asserts that 23:00–01:00 starts at 23:30 and 00:30 and not at 22:59 or 01:00, and that ranges overlapping only after midnight expansion are rejected while adjacent ones are admitted.

<a id="CC-SCHEDULE-SCHEDULING-MODEL-REPEAT"></a>

### CC-SCHEDULE-SCHEDULING-MODEL-REPEAT

- Case: `CC-SCHEDULE-SCHEDULING-MODEL-REPEAT`
- Result: passed
- Test: `TestCcScheduleSchedulingModelRepeat` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcScheduleSchedulingModelRepeat$' .`
- Scope: Synthetic core contract cases; Distinguish none and interval, disallow interval-only fields for none and generate the bounded number of runs.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0018: Asserts through `Client.AdmitTask` and `Client.StartRun` on the injected clock that omitted and `none` repeat create one run over six host calls, that `interval` creates exactly `max_runs` runs (1, 2 and 5) over ten calls, each writing one result, with distinct run identities among the results the pending store still retains, and that interval fields with `none`, an unregistered or missing mode and unknown repeat members are `task_schema_invalid`.

<a id="CC-SCHEDULE-SCHEDULING-MODEL-RETRY"></a>

### CC-SCHEDULE-SCHEDULING-MODEL-RETRY

- Case: `CC-SCHEDULE-SCHEDULING-MODEL-RETRY`
- Result: passed
- Test: `TestCcScheduleSchedulingModelRetry` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcScheduleSchedulingModelRetry$' .`
- Scope: Synthetic core contract cases; Keep run identity/max_runs accounting stable across the second attempt; a terminal negative measurement creates no automatic retry.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0022: Asserts through `Client.AdmitTask` and `Client.StartRun` on the injected clock that a retryable `internal_error` produces a second attempt after the 30-second fixed backoff.
  - REQ-SCHEDULE-0023: Asserts that the second attempt is the last one: a run has at most two attempts.
  - REQ-SCHEDULE-0024: Asserts that the retry keeps the run identity and run count, so `max_runs` 2 still permits exactly two runs.
  - REQ-SCHEDULE-0025: Asserts that a `connection_refused` outcome is a `completed` attempt with its measurement and is not retried although `internal_error` is retryable.

<a id="CC-SCHEDULE-SCHEDULING-MODEL-RETRY-RETRY-FIELDS"></a>

### CC-SCHEDULE-SCHEDULING-MODEL-RETRY-RETRY-FIELDS

- Case: `CC-SCHEDULE-SCHEDULING-MODEL-RETRY-RETRY-FIELDS`
- Result: passed
- Test: `TestCcScheduleSchedulingModelRetryRetryFields` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcScheduleSchedulingModelRetryRetryFields$' .`
- Scope: Synthetic core contract cases; Enumerate permitted and forbidden reasons; exercise attempts=1/2 conditional fields, fixed/exponential delays and jitter at expiry and delay bounds.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0026: Asserts through `Client.AdmitTask` and `Client.StartRun` on the injected clock that `max_attempts` 2 needs a non-empty, duplicate-free `retry_on` and a `backoff`, that `max_attempts` is required, and that `max_attempts` 1 is admitted.
  - REQ-SCHEDULE-0027: Asserts that each of the four registered retry reasons and all four together are admitted, and that `cancelled_by_host`, `constraints_unsatisfied`, `proxy_reconfigured`, `resource_limit`, `attempt_deadline_exceeded` and a measurement outcome are rejected.
  - REQ-SCHEDULE-0028: Asserts that `fixed` with a multiplier, `exponential` without one and an unregistered mode are rejected, that multiplier 3 is `task_limit_exceeded`, and that fixed and exponential backoff each produce one delay of the initial delay plus local jitter.
  - REQ-SCHEDULE-0029: Asserts that `retry_on` or `backoff` with one attempt is rejected, that the default 10-second jitter cap and an explicit cap apply, that the realized delay is capped at `max_delay_ms`, and that a delay reaching `not_after` is not taken.

<a id="CC-SCHEDULE-VALIDATION"></a>

### CC-SCHEDULE-VALIDATION

- Case: `CC-SCHEDULE-VALIDATION`
- Result: passed
- Test: `TestCcScheduleValidation` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcScheduleValidation$' .`
- Scope: Synthetic core contract cases; Mutate each listed schedule constraint; reject expired/invalid schedules, and redeliver identical tasks without extending expiry or adding schedules.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0033: Asserts through `Client.AdmitTask` and `Client.StartRun` on the injected clock that unknown members of the schedule, time window, range, repeat, retry, backoff and limits objects are `task_schema_invalid`.
  - REQ-SCHEDULE-0036: Asserts that timestamps with an offset or without a zone are rejected and that `starts_at` equal to or after `not_after` is rejected while an earlier one is admitted.
  - REQ-SCHEDULE-0037: Asserts that duplicate ranges and ranges overlapping after midnight expansion are rejected.
  - REQ-SCHEDULE-0038: Asserts that an omitted repeat is admitted as `none` and that interval fields with `none` are rejected.
  - REQ-SCHEDULE-0039: Asserts that a repeated task without `not_after`, `max_runs` or jitter is rejected and with all three is admitted.
  - REQ-SCHEDULE-0040: Asserts that `retry_on` without a second attempt, a second attempt without backoff and a measurement-outcome reason are rejected.
  - REQ-SCHEDULE-0041: Asserts that an exact redelivery an hour later returns the same task handle with the original expiry, and that the run count and expiry continue: the next run starts at its interval and no run starts after the accepted `not_after`.
  - REQ-SCHEDULE-0042: Asserts that an expired or invalid schedule is an admission failure with no handle and no Measurement Result written.

<a id="CC-SCHEDULE-VALIDATION-ELIGIBILITY-RULES"></a>

### CC-SCHEDULE-VALIDATION-ELIGIBILITY-RULES

- Case: `CC-SCHEDULE-VALIDATION-ELIGIBILITY-RULES`
- Result: passed
- Test: `TestCcScheduleValidationEligibilityRules` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcScheduleValidationEligibilityRules$' .`
- Scope: Synthetic core contract cases; Deny each eligibility predicate independently for initial and retry attempts; assert no early/out-of-window/expired execution and preserve stricter local veto.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0034: Asserts through `Client.AdmitTask` and `Client.StartRun` on the injected clock that denying one predicate at a time (no registered runner, before `starts_at`, at `not_after`, outside the window, `max_runs` reached, an unsatisfied device constraint, absent consent) starts no run, reserves no run and reaches no path, runner or output, and that a constraint failure, ended consent, `not_after` or a closing window between attempt 1 and the retry prevents the retry.
  - REQ-SCHEDULE-0035: Asserts that the probe declines an otherwise eligible run when the production pending store has no capacity because its only bundle slot holds a stored bundle.

<a id="CC-LIMIT-SCHEDULE-AND-RETRY-LIMITS"></a>

### CC-LIMIT-SCHEDULE-AND-RETRY-LIMITS

- Case: `CC-LIMIT-SCHEDULE-AND-RETRY-LIMITS`
- Result: passed
- Test: `TestCcLimitScheduleAndRetryLimits` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcLimitScheduleAndRetryLimits$' .`
- Scope: Synthetic core contract cases; Parameterize every bound/default, issuance skew and trusted-time absence, embedded one-time exception and forbidden measurement-outcome retries.
- Requirement aspects exercised:

  - REQ-LIMIT-0009: Asserts for signed remote packages that `not_after` one week after issuance is admitted, one second more is `task_limit_exceeded` and a missing `not_after` is rejected.
  - REQ-LIMIT-0010: Asserts that a package issued 300 seconds after the trusted clock is admitted, 301 seconds is `task_source_unauthenticated`, and unavailable trusted time is `task_trust_unavailable`.
  - REQ-LIMIT-0011: Asserts through `Client.AdmitTask` and `Client.StartRun` on the injected clock every registered schedule and retry bound: runs 1 and 28 admitted, 0 and 29 rejected; ranges 4 admitted, 5 rejected; interval 21,600 and 604,800 admitted, 21,599 and 604,801 rejected; jitter 300 and 1,800 admitted, 299 and 1,801 rejected; attempts 3 rejected; initial delay 30,000 and 300,000 admitted, 29,999 and 300,001 rejected; maximum delay below the initial delay and 600,001 rejected; retry jitter 0 and 60,000 admitted, 60,001 rejected; and the defaults of one run, one attempt and a 10-second retry jitter cap.
  - REQ-LIMIT-0012: Asserts that a measurement outcome in `retry_on` is rejected at admission.
  - REQ-LIMIT-0013: Asserts through `validation.ScheduleAt` that the embedded one-time exception admits an embedded task without `not_after` only for one run of a non-repeating task, and that a repeating embedded task is limited to the one-week lifetime.

<a id="CC-LIMIT-RESULT-AND-BUNDLE-LIMITS"></a>

### CC-LIMIT-RESULT-AND-BUNDLE-LIMITS

- Case: `CC-LIMIT-RESULT-AND-BUNDLE-LIMITS`
- Result: passed
- Test: `TestCcLimitResultAndBundleLimits` in `internal/bundle/cross_cutting_conformance_test.go` (package `internal/bundle`)
- Reproduce one case: `go test -count=1 -run '^TestCcLimitResultAndBundleLimits$' ./internal/bundle`
- Scope: Synthetic core contract cases; Exercise every encoded/count/pending-storage maximum; split bundles or return bounded persistence error without truncating valid results.
- Requirement aspects exercised:

  - REQ-LIMIT-0021: Asserts that an encoded Measurement Result of 65,535 and 65,536 bytes is accepted and 65,537 bytes is rejected with the bounds class on the strict import boundary; that through the production pending store the 16th result joins its bundle and the 17th starts another; that a bundle of 262,143 and 262,144 encoded bytes is stored as one bundle; that 32 sealed bundles are held and a 33rd is rejected without changing storage; and that sixteen bundles totalling 4,194,303 and 4,194,304 bytes are held, after which a further result is rejected and capacity is zero.
  - REQ-LIMIT-0022: Asserts that a sixth result that would make its bundle 262,145 bytes starts another bundle instead, that results rejected at the bundle or aggregate maximum leave storage byte-for-byte unchanged, that every stored result equals its accepted encoding byte for byte, and that results of the largest fixture size are accepted while reported capacity is positive until it reaches zero.

<a id="CC-LINKAGE-CLEAR-DATA-CONTRACT"></a>

### CC-LINKAGE-CLEAR-DATA-CONTRACT

- Case: `CC-LINKAGE-CLEAR-DATA-CONTRACT`
- Result: passed
- Test: `TestCcLinkageClearDataContract` in `internal/persistence/cross_cutting_conformance_test.go` (package `internal/persistence`)
- Reproduce one case: `go test -count=1 -run '^TestCcLinkageClearDataContract$' ./internal/persistence`
- Scope: Synthetic core contract cases; Go-owned rows only. Seed a storage scope with pending bundles, run accounting, replay records and a synthetic linkage key, and fault-inject each clearing step before and after it takes effect. Assert an explicit incomplete error for every failure, a durable clearing marker that persists across a simulated restart while any deletion step may have run, idempotent recovery clearing from every partial state that leaves every record class and the key deleted, and that clearing derives no replacement key and has no transport. Cancellation of active attempts with bounded cleanup is asserted by the root UT cases. Host user-interface and operating-system reset wiring, native keystores, reinstall sentinels and host-owned copies are planned in CC-LINKAGE-CLEAR-DATA-CONTRACT-HOST (DEV-455). Requirements: REQ-LINKAGE-0023, REQ-LINKAGE-0024.
- Requirement aspects exercised:

  - REQ-LINKAGE-0023: Asserts with `persistence.Clear`, over a scope seeded with a replay record, its run record and a pending record and with a synthetic linkage key, that each of the 13 record-port steps (the marker write; the listing, deletion and confirming listing of pending, run and replay records; and the listing, deletion and confirming listing of the marker class), failed before and after it takes effect, returns `ErrClearIncomplete` and never success (item 7); that after a simulated restart `Clearing` reports the marker exactly while it is stored and no partial clear leaves the scope unmarked; that a later clear deletes the pending, run and replay records and the linkage key (items 3 to 5); and that a failed key deletion keeps the scope marked until a retry deletes the key. Active attempts, caches, epochs, sentinels and in-memory linkage material are not exercised by this case.
  - REQ-LINKAGE-0024: Asserts that after every injected failure two consecutive clears both complete and leave the scope empty, and that `Clear` takes only the record port and erasers, so no key derivation or transport is reachable from it; the absence of network packages in its dependency closure is shown by `TestClearingHasNoNetworkDependency`, not by observed traffic.

<a id="CC-TRUST-REPLAY-AFTER-CLEARING"></a>

### CC-TRUST-REPLAY-AFTER-CLEARING

- Case: `CC-TRUST-REPLAY-AFTER-CLEARING`
- Result: passed
- Test: `TestCcTrustReplayAfterClearing` in `internal/admission/cross_cutting_conformance_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestCcTrustReplayAfterClearing$' ./internal/admission`
- Scope: Synthetic core contract cases; After clearMeasurementData deletes replay state with a deterministic clock and fake persistent storage, redeliver previously accepted exact task bytes: an unexpired task is evaluated as new, receives a fresh replay record with unchanged expiry derived from its not_after and creates no record for the cleared acceptance; an expired task is still rejected as task_expired; a failed or incomplete deletion is reported explicitly and leaves no task admitted without replay state. Requirements: REQ-TRUST-0024, REQ-TRUST-0025.
- Requirement aspects exercised:

  - REQ-TRUST-0024: Asserts against the production `persistence.State` over the test-only record store that, after `persistence.Clear` and `State.Forget`, exact embedded and remote task bytes redelivered one hour after acceptance are not a redelivery and write exactly one fresh replay record holding the key, digest, generation, origin and expiry derived from the task's `not_after`, which then resolves a later redelivery; and that after a failed clearing the retained record still resolves redelivery to the original acceptance.
  - REQ-TRUST-0025: Asserts that clearing removes replay state, so an unexpired redelivered task is evaluated as new within its expiry, while the same task redelivered at its `not_after` after clearing is rejected as `task_expired` with no replay write; and that a failed clearing is reported as incomplete and leaves the scope marked, after which recovery clearing removes the record. The root client's refusal of admission while the marker exists is asserted by `TestClearFaultAtEachStep`, not by this case.

## Cases

<a id="UT.internal.admission.TestDecodePackageAndMessage"></a>

### UT.internal.admission.TestDecodePackageAndMessage

- Case: `UT.internal.admission.TestDecodePackageAndMessage`
- Result: passed
- Test: `TestDecodePackageAndMessage` in `internal/admission/signed_internal_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestDecodePackageAndMessage$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestDecodePackageAndMessage in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0048: Asserts that a valid encoded package decodes to its exact members, including task bytes containing NUL and 0xff, and that formatting a decoded package exposes no member.
  - REQ-TRUST-0049: Asserts that the message is the length-prefixed concatenation of `barometer-signed-task-v1`, `key_id`, `issuer_id`, `issued_at` and the task bytes, byte-identical to an independent construction, and that moving a byte across the `key_id` and `issuer_id` boundary changes the message and fails verification.

<a id="UT.internal.admission.TestGateAcceptsReviewedOrigins"></a>

### UT.internal.admission.TestGateAcceptsReviewedOrigins

- Case: `UT.internal.admission.TestGateAcceptsReviewedOrigins`
- Result: passed
- Test: `TestGateAcceptsReviewedOrigins` in `internal/admission/admit_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateAcceptsReviewedOrigins$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateAcceptsReviewedOrigins in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0040: Asserts that a remote task passes the gates only as a signed package verified against a trusted key of its issuer in the active policy, and an embedded task only through the reviewed embedded-origin installer.
  - REQ-TRUST-0041: Asserts that an embedded task installed through the embedded-origin installer passes the gates with embedded origin, its issuer, the trusted acceptance time and the registry entry, decoded through the registry-selected synthetic codec with defaults applied, and that the same bytes pass as a signed remote package.
  - REQ-TRUST-0044: Asserts that remote packages issued one hour before, at and 300 seconds after the fake trusted clock pass the gates with a verified remote source whose issuer and issuance time are those of the package.

<a id="UT.internal.admission.TestGateAdmissionLimitHandling"></a>

### UT.internal.admission.TestGateAdmissionLimitHandling

- Case: `UT.internal.admission.TestGateAdmissionLimitHandling`
- Result: passed
- Test: `TestGateAdmissionLimitHandling` in `internal/admission/authority_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateAdmissionLimitHandling$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateAdmissionLimitHandling in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0026: Asserts that complete task JSON, task_id, input, options, label count, max_runs, max_duration_ms, max_bytes_sent, max_bytes_received and max_connections are each accepted at their registered maximum and rejected one unit above it as task_limit_exceeded at admission, with an unused recording transport that sees zero calls. The NET-05 nine-step row and runtime limit handling are not tested.
  - REQ-LIMIT-0030: Asserts that label values naming limits are kept as opaque labels with effective limits unchanged, and that limit-like members in input, options, schedule, execution and safety and an unknown top-level field are task_schema_invalid; also asserts that issuer or deployment policy ceilings above the registered attempt, run, attempt-count and validity maxima are refused by policy construction.

<a id="UT.internal.admission.TestGateAuthorityPrecedence"></a>

### UT.internal.admission.TestGateAuthorityPrecedence

- Case: `UT.internal.admission.TestGateAuthorityPrecedence`
- Result: passed
- Test: `TestGateAuthorityPrecedence` in `internal/admission/authority_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateAuthorityPrecedence$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateAuthorityPrecedence in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0047: Asserts first-failing-step precedence around step 7 and step 8: an unknown top-level field, an expired task and a value above a registered maximum each keep their earlier class with an unknown issuer, an unknown issuer precedes invalid input, a policy ceiling precedes path authority, invalid input precedes path authority, and path authority precedes the risk bound.

<a id="UT.internal.admission.TestGateCompositeBeforeProxyCompatibility"></a>

### UT.internal.admission.TestGateCompositeBeforeProxyCompatibility

- Case: `UT.internal.admission.TestGateCompositeBeforeProxyCompatibility`
- Result: passed
- Test: `TestGateCompositeBeforeProxyCompatibility` in `internal/admission/proxy_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateCompositeBeforeProxyCompatibility$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateCompositeBeforeProxyCompatibility in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0018: Asserts that the rejected NET-05 proxy task causes no call or open connection on the unused recording transport.
  - REQ-TRUST-0036: Asserts that a registered and enabled NET-05 proxy task is `task_unsupported` because composite steps cannot be matched.
  - REQ-TRUST-0039: Asserts that the NET-05 registry rule is not consulted, so a NET-05 profile is not classified by a placeholder rule; rejecting a profile with an incompatible step as `task_schema_invalid` needs the NET-05 codec and is not tested.

<a id="UT.internal.admission.TestGateConsentRequirement"></a>

### UT.internal.admission.TestGateConsentRequirement

- Case: `UT.internal.admission.TestGateConsentRequirement`
- Result: passed
- Test: `TestGateConsentRequirement` in `internal/admission/safety_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateConsentRequirement$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateConsentRequirement in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0021: Asserts that a task requires_user_consent of false, with issuer and deployment consent requirements off, is admitted with no consent requirement recorded on the gated value.
  - REQ-SAFETY-0022: Asserts for all eight combinations of task, issuer and deployment consent requirement that the recorded requirement is set whenever the issuer or the deployment requires consent, whatever the task value. Consent validity at attempt start is not tested.
  - REQ-SAFETY-0028: Asserts that a task value of false never clears an issuer or deployment consent requirement, and that declaring requires_user_consent true does not relax a risk bound that rejects the same task.

<a id="UT.internal.admission.TestGateEffectiveRisk"></a>

### UT.internal.admission.TestGateEffectiveRisk

- Case: `UT.internal.admission.TestGateEffectiveRisk`
- Result: passed
- Test: `TestGateEffectiveRisk` in `internal/admission/safety_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateEffectiveRisk$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateEffectiveRisk in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0004: Asserts with synthetic registry baselines low, medium and unknown that the registered baseline is an input to the effective risk recorded on the gated value.
  - REQ-SAFETY-0005: Asserts that a declared level above the registered baseline raises effective risk and that no declared level reduces it below the baseline.
  - REQ-SAFETY-0023: Asserts for every combination of registered baseline (low, medium, unknown), declared level and matched target-registration baseline (low, medium, high, unknown), under a policy permitting every level, that the recorded effective risk equals the most restrictive input as computed independently in the test, with unknown dominating. Options, schedule or other local assessments are not inputs and are not tested.
  - REQ-SAFETY-0024: Asserts that effective risk is never below the registered baseline for any declared level.
  - REQ-SAFETY-0025: Asserts that a low declaration on a medium baseline is admitted with effective risk medium, never low.
  - REQ-SAFETY-0026: Asserts that a declaration higher than the registered baseline is accepted and becomes the effective risk.

<a id="UT.internal.admission.TestGateEgressContextPrecedence"></a>

### UT.internal.admission.TestGateEgressContextPrecedence

- Case: `UT.internal.admission.TestGateEgressContextPrecedence`
- Result: passed
- Test: `TestGateEgressContextPrecedence` in `internal/admission/capability_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateEgressContextPrecedence$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateEgressContextPrecedence in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0036: Asserts that an unrecognized risk factor type together with an egress context requirement is rejected as task_unsupported.
  - REQ-TRUST-0047: Asserts that an unknown field, task_id over 128 bytes, expiry, a value above a registered limit and an invalid risk factor keep their own class ahead of the step-6 egress check, and that the egress check precedes invalid input, a test/version not permitted, policy rollback, a label outside the permitted namespaces, a path not permitted, an unregistered target and a risk bound, with the registry codec never reached.

<a id="UT.internal.admission.TestGateErrorsCarryNoData"></a>

### UT.internal.admission.TestGateErrorsCarryNoData

- Case: `UT.internal.admission.TestGateErrorsCarryNoData`
- Result: passed
- Test: `TestGateErrorsCarryNoData` in `internal/admission/admit_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateErrorsCarryNoData$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateErrorsCarryNoData in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0033: Asserts that each rejection is a task.AdmissionError whose text is the fixed class string, that repeating an identical failing delivery returns the same error, and that rejection leaves the delivery record resolvable and a later valid task accepted.
  - REQ-TRUST-0035: Embeds canaries in payload, target, labels, issuer, codec error and clock error text and asserts none appears in the error text, %v, %+v or %#v formatting, and that errors.Unwrap returns nil.

<a id="UT.internal.admission.TestGateIssuerAndLocalCeilings"></a>

### UT.internal.admission.TestGateIssuerAndLocalCeilings

- Case: `UT.internal.admission.TestGateIssuerAndLocalCeilings`
- Result: passed
- Test: `TestGateIssuerAndLocalCeilings` in `internal/admission/authority_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateIssuerAndLocalCeilings$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateIssuerAndLocalCeilings in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0030: Asserts that a task value one unit beyond an issuer or deployment ceiling is task_limit_exceeded for every task-bounded ceiling, including repeat interval, jitter and retry delay minima, and is never clamped.
  - REQ-TRUST-0010: Asserts that each numeric ceiling narrowed in the issuer, and separately in the deployment policy, binds while the other side is permissive, and that result and retention ceilings are carried as the lower of the two sides in both directions.
  - REQ-TRUST-0011: Asserts that issuer and deployment ceilings bound attempt duration, bytes sent and received, connections, max_runs, retry max_attempts, validity lifetime, repeat interval and jitter minima, and retry initial delay, maximum delay and jitter.
  - REQ-TRUST-0012: Asserts for each task-bounded ceiling that a value at the bound and one unit below it is accepted, the lower value is carried as the effective value, and one unit beyond is rejected as task_limit_exceeded.

<a id="UT.internal.admission.TestGateIssuerAuthorization"></a>

### UT.internal.admission.TestGateIssuerAuthorization

- Case: `UT.internal.admission.TestGateIssuerAuthorization`
- Result: passed
- Test: `TestGateIssuerAuthorization` in `internal/admission/authority_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateIssuerAuthorization$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateIssuerAuthorization in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0010: Narrows test/version, network path, maximum risk, medium-risk enablement, unknown-risk exception, label namespaces and NET-05 composite enablement in the issuer and then in the deployment policy, with the other side permissive, and asserts each narrowing binds.
  - REQ-TRUST-0011: Asserts that permitted test versions, direct and proxy path use, maximum risk, medium-risk and NET-05 enablement, the unknown-risk exception and label namespaces at a '.' boundary are each enforced as task_issuer_not_authorized when not granted. Registration-reference matching and consent evaluation are not tested.
  - REQ-TRUST-0047: Asserts task_issuer_not_authorized for a direct or proxy path not permitted by issuer or deployment policy and for risk above an issuer or local bound, or unknown risk without the reviewed exception on both sides; locally determined risk is not tested.

<a id="UT.internal.admission.TestGateIssuerLocalConflicts"></a>

### UT.internal.admission.TestGateIssuerLocalConflicts

- Case: `UT.internal.admission.TestGateIssuerLocalConflicts`
- Result: passed
- Test: `TestGateIssuerLocalConflicts` in `internal/admission/authority_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateIssuerLocalConflicts$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateIssuerLocalConflicts in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0010: Asserts that when issuer and deployment disagree the stricter attempt duration binds in both directions, disjoint path grants permit neither path, registration IDs intersect, and consent is required when the issuer or the task requires it, with effective paths narrowed to the task's path.

<a id="UT.internal.admission.TestGateLocallyDeterminedRiskBounds"></a>

### UT.internal.admission.TestGateLocallyDeterminedRiskBounds

- Case: `UT.internal.admission.TestGateLocallyDeterminedRiskBounds`
- Result: passed
- Test: `TestGateLocallyDeterminedRiskBounds` in `internal/admission/safety_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateLocallyDeterminedRiskBounds$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateLocallyDeterminedRiskBounds in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0010: Asserts that unknown locally determined risk from a matched registration is rejected unless the reviewed unknown-risk exception is granted by both issuer and deployment, and admitted when both grant it.
  - REQ-SAFETY-0023: Asserts that a matched registration's reviewed baseline raises effective risk on a low registered baseline and low declaration, and is carried as the effective risk when permitted.
  - REQ-SAFETY-0027: Asserts at admission that effective risk raised by a matched registration must be permitted by local policy: above the lower maximum or without medium-risk enablement it is rejected as task_issuer_not_authorized, narrowed in the issuer and then in the deployment. Consent and host or device conditions are not tested.
  - REQ-TRUST-0011: Asserts that the issuer maximum risk, medium-risk enablement and unknown-risk exception bound locally determined risk, each narrowed in the issuer with the deployment permissive.

<a id="UT.internal.admission.TestGateNoTaskControlledAuthorityOrInterpolation"></a>

### UT.internal.admission.TestGateNoTaskControlledAuthorityOrInterpolation

- Case: `UT.internal.admission.TestGateNoTaskControlledAuthorityOrInterpolation`
- Result: passed
- Test: `TestGateNoTaskControlledAuthorityOrInterpolation` in `internal/admission/admit_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateNoTaskControlledAuthorityOrInterpolation$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateNoTaskControlledAuthorityOrInterpolation in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0026: Asserts that recipient selectors (installation, linkage, bundle, run, account, advertising, device, push, IP, location, eligibility and selector fields, top-level and nested) are task_schema_invalid at admission.
  - REQ-TRUST-0027: Asserts that placeholder text for country, application version, cohort and network in input, options and labels is kept literal in the gated candidate and that identical bytes on two instances yield equal candidates. Host-side eligibility policy is not tested.
  - REQ-TRUST-0040: Asserts that top-level authenticated, issuer_id, policy_generation, trusted and source fields are task_schema_invalid, and that a label reading 'authenticated: true' leaves the gated source context embedded origin with its installed issuer.

<a id="UT.internal.admission.TestGatePolicyAndIssuerResolution"></a>

### UT.internal.admission.TestGatePolicyAndIssuerResolution

- Case: `UT.internal.admission.TestGatePolicyAndIssuerResolution`
- Result: passed
- Test: `TestGatePolicyAndIssuerResolution` in `internal/admission/authority_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGatePolicyAndIssuerResolution$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGatePolicyAndIssuerResolution in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0010: Asserts that issuer authority is resolved by exact issuer ID for embedded and verified remote sources, that an embedded issuer absent from the policy is task_issuer_not_authorized while a policy without the remote issuer holds none of its keys, so its package fails verification as task_source_unauthenticated, and that remote validity is bounded from the earlier of issuance and acceptance.
  - REQ-TRUST-0047: Asserts at step 7 that no installed policy and a source policy_generation above the active one are task_policy_unavailable, one below is policy_rollback, and an unknown issuer is task_issuer_not_authorized, with generation checked before the issuer and the codec not reached.

<a id="UT.internal.admission.TestGatePrecedence"></a>

### UT.internal.admission.TestGatePrecedence

- Case: `UT.internal.admission.TestGatePrecedence`
- Result: passed
- Test: `TestGatePrecedence` in `internal/admission/admit_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGatePrecedence$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGatePrecedence in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0047: Asserts first-failing-step precedence among implemented steps: unregistered version before a 129-byte task_id, unknown field before unregistered version or over-limit labels, task_id characters before a field limit, task_id length before expiry, expiry and unrecognized risk factor type before invalid input, and provenance (an unsigned task presented as remote), freshness (a package issued 301 seconds ahead) and trusted time before the byte limit, while a validly signed task over 65,536 bytes fails the byte limit after verification and one at the limit passes.

<a id="UT.internal.admission.TestGateProxyCompatibility"></a>

### UT.internal.admission.TestGateProxyCompatibility

- Case: `UT.internal.admission.TestGateProxyCompatibility`
- Result: passed
- Test: `TestGateProxyCompatibility` in `internal/admission/proxy_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateProxyCompatibility$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateProxyCompatibility in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0026: Asserts that a proxy task of a test whose registered rule is never compatible (`registry.NoProxy`, the form a test listed as proxy-incompatible would use) is rejected, and that a rule's per-parameter verdict decides acceptance. The per-test compatibility list of REQ-PROXY-0026 is not encoded by any production rule and is not tested.
  - REQ-PROXY-0027: Asserts that a proxy task whose synthetic rule reports its parameters incompatible is rejected atomically at admission, for embedded and remote origin, with no call or open connection on the unused recording transport; no compatible subset is admitted.
  - REQ-PROXY-0042: Asserts that an incompatible proxy task is rejected at admission as `task_schema_invalid` rather than returned as a gated task that could end in `proxy_path_unsupported`. The execution-time state is not tested.
  - REQ-TRUST-0018: Asserts that every compatibility rejection and every acceptance causes no call or open connection on the unused recording transport.
  - REQ-TRUST-0039: Asserts that a `proxy` task whose synthetic rule reports its parameters incompatible is `task_schema_invalid`, that a compatible one is accepted with its effective path narrowed to proxy, that a `direct` task is not checked, that a never-compatible rule (`registry.NoProxy`) rejects every proxy task and no direct task, and that the check runs after input and options and before issuer and local path authority, target registration and the step-9 risk bound. NET-05 profiles and `task_unsupported` for a permitted but unprovided combination are not tested.
  - REQ-TRUST-0047: Asserts that invalid input, invalid options, a codec failure, an egress requirement, a test not permitted by the issuer and policy rollback keep their classes ahead of the compatibility check and leave the rule unconsulted, and that a compatible task still receives `task_issuer_not_authorized` or `target_not_registered` from the later step-8 checks.

<a id="UT.internal.admission.TestGateRejectsCompositeProfiles"></a>

### UT.internal.admission.TestGateRejectsCompositeProfiles

- Case: `UT.internal.admission.TestGateRejectsCompositeProfiles`
- Result: passed
- Test: `TestGateRejectsCompositeProfiles` in `internal/admission/capability_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateRejectsCompositeProfiles$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateRejectsCompositeProfiles in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0018: Asserts that every NET-05 rejection, unregistered, without composite enablement or registered and enabled, causes no call or open connection on the unused recording transport.
  - REQ-TRUST-0036: Asserts that NET-05 absent from the registry, at the registered or another version, is task_unsupported even under a policy that enables composite tasks, and that a registered NET-05 composite enabled by both issuer and deployment is task_unsupported because composite steps cannot be matched, while SYN-01 is admitted under the same policy. Composite-step matching is not provided and is not tested.
  - REQ-TRUST-0047: Asserts that policy rollback and missing composite enablement in the issuer or in the deployment (task_issuer_not_authorized) precede the step-8 composite check, that invalid input and an egress requirement precede it, and that it precedes path authority, target registration and risk bounds.

<a id="UT.internal.admission.TestGateRejectsEgressContextRequirements"></a>

### UT.internal.admission.TestGateRejectsEgressContextRequirements

- Case: `UT.internal.admission.TestGateRejectsEgressContextRequirements`
- Result: passed
- Test: `TestGateRejectsEgressContextRequirements` in `internal/admission/capability_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateRejectsEgressContextRequirements$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateRejectsEgressContextRequirements in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0018: Asserts that rejecting a task with an egress context requirement, for embedded and remote origin, causes no call or open connection on the unused recording transport, so admission performs no context lookup, endpoint call or other network activity.
  - REQ-TRUST-0036: Asserts that a valid task requiring country_code, region_code or asn individually or all three, a country allowlist alone or with an empty required_fields, or required fields with an allowlist, is rejected as task_unsupported for embedded and remote origin; that tasks without a network object, without context, with an empty context or with an empty required_fields are admitted; and that an invalid context (unknown field name, duplicate field, empty allowlist, lowercase code or unknown member) is task_schema_invalid instead. Egress acquisition is not provided and is not tested.
  - REQ-TRUST-0047: Asserts the step-6 row for a valid field that requires a capability the implementation does not provide (task_unsupported), with the registry codec never reached.

<a id="UT.internal.admission.TestGateStepClasses"></a>

### UT.internal.admission.TestGateStepClasses

- Case: `UT.internal.admission.TestGateStepClasses`
- Result: passed
- Test: `TestGateStepClasses` in `internal/admission/admit_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateStepClasses$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateStepClasses in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0004: Asserts on original encoded bytes: complete task 65,535 and 65,536 accepted and 65,537 task_limit_exceeded; input 49,152 accepted and 49,153 rejected, including an escaped value whose decoded length is smaller; options 8,192 and 8,193; 16 and 17 labels; labels of 2,048 and 2,049 encoded bytes.
  - REQ-LIMIT-0005: Asserts that an unknown field taking a task to 65,537 bytes is task_limit_exceeded while the same field within the limit is task_schema_invalid, and that 65,537 bytes of malformed input are task_limit_exceeded before parsing.
  - REQ-LIMIT-0010: Asserts that remote issued_at up to 300 seconds after trusted time passes and 300.5 or 301 seconds is rejected, and that unavailable trusted time fails remote and embedded admission.
  - REQ-SAFETY-0017: Asserts that a task with an unrecognized risk factor type is rejected as task_unsupported rather than silently ignored, and that the same type with an invalid level is task_schema_invalid.
  - REQ-TRUST-0018: Every rejection and acceptance runs with an unused recording transport that must report zero calls and no open bodies; admission holds no transport and its import rule forbids network and file I/O.
  - REQ-TRUST-0036: Asserts task_unsupported, not task_schema_invalid, for a well-formed unregistered test_version or test_id and for an unrecognized risk factor type; a malformed test_id or test_version and an unsupported schema_version are task_schema_invalid.
  - REQ-TRUST-0043: Asserts task_source_unauthenticated for a missing, malformed, foreign-instance, byte-mismatched or expired delivery record and for a bare task presented as remote without a signed package, and task_trust_unavailable when trusted time is unavailable at resolution or at step 3.
  - REQ-TRUST-0046: Asserts the class of each implemented step (1 to 6 and the input/options part of 8) for tasks that each fail one step, and that the synthetic codec is not invoked when a step before 8 fails.
  - REQ-TRUST-0047: Asserts the REQ-TRUST-0047 classes for the implemented rows: step 1 binding failures and a remote package issued 301 seconds ahead as task_source_unauthenticated, unavailable trusted time as task_trust_unavailable, and unregistered pairs and unrecognized risk factor types as task_unsupported.

<a id="UT.internal.admission.TestGateTargetRegistrationAuthority"></a>

### UT.internal.admission.TestGateTargetRegistrationAuthority

- Case: `UT.internal.admission.TestGateTargetRegistrationAuthority`
- Result: passed
- Test: `TestGateTargetRegistrationAuthority` in `internal/admission/target_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTargetRegistrationAuthority$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTargetRegistrationAuthority in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0011: Asserts that a registration, including a redirect registration, omitted from the issuer's or the deployment's registration list is target_not_registered, while one omitted only by another issuer still matches.
  - REQ-TRUST-0013: Asserts target_not_registered for a registration of another test or version, of the validation or resolver role instead of measurement, and for a measurement tuple registered only with the redirect role.
  - REQ-TRUST-0014: Asserts that the matched registration's direct or proxy path permission binds, including for a redirect registration, that a redirect not listed by the matched registration is rejected, and that a proxy-only duplicate tuple does not hide a direct registration.

<a id="UT.internal.admission.TestGateTargetRegistrationExactMatch"></a>

### UT.internal.admission.TestGateTargetRegistrationExactMatch

- Case: `UT.internal.admission.TestGateTargetRegistrationExactMatch`
- Result: passed
- Test: `TestGateTargetRegistrationExactMatch` in `internal/admission/target_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTargetRegistrationExactMatch$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTargetRegistrationExactMatch in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0013: Asserts that synthetic TCP-01, TLS-01, HTTP-01 and DNS-01 tasks whose every reported tuple exactly matches a current, permitted registration pass step 8, and that the carried registration IDs narrow to the matched measurement and redirect registrations; acceptance (steps 10 and 11) is not tested.
  - REQ-TRUST-0014: Asserts matching on address, port and scheme (TCP-01), plus TLS server name and ALPN values (TLS-01), URL, expected status, body and header checks and a permitted redirect target (HTTP-01), and hostname, resolver name and bootstrap addresses (DNS-01), using test-scoped synthetic codecs and rules, not the public test schemas.

<a id="UT.internal.admission.TestGateTargetRegistrationExpiry"></a>

### UT.internal.admission.TestGateTargetRegistrationExpiry

- Case: `UT.internal.admission.TestGateTargetRegistrationExpiry`
- Result: passed
- Test: `TestGateTargetRegistrationExpiry` in `internal/admission/target_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTargetRegistrationExpiry$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTargetRegistrationExpiry in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0013: Asserts that a registration expiring one second after trusted now matches and one expiring at or one second before trusted now is target_not_registered, for each synthetic test, with the deterministic clock.
  - REQ-TRUST-0014: Asserts that an expired redirect registration fails an HTTP-01 task whose measurement registration is current.

<a id="UT.internal.admission.TestGateTargetRegistrationNormalization"></a>

### UT.internal.admission.TestGateTargetRegistrationNormalization

- Case: `UT.internal.admission.TestGateTargetRegistrationNormalization`
- Result: passed
- Test: `TestGateTargetRegistrationNormalization` in `internal/admission/target_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTargetRegistrationNormalization$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTargetRegistrationNormalization in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0015: Asserts that forms the synthetic codec normalizes to a registered value (case of names, URL scheme and host, noncanonical IPv6 text) match, while suffix (evil-example.test), subdomain, parent-domain, trailing-dot, wildcard, URL path-prefix and extra-path, URL host-suffix, URL path-case and IPv4-mapped IPv6 forms are target_not_registered; a malformed address is task_schema_invalid at the codec. Public normalization rules for each registered test are not tested.

<a id="UT.internal.admission.TestGateTargetRegistrationPrecedence"></a>

### UT.internal.admission.TestGateTargetRegistrationPrecedence

- Case: `UT.internal.admission.TestGateTargetRegistrationPrecedence`
- Result: passed
- Test: `TestGateTargetRegistrationPrecedence` in `internal/admission/target_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTargetRegistrationPrecedence$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTargetRegistrationPrecedence in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0047: Asserts the step-8 target-registration row and first-failing-step precedence around it: an expired task with an unregistered target is task_expired, an unknown field task_schema_invalid, an issuer path refusal task_issuer_not_authorized, and an unregistered target with risk above the issuer maximum target_not_registered.

<a id="UT.internal.admission.TestGateTargetRegistrationRequiresCompleteTuples"></a>

### UT.internal.admission.TestGateTargetRegistrationRequiresCompleteTuples

- Case: `UT.internal.admission.TestGateTargetRegistrationRequiresCompleteTuples`
- Result: passed
- Test: `TestGateTargetRegistrationRequiresCompleteTuples` in `internal/admission/target_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTargetRegistrationRequiresCompleteTuples$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTargetRegistrationRequiresCompleteTuples in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0013: Asserts target_not_registered when the test's rule reports no target, only a redirect or only a resolver, parameters of an unexpected type, or more than the local bound of 64 tuples, and that 64 tuples are accepted.
  - REQ-TRUST-0014: Asserts target_not_registered when a synthetic rule reports an incomplete tuple: TCP-01 without address, port or scheme, TLS-01 without server name or ALPN, HTTP-01 without URL or status, DNS-01 without resolver or bootstrap addresses, or a redirect on a non-HTTP test.

<a id="UT.internal.admission.TestGateTargetRegistrationSingleComponentMutations"></a>

### UT.internal.admission.TestGateTargetRegistrationSingleComponentMutations

- Case: `UT.internal.admission.TestGateTargetRegistrationSingleComponentMutations`
- Result: passed
- Test: `TestGateTargetRegistrationSingleComponentMutations` in `internal/admission/target_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTargetRegistrationSingleComponentMutations$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTargetRegistrationSingleComponentMutations in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0014: Changes one component at a time (IP address and family, port, scheme, added hostname, server name, each ALPN change including order, URL host, path, port and query, expected status, body and header checks, redirect target, an extra redirect, queried name, resolver name, bootstrap address including order, and direct versus proxy path) and asserts target_not_registered for each.
  - REQ-TRUST-0015: Asserts that matching compares the complete normalized typed tuple, so no single-component variation of a registered tuple is authorized.

<a id="UT.internal.admission.TestGateTaskCannotLowerBaseline"></a>

### UT.internal.admission.TestGateTaskCannotLowerBaseline

- Case: `UT.internal.admission.TestGateTaskCannotLowerBaseline`
- Result: passed
- Test: `TestGateTaskCannotLowerBaseline` in `internal/admission/safety_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTaskCannotLowerBaseline$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTaskCannotLowerBaseline in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0010: Asserts that an unknown registered baseline with a low declaration fails closed as task_issuer_not_authorized when the unknown-risk exception is granted on one side only, and is admitted as unknown only with both.
  - REQ-SAFETY-0024: Asserts that a low declaration cannot lower a medium baseline below an issuer or deployment maximum of low, and cannot lower an unknown baseline.
  - REQ-SAFETY-0025: Asserts that a low declaration on a medium baseline is evaluated as medium and rejected as task_issuer_not_authorized under a low maximum on either side.

<a id="UT.internal.admission.TestGateTaskIdentifierRules"></a>

### UT.internal.admission.TestGateTaskIdentifierRules

- Case: `UT.internal.admission.TestGateTaskIdentifierRules`
- Result: passed
- Test: `TestGateTaskIdentifierRules` in `internal/admission/admit_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateTaskIdentifierRules$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateTaskIdentifierRules in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0004: Asserts 128-byte task_id values pass and 129-byte values are task_limit_exceeded, never task_schema_invalid, including 129 uppercase bytes.
  - REQ-TRUST-0019: Asserts task_id values of 1, 127 and 128 allowed bytes pass; empty, uppercase, space, '/', non-ASCII, leading '.', '_' or '-', placeholder syntax and a numeric task_id are task_schema_invalid. Campaign scope and non-user-specificity are not tested.

<a id="UT.internal.admission.TestGateUnreliableSafetyInputsFailClosed"></a>

### UT.internal.admission.TestGateUnreliableSafetyInputsFailClosed

- Case: `UT.internal.admission.TestGateUnreliableSafetyInputsFailClosed`
- Result: passed
- Test: `TestGateUnreliableSafetyInputsFailClosed` in `internal/admission/safety_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateUnreliableSafetyInputsFailClosed$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateUnreliableSafetyInputsFailClosed in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0017: Asserts that unrecognized factor types (battery_drain, bandwidth, a case variant, a hyphen variant and the empty string) are task_unsupported at admission even when both sides grant the unknown-risk exception, never accepted with the factor ignored.
  - REQ-SAFETY-0019: Asserts that an unknown factor under overall low, medium or high is rejected as task_schema_invalid, because no trusted local assessment exists to resolve it. Resolution through a trusted assessment is not implemented or tested.
  - REQ-SAFETY-0029: Asserts that declarations whose risk cannot be evaluated reliably fail closed before effective-risk evaluation.
  - REQ-TRUST-0047: Asserts that an unregistered target at step 8 returns target_not_registered before an unknown-risk rejection at step 9.

<a id="UT.internal.admission.TestPackageMemberFormats"></a>

### UT.internal.admission.TestPackageMemberFormats

- Case: `UT.internal.admission.TestPackageMemberFormats`
- Result: passed
- Test: `TestPackageMemberFormats` in `internal/admission/signed_internal_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestPackageMemberFormats$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestPackageMemberFormats in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0048: Asserts that `issued_at` accepts exactly `YYYY-MM-DDTHH:MM:SSZ` for valid dates from 0001 to 9999, including a leap day, and rejects a missing `Z`, lowercase `t` or `z`, a space separator, fractions including `.0`, numeric offsets, second 60, invalid days and months, year 0000, hour 24, signs, leading spaces and short fields; that identifiers accept 1 to 64 lowercase letters, digits, `.`, `_` and `-` and reject empty, uppercase, space, `/`, non-ASCII, 65-byte and NUL-containing values; and that base64url accepts only the canonical unpadded alphabet, rejecting padding, whitespace, the standard alphabet, nonzero trailing bits and impossible lengths.

<a id="UT.internal.admission.TestRemoteRecordsAreSingleUse"></a>

### UT.internal.admission.TestRemoteRecordsAreSingleUse

- Case: `UT.internal.admission.TestRemoteRecordsAreSingleUse`
- Result: passed
- Test: `TestRemoteRecordsAreSingleUse` in `internal/admission/delivery_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestRemoteRecordsAreSingleUse$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestRemoteRecordsAreSingleUse in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0053: Asserts that a remote delivery record resolves once and is then gone, that a resolution with mismatched bytes or without trusted time also drops it, and that each retry through a newly installed reference resolves.

<a id="UT.internal.admission.TestResolveChecksBindingBeforeTrustedTime"></a>

### UT.internal.admission.TestResolveChecksBindingBeforeTrustedTime

- Case: `UT.internal.admission.TestResolveChecksBindingBeforeTrustedTime`
- Result: passed
- Test: `TestResolveChecksBindingBeforeTrustedTime` in `internal/admission/delivery_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestResolveChecksBindingBeforeTrustedTime$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestResolveChecksBindingBeforeTrustedTime in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0047: Asserts that with the trusted clock unavailable, an unknown well-formed reference, a reference from another store instance, an embedded or remote record presented with bytes differing by one byte, a record installed under a replaced policy generation, an invalidated record and a record of a closed store are each `task_source_unauthenticated`, the step-1 class, rather than the trusted-time class; that a held embedded or remote record presented with its exact bytes is `task_trust_unavailable` while the clock is unavailable; and that a held record past its lifetime is `task_trust_unavailable` while the clock is unavailable and `task_source_unauthenticated` once it reads.
  - REQ-TRUST-0053: Asserts that a remote record presented with mismatched bytes or without trusted time is dropped on that first resolution and does not resolve again, while an embedded record survives both.

<a id="UT.internal.admission.TestRestrictiveFailsClosed"></a>

### UT.internal.admission.TestRestrictiveFailsClosed

- Case: `UT.internal.admission.TestRestrictiveFailsClosed`
- Result: passed
- Test: `TestRestrictiveFailsClosed` in `internal/admission/risk_internal_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestRestrictiveFailsClosed$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestRestrictiveFailsClosed in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0029: Asserts that the effective-risk combination treats unknown and out-of-range levels, including the zero value, as unknown; that unrecognized declared strings map to unknown; and that a matched registration missing from the policy snapshot is evaluated as unknown and rejected without the unknown-risk exception.

<a id="UT.internal.admission.TestSignatureStrictnessEdgeVectors"></a>

### UT.internal.admission.TestSignatureStrictnessEdgeVectors

- Case: `UT.internal.admission.TestSignatureStrictnessEdgeVectors`
- Result: passed
- Test: `TestSignatureStrictnessEdgeVectors` in `internal/admission/signed_internal_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestSignatureStrictnessEdgeVectors$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestSignatureStrictnessEdgeVectors in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0050: Asserts, with signatures built from synthetic keys so that the verification equation holds for the decoded values, that an identity R in canonical encoding is accepted while the identity encoded with y = p + 1 or with the sign bit set is rejected, that S + L, S + 2L, S = L and S with its top bits set are rejected, and that flipping any one of the 64 signature bytes fails verification. This checks Go 1.26 `crypto/ed25519` behavior; it is not a cryptographic review.

<a id="UT.internal.admission.TestTargetRegistrationRejectsCredentialsAndInterpolation"></a>

### UT.internal.admission.TestTargetRegistrationRejectsCredentialsAndInterpolation

- Case: `UT.internal.admission.TestTargetRegistrationRejectsCredentialsAndInterpolation`
- Result: passed
- Test: `TestTargetRegistrationRejectsCredentialsAndInterpolation` in `internal/admission/target_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestTargetRegistrationRejectsCredentialsAndInterpolation$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestTargetRegistrationRejectsCredentialsAndInterpolation in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0015: Asserts that local policy construction refuses registrations with URL credentials or user info, interpolation syntax in URL, host or ALPN, wildcard or suffix-rule hosts and IPv4-mapped addresses, so no task can match them.

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

<a id="UT.internal.execution.TestAccountingAnchor"></a>

### UT.internal.execution.TestAccountingAnchor

- Case: `UT.internal.execution.TestAccountingAnchor`
- Result: passed
- Test: `TestAccountingAnchor` in `internal/execution/occurrence_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestAccountingAnchor$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestAccountingAnchor in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0007: Asserts through `NewAccounting` that occurrence 0 of a task with `starts_at` is anchored at `starts_at`, and that a fractional `starts_at` is rounded up to the next whole second, so no run becomes due before it.
  - REQ-SCHEDULE-0009: Asserts that without `starts_at` the anchor and the first due time are the acceptance time, with no reserved run.

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

<a id="UT.internal.execution.TestAttemptDeadlineBeforeTheTestStarts"></a>

### UT.internal.execution.TestAttemptDeadlineBeforeTheTestStarts

- Case: `UT.internal.execution.TestAttemptDeadlineBeforeTheTestStarts`
- Result: passed
- Test: `TestAttemptDeadlineBeforeTheTestStarts` in `internal/execution/deadline_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestAttemptDeadlineBeforeTheTestStarts$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestAttemptDeadlineBeforeTheTestStarts in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0006: Asserts that the attempt deadline covers Network Context capture: expiry while capture is outstanding ends the attempt with `attempt_deadline_exceeded`.
  - REQ-RESULT-0051: Asserts that deadline expiry during context capture and in the runner before its first network operation yields `not_run`, not `incomplete`.
  - REQ-RESULT-0053: Asserts that both `not_run` results carry no `network_context`, including the one whose context capture had succeeded.

<a id="UT.internal.execution.TestAttemptDeadlineBoundaries"></a>

### UT.internal.execution.TestAttemptDeadlineBoundaries

- Case: `UT.internal.execution.TestAttemptDeadlineBoundaries`
- Result: passed
- Test: `TestAttemptDeadlineBoundaries` in `internal/execution/deadline_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestAttemptDeadlineBoundaries$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestAttemptDeadlineBoundaries in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0006: Asserts on the injected engine clock with a 30,000 ms `max_duration_ms` that the attempt context is still live after 29,999 ms and has ended at 30,000 ms and at 30,001 ms; an expired attempt is `incomplete` with `attempt_deadline_exceeded`, an attempt finishing before the deadline is `completed`, and no deadline timer stays armed after the run.
  - REQ-RESULT-0051: Asserts that an attempt whose runner started the test and is stopped by the deadline is `incomplete`.

<a id="UT.internal.execution.TestAttemptScopeReleasesSockets"></a>

### UT.internal.execution.TestAttemptScopeReleasesSockets

- Case: `UT.internal.execution.TestAttemptScopeReleasesSockets`
- Result: passed
- Test: `TestAttemptScopeReleasesSockets` in `internal/execution/deadline_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestAttemptScopeReleasesSockets$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestAttemptScopeReleasesSockets in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0008: Asserts that a loopback socket the runner leaves open is closed when the runner returns, as observed by the server, and that a dial after the attempt was cancelled is refused without reaching the path's dialer.

<a id="UT.internal.execution.TestBudgetAcrossConnectionReuse"></a>

### UT.internal.execution.TestBudgetAcrossConnectionReuse

- Case: `UT.internal.execution.TestBudgetAcrossConnectionReuse`
- Result: passed
- Test: `TestBudgetAcrossConnectionReuse` in `internal/execution/budget_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestBudgetAcrossConnectionReuse$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestBudgetAcrossConnectionReuse in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0034: Asserts that four request/reply exchanges on one reused loopback connection accumulate 400 sent and 200 received metered bytes on one connection, and that with `max_bytes_sent` 399 the fourth request is not issued, leaving 300 bytes read by the server and the attempt stopped with `resource_limit`.

<a id="UT.internal.execution.TestCancellationAtEveryStage"></a>

### UT.internal.execution.TestCancellationAtEveryStage

- Case: `UT.internal.execution.TestCancellationAtEveryStage`
- Result: passed
- Test: `TestCancellationAtEveryStage` in `internal/execution/deadline_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCancellationAtEveryStage$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestCancellationAtEveryStage in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0051: Asserts that host cancellation during context capture and in the runner before its first network operation yields `not_run` with `cancelled_by_host`, while cancellation during a dial and during a blocked read on a real loopback socket, and deadline expiry during that read, yield `incomplete`; the loopback server observes the socket closed after each real-socket stop.

<a id="UT.internal.execution.TestCancellationRacesWithCompletion"></a>

### UT.internal.execution.TestCancellationRacesWithCompletion

- Case: `UT.internal.execution.TestCancellationRacesWithCompletion`
- Result: passed
- Test: `TestCancellationRacesWithCompletion` in `internal/execution/deadline_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCancellationRacesWithCompletion$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestCancellationRacesWithCompletion in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0052: Asserts over 200 runs that race host cancellation with a runner returning an outcome that each run writes exactly one result, that a recorded outcome stays `completed` with its measurement, and that the reported attempt phase matches the written state.

<a id="UT.internal.execution.TestConnectionBudgetBoundaries"></a>

### UT.internal.execution.TestConnectionBudgetBoundaries

- Case: `UT.internal.execution.TestConnectionBudgetBoundaries`
- Result: passed
- Test: `TestConnectionBudgetBoundaries` in `internal/execution/budget_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestConnectionBudgetBoundaries$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestConnectionBudgetBoundaries in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0006: Asserts with `max_connections` 3 that two and three sequential loopback connections complete and that a fourth dial is not issued: the server accepts three, the path dialer dials three and the attempt is `incomplete` with `resource_limit`.
  - REQ-LIMIT-0027: Asserts that the dial beyond the connection limit stops the attempt with `resource_limit`.

<a id="UT.internal.execution.TestConsentGate"></a>

### UT.internal.execution.TestConsentGate

- Case: `UT.internal.execution.TestConsentGate`
- Result: passed
- Test: `TestConsentGate` in `internal/execution/gates_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestConsentGate$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestConsentGate in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0020: Asserts that for a task whose consent requirement is recorded only valid consent passes the start check: absent, ended and unavailable statuses, a failing provider and a missing provider are unsatisfied; each check queries the provider once, a task without the requirement does not query it, and a cancelled context is unsatisfied.

<a id="UT.internal.execution.TestDeviceGate"></a>

### UT.internal.execution.TestDeviceGate

- Case: `UT.internal.execution.TestDeviceGate`
- Result: passed
- Test: `TestDeviceGate` in `internal/execution/gates_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestDeviceGate$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestDeviceGate in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0028: Asserts the battery condition at limit −1 (unsatisfied), limit and limit +1 (satisfied) for `minimum_battery_percent` 20, that an unknown battery level, an unknown or false charging state for `allow_charging_only`, a failing capability and a missing capability are unsatisfied, and that device state is not queried without a requirement.

<a id="UT.internal.execution.TestEachAttemptHasItsOwnBudget"></a>

### UT.internal.execution.TestEachAttemptHasItsOwnBudget

- Case: `UT.internal.execution.TestEachAttemptHasItsOwnBudget`
- Result: passed
- Test: `TestEachAttemptHasItsOwnBudget` in `internal/execution/budget_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestEachAttemptHasItsOwnBudget$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestEachAttemptHasItsOwnBudget in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0006: Asserts that a retried attempt after `internal_error` gets the full byte budget again, each attempt metering 600 of 1,000 sent bytes and the server reading 1,200, and that an attempt stopped by `resource_limit`, which is not a retry candidate, requests no second dialer.

<a id="UT.internal.execution.TestExhaustedBudgetDispatchesNoFurtherIO"></a>

### UT.internal.execution.TestExhaustedBudgetDispatchesNoFurtherIO

- Case: `UT.internal.execution.TestExhaustedBudgetDispatchesNoFurtherIO`
- Result: passed
- Test: `TestExhaustedBudgetDispatchesNoFurtherIO` in `internal/execution/budget_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestExhaustedBudgetDispatchesNoFurtherIO$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestExhaustedBudgetDispatchesNoFurtherIO in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0027: Asserts that after the sent budget is exhausted on one of two open loopback connections, a later write and read on the other connection and a later dial all fail, and that an outcome the runner returns afterwards is recorded as `incomplete` with `resource_limit`, not `completed`.
  - REQ-LIMIT-0028: Asserts that the server reads and writes no byte, observes both sockets closed, and that the path dialer is neither called again nor replaced by another dialer.

<a id="UT.internal.execution.TestExhaustionReleasesBlockedSockets"></a>

### UT.internal.execution.TestExhaustionReleasesBlockedSockets

- Case: `UT.internal.execution.TestExhaustionReleasesBlockedSockets`
- Result: passed
- Test: `TestExhaustionReleasesBlockedSockets` in `internal/execution/budget_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestExhaustionReleasesBlockedSockets$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestExhaustionReleasesBlockedSockets in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0008: Asserts that budget exhaustion on one connection interrupts a read blocked on another real loopback socket and that the attempt ends with no armed timer, so cleanup does not wait for the blocked operation.
  - REQ-LIMIT-0028: Asserts that the holding server observes every socket closed after the attempt stops with `resource_limit`.

<a id="UT.internal.execution.TestFailedConnectCountsTowardConnections"></a>

### UT.internal.execution.TestFailedConnectCountsTowardConnections

- Case: `UT.internal.execution.TestFailedConnectCountsTowardConnections`
- Result: passed
- Test: `TestFailedConnectCountsTowardConnections` in `internal/execution/budget_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestFailedConnectCountsTowardConnections$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestFailedConnectCountsTowardConnections in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0006: Asserts that a refused loopback connect counts as one connection, recorded as a completed `socket_error` outcome with one metered connection, and that with `max_connections` 1 a second dial after the refused one is not issued and the attempt stops with `resource_limit`.

<a id="UT.internal.execution.TestFirstTerminationIsKept"></a>

### UT.internal.execution.TestFirstTerminationIsKept

- Case: `UT.internal.execution.TestFirstTerminationIsKept`
- Result: passed
- Test: `TestFirstTerminationIsKept` in `internal/execution/deadline_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestFirstTerminationIsKept$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestFirstTerminationIsKept in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0021: Asserts that only the first outer condition is recorded: a deadline that expired before a host cancellation yields `attempt_deadline_exceeded`, and a host cancellation before the deadline yields `cancelled_by_host`.

<a id="UT.internal.execution.TestHostCancellation"></a>

### UT.internal.execution.TestHostCancellation

- Case: `UT.internal.execution.TestHostCancellation`
- Result: passed
- Test: `TestHostCancellation` in `internal/execution/run_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestHostCancellation$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestHostCancellation in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0051: Asserts that a host cancellation after the test started yields `incomplete` and one before it yields `not_run`.
  - REQ-RESULT-0052: Asserts that a cancellation after the runner returned its outcome leaves the attempt `completed` with that outcome and starts no retry.
  - REQ-SCHEDULE-0027: Asserts that an attempt stopped by host cancellation is recorded with `cancelled_by_host` and is never retried, also when the cancellation arrives during the backoff wait, and that a run cancelled before its first attempt records only that attempt, as `not_run`, and does no other work.

<a id="UT.internal.execution.TestIntervalDueTimes"></a>

### UT.internal.execution.TestIntervalDueTimes

- Case: `UT.internal.execution.TestIntervalDueTimes`
- Result: passed
- Test: `TestIntervalDueTimes` in `internal/execution/occurrence_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestIntervalDueTimes$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestIntervalDueTimes in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0020: Asserts under the interim due-time rule of GO-SPEC-ISSUE-011 that occurrence 0 is due at the anchor, not one second before, and that after its reservation occurrence 1 is due at the anchor plus the 21,600-second interval plus jitter: one second before that it is not due, at and after it it is.
  - REQ-SCHEDULE-0021: Asserts that the jitter of a due run comes from the injected local jitter source and is capped at the task's `jitter`: a source offering one hour yields 300 seconds.

<a id="UT.internal.execution.TestMissedOccurrencesAreSkipped"></a>

### UT.internal.execution.TestMissedOccurrencesAreSkipped

- Case: `UT.internal.execution.TestMissedOccurrencesAreSkipped`
- Result: passed
- Test: `TestMissedOccurrencesAreSkipped` in `internal/execution/occurrence_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestMissedOccurrencesAreSkipped$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestMissedOccurrencesAreSkipped in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0020: Asserts under the interim rule of GO-SPEC-ISSUE-011 that when occurrence 1 is missed, occurrence 2 becomes the due run without consuming a run for the skipped one, and that after the last occurrence of `max_runs` 4 no run is due.
  - REQ-SCHEDULE-0021: Asserts that the jitter of a due run is drawn once: advancing within the current occurrence with a different jitter source leaves its due time unchanged.

<a id="UT.internal.execution.TestNetworkConditionsFailClosed"></a>

### UT.internal.execution.TestNetworkConditionsFailClosed

- Case: `UT.internal.execution.TestNetworkConditionsFailClosed`
- Result: passed
- Test: `TestNetworkConditionsFailClosed` in `internal/execution/gates_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestNetworkConditionsFailClosed$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestNetworkConditionsFailClosed in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0028: Asserts through `execution.Gates` that a network condition this build cannot determine is unsatisfied: a proxy or unset path, an allowed-transport list, a `vpn_policy` other than `any`, `require_validated`, required context fields and a country allowlist each fail the start check, while a direct path without conditions passes.

<a id="UT.internal.execution.TestPartialTransfersAreCountedExactly"></a>

### UT.internal.execution.TestPartialTransfersAreCountedExactly

- Case: `UT.internal.execution.TestPartialTransfersAreCountedExactly`
- Result: passed
- Test: `TestPartialTransfersAreCountedExactly` in `internal/execution/budget_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestPartialTransfersAreCountedExactly$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestPartialTransfersAreCountedExactly in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0034: Asserts on a scripted connection that a write accepting 3 of 8 bytes and a write accepting 3 of 7 bytes meter 6 sent bytes, and that two reads with 100-byte buffers request 10 and then 6 bytes from the stream under a 10-byte received budget and meter the 8 bytes returned. The scripted connection complements the loopback cases and is not runtime evidence on its own.

<a id="UT.internal.execution.TestPlanLimitsAreBounded"></a>

### UT.internal.execution.TestPlanLimitsAreBounded

- Case: `UT.internal.execution.TestPlanLimitsAreBounded`
- Result: passed
- Test: `TestPlanLimitsAreBounded` in `internal/execution/budget_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestPlanLimitsAreBounded$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestPlanLimitsAreBounded in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0006: Asserts that a plan whose effective `max_bytes_sent`, `max_bytes_received` or `max_connections` is one above its registered maximum or negative fails with `ErrPlan` before the runner is called, and that a plan at every registered maximum runs; no value is clamped.

<a id="UT.internal.execution.TestPreconditionsAreRecheckedBeforeEachAttempt"></a>

### UT.internal.execution.TestPreconditionsAreRecheckedBeforeEachAttempt

- Case: `UT.internal.execution.TestPreconditionsAreRecheckedBeforeEachAttempt`
- Result: passed
- Test: `TestPreconditionsAreRecheckedBeforeEachAttempt` in `internal/execution/run_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestPreconditionsAreRecheckedBeforeEachAttempt$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestPreconditionsAreRecheckedBeforeEachAttempt in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0028: Asserts through `Engine.Run` that a run whose start conditions are unsatisfied before the first attempt fails closed with no result, no context capture and no runner call, and that conditions unsatisfied before the retry end the run with only the first attempt's result.

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

<a id="UT.internal.execution.TestReadAheadIsCountedAndCapped"></a>

### UT.internal.execution.TestReadAheadIsCountedAndCapped

- Case: `UT.internal.execution.TestReadAheadIsCountedAndCapped`
- Result: passed
- Test: `TestReadAheadIsCountedAndCapped` in `internal/execution/budget_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestReadAheadIsCountedAndCapped$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestReadAheadIsCountedAndCapped in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0034: Asserts that a runner reading a 10-byte reply through a 4,096-byte read-ahead buffer from a server sending 500 bytes has its reads capped by the 300-byte received budget rather than its buffer, and that the metered count is between the 10 bytes it consumed and the limit and at most what the server wrote.
  - REQ-EXECUTION-0032: Asserts that bytes read ahead into a buffer are counted at the stream boundary when read, not when consumed.

<a id="UT.internal.execution.TestReceivedBudgetBoundaries"></a>

### UT.internal.execution.TestReceivedBudgetBoundaries

- Case: `UT.internal.execution.TestReceivedBudgetBoundaries`
- Result: passed
- Test: `TestReceivedBudgetBoundaries` in `internal/execution/budget_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestReceivedBudgetBoundaries$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestReceivedBudgetBoundaries in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0006: Asserts on real loopback sockets with `max_bytes_received` 1,000 that replies of 999 and 1,000 bytes complete with that many bytes metered, and that a 1,001-byte reply stops the attempt as `incomplete` with `resource_limit` after exactly 1,000 metered bytes.
  - REQ-LIMIT-0027: Asserts that the read with no received budget left stops the active operation and the attempt with `resource_limit`.
  - REQ-LIMIT-0034: Asserts that each read requests no more than the remaining received budget, so the runner can consume exactly the limit, and that the metered count never exceeds the bytes the server wrote.
  - REQ-EXECUTION-0032: Asserts that reads stop at the counted stream boundary once the received budget is used.

<a id="UT.internal.execution.TestRetryDelayIsCappedAndBoundedByExpiry"></a>

### UT.internal.execution.TestRetryDelayIsCappedAndBoundedByExpiry

- Case: `UT.internal.execution.TestRetryDelayIsCappedAndBoundedByExpiry`
- Result: passed
- Test: `TestRetryDelayIsCappedAndBoundedByExpiry` in `internal/execution/run_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestRetryDelayIsCappedAndBoundedByExpiry$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestRetryDelayIsCappedAndBoundedByExpiry in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0028: Asserts that `fixed` backoff with two attempts produces one bounded delay of `initial_delay_ms` plus jitter, capped at `max_delay_ms`.
  - REQ-SCHEDULE-0029: Asserts that a realized delay of 30,000 ms plus 60,000 ms jitter is capped at `max_delay_ms` 60,000 ms, and that a retry whose realized delay would reach `not_after` is not made and no wait occurs.
  - REQ-LIMIT-0011: Asserts that the realized retry delay stays within the registered retry maximum delay for the task (`max_delay_ms`), including when jitter would exceed it.

<a id="UT.internal.execution.TestRetryRechecksEligibilityAfterWait"></a>

### UT.internal.execution.TestRetryRechecksEligibilityAfterWait

- Case: `UT.internal.execution.TestRetryRechecksEligibilityAfterWait`
- Result: passed
- Test: `TestRetryRechecksEligibilityAfterWait` in `internal/execution/run_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestRetryRechecksEligibilityAfterWait$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestRetryRechecksEligibilityAfterWait in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0022: Asserts that a retry candidate is not permission to execute: when the task's time window closes during the backoff wait, eligibility is checked again after the wait and no second attempt runs.

<a id="UT.internal.execution.TestRunBudget"></a>

### UT.internal.execution.TestRunBudget

- Case: `UT.internal.execution.TestRunBudget`
- Result: passed
- Test: `TestRunBudget` in `internal/execution/occurrence_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestRunBudget$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestRunBudget in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0030: Asserts that run accounting reserves exactly `max_runs` runs for limits 1, 2 and 28 when offered one more due occurrence than the limit, that repeat mode `none` permits one run, and that a schedule without `max_runs` or an invalid accounting record permits none.

<a id="UT.internal.execution.TestRunFailsClosedBeforeAnyWork"></a>

### UT.internal.execution.TestRunFailsClosedBeforeAnyWork

- Case: `UT.internal.execution.TestRunFailsClosedBeforeAnyWork`
- Result: passed
- Test: `TestRunFailsClosedBeforeAnyWork` in `internal/execution/run_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestRunFailsClosedBeforeAnyWork$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestRunFailsClosedBeforeAnyWork in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0023: Asserts that a plan with `max_attempts` 0 or 3 is rejected with `ErrPlan` before any context capture, dialer, runner call or result write.
  - REQ-LIMIT-0011: Asserts that the engine enforces the registered bound of at most two attempts per run, including the initial attempt, by rejecting a larger `max_attempts` before any work.

<a id="UT.internal.execution.TestRunSequencesAttemptsAndRetries"></a>

### UT.internal.execution.TestRunSequencesAttemptsAndRetries

- Case: `UT.internal.execution.TestRunSequencesAttemptsAndRetries`
- Result: passed
- Test: `TestRunSequencesAttemptsAndRetries` in `internal/execution/run_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestRunSequencesAttemptsAndRetries$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestRunSequencesAttemptsAndRetries in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0022: Asserts through `Engine.Run` that additional attempts occur only for an execution failure the retry object names: a runner failure with `max_attempts` 2 and `retry_on` `internal_error` gets exactly one more attempt, while the same failure with `max_attempts` 1 or with `internal_error` outside `retry_on` gets none, and a completed attempt is never followed by another.
  - REQ-SCHEDULE-0023: Asserts that a run with `max_attempts` 2 makes at most two attempts including the initial one: two consecutive runner failures produce exactly two attempt results numbered 1 and 2, and the runner, dialer source and report each record exactly that many attempts.
  - REQ-SCHEDULE-0024: Asserts that the retry belongs to the same run: both attempt results carry the run ID of the run report.
  - REQ-SCHEDULE-0025: Asserts that a negative protocol outcome is a completed measurement, not an execution failure: a `tcp_connect` failure with outcome `connection_refused` is recorded as `completed` with its measurement and is not retried although `retry_on` contains `internal_error`.
  - REQ-SCHEDULE-0029: Asserts on the injected clock that the single wait before the retry equals `initial_delay_ms` (30,000 ms) plus the locally generated jitter (5,000 ms within `jitter_ms` 10,000), with no sleep.
  - REQ-LIMIT-0012: Asserts that only an execution reason triggers a retry: runner failures recorded as `incomplete` with `internal_error` retry, while completed attempts with positive or negative protocol outcomes are not retried.

<a id="UT.internal.execution.TestSentBudgetBoundaries"></a>

### UT.internal.execution.TestSentBudgetBoundaries

- Case: `UT.internal.execution.TestSentBudgetBoundaries`
- Result: passed
- Test: `TestSentBudgetBoundaries` in `internal/execution/budget_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestSentBudgetBoundaries$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestSentBudgetBoundaries in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0006: Asserts on real loopback sockets with `max_bytes_sent` 1,000 that one write of 999 and of 1,000 bytes completes with that many bytes metered, and that a write of 1,001 bytes stops the attempt as `incomplete` with `resource_limit`.
  - REQ-LIMIT-0027: Asserts that the write exceeding the remaining sent budget stops the active operation and the attempt with the common execution reason `resource_limit`, with no armed timer left.
  - REQ-LIMIT-0034: Asserts that the write that would exceed the remaining sent budget is not issued: the exchange server reads 0 bytes at limit +1, and the metered sent count equals the bytes the server read below and at the limit.
  - REQ-EXECUTION-0032: Asserts that the write is stopped at the counted stream boundary before any byte reaches the socket.

<a id="UT.internal.execution.TestZeroLimitsPermitNoTraffic"></a>

### UT.internal.execution.TestZeroLimitsPermitNoTraffic

- Case: `UT.internal.execution.TestZeroLimitsPermitNoTraffic`
- Result: passed
- Test: `TestZeroLimitsPermitNoTraffic` in `internal/execution/budget_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestZeroLimitsPermitNoTraffic$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestZeroLimitsPermitNoTraffic in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0007: Asserts that with zero byte and connection limits, as for NET-01, a dial is refused before the path dialer is called, the server accepts nothing, nothing is metered and the attempt stops with `resource_limit`.

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

<a id="UT.internal.model.task.TestGeneratedKeysAccepted"></a>

### UT.internal.model.task.TestGeneratedKeysAccepted

- Case: `UT.internal.model.task.TestGeneratedKeysAccepted`
- Result: passed
- Test: `TestGeneratedKeysAccepted` in `internal/model/task/keys_internal_test.go` (package `internal/model/task`)
- Reproduce one case: `go test -count=1 -run '^TestGeneratedKeysAccepted$' ./internal/model/task`
- Scope: Existing synthetic Go unit test TestGeneratedKeysAccepted in package internal/model/task, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0051: Asserts that 32 public keys generated by `crypto/ed25519` from fixed seeds decode, re-encode to the same bytes and are accepted.

<a id="UT.internal.model.task.TestMalformedSourceRejectedAtomically"></a>

### UT.internal.model.task.TestMalformedSourceRejectedAtomically

- Case: `UT.internal.model.task.TestMalformedSourceRejectedAtomically`
- Result: passed
- Test: `TestMalformedSourceRejectedAtomically` in `internal/model/task/trust_test.go` (package `internal/model/task`)
- Reproduce one case: `go test -count=1 -run '^TestMalformedSourceRejectedAtomically$' ./internal/model/task`
- Scope: Existing synthetic Go unit test TestMalformedSourceRejectedAtomically in package internal/model/task, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0043: Asserts that a remote source cannot be completed with an empty, uppercase or over-long issuer or a zero, fractional, sub-second or non-UTC issuance time, and that a zero source cannot be completed, so no host-built value stands in for a verified package.
  - REQ-TRUST-0044: Asserts that NewSourceContext returns ErrTrustModel and an invalid context for a remote source with a host-supplied issuer, an unknown origin, a zero policy generation, or an embedded source without a valid issuer.

<a id="UT.internal.model.task.TestNonCanonicalAndOffCurveKeysRejected"></a>

### UT.internal.model.task.TestNonCanonicalAndOffCurveKeysRejected

- Case: `UT.internal.model.task.TestNonCanonicalAndOffCurveKeysRejected`
- Result: passed
- Test: `TestNonCanonicalAndOffCurveKeysRejected` in `internal/model/task/keys_internal_test.go` (package `internal/model/task`)
- Reproduce one case: `go test -count=1 -run '^TestNonCanonicalAndOffCurveKeysRejected$' ./internal/model/task`
- Scope: Existing synthetic Go unit test TestNonCanonicalAndOffCurveKeysRejected in package internal/model/task, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0051: Asserts that encodings with y equal to p, p + 1 or 2^255 - 1, and with x = 0 and the sign bit set, fail canonical decoding, and that small y values with no curve point fail decoding and are rejected as keys.

<a id="UT.internal.model.task.TestPolicyCarriesBoundedTrustedKeys"></a>

### UT.internal.model.task.TestPolicyCarriesBoundedTrustedKeys

- Case: `UT.internal.model.task.TestPolicyCarriesBoundedTrustedKeys`
- Result: passed
- Test: `TestPolicyCarriesBoundedTrustedKeys` in `internal/model/task/keys_test.go` (package `internal/model/task`)
- Reproduce one case: `go test -count=1 -run '^TestPolicyCarriesBoundedTrustedKeys$' ./internal/model/task`
- Scope: Existing synthetic Go unit test TestPolicyCarriesBoundedTrustedKeys in package internal/model/task, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0032: Asserts that four keys for one issuer and 32 keys across eight issuers, the registered maxima, are accepted with a 365-day validity.
  - REQ-TRUST-0051: Asserts that a valid key record is returned by exact key ID without aliasing caller or getter memory, that an unknown, case-different or zero-policy lookup finds nothing, and that a policy without keys stays valid.

<a id="UT.internal.model.task.TestPolicyRejectsMalformedOrAmbiguousKeys"></a>

### UT.internal.model.task.TestPolicyRejectsMalformedOrAmbiguousKeys

- Case: `UT.internal.model.task.TestPolicyRejectsMalformedOrAmbiguousKeys`
- Result: passed
- Test: `TestPolicyRejectsMalformedOrAmbiguousKeys` in `internal/model/task/keys_test.go` (package `internal/model/task`)
- Reproduce one case: `go test -count=1 -run '^TestPolicyRejectsMalformedOrAmbiguousKeys$' ./internal/model/task`
- Scope: Existing synthetic Go unit test TestPolicyRejectsMalformedOrAmbiguousKeys in package internal/model/task, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0032: Asserts rejection of the whole policy for a validity one second above 365 days, a fifth key for one issuer and a 33rd key in one policy.
  - REQ-TRUST-0051: Asserts that the whole policy is rejected, with the fixed error and no exposed key, for an empty, uppercase, over-long or space-containing key ID, a key ID duplicated for the same or another issuer, an unknown or empty issuer, a zero, identity or non-canonical public key, and equal, reversed, zero, fractional, sub-second or non-UTC validity times.

<a id="UT.internal.model.task.TestSmallOrderPointsDecodeAndAreRejected"></a>

### UT.internal.model.task.TestSmallOrderPointsDecodeAndAreRejected

- Case: `UT.internal.model.task.TestSmallOrderPointsDecodeAndAreRejected`
- Result: passed
- Test: `TestSmallOrderPointsDecodeAndAreRejected` in `internal/model/task/keys_internal_test.go` (package `internal/model/task`)
- Reproduce one case: `go test -count=1 -run '^TestSmallOrderPointsDecodeAndAreRejected$' ./internal/model/task`
- Scope: Existing synthetic Go unit test TestSmallOrderPointsDecodeAndAreRejected in package internal/model/task, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0051: Asserts that the canonical encodings of the eight edwards25519 points of small order decode as eight distinct curve points whose eightfold multiple is the identity, and that each is rejected as a public key.

<a id="UT.internal.model.task.TestTrustedKeyCoversInclusiveValidity"></a>

### UT.internal.model.task.TestTrustedKeyCoversInclusiveValidity

- Case: `UT.internal.model.task.TestTrustedKeyCoversInclusiveValidity`
- Result: passed
- Test: `TestTrustedKeyCoversInclusiveValidity` in `internal/model/task/keys_test.go` (package `internal/model/task`)
- Reproduce one case: `go test -count=1 -run '^TestTrustedKeyCoversInclusiveValidity$' ./internal/model/task`
- Scope: Existing synthetic Go unit test TestTrustedKeyCoversInclusiveValidity in package internal/model/task, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0052: Asserts that the check-7 validity test includes `not_before` and `not_after` themselves and excludes one second or one nanosecond outside either end.

<a id="UT.internal.persistence.TestReplayRecordRetention"></a>

### UT.internal.persistence.TestReplayRecordRetention

- Case: `UT.internal.persistence.TestReplayRecordRetention`
- Result: passed
- Test: `TestReplayRecordRetention` in `internal/persistence/replay_test.go` (package `internal/persistence`)
- Reproduce one case: `go test -count=1 -run '^TestReplayRecordRetention$' ./internal/persistence`
- Scope: Existing synthetic Go unit test TestReplayRecordRetention in package internal/persistence, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0023: Asserts that persistence.NewReplayRecord retains a record until 86,400 seconds after the task's not_after, rounded up to a whole second for a fractional not_after, capped at 691,200 seconds after the acceptance time truncated to a whole second, and that the record is retained one nanosecond before and overdue at and after its retention end.
  - REQ-TRUST-0021: Asserts that the record built for an accepted task carries its replay key, digest, accepted policy generation and origin and is valid for task validities from one second to seven days.

<a id="UT.internal.persistence.TestStateReplayCommitsConditionallyWithinCapacity"></a>

### UT.internal.persistence.TestStateReplayCommitsConditionallyWithinCapacity

- Case: `UT.internal.persistence.TestStateReplayCommitsConditionallyWithinCapacity`
- Result: passed
- Test: `TestStateReplayCommitsConditionallyWithinCapacity` in `internal/persistence/state_test.go` (package `internal/persistence`)
- Reproduce one case: `go test -count=1 -run '^TestStateReplayCommitsConditionallyWithinCapacity$' ./internal/persistence`
- Scope: Existing synthetic Go unit test TestStateReplayCommitsConditionallyWithinCapacity in package internal/persistence, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0021: Asserts that `persistence.State` over the fault-injectable record store commits a new replay record, returns the retained record without writing for the same key with another digest, rejects a commit beyond its capacity of two without evicting a retained record, reports an absent key as absent and returns the highest retained accepted generation.
  - REQ-TRUST-0022: Asserts that the record looked up after commit equals the committed key, digest, policy generation, origin and expiry classes, and that stored record names are the digests of the two committed replay keys rather than issuer or task identifiers.

<a id="UT.internal.persistence.TestStateRetentionBoundaries"></a>

### UT.internal.persistence.TestStateRetentionBoundaries

- Case: `UT.internal.persistence.TestStateRetentionBoundaries`
- Result: passed
- Test: `TestStateRetentionBoundaries` in `internal/persistence/state_test.go` (package `internal/persistence`)
- Reproduce one case: `go test -count=1 -run '^TestStateRetentionBoundaries$' ./internal/persistence`
- Scope: Existing synthetic Go unit test TestStateRetentionBoundaries in package internal/persistence, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0023: Asserts on the injected time that a replay record and the run record of its key are retained one nanosecond before the replay retention end and are reported absent, excluded from the highest generation and deleted from the record store at and one second after it, both on access and when the state is opened after a simulated restart at those times, and that a run record cannot be written at the retention end.

<a id="UT.internal.persistence.TestStateStoresOnlyRecordFields"></a>

### UT.internal.persistence.TestStateStoresOnlyRecordFields

- Case: `UT.internal.persistence.TestStateStoresOnlyRecordFields`
- Result: passed
- Test: `TestStateStoresOnlyRecordFields` in `internal/persistence/state_test.go` (package `internal/persistence`)
- Reproduce one case: `go test -count=1 -run '^TestStateStoresOnlyRecordFields$' ./internal/persistence`
- Scope: Existing synthetic Go unit test TestStateStoresOnlyRecordFields in package internal/persistence, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0023: Asserts that the stored replay value is exactly the size of its encoded key, digest, generation, origin and two expiry fields, and the stored run value exactly the size of its key, anchor, two counters, due time and retention end, so neither holds task bytes, input, targets, labels or other values.

<a id="UT.internal.persistence.pending.TestAggregateBytesAndCapacity"></a>

### UT.internal.persistence.pending.TestAggregateBytesAndCapacity

- Case: `UT.internal.persistence.pending.TestAggregateBytesAndCapacity`
- Result: passed
- Test: `TestAggregateBytesAndCapacity` in `internal/persistence/pending/pending_test.go` (package `internal/persistence/pending`)
- Reproduce one case: `go test -count=1 -run '^TestAggregateBytesAndCapacity$' ./internal/persistence/pending`
- Scope: Existing synthetic Go unit test TestAggregateBytesAndCapacity in package internal/persistence/pending, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0022: Asserts on the production pending store that a result exceeding a lowered aggregate byte limit is rejected with a bounded error and leaves storage byte-for-byte unchanged, that the stored total never exceeds the limit, and that capacity counts maximum-size results only in new bundles: 63 at the registered maxima, 2 with room for exactly two maximum results in one bundle, 1 one byte short of that, and 0 once a result is stored in the smallest store.

<a id="UT.internal.persistence.pending.TestCountLimits"></a>

### UT.internal.persistence.pending.TestCountLimits

- Case: `UT.internal.persistence.pending.TestCountLimits`
- Result: passed
- Test: `TestCountLimits` in `internal/persistence/pending/pending_test.go` (package `internal/persistence/pending`)
- Reproduce one case: `go test -count=1 -run '^TestCountLimits$' ./internal/persistence/pending`
- Scope: Existing synthetic Go unit test TestCountLimits in package internal/persistence/pending, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0021: Asserts on the production pending store that results 1 to 16 share one bundle and the 17th starts a second, and that 32 single-result bundles are held while a 33rd is rejected with capacity zero.
  - REQ-LIMIT-0022: Asserts that the result rejected at the 32-bundle maximum is not stored and leaves storage byte-for-byte unchanged, and that acknowledging a bundle frees capacity so a later result is accepted.

<a id="UT.internal.persistence.pending.TestRetentionBoundary"></a>

### UT.internal.persistence.pending.TestRetentionBoundary

- Case: `UT.internal.persistence.pending.TestRetentionBoundary`
- Result: passed
- Test: `TestRetentionBoundary` in `internal/persistence/pending/pending_test.go` (package `internal/persistence/pending`)
- Reproduce one case: `go test -count=1 -run '^TestRetentionBoundary$' ./internal/persistence/pending`
- Scope: Existing synthetic Go unit test TestRetentionBoundary in package internal/persistence/pending, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0024: Asserts on the injected clock that a pending bundle is returned one nanosecond before 86,400 seconds after its first result was stored and is deleted, never returned, at and one second after that time, on access and when the store is opened after a simulated restart; that a later result does not extend the bundle's retention; that an expired returned bundle cannot be acknowledged and is removed; and that while an expiry deletion fails the bundle is not returned and no result is stored, until a retry deletes it.
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
  - REQ-RESULT-0029: Rejects TCP-01 incomplete results that include a measurement.
  - REQ-RESULT-0034: Rejects incomplete/not_run results without reason or with measurement, and not_run results that carry network_context.
  - REQ-RESULT-0042: For incomplete and not_run TCP-01 results, accepts each of 15 registered reasons and rejects a missing reason and reason 'timeout'.
  - REQ-RESULT-0053: For each of the 15 registered reasons, rejects a `not_run` TCP-01 result that carries `network_context` and accepts it without one; rejects an `incomplete` result without context.

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
  - REQ-RESULT-0026: Accepts registered outcomes with ordered observations for TCP-01, TLS-01, DNS-01, DNS-03, DNS-04, HTTP-01, NET-01, and NET-03, rejects outcome 'censored', and rejects out-of-order or mismatched observation sequences.
  - REQ-RESULT-0027: For each initial test, asserts a measurement containing an extra raw_body field is rejected.
  - REQ-RESULT-0041: Rejects outcome 'censored' for each initial test, an incomplete tcp_connect observation in a completed TCP-01 result, and TLS-01 outcomes inconsistent with observations.
  - REQ-RESULT-0043: Rejects a TCP-01 result carrying a tls_handshake observation and TLS-01 results with wrong observation order or missing handshake.
  - REQ-RESULT-0046: Rejects a measurement with an added raw_body field for each initial test.
  - REQ-RESULT-0053: Asserts a NET-01 result that adds a network_context object is rejected, NET-01 using its access-only representation instead.
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

<a id="UT.internal.taskdecode.TestStagedDecoding"></a>

### UT.internal.taskdecode.TestStagedDecoding

- Case: `UT.internal.taskdecode.TestStagedDecoding`
- Result: passed
- Test: `TestStagedDecoding` in `internal/taskdecode/staged_test.go` (package `internal/taskdecode`)
- Reproduce one case: `go test -count=1 -run '^TestStagedDecoding$' ./internal/taskdecode`
- Scope: Existing synthetic Go unit test TestStagedDecoding in package internal/taskdecode, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0008: Asserts the staged decoder selects the exact (test_id, test_version) after shape validation, returns ErrUnregistered (also ErrVersion) only for a well-formed unregistered pair and ErrVersion for a malformed pair or unsupported schema_version, and validates input against the selected test schema only in Parameters, without invoking the codec on failure.

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

<a id="UT.internal.validation.TestRefiningClasses"></a>

### UT.internal.validation.TestRefiningClasses

- Case: `UT.internal.validation.TestRefiningClasses`
- Result: passed
- Test: `TestRefiningClasses` in `internal/validation/limits_test.go` (package `internal/validation`)
- Reproduce one case: `go test -count=1 -run '^TestRefiningClasses$' ./internal/validation`
- Scope: Existing synthetic Go unit test TestRefiningClasses in package internal/validation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0002: Asserts values above registered duration, connection, run, attempt, label and repeat-jitter maxima are rejected matching ErrLimit, values below minima are rejected without ErrLimit, and a not_after beyond seven days matches ErrLimit; nothing is accepted after clamping.
  - REQ-SAFETY-0017: Asserts that validation.Common rejects an unrecognized risk factor type matching ErrUnsupported and ErrTask, and that an otherwise invalid factor is not reported as unsupported.

<a id="UT.internal.validation.TestTaskIDRulesAndLimitClass"></a>

### UT.internal.validation.TestTaskIDRulesAndLimitClass

- Case: `UT.internal.validation.TestTaskIDRulesAndLimitClass`
- Result: passed
- Test: `TestTaskIDRulesAndLimitClass` in `internal/validation/limits_test.go` (package `internal/validation`)
- Reproduce one case: `go test -count=1 -run '^TestTaskIDRulesAndLimitClass$' ./internal/validation`
- Scope: Existing synthetic Go unit test TestTaskIDRulesAndLimitClass in package internal/validation, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0004: Asserts 129-byte task_id values match ErrLimit as well as ErrTask, whatever their characters, and that no shorter invalid identifier matches ErrLimit.
  - REQ-TRUST-0019: Asserts validation.TaskID accepts 1 and 128 allowed bytes and rejects empty, uppercase, space, '/', non-ASCII and leading '-' identifiers with ErrTask.

<a id="UT.root.TestAdmissionHasNoNetworkDependency"></a>

### UT.root.TestAdmissionHasNoNetworkDependency

- Case: `UT.root.TestAdmissionHasNoNetworkDependency`
- Result: passed
- Test: `TestAdmissionHasNoNetworkDependency` in `dependency_rules_test.go` (root package, Go package `barometer_test`)
- Reproduce one case: `go test -count=1 -run '^TestAdmissionHasNoNetworkDependency$' .`
- Scope: Existing synthetic Go unit test TestAdmissionHasNoNetworkDependency in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0018: Asserts with `go list -deps`, for the host build, that the transitive dependency closure of the root package and of `internal/admission` contains no `net` package other than `net/netip` and `net/url` and no `crypto/tls`, `os/exec`, `plugin` or `internal/testkit` package, so admission through the facade links no code that can open a network connection. This is structural evidence: it does not observe runtime behavior, cover mobile builds or constrain a future trusted adapter seam.

<a id="UT.root.TestAdmitTaskAcceptsReviewedOrigins"></a>

### UT.root.TestAdmitTaskAcceptsReviewedOrigins

- Case: `UT.root.TestAdmitTaskAcceptsReviewedOrigins`
- Result: passed
- Test: `TestAdmitTaskAcceptsReviewedOrigins` in `admit_task_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestAdmitTaskAcceptsReviewedOrigins$' .`
- Scope: Existing synthetic Go unit test TestAdmitTaskAcceptsReviewedOrigins in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0046: Asserts that `Client.AdmitTask`, for an embedded-origin record and for a remote record presenting a signed package, resolves the delivery record, runs the implemented admission steps through `Admitter.Admit`, commits exactly one replay record and only then returns a `T.`-prefixed task handle that resolves to the accepted task with its `not_after` expiry, allocating no run.

<a id="UT.root.TestAdmitTaskErrorsCarryNoInput"></a>

### UT.root.TestAdmitTaskErrorsCarryNoInput

- Case: `UT.root.TestAdmitTaskErrorsCarryNoInput`
- Result: passed
- Test: `TestAdmitTaskErrorsCarryNoInput` in `admit_task_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestAdmitTaskErrorsCarryNoInput$' .`
- Scope: Existing synthetic Go unit test TestAdmitTaskErrorsCarryNoInput in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0035: Asserts that `AdmitTask` errors for an unknown source reference, an invalid task, a `task_id` conflict and a closed client, with a canary in the `task_id`, a task member name and the source reference, format under `%v`, `%+v`, `%#v` and `%s` without the canary, the published handle or the registered target name, and that each is exactly one unwrapped fixed code: either a `task.AdmissionError` class or an API-local code.

<a id="UT.root.TestAdmitTaskExactRedeliveryReturnsTheSameHandle"></a>

### UT.root.TestAdmitTaskExactRedeliveryReturnsTheSameHandle

- Case: `UT.root.TestAdmitTaskExactRedeliveryReturnsTheSameHandle`
- Result: passed
- Test: `TestAdmitTaskExactRedeliveryReturnsTheSameHandle` in `admit_task_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestAdmitTaskExactRedeliveryReturnsTheSameHandle$' .`
- Scope: Existing synthetic Go unit test TestAdmitTaskExactRedeliveryReturnsTheSameHandle in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0024: Asserts through `Client.AdmitTask` that redelivery of the exact bytes 30 seconds later, through the same delivery reference and through a newly installed one, returns the same task handle, commits nothing, allocates no run and leaves the retained replay record, including its expiry and retention end, unchanged; and that bytes differing only in `not_after` under the same `task_id` return `task_id_conflict` with no handle and no state change.

<a id="UT.root.TestAdmitTaskHostileSources"></a>

### UT.root.TestAdmitTaskHostileSources

- Case: `UT.root.TestAdmitTaskHostileSources`
- Result: passed
- Test: `TestAdmitTaskHostileSources` in `admit_task_hostile_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestAdmitTaskHostileSources$' .`
- Scope: Existing synthetic Go unit test TestAdmitTaskHostileSources in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0042: Asserts through `Client.AdmitTask` that a bare task presented as remote, without a signed package, is rejected as `task_source_unauthenticated` with no task handle, replay record or run.
  - REQ-TRUST-0043: Asserts through `Client.AdmitTask` that a package with an altered task byte, a package signed by an unknown key, a package re-signed under another issuer's key and a package issued 301 seconds after the trusted clock are each rejected as `task_source_unauthenticated` with no task handle, replay record or run.
  - REQ-TRUST-0046: Asserts that step 1 rejects an empty reference, a forged reference, a reference from another delivery-store instance, a record bound to other task bytes, a remote reference presented a second time and an expired record as `task_source_unauthenticated`, with no task handle, replay record or run.
  - REQ-TRUST-0052: Asserts through `Client.AdmitTask` the check-4, check-5, check-8 and check-9 rejections for an unknown key, a key of another issuer, an issuance 301 seconds ahead and an altered task byte.

<a id="UT.root.TestAdmitTaskRejectionsAreAtomic"></a>

### UT.root.TestAdmitTaskRejectionsAreAtomic

- Case: `UT.root.TestAdmitTaskRejectionsAreAtomic`
- Result: passed
- Test: `TestAdmitTaskRejectionsAreAtomic` in `admit_task_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestAdmitTaskRejectionsAreAtomic$' .`
- Scope: Existing synthetic Go unit test TestAdmitTaskRejectionsAreAtomic in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0038: Asserts that for embedded-origin tasks an unavailable replay lookup and a commit that crashes before becoming durable both return `task_trust_unavailable` through `Client.AdmitTask`, with no handle, no durable replay record and no run.
  - REQ-TRUST-0047: Asserts through `Client.AdmitTask` the class of the first failing step: `task_source_unauthenticated` for an unknown reference, for bytes that differ from the record and for 65,537 bytes presented on another task's reference (step 1 before the step-3 byte limit); `task_limit_exceeded` for a recorded complete task one byte over 65,536 bytes; `task_schema_invalid`, `task_expired` and `task_unsupported` for an invalid schema, an expired task and an unregistered test version; `task_limit_exceeded` for `max_duration_ms` 60,001; `task_issuer_not_authorized` for an issuer the policy does not list; and `policy_rollback` while a record from a higher generation is retained. Each rejection is a `task.AdmissionError` and leaves no handle, new replay record or run.

<a id="UT.root.TestAdmitTaskSignedPackageStrictJSON"></a>

### UT.root.TestAdmitTaskSignedPackageStrictJSON

- Case: `UT.root.TestAdmitTaskSignedPackageStrictJSON`
- Result: passed
- Test: `TestAdmitTaskSignedPackageStrictJSON` in `admit_task_hostile_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestAdmitTaskSignedPackageStrictJSON$' .`
- Scope: Existing synthetic Go unit test TestAdmitTaskSignedPackageStrictJSON in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0001: Asserts through `Client.AdmitTask`, for a remote signed package, that a package followed by a second value and a package preceded by a byte-order mark are rejected as `task_source_unauthenticated`.
  - REQ-JSON-0002: Asserts that an invalid UTF-8 byte in a member name or in a member value, and an unpaired `\ud800` escape in a member value, are rejected as `task_source_unauthenticated`.
  - REQ-JSON-0003: Asserts that an equal and a conflicting repetition of `key_id`, and escaped-equivalent duplicates of `key_id` and `task`, are each rejected as `task_source_unauthenticated`.
  - REQ-JSON-0004: Asserts that an unknown member, an escaped spelling of an unknown name and a case variant of `key_id` are rejected as `task_source_unauthenticated`, and that an escaped spelling of the known name `key_id` is admitted.
  - REQ-JSON-0005: Asserts that an object where the string `key_id` is required is rejected as `task_source_unauthenticated`.
  - REQ-JSON-0009: Asserts that a package whose fully escaped `task` member, padded with whitespace, is one byte over 98,304 original bytes is rejected as `task_limit_exceeded`, while the same package at exactly 98,304 bytes is admitted.
  - REQ-JSON-0010: Asserts that nesting 200 levels deep is rejected as `task_source_unauthenticated`, and that every rejection in this test is exactly one fixed, unwrapped public admission class leaving no task handle, replay record or run.
  - REQ-TRUST-0048: Asserts through the facade that the package is a closed object of string members: duplicate, unknown, case-variant and nested members are rejected before the task is parsed, and escaped spellings of a member name and of a member value decode to the same verified package and admit.
  - REQ-LIMIT-0032: Asserts that the signed-package byte limit is applied to the original presented bytes, including JSON escapes and whitespace: one byte over is `task_limit_exceeded` and exactly at the limit admits.

<a id="UT.root.TestAdmitTaskStrictJSON"></a>

### UT.root.TestAdmitTaskStrictJSON

- Case: `UT.root.TestAdmitTaskStrictJSON`
- Result: passed
- Test: `TestAdmitTaskStrictJSON` in `admit_task_hostile_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestAdmitTaskStrictJSON$' .`
- Scope: Existing synthetic Go unit test TestAdmitTaskStrictJSON in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0001: Asserts through `Client.AdmitTask`, for embedded and remote origin, that a task followed by a second value and a task preceded by a byte-order mark are rejected as `task_schema_invalid`.
  - REQ-JSON-0002: Asserts that a member name containing an invalid UTF-8 byte is rejected as `task_schema_invalid`.
  - REQ-JSON-0003: Asserts that equal duplicate names at the top level and in `input` and `execution.limits`, and escaped-equivalent duplicates of `task_id` and `schedule.not_after`, are each rejected as `task_schema_invalid`, even when the repeated values are equal.
  - REQ-JSON-0004: Asserts that an escaped spelling of an unknown top-level name and the unknown `authenticated` and `input.extra` members are rejected as `task_schema_invalid`, and that an escaped spelling of the exact `schema_version` value is admitted as that version.
  - REQ-JSON-0006: Asserts that `-0`, `1.0` and `1e0` for the integer `max_connections` are rejected as `task_schema_invalid`.
  - REQ-JSON-0007: Asserts that `max_connections` 9,007,199,254,740,993, beyond the exact float range and above its registered maximum, is rejected as `task_limit_exceeded` rather than rounded.
  - REQ-JSON-0008: Asserts that `schema_version` `0.1.00` is rejected as `task_schema_invalid` with no fallback to `0.1.0`.
  - REQ-JSON-0009: Asserts that a task whose escaped string makes its original encoded bytes exceed 65,536, while its decoded text would not, is rejected as `task_limit_exceeded`.
  - REQ-JSON-0010: Asserts that nesting 200 levels deep is rejected as `task_schema_invalid`, and that every rejection in this test is exactly one fixed, unwrapped public admission class leaving no task handle, replay record or run.

<a id="UT.root.TestCancelRunCleanupIncomplete"></a>

### UT.root.TestCancelRunCleanupIncomplete

- Case: `UT.root.TestCancelRunCleanupIncomplete`
- Result: passed
- Test: `TestCancelRunCleanupIncomplete` in `cancel_run_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCancelRunCleanupIncomplete$' .`
- Scope: Existing synthetic Go unit test TestCancelRunCleanupIncomplete in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0008: Asserts with a runner that ignores cancellation that `Client.CancelRun` is still waiting one millisecond before the 2,000 ms cleanup bound on the engine clock and returns `cleanup_incomplete` at it, that the run stays `active` and keeps its exclusive reservation, and that `Close` returns `cleanup_incomplete` until the runner returns.

<a id="UT.root.TestCancelRunStopsTheRun"></a>

### UT.root.TestCancelRunStopsTheRun

- Case: `UT.root.TestCancelRunStopsTheRun`
- Result: passed
- Test: `TestCancelRunStopsTheRun` in `cancel_run_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCancelRunStopsTheRun$' .`
- Scope: Existing synthetic Go unit test TestCancelRunStopsTheRun in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0020: Asserts that `Client.CancelRun` on a held run records its attempt with `cancelled_by_host`, starts no permitted retry and returns with the run terminal, and that three repeated cancellations and a cancellation of a completed run leave the written results byte-identical.
  - REQ-RESULT-0051: Asserts that the cancelled attempt, whose test had started, is `incomplete`.

<a id="UT.root.TestClearInvalidatesHandlesAndDeliveryReferences"></a>

### UT.root.TestClearInvalidatesHandlesAndDeliveryReferences

- Case: `UT.root.TestClearInvalidatesHandlesAndDeliveryReferences`
- Result: passed
- Test: `TestClearInvalidatesHandlesAndDeliveryReferences` in `client_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestClearInvalidatesHandlesAndDeliveryReferences$' .`
- Scope: Existing synthetic Go unit test TestClearInvalidatesHandlesAndDeliveryReferences in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LINKAGE-0023: Asserts that ClearMeasurementData on an initialized client whose record-port deletions fail returns clearing_incomplete and never success (item 7), leaves the client blocked so that a new entry-point reservation for admission, runs or bundles returns clearing_incomplete (item 1), and invalidates an issued task handle and an installed delivery reference.

<a id="UT.root.TestConsentIsRecheckedBeforeARetry"></a>

### UT.root.TestConsentIsRecheckedBeforeARetry

- Case: `UT.root.TestConsentIsRecheckedBeforeARetry`
- Result: passed
- Test: `TestConsentIsRecheckedBeforeARetry` in `runs_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestConsentIsRecheckedBeforeARetry$' .`
- Scope: Existing synthetic Go unit test TestConsentIsRecheckedBeforeARetry in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0020: Asserts through `Client.StartRun` that consent withdrawn during the first attempt of a consent-requiring task prevents its retry: the run ends with one `incomplete` result and one runner call.

<a id="UT.root.TestInitializeStrictJSON"></a>

### UT.root.TestInitializeStrictJSON

- Case: `UT.root.TestInitializeStrictJSON`
- Result: passed
- Test: `TestInitializeStrictJSON` in `initialize_hostile_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestInitializeStrictJSON$' .`
- Scope: Existing synthetic Go unit test TestInitializeStrictJSON in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0001: Asserts through `Client.Initialize` that a configuration followed by a second value and a configuration preceded by a byte-order mark are rejected as `not_initialized`.
  - REQ-JSON-0002: Asserts that an invalid UTF-8 byte in a member name or in a member value, and an unpaired `\ud800` escape in a member value, are rejected as `not_initialized`.
  - REQ-JSON-0003: Asserts that an equal and an escaped-equivalent repetition of `proxy`, a conflicting repetition of `port`, an escaped-equivalent repetition of `host`, a repetition of `fallback_endpoint` and an escaped-equivalent repetition of `mode` are each rejected as `not_initialized`.
  - REQ-JSON-0004: Asserts that escaped spellings of an unknown top-level name and of an unknown `proxy` member, and the case variant `Proxy`, are rejected as `not_initialized`, and that an escaped spelling of the known name `proxy` is accepted.
  - REQ-JSON-0005: Asserts that the quoted string `"54321"` for the integer `port` is rejected as `not_initialized` rather than coerced.
  - REQ-JSON-0006: Asserts that `-0`, `54321.0` and `5.4321e4` for `port`, and `604800.0`, `6.048e5` and `-0` for `rotation_period_seconds`, are rejected as `not_initialized`.
  - REQ-JSON-0007: Asserts that `port` 9,007,199,254,740,993, beyond the exact float range and above its maximum, is rejected as `not_initialized` rather than rounded.
  - REQ-JSON-0009: Asserts that a configuration whose escaped `scope` value, padded with whitespace, is one byte over 16,384 original bytes is rejected as `not_initialized`, while the same configuration at exactly 16,384 bytes is accepted.
  - REQ-JSON-0010: Asserts that nesting 200 levels deep is rejected as `not_initialized`, and that every rejection leaves the client uninitialized with no runtime generation and no linkage key store call, after which a valid initialization succeeds.

<a id="UT.root.TestLifecycleTransitionTable"></a>

### UT.root.TestLifecycleTransitionTable

- Case: `UT.root.TestLifecycleTransitionTable`
- Result: passed
- Test: `TestLifecycleTransitionTable` in `client_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestLifecycleTransitionTable$' .`
- Scope: Existing synthetic Go unit test TestLifecycleTransitionTable in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LINKAGE-0023: Asserts that clearing a zero-value client, which has no storage scope, returns clearing_incomplete and never success; that with a failing record-port deletion clearing an initialized client returns clearing_incomplete and leaves it blocked, and a repeated clear stays incomplete and blocked; and that with the clearing capability derived from the synthetic record port, clearing from the uninitialized, initialized and close-incomplete states succeeds and leaves the client uninitialized, and from the closed state succeeds and keeps it closed (item 7).
  - REQ-LINKAGE-0024: Asserts that clearing is safe to call repeatedly: before initialization, after an incomplete clear, after an incomplete close and after close it returns a fixed result without panicking, and it never reopens a closed client.

<a id="UT.root.TestLifecycleTransitionsCancelRuns"></a>

### UT.root.TestLifecycleTransitionsCancelRuns

- Case: `UT.root.TestLifecycleTransitionsCancelRuns`
- Result: passed
- Test: `TestLifecycleTransitionsCancelRuns` in `cancel_run_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestLifecycleTransitionsCancelRuns$' .`
- Scope: Existing synthetic Go unit test TestLifecycleTransitionsCancelRuns in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LINKAGE-0023: Asserts that clearing, like close and policy replacement, cancels an active run and waits for it before returning, and that the run's attempt is recorded as `cancelled_by_host`.

<a id="UT.root.TestPanicsAreRecoveredToFixedCodes"></a>

### UT.root.TestPanicsAreRecoveredToFixedCodes

- Case: `UT.root.TestPanicsAreRecoveredToFixedCodes`
- Result: passed
- Test: `TestPanicsAreRecoveredToFixedCodes` in `client_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestPanicsAreRecoveredToFixedCodes$' .`
- Scope: Existing synthetic Go unit test TestPanicsAreRecoveredToFixedCodes in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LINKAGE-0023: Asserts that a panicking synthetic persistent-clear capability is reported as clearing_incomplete with no panic text in the formatted error, keeps an initialized client blocked until close, and keeps a closed client closed (item 7).

<a id="UT.root.TestRootPackageImportsNoNetworkPackages"></a>

### UT.root.TestRootPackageImportsNoNetworkPackages

- Case: `UT.root.TestRootPackageImportsNoNetworkPackages`
- Result: passed
- Test: `TestRootPackageImportsNoNetworkPackages` in `public_api_test.go` (root package, Go package `barometer_test`)
- Reproduce one case: `go test -count=1 -run '^TestRootPackageImportsNoNetworkPackages$' .`
- Scope: Existing synthetic Go unit test TestRootPackageImportsNoNetworkPackages in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LINKAGE-0024: Asserts that the root package's production files, which implement ClearMeasurementData, import no network package other than net/netip and net/url and no test support. This is an import-level check of the clearing code itself, not an observation of absent traffic through adapters.

<a id="UT.root.TestStartRunBudget"></a>

### UT.root.TestStartRunBudget

- Case: `UT.root.TestStartRunBudget`
- Result: passed
- Test: `TestStartRunBudget` in `runs_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestStartRunBudget$' .`
- Scope: Existing synthetic Go unit test TestStartRunBudget in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0030: Asserts through `Client.StartRun` that a repeated task with `max_runs` 1, 2 and 3, offered one more due occurrence than its limit, starts exactly `max_runs` runs, with the ledger and the runner recording the same count.

<a id="UT.root.TestStartRunConcurrency"></a>

### UT.root.TestStartRunConcurrency

- Case: `UT.root.TestStartRunConcurrency`
- Result: passed
- Test: `TestStartRunConcurrency` in `runs_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestStartRunConcurrency$' .`
- Scope: Existing synthetic Go unit test TestStartRunConcurrency in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0027: Asserts through `Client.StartRun` that an `exclusive` task does not start while a `shared` run is active and starts once it is terminal, and that no other run starts while the exclusive run is active.

<a id="UT.root.TestStartRunDuplicatesShareOneRun"></a>

### UT.root.TestStartRunDuplicatesShareOneRun

- Case: `UT.root.TestStartRunDuplicatesShareOneRun`
- Result: passed
- Test: `TestStartRunDuplicatesShareOneRun` in `runs_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestStartRunDuplicatesShareOneRun$' .`
- Scope: Existing synthetic Go unit test TestStartRunDuplicatesShareOneRun in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0030: Asserts that 16 concurrent starts of a one-time task return one run handle and allocate one run, and that a start after the run is terminal returns the same handle: the ledger, the runner, the results and the run table each record one run.

<a id="UT.root.TestStartRunEnforcesTheAttemptDeadline"></a>

### UT.root.TestStartRunEnforcesTheAttemptDeadline

- Case: `UT.root.TestStartRunEnforcesTheAttemptDeadline`
- Result: passed
- Test: `TestStartRunEnforcesTheAttemptDeadline` in `cancel_run_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestStartRunEnforcesTheAttemptDeadline$' .`
- Scope: Existing synthetic Go unit test TestStartRunEnforcesTheAttemptDeadline in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0006: Asserts through `Client.StartRun` that a task's `max_duration_ms` of 1,000 ms leaves the run active after 999 ms on the engine clock and ends its attempt at 1,000 ms as `incomplete` with `attempt_deadline_exceeded`.

<a id="UT.root.TestStartRunScheduleBoundaries"></a>

### UT.root.TestStartRunScheduleBoundaries

- Case: `UT.root.TestStartRunScheduleBoundaries`
- Result: passed
- Test: `TestStartRunScheduleBoundaries` in `runs_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestStartRunScheduleBoundaries$' .`
- Scope: Existing synthetic Go unit test TestStartRunScheduleBoundaries in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0007: Asserts through `Client.StartRun` on the injected clock that a start one second before `starts_at` returns `run_not_due` and reserves nothing, while starts at and one second after it run.
  - REQ-SCHEDULE-0034: Asserts that a start runs one second before `not_after` and not at it, that a 12:00–18:00 window permits starts at 12:00 and 17:59 and not at 11:59 or 18:00, that a 22:00–06:00 window permits 22:00 and 05:59 and not 21:59 or 06:00, and that a window without a known device time zone is unsatisfied; every refused start reserves no run and invokes no runner.

<a id="UT.root.TestNextBundleAndAcknowledge"></a>

### UT.root.TestNextBundleAndAcknowledge

- Case: `UT.root.TestNextBundleAndAcknowledge`
- Result: passed
- Test: `TestNextBundleAndAcknowledge` in `bundles_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestNextBundleAndAcknowledge$' .`
- Scope: Existing synthetic Go unit test TestNextBundleAndAcknowledge in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0024: Asserts through `Client.NextBundle` and `Client.AcknowledgeBundle` that a completed run's result is returned in a canonical bundle with the same bytes on every call, also after a later result, until acknowledgement; that acknowledgement deletes the bundle immediately, leaving its identifier absent from the record store; that a repeated, empty or unrelated identifier returns `bundle_unavailable`; and that once every bundle is acknowledged no pending record remains and the next call returns the empty string.

<a id="UT.root.TestClearDeletesEveryStateClass"></a>

### UT.root.TestClearDeletesEveryStateClass

- Case: `UT.root.TestClearDeletesEveryStateClass`
- Result: passed
- Test: `TestClearDeletesEveryStateClass` in `clear_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestClearDeletesEveryStateClass$' .`
- Scope: Existing synthetic Go unit test TestClearDeletesEveryStateClass in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LINKAGE-0023: Asserts through `Client.ClearMeasurementData`, with the production `persistence.State` and pending store over the test-only record store, that after a completed run leaves one replay record, one run record and one pending bundle, clearing succeeds, leaves no record of any class and no stored byte, deletes the synthetic linkage key once and derives none (items 3 to 5 and 8), and leaves the client uninitialized, so a reservation returns `not_initialized` until an explicit `Initialize`.
  - REQ-LINKAGE-0024: Asserts that the pending bundle is deleted rather than delivered: `NextBundle` after reinitialization returns no bundle, and neither clearing nor reinitialization derives a linkage key.

<a id="UT.internal.persistence.TestClearOrder"></a>

### UT.internal.persistence.TestClearOrder

- Case: `UT.internal.persistence.TestClearOrder`
- Result: passed
- Test: `TestClearOrder` in `internal/persistence/clear_test.go` (package `internal/persistence`)
- Reproduce one case: `go test -count=1 -run '^TestClearOrder$' ./internal/persistence`
- Scope: Existing synthetic Go unit test TestClearOrder in package internal/persistence, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LINKAGE-0023: Asserts the exact order of record-port calls made by `persistence.Clear` over a seeded scope: the marker write; the listing, deletion and confirming listing of pending, then run, then replay records; the eraser; and the listing, deletion and confirming listing of the marker class, leaving every class empty.

<a id="UT.internal.persistence.TestStateForgetAfterClear"></a>

### UT.internal.persistence.TestStateForgetAfterClear

- Case: `UT.internal.persistence.TestStateForgetAfterClear`
- Result: passed
- Test: `TestStateForgetAfterClear` in `internal/persistence/clear_test.go` (package `internal/persistence`)
- Reproduce one case: `go test -count=1 -run '^TestStateForgetAfterClear$' ./internal/persistence`
- Scope: Existing synthetic Go unit test TestStateForgetAfterClear in package internal/persistence, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0025: Asserts that after `persistence.Clear` and `Forget` on both views of the production `State`, a cleared generation-7 replay record no longer counts toward `HighestGeneration`, which returns zero, or toward a capacity of one, so another key commits as new.
