# Go Unit-Test Evidence at 13c5251

This Phase 3 M3 milestone update moves both manifests to this baseline and to public `da7517c`, which allocates REQ-OBS-0052 and REQ-TRUST-0060 in the residual source review. It records the first execution of the PRECONDITIONS, MINIMIZATION, INTERPRETATION and SAFETY families of all nine measurements, the four Go-owned result-linkage families, and four execution, context and schedule families whose rows were completed by the platform suspension signal and the trusted local restriction inputs. Each case section below is the conformance evidence for one manifest case: 108 measurement family cases and 94 measurement `UT.*` cases, and 64 cross-cutting family cases and 296 cross-cutting `UT.*` cases. 34 `UT.*` cases belong to both manifests and have one section, listing the aspects each manifest links.

## Execution Record

- Implementation: `https://github.com/tv-rain/barometer-go` commit `13c525135cdd4614c7ba26b3b4e43e168fefc932` (the merge of barometer-go PR #134 on `main`), exported with `git archive` into an empty directory outside any checkout.
- Specification baseline: `https://github.com/tv-rain/barometer-spec` commit `da7517c07af7008f432b8e7bf4c4db41b87bd696` (Draft `0.1.0`, public PR #49 merge), which allocates REQ-OBS-0052 and REQ-TRUST-0060, adds no measurement unit and retires no requirement; it contains everything in the previous baseline `45bc424`.
- Environment: go1.26.0 darwin/arm64 (`GOTOOLCHAIN=local`, `GOFLAGS` empty, `CGO_ENABLED=1`), default test timeout, no race detector.
- Command, from the export root: `go test -count=1 -json ./...`
- Outcome: exit status 0; all 758 top-level tests passed, none failed or skipped. Each case below names its passing top-level tests; a top-level pass includes all of its subtests. Fuzz targets ran their seed corpora only.
- Executed on November 13, 2026. Inputs are the synthetic, in-repository fixtures at the implementation commit: public IP literals and synthetic names that test transports map to loopback servers, the test-only scripted HTTP/1.1 CONNECT proxy, loopback DNS, TLS, HTTPS and DNS-over-TLS servers with certificates from a synthetic certificate authority generated for each run, an injected system resolver, a fake platform network, access, suspension and device state, synthetic local policies, test linkage key stores holding synthetic secrets, recording sinks, synthetic canary values, synthetic native adapters for the gomobile facade and deterministic clocks. No public destination is contacted, and no deployment key, endpoint, proxy, participant, measurement or private data is involved.

## Carried-Forward Cases

Every case of the [5bc1335 report](go-unit-5bc1335.md) was re-executed at this baseline with its recorded aspects, except for the changes listed here. The commits since then add the reviewed Android trust-root set, the platform suspension signal on the Network port, trusted local restriction inputs (release-reviewed one-time tasks, device constraints and Network Context eligibility), the shared minimization, interpretation, preconditions and safety harnesses with their family suites, the Go-owned linkage family suites and the residual allocation references. They add 66 top-level tests, of which 45 are cited by the new cases below; the other 21, mostly tests of the trust-root loader, the trust-document decoder, the platform network port, local restriction inputs and the facade, passed and are not mapped to manifest cases at this baseline.

Carried tests whose assertions changed are recorded with their current aspects: `TestCcScheduleSchedulingModelTimeLimit`, `TestGateIssuerAndLocalCeilings` and `TestGateStepClasses` admit the embedded one-time task without `not_after` only under a local policy that lists it as release-reviewed by issuer and exact-bytes digest, where they earlier relied on every embedded source carrying the exception; `TestCcLimitScheduleAndRetryLimits` adds the release-review rows through `Client` (a listed task is admitted, an unlisted one, other bytes of the listed task, a listed task with two runs and a remote package of the listed bytes are `task_schema_invalid`, and without a listing no embedded task has the exception); and `TestCcContextContextLayersAccessContext` expects the platform access state to carry the local `Suspended` field, which still never reaches the access object. `TestMNet01Admission`, `TestMNet05CodecBoundsAndComposition`, `TestMNet05PrimitiveRiskAndMissingCapability` and `TestProductionRegistryContainsOnlyDeliveredMeasurements` build the production registry with the trust-root flag that now gates the TLS-dependent tests, with no change to their assertions. The changed rows fall within each test's recorded aspects except where noted in its section.

## New Cases

- **Measurement family cases:** the PRECONDITIONS, MINIMIZATION, INTERPRETATION and SAFETY families of NET-01, NET-03, NET-05, DNS-01, DNS-03, DNS-04, TCP-01, TLS-01 and HTTP-01, 36 cases, executed as root-package suites through the production client on the shared preconditions, minimization, interpretation and safety harnesses. Their manifest locations move from the planned `internal/measurement`, `internal/result` and `internal/admission` paths to the root files that hold them. All 108 measurement family cases are now passed.
- **Cross-cutting family cases:** `CC-EXECUTION-FIELDS-NETWORK-FIELDS-NETWORK-CONTEXT-REQUIREMENTS`, `CC-CONTEXT-EXECUTION-INTEGRATION`, `CC-SCHEDULE-ENVELOPE` and `CC-EXECUTION-FIELDS-DEVICE-FIELDS` at their existing locations, and `CC-LINKAGE-GOALS-AND-NON-GOALS`, `CC-LINKAGE-INSTALLATION-SECRET`, `CC-LINKAGE-STORAGE-FAILURE` and `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`, which move to `linkage_conformance_test.go`. The four native linkage families stay planned; their rows map to plugin evidence in `docs/conformance/native-linkage.md`.
- **Cross-cutting unit-test cases:** `UT.internal.admission.TestGateReleaseReviewedOneTimeException`, linked to REQ-LIMIT-0013, and `UT.root.TestLocalPolicyContextEligibility`, linked to REQ-EXECUTION-0022 to REQ-EXECUTION-0024.
- **Newly allocated requirements:** REQ-OBS-0052 is linked to `UT.root.TestRuntimeTLSNegotiatedVocabularies`, `UT.root.TestRuntimeTLSRejectsUnofferedALPN`, `UT.root.TestMTls01Normalization` and `UT.internal.model.observation.TestAcceptedProducerMappings`, and REQ-TRUST-0060 to `UT.internal.model.task.TestPolicyCarriesBoundedTrustedKeys` and `UT.internal.model.task.TestPolicyRejectsMalformedOrAmbiguousKeys`; their sections below add one aspect each. Their family cases, `CC-OBS-VOCAB-TLS-RETAINED-ALPN` and `CC-TRUST-KEY-POLICY-INTEGRITY`, are planned.

## Limits

Every case is partial evidence. A passing case shows only the stated aspects of each linked requirement; the manifest notes record the remaining obligations. Destinations are public literals mapped to loopback inside test transports, the trust store is a synthetic certificate authority, the system resolver and platform access, suspension and device state are injected, and the linkage key stores are test stores, so no native resolver, trust store, network-state API, captive-portal signal, battery or charging reading, key protection, routing, suspension, IPv6 reachability or proxy behavior is shown; the Android and iOS plugins record their native evidence separately. The facade tests run the Go package with synthetic native adapters; no gomobile binding, native build or device is involved. Overlap checks for equivalent tuples wait a fixed 20 ms, so on a slow host they can only miss an overlap, never report a false one. The minimization cases search the surfaces the library exposes (results, bundles, the pending store's bytes and returned errors) for synthetic canaries; they cannot show that an item absent from every fixture is never retained. The interpretation cases scan production comments, string literals, examples and documentation for prohibited claims and required qualifiers; they do not show how a reader interprets a result. The safety cases probe effective risk, consent and the deployment's authority with synthetic policies; deployment review, consent user interfaces and field risk are not involved. No `RV-*` review case, platform or device validation or field measurement is involved. DNS, TCP, TLS, HTTP and network-state observations, NET-03 validation results and NET-05 profiles are not evidence of censorship, and linkage tokens are pseudonymous, not anonymous. NET-05 validation context is only the platform's own state, which can be wrong, stale or unavailable, and Barometer does not detect an issuer that reuses or lowers a `profile_version`. An admission rejection is not evidence of interference, and admission does not show that a NET-05 run fits. These reports establish no implemented requirement, released API, platform or field support.

## Measurement Conformance Family Cases

<a id="M-TCP-01-ADMISSION"></a>

### M-TCP-01-ADMISSION

- Case: `M-TCP-01-ADMISSION`
- Result: passed
- Tests: `TestMTcp01Admission` in `tcp01_conformance_test.go` (package `root`); `TestTcp01StrictSchemaBothOrigins` in `tcp01_conformance_test.go` (package `root`); `FuzzTcp01Admission` in `tcp01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTcp01Admission$' ./.`
- Scope: Vary canonical public IPv4/IPv6, ports 1/65535 and outside bounds, exact local registration, prohibited/special-purpose ranges, hostnames, arrays, port ranges, interpolation and unknown fields. Preserve REQ-TCP-01-0001: rejection starts no socket and produces no network-interference Measurement Result.
- Requirement aspects exercised:

  - REQ-TCP-01-0001: Every rejected input, through Client.AdmitTask, returns no handle, writes no replay record, starts no run and makes no transport call, so no Measurement Result exists; FuzzTcp01Admission checks the same property on its seed corpus.
  - REQ-TCP-01-0005: Admits exactly one canonical public IPv4 or IPv6 literal and one port only when they match a local registration; an unregistered port is target_not_registered with no contact.
  - REQ-TCP-01-0010: Accepts ports 1 and 65535 with IPv4 and IPv6 literals and rejects 0, 65536, negative, fractional, exponent, string, null and over-precision ports and a missing port as task_schema_invalid.
  - REQ-TCP-01-0011: Rejects a hostname in ip_address as task_schema_invalid before any contact.
  - REQ-TCP-01-0012: Rejects unspecified, private, shared, loopback, link-local, documentation, benchmarking, multicast, reserved, IPv4-mapped and IPv6 special-purpose literals, non-canonical spellings and zoned addresses as task_schema_invalid.
  - REQ-TCP-01-0013: Rejects an array address, an array port, a port range string and an interpolation placeholder.
  - REQ-TCP-01-0014: Rejects unknown members, duplicate members and escaped-equivalent duplicate names for embedded and signed remote delivery (TestTcp01StrictSchemaBothOrigins).

<a id="M-TCP-01-OPTIONS"></a>

### M-TCP-01-OPTIONS

- Case: `M-TCP-01-OPTIONS`
- Result: passed
- Tests: `TestMTcp01Options` in `tcp01_conformance_test.go` (package `root`); `TestMTcp01AddressFamilies` in `tcp01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTcp01Options$' ./.`
- Scope: Accept empty options only; reject extras and wrong types. Derive the family solely from the validated explicit IP address.
- Requirement aspects exercised:

  - REQ-TCP-01-0015: Admits only an empty options object; any member is rejected, and the original encoded input and options byte caps are checked at limit -1, limit and limit +1.
  - REQ-TCP-01-0016: Rejects additional option properties and wrong option types; the recorded address_family follows the IP literal for IPv4 and IPv6 (TestMTcp01AddressFamilies).

<a id="M-TCP-01-PRECONDITIONS"></a>

### M-TCP-01-PRECONDITIONS

- Case: `M-TCP-01-PRECONDITIONS`
- Result: passed
- Tests: `TestMTcp01Preconditions` in `tcp01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTcp01Preconditions$' ./.`
- Scope: Vary family support, registered address/port approval, stable generation/deadline and intended path capability. Reject before contact when routing would use an unintended path, including stricter host/task policy. Portable gate tests plus separately required native/host validation; no platform claim.
- Requirement aspects exercised:

  - REQ-TCP-01-0017: With the platform reporting IPv4 availability `no` or `unknown` for an IPv4 target, the run ends `incomplete` with `constraints_unsatisfied` and no dial.
  - REQ-TCP-01-0018: Tasks whose port (`444`) or address (`8.8.4.4`) is not registered are rejected at admission with no dial; a withheld registration gives `target_not_registered` and a registration lapsed after admission `local_policy_rejected`, both with no dial. Through the `TestMTcp01Termination` subtest, a local address policy that denies the destination ends the run `incomplete` with `local_policy_rejected` and no dial.
  - REQ-TCP-01-0019: Through the shared gates, an unreadable network access state at reservation defers the run (`run_not_due`, no result), and a network change after reservation to a transport the task does not allow gives `not_run` with `constraints_unsatisfied`, both with zero contacts. Through the `TestMTcp01Termination` subtest, a network generation change during the dial gives `incomplete` with `network_changed`, a host cancel gives `incomplete` with `cancelled_by_host`, and a generation change during close after the outcome leaves `completed` with `connected`. Execution deadline availability is not asserted.
  - REQ-TCP-01-0020: Only the VPN and proxy rows of the shared gates are asserted: a task excluding VPN on a VPN-detected network is deferred with `run_not_due`, and a proxy-path task without a proxy never contacts directly. Routing through an unintended path by platform policy is not asserted; native row, see docs/measurement-requirements.md.
  - REQ-TCP-01-0021: The shared portable gates assert, each with zero contacts: a task transport condition the network does not meet, a task excluding VPN on a VPN-detected network and a local charging-only device restriction defer the run at `StartRun` (`run_not_due`, no result); a transport condition lost after reservation gives `not_run` with `constraints_unsatisfied`; and a proxy-path task with no proxy configured is either rejected at admission with `task_schema_invalid` or ends `not_run` with `proxy_not_configured` (the test accepts either); a deployment that withholds the target registration rejects the task with `target_not_registered`, and a registration that lapses after admission gives a result with reason `local_policy_rejected`.

<a id="M-TCP-01-EXECUTION"></a>

### M-TCP-01-EXECUTION

- Case: `M-TCP-01-EXECUTION`
- Result: passed
- Tests: `TestMTcp01ExecutionAndOutput` in `tcp01_conformance_test.go` (package `root`); `TestMTcp01ConnectionBudget` in `tcp01_conformance_test.go` (package `root`); `TestTcp01ConnectionMeter` in `internal/execution/tcp01_termination_test.go` (package `internal/execution`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTcp01ExecutionAndOutput$' ./.`
- Scope: Record one socket connect to the exact accepted address/port and immediate close without application bytes. No DNS lookup, alternate address/port, retry or additional protocol may run.
- Requirement aspects exercised:

  - REQ-TCP-01-0006: On the direct path and through the scripted CONNECT proxy, the loopback server accepts exactly one connection, receives and sends zero application bytes and sees it closed; the transport is dialed once for the exact checked authority.
  - REQ-TCP-01-0007: The production runner checks the target, confirms the network generation and family, issues one metered connect and closes it after recording the outcome; the meter counts exactly one connection and no bytes (TestTcp01ConnectionMeter).

<a id="M-TCP-01-NORMALIZATION"></a>

### M-TCP-01-NORMALIZATION

- Case: `M-TCP-01-NORMALIZATION`
- Result: passed
- Tests: `TestMTcp01Normalization` in `tcp01_conformance_test.go` (package `root`); `TestMTcp01Limits` in `tcp01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTcp01Normalization$' ./.`
- Scope: Exercise every failure/outcome row, including refusal/reset/timeout/unreachable/permission and coarse inconclusive errors, using deterministic socket fixtures. Distinguish socket outcomes from common interruption and map to exactly one tcp_connect.
- Requirement aspects exercised:

  - REQ-TCP-01-0008: A completed connect to the checked authority yields outcome connected with one succeeded tcp_connect observation.
  - REQ-TCP-01-0009: Wrapped typed socket errors map to timeout, connection_refused, connection_reset, network_unreachable, address_unreachable, permission_denied and socket_error, and an untyped error to inconclusive, without retaining raw error text or numbers.
  - REQ-TCP-01-0025: Each failure maps to the outcome of the same name with one failed tcp_connect, and the 15-second operation bound yields completed with outcome timeout rather than an interruption (TestMTcp01Limits).

<a id="M-TCP-01-LIMITS"></a>

### M-TCP-01-LIMITS

- Case: `M-TCP-01-LIMITS`
- Result: passed
- Tests: `TestMTcp01Limits` in `tcp01_conformance_test.go` (package `root`); `TestMTcp01ConnectionBudget` in `tcp01_conformance_test.go` (package `root`); `TestMTcp01TupleConcurrency` in `tcp01_conformance_test.go` (package `root`); `TestMTcp01EquivalentRunsSerialize` in `tcp01_conformance_test.go` (package `root`); `TestTcp01ConnectionMeter` in `internal/execution/tcp01_termination_test.go` (package `internal/execution`); `TestTcp01ReleasesSocketOnEveryPath` in `internal/execution/tcp01_termination_test.go` (package `internal/execution`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTcp01Limits$' ./.`
- Scope: Advance the operation and outer deadlines independently; assert one connection, handshake traffic only, no retries and deduplication by address/port tuple. Verify socket release on all failure/cancellation paths.
- Requirement aspects exercised:

  - REQ-TCP-01-0022: The operation bound and the outer attempt deadline are advanced independently; max_connections 0 refuses the dial before the transport with resource_limit, and at 1 or 2 exactly one connection is metered with no application bytes; equivalent address/port tuples wait without contact and run after release; a socket returned after host cancellation, the attempt deadline or the operation bound is closed exactly once, with no second dial.

<a id="M-TCP-01-TERMINATION"></a>

### M-TCP-01-TERMINATION

- Case: `M-TCP-01-TERMINATION`
- Result: passed
- Tests: `TestMTcp01Termination` in `tcp01_conformance_test.go` (package `root`); `TestMTcp01ProxyTermination` in `tcp01_conformance_test.go` (package `root`); `TestTcp01AllTerminationBoundaries` in `internal/execution/tcp01_termination_test.go` (package `internal/execution`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTcp01Termination$' ./.`
- Scope: Exercise a valid synthetic TCP-01 task against a scripted transport, fake platform and deterministic clock, and interrupt at each step of the TCP-01 Procedure section: during address/port validation, during the stable network-generation and address-family check, after the connect attempt starts but before the socket reports an outcome, after a success or socket outcome is recorded but before close, and during the immediate close. At each point inject separately a network-generation change, host cancellation, expiry of the outer attempt deadline, platform suspension, and (where a proxy path is configured) proxy failure or proxy reconfiguration. Expected states follow the GO-SPEC-ISSUE-004 disposition: assert an interruption during address/port validation is not_run with exactly that reason, no measurement and no network_context; at the network-generation and address-family check, which starts the test, and while the connect is pending, assert execution.state is incomplete with exactly the first applicable reason (network_changed, cancelled_by_host, attempt_deadline_exceeded, platform_suspended, or the matching proxy reason such as proxy_unavailable, proxy_connect_failed or proxy_reconfigured), that no measurement object and no partial tcp_connect observation is emitted (TCP-01 is not NET-05); after a success or socket outcome is recorded and during the immediate close, assert execution.state stays completed with the recorded outcome unchanged, and that network_changed never appears as a measurement outcome or tcp_connect failure. Contrast with expiry of the operation deadline inside the connect step, which must yield state completed, outcome timeout and one failed tcp_connect with failure timeout. Verify on every path that the socket is released, no application bytes are written, and no retry, alternate address/port or invented terminal TCP outcome follows the interruption. Requirements: REQ-TCP-01-0026.
- Requirement aspects exercised:

  - REQ-TCP-01-0026: For host cancellation, network change, attempt deadline, injected suspension and the three proxy reasons: an interruption before the test starts is not_run with no network_context; at the generation and family check and during the connect it is incomplete with the first reason, no measurement and no dial after the interruption; during the close after the outcome it stays completed with outcome connected. The scripted CONNECT proxy produces proxy_connect_failed, proxy_reconfigured, cancelled_by_host and proxy_unavailable without a TCP outcome.

<a id="M-TCP-01-MINIMIZATION"></a>

### M-TCP-01-MINIMIZATION

- Case: `M-TCP-01-MINIMIZATION`
- Result: passed
- Tests: `TestMTcp01Minimization` in `tcp01_conformance_test.go` (package `root`), `TestMTcp01Safety` in `measurement_safety_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTcp01Minimization$' ./.`
- Scope: Exercise a synthetic accepted TCP-01 target through a scripted transport and fake platform that surface one unique canary per item prohibited by REQ-TCP-01-0027: a socket error string (CANARY-SOCKERR-TEXT), a socket error number (91337), the destination address (the synthetic configured address string), the destination port (synthetic port 47391), a local address (CANARY-LOCAL-ADDR), a local port (51873), TCP flags (CANARY-TCP-FLAGS), packet counts (73191), payload data (CANARY-PAYLOAD-BYTES) and retry metadata (CANARY-RETRY-META). Iterate the connected outcome and every failed outcome (timeout, connection_refused, connection_reset, network_unreachable, address_unreachable, permission_denied, socket_error, inconclusive) plus the incomplete error path. Assert the serialized result, error values and all nested observations contain none of the canaries and no field outside the allowed set; the measurement retains only outcome and one tcp_connect observation with type, status, duration_ms as a non-negative multiple of 250, failure only when failed, and details containing only address_family (ipv4 or ipv6). Assert that an injected extra detail or measurement field is rejected rather than dropped. No participant or live measurement data is used. Requirements: REQ-TCP-01-0027.
- Requirement aspects exercised:

  - REQ-TCP-01-0027: Unique canaries are placed in the destination address and port (task input and registration) and, through the scripted transport, in the socket error text and number, local address and port, TCP flags, packet count, payload and retry metadata carried by the returned error and the connected socket. Over `connected`, `timeout`, `connection_refused`, `connection_reset`, `network_unreachable`, `address_unreachable`, `permission_denied`, `socket_error`, `inconclusive` and the `incomplete` runs ended by a network change and by host cancellation, no canary appears in any searched encoding in any result, the bundle, the pending store's bytes, any returned error or the task-validation rejection of a task whose unknown input member carries the canaries; each outcome's measurement holds only `outcome` and one `tcp_connect` observation with `type`, `status`, `duration_ms` (a non-negative multiple of 250), `failure` and `details.address_family`, interrupted runs keep no measurement, and the strict decoder rejects an injected `port`, `local_address`, `error`, `retries`, `destination` or `payload` member and an unquantized `duration_ms`.

<a id="M-TCP-01-INTERPRETATION"></a>

### M-TCP-01-INTERPRETATION

- Case: `M-TCP-01-INTERPRETATION`
- Result: passed
- Tests: `TestMTcp01Interpretation` in `tcp01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTcp01Interpretation$' ./.`
- Scope: Assert the serialized TCP-01 result schema, the Go field descriptions and documentation strings, and the sanitized example and host-facing text for every outcome (connected, timeout, connection_refused, connection_reset, network_unreachable, address_unreachable, permission_denied, socket_error, inconclusive) contain none of these prohibited claims: that DNS resolution was performed or measured; that TLS, HTTP or another application protocol is usable or unusable; that a timeout or reset establishes intentional filtering; that a single connection result is definitive proof of censorship; or that outcome is a generic censorship, blocking, anomaly or success verdict. Assert no interpretation, verdict, censored, blocked or filtered field exists and that injecting such a field into a measurement is rejected as unknown. Assert text describing a result stays within one execution's claim that a TCP connection to the configured address and port completed or produced a particular observable socket outcome. Requirements: REQ-TCP-01-0002, REQ-TCP-01-0003, REQ-TCP-01-0004, REQ-TCP-01-0028, REQ-TCP-01-0029, REQ-TCP-01-0030. Judgement-based completion is the separate review case RV-M-TCP-01-INTERPRETATION; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-TCP-01-0002: No non-negated sentence in the comments and string literals of production Go files and sanitized examples and the shipped Markdown documentation (case plans, manifests, evidence, changelog and specification-issue records excluded) matches the "TCP-01 performed or measured DNS" claim pattern, and the TCP-01 section of the interpretation guidance contains "performs no dns resolution". This is a textual scan of shipped text; it does not prove how a reader interprets a result.
  - REQ-TCP-01-0003: The same scan finds no non-negated sentence stating that TCP-01 shows TLS, HTTP or an application to be usable or available; no guidance qualifier is required for this item.
  - REQ-TCP-01-0004: The same scan finds no non-negated sentence stating that a timeout, reset or refusal establishes, indicates, shows, proves or means filtering, blocking or censorship, the guidance contains "does not establish intentional filtering", and no member name or string value of any run's measurement contains a verdict token (such as `verdict`, `censor`, `block`, `filter`, `anomal`, `confiden`) and the strict result decoder rejects `verdict`, `interpretation`, `censored`, `blocked`, `filtered`, `anomaly`, `confidence`, `reachable` and `authoritative` fields injected into a valid result.
  - REQ-TCP-01-0028: The guidance contains "establishes only whether a tcp connection to the configured address and port completed or produced a particular observable socket outcome", and the nine runs, one per registered outcome, each record that outcome with a single `tcp_connect` observation.
  - REQ-TCP-01-0029: Only the fourth item is asserted: the guidance contains "independent measurements from comparable times and network contexts"; repeated attempts, control endpoints and related DNS, TLS and HTTPS observations are not checked.
  - REQ-TCP-01-0030: No non-negated sentence of the scanned text matches the common claims (a result is or provides definitive or conclusive proof of censorship or blocking; a result proves, demonstrates or confirms censorship or blocking; a censorship, blocking, anomaly or success verdict), the guidance contains "is not definitive proof of censorship", and no member name or string value of any run's measurement contains a verdict token (such as `verdict`, `censor`, `block`, `filter`, `anomal`, `confiden`) and the strict result decoder rejects `verdict`, `interpretation`, `censored`, `blocked`, `filtered`, `anomaly`, `confidence`, `reachable` and `authoritative` fields injected into a valid result.

<a id="M-TCP-01-SAFETY"></a>

### M-TCP-01-SAFETY

- Case: `M-TCP-01-SAFETY`
- Result: passed
- Tests: `TestMTcp01Safety` in `measurement_safety_conformance_test.go` (package `root`); `TestMTcp01Minimization` in `tcp01_conformance_test.go` (package `root`); `TestMTcp01EquivalentRunsSerialize` in `tcp01_conformance_test.go` (package `root`); `TestMTcp01TupleConcurrency` in `tcp01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTcp01Safety$' ./.`
- Scope: Iterate each Baseline Risk row and each Risk Factors row with synthetic tasks, a fake local policy and host, and a recording transport. Baseline: assert the registered risk level is low and requires_user_consent false is evaluated subject to host and local policy, so a fake host that requires consent blocks execution without valid consent. Factors network_observability, traffic_distinctiveness, target_sensitivity, data_sensitivity and legal_or_social_context: assert each registered low level is kept, a task declaring medium, high or unknown raises the effective risk accordingly, and effective risk is never below the registered baseline. Exercise each normative safety action: data minimization (serialized output holds only outcome, address family and quantized duration); traffic distinctiveness (recording transport sees one connect per run, port arrays or ranges rejected, equivalent concurrent runs for one address/port tuple deduplicated or delayed, no retry); user-identification risk (zero application bytes written and no identifier such as run_id in traffic); target risk (reject loopback, unspecified, multicast, link-local, private, documentation, benchmarking and other special-purpose synthetic addresses, hostnames, and targets the fake policy marks unreviewed or sensitive, with no socket opened and no measurement); and probe final authority (fake local policy delaying, skipping or rejecting a valid task prevails). Requirements: REQ-TCP-01-0031, REQ-TCP-01-0032, REQ-TCP-01-0033, REQ-TCP-01-0034, REQ-TCP-01-0035, REQ-TCP-01-0036, REQ-TCP-01-0037. Judgement-based completion is the separate review case RV-M-TCP-01-SAFETY; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-TCP-01-0031: The test asserts that the production registry records a `low` baseline for TCP-01 and that a task declaring `low`, `medium`, `high` or `unknown` has effective risk exactly that level (admitted under issuer and local bounds permitting it, rejected `task_issuer_not_authorized` with no contact when either side permits only lower levels). For the consent baseline it asserts that `false` adds no requirement (with host consent absent the run starts, yields one result and never queries the consent capability), while consent required by the task, by an issuer or by the local policy blocks every result and contact unless the fake host reports valid consent (absent, ended and unavailable all block).
  - REQ-TCP-01-0032: For each of the five factor types the test asserts that the registered baseline equals the most restrictive of the fixture's row levels (all five `low`), that declaring the factor at `low`, `medium` or `high` gives effective risk exactly the maximum of baseline, row level and declared level, that an `unknown` factor gives `unknown`, and that an unknown factor under a known overall level or a factor above the overall level is rejected `task_schema_invalid` and an unrecognized factor type `task_unsupported`, all with no contact. The row levels are fixture constants checked only against the registered baseline, and the rationale text is not asserted.
  - REQ-TCP-01-0033: For the connected outcome, eight failure classes and two interrupted runs, the test asserts that results, the pending store, the bundle and error text contain none of ten transport canaries (destination and local address and port, socket error text and number, TCP flags, packet count, payload, retry metadata), that the measurement holds only outcome, observation type, status, `duration_ms`, failure and `details.address_family`, and that interrupted runs retain no measurement. The strict result decoder rejects injected port, local address, error text, retry count, destination, payload and an unquantized `duration_ms` of 375.
  - REQ-TCP-01-0034: The test asserts exactly one connect per run (a refused connect is not retried), that port arrays, port ranges and a `ports` member are rejected `task_schema_invalid` with no connect, and that a second equivalent run to the same address and port does not dial while the first holds the tuple (it dials only after release, or is cancelled without a dial, and the tuple is released afterwards). Synchronized schedules are not asserted beyond the absence of self-scheduled repeats.
  - REQ-TCP-01-0035: With a connected recording socket the test asserts zero application bytes written and zero reads. Identifiers are covered only by that zero-byte check; there is no separate assertion on identifier content.
  - REQ-TCP-01-0036: The test asserts that loopback, unspecified, multicast, link-local, private, documentation and benchmarking addresses (IPv4 and IPv6) registered in place of the reviewed target produce no contact and no completed result, and that an unreviewed address, an unreviewed port and a hostname are rejected at admission with no connect. For repeated execution it asserts that one admitted run followed by a 48-hour clock advance makes exactly one connect and one result; the residual-risk statement is not asserted.
  - REQ-TCP-01-0037: The test asserts that a local policy not permitting the task's network path rejects a valid task at admission with no contact, and that a local charging-only device constraint holds an admitted valid task (`run_not_due` twice, no result, no contact) until satisfied, after which the same task runs; a target registration that expires after admission ends the attempt `local_policy_rejected` with no contact. The no-lowering clause is asserted through the declared-level rows: no task declaration yields effective risk below the `low` baseline.

<a id="M-TCP-01-OUTPUT"></a>

### M-TCP-01-OUTPUT

- Case: `M-TCP-01-OUTPUT`
- Result: passed
- Tests: `TestMTcp01ExecutionAndOutput` in `tcp01_conformance_test.go` (package `root`); `TestMTcp01Normalization` in `tcp01_conformance_test.go` (package `root`); `TestMTcp01AddressFamilies` in `tcp01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTcp01ExecutionAndOutput$' ./.`
- Scope: Require exactly one tcp_connect with the selected family, normalized outcome and quantized duration. Reuse shared TCP observation validation; resolve the target through task identity rather than duplicating destination fields.
- Requirement aspects exercised:

  - REQ-TCP-01-0024: Each completed result carries exactly one tcp_connect with the selected family and a 250-ms quantized duration, repeats neither ip_address nor port, and passes production result and bundle validation.

<a id="M-TCP-01-VERSIONS"></a>

### M-TCP-01-VERSIONS

- Case: `M-TCP-01-VERSIONS`
- Result: passed
- Tests: `TestMTcp01Versions` in `tcp01_conformance_test.go` (package `root`); `TestProductionRegistryContainsOnlyDeliveredMeasurements` in `registry_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTcp01Versions$' ./.`
- Scope: Exercise result construction and validation for test.id TCP-01 with test.version 0.1.0, schema_version 0.1.0 and test.result_version 0.1.0, and assert this exact registered combination is the only one emitted and accepted. Iterate rejected variants: schema_version 0.2.0, 1.0.0, 0.1, v0.1.0, empty string and numeric 0.1; test.version 0.1.1 and 0.2.0; test.result_version 0.2.0, 0.1 and missing; cross-test combinations such as test.id DNS-01 or TLS-01 carrying the TCP-01 measurement, and test.id TCP-01 carrying a dns_query or tls_handshake observation instead of tcp_connect; and an unknown field in the test object. Assert an unsupported or malformed version selected for execution is rejected before any socket is opened, and that a received result with an unsupported version is never interpreted best-effort, coerced to 0.1.0, or analyzed after discarding unknown fields (the strict decoder rejects it; retaining an opaque bounded object in a quarantine path is a receiver option in the public text, not a library obligation). Assert the Draft status and registration metadata alone do not add any supported version. Requirements: REQ-TCP-01-0023.
- Requirement aspects exercised:

  - REQ-TCP-01-0023: Task test_version 0.1.1, 0.2.0, 1.0.0, empty, v0.1.0, 0.1, numeric and null are rejected before any dial; the decoder accepts only schema_version, test.version and test.result_version 0.1.0 and rejects 0.1.1, 0.2.0, 1.0.0, 0.1, v0.1.0, empty, numeric and null values, a missing result_version, other test IDs, an extra test member and other observation types.

<a id="M-DNS-01-ADMISSION"></a>

### M-DNS-01-ADMISSION

- Case: `M-DNS-01-ADMISSION`
- Result: passed
- Tests: `TestMDns01Admission` in `dns01_conformance_test.go` (package `root`); `TestDns01NormalizedNamesAndResolverBoundaries` in `dns01_conformance_test.go` (package `root`); `TestDns01StrictSchemaBothOrigins` in `dns01_conformance_test.go` (package `root`); `TestDns01RechecksBothRegistrationExpiries` in `dns01_conformance_test.go` (package `root`); `TestDns01CNAMEAnswersCauseNoContact` in `dns01_conformance_test.go` (package `root`); `TestCodecNormalizedNameBoundaries` in `internal/measurement/dns01/codec_test.go` (package `internal/measurement/dns01`); `TestCodecFieldTypesAndModeRelations` in `internal/measurement/dns01/codec_test.go` (package `internal/measurement/dns01`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns01Admission$' ./.`
- Scope: Vary IDNA-normalized hostname boundaries through 253 ASCII bytes, A/AAAA enums, local/root/single-label/private names and unknown fields. Independently test resolver presence by mode, canonical public address/port syntax, default port and exact local registration. Invalid input never resolves or emits a network-interference result; observed CNAMEs never cause application connections.
- Requirement aspects exercised:

  - REQ-DNS-01-0005: Admits one hostname only when it matches a local query-name registration; another name is target_not_registered with no contact.
  - REQ-DNS-01-0006: System mode needs only the query-name registration; direct mode needs one registered resolver address and port, and another resolver address or port is target_not_registered.
  - REQ-DNS-01-0014: Rejects missing, null, wrong-type and other-enum hostname, record_type and resolver members, and admits A and AAAA.
  - REQ-DNS-01-0015: Admits resolver ports 1 and 65535 for IPv4 and IPv6 resolvers, applies the default port 53 when the port is omitted, and rejects ports 0 and 65536.
  - REQ-DNS-01-0016: Normalizes upper-case and internationalized names deterministically (bücher.example.org to xn--bcher-kva.example.org), admits 252- and 253-byte names and rejects 254 bytes, invalid punycode, joiners, spaces and URL-like names.
  - REQ-DNS-01-0017: Rejects root, single-label, .local, .localhost, home.arpa, .internal, IP-literal, URL, wildcard, empty-label, hyphen-edge, over-long-label, trailing-dot and joiner names with no handle, replay record, contact or result.
  - REQ-DNS-01-0018: The query name must match its registration at admission and again at execution: an expired registration stops the attempt before contact (TestDns01RechecksBothRegistrationExpiries).
  - REQ-DNS-01-0019: A resolver is rejected in system mode and required in direct mode.
  - REQ-DNS-01-0020: A direct resolver must match a registered resolver tuple at admission and its expiry is re-checked before contact.
  - REQ-DNS-01-0021: Rejects loopback, unspecified, private, shared, link-local, documentation, benchmarking, multicast, reserved, IPv4-mapped and IPv6 special-purpose resolver literals, non-canonical spellings, zoned addresses and a hostname, with no contact.
  - REQ-DNS-01-0022: A CNAME chain in a UDP or TCP answer is retained as observed data while exactly one query and one dial are made (TestDns01CNAMEAnswersCauseNoContact).
  - REQ-DNS-01-0023: Rejects unknown input members, including duplicate and escaped-equivalent names, for embedded and signed remote delivery.
  - REQ-DNS-01-0024: Every rejected input returns no handle, writes no replay record, makes no transport call and produces no Measurement Result.

<a id="M-DNS-01-OPTIONS"></a>

### M-DNS-01-OPTIONS

- Case: `M-DNS-01-OPTIONS`
- Result: passed
- Tests: `TestMDns01Options` in `dns01_conformance_test.go` (package `root`); `TestCodecFieldTypesAndModeRelations` in `internal/measurement/dns01/codec_test.go` (package `internal/measurement/dns01`); `TestDns01StrictSchemaBothOrigins` in `dns01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns01Options$' ./.`
- Scope: Cover omitted/system/direct modes, direct-only UDP/TCP, collect_additional_responses and conditional 1..1,000 ms window: the shared registered maximum of 1,000 ms is stricter than the test's option table and controls admission, so 1,001 ms and above are rejected without clamping. Test every missing/null/wrong/unknown option and prohibited combination, with no fallback to another mode, resolver or transport.
- Requirement aspects exercised:

  - REQ-DNS-01-0025: Defaults to system mode and UDP; accepts the additional-response window of 1 and 1,000 ms only with collection enabled in direct UDP; rejects window 0, wrong scalar types, null, case variants, a transport or collection in system mode and collection over TCP as task_schema_invalid, and windows 1,001 and 10,001 as task_limit_exceeded.
  - REQ-DNS-01-0026: Rejects unknown option members with no contact; each admitted mode and transport is the one dialed, with no fallback (see M-DNS-01-TERMINATION for the zero-dial system rows).

<a id="M-DNS-01-PRECONDITIONS"></a>

### M-DNS-01-PRECONDITIONS

- Case: `M-DNS-01-PRECONDITIONS`
- Result: passed
- Tests: `TestMDns01Preconditions` in `dns01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns01Preconditions$' ./.`
- Scope: Vary generation stability, hostname ownership/review, resolver registration, availability of a narrow resolver API and explicit UDP window/count bounds. Reject APIs exposing unrelated DNS activity; host review and platform visibility require separate evidence. Portable gate tests plus separately required native/host validation; no platform claim.
- Requirement aspects exercised:

  - REQ-DNS-01-0027: Through the shared gates, an unreadable network access state at reservation defers the run (`run_not_due`, no result), and a network change after reservation to a transport the task does not allow gives `not_run` with `constraints_unsatisfied`, both with zero contacts (no transport dial).
  - REQ-DNS-01-0028: A task whose hostname differs from the registered name (`unreviewed.example`) is rejected at admission with no transport contact. The host review of name ownership is not asserted; host row, see docs/measurement-requirements.md.
  - REQ-DNS-01-0029: A task naming a direct resolver address that is not registered (`8.8.4.4`) is rejected at admission with no transport contact, and a deployment that withholds the registration rejects the task with `target_not_registered`.
  - REQ-DNS-01-0030: With the system engine, the platform resolver port is called exactly once, with the admitted name (`ip4 example.org.`). Whether a native binding exposes or retains unrelated DNS activity, and rejection of such an API, are not asserted; native row, see docs/measurement-requirements.md.
  - REQ-DNS-01-0031: Through the `TestMDns01CollectionOrderAndBounds` subtest, with additional-response collection enabled and a 500 ms window two responses are recorded in arrival order with offsets 0 and 250 ms; with four responses available three are recorded with `response_limit_reached`, the limit flag set, three reads and one query; with collection disabled only one response is read. A task enabling collection without explicit bounds is not exercised.
  - REQ-DNS-01-0032: The shared portable gates assert, each with zero contacts: a task transport condition the network does not meet, a task excluding VPN on a VPN-detected network and a local charging-only device restriction defer the run at `StartRun` (`run_not_due`, no result); a transport condition lost after reservation gives `not_run` with `constraints_unsatisfied`; and a proxy-path task with no proxy configured is either rejected at admission with `task_schema_invalid` or ends `not_run` with `proxy_not_configured` (the test accepts either); a deployment that withholds the target registration rejects the task with `target_not_registered`, and a registration that lapses after admission gives a result with reason `local_policy_rejected`.

<a id="M-DNS-01-EXECUTION"></a>

### M-DNS-01-EXECUTION

- Case: `M-DNS-01-EXECUTION`
- Result: passed
- Tests: `TestMDns01ExecutionAndOutput` in `dns01_conformance_test.go` (package `root`); `TestDns01SystemAPIOutputAndHiddenBytes` in `dns01_conformance_test.go` (package `root`); `TestDns01AAAAExecution` in `dns01_conformance_test.go` (package `root`); `TestMDns01CollectionOrderAndBounds` in `dns01_conformance_test.go` (package `root`); `TestDns01CNAMEAnswersCauseNoContact` in `dns01_conformance_test.go` (package `root`); `TestQueryCorrelationAndCollectionPrecedence` in `internal/protocol/dns/query_test.go` (package `internal/protocol/dns`); `TestDnsUDPSourceAndTruncation` in `internal/execution/dns01_transport_test.go` (package `internal/execution`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns01ExecutionAndOutput$' ./.`
- Scope: Record one ordinary system lookup or one direct UDP/TCP query. In UDP, vary source, transaction ID and question, unrelated datagrams, multiple matching datagrams, collection enablement and generation changes; preserve matching arrival order and coarse offsets without changing resolver configuration or contacting another name/server/transport.
- Requirement aspects exercised:

  - REQ-DNS-01-0007: System mode calls the injected ordinary resolver once and normalizes only what it returns.
  - REQ-DNS-01-0008: Direct UDP and TCP each send one query to the loopback resolver, counted by the server, with one dial; collection keeps reading matching datagrams until the window or the three-response bound.
  - REQ-DNS-01-0009: No second name, server or transport is contacted: system mode makes no direct dial, and direct mode makes one dial and one query.
  - REQ-DNS-01-0010: Responses from another source, with another ID or another question are discarded without overriding a later usable reply; matching responses keep arrival indexes 1 and 2 and coarse offsets 0 and 250 ms; with collection disabled only the first reply is read.

<a id="M-DNS-01-NORMALIZATION"></a>

### M-DNS-01-NORMALIZATION

- Case: `M-DNS-01-NORMALIZATION`
- Result: passed
- Tests: `TestMDns01Normalization` in `dns01_conformance_test.go` (package `root`); `TestMDns01CollectionOrderAndBounds` in `dns01_conformance_test.go` (package `root`); `TestDns01SystemAPIOutputAndHiddenBytes` in `dns01_conformance_test.go` (package `root`); `TestQueryCorrelationAndCollectionPrecedence` in `internal/protocol/dns/query_test.go` (package `internal/protocol/dns`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns01Normalization$' ./.`
- Scope: Exercise all failure/outcome rows and precedence with usable, unrelated, mismatching, identical and conflicting responses. Conflict outranks resolved; a count-truncated collection yields response_limit_reached as specified, with the flag. Test NXDOMAIN/empty answers without classifying them as successful address resolution and preserve opaque system error limits.
- Requirement aspects exercised:

  - REQ-DNS-01-0011: A NOERROR reply with a requested-type answer is resolved; NXDOMAIN and empty answers are failures, not resolution.
  - REQ-DNS-01-0012: Two matching responses with different answers are retained as conflicting_responses with both responses.
  - REQ-DNS-01-0013: Exercises nxdomain, no_data, servfail, refused, timeout, malformed_response, query_mismatch, unexpected_source, resolver_error and inconclusive, including opaque system resolver errors.
  - REQ-DNS-01-0037: Each outcome row maps from its observation, with response_limit_reached and the flag at three responses.
  - REQ-DNS-01-0038: Unrelated and mismatched datagrams do not override a usable reply; conflicting_responses outranks resolved, and the count bound outranks conflict.

<a id="M-DNS-01-LIMITS"></a>

### M-DNS-01-LIMITS

- Case: `M-DNS-01-LIMITS`
- Result: passed
- Tests: `TestMDns01Limits` in `dns01_conformance_test.go` (package `root`); `TestMDns01CollectionOrderAndBounds` in `dns01_conformance_test.go` (package `root`); `TestCcLimitPurposeDNSRows` in `dns01_conformance_test.go` (package `root`); `TestMDns01ConnectionBudget` in `dns01_conformance_test.go` (package `root`); `TestDns01EquivalentRunsSerialize` in `dns01_conformance_test.go` (package `root`); `TestDns01TupleComponentsDecideSerialization` in `dns01_conformance_test.go` (package `root`); `TestWireMessageBoundaries` in `internal/protocol/dns/wire_test.go` (package `internal/protocol/dns`); `TestWireAnswerAndCNAMEBounds` in `internal/protocol/dns/wire_test.go` (package `internal/protocol/dns`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns01Limits$' ./.`
- Scope: Exercise UDP 4,096-byte/TCP 65,535-byte messages, three matching responses, 16 normalized entries/eight CNAMEs and 16,384 encoded observation bytes at below/equal/above boundaries; test the additional window and outer deadline independently. Assert no test-layer retries and deduplication by the complete hostname/type/mode/resolver/generation tuple; the implementation may conservatively also serialize tuples that differ only in transport or network generation, never fewer.
- Requirement aspects exercised:

  - REQ-DNS-01-0033: The 15-second operation bound yields timeout in each mode; UDP 4,096 and TCP 65,535-byte messages are parsed at and below the limit and UDP above it is malformed; a fourth response is never read once three are retained; 15 to 17 answers and 7 to 9 CNAMEs are bounded; DNS messages and TCP length prefixes count at the metered boundary below, at and above the byte limits; one socket or connection is metered; equivalent hostname/type/mode/resolver tuples wait, while a different hostname, record type or resolver proceeds.

<a id="M-DNS-01-TERMINATION"></a>

### M-DNS-01-TERMINATION

- Case: `M-DNS-01-TERMINATION`
- Result: passed
- Tests: `TestMDns01Termination` in `dns01_conformance_test.go` (package `root`); `TestDns01AllTerminationBoundaries` in `internal/execution/dns01_termination_test.go` (package `internal/execution`); `TestMDns01Limits` in `dns01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns01Termination$' ./.`
- Scope: Drive DNS-01 with a fake platform resolver, a scripted UDP/TCP recording transport, a deterministic clock and a controllable network generation, and inject each interruption at each Procedure stage after the network generation is recorded: after step 2 (generation recorded, no query issued), during the system-mode lookup (step 3), after the direct UDP or TCP query is sent and before any response (step 4), after the first matching UDP response while response-source, transaction ID and question validation runs (step 5), inside the enabled additional-response window with one and with two matching responses already read (step 5), and during response normalization (steps 6-7). Expected states follow the GO-SPEC-ISSUE-004 disposition: every listed point follows the recorded network generation, so each is incomplete as asserted below, an interruption before the generation is recorded is not_run with that reason, no measurement and no network_context, and an interruption after the outcome is recorded, while the socket closes, stays completed with the recorded outcome. Iterate the interruptions network generation change, host cancellation, outer attempt deadline expiry, proxy failure and proxy replacement or clearing where a proxy path is configured, and platform suspension. Assert execution.state is incomplete with execution.reason respectively network_changed, cancelled_by_host, attempt_deadline_exceeded, the registered proxy reason (proxy_unavailable, proxy_connect_failed or proxy_reconfigured) and platform_suspended, that reason names only the first condition, and that no measurement object is present even when matching responses were already read (DNS-01 has no partial output, unlike NET-05). Contrast a bounded resolution timeout, which must remain a completed result with outcome timeout, against attempt_deadline_exceeded. Assert the recording transport shows no further read, no second query and no fallback to another mode, server or transport after the interruption, and that no terminal DNS outcome such as network_changed, timeout or response_limit_reached is invented for an interrupted attempt. Requirements: REQ-DNS-01-0039.
- Requirement aspects exercised:

  - REQ-DNS-01-0039: For host cancellation, network change, attempt deadline, injected suspension and the three proxy reasons over UDP and TCP: before the test starts is not_run with no network_context; at the generation check, connect, write, read, normalization, inside the collection window and in system mode it is incomplete with the first reason and no measurement, with no dial, read or write after the interruption and no direct dial in system mode; during the close after the outcome it stays completed. A bounded operation timeout remains completed with outcome timeout.

<a id="M-DNS-01-MINIMIZATION"></a>

### M-DNS-01-MINIMIZATION

- Case: `M-DNS-01-MINIMIZATION`
- Result: passed
- Tests: `TestMDns01Minimization` in `dns01_conformance_test.go` (package `root`), `TestMDns01Safety` in `measurement_safety_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns01Minimization$' ./.`
- Scope: Run system mode and direct UDP and TCP mode against a fake platform resolver and scripted recording transports, injecting one unique synthetic canary per prohibited item: a canary queried hostname approved only by the synthetic local policy, a canary direct resolver address and a canary non-default port in the fake resolver registry, a canary system resolver configuration (resolver list and search domain) exposed by the fake platform, a canary 16-bit query identifier, a canary byte pattern in the raw DNS message, canary records in the authority and additional sections, a canary TTL on every answer, canary answer ASN and organization strings offered by a fake enrichment source, a canary local socket address, and canary platform error text on a failing system lookup. Iterate success, nxdomain, servfail, malformed_response, query_mismatch, conflicting_responses and response_limit_reached paths. Assert the serialized Measurement Result, every nested dns_query observation and every error path contain none of the canaries and no field beyond engine, address_family, query_type, responses, response_limit_reached and, per response, arrival_index, arrival_offset_ms, rcode and answers with only type and value. Assert retained answers are only the bounded normalized A, AAAA and intervening CNAME values from the scripted responses, deduplicated in first-occurrence order, and that no participant or live measurement data is used. Requirements: REQ-DNS-01-0040.
- Requirement aspects exercised:

  - REQ-DNS-01-0040: Unique canaries cover the queried name, direct resolver address and port, a system resolver list and search domain carried in a scripted platform error, the query identifier, a raw-message pattern, authority and additional records and an additional address, the answer TTL, answer ASN and organization strings offered in a TXT record (the library has no enrichment source), the loopback local socket address and the platform error text. Over direct UDP and TCP `resolved`, `nxdomain`, `servfail`, `malformed_response`, `query_mismatch`, `conflicting_responses`, `response_limit_reached` and system-mode `resolved` and `resolver_error`, no canary appears in any searched encoding in any result, the bundle, the pending store's bytes, any returned error or the task-validation rejection of a task whose unknown input member carries the canaries, every measurement path is within the registered `dns_query` set, retained answers are the scripted A values deduplicated in first-occurrence order (three responses kept when four arrive), and the strict decoder rejects 13 injected members (including `hostname`, `query_id`, `ttl`, `authority`, `asn`, `local_address` and `raw`), a duplicate answer and a TXT answer.

<a id="M-DNS-01-INTERPRETATION"></a>

### M-DNS-01-INTERPRETATION

- Case: `M-DNS-01-INTERPRETATION`
- Result: passed
- Tests: `TestMDns01Interpretation` in `dns01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns01Interpretation$' ./.`
- Scope: Assert the DNS-01 Output schema is a closed field set (outcome, observations and the registered dns_query details) with no verdict, injected, spoofed, legitimate, authoritative, correct, censored, blocked, reachable, anomaly or interference field, and that outcome values are only the registered DNS-01 outcomes. Exercise serialized results for resolved, nxdomain, no_data, conflicting_responses and response_limit_reached, the documented field descriptions and the sanitized conflicting-response example and host-facing text, and assert none states or implies that an unexpected answer was introduced intentionally, that a returned address is reachable, that answers were compared with authoritative ground truth in the run, that the first or a later response is spoofed, injected, legitimate or authoritative, that either differing resolver is correct, that censorship or blocking occurred, or that NXDOMAIN or an empty answer is a successful address resolution. Verify no answer is enriched with ASN, organization, geolocation, reputation or a censorship verdict. Requirements: REQ-DNS-01-0001, REQ-DNS-01-0002, REQ-DNS-01-0003, REQ-DNS-01-0004, REQ-DNS-01-0041, REQ-DNS-01-0042, REQ-DNS-01-0043, REQ-DNS-01-0044. Judgement-based completion is the separate review case RV-M-DNS-01-INTERPRETATION; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-DNS-01-0001: No non-negated sentence in the comments and string literals of production Go files and sanitized examples and the shipped Markdown documentation (case plans, manifests, evidence, changelog and specification-issue records excluded) matches the "unexpected answer introduced intentionally" claim pattern; no guidance qualifier is required for this item; no run's measurement contains `intent` or `inject`. This is a textual scan of shipped text; it does not prove how a reader interprets a result.
  - REQ-DNS-01-0002: The same scan finds no non-negated sentence stating that a returned address is reachable, the guidance contains "is not shown to be reachable", and the strict decoder rejects an injected `reachable` field.
  - REQ-DNS-01-0003: The same scan finds no non-negated sentence stating that answers were compared with or verified against authoritative ground truth, data or answers, and the guidance contains "answers are not compared with authoritative data in the run".
  - REQ-DNS-01-0004: The same scan finds no non-negated sentence labelling the first, a later, the second or this response spoofed, injected, legitimate or authoritative, the guidance contains "it cannot establish which response was injected or correct", no run's measurement contains `spoof`, `inject`, `legitimate` or `authoritative`, and the strict decoder rejects an injected `authoritative` field.
  - REQ-DNS-01-0041: The guidance contains "establishes only the result exposed by the selected resolver path for the configured name, record type, time and network generation", and each of the six runs (UDP and TCP `resolved`, `nxdomain`, `no_data`, `conflicting_responses`, `response_limit_reached`) records its expected outcome with a single `dns_query` observation.
  - REQ-DNS-01-0042: A direct UDP run receiving two differing datagrams records `conflicting_responses` and one receiving four records `response_limit_reached`, and the guidance contains "it cannot establish which response was injected or correct"; this test does not inspect the retained responses' contents.
  - REQ-DNS-01-0043: Not asserted: the test requires none of the listed evidence items in the guidance and checks none of them.
  - REQ-DNS-01-0044: The same scan finds no non-negated sentence stating that either, one or the differing resolver is correct, the guidance contains "not proof that either resolver is correct or that censorship occurred", and no non-negated sentence of the scanned text matches the common claims (a result is or provides definitive or conclusive proof of censorship or blocking; a result proves, demonstrates or confirms censorship or blocking; a censorship, blocking, anomaly or success verdict); no run compares two resolvers.

<a id="M-DNS-01-SAFETY"></a>

### M-DNS-01-SAFETY

- Case: `M-DNS-01-SAFETY`
- Result: passed
- Tests: `TestMDns01Safety` in `measurement_safety_conformance_test.go` (package `root`); `TestMDns01Minimization` in `dns01_conformance_test.go` (package `root`); `TestDns01EquivalentRunsSerialize` in `dns01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns01Safety$' ./.`
- Scope: Iterate the Baseline Risk rows (risk level low; requires user consent false, subject to host and local policy) and each Risk Factors row (network_observability, traffic_distinctiveness, target_sensitivity, data_sensitivity, legal_or_social_context, each low) using synthetic tasks: assert a task declaring risk_level medium, high or unknown is evaluated at that level, the effective risk is never below the registered low baseline, and a task or fake host policy requiring consent prevents execution without fake host consent. Exercise each Safety Analysis action with synthetic targets and recording transports: data minimization, asserting retained data is limited to requested type, coarse outcome, timing and normalized answers; traffic distinctiveness, asserting exactly one ordinary query of type A or AAAA per run with no test-layer retry or repetition; user-identification risk, asserting no installation identifier and no canary entries from a fake unrelated DNS cache; target risk, asserting a non-reviewed hostname or unregistered direct resolver is rejected with zero recorded traffic; repeated-execution risk, asserting the test layer does not self-schedule repeats; and probe authority, asserting a fake local policy can delay, skip or reject (not_run with local_policy_rejected) with no traffic. Requirements: REQ-DNS-01-0045, REQ-DNS-01-0046, REQ-DNS-01-0047, REQ-DNS-01-0048, REQ-DNS-01-0049, REQ-DNS-01-0050, REQ-DNS-01-0051, REQ-DNS-01-0052. Judgement-based completion is the separate review case RV-M-DNS-01-SAFETY; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-DNS-01-0045: The test asserts that the production registry records a `low` baseline for DNS-01 and that a task declaring `low`, `medium`, `high` or `unknown` has effective risk exactly that level (admitted under issuer and local bounds permitting it, rejected `task_issuer_not_authorized` with no contact when either side permits only lower levels). For the consent baseline it asserts that `false` adds no requirement (with host consent absent the run starts, yields one result and never queries the consent capability), while consent required by the task, by an issuer or by the local policy blocks every result and contact unless the fake host reports valid consent (absent, ended and unavailable all block).
  - REQ-DNS-01-0046: For each of the five factor types the test asserts that the registered baseline equals the most restrictive of the fixture's row levels (all five `low`), that declaring the factor at `low`, `medium` or `high` gives effective risk exactly the maximum of baseline, row level and declared level, that an `unknown` factor gives `unknown`, and that an unknown factor under a known overall level or a factor above the overall level is rejected `task_schema_invalid` and an unrecognized factor type `task_unsupported`, all with no contact. The row levels are fixture constants checked only against the registered baseline, and the rationale text is not asserted.
  - REQ-DNS-01-0047: The test asserts that no surface contains the query hostname, resolver address and port, system resolver list, search domain, query identifier, raw message bytes, authority or additional records, TTL, ASN, organization, local socket address or platform error text, and that the measurement holds only outcome, observation basics, engine, address family, query type, response-limit flag and per-response arrival index, offset, rcode and answer type and value, with answer values as scripted. The decoder rejects injected hostname, resolver, TTL, authority, extra answers and non-A answers.
  - REQ-DNS-01-0048: The test asserts exactly one system lookup per run (IPv4 family, the admitted name), a single UDP resolver contact after a failure (no retry), and that `TXT`, `ANY`, `MX` and `PTR` record types are rejected at admission with no contact. DNS message flags and conservative scheduling of direct resolver use are not asserted.
  - REQ-DNS-01-0049: Through a counting system resolver the test asserts that exactly one name reaches the resolver API (the traffic row confirms it is the admitted `example.org.`) and that it contains no `safety` text, and it reruns the closed-result minimization row. No fake unrelated DNS cache is installed, so the absence of copied cache contents is shown only through the closed result field set and the absent system-resolver-list and search-domain canaries.
  - REQ-DNS-01-0050: The test asserts that an unreviewed hostname and unreviewed resolver addresses (public, private and loopback) are rejected at admission with no contact, and that loopback, private, link-local and IPv6 loopback resolver addresses registered in place of the reviewed resolver are never contacted. Sensitivity review of names beyond registration is not asserted.
  - REQ-DNS-01-0051: The test asserts that a second equivalent run does not query while the first holds the resolver tuple (both then complete with two contacts in total), and that one admitted run followed by a 48-hour clock advance makes exactly one contact and one result, so the test layer does not self-schedule repeats. The residual-risk statement is not asserted.
  - REQ-DNS-01-0052: The test asserts that a local policy not permitting the task's network path rejects a valid task at admission with no contact, and that a local charging-only device constraint holds an admitted valid task (`run_not_due` twice, no result, no contact) until satisfied, after which the same task runs; a target registration that expires after admission ends the attempt `local_policy_rejected` with no contact. The no-lowering clause is asserted through the declared-level rows: no task declaration yields effective risk below the `low` baseline.

<a id="M-DNS-01-OUTPUT"></a>

### M-DNS-01-OUTPUT

- Case: `M-DNS-01-OUTPUT`
- Result: passed
- Tests: `TestMDns01ExecutionAndOutput` in `dns01_conformance_test.go` (package `root`); `TestDns01SystemAPIOutputAndHiddenBytes` in `dns01_conformance_test.go` (package `root`); `TestDns01AAAAExecution` in `dns01_conformance_test.go` (package `root`); `TestMDns01CollectionOrderAndBounds` in `dns01_conformance_test.go` (package `root`); `TestDns01ResultVersionsAndClosedOutput` in `dns01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns01ExecutionAndOutput$' ./.`
- Scope: Require exactly one dns_query with mode-correct engine, observable fields only and bounded ordered responses. System API omissions must remain omissions rather than invented codes/datagrams/timing; reuse the shared DNS observation schema for answer normalization and length/count limits.
- Requirement aspects exercised:

  - REQ-DNS-01-0035: Each completed result carries exactly one dns_query with the engine of its mode; system mode retains no response records, codes or datagram timing the API did not expose; results pass production result and bundle validation.
  - REQ-DNS-01-0036: Responses are retained in arrival order up to three, with only requested-type and chain CNAME answers, and no query name, resolver address, port, TTL or raw message appears in the result.

<a id="M-DNS-01-VERSIONS"></a>

### M-DNS-01-VERSIONS

- Case: `M-DNS-01-VERSIONS`
- Result: passed
- Tests: `TestMDns01Versions` in `dns01_conformance_test.go` (package `root`); `TestDns01ResultVersionsAndClosedOutput` in `dns01_conformance_test.go` (package `root`); `TestProductionRegistryContainsOnlyDeliveredMeasurements` in `registry_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns01Versions$' ./.`
- Scope: Select test_id DNS-01 with task test_version 0.1.0 and expect a result with schema_version 0.1.0, test.id DNS-01, test.version 0.1.0 and test.result_version 0.1.0 as the only registered combination. Reject before any recording-transport traffic, as a task-validation failure rather than a Measurement Result, each task test_version variant 0.2.0, 1.0.0, 0.1, 0.1.0-draft, v0.1.0, an empty string, a non-string value and a missing field. Verify result validation rejects schema_version or test.result_version set to 0.2.0, 0.1 or missing, a DNS-01 test.version paired with a different result_version, and cross-test combinations: test.id DNS-01 carrying another test's Output (a dns_query with engine doh or dot, or NET-05 profile and step fields) and test.id DNS-03 or DNS-04 carrying a DNS-01 direct UDP output. Assert no version is coerced, guessed or added and that the Draft status and registration metadata alone confer no support. Requirements: REQ-DNS-01-0034.
- Requirement aspects exercised:

  - REQ-DNS-01-0034: Task test_version 0.1.1, 0.2.0 and 1.0.0 are task_unsupported and 0.1, v0.1.0, 0.1.0-draft, empty, null, numeric and missing are task_schema_invalid, all with no contact or result; the decoder accepts only the 0.1.0 tuple and rejects other version values, a missing schema_version or result_version, test IDs DNS-03, DNS-04 and NET-05, NET-05 profile and step fields, doh and dot engines and an extra test member.

<a id="M-TLS-01-ADMISSION"></a>

### M-TLS-01-ADMISSION

- Case: `M-TLS-01-ADMISSION`
- Result: passed
- Tests: `TestMTls01Admission` in `tls01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTls01Admission$' ./.`
- Scope: Vary canonical public IP, port, deterministic normalized DNS server_name through 253 bytes, exact reviewed tuple and unknown fields. Reject literal/wildcard/derived SNI, private/special/interpolated targets and every certificate/hostname-verification bypass before socket activity.
- Requirement aspects exercised:

  - REQ-TLS-01-0004: Admits one public IPv4 or IPv6 literal, a port and a DNS server_name only when the tuple matches a local registration; a different address, port, server name or ALPN order, or no ALPN offer, is target_not_registered with no contact or replay commit.
  - REQ-TLS-01-0009: Accepts ports 1 and 65535, IPv4 and IPv6 literals, a mixed-case name normalized to lower case, an IDNA name registered as its A-label and a 253-byte name; rejects a missing field, port 0, 65536 or a string, a null or array server_name, a 254-byte name, a 64-byte label and IP-literal, bracketed, wildcard, single-label, trailing-dot, empty-label, leading-hyphen, underscore and leading-space names as task_schema_invalid.
  - REQ-TLS-01-0010: Admission requires the complete registered tuple of address, port, normalized server_name and ordered ALPN offer; changing any one element is target_not_registered.
  - REQ-TLS-01-0011: Rejects hostnames, interpolation placeholders, non-canonical, unspecified, loopback, private, shared, link-local, documentation, benchmarking, multicast, IPv4-mapped, uppercase IPv6 and zoned literals, and localhost, .local, .test and interpolated server names, as task_schema_invalid before contact.
  - REQ-TLS-01-0012: A missing server_name and an IP-literal server_name are rejected, so no name is derived from the address.
  - REQ-TLS-01-0013: insecure_skip_verify and verify_hostname:false input members are rejected as task_schema_invalid.
  - REQ-TLS-01-0014: An unknown member and a duplicated server_name member are rejected.
  - REQ-TLS-01-0015: Every rejection returns no handle and leaves no task handle, replay commit, run, transport contact or result record.

<a id="M-TLS-01-OPTIONS"></a>

### M-TLS-01-OPTIONS

- Case: `M-TLS-01-OPTIONS`
- Result: passed
- Tests: `TestMTls01Options` in `tls01_conformance_test.go` (package `root`); `TestMTls01ALPNBoundsIsolated` in `internal/measurement/tls01/tls01_options_test.go` (package `internal/measurement/tls01`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTls01Options$' ./.`
- Scope: Exercise omitted/empty ALPN, 1..8 identifiers, 1..32 ASCII bytes, ordering, registered values and duplicates at all boundaries. Only h2, http/1.1 and dot are registered offer identifiers, so through admission registry membership is the stricter check and subsumes the 1..32-byte length bound, and the 8-identifier count bound is isolated at the closed options descriptor. Reject custom cipher/version profiles, invalid-certificate controls, alternate SNI, client certificates, resumption and insecure verification.
- Requirement aspects exercised:

  - REQ-TLS-01-0016: Through Client.AdmitTask, omitted and empty offers, h2, dot, http/1.1 and the ordered offer http/1.1,h2,dot are admitted, while null, string, numeric, empty, unregistered (unknown, h2c, H2, h3, spdy/3, 33 bytes), duplicate and nine-entry offers and a duplicated member are task_schema_invalid with no contact. At the closed options descriptor eight distinct well-formed identifiers pass the count bound and nine do not; registry membership then rejects unregistered 1-, 32- and 33-byte identifiers and keeps the registered identifiers in their given order (TestMTls01ALPNBoundsIsolated).
  - REQ-TLS-01-0017: cipher_suites, tls_version, min_version, insecure_skip_verify, allow_invalid_certificates, server_name, sni, client_certificate, session_resumption and session_ticket options are task_schema_invalid with no contact.

<a id="M-TLS-01-PRECONDITIONS"></a>

### M-TLS-01-PRECONDITIONS

- Case: `M-TLS-01-PRECONDITIONS`
- Result: passed
- Tests: `TestMTls01Preconditions` in `tls01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTls01Preconditions$' ./.`
- Scope: Run the portable TLS-01 gate against a fake platform, synthetic target registry and recording transport. Stable network generation unavailable, and execution deadline unavailable: assert execution.state not_run with no TCP connection opened. Target not registered as expected to support normal publicly authenticated TLS: assert not_run. Target registered as requiring client authentication, credentials or a sensitive server name (iterate each): assert not_run and no ClientHello sent. All preconditions met: assert exactly one TCP connection and one handshake. Stricter task execution constraint or local policy denial: assert constraints_unsatisfied or local_policy_rejected; other denials use a registered reason per the public table. Requirements: REQ-TLS-01-0018, REQ-TLS-01-0019, REQ-TLS-01-0020, REQ-TLS-01-0021, REQ-TLS-01-0022. Judgement-based completion is the separate review case RV-M-TLS-01-PRECONDITIONS; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-TLS-01-0018: Through the shared gates, an unreadable network access state at reservation defers the run (`run_not_due`, no result), and a network change after reservation to a transport the task does not allow gives `not_run` with `constraints_unsatisfied`, both with zero contacts (no dial). Execution deadline availability is not asserted.
  - REQ-TLS-01-0019: A TCP-01 registration of the same address and port does not admit a TLS-01 task, and tasks with another server name or ALPN offer are rejected at admission, all with no dial. Through the `TestMTls01ExecutionAndOutput` subtest, with all preconditions met on the direct and proxy paths the loopback server accepts exactly one connection and one ClientHello carrying the registered server name and ALPN list, with one direct dial or one proxy CONNECT, and the result passes bundle validation.
  - REQ-TLS-01-0020: Not asserted by this case; documentation of the trust policy and time source is the review case `RV-M-TLS-01-PRECONDITIONS`, see docs/measurement-requirements.md.
  - REQ-TLS-01-0021: A task whose server name differs from the registration (`sensitive.example`) is rejected at admission with no dial. Registrations carry no client-authentication or credential marking, so those rows are not asserted here; see docs/measurement-requirements.md (`RV-M-TLS-01-PRECONDITIONS`, `M-TLS-01-SAFETY`).
  - REQ-TLS-01-0022: The shared portable gates assert, each with zero contacts: a task transport condition the network does not meet, a task excluding VPN on a VPN-detected network and a local charging-only device restriction defer the run at `StartRun` (`run_not_due`, no result); a transport condition lost after reservation gives `not_run` with `constraints_unsatisfied`; and a proxy-path task with no proxy configured is either rejected at admission with `task_schema_invalid` or ends `not_run` with `proxy_not_configured` (the test accepts either); a deployment that withholds the target registration rejects the task with `target_not_registered`, and a registration that lapses after admission gives a result with reason `local_policy_rejected`.

<a id="M-TLS-01-EXECUTION"></a>

### M-TLS-01-EXECUTION

- Case: `M-TLS-01-EXECUTION`
- Result: passed
- Tests: `TestMTls01ExecutionAndOutput` in `tls01_conformance_test.go` (package `root`); `TestMTls01NoImplicitDNS` in `tls01_execution_closeout_test.go` (package `root`); `TestMTls01NoSessionResumption` in `tls01_execution_closeout_test.go` (package `root`); `TestTls01PlaintextAccounting` in `internal/execution/tls01_test.go` (package `internal/execution`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTls01ExecutionAndOutput$' ./.`
- Scope: Record explicit-IP TCP then one normal SNI/ALPN handshake with certificate/hostname validation; retain first failed stage or minimized negotiated properties and close before application data. No implicit DNS, alternate profiles, SNI comparison, sessions or retries.
- Requirement aspects exercised:

  - REQ-TLS-01-0005: On the direct path and through the scripted CONNECT proxy, the loopback server accepts one connection and sees one ClientHello with the configured server name and the ALPN offer h2,http/1.1 in order; the chain validates against the synthetic CA, no plaintext arrives and the connection closes. Against a session-ticket endpoint that a caching control client does resume, two consecutive TLS-01 runs both perform full handshakes (TestMTls01NoSessionResumption).
  - REQ-TLS-01-0006: The direct path dials once for the literal authority and the proxy path issues one CONNECT with no direct dial; an installed system resolver is never called for a handshake or a refused connect (TestMTls01NoImplicitDNS); the meter records one connection and zero application bytes while handshake bytes cross the socket and tunnel (TestTls01PlaintextAccounting); the result passes bundle validation.

<a id="M-TLS-01-NORMALIZATION"></a>

### M-TLS-01-NORMALIZATION

- Case: `M-TLS-01-NORMALIZATION`
- Result: passed
- Tests: `TestMTls01Normalization` in `tls01_conformance_test.go` (package `root`); `TestRuntimeTLSRejectsUnofferedALPN` in `runtime_vocabularies_test.go` (package `root`); `TestMTls01Limits` in `tls01_conformance_test.go` (package `root`); `TestMTls01Termination` in `tls01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTls01Normalization$' ./.`
- Scope: Drive TLS-01 through a scripted transport and deterministic clock and iterate every outcome row. Assert each outcome and its terminal observation: handshake_succeeded with tls_handshake succeeded and certificate_validation trusted; tcp_timeout, tcp_refused with tcp_connect failed timeout or connection_refused and no tls_handshake observation; tcp_reset from tcp_connect failed connection_reset, and separately from tls_handshake failed connection_reset after a succeeded tcp_connect; tls_timeout with tls_handshake failed timeout; tls_alert with tls_handshake failed alert, with details.alert either absent or a registered coarse alert class and no raw alert; protocol_error with tls_handshake failed protocol_error; certificate_untrusted, certificate_name_mismatch, certificate_time_invalid with the same failure value and certificate_validation untrusted, name_mismatch or time_invalid respectively; alpn_no_overlap with tls_handshake failed alpn_no_overlap; inconclusive when the terminal TCP or TLS observation failed inconclusive. Assert the required certificate_validation value is present on every tls_handshake observation and the stage-specific failure is preserved rather than collapsed into a single filtering outcome. Inject network_changed and the outer deadline and assert an incomplete execution state, not a TLS outcome. Assert success only when the handshake completes, the chain is accepted by the documented trust policy and the certificate is valid for server_name. Requirements: REQ-TLS-01-0007, REQ-TLS-01-0008, REQ-TLS-01-0026.
- Requirement aspects exercised:

  - REQ-TLS-01-0007: handshake_succeeded is emitted only after a completed handshake with certificate_validation trusted against the synthetic CA, with and without a negotiated ALPN identifier; a chain from another CA, a certificate for another name and an expired leaf each fail.
  - REQ-TLS-01-0008: Each failure class arises from real loopback behavior or a wrapped typed dial error: refused, reset and timeout dial errors; a reset and a stall during the handshake; peer alerts 40, 42, 49, 70, 80 and 200 as handshake_failure, certificate_error, access_denied, protocol_error, internal_error and other; alert 120 without an offer as an alert; a server supporting only dot against the h2,http/1.1 offer as alpn_no_overlap; non-TLS bytes as protocol_error; and untrusted, name-mismatched and expired certificates. Exactly one dial is made and no application data is sent.
  - REQ-TLS-01-0026: Each outcome has its terminal observation: tcp_refused, tcp_reset and tcp_timeout with one failed tcp_connect and no tls_handshake; tcp_reset from a TLS connection_reset after a succeeded connect; tls_timeout at the 15-second handshake bound; tls_alert with only a coarse alert class; protocol_error; the three certificate outcomes with certificate_validation untrusted, name_mismatch or time_invalid; alpn_no_overlap; and inconclusive from an untyped dial error, from an unreachable dial error without a TLS-01 class, and from a TLS inconclusive failure after an unoffered ALPN selection (TestRuntimeTLSRejectsUnofferedALPN). certificate_validation is present on every tls_handshake. A network change or host cancellation during a stalled handshake (TestMTls01Termination) and the outer deadline (TestMTls01Limits) give incomplete with no measurement.

<a id="M-TLS-01-LIMITS"></a>

### M-TLS-01-LIMITS

- Case: `M-TLS-01-LIMITS`
- Result: passed
- Tests: `TestMTls01Limits` in `tls01_conformance_test.go` (package `root`); `TestMTls01TupleConcurrency` in `tls01_conformance_test.go` (package `root`); `TestTls01PlaintextAccounting` in `internal/execution/tls01_test.go` (package `internal/execution`); `TestScopeSessionMetersPlaintext` in `internal/execution/tls01_test.go` (package `internal/execution`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTls01Limits$' ./.`
- Scope: Assert one connection/handshake, full-chain deadline, no application payload/retries and deduplication by endpoint/server-name tuple. Test cancellation and close while the peer stalls in handshake.
- Requirement aspects exercised:

  - REQ-TLS-01-0023: Against a stalled handshake the 15-second operation bound gives completed tls_timeout, a 1,000 ms outer deadline gives incomplete attempt_deadline_exceeded and host cancellation gives incomplete cancelled_by_host, each with one dial, one accepted connection and the socket released; max_connections 1 and 2 with one-byte byte limits complete with exactly one connection and no retry; an equivalent endpoint and server-name tuple waits without contact and is not_run when cancelled; handshake and record bytes are not metered while plaintext bytes count against the same budget (TestScopeSessionMetersPlaintext).

<a id="M-TLS-01-TERMINATION"></a>

### M-TLS-01-TERMINATION

- Case: `M-TLS-01-TERMINATION`
- Result: passed
- Tests: `TestMTls01Termination` in `tls01_conformance_test.go` (package `root`); `TestTls01AllTerminationBoundaries` in `internal/execution/tls01_test.go` (package `internal/execution`); `TestMTls01Limits` in `tls01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTls01Termination$' ./.`
- Scope: Exercise a valid TLS-01 task against a synthetic target accepted by a test-only local policy fixture, using a deterministic clock, fake platform and scripted TCP/TLS transport, and interrupt at each Procedure point: after input validation (step 1), after network-generation and address-family confirmation (step 2), while the TCP connect is pending (step 3), after TCP success before the ClientHello is sent, while the TLS handshake is stalled mid-exchange (step 4), and after negotiated properties are recorded but before close (steps 5-6). Expected states follow the GO-SPEC-ISSUE-004 disposition: after input validation (step 1) the attempt is not_run with that reason, no measurement and no network_context; step 2 starts the test; after the negotiated properties are recorded and before close it stays completed with the recorded outcome; every other point is incomplete as asserted below. At each point inject in turn a network-generation change, host cancellation, outer attempt-deadline expiry, proxy reconfiguration (and other registered proxy_* reasons where a proxy path is configured, per the public table) and platform suspension. At the incomplete points, assert execution.state is incomplete with execution.reason exactly network_changed, cancelled_by_host, attempt_deadline_exceeded, proxy_reconfigured (or the matching proxy_* reason) or platform_suspended, naming only the first condition; assert measurement is absent at those points because TLS-01 defines no partial Output (only NET-05 may retain partial measurement); assert no outcome is network_changed or any other invented TLS outcome, no tcp_connect or tls_handshake observation is emitted for the interrupted stage, the socket is closed, no application payload is sent, and no second address, server name, TLS profile or retry is attempted. Verify the contrast cases: a bounded TCP or TLS operation timeout before the outer deadline yields execution.state completed with outcome tcp_timeout or tls_timeout, while expiry of the outer deadline yields attempt_deadline_exceeded with no measurement. Requirements: REQ-TLS-01-0027.
- Requirement aspects exercised:

  - REQ-TLS-01-0027: For host cancellation, network change, attempt deadline, injected platform suspension, proxy_unavailable, proxy_connect_failed and proxy_reconfigured, with a second cause injected after the first: an interruption before the test starts is not_run with no network_context; at the generation and family check, during the connect, after TCP success before the ClientHello and during a stalled handshake it is incomplete with the first reason, no measurement, at most one accepted connection, no plaintext and no timer left; during close after the outcome is recorded it stays completed with handshake_succeeded. Through the Client a network change and host cancellation during a stalled handshake are incomplete with the socket closed, and a refusing CONNECT proxy gives proxy_connect_failed with no direct dial; the 15-second handshake bound gives completed tls_timeout while the outer deadline gives attempt_deadline_exceeded (TestMTls01Limits).

<a id="M-TLS-01-MINIMIZATION"></a>

### M-TLS-01-MINIMIZATION

- Case: `M-TLS-01-MINIMIZATION`
- Result: passed
- Tests: `TestMTls01Minimization` in `tls01_conformance_test.go` (package `root`), `TestMTls01Safety` in `measurement_safety_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTls01Minimization$' ./.`
- Scope: Exercise TLS-01 with a synthetic task and scripted TCP/TLS transport and fake trust store, and inject one unique synthetic canary into each item REQ-TLS-01-0028 names and that the shared tls_handshake observation contract also prohibits: the destination IP address, the destination port, server_name, a certificate DER blob and a certificate-chain element, certificate subject, issuer and SAN names, certificate serial number, public key and signature, certificate fingerprint, the raw TLS alert code and description, ClientHello fields (client random, extension bytes), ServerHello fields (server random, extension bytes), the offered cipher-suite list, the negotiated cipher suite, the session identifier and session ticket, TLS secrets (key-log material), and the raw library error string. Drive the success path and each failure class (tcp_timeout, tcp_refused, tcp_reset, tls_timeout, tls_alert, protocol_error, certificate_untrusted, certificate_name_mismatch, certificate_time_invalid, alpn_no_overlap, inconclusive) plus incomplete and not_run paths. Assert serialized result and error paths, including every nested observation, contain no canary and no input digest, and contain only the allowed fields: measurement.outcome, and per observation type, status, duration_ms, failure, details.address_family, and for tls_handshake details.tls_version, details.alpn, details.certificate_validation and a registered coarse details.alert class; assert any extra field is rejected. No participant or live measurement data is used. Requirements: REQ-TLS-01-0028.
- Requirement aspects exercised:

  - REQ-TLS-01-0028: Unique canaries cover the destination address and port, `server_name`, a mismatched SAN name, the certificate subject, issuer and serials, the raw alert code, raw library, x509, TLS and transport error texts and the ALPN offer list, plus hexadecimal and base64 forms of each test certificate's DER, chain element, public key, signature and SHA-256 fingerprint and, for runs that reached a TLS server, the recorded ClientHello and ServerHello randoms, session identifiers, offered cipher suites and extension bytes. Over `handshake_succeeded`, `tcp_timeout`, `tcp_refused`, `tcp_reset`, `tls_timeout`, `tls_alert`, `protocol_error`, the three certificate outcomes, `alpn_no_overlap`, `inconclusive`, host-cancelled `incomplete` and `not_run`, no canary appears in any result, the bundle, the pending store's bytes, any returned error or the task-validation rejection of a task whose unknown input member carries the canaries and every measurement path is within the registered TLS-01 set, and the strict decoder rejects 23 injected members (including `cipher_suite`, `session_id`, `session_ticket`, `key_log`, `client_random`, `fingerprint` and `input_digest`) and a raw `alert(200)` value. Handshake secrets, session tickets and the negotiated cipher suite have no canary and are covered only by the closed field set and these injections.

<a id="M-TLS-01-INTERPRETATION"></a>

### M-TLS-01-INTERPRETATION

- Case: `M-TLS-01-INTERPRETATION`
- Result: passed
- Tests: `TestMTls01Interpretation` in `tls01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTls01Interpretation$' ./.`
- Scope: Assert the TLS-01 result schema has no interpretation, verdict, anomaly, blocked, censored, filtered or confidence field and that such a field is rejected as unknown. Assert result field descriptions, documented outcome descriptions and the sanitized example and host-facing text contain none of these prohibited claims: that a handshake failure establishes TLS or SNI filtering, censorship, blocking, interference or intent; that a certificate_untrusted, certificate_name_mismatch or certificate_time_invalid outcome establishes interception or tampering rather than target, CDN, trust-store, certificate or device-time misconfiguration; that tcp_timeout, tcp_reset or tls_timeout excludes routing, packet loss, mobile handover, VPN, firewall, proxy, rate limiting or overload; that alpn_no_overlap or protocol_error excludes unsupported TLS version, algorithm or ALPN configuration; that handshake_succeeded establishes HTTP or application-content availability; that a result compares different SNI values; or that a failure excludes a stale address-to-name mapping. Verify outcome text is limited to whether TCP and normal authenticated TLS completed for the configured address, port, server name and ALPN offer. Requirements: REQ-TLS-01-0001, REQ-TLS-01-0002, REQ-TLS-01-0003, REQ-TLS-01-0029, REQ-TLS-01-0030, REQ-TLS-01-0031. Judgement-based completion is the separate review case RV-M-TLS-01-INTERPRETATION; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-TLS-01-0001: No non-negated sentence in the comments and string literals of production Go files and sanitized examples and the shipped Markdown documentation (case plans, manifests, evidence, changelog and specification-issue records excluded) matches the "handshake_succeeded establishes HTTP availability" claim pattern, and the TLS-01 section of the interpretation guidance contains "does not establish http or application-content availability"; the test does not check that no HTTP request is sent. This is a textual scan of shipped text; it does not prove how a reader interprets a result.
  - REQ-TLS-01-0002: The same scan finds no non-negated sentence stating that TLS-01 compares SNI values or server names, and the guidance contains "does not compare different sni values".
  - REQ-TLS-01-0003: The same scan finds no non-negated sentence stating that a handshake or TLS failure establishes filtering, censorship, blocking, interference or intent, or that a certificate outcome establishes interception or tampering, the guidance contains "does not establish tls or sni filtering, censorship, blocking, interference or intent" and "does not establish interception or tampering", and no member name or string value of any run's measurement contains a verdict token (such as `verdict`, `censor`, `block`, `filter`, `anomal`, `confiden`) and the strict result decoder rejects `verdict`, `interpretation`, `censored`, `blocked`, `filtered`, `anomaly`, `confidence`, `reachable` and `authoritative` fields injected into a valid result.
  - REQ-TLS-01-0029: The guidance contains "establishes whether tcp and normal authenticated tls completed for the configured address, port, server name and alpn offer", and the eleven runs (success, three TCP errors, a generic dial error, alert, garbage, untrusted, mismatched and expired certificates, ALPN mismatch) all carry registered outcomes; which outcome each run produced is not checked here.
  - REQ-TLS-01-0030: Not asserted: the test requires none of the listed evidence items in the guidance and checks none of them.
  - REQ-TLS-01-0031: The same SNI-comparison claim pattern and the "does not compare different sni values" qualifier apply; that such a comparison would need a separately reviewed measurement is not checked.

<a id="M-TLS-01-SAFETY"></a>

### M-TLS-01-SAFETY

- Case: `M-TLS-01-SAFETY`
- Result: passed
- Tests: `TestMTls01Safety` in `measurement_safety_conformance_test.go` (package `root`); `TestMTls01Minimization` in `tls01_conformance_test.go` (package `root`); `TestMTls01ExecutionAndOutput` in `tls01_conformance_test.go` (package `root`); `TestMTls01TupleConcurrency` in `tls01_conformance_test.go` (package `root`); `TestMTls01NoImplicitDNS` in `tls01_execution_closeout_test.go` (package `root`); `TestMTls01NoSessionResumption` in `tls01_execution_closeout_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTls01Safety$' ./.`
- Scope: Iterate every Baseline Risk and Risk Factors row of TLS-01 using synthetic targets, a fake local policy and a recording TCP/TLS transport. Baseline Risk rows: Assert registered risk level low and requires_user_consent false, and Verify a fake host or local policy that requires consent or declines execution still blocks the run (not_run with the applicable reason per the public table). Risk Factors rows network_observability, traffic_distinctiveness, target_sensitivity, data_sensitivity and legal_or_social_context, each low: Assert a task declaring risk_level low is evaluated at the registered baseline, a task adding a medium, high or unknown factor raises effective risk to the most restrictive value, and no task declaration lowers the baseline. Exercise each normative safety action: data minimization, asserting no full certificate chain, raw ClientHello or ServerHello, session key or packet trace is retained; traffic distinctiveness, asserting the recording transport sees exactly one TCP connection to the configured IP and port, one standards-compliant ClientHello with the configured server_name and registered ALPN list in order, no malformed, alternate-SNI or comparative handshake, no implicit DNS lookup and no retry; user-identification risk, asserting no client certificate, identifier, cookie or application data byte is sent and the session closes after the handshake; target risk, asserting an unreviewed address, port and server-name tuple is rejected before any traffic; repeated-execution risk, asserting one active run per endpoint and server-name tuple with equivalent runs deduplicated or delayed. Verify the probe can delay, skip or reject a run that the task permits. Requirements: REQ-TLS-01-0032, REQ-TLS-01-0033, REQ-TLS-01-0034, REQ-TLS-01-0035, REQ-TLS-01-0036, REQ-TLS-01-0037, REQ-TLS-01-0038. Judgement-based completion is the separate review case RV-M-TLS-01-SAFETY; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-TLS-01-0032: The test asserts that the production registry records a `low` baseline for TLS-01 and that a task declaring `low`, `medium`, `high` or `unknown` has effective risk exactly that level (admitted under issuer and local bounds permitting it, rejected `task_issuer_not_authorized` with no contact when either side permits only lower levels). For the consent baseline it asserts that `false` adds no requirement (with host consent absent the run starts, yields one result and never queries the consent capability), while consent required by the task, by an issuer or by the local policy blocks every result and contact unless the fake host reports valid consent (absent, ended and unavailable all block).
  - REQ-TLS-01-0033: For each of the five factor types the test asserts that the registered baseline equals the most restrictive of the fixture's row levels (all five `low`), that declaring the factor at `low`, `medium` or `high` gives effective risk exactly the maximum of baseline, row level and declared level, that an `unknown` factor gives `unknown`, and that an unknown factor under a known overall level or a factor above the overall level is rejected `task_schema_invalid` and an unrecognized factor type `task_unsupported`, all with no contact. The row levels are fixture constants checked only against the registered baseline, and the rationale text is not asserted.
  - REQ-TLS-01-0034: Across the success path, every failure class and the incomplete and `not_run` paths, the test asserts that no surface (results, pending store, bundle, error text) contains leaf or chain certificate DER, public key, signature, fingerprint, serials, subject or issuer, raw alert or error text, the ALPN offer list, or the ClientHello and ServerHello random, session identifier, cipher-suite list and extension bytes, and that the measurement holds only the closed field set (outcome, observation type, status, duration, failure, address family, TLS version, ALPN, certificate validation, alert). The decoder rejects injected certificate, chain, session ticket, key log and random fields; session keys and packet traces are covered only by that closed field set.
  - REQ-TLS-01-0035: On direct and proxy paths the test asserts exactly one accepted connection carrying one ClientHello with the configured server name and ALPN list `h2,http/1.1` in order, no plaintext after the handshake and a closed connection; it also asserts zero system resolver lookups and a single dial after a refused connect (no retry). Platform-typical TLS behavior beyond that single standard handshake is not asserted.
  - REQ-TLS-01-0036: The test asserts that a server requesting a client certificate receives none, that two consecutive runs do not resume a TLS session (a control client against the same endpoint does), and, in the traffic row, that zero application bytes follow the handshake. Cookies and identifiers are covered only by that zero-byte check.
  - REQ-TLS-01-0037: The test asserts that an unreviewed server name, port or address is rejected at admission with no dial, and that loopback, private, link-local, documentation and IPv6 loopback addresses registered in place of the reviewed tuple are never contacted. For repeated execution it asserts that a second run on the same tuple makes no connection while the first is in its handshake and can be cancelled `not_run`; the residual-risk statement is not asserted.
  - REQ-TLS-01-0038: The test asserts that a local policy not permitting the task's network path rejects a valid task at admission with no contact, and that a local charging-only device constraint holds an admitted valid task (`run_not_due` twice, no result, no contact) until satisfied, after which the same task runs; a target registration that expires after admission ends the attempt `local_policy_rejected` with no contact. The no-lowering clause is asserted through the declared-level rows: no task declaration yields effective risk below the `low` baseline.

<a id="M-TLS-01-OUTPUT"></a>

### M-TLS-01-OUTPUT

- Case: `M-TLS-01-OUTPUT`
- Result: passed
- Tests: `TestMTls01ExecutionAndOutput` in `tls01_conformance_test.go` (package `root`); `TestMTls01Normalization` in `tls01_conformance_test.go` (package `root`); `TestMTls01Versions` in `tls01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTls01ExecutionAndOutput$' ./.`
- Scope: Require TCP then TLS only when handshake activity began, in execution order. Check minimized allowed negotiated properties against common TCP/TLS observations with certificate-validation classes preserved.
- Requirement aspects exercised:

  - REQ-TLS-01-0025: Completed results carry tcp_connect then tls_handshake when TLS began and only tcp_connect when the connect failed, with 250-ms quantized durations, address_family ipv4 and no details other than address_family, tls_version, alpn, certificate_validation and alert; the decoder rejects a truncated or reversed sequence, other observation types, a changed or missing certificate_validation, an unregistered alpn or tls_version, an alert on a success and server_name, cipher_suite and certificate details.

<a id="M-TLS-01-VERSIONS"></a>

### M-TLS-01-VERSIONS

- Case: `M-TLS-01-VERSIONS`
- Result: passed
- Tests: `TestMTls01Versions` in `tls01_conformance_test.go` (package `root`); `TestProductionRegistryContainsOnlyDeliveredMeasurements` in `registry_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMTls01Versions$' ./.`
- Scope: Select the exact registered combination test.id TLS-01, test.version 0.1.0, Measurement Result schema_version 0.1.0 and test.result_version 0.1.0, and Assert a synthetic completed TLS-01 result validates only with that combination. Reject, before execution or interpretation and without coercion or best-effort parsing: schema_version 0.2.0, 1.0.0, 0.1 and empty or non-string values; test.version 0.2.0 or 0.1; test.result_version 0.2.0 or missing; a TLS-01 measurement labelled test.id TCP-01 or HTTP-01; a TLS-01 envelope carrying the TCP-01 outcome connected or any outcome not in the TLS-01 Outcomes table; a TLS-01 result containing an observation type other than tcp_connect or tls_handshake; and unknown fields in the envelope, test object or measurement. Verify a task referencing an unsupported TLS-01 version is a task-validation failure reported outside Measurement Results, that no additional version is added or inferred, and that the Draft status and registration metadata of TLS-01 confer no support claim. Requirements: REQ-TLS-01-0024.
- Requirement aspects exercised:

  - REQ-TLS-01-0024: Task test_version 0.1.1, 0.2.0 and 1.0.0 are task_unsupported and empty, v0.1.0, 0.1, padded, numeric, null and missing values are task_schema_invalid, all with no handle, replay commit, contact or result; the decoder accepts only the 0.1.0 tuple and rejects 0.1.1, 0.2.0, 1.0.0, 0.1, v0.1.0, empty, numeric and null values of schema_version, test.version and test.result_version, a missing result_version, test IDs TCP-01 and HTTP-01, the outcomes connected and blocked, tls_alert on a successful handshake, extra envelope, test and measurement members and other observation types.

<a id="M-HTTP-01-ADMISSION"></a>

### M-HTTP-01-ADMISSION

- Case: `M-HTTP-01-ADMISSION`
- Result: passed
- Tests: `TestMHttp01Admission` in `http01_conformance_test.go` (package `root`); `TestMHttp01ProxyAdmission` in `http01_conformance_test.go` (package `root`); `TestCodecBoundsAndNormalization` in `internal/measurement/http01/codec_test.go` (package `internal/measurement/http01`); `TestMHttp01CheckSemantics` in `http01_closeout_test.go` (package `root`); `TestMHttp01Redirects` in `http01_conformance_test.go` (package `root`); `TestMHttp01Normalization` in `http01_conformance_test.go` (package `root`); `TestHTTPRedirectUsesProductionDestinationGate` in `internal/execution/http_primitive_test.go` (package `internal/execution`); `FuzzHttp01Parameters` in `internal/measurement/http01/codec_test.go` (package `internal/measurement/http01`)
- Reproduce the primary test: `go test -count=1 -run '^TestMHttp01Admission$' ./.`
- Scope: Vary HTTPS URL/2,048-byte bounds, optional status, body/header arrays 0..8 and value bounds through 256 UTF-8 bytes; check all nested types/enums/conditional values, SHA-256 syntax and unknown fields. Exercise case-sensitive body operations, case-insensitive header names, multivalue any-match and all-checks conjunction. Reject sensitive headers, unsafe URLs/addresses and unregistered redirect targets before contact.
- Requirement aspects exercised:

  - REQ-HTTP-01-0004: A task is admitted only for the registered normalized URL and checks; a different path is target_not_registered with no contact or replay commit.
  - REQ-HTTP-01-0010: Rejects empty, null and array input, a null URL, http, credential, fragment and loopback URLs, expected_status 600 and "200", and unknown or duplicated members; the codec accepts URLs of 2,047 and 2,048 bytes and rejects 2,049, and accepts 0, 1 and 8 body or header checks with 1- and 256-byte values while rejecting 9 checks and 257-byte values (TestCodecBoundsAndNormalization).
  - REQ-HTTP-01-0011: Rejects a regex body check, uppercase SHA-256 hex, an empty contains value and a null prefix value; accepts a 64-character lowercase sha256 value.
  - REQ-HTTP-01-0012: Through the production runner, contains, prefix and suffix comparisons are case-sensitive, and with several body checks a single mismatch gives body_check_failed with per-index results (TestMHttp01CheckSemantics).
  - REQ-HTTP-01-0013: Rejects Set-Cookie, an invalid header name, a value on present and a missing or empty value on equals; header names match case-insensitively and are stored in lower case in the immutable tuple (TestMHttp01CheckSemantics, TestCodecBoundsAndNormalization).
  - REQ-HTTP-01-0014: equals and contains succeed when any value of a repeated header matches, a header with no matching value fails, every header check must match, and results retain indexed outcomes without header names or values (TestMHttp01CheckSemantics).
  - REQ-HTTP-01-0015: An http URL is rejected as task_schema_invalid.
  - REQ-HTTP-01-0016: Credential, loopback, interpolated-host, IPv4-mapped, fragment and invalid-port URLs are rejected.
  - REQ-HTTP-01-0017: The exact normalized URL, including its path, must match a registration; a different path is target_not_registered. No query-parameter case is included.
  - REQ-HTTP-01-0018: A redirect target is checked before contact: an unregistered next URL is redirect_rejected after one connection (TestMHttp01Redirects), and at the primitive level loopback and unregistered redirect targets are rejected without a second connection on the direct and proxy paths (TestHTTPRedirectUsesProductionDestinationGate).
  - REQ-HTTP-01-0019: A redirect to an http URL is redirect_rejected (TestMHttp01Normalization), and loopback and unregistered redirect targets are rejected before contact (TestHTTPRedirectUsesProductionDestinationGate).
  - REQ-HTTP-01-0020: Body and header check counts are bounded at 8 entries and string values at 256 bytes (TestCodecBoundsAndNormalization); normalized parameters from the fuzz seed corpus keep a body bound of 1 to 65,536 and valid checks.
  - REQ-HTTP-01-0021: A Set-Cookie header check is rejected, and successful results contain no header names or values.
  - REQ-HTTP-01-0022: Unknown and duplicated input members are rejected.
  - REQ-HTTP-01-0023: Every rejection returns no handle and leaves no task handle, replay commit, run or transport contact.
  - REQ-HTTP-01-0051: With network_path proxy, a hostname URL and an IP URL with address_family ipv4 are task_schema_invalid at admission (TestMHttp01ProxyAdmission).

<a id="M-HTTP-01-OPTIONS"></a>

### M-HTTP-01-OPTIONS

- Case: `M-HTTP-01-OPTIONS`
- Result: passed
- Tests: `TestMHttp01Options` in `http01_conformance_test.go` (package `root`); `TestMHttp01OptionsAddressFamilyAndDefaults` in `http01_closeout_test.go` (package `root`); `TestMHttp01OptionsOverridesRejected` in `http01_closeout_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMHttp01Options$' ./.`
- Scope: Test auto/ipv4/ipv6, max_body_bytes default 4096 and 1..65536 with stricter task bounds, follow_redirects default false and conditional max_redirects 1/2. Reject max_redirects when disabled and every custom method/header/body/cookie/credential/proxy/compression/TLS override.
- Requirement aspects exercised:

  - REQ-HTTP-01-0024: Admits empty options, ipv4, max_body_bytes 1 and 65,536 and follow_redirects with max_redirects 1 or 2; rejects an unknown or null address_family and max_body_bytes 0 or a string (task_schema_invalid), max_body_bytes 65,537 and max_redirects 3 (task_limit_exceeded), and max_body_bytes 4,096 against max_bytes_received 4,095. ipv6 is rejected for an IPv4 literal and ipv4 for an IPv6 literal; an IPv6 literal URL completes with ipv6 observations under ipv6, auto and the default; the default 4,096-byte body bound truncates a 4,097-byte body and is rejected against max_bytes_received 4,095 (task_limit_exceeded).
  - REQ-HTTP-01-0025: follow_redirects true without max_redirects, max_redirects without following or with following false, and method, headers, user_agent, body, request_body, cookies, cookie, credentials, authorization, proxy, network_path, compression, accept_encoding, tls, server_name, alpn_protocols and insecure_skip_verify options and null option values are task_schema_invalid with no contact.

<a id="M-HTTP-01-PRECONDITIONS"></a>

### M-HTTP-01-PRECONDITIONS

- Case: `M-HTTP-01-PRECONDITIONS`
- Result: passed
- Tests: `TestMHttp01Preconditions` in `http01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMHttp01Preconditions$' ./.`
- Scope: Run the portable HTTP-01 gate with a synthetic local target registry, deterministic clock, fake platform and recording transport. URL not locally registered, or registered but not current: assert execution.state not_run and no DNS, TCP, TLS or HTTP activity. Expected response properties not registered or not current: assert not_run. A permitted redirect target or response check not registered with the measurement target (iterate each): assert not_run. Stable network generation or task execution deadline unavailable: assert not_run. Target whose local registration does not mark it as a small response suitable for bounded mobile measurement: assert not_run. Fake HTTP stack reporting that credentials, cookies or identifying headers would be attached and cannot be disabled (iterate each): assert not_run and no request sent; when they can be disabled, assert the recorded request carries none of them. Stricter execution constraint or local policy denial: assert constraints_unsatisfied or local_policy_rejected; other denials use a registered reason per the public table. Requirements: REQ-HTTP-01-0026, REQ-HTTP-01-0027, REQ-HTTP-01-0028, REQ-HTTP-01-0029, REQ-HTTP-01-0030, REQ-HTTP-01-0031. Judgement-based completion is the separate review case RV-M-HTTP-01-PRECONDITIONS; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-HTTP-01-0026: Tasks whose URL path or expected status differs from the local registration are rejected at admission with no dial, as is a task whose registration expired before admission; a withheld registration gives `target_not_registered` and a registration lapsed after admission `local_policy_rejected`, both with no dial.
  - REQ-HTTP-01-0027: Through the shared gates, an unreadable network access state at reservation defers the run (`run_not_due`, no result), and a network change after reservation to a transport the task does not allow gives `not_run` with `constraints_unsatisfied`, both with zero contacts (no dial). Execution deadline availability is not asserted.
  - REQ-HTTP-01-0028: Through the `TestMHttp01Limits` subtest, for bodies around the 4, 4096 and 65536 byte bounds the run completes with `expected_response`, records `bytes_processed` as the smaller of body size and `max_body_bytes`, and sets `truncated` only above the bound. A gate on a small-response marking is not asserted; registrations carry none (`RV-M-HTTP-01-PRECONDITIONS`), see docs/measurement-requirements.md.
  - REQ-HTTP-01-0029: Tasks adding a header check or body check that is not registered with the target are rejected at admission with no dial. Through the `TestMHttp01Redirects` subtest, a redirect to an unregistered target, or to a hostname target on the proxy path, ends `completed` with `redirect_rejected` after one hop, while registered targets are followed up to `max_redirects` with one dial per hop.
  - REQ-HTTP-01-0030: Not asserted by this case; the library's HTTP stack attaches no credential, cookie or identifying header, so the cannot-be-disabled row does not arise in the portable core, and recorded requests are checked in `M-HTTP-01-SAFETY`, see docs/measurement-requirements.md.
  - REQ-HTTP-01-0031: The shared portable gates assert, each with zero contacts: a task transport condition the network does not meet, a task excluding VPN on a VPN-detected network and a local charging-only device restriction defer the run at `StartRun` (`run_not_due`, no result); a transport condition lost after reservation gives `not_run` with `constraints_unsatisfied`; and a proxy-path task with no proxy configured is either rejected at admission with `task_schema_invalid` or ends `not_run` with `proxy_not_configured` (the test accepts either); a deployment that withholds the target registration rejects the task with `target_not_registered`, and a registration that lapses after admission gives a result with reason `local_policy_rejected`.

<a id="M-HTTP-01-EXECUTION"></a>

### M-HTTP-01-EXECUTION

- Case: `M-HTTP-01-EXECUTION`
- Result: passed
- Tests: `TestMHttp01Execution` in `http01_conformance_test.go` (package `root`); `TestMHttp01DNSStages` in `http01_conformance_test.go` (package `root`); `TestMHttp01Redirects` in `http01_conformance_test.go` (package `root`); `TestMHttp01PrecedingFailures` in `http01_conformance_test.go` (package `root`); `TestMHttp01Limits` in `http01_conformance_test.go` (package `root`); `TestHTTPRedirectUsesProductionDestinationGate` in `internal/execution/http_primitive_test.go` (package `internal/execution`); `TestHTTPApprovedRedirectChain` in `internal/execution/http_primitive_test.go` (package `internal/execution`)
- Reproduce the primary test: `go test -count=1 -run '^TestMHttp01Execution$' ./.`
- Scope: Record system DNS, TCP, authenticated TLS, ordinary bodyless credential-free GET requesting identity encoding, bounded redirect decision and terminal checks/body close. Validate each next target and resolved address before contact; no downgrade, unsafe redirects, excessive redirect contact or test-layer retries.
- Requirement aspects exercised:

  - REQ-HTTP-01-0005: Direct and CONNECT-proxy runs send one GET with no body, Accept-Encoding identity and no User-Agent, Cookie or Authorization header, validate the chain against the synthetic CA and complete with expected_response; following is used only when enabled and only for registered targets within max_redirects.
  - REQ-HTTP-01-0006: body.bytes_processed counts application body bytes up to max_body_bytes, independently of TLS records and socket buffering (TestMHttp01Limits).
  - REQ-HTTP-01-0007: A hostname URL is resolved once through the system resolver for ipv4 and dns_query precedes tcp_connect, tls_handshake and http_exchange; a resolver timeout is dns_failure and an answer containing a loopback address is incomplete local_policy_rejected, both with no contact and no second lookup; a refused connect is tcp_failure with one observation and an untrusted chain tls_failure with two; a followed redirect makes one new connection per hop with no retry, and over the CONNECT proxy no direct dial occurs.
  - REQ-HTTP-01-0051: Through the CONNECT proxy, a redirect to a hostname URL is redirect_rejected with only the first hop observed (TestMHttp01Redirects); at the primitive level a hostname redirect on the proxy path is rejected before any resolver call or second connection (TestHTTPRedirectUsesProductionDestinationGate).

<a id="M-HTTP-01-NORMALIZATION"></a>

### M-HTTP-01-NORMALIZATION

- Case: `M-HTTP-01-NORMALIZATION`
- Result: passed
- Tests: `TestMHttp01Normalization` in `http01_conformance_test.go` (package `root`); `TestMHttp01CheckSemantics` in `http01_closeout_test.go` (package `root`); `TestMHttp01PrecedingFailures` in `http01_conformance_test.go` (package `root`); `TestMHttp01DNSStages` in `http01_conformance_test.go` (package `root`); `TestMHttp01Redirects` in `http01_conformance_test.go` (package `root`); `TestMHttp01Termination` in `http01_conformance_test.go` (package `root`); `TestMHttp01HeaderBoundary` in `http01_conformance_test.go` (package `root`); `TestMHttp01NormalizationInconclusive` in `internal/protocol/http/http01_inconclusive_test.go` (package `internal/protocol/http`); `TestMHttp01InconclusiveResultShape` in `http01_closeout_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMHttp01Normalization$' ./.`
- Scope: Exercise every outcome row plus configured-check combinations and partial/unevaluated checks, body completion versus truncation and unsupported content encoding. Preserve failure at the observed stage and indexed decidable check outcomes; do not invent verdicts for checks not evaluated or select undocumented precedence. The inconclusive outcome cannot arise through the runner, because crypto/tls never returns a zero-byte read without an error; it is exercised at the bounded HTTP exchange primitive and by decoding the registered result shape.
- Requirement aspects exercised:

  - REQ-HTTP-01-0008: expected_response is selected only when the status and every configured header and body check match; any single mismatch among several checks selects the failing class.
  - REQ-HTTP-01-0009: Real loopback exchanges produce dns_failure, tcp_failure, tls_failure, http_timeout at the 15-second bound, protocol_error for a malformed response, a premature body and 16,385 header bytes, redirect_observed, redirect_rejected for an http or unregistered target, redirect_limit_reached, unexpected_status, header_check_failed, body_check_failed, body_limit_reached and unsupported_content_encoding.
  - REQ-HTTP-01-0036: Every outcome row except inconclusive is produced through the Client with its terminal observation and no retained URL, header, body or raw error; inconclusive is produced by the exchange primitive from an inconclusive body read (TestMHttp01NormalizationInconclusive), and the bound result decoder accepts the registered inconclusive shape and rejects it with a timeout failure (TestMHttp01InconclusiveResultShape).
  - REQ-HTTP-01-0037: Results retain the indexed outcome of each decided status, header and body check and omit checks left undecidable by truncation or by an inconclusive body read; a decided mismatch is retained while an undecidable check beside it is omitted.
  - REQ-HTTP-01-0052: With a body check configured, unsupported content encoding wins over a failed status; a premature body (protocol_error) and an inconclusive body read win over a failed status; otherwise status precedes header, header precedes body and a decided body mismatch precedes the body limit: a 503 with a truncated body and an undecidable suffix is unexpected_status, a truncated body with a decided prefix is expected_response and an undecidable suffix is body_limit_reached. At the exchange layer no further read or connection follows the inconclusive read.

<a id="M-HTTP-01-LIMITS"></a>

### M-HTTP-01-LIMITS

- Case: `M-HTTP-01-LIMITS`
- Result: passed
- Tests: `TestMHttp01Limits` in `http01_conformance_test.go` (package `root`); `TestMHttp01HeaderBoundary` in `http01_conformance_test.go` (package `root`); `TestMHttp01Redirects` in `http01_conformance_test.go` (package `root`); `TestMHttp01OuterBoundsBoundary` in `http01_closeout_test.go` (package `root`); `TestMHttp01OuterBudgets` in `http01_conformance_test.go` (package `root`); `TestMHttp01EquivalentURLsSerialize` in `http01_conformance_test.go` (package `root`); `TestMHttp01EquivalentSpellingSerialized` in `http01_closeout_test.go` (package `root`); `TestMHttp01Termination` in `http01_conformance_test.go` (package `root`); `TestCcLimitPurposeHTTPRows` in `internal/execution/http_primitive_test.go` (package `internal/execution`)
- Reproduce the primary test: `go test -count=1 -run '^TestMHttp01Limits$' ./.`
- Scope: Test 16,384 response-header bytes, configured body bound and at most two followed redirects within outer connection/byte/time bounds at below/equal/above boundaries. Verify close at the active bound, no test retries and URL deduplication; distinguish application bytes processed from lower-layer buffering/on-wire use.
- Requirement aspects exercised:

  - REQ-HTTP-01-0032: Response headers of 16,383 and 16,384 bytes complete and 16,385 is protocol_error; bodies of 0 to 65,537 bytes against bounds 4, 4,096 and 65,536 report bytes_processed min(size, bound) and truncation exactly above the bound; max_redirects 1 and 2 follow at most that many registered hops and then redirect_limit_reached, with one dial per hop. On the direct and CONNECT paths, max_bytes_sent, max_bytes_received and max_connections at measured usage +1 and equal complete and at -1 stop with incomplete resource_limit, the CONNECT exchange not being charged; a run completes before the 30-second outer deadline and expires at it. Equivalent and differently spelled equivalent URLs wait without contact and are not_run when cancelled, and no test-level retry occurs.

<a id="M-HTTP-01-TERMINATION"></a>

### M-HTTP-01-TERMINATION

- Case: `M-HTTP-01-TERMINATION`
- Result: passed
- Tests: `TestMHttp01Termination` in `http01_conformance_test.go` (package `root`); `TestHttp01AllTerminationBoundaries` in `internal/execution/http01_termination_test.go` (package `internal/execution`); `TestHttp01RedirectAndHopTerminationPoints` in `internal/execution/http01_termination_points_test.go` (package `internal/execution`); `TestMHttp01Execution` in `http01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMHttp01Termination$' ./.`
- Scope: Drive HTTP-01 with a synthetic registered HTTPS target, a scripted resolver/TCP/TLS/HTTP transport, a fake network-generation source and a deterministic clock, and inject each interruption at each Procedure point: after input validation and before DNS (step 1-2), during the system DNS query (step 3), after DNS before TCP, during TCP connect and during the TLS handshake (step 4), after the GET is sent while awaiting response headers and during header read (step 5), after a redirect response is received with following disabled and with following enabled before next-URL validation (steps 6-7), during each redirect hop's DNS, TCP, TLS and HTTP stages for max_redirects 1 and 2, and during terminal-response body reading below max_body_bytes (step 8). Expected states follow the GO-SPEC-ISSUE-004 disposition: after input validation and before the test's first network-state observation or DNS query the attempt is not_run with that reason, no measurement and no network_context; every other listed point precedes the terminal outcome and is incomplete as asserted below. Inject in turn a network-generation change, host cancellation, outer attempt deadline expiry, a non-time outer connection or data bound, platform suspension and, where the host has configured a proxy path, proxy_unavailable, proxy_connect_failed and proxy_reconfigured. At the incomplete points, assert execution.state is incomplete with exactly the first matching reason from the common table (network_changed, cancelled_by_host, attempt_deadline_exceeded, resource_limit, platform_suspended, proxy_unavailable, proxy_connect_failed, proxy_reconfigured), measurement is absent at those points because HTTP-01 does not register bounded partial output, and no HTTP outcome (including network_changed, http_timeout or inconclusive) or check outcome is invented. Verify a bounded protocol timeout before the outer deadline still yields state completed with outcome http_timeout, that an uninterrupted run yields completed with measurement, that the response body and connections are closed, and that the recording transport shows no retry, fallback stack, further redirect hop or new connection after the interruption. Requirements: REQ-HTTP-01-0038.
- Requirement aspects exercised:

  - REQ-HTTP-01-0038: For host cancellation, network change, attempt deadline, resource_limit, injected platform suspension, proxy_unavailable, proxy_connect_failed and proxy_reconfigured: before the test starts the attempt is not_run with no network_context; during DNS, after DNS before TCP, during the connect, the TLS handshake, header and body reads, after a redirect response with following disabled or enabled, and during each redirect hop's DNS, TCP, TLS and header stage for max_redirects 1 and 2, it is incomplete with exactly the first reason and no measurement, no further dial or lookup and every opened connection closed; during cleanup after the outcome is recorded it stays completed with expected_response. Through the Client host cancellation and the outer deadline are incomplete while the 15-second protocol bound is completed http_timeout, and an uninterrupted run completes (TestMHttp01Execution).

<a id="M-HTTP-01-MINIMIZATION"></a>

### M-HTTP-01-MINIMIZATION

- Case: `M-HTTP-01-MINIMIZATION`
- Result: passed
- Tests: `TestMHttp01Minimization` in `http01_conformance_test.go` (package `root`), `TestMHttp01Safety` in `measurement_safety_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMHttp01Minimization$' ./.`
- Scope: Exercise HTTP-01 against scripted transports with a unique synthetic canary per prohibited item: requested URL, redirect Location URL, hostname, path, query, resolved destination address, request header value, response header name, response header value, request-side configured header_checks value, response body bytes, expected sha256 body digest in body_checks, computed observed body digest, matched contains/prefix/suffix strings, a Set-Cookie value, a credential in a WWW-Authenticate challenge and in redirect userinfo, certificate subject/issuer/serial, a cached response body, and raw transport/TLS/HTTP error text. Cover expected_response, redirect_observed, redirect_rejected, redirect_limit_reached, unexpected_status, header_check_failed, body_check_failed, body_limit_reached, unsupported_content_encoding and DNS/TCP/TLS failure paths. Assert the serialized result, every nested observation and every error/task-validation path contain no canary in any encoding and no field beyond the registered set: outcome, observations[].type, status, duration_ms, failure; dns_query engine, query_type, responses; tcp_connect address_family; tls_handshake address_family, tls_version, alpn, certificate_validation, alert; http_exchange address_family, protocol, status_code, redirect.followed_count, redirect.terminal_location_class, body.bytes_processed, body.truncated, checks[].kind/index/matched and content_encoding. Requirements: REQ-HTTP-01-0039.
- Requirement aspects exercised:

  - REQ-HTTP-01-0039: Unique canaries cover the URL hostname (also the Host header, the only task-derived request header value), path and query, the resolved address, a header-check name and value, a response header name and value, the response body, the expected and observed body digests, contains, prefix and suffix strings, a Set-Cookie value, a WWW-Authenticate realm and credential, the redirect userinfo, host, path, query and fragment, a registered redirect path and query, the certificate subject, issuer and serials, and raw transport, x509, TLS and status-text errors. Over `expected_response`, `redirect_observed`, `redirect_rejected`, `redirect_limit_reached`, `unexpected_status`, `header_check_failed`, `body_check_failed`, `body_limit_reached`, `unsupported_content_encoding`, `dns_failure`, `tcp_failure` and `tls_failure`, no canary appears in any result, the bundle, the pending store's bytes, any returned error or the task-validation rejection of a task whose unknown input member carries the canaries, every measurement path is within the registered HTTP-01 set, and the strict decoder rejects 16 injected members (including `url`, `location`, `headers`, `body`, `digest`, `cookie`, `credential`, `certificate` and `error`) in the last observation, its details, its body object and its first check. Cache contents and request bodies are not seeded with canaries, and no incomplete path is run.

<a id="M-HTTP-01-INTERPRETATION"></a>

### M-HTTP-01-INTERPRETATION

- Case: `M-HTTP-01-INTERPRETATION`
- Result: passed
- Tests: `TestMHttp01Interpretation` in `http01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMHttp01Interpretation$' ./.`
- Scope: Assert the HTTP-01 Output schema, serialized results for every registered outcome and every documented field description and sanitized example or host-facing text contain no field outside outcome and observations, no interpretation, verdict, blocking, censorship, block-page or manipulation field, and none of these prohibited claims: that the test browsed or rendered a page; that it authenticated, submitted user data or executed scripts; that an unexpected status, redirect_observed, redirect_rejected, header_check_failed or body_check_failed outcome establishes a block page, intentional manipulation or blocking; that body.bytes_processed or max_body_bytes is an exact on-wire byte limit; or that a result excludes origin, CDN, DNS, routing, certificate or deployment failure, intentional redirect or changed content, captive portal, enterprise proxy, VPN, antivirus or local filtering, server-side geolocation, rate limiting, load balancing or maintenance, IPv4/IPv6 or transient mobile-network differences, stale expected status, body pattern, digest or header expectation, or legitimate redirect-chain changes. Verify each documented outcome is described only as whether the endpoint or approved terminal redirect target responded and whether bounded status, header and body checks matched. Requirements: REQ-HTTP-01-0001, REQ-HTTP-01-0002, REQ-HTTP-01-0003, REQ-HTTP-01-0040, REQ-HTTP-01-0041, REQ-HTTP-01-0042, REQ-HTTP-01-0043. Judgement-based completion is the separate review case RV-M-HTTP-01-INTERPRETATION; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-HTTP-01-0001: No non-negated sentence in the comments and string literals of production Go files and sanitized examples and the shipped Markdown documentation (case plans, manifests, evidence, changelog and specification-issue records excluded) matches the "HTTP-01 browsed or rendered a page" claim pattern, and the HTTP-01 section of the interpretation guidance contains "does not browse or render a page". This is a textual scan of shipped text; it does not prove how a reader interprets a result.
  - REQ-HTTP-01-0002: The same scan finds no non-negated sentence stating that HTTP-01 authenticates, submits data or executes or runs scripts, and the guidance contains "does not authenticate, submit user data or execute scripts"; the requests actually sent are not checked here.
  - REQ-HTTP-01-0003: The same scan finds no non-negated sentence stating that an `unexpected_status`, redirect, header-check or body-check outcome establishes blocking, manipulation or censorship, or that a block page was detected, identified, found or observed, the guidance contains "does not establish a block page, intentional manipulation or blocking", and no member name or string value of any run's measurement contains a verdict token (such as `verdict`, `censor`, `block`, `filter`, `anomal`, `confiden`) and the strict result decoder rejects `verdict`, `interpretation`, `censored`, `blocked`, `filtered`, `anomaly`, `confidence`, `reachable` and `authoritative` fields injected into a valid result.
  - REQ-HTTP-01-0040: The guidance contains "whether the configured endpoint or an approved terminal redirect target responded" and "whether its bounded status, header and body checks matched", and each of the twelve runs records its expected outcome with a measurement holding only `outcome` and `observations`.
  - REQ-HTTP-01-0041: The same scan finds no non-negated sentence stating that `bytes_processed` or `max_body_bytes` is an exact on-wire limit, and the guidance contains "are not an exact on-wire limit"; a `body_limit_reached` run is included, but on-wire bytes are not measured.
  - REQ-HTTP-01-0042: Not asserted: the test requires none of the listed evidence items in the guidance and checks none of them.
  - REQ-HTTP-01-0043: Each run's measurement holds only `outcome` and `observations`, so no classification member is present, no sentence claims that redirect or response-difference outcomes establish blocking, and no member name or string value of any run's measurement contains a verdict token (such as `verdict`, `censor`, `block`, `filter`, `anomal`, `confiden`) and the strict result decoder rejects `verdict`, `interpretation`, `censored`, `blocked`, `filtered`, `anomaly`, `confidence`, `reachable` and `authoritative` fields injected into a valid result; that a manual or separately reviewed classification is required is not checked.

<a id="M-HTTP-01-SAFETY"></a>

### M-HTTP-01-SAFETY

- Case: `M-HTTP-01-SAFETY`
- Result: passed
- Tests: `TestMHttp01Safety` in `measurement_safety_conformance_test.go` (package `root`); `TestMHttp01EquivalentURLsSerialize` in `http01_conformance_test.go` (package `root`); `TestMHttp01Limits` in `http01_conformance_test.go` (package `root`); `TestMHttp01Minimization` in `http01_conformance_test.go` (package `root`); `TestMHttp01Redirects` in `http01_conformance_test.go` (package `root`); `TestMHttp01EquivalentSpellingSerialized` in `http01_closeout_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMHttp01Safety$' ./.`
- Scope: Exercise task safety evaluation and execution for HTTP-01 using synthetic registered targets, a fake local policy and host consent source, and recording transports. Iterate Baseline Risk rows: risk level low (task declaring low, medium, high and unknown yields effective risk low, medium, high and unknown respectively, never below low; a task with a factor more restrictive than its overall risk_level or an unknown factor type is rejected or treated as unknown) and requires user consent false subject to host and local policy (task false with a host policy requiring consent and no consent yields not_run local_policy_rejected with no network contact; task true without valid consent does not execute). Iterate each Risk Factors row: network_observability (recording transport shows only the registered origin over authenticated TLS, no plaintext HTTP), traffic_distinctiveness (one ordinary bodyless GET with identity encoding, no measurement identifiers or unusual headers), target_sensitivity (unregistered, private or local target rejected before DNS), data_sensitivity (body read stops at max_body_bytes and is not retained) and legal_or_social_context (low baseline cannot be lowered by the task). Exercise each Safety Analysis action: data minimization (no body, full headers, matched strings, cookies, full redirect URL or unrelated certificate metadata in output), traffic distinctiveness (no repeated redirect chain; equivalent runs for one normalized URL deduplicated or delayed), user-identification risk (synthetic cookie jar, credential store, user-state cache and app identifier are present yet never sent or used), target risk (unregistered or non-HTTPS redirect from a registered URL yields redirect_rejected and is not contacted) and probe final authority (fake local policy delay, skip and reject each prevent contact). Requirements: REQ-HTTP-01-0044, REQ-HTTP-01-0045, REQ-HTTP-01-0046, REQ-HTTP-01-0047, REQ-HTTP-01-0048, REQ-HTTP-01-0049, REQ-HTTP-01-0050. Judgement-based completion is the separate review case RV-M-HTTP-01-SAFETY; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-HTTP-01-0044: The test asserts that the production registry records a `low` baseline for HTTP-01 and that a task declaring `low`, `medium`, `high` or `unknown` has effective risk exactly that level (admitted under issuer and local bounds permitting it, rejected `task_issuer_not_authorized` with no contact when either side permits only lower levels). For the consent baseline it asserts that `false` adds no requirement (with host consent absent the run starts, yields one result and never queries the consent capability), while consent required by the task, by an issuer or by the local policy blocks every result and contact unless the fake host reports valid consent (absent, ended and unavailable all block).
  - REQ-HTTP-01-0045: For each of the five factor types the test asserts that the registered baseline equals the most restrictive of the fixture's row levels (all five `low`), that declaring the factor at `low`, `medium` or `high` gives effective risk exactly the maximum of baseline, row level and declared level, that an `unknown` factor gives `unknown`, and that an unknown factor under a known overall level or a factor above the overall level is rejected `task_schema_invalid` and an unrecognized factor type `task_unsupported`, all with no contact. The row levels are fixture constants checked only against the registered baseline, and the rationale text is not asserted; per-factor action rows additionally assert that a plaintext `http://` URL and an `https` URL on port 80 are rejected with no contact and a run makes one TLS connection with one ClientHello (`network_observability`), one bodyless GET (`traffic_distinctiveness`), rejection of unregistered, private, loopback and `localhost` targets with no dial (`target_sensitivity`), and `bytes_processed` equal to the smaller of body size and `max_body_bytes` with `truncated` set only above the bound (`data_sensitivity`).
  - REQ-HTTP-01-0046: The test asserts that no surface contains the requested host, URL path and query, resolved address, header-check and response header names and values, body bytes, body digests, contains/prefix/suffix strings, `Set-Cookie` value, `WWW-Authenticate` realm and credential, redirect userinfo, host, path, query and fragment, certificate fields or raw errors, and that the measurement holds only the closed field set. The decoder rejects injected URL, location, headers, body, digest, value, cookie, credential, certificate and error fields.
  - REQ-HTTP-01-0047: The test asserts one bodyless GET to `/check` over one connection, with `Accept-Encoding` absent or `identity`, none of the listed identifying headers and no header value containing `safety-` or `barometer`, and that a second run for the same URL or an equivalent spelling makes no dial or request while the first is in progress. Redirect chains are bounded by the redirect rows (the configured limit ends with `redirect_limit_reached` and the exact number of dials).
  - REQ-HTTP-01-0048: The test asserts that after a response setting a cookie, the next request does not return it, that neither request carries `Cookie`, `Authorization`, `Proxy-Authorization`, cache validators or the listed application headers, and that a server requesting a client certificate receives none. No synthetic credential store, user-state cache or application identifier is installed, so those are covered only by the absent headers.
  - REQ-HTTP-01-0049: The test asserts that unregistered addresses and paths and private, loopback and `localhost` targets are rejected at admission with no dial, that loopback, private and link-local registered addresses are refused or never contacted, that a redirect to an unregistered target or to a hostname on the proxy path ends `redirect_rejected` without contacting it, and that registered redirects are followed only within `max_redirects`. A non-HTTPS redirect is not exercised, repeated execution is covered only by equivalent-URL serialization, and the residual-risk statement is not asserted.
  - REQ-HTTP-01-0050: The test asserts that a local policy not permitting the task's network path rejects a valid task at admission with no contact, and that a local charging-only device constraint holds an admitted valid task (`run_not_due` twice, no result, no contact) until satisfied, after which the same task runs; a target registration that expires after admission ends the attempt `local_policy_rejected` with no contact. The no-lowering clause is asserted through the declared-level rows: no task declaration yields effective risk below the `low` baseline.

<a id="M-HTTP-01-OUTPUT"></a>

### M-HTTP-01-OUTPUT

- Case: `M-HTTP-01-OUTPUT`
- Result: passed
- Tests: `TestMHttp01Output` in `http01_conformance_test.go` (package `root`); `TestMHttp01Execution` in `http01_conformance_test.go` (package `root`); `TestMHttp01DNSStages` in `http01_conformance_test.go` (package `root`); `TestMHttp01Redirects` in `http01_conformance_test.go` (package `root`); `TestMHttp01PrecedingFailures` in `http01_conformance_test.go` (package `root`); `TestMHttp01ResultVersionsAndClosedOutput` in `http01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMHttp01Output$' ./.`
- Scope: Validate ordered observations only for performed DNS/TCP/TLS/HTTP operations, including bounded redirect chains and connection-reuse omissions. Terminal HTTP retains only status, minimized redirect class, bounded body metadata and indexed checks; no request/response content.
- Requirement aspects exercised:

  - REQ-HTTP-01-0034: An IP-literal URL yields tcp_connect, tls_handshake and http_exchange in order with no dns_query; a hostname URL adds one dns_query first; a TCP failure yields only tcp_connect and a TLS failure no http_exchange; each followed hop adds its own TCP, TLS and HTTP observations, since no connection is reused.
  - REQ-HTTP-01-0035: The terminal http_exchange keeps the status, body metadata and one indexed result per configured check, and the result contains no URL, path, address, header name or value, body or certificate canary; the decoder rejects added body_content, expected_digest, url, header_value, debug_log and censorship details.

<a id="M-HTTP-01-VERSIONS"></a>

### M-HTTP-01-VERSIONS

- Case: `M-HTTP-01-VERSIONS`
- Result: passed
- Tests: `TestMHttp01Versions` in `http01_conformance_test.go` (package `root`); `TestMHttp01TaskVersionShapes` in `http01_closeout_test.go` (package `root`); `TestMHttp01ResultVersionShapes` in `http01_closeout_test.go` (package `root`); `TestMHttp01ResultVersionsAndClosedOutput` in `http01_conformance_test.go` (package `root`); `TestMHttp01Output` in `http01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMHttp01Versions$' ./.`
- Scope: Exercise task acceptance and result validation with the single registered combination test.id HTTP-01, test.version 0.1.0, schema_version 0.1.0 and test.result_version 0.1.0, and assert that an emitted result carries exactly that combination. Iterate unsupported and malformed values for each field independently: 0.2.0, 1.0.0, 0.1, v0.1.0, 0.1.0 with trailing whitespace, empty string, numeric 0.1, null and missing. Iterate cross-test combinations: test.id HTTP-01 with a measurement shaped as the NET-01 access-only object or carrying NET-05 profile/step fields, NET-03 or TLS-01 test.id with an HTTP-01 measurement object, and lowercase http-01. Assert task-side unsupported versions are task-validation failures that produce no Measurement Result and no network contact on the recording transport, and that a receiver rejects or quarantines unsupported result versions without best-effort parsing, coercion to 0.1.0, dropping unknown fields or adding another version; Draft status confers no support. Requirements: REQ-HTTP-01-0033.
- Requirement aspects exercised:

  - REQ-HTTP-01-0033: An emitted result carries schema_version, test.version and test.result_version 0.1.0 with test.id HTTP-01. Task test_version 0.2.0, 1.0.0, 0.1, v0.1.0, padded, empty, numeric, null and missing, and test_id http-01, NET-03, TLS-01, NET-01 and NET-05, are rejected at admission with no transport contact (the numeric, null, tab-padded, missing and cross-test cases also assert no handle or replay commit). The decoder rejects 0.2.0, 1.0.0, 0.1, v0.1.0, padded, empty, numeric and null values of each version field, a missing schema_version, test.version, result_version or test.id, test IDs NET-01, NET-03, NET-05, TLS-01 and http-01, a NET-01 access object or access-only measurement and NET-05 profile, steps or network_validation fields.

<a id="M-NET-05-ADMISSION"></a>

### M-NET-05-ADMISSION

- Case: `M-NET-05-ADMISSION`
- Result: passed
- Tests: `TestMNet05Admission` in `net05_conformance_test.go` (package `root`); `TestMNet05ProfileVersionWithoutHistory` in `net05_conformance_test.go` (package `root`); `TestMNet05AtomicTargetAndProxyRules` in `net05_conformance_test.go` (package `root`); `TestMNet05CodecBoundsAndComposition` in `net05_conformance_test.go` (package `root`); `TestMNet05Versions` in `net05_conformance_test.go` (package `root`); `TestMNet05RegisteredCeilingAdmission` in `net05_conformance_test.go` (package `root`); `TestNet05StepTargetAndRedirectIsolation` in `internal/execution/net05_conformance_test.go` (package `internal/execution`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet05Admission$' ./.`
- Scope: Validate the outer profile, profile_id/version, each step_id/test_id/exact version/input/options and all nested closed schemas. Test 0/1/8/9 steps, identifier lengths 0/1/64/65, duplicate IDs, nonpositive versions, equal or lower versions for one profile_id admitted without history, unsupported/nested tests, forbidden nested envelopes and interpolation. Put an invalid or unsafe final step after valid earlier steps and assert zero primitive operations.
- Requirement aspects exercised:

  - REQ-NET-05-0005: Profiles of 1, 4 and 8 primitive steps, each with its own target in its primitive input, are admitted through `Client` with one replay commit, no contact and the NET-05 runner installed.
  - REQ-NET-05-0014: Empty input, null input and a null profile are rejected with no handle, replay commit or contact.
  - REQ-NET-05-0016: Zero and nine steps are rejected; step_id lengths 1 and 64 are accepted and 0, 65, non-ASCII, control and interpolated identifiers rejected; a step missing test_version, input or options is rejected and nested step options keep the common 8,192-byte cap (TestMNet05CodecBoundsAndComposition); the normalized step container is immutable.
  - REQ-NET-05-0017: Only an inline profile object is accepted; null and missing profiles are rejected.
  - REQ-NET-05-0018: A profile whose final step has a credential URL or whose second step targets a loopback address is rejected with no handle, run or contact; a changed final-step registration, a registration for NET-05 instead of the primitive, a proxy-only path, an expired registration or an issuer missing one step's registration rejects the whole profile as target_not_registered with no contact (TestMNet05AtomicTargetAndProxyRules).
  - REQ-NET-05-0019: A loopback TCP-01 target and a credential HTTP-01 URL are rejected by their primitive validation inside the profile; each step is matched against its own primitive registration and authority (TestMNet05AtomicTargetAndProxyRules), and each step gate grants only its own registered tuple, never another step's redirect link (TestNet05StepTargetAndRedirectIsolation).
  - REQ-NET-05-0020: A profile repeating a step_id is rejected.
  - REQ-NET-05-0021: A step whose test_id is NET-05, NET-01, NET-03, DNS-03, SYN-01 or tcp-01, after a valid step, rejects the whole profile as task_schema_invalid with no handle, replay commit or contact (TestMNet05Versions).
  - REQ-NET-05-0022: Steps containing schedule, execution, safety, labels, script or conditions members are rejected.
  - REQ-NET-05-0023: Credential URLs, loopback targets, interpolated identifiers and more than eight steps are rejected.
  - REQ-NET-05-0024: An unknown profile member and a duplicated profile_version member are rejected.
  - REQ-NET-05-0025: Through `Client` a 49,140-byte profile, the largest the common 49,152-byte input bound admits with its 12-byte wrapper, is admitted with one replay commit and no contact, and profiles of 49,141, 49,151, 49,152 and 49,153 bytes are task_limit_exceeded; the codec alone accepts encoded profiles up to 49,152 bytes and rejects a profile over its own cap (TestMNet05CodecBoundsAndComposition).
  - REQ-NET-05-0026: Profiles whose summed declared maxima exceed the task connection or received-byte bound or the composite connection ceiling are task_limit_exceeded through `Client` with no handle, replay record, run or contact, and the equal profiles are admitted (TestMNet05RegisteredCeilingAdmission); the boundaries are recorded under M-NET-05-LIMITS.
  - REQ-NET-05-0027: Every rejection returns no handle and leaves no task handle, replay commit, run or transport contact.
  - REQ-NET-05-0062: profile_id lengths 1 and 64 are accepted and 0 and 65, non-ASCII, control and interpolated identifiers rejected; profile_version 0, -1, 1.5, a string and null are rejected, and positive versions up to the signed 64-bit maximum are admitted (TestMNet05ProfileVersionWithoutHistory).
  - REQ-NET-05-0063: In one client instance, profile_version 5, 3, 5, 1, the signed 64-bit maximum and 2 for the same profile_id are each admitted through `Client` and retained unchanged in the normalized profile, with six task handles and replay commits and no contact; no version is compared with another task's (TestMNet05ProfileVersionWithoutHistory).

<a id="M-NET-05-OPTIONS"></a>

### M-NET-05-OPTIONS

- Case: `M-NET-05-OPTIONS`
- Result: passed
- Tests: `TestMNet05Options` in `net05_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet05Options$' ./.`
- Scope: Accept only empty composite options. Independently vary each primitive option via its exact versioned schema; reject attempts to override sequence, outer execution/safety, primitive limits or no-retry behavior.
- Requirement aspects exercised:

  - REQ-NET-05-0028: An empty composite options object is admitted.
  - REQ-NET-05-0029: null, array and every composite member (sequence, retry, retries, execution, safety, limits, max_duration_ms, max_attempts, parallel, order, fallback, network_path and an empty name) are task_schema_invalid with no contact.
  - REQ-NET-05-0030: Each primitive option is checked by its own exact schema in a later step after a valid one: TCP-01 limit, duration, timeout and retry members, TLS-01 h3, duplicate and string offers, resumption and retries, DNS-01 doh, a non-boolean collection flag, a window without collection and max_attempts, and HTTP-01 max_redirects 3, wrong types and limit or retry members are rejected; limit, retry, order, parallel, network_path and risk_level members beside a step's input and options are rejected; max_body_bytes 65,537 and 4,096 against max_bytes_received 4,095 are task_limit_exceeded, and a narrower primitive option within the envelope is admitted without contact.

<a id="M-NET-05-PRECONDITIONS"></a>

### M-NET-05-PRECONDITIONS

- Case: `M-NET-05-PRECONDITIONS`
- Result: passed
- Tests: `TestMNet05Preconditions` in `net05_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet05Preconditions$' ./.`
- Scope: Run the portable NET-05 gate with a synthetic host policy, synthetic inline profile, fake platform and recording transport. Iterate single-precondition failures and assert execution.state not_run with no profile step started: task source or inline profile untrusted by host policy; any target failing local privacy and safety evaluation; NET-05 neither explicitly enabled by host policy nor triggered by repeated baseline anomalies under that policy; network not stable; validation or captive-portal context unavailable; outer execution envelope unable to accommodate the complete inline profile within primitive limits (for example step durations exceeding max_duration_ms). Assert envelope fit never weakens primitive limits. Inject host-policy withdrawal after step 1 and assert no further steps start. Assert a stricter execution constraint or local policy denial gives constraints_unsatisfied or local_policy_rejected; other denials use a registered reason per the public table. Requirements: REQ-NET-05-0031, REQ-NET-05-0032, REQ-NET-05-0033, REQ-NET-05-0034, REQ-NET-05-0035, REQ-NET-05-0036. Judgement-based completion is the separate review case RV-M-NET-05-PRECONDITIONS; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-NET-05-0031: A task embedded from an issuer the host policy does not trust is refused by `AdmitTask` with zero dials and system-resolver lookups. Through the `TestMNet05AtomicTargetAndProxyRules` subtest, a profile step target with a different URL, a different test, a proxy-only path, an expired registration or a registration the issuer lacks is rejected with `target_not_registered`, a step target registered above the permitted risk with `task_issuer_not_authorized`, and a proxy profile with a system or direct-UDP DNS step or a hostname HTTP step with `task_schema_invalid`, each with no published state or transport contact.
  - REQ-NET-05-0032: A deployment that lists only another test, no registrations and no composite permission rejects NET-05 with `task_issuer_not_authorized` and no contact, as does a deployment or issuer lacking a primitive test or composite permission. Replacing the policy with one that no longer enables NET-05 after step 1 completes and while step 2 is dialing ends the run `incomplete` with `cancelled_by_host`, with one dial, no HTTP request and one result; no anomaly-triggered path is asserted.
  - REQ-NET-05-0033: Through the `TestMNet05PlatformValidationGate` subtest, platform validation reported `no` or `unknown` gives `completed` with outcome `network_not_validated` or `inconclusive` (validation state `not_validated` or `unknown`), no steps and no dial, and a task requiring a validated network is deferred with `run_not_due`. Network stability is asserted only through the shared gates: an unreadable network access state at reservation defers the run (`run_not_due`, no result), and a network change after reservation to a transport the task does not allow gives `not_run` with `constraints_unsatisfied`, both with zero contacts.
  - REQ-NET-05-0034: Through the `TestMNet05RegisteredCeilingAdmission` subtest, task limits at or below the registered ceilings (`max_duration_ms` 60000, `max_connections` 12, `max_bytes_sent` 65536, `max_bytes_received` 262144) are admitted and one above is rejected with `task_limit_exceeded`, and a task bound below the summed declared step maxima for connections or received bytes is rejected the same way, all without transport contact. A profile whose step durations exceed `max_duration_ms` is not exercised.
  - REQ-NET-05-0035: Not asserted by this case; the traffic-pattern criterion is the review case `RV-M-NET-05-PRECONDITIONS`, see docs/measurement-requirements.md.
  - REQ-NET-05-0036: The shared portable gates assert, each with zero contacts: a task transport condition the network does not meet, a task excluding VPN on a VPN-detected network and a local charging-only device restriction defer the run at `StartRun` (`run_not_due`, no result); a transport condition lost after reservation gives `not_run` with `constraints_unsatisfied`; and a proxy-path task with no proxy configured is either rejected at admission with `task_schema_invalid` or ends `not_run` with `proxy_not_configured` (the test accepts either); a deployment that withholds the target registration rejects the task with `target_not_registered`, and a registration that lapses after admission gives a result with reason `local_policy_rejected`; contacts count both transport dials and system-resolver lookups.

<a id="M-NET-05-EXECUTION"></a>

### M-NET-05-EXECUTION

- Case: `M-NET-05-EXECUTION`
- Result: passed
- Tests: `TestMNet05ExecutionAndOutput` in `net05_runner_test.go` (package `root`); `TestNet05StepCountExecution` in `net05_runner_test.go` (package `root`); `TestNet05DirectDNSPrimitives` and `TestNet05TLSPrimitiveSuccess` in `net05_runner_test.go` (package `root`); `TestNet05WholeProfileSerialization` in `internal/execution/net05_conformance_test.go` (package `internal/execution`); `TestMNet05PrimitiveRiskAndMissingCapability` in `net05_conformance_test.go` (package `root`); `TestMNet05Admission` in `net05_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet05ExecutionAndOutput$' ./.`
- Scope: Record admission of the entire profile and effective risk before any connection, current NET-01, platform validation context only with source platform and no NET-03 request, and strictly sequential DNS-01/TCP-01/TLS-01/HTTP-01 dispatch. Each primitive must use its own schema and shared protocol operations, with no parallel steps, fallback, retries, target expansion or combined verdict.
- Requirement aspects exercised:

  - REQ-NET-05-0006: Through `Client` admission and `StartRun`, DNS-01, TCP-01, TLS-01 and HTTP-01 steps run in profile order with three dials and their own outcomes resolved, connected, alpn_no_overlap and expected_response; profiles of one and eight TCP-01 steps execute every step once, in order (TestNet05StepCountExecution); a direct DNS-01 step over UDP and over TCP contacts only the registered resolver once (TestNet05DirectDNSPrimitives); a TLS-01 step without an ALPN offer completes with one dial (TestNet05TLSPrimitiveSuccess); and primitive operations of one profile never overlap (TestNet05WholeProfileSerialization).
  - REQ-NET-05-0007: The run is admitted and started under the outer NET-05 task's schedule, execution envelope and safety declaration, and each step carries only step_id, test_id, test_version, input and options; steps carrying schedule, execution, safety, labels, script or conditions members are rejected at admission (TestMNet05Admission).
  - REQ-NET-05-0064: Validation context is the platform's validated state, and the run sends only the declared steps' traffic: three dials for the TCP-01, TLS-01 and HTTP-01 steps and one HTTP request, the HTTP-01 step's, so no `NET-03` request is sent to obtain context.
  - REQ-NET-05-0065: A primitive with a high baseline risk makes the profile task_issuer_not_authorized under a medium-risk policy and a missing primitive registry entry makes it task_unsupported, both before any replay commit or contact (TestMNet05PrimitiveRiskAndMissingCapability); an admitted profile reads the platform validation state, obtains or reuses no `NET-03` observation and dispatches each step in order through the shared protocol operations.

<a id="M-NET-05-NORMALIZATION"></a>

### M-NET-05-NORMALIZATION

- Case: `M-NET-05-NORMALIZATION`
- Result: passed
- Tests: `TestMNet05Normalization` in `internal/execution/net05_conformance_test.go` (package `internal/execution`); `TestMNet05PlatformValidationGate` in `net05_runner_test.go` (package `root`); `TestMNet05StepNegativeOutcomesStayCompleted` in `internal/execution/net05_closeout_test.go` (package `internal/execution`); `TestMNet05Termination` in `internal/execution/net05_conformance_test.go` (package `internal/execution`); `TestNet05TaskValidationBeforeStart` in `internal/execution/net05_conformance_test.go` (package `internal/execution`); `TestMNet05Admission` in `net05_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet05Normalization$' ./internal/execution`
- Scope: Exercise every composite failure/outcome row using positive and negative terminal primitive results. All completed primitive results yield completed composite execution; missing validation under requiring policy runs no primitive. Represent network/resource termination at the common execution layer without rewriting primitive failures.
- Requirement aspects exercised:

  - REQ-NET-05-0010: With validated platform context, three TCP-01 steps that are all refused complete the profile with every step present and completed; a five-step profile with TCP timeout, TLS timeout, untrusted certificate, refusal and success completes with each primitive outcome retained (TestMNet05StepNegativeOutcomesStayCompleted).
  - REQ-NET-05-0011: Not-validated platform context gives network_not_validated and unknown context gives inconclusive, both with no contact; completed is selected when every step reaches a terminal primitive result; network change, resource exhaustion and the other common reasons give outcome incomplete with the matching execution.reason (TestMNet05Termination).
  - REQ-NET-05-0012: Negative primitive outcomes (connection_refused, timeout, tls_timeout, certificate_untrusted) are kept unchanged inside completed steps and the composite outcome stays completed; no other composite classification is emitted.
  - REQ-NET-05-0013: Invalid profiles are rejected at admission with no handle, run, contact or result (TestMNet05Admission), and an unsatisfied task-level require_validated constraint is not_run constraints_unsatisfied with no measurement or contact (TestNet05TaskValidationBeforeStart).
  - REQ-NET-05-0046: completed is recorded with every step completed and execution.state completed, including with negative primitive outcomes; network_not_validated and inconclusive have no steps and no contact, also through the Client (TestMNet05PlatformValidationGate); incomplete is recorded with execution.state incomplete and the stopping reason.

<a id="M-NET-05-LIMITS"></a>

### M-NET-05-LIMITS

- Case: `M-NET-05-LIMITS`
- Result: passed
- Tests: `TestMNet05Limits` in `internal/execution/net05_conformance_test.go` (package `internal/execution`); `TestMNet05RegisteredCeilingAdmission` in `net05_conformance_test.go` (package `root`); `TestDeclaredMaximaFailClosed` in `internal/measurement/net05/codec_test.go` (package `internal/measurement/net05`); `TestMNet05AggregateBytes` and `TestMNet05AggregateDeadline` in `internal/execution/net05_conformance_test.go` (package `internal/execution`); `TestMNet05ConnectionCeiling`, `TestMNet05ReceivedCeiling` and `TestMNet05DurationCeiling` in `internal/execution/net05_closeout_test.go` (package `internal/execution`); `TestNet05WholeProfileSerialization` in `internal/execution/net05_conformance_test.go` (package `internal/execution`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet05Limits$' ./internal/execution`
- Scope: Test 49,152 encoded profile bytes, 60,000 ms, 12 connections/tunnels, 65,536 application bytes sent and 262,144 received at below/equal/above boundaries and against stricter primitive/task bounds. Reject over-budget declared maxima atomically before contact, with the composite duration and sent ceilings recorded as unreachable from declarations; enforce actual aggregate use and one sequential profile run per generation.
- Requirement aspects exercised:

  - REQ-NET-05-0026: Each primitive declares its registered maxima: DNS-01 system mode nothing, direct UDP 4,096 received bytes, or 1,000 ms and 12,288 bytes when collecting, direct TCP one connection and 65,537 bytes; TCP-01 and TLS-01 one connection; HTTP-01 one connection and 20,480 bytes by default, or three connections and 114,688 bytes with two redirects and a 65,536-byte body. Declared sums of 11, 12 and 13 connections and 262,143, 262,144 and 262,145 received bytes against the composite ceilings, and of 3 connections, 20,480 bytes and 3,000 ms against task bounds one above, equal and one below, are admitted at or under the bound and rejected above it; undeclared sent bytes add nothing even with max_bytes_sent 0, and a step without declarations fails closed. Through Client, profiles over the task connection or received-byte bound and a 13-connection profile over the composite ceiling are task_limit_exceeded with no handle, replay record, run or contact, and the equal profiles are admitted. The 60,000 ms and 65,536-byte sent ceilings cannot be reached from declarations (owner-selected rule, GO-SPEC-ISSUE-021).
  - REQ-NET-05-0037: The duration, connection and received-byte ceilings and stricter task deadlines stop the profile with resource_limit or attempt_deadline_exceeded against actual aggregate use on direct and CONNECT paths, usage never exceeds a limit, and a second profile on one runner waits and is reused only after the first completes. These run-time cases build their plans without admission and are unchanged.
  - REQ-NET-05-0038: max_duration_ms 60,000, max_connections 12, max_bytes_sent 65,536 and max_bytes_received 262,144 are admitted at the ceiling and one below and rejected one above at admission, each independently; the profile's declared sums are checked against the same envelope before execution.

<a id="M-NET-05-TERMINATION"></a>

### M-NET-05-TERMINATION

- Case: `M-NET-05-TERMINATION`
- Result: passed
- Tests: `TestMNet05Termination` in `internal/execution/net05_conformance_test.go` (package `internal/execution`); `TestNet05MixedPrimitiveTermination` in `internal/execution/net05_conformance_test.go` (package `internal/execution`); `TestMNet05StepNegativeOutcomesStayCompleted` in `internal/execution/net05_closeout_test.go` (package `internal/execution`); `TestNet05TaskValidationBeforeStart` in `internal/execution/net05_conformance_test.go` (package `internal/execution`); `TestMNet05Normalization` in `internal/execution/net05_conformance_test.go` (package `internal/execution`); `TestCcNet05ProxyLifecycleAndRenewal` in `net05_runner_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet05Termination$' ./internal/execution`
- Scope: Use a validated synthetic three-step profile (DNS-01, TCP-01, HTTP-01 steps with synthetic .test targets) driven by scripted per-primitive transports, a deterministic clock and a fake platform that can raise a network-generation change, host cancellation, outer attempt-deadline expiry, non-time connection/data bound exhaustion, platform suspension and proxy reconfiguration. Iterate each interruption point: (a) after NET-01/NET-03 context capture and before step 1; (b) before each later step k; (c) inside each started step k while its primitive operation is in flight; (d) after the final step reaches a terminal primitive outcome. Expected states follow the GO-SPEC-ISSUE-004 disposition: point (a) follows the stable-network check of Procedure step 3, so points (a) to (c) are incomplete as asserted below; a task-level require_validated constraint that is unsatisfied before that check is not_run with constraints_unsatisfied and no measurement; and the composite outcomes completed, network_not_validated and inconclusive are retained with execution.state completed while composite outcome incomplete keeps execution.state incomplete with the stopping reason. For each of the common reasons network_changed, resource_limit, attempt_deadline_exceeded, cancelled_by_host, platform_suspended and proxy_reconfigured, assert execution.state is incomplete with exactly that reason (only the first cross-cutting condition is recorded) and measurement.outcome is incomplete; for network_changed and resource_limit assert no network_changed or resource_limit outcome value is emitted, only outcome incomplete plus the matching execution.reason. Assert permitted partial profile output at (a)/(b): steps contains only the earlier steps, each with state completed and its primitive measurement, no placeholder entry for unstarted steps, and stopped_before_step equals the step_id of step k; at (c): step k is the single final entry with state incomplete and no measurement object, and stopped_before_step equals the next profile step_id when one remains and is absent when step k is the last; at (d): outcome completed, execution.state completed, no reason, no stopped_before_step. Assert a primitive protocol timeout, refusal or TLS failure scripted inside a step yields a completed step with the primitive negative outcome rather than an interruption, and that nested step measurement objects never carry their own common envelope, execution object or Network Context while one outer Network Context remains present. Verify after each interruption that the recording transport shows no further connection, retry, fallback to the direct path, new proxy endpoint or mixed proxy generation, that all sockets and timers are released, and that no terminal protocol outcome is invented for the interrupted or unstarted steps. Requirements: REQ-NET-05-0047, REQ-NET-05-0061.
- Requirement aspects exercised:

  - REQ-NET-05-0047: network_changed and resource_limit, like host cancellation, attempt deadline, injected suspension and the three proxy reasons, appear only as execution.reason with measurement.outcome incomplete; no step or composite outcome carries them.
  - REQ-NET-05-0061: For each common reason with a second cause injected after the first, interruptions before step 1, before later steps and inside each of three TCP-01 steps are incomplete with exactly the first reason and the partial steps and stopped_before_step expected, with no later dial and released connections; in a DNS-01, TCP-01, HTTP-01 profile, interruptions during resolution, TCP, the HTTP connect, TLS and the body read are incomplete with the interrupted step last and no leaked timer, while an interruption during final cleanup keeps completed. Primitive timeouts, a refusal and a certificate failure inside steps stay completed with one envelope, execution object and Network Context (TestMNet05StepNegativeOutcomesStayCompleted); network_not_validated and inconclusive are completed (TestMNet05Normalization); an unsatisfied require_validated constraint is not_run constraints_unsatisfied with no measurement (TestNet05TaskValidationBeforeStart); a proxy replacement or clearing during the HTTP step is incomplete proxy_reconfigured with no new proxy, direct fallback or later step (TestCcNet05ProxyLifecycleAndRenewal).

<a id="M-NET-05-MINIMIZATION"></a>

### M-NET-05-MINIMIZATION

- Case: `M-NET-05-MINIMIZATION`
- Result: passed
- Tests: `TestMNet05Minimization` in `net05_conformance_test.go` (package `root`), `TestMNet05Safety` in `measurement_safety_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet05Minimization$' ./.`
- Scope: Run a synthetic profile covering DNS-01, TCP-01, TLS-01 and HTTP-01 steps through scripted transports, inserting one unique canary per prohibited item named by the linked unit: an outer task schedule canary, an outer safety-declaration canary and an outer labels canary (nested task envelope, schedule, safety and labels must not be copied); a canary hostname and a canary HTTPS URL path in step inputs (duplicated targets and target URLs); a canary TCP-01 target address from 192.0.2.0/24 (IP addresses beyond bounded DNS answer values); a canary profile field value such as a step option value (profile contents); a canary response-body string and a canary HTTP header value returned by the scripted HTTP server; a canary certificate subject and serial from the synthetic TLS chain (certificate data); and a canary platform error string raised by the fake platform (platform errors). Assert serialized output, including every nested step measurement and observation and any result-validation error path, contains none of these canaries, no nested schedule, execution, safety or labels object, no per-step Network Context and no field named or valued as a combined restriction verdict (for example broad_restriction_suspected or allowlist_detected). Assert the only retained NET-05 fields are outcome, observations, profile.profile_id, profile.profile_version, network_validation.state, network_validation.source, steps[].step_id, steps[].test.id, steps[].test.version, steps[].test.result_version, steps[].state, steps[].measurement (the primitive Output object as registered by that primitive) and stopped_before_step when required; the DNS-01 step may retain bounded normalized DNS answer values (a synthetic 198.51.100.0/24 answer is present only in the DNS answers path). Reject any result carrying an extra field before bundling. No participant or live measurement data is used. Requirements: REQ-NET-05-0048.
- Requirement aspects exercised:

  - REQ-NET-05-0048: A five-step profile (direct DNS-01, failing system DNS-01, TCP-01, TLS-01, HTTP-01) and its outer task carry unique canaries for the schedule time, a label key and value, step hostnames, the URL path, the target address and port, the TLS server name, a step option value, the response body and a header value, the certificate subject, issuer and serials, platform and raw TCP errors, the DNS TTL, authority, additional and raw-message patterns, and the strings `broad_restriction_suspected` and `allowlist_detected`. For the completed run and a run cancelled during the HTTP step, no canary appears in any searched encoding in any result, the bundle, the pending store's bytes, any returned error or the task-validation rejection of a task whose unknown input member carries the canaries and every measurement path is within the NET-05 and primitive field sets; the completed run records step outcomes `resolved`, `resolver_error`, `connected`, `handshake_succeeded` and `expected_response` in order, retains the DNS answer `1.1.1.1` exactly once and only in the DNS answers path, and copies no `schedule`, `safety`, `labels`, `risk_level`, `requires_user_consent`, `input` or `options` member, no second `execution` and no per-step `network_context`, and the strict decoder rejects those members, `target`, `url`, `error`, `classification` or either verdict name injected into the measurement, a step, a step measurement or the profile, and either verdict name as the outcome.

<a id="M-NET-05-INTERPRETATION"></a>

### M-NET-05-INTERPRETATION

- Case: `M-NET-05-INTERPRETATION`
- Result: passed
- Tests: `TestMNet05Interpretation` in `net05_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet05Interpretation$' ./.`
- Scope: Assert the serialized NET-05 result for completed, incomplete, network_not_validated and inconclusive synthetic runs contains no interpretation, verdict, classification, anomaly, censorship, allowlist or broad-restriction field, and that outcome takes only completed, network_not_validated, incomplete or inconclusive, never broad_restriction_suspected, allowlist_detected or another censorship classification, including when every step reports a primitive negative outcome. Assert Go field documentation, the result-type doc comments and sanitized example and host-facing text for NET-05 contain none of these prohibited claims: that the result proves an allowlist policy exists or identifies who operates it; that combined results are classified as a broad restriction, allowlist or censorship event; that a profile tests sensitive, political, circumvention or broad public website lists; that NET-05 replaces outage, captive-portal, routing or independent reference analysis; that correlated observations establish censor intent or policy with certainty; and that external datasets are ground truth. Verify documentation states that one complete execution establishes only the ordered primitive results on one network generation, that correlation with target knowledge, repeated runs, controls and independent signals is outside NET-05 version 0.1.0, and that the listed alternative explanations (simultaneous outage or misconfiguration of external targets, shared hidden CDN/DNS/routing/certificate dependencies, transit failure or congestion, captive portal/subscription/VPN/enterprise policy/local firewall, stale profile definitions or changed endpoint behavior, address-family failure or mobile handover) are present. Requirements: REQ-NET-05-0001, REQ-NET-05-0002, REQ-NET-05-0003, REQ-NET-05-0004, REQ-NET-05-0049, REQ-NET-05-0050, REQ-NET-05-0051, REQ-NET-05-0052. Judgement-based completion is the separate review case RV-M-NET-05-INTERPRETATION; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-NET-05-0001: No non-negated sentence in the comments and string literals of production Go files and sanitized examples and the shipped Markdown documentation (case plans, manifests, evidence, changelog and specification-issue records excluded) matches the "NET-05 proves an allowlist or identifies its operator" claim pattern, and the NET-05 section of the interpretation guidance contains "does not prove that an allowlist policy exists or identify who operates it". This is a textual scan of shipped text; it does not prove how a reader interprets a result.
  - REQ-NET-05-0002: The same scan finds no non-negated sentence stating that combined or correlated results are classified as a broad restriction, allowlist or censorship event, and the guidance contains "does not classify the combined results as a broad restriction, allowlist or censorship event"; the profile run, which includes a failing DNS step, has outcome `completed`, and no member name or string value of any run's measurement contains a verdict token (such as `verdict`, `censor`, `block`, `filter`, `anomal`, `confiden`) and the strict result decoder rejects `verdict`, `interpretation`, `censored`, `blocked`, `filtered`, `anomaly`, `confidence`, `reachable` and `authoritative` fields injected into a valid result.
  - REQ-NET-05-0003: The same scan finds no non-negated sentence stating that a profile tests sensitive, political, circumvention or broad public website lists, and the guidance contains "is not for sensitive, political, circumvention or broad public website lists"; profile contents are not inspected.
  - REQ-NET-05-0004: The same scan finds no non-negated sentence stating that NET-05 replaces outage, captive-portal, routing or reference analysis, and the guidance contains "does not replace outage, captive-portal, routing or independent reference analysis".
  - REQ-NET-05-0049: The guidance contains "establishes only the ordered primitive results the inline profile produced on one network generation", the runs with platform validation yes, no and unknown record `completed`, `network_not_validated` and `inconclusive`, and every outcome present, including in a host-cancelled run, is registered; step order is not checked by this test.
  - REQ-NET-05-0050: Only the guidance phrase "is outside net-05 version `0.1.0`" is required; nothing else about separate correlation analysis is checked.
  - REQ-NET-05-0051: Not asserted: the test requires none of the listed evidence items in the guidance and checks none of them.
  - REQ-NET-05-0052: The same scan finds no non-negated sentence stating that correlation establishes or proves intent or policy with certainty, or that external datasets are ground truth, and the guidance contains "cannot establish censor intent or policy with certainty" and "external datasets are supporting reference signals, not ground truth".

<a id="M-NET-05-SAFETY"></a>

### M-NET-05-SAFETY

- Case: `M-NET-05-SAFETY`
- Result: passed
- Tests: `TestMNet05Safety` in `measurement_safety_conformance_test.go` (package `root`); `TestMNet05PrimitiveRiskAndMissingCapability` in `net05_conformance_test.go` (package `root`); `TestMNet05Minimization` in `net05_conformance_test.go` (package `root`); `TestMNet05Admission` in `net05_conformance_test.go` (package `root`); `TestMNet05AtomicTargetAndProxyRules` in `net05_conformance_test.go` (package `root`); `TestMNet05ExecutionAndOutput` in `net05_runner_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet05Safety$' ./.`
- Scope: Iterate the Baseline Risk rows using synthetic targets and recording transports: risk level medium and requires user consent false subject to host and local policy; assert a task declaring low is rejected or evaluated as at least medium, a task declaring high raises effective risk to high, and a host/local policy requiring consent blocks execution without consent despite the false baseline. Iterate each Risk Factors row (network_observability medium, traffic_distinctiveness medium, target_sensitivity low, data_sensitivity medium, legal_or_social_context medium), asserting effective risk is the most restrictive of NET-05, every referenced primitive (DNS-01, TCP-01, TLS-01, HTTP-01, each registered low), the task declaration and local policy, and that an unknown factor type is rejected or treated as unknown rather than ignored. Exercise each normative safety action separately: data minimization (retain primitive classifications and safe step identifiers only, no duplicated raw target lists, response bodies, full addresses or detailed network identity); traffic distinctiveness (recording transport shows strictly sequential steps, no automatic retries, no port variation, no target expansion, and broad lists rejected; no synchronized fleet-wide cadence is introduced by the library); user-identification risk (no per-installation identifier and no unrelated device or application telemetry in output or traffic); target risk (each inline target independently passes schema and local safety policy, and one rejected target rejects the whole profile before contact); repeated-execution risk (an overlapping second profile run on the same network generation is delayed or rejected and scheduling remains host/local controlled); and final probe authority (local policy can delay, skip or reject, producing local_policy_rejected with no measurement traffic). Requirements: REQ-NET-05-0053, REQ-NET-05-0054, REQ-NET-05-0055, REQ-NET-05-0056, REQ-NET-05-0057, REQ-NET-05-0058, REQ-NET-05-0059, REQ-NET-05-0060. Judgement-based completion is the separate review case RV-M-NET-05-SAFETY; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-NET-05-0053: The test asserts that the production registry records a `medium` baseline for NET-05, that each referenced primitive (DNS-01, TCP-01, TLS-01, HTTP-01) is a registered `low` step, and that a task declaring `low` or `medium` has effective risk exactly `medium`, `high` exactly `high` and `unknown` exactly `unknown` (admitted under issuer and local bounds permitting that level, rejected `task_issuer_not_authorized` with no contact when either side permits only lower levels). For the consent baseline it asserts that `false` adds no requirement (with host consent absent the run starts, yields one result and never queries the consent capability), while consent required by the task, by an issuer or by the local policy blocks every result and contact unless the fake host reports valid consent (absent, ended and unavailable all block).
  - REQ-NET-05-0054: For each of the five factor types the test asserts that the registered baseline equals the most restrictive of the fixture's row levels (`target_sensitivity` `low`, the other four `medium`), that declaring the factor at `low`, `medium` or `high` gives effective risk exactly the maximum of baseline, row level and declared level, that an `unknown` factor gives `unknown`, and that an unknown factor under a known overall level or a factor above the overall level is rejected `task_schema_invalid` and an unrecognized factor type `task_unsupported`, all with no contact. The row levels are fixture constants checked only against the registered baseline, and the rationale text is not asserted; a further row asserts that raising any one primitive's target registration to `high` makes effective risk exactly `high`, that a registry giving HTTP-01 a `high` baseline is rejected under a `medium` bound and one lacking HTTP-01 is rejected `task_unsupported`, and that disabling composite execution on either side rejects the task with no contact.
  - REQ-NET-05-0055: The test asserts that no surface contains the outer schedule and labels, step hostnames, URL path, address, port, server name, option value, response body and header value, certificate fields, platform and raw TCP errors, DNS TTL, records and raw bytes, or the combined verdict strings, and that the measurement holds only profile identity, validation state, step identity, test version and state, `stopped_before_step` and each primitive's closed field set. It also asserts that the DNS answer appears only in its answers path, that schedule, safety, labels, input and options are not copied and that there is one `network_context`; the decoder rejects such fields injected.
  - REQ-NET-05-0056: The test asserts that the four declared steps run in order with three dials, one lookup and one HTTP request and no overlapping dials, that profiles of one, four and eight steps are admitted without contact while a nine-step profile (which also repeats step identifiers), a duplicated step and an oversized profile are rejected with no contact or state. Retries and port variation are covered only by those exact counts, and fleet-wide cadence is not asserted.
  - REQ-NET-05-0057: The test asserts that the HTTP step's request carries none of the listed identifying headers and no header value containing `safety-` or `barometer`. Combination with unrelated device or application telemetry is covered only by the closed result field set.
  - REQ-NET-05-0058: The test asserts that changing any single step's registration (URL, test, proxy-only path, expiry, or the issuer no longer listing it) rejects the whole profile `target_not_registered`, that an issuer or deployment lacking HTTP-01 or composite permission rejects it, that a proxy-incompatible DNS or hostname step rejects it, and that a loopback step or a URL with userinfo is rejected, each with no contact or published state. Documented ownership and operational expectations of endpoints are not asserted.
  - REQ-NET-05-0059: The test asserts that a second profile run admitted while the first is held at its first dial makes no additional dial or lookup and can be cancelled, and the shared harness shows the local device constraint holding a run. Host scheduling policy and the residual-risk statement are not asserted.
  - REQ-NET-05-0060: The test asserts that a local policy not permitting the task's network path rejects a valid task at admission with no contact, that a local risk bound below `medium` rejects it, and that a local charging-only device constraint holds an admitted valid task (`run_not_due` twice, no result, no contact) until satisfied, after which the same task runs; a target registration that expires after admission ends the attempt `local_policy_rejected` with no contact. The no-lowering clause is asserted through the declared-level rows: no task declaration yields effective risk below the `medium` baseline.

<a id="M-NET-05-OUTPUT"></a>

### M-NET-05-OUTPUT

- Case: `M-NET-05-OUTPUT`
- Result: passed
- Tests: `TestMNet05ExecutionAndOutput` in `net05_runner_test.go` (package `root`); `TestMNet05Termination` in `internal/execution/net05_conformance_test.go` (package `internal/execution`); `TestCcNet05ProxyLifecycleAndRenewal` in `net05_runner_test.go` (package `root`); `TestCcNet05PathContextAndProxyGeneration` in `internal/execution/net05_conformance_test.go` (package `internal/execution`); `TestMNet05ReceiverRejectsVersions` in `internal/execution/net05_closeout_test.go` (package `internal/execution`); `TestMNet05PlatformValidationGate` in `net05_runner_test.go` (package `root`); `TestMNet05StepNegativeOutcomesStayCompleted` in `internal/execution/net05_closeout_test.go` (package `internal/execution`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet05ExecutionAndOutput$' ./.`
- Scope: Validate exact profile identity, required network_validation state/source, ordered bounded started steps, exact primitive/result versions and conditional primitive measurement. Test interruption before each step and inside the last started step: at most one final incomplete step, no placeholders, exact stopped_before_step and one outer Network Context.
- Requirement aspects exercised:

  - REQ-NET-05-0040: Each result has exactly one network_context (direct and proxy paths, and a proxy-reconfigured partial result keeps the proxy path snapshot), ordered primitive step results and no target literal in the measurement.
  - REQ-NET-05-0041: Results carry profile, network_validation and steps no longer than the profile; stopped_before_step names the next unstarted step after an interruption before or inside a step and is absent after the last step or on completion.
  - REQ-NET-05-0042: The emitted profile carries the exact validated profile_id baseline and profile_version 1.
  - REQ-NET-05-0043: network_validation.source is platform and state is validated, not_validated or unknown from the platform context; the decoder rejects other or missing states, not_validated on a completed run and another source (TestMNet05ReceiverRejectsVersions).
  - REQ-NET-05-0066: Every emitted network_validation has source platform and no fresh or cached value appears in the serialized result; the decoder rejects another source (TestMNet05ReceiverRejectsVersions).
  - REQ-NET-05-0044: Each step carries its exact step_id, primitive test ID, version and result version 0.1.0; completed steps carry a primitive measurement and an interrupted step carries none; the decoder rejects a step with a common envelope or a Network Context.
  - REQ-NET-05-0045: Interruption before each step emits only earlier completed steps and inside a step emits that step as the single final incomplete entry; no placeholder is emitted for unstarted steps.

<a id="M-NET-05-VERSIONS"></a>

### M-NET-05-VERSIONS

- Case: `M-NET-05-VERSIONS`
- Result: passed
- Tests: `TestMNet05Versions` in `net05_conformance_test.go` (package `root`); `TestMNet05ReceiverRejectsVersions` in `internal/execution/net05_closeout_test.go` (package `internal/execution`); `TestMNet05ExecutionAndOutput` in `net05_runner_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet05Versions$' ./.`
- Scope: Select the exact registered combination for NET-05: test.id NET-05, test.version 0.1.0, Measurement Result schema_version 0.1.0 and test.result_version 0.1.0, and for every dispatched primitive step test_id DNS-01, TCP-01, TLS-01 or HTTP-01 with test_version 0.1.0 and emitted step test.result_version 0.1.0. Assert a valid profile produces exactly these strings in the outer envelope and in each steps[].test object. Iterate rejection inputs using a recording transport: outer NET-05 version 0.1.1, 0.2.0, 1.0.0, empty string, v0.1.0 and non-string values; each step test_version set to an unsupported, empty, padded or non-string value while other steps stay valid; a step test_id of NET-05 (nested), NET-01, NET-03, DNS-03 and an unregistered ID; a DNS-01 step whose input/options belong to HTTP-01 (cross-test input); and an emitted result whose schema_version, outer result_version or a step result_version differs from 0.1.0. Assert every unsupported task or step version, malformed version and cross-test combination is a task-validation failure reported outside Measurement Results, that the whole profile is rejected before any primitive step or connection (zero recorded contacts; no compatible subset runs), and that no version is guessed, coerced, defaulted or added. Assert a receiver validator refuses best-effort interpretation of a result carrying an unknown schema, test or result version. Registration metadata and Draft status confer no support claim. Requirements: REQ-NET-05-0039.
- Requirement aspects exercised:

  - REQ-NET-05-0039: Outer and per-step test_version 0.1.1, 0.2.0, 1.0.0, empty, v0.1.0, padded, numeric and null, step test IDs NET-05, NET-01, NET-03, DNS-03, SYN-01 and tcp-01 and a DNS-01 step with HTTP-01 input reject the whole profile with no handle or replay commit, and the step-level cases also with no contact; accepted steps keep version 0.1.0 and order. Emitted results carry schema_version 0.1.0, NET-05 0.1.0 and result 0.1.0 and each step its primitive 0.1.0 tuple; the decoder rejects 0.1.1, 0.2.0, 1.0.0, 9.0.0, 0.1, v0.1.0, padded, empty, numeric and null outer, test, result and step versions, missing version fields, other outer and step test IDs and an extra step test member.

<a id="M-NET-01-ADMISSION"></a>

### M-NET-01-ADMISSION

- Case: `M-NET-01-ADMISSION`
- Result: passed
- Test: `TestMNet01Admission` in `net01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet01Admission$' ./.`
- Scope: Accept only empty input; vary wrong root types, null, missing input and every extra property, including interface names and identifiers. Assert rejection before platform collection, traffic, or Measurement Result construction.
- Requirement aspects exercised:

  - REQ-NET-01-0005: The empty-input task is admitted through Client.AdmitTask with a local policy carrying no target registrations, and the production registry entry for NET-01 0.1.0 uses registry.NoTargets and registry.NoProxy, is low risk and is neither a NET-05 step nor a composite.
  - REQ-NET-01-0010: The task with input {} is admitted with no platform read, transport use or result written at admission.
  - REQ-NET-01-0011: Objects with any added property (interface, interface_name, interface_index, ip_address, local_address, device_id, ssid, bssid, imsi, hostname, an empty name and x) are rejected as task_schema_invalid.
  - REQ-NET-01-0012: null, arrays, strings (including "{}"), numbers, booleans and a missing input member are rejected as task_schema_invalid; only {} is admitted.
  - REQ-NET-01-0013: Input members carrying interface names or indexes, local IPv4/IPv6 addresses, a device ID, SSID, BSSID and IMSI are rejected as task_schema_invalid.
  - REQ-NET-01-0014: Every rejected input leaves no handle, replay commit, run, platform read, dialer use or Measurement Result.

<a id="M-NET-01-OPTIONS"></a>

### M-NET-01-OPTIONS

- Case: `M-NET-01-OPTIONS`
- Result: passed
- Test: `TestMNet01Options` in `net01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet01Options$' ./.`
- Scope: Accept the empty options object and reject every added property or wrong type without changing the selected snapshot behavior.
- Requirement aspects exercised:

  - REQ-NET-01-0015: The empty options object is admitted and runs to a completed snapshot result with the six-field access object.
  - REQ-NET-01-0016: null, arrays, strings, numbers, booleans, a missing options member and objects with x, resolver_mode, egress or interface members are rejected as task_schema_invalid with no replay commit, platform read or dialer use.

<a id="M-NET-01-PRECONDITIONS"></a>

### M-NET-01-PRECONDITIONS

- Case: `M-NET-01-PRECONDITIONS`
- Result: passed
- Tests: `TestMNet01Preconditions` in `net01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet01Preconditions$' ./.`
- Scope: Run the portable NET-01 gate against a fake platform with a synthetic capability matrix and assert no network operation is attempted. Supported network-state API absent: assert execution.state not_run with a registered reason per the public table and no measurement. Required state API present but the only available implementation path requires interface addresses, SSIDs, BSSIDs, subscriber identifiers or location (iterate each): assert the gate denies before any such access occurs and the fake platform records zero calls to those accessors. Assert the gate requests from the fake platform only the network-state capability. Stricter task execution constraint or local probe policy denial: assert not_run with constraints_unsatisfied or local_policy_rejected respectively. Requirements: REQ-NET-01-0017, REQ-NET-01-0018, REQ-NET-01-0019, REQ-NET-01-0020. Judgement-based completion is the separate review case RV-M-NET-01-PRECONDITIONS; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-NET-01-0017: With the fake network-state port reporting the API missing (`platform.ErrNetworkUnavailable`) at reservation, `StartRun` returns `run_not_due`, no result is written and no dialer is called. Through the `TestMNet01Normalization` subtest, a missing or wrapped-missing API at the test's own read gives `completed` with outcome `platform_unavailable` and no raw error, dial or armed watch; no `not_run` result is produced for this row.
  - REQ-NET-01-0018: A completed NET-01 run reads the network-state port and makes zero calls to the direct dialers, the consent port and the device port, so the library uses only the network-state capability. What the host application grants on a platform is not asserted; native/host row, see docs/measurement-requirements.md.
  - REQ-NET-01-0019: Not asserted by this case; the portable platform port has no accessor for interface addresses, SSIDs, BSSIDs, subscriber identifiers or location, so the plan's per-accessor zero-call iteration is not run, and the native APIs are a native/host row (`RV-M-NET-01-PRECONDITIONS`), see docs/measurement-requirements.md.
  - REQ-NET-01-0020: The shared portable gates assert, each with zero contacts: a task transport condition the network does not meet, a task excluding VPN on a VPN-detected network and a local charging-only device restriction defer the run at `StartRun` (`run_not_due`, no result); a transport condition lost after reservation gives `not_run` with `constraints_unsatisfied`; and a proxy-path task with no proxy configured is either rejected at admission with `task_schema_invalid` or ends `not_run` with `proxy_not_configured` (the test accepts either). NET-01 has no target registration, so the withheld and lapsed registration rows do not run for this family.

<a id="M-NET-01-EXECUTION"></a>

### M-NET-01-EXECUTION

- Case: `M-NET-01-EXECUTION`
- Result: passed
- Test: `TestMNet01ExecutionAndOutput` in `net01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet01ExecutionAndOutput$' ./.`
- Scope: Use a recording platform provider to assert normalized access reads, unknown values, transition detection and release of handles/callbacks. Install a transport that fails on any use; no primary or fallback egress endpoint may be contacted.
- Requirement aspects exercised:

  - REQ-NET-01-0006: A run through Client with the production registry produces a snapshot whose access holds the platform's network type and IPv4 value and unknown for the VPN, IPv6 and validation properties the fake platform does not report.
  - REQ-NET-01-0007: The run makes exactly one platform read after the reservation check, requests no path dialer and makes no connection (zero calls on the recording transport), releases the generation watch, and a stated egress requirement ends not_run with network_context_unavailable without starting the test; a stated access condition unsatisfied at the attempt's gate read ends not_run with constraints_unsatisfied.

<a id="M-NET-01-NORMALIZATION"></a>

### M-NET-01-NORMALIZATION

- Case: `M-NET-01-NORMALIZATION`
- Result: passed
- Tests: `TestMNet01Normalization` in `net01_conformance_test.go` (package `root`); `TestMNet01Termination` in `net01_conformance_test.go` (package `root`); `TestMNet01Limits` in `net01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet01Normalization$' ./.`
- Scope: Supply coherent, unavailable, failed and inconsistent snapshots; exercise every failure/outcome row and distinguish unknown properties from unavailable/inconsistent snapshots. Verify completed snapshots cannot combine generations.
- Requirement aspects exercised:

  - REQ-NET-01-0008: Coherent platform reads with every network type, VPN and availability value, with zero values and with unrecognized values each complete as snapshot with the mapped names and explicit unknown where the platform reports nothing recognizable.
  - REQ-NET-01-0009: The platform's no-API signal (direct or wrapped) yields completed platform_unavailable, any other platform error or a zero generation yields completed platform_error, neither carries access or the raw error text, a platform unavailable at StartRun leaves no run, and a generation change during the read yields incomplete network_changed with no measurement (TestMNet01Termination); the inconclusive row is covered by disposition, not by a test.
  - REQ-NET-01-0026: snapshot carries the complete six-field access object with network_changed false, platform_unavailable and platform_error carry only outcome and an empty observations array, inconclusive is never emitted, and concurrent runs across a generation transition each report one generation's values or incomplete network_changed, never a merged snapshot (TestMNet01Limits).

<a id="M-NET-01-LIMITS"></a>

### M-NET-01-LIMITS

- Case: `M-NET-01-LIMITS`
- Result: passed
- Tests: `TestMNet01Limits` in `net01_conformance_test.go` (package `root`); `TestNet01DeadlineReleasesRead` in `internal/execution/net01_termination_test.go` (package `internal/execution`); `TestNet01GenerationChange` in `internal/execution/net01_termination_test.go` (package `internal/execution`); `TestMNet01ExecutionAndOutput` in `net01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet01Limits$' ./.`
- Scope: Drive the fake clock to the outer deadline and assert prompt release with zero network operations/bytes. Run concurrent snapshots with a generation transition and assert coherent values or the prescribed incomplete state, never merged state.
- Requirement aspects exercised:

  - REQ-NET-01-0021: A platform read left unanswered is ended by the 30-second outer deadline on the fake clock as incomplete attempt_deadline_exceeded with the read, timer and watch released; a failed platform read completes as platform_error after exactly one attempt read with no retry; no run requests a path dialer or connection; and concurrent runs across a generation transition keep one coherent generation each or end incomplete network_changed.

<a id="M-NET-01-TERMINATION"></a>

### M-NET-01-TERMINATION

- Case: `M-NET-01-TERMINATION`
- Result: passed
- Tests: `TestMNet01Termination` in `net01_conformance_test.go` (package `root`); `TestNet01AllTerminationBoundaries` in `internal/execution/net01_termination_test.go` (package `internal/execution`); `TestNet01GenerationChange` in `internal/execution/net01_termination_test.go` (package `internal/execution`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet01Termination$' ./.`
- Scope: Use a fake platform network-state provider, a deterministic clock and a recording transport that fails on any use, and inject interruptions at each NET-01 Procedure point: before step 1 (read current path/connectivity state through the shared Network Context collector), during step 1 (provider call outstanding), between steps 1 and 2 (raw state read, not yet normalized into `access`), during step 3 (network-generation check) and during step 4 (release of callbacks/handles). Expected states follow the GO-SPEC-ISSUE-004 disposition: before step 1 the attempt is not_run with that reason, no measurement and no network_context; collection starts the test, so steps 1 to 3 are incomplete as asserted below; during step 4, after the outcome is recorded, the attempt stays completed with the recorded outcome and the release still completes. At each point inject separately: a network-generation change, host cancellation, outer attempt-deadline expiry on the fake clock, and platform suspension. At the incomplete points, assert a generation change during collection yields `execution.state: incomplete` with `execution.reason: network_changed` and no `measurement` object (no `outcome`, `observations` or `access`); assert cancellation, deadline expiry and suspension yield the common execution layer reasons `cancelled_by_host`, `attempt_deadline_exceeded` and `platform_suspended` with no `measurement`, never a NET-01 `outcome` value such as `platform_error` or `inconclusive`. Inject two conditions in sequence and assert `reason` names only the first. Assert no top-level `network_context` or `egress` is added on any interrupted path, every provider handle and callback created for the run is released, the recording transport records zero connections and zero bytes, and no retry or fallback (including any egress-classification endpoint) is attempted within the run. Proxy loss/reconfiguration is not exercised because NET-01 opens no network path. Requirements: REQ-NET-01-0027.
- Requirement aspects exercised:

  - REQ-NET-01-0027: Host cancellation, a generation change, outer-deadline expiry and injected platform suspension, each followed by a second condition, give not_run before the read, incomplete with only the first reason and no measurement during and after the read, and completed snapshot when injected after the outcome is recorded, with no top-level network_context, zero usage and dialer requests, and no timer or watch left; through Client, a generation change, cancellation and deadline while the read is outstanding give incomplete with that reason and no measurement.

<a id="M-NET-01-MINIMIZATION"></a>

### M-NET-01-MINIMIZATION

- Case: `M-NET-01-MINIMIZATION`
- Result: passed
- Tests: `TestMNet01Minimization` in `net01_conformance_test.go` (package `root`), `TestMNet01Safety` in `measurement_safety_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet01Minimization$' ./.`
- Scope: Configure the fake platform provider to expose one unique synthetic canary string or value in each item REQ-NET-01-0028 prohibits: interface name, interface index, local address, public address, route, DNS configuration, SSID, BSSID, Wi-Fi scan data, cell identifier, SIM data, subscriber data, carrier name, signal strength, VPN application identity, network generation, platform object description and raw platform error text (the last via a scripted `platform_error` provider failure). Run NET-01 to `snapshot`, `platform_error`, `inconclusive` and `incomplete`/`network_changed` and assert the serialized Measurement Result, every nested object, the empty `observations` array, any task-validation or error path and any log or diagnostic surface exposed by the result path contain none of the canaries. Assert the only retained measurement fields are `outcome`, `observations` (empty) and, where permitted, `access` with exactly `network_type`, `vpn_state`, `ipv4_available`, `ipv6_available`, `validated` and `network_changed` drawn from their enumerated values; assert no top-level `network_context`, no `egress` object and no extra field. Only synthetic provider data is used. Requirements: REQ-NET-01-0028.
- Requirement aspects exercised:

  - REQ-NET-01-0028: Each listed item gets one unique synthetic canary, as do the fake platform's withheld values: the network generation is reported by the fake as a canary generation, and every other item reaches the core only inside the error text, wrapped platform object and description of a scripted platform failure. Over the `snapshot`, `platform_error`, `platform_unavailable` and `incomplete`/`network_changed` runs, no canary appears raw, case-insensitively, JSON-escaped, percent-encoded, in hexadecimal or in base64 (a digit-only canary as a whole number) in any result, the bundle, the pending store's bytes, any returned error or the task-validation rejection of a task whose unknown input member carries the canaries; each result has no top-level `network_context` and no `egress`, an outcome's measurement holds exactly `outcome`, an empty `observations` array and, for `snapshot`, `access` whose five enumerated fields take enumerated values beside `network_changed`, the interrupted run keeps no measurement, and the strict decoder rejects each canary injected into `access` or the measurement, an `inconclusive` member, `egress`, `network_context` and a non-empty `observations` entry. The `inconclusive` path is not driven, because the runner never emits it.

<a id="M-NET-01-INTERPRETATION"></a>

### M-NET-01-INTERPRETATION

- Case: `M-NET-01-INTERPRETATION`
- Result: passed
- Tests: `TestMNet01Interpretation` in `net01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet01Interpretation$' ./.`
- Scope: Assert the serialized NET-01 `measurement` contains only `outcome`, `observations` and `access`, that `outcome` takes only `snapshot`, `platform_unavailable`, `platform_error` or `inconclusive`, and that no field names or values express a verdict or interpretation (for example censorship, blocking, anomaly, reachability or operator conclusions). Assert Go doc strings for the NET-01 result and access types, documented field descriptions and every sanitized example or host-facing text string shipped with the implementation contain none of these prohibited claims: that the Internet is reachable; that `validated: yes`, `ipv4_available: yes` or `ipv6_available: yes` establishes Internet reachability; identification of a network operator, subscriber, device or physical location; a country, region, ASN, provider, mobile operator or public IP address derived from NET-01; that `vpn_state: detected` establishes traffic is routed through a VPN; and that a result is definitive proof of censorship. Verify the field descriptions retain the documented qualifiers that availability values are platform-reported and not proven reachability. Requirements: REQ-NET-01-0001, REQ-NET-01-0002, REQ-NET-01-0003, REQ-NET-01-0004, REQ-NET-01-0029, REQ-NET-01-0030, REQ-NET-01-0031. Judgement-based completion is the separate review case RV-M-NET-01-INTERPRETATION; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-NET-01-0001: No non-negated sentence in the comments and string literals of production Go files and sanitized examples and the shipped Markdown documentation (case plans, manifests, evidence, changelog and specification-issue records excluded) matches the "the Internet is reachable" or "availability establishes reachability" claim pattern, and the NET-01 section of the interpretation guidance contains "does not establish that the internet is reachable" and "are platform-reported and are not proven reachability"; the Go availability description in `internal/platform/network.go` keeps its platform-reported qualifier, no run's measurement contains `reachab`, and the strict decoder rejects an injected `reachable` field. This is a textual scan of shipped text; it does not prove how a reader interprets a result.
  - REQ-NET-01-0002: The same scan finds no non-negated sentence stating that NET-01 identifies or reveals an operator, subscriber, device or physical location, and the guidance contains "does not identify a network operator, subscriber, device or physical location".
  - REQ-NET-01-0003: The same scan finds no non-negated sentence stating that NET-01 derives, determines, infers or reports a country, region, ASN, provider, mobile operator or public IP address, and the guidance contains "country, region, asn, provider, mobile operator and public ip address are not derived from net-01"; the four results (two snapshots, `platform_error`, `platform_unavailable`) have no measurement member beyond `outcome`, `observations` and `access`.
  - REQ-NET-01-0004: The same scan finds no non-negated sentence stating that a detected VPN establishes, proves, means or shows routing, and the guidance contains "does not establish that traffic is routed through a vpn".
  - REQ-NET-01-0029: Each of the four results carries a registered outcome and only `outcome`, `observations` and `access` in its measurement, and the guidance keeps the platform-reported and country, region, ASN, provider and operator non-derivation qualifiers; the test does not otherwise check what a single snapshot establishes.
  - REQ-NET-01-0030: Only the third item is asserted: the guidance contains "repeated snapshots after a reported network change"; the NET-03 results and same-generation protocol measurements items are not checked.
  - REQ-NET-01-0031: No non-negated sentence of the scanned text matches the common claims (a result is or provides definitive or conclusive proof of censorship or blocking; a result proves, demonstrates or confirms censorship or blocking; a censorship, blocking, anomaly or success verdict), the guidance contains "a single result is not definitive proof of censorship", and no member name or string value of any run's measurement contains a verdict token (such as `verdict`, `censor`, `block`, `filter`, `anomal`, `confiden`) and the strict result decoder rejects `verdict`, `interpretation`, `censored`, `blocked`, `filtered`, `anomaly`, `confidence`, `reachable` and `authoritative` fields injected into a valid result.

<a id="M-NET-01-SAFETY"></a>

### M-NET-01-SAFETY

- Case: `M-NET-01-SAFETY`
- Result: passed
- Tests: `TestMNet01Safety` in `measurement_safety_conformance_test.go` (package `root`); `TestMNet01Minimization` in `net01_conformance_test.go` (package `root`); `TestMNet01ExecutionAndOutput` in `net01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet01Safety$' ./.`
- Scope: Run NET-01 with a fake platform provider and a recording transport for each registered row and action. Baseline Risk: verify registered `risk_level: low` and `requires_user_consent: false`; exercise a task declaring `low` (accepted), a task declaring `medium` or `high` (effective risk at least the declared level), a task declaring `requires_user_consent: true` with no host-established consent (no execution), and a trusted local policy that rejects or delays (`not_run` with `local_policy_rejected`, no measurement). Risk Factors: for `network_observability` and `traffic_distinctiveness` assert zero connections, zero bytes and no DNS or socket use on the recording transport; for `target_sensitivity` assert no remote target is accepted (non-empty input rejected before execution); for `data_sensitivity` assert output limited to the six coarse `access` classifications with no interface or subscriber identifiers; for `legal_or_social_context` assert no network-visible participation signal (zero traffic). Safety Analysis actions: data minimization (no `egress`, no carrier identity), traffic distinctiveness (no traffic), user-identification risk (synthetic interface names, addresses, SSIDs, BSSIDs, carrier names and subscriber identifiers from the provider are absent), target risk (no target), repeated-execution risk (N repeated runs record zero network operations), residual risk (no `network_context` or `egress` joined to output), and probe final authority to delay, skip or reject. Requirements: REQ-NET-01-0032, REQ-NET-01-0033, REQ-NET-01-0034, REQ-NET-01-0035, REQ-NET-01-0036, REQ-NET-01-0037, REQ-NET-01-0038. Judgement-based completion is the separate review case RV-M-NET-01-SAFETY; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-NET-01-0032: The test asserts that the production registry records a `low` baseline for NET-01 and that a task declaring `low`, `medium`, `high` or `unknown` has effective risk exactly that level (admitted under issuer and local bounds permitting it, rejected `task_issuer_not_authorized` with no contact when either side permits only lower levels). For the consent baseline it asserts that `false` adds no requirement (with host consent absent the run starts, yields one result and never queries the consent capability), while consent required by the task, by an issuer or by the local policy blocks every result and contact unless the fake host reports valid consent (absent, ended and unavailable all block). Every run in these rows records zero calls on the recording direct dialer.
  - REQ-NET-01-0033: For each of the five factor types the test asserts that the registered baseline equals the most restrictive of the fixture's row levels (all five `low`), that declaring the factor at `low`, `medium` or `high` gives effective risk exactly the maximum of baseline, row level and declared level, that an `unknown` factor gives `unknown`, and that an unknown factor under a known overall level or a factor above the overall level is rejected `task_schema_invalid` and an unrecognized factor type `task_unsupported`, all with no contact. The row levels are fixture constants checked only against the registered baseline, and the rationale text is not asserted; for NET-01 the fixture expects no traffic, so every running row also asserts zero dialer calls, inputs naming a hostname, address or URL target are rejected `task_schema_invalid` with no dialer call, and the minimization row bounds `data_sensitivity`. Socket use is measured through the direct dialer only; no separate DNS counter is installed.
  - REQ-NET-01-0034: The test asserts that the measurement holds only `outcome`, `observations` and the six `access` fields (`network_type`, `vpn_state`, `ipv4_available`, `ipv6_available`, `validated`, `network_changed`) with each value inside its enumeration, that results carry no top-level `network_context`, and that the decoder rejects injected `egress`, `network_context` and canary members. Carrier identity is covered by the absent carrier-name canary.
  - REQ-NET-01-0035: The test asserts zero calls on the recording direct dialer in the consent, delay, execution and five-run repetition rows, and the execution row also asserts no platform network watch is left open.
  - REQ-NET-01-0036: The test asserts that interface name and index, local and public addresses, route, DNS configuration, SSID, BSSID, Wi-Fi scan data, cell identifier, SIM and subscriber data, carrier name, signal strength, VPN application identity, network generation and raw platform error text, supplied through the fake platform's error and object text and its withheld values, appear in no result, pending store, bundle or error text. The decoder rejects each canary injected as an `access` or measurement member.
  - REQ-NET-01-0037: The test asserts that inputs naming a remote target are rejected at admission with no dialer call, that five consecutive snapshots produce five results with zero dialer calls, and, for residual risk, that results never carry top-level `network_context` or `egress` and a task requiring egress context ends `not_run` with `network_context_unavailable` without an egress lookup or dialer call.
  - REQ-NET-01-0038: The test asserts that a local policy not permitting the task's network path rejects a valid task at admission with no contact, and that a local charging-only device constraint holds an admitted valid task (`run_not_due` twice, no result, no contact) until satisfied, after which the same task runs. The no-lowering clause is asserted through the declared-level rows: no task declaration yields effective risk below the `low` baseline. The registration-withdrawal row is skipped because NET-01 has no target registration.

<a id="M-NET-01-OUTPUT"></a>

### M-NET-01-OUTPUT

- Case: `M-NET-01-OUTPUT`
- Result: passed
- Tests: `TestMNet01ExecutionAndOutput` in `net01_conformance_test.go` (package `root`); `TestMNet01Versions` in `net01_conformance_test.go` (package `root`); `TestMNet01Normalization` in `net01_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet01ExecutionAndOutput$' ./.`
- Scope: Verify no top-level network_context, no egress, empty observations and the conditional complete access object for snapshot/inconclusive. Exercise all access enums through the canonical shared format suite.
- Requirement aspects exercised:

  - REQ-NET-01-0023: Every emitted NET-01 result has no top-level network_context, and a completed measurement holds exactly outcome, an empty observations array and, for snapshot, the six-field access object; the result decoder rejects a top-level network_context, a missing or incomplete access object on snapshot, an added access field, an unregistered value for each access field and network_changed true.
  - REQ-NET-01-0024: No emitted result contains an egress member, and the result decoder rejects an egress member in the measurement.
  - REQ-NET-01-0025: The result decoder accepts inconclusive with or without access, accepts platform_unavailable and platform_error only without access, and rejects snapshot without access.

<a id="M-NET-01-VERSIONS"></a>

### M-NET-01-VERSIONS

- Case: `M-NET-01-VERSIONS`
- Result: passed
- Test: `TestMNet01Versions` in `net01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet01Versions$' ./.`
- Scope: Select the exact registered combination `test.id: NET-01`, `test.version: 0.1.0`, `test.result_version: 0.1.0` with Measurement Result `schema_version: 0.1.0` and verify it is accepted and emitted as exactly that combination. Iterate rejections on each field independently, with the other fields at their registered values: an unregistered but well-formed version (`0.2.0`, `1.0.0`), malformed strings (`0.1`, `v0.1.0`, `0.1.0-draft`, empty string, non-string value) and a missing field, for `schema_version`, `test.version` and `test.result_version`. Iterate cross-test combinations: a result with `test.id: NET-03` carrying the NET-01 `access` measurement shape, and a NET-01 result whose `measurement` follows another registered test's Output shape. Assert unsupported task versions are rejected as task-validation failures before execution through the operational interface and produce no Measurement Result, and that unsupported result combinations are rejected rather than coerced, interpreted best-effort or mapped to another version. Assert the `Draft` status and `0.1.0` registration metadata alone add no supported combination. Requirements: REQ-NET-01-0022.
- Requirement aspects exercised:

  - REQ-NET-01-0022: The emitted result has schema_version 0.1.0 and test NET-01 0.1.0 and is accepted by the result decoder; tasks with a well-formed unregistered test_version are task_unsupported and every other unsupported, malformed, non-string, null or missing test_version or schema_version is task_schema_invalid, with no handle, replay commit, platform read, dialer use or result; results with an unsupported, malformed, non-string, null or missing schema_version, test.version or test.result_version, test.id NET-03 or TCP-01 with the NET-01 shape, or a TCP-01 measurement shape under NET-01 are rejected.

<a id="M-NET-03-ADMISSION"></a>

### M-NET-03-ADMISSION

- Case: `M-NET-03-ADMISSION`
- Result: passed
- Tests: `TestMNet03Admission` in `net03_conformance_test.go` (package `root`); `TestMNet03AdmissionTypesAndStatuses` in `net03_closeout_test.go` (package `root`); `TestMNet03Execution` in `net03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet03Admission$' ./.`
- Scope: Validate HTTP and HTTPS registered endpoints, required expected_status and optional lowercase SHA-256; vary all types, presence, URL length and digest syntax. Accept every final non-redirect expected_status from 200 to 599 and reject 301, 302, 303, 307, 308 and out-of-range statuses (REQ-NET-03-0043). Reject unsafe/user-specific targets and unknown fields before contact; use a redirect canary destination to prove it is never contacted.
- Requirement aspects exercised:

  - REQ-NET-03-0004: Only a task whose complete URL, scheme, expected status and digest tuple matches the one locally registered validation profile is admitted (any other is target_not_registered with no contact), URLs with credentials are rejected, and the executed GET carries no Cookie, Authorization or User-Agent header.
  - REQ-NET-03-0009: https and http URLs with and without a lowercase digest are admitted; a missing url or expected_status, a non-string url, a string, fractional, null, boolean or array expected_status, a null, numeric, uppercase, 63-, 65-character or non-hex digest, a non-object input and a relative URL are task_schema_invalid, a 2,049-byte URL is task_limit_exceeded and a 2,048-byte URL passes validation, all with no handle, replay commit, run or contact.
  - REQ-NET-03-0010: A different path, user-specific query, scheme, expected status or digest than the registered profile is target_not_registered before any contact.
  - REQ-NET-03-0011: Credentials, fragments, ftp, loopback (127.0.0.1, ::1), private (10.0.0.1, 192.168.1.1), link-local (169.254.1.1, fe80::1), unspecified (0.0.0.0) and IPv4-mapped IPv6 URLs are task_schema_invalid with no contact.
  - REQ-NET-03-0012: A 302 redirect to a public https, private, plaintext or hostname location is classified captive_portal_suspected/redirect with exactly one dial, no dial to any other destination and one server request.
  - REQ-NET-03-0013: An unknown input field and a duplicate url member are task_schema_invalid before contact.
  - REQ-NET-03-0014: Every invalid input is rejected at admission with no task handle, replay commit, run entry or transport contact, so no Measurement Result is produced.
  - REQ-NET-03-0043: expected_status 100, 101, 199, 301, 302, 303, 307, 308 and 600 are task_schema_invalid, while 200, 204, 300, 304, 305, 306, 404 and 599 are admitted with no contact.

<a id="M-NET-03-OPTIONS"></a>

### M-NET-03-OPTIONS

- Case: `M-NET-03-OPTIONS`
- Result: passed
- Tests: `TestMNet03Options` in `net03_conformance_test.go` (package `root`); `TestMNet03OptionsDefaultAndUnsupportedFamily` in `net03_closeout_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet03Options$' ./.`
- Scope: Exercise auto/ipv4/ipv6, omitted default and unsupported family; reject redirect-following, custom-header, credential and body options. No option may bypass local policy.
- Requirement aspects exercised:

  - REQ-NET-03-0015: Omitted, auto and ipv4 families are admitted for an IPv4 literal and ipv6 for a hostname, and the omitted family normalizes to auto; unknown and null families and ipv6 for an IPv4 literal are task_schema_invalid; ipv6 on a network without IPv6 ends incomplete with constraints_unsatisfied, no measurement and no contact.
  - REQ-NET-03-0016: follow_redirects, max_redirects, headers, cookies, credentials, method, body and max_body_bytes options and a proxy network_path are task_schema_invalid with no handle, replay commit, run or contact.

<a id="M-NET-03-PRECONDITIONS"></a>

### M-NET-03-PRECONDITIONS

- Case: `M-NET-03-PRECONDITIONS`
- Result: passed
- Tests: `TestMNet03Preconditions` in `net03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet03Preconditions$' ./.`
- Scope: Run the portable NET-03 gate against a fake platform and synthetic local profile registry with a deterministic clock. Stable network generation unavailable: assert execution.state not_run, no request sent. Validation target or expected response profile not locally available, and present but not current: assert not_run before any request. All preconditions satisfied: assert exactly one request proceeds. Fake platform captive-portal flow that would display, submit or retain user data (iterate each): assert the gate denies, no request is sent and no flow data is retained. Stricter task execution constraint or local probe policy denial: assert not_run with constraints_unsatisfied or local_policy_rejected respectively; other denials use a registered reason per the public table. Requirements: REQ-NET-03-0017, REQ-NET-03-0018, REQ-NET-03-0019, REQ-NET-03-0020. Judgement-based completion is the separate review case RV-M-NET-03-PRECONDITIONS; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-NET-03-0017: Through the shared gates, an unreadable network access state at reservation defers the run (`run_not_due`, no result), and a network change after reservation to a transport the task does not allow gives `not_run` with `constraints_unsatisfied`, both with zero contacts (no dial and no other contact).
  - REQ-NET-03-0018: Tasks whose validation URL path, expected status or expected body digest differs from the local registration are rejected at admission with zero contacts, and a registration that lapses after admission gives `local_policy_rejected` with no contact. With all preconditions met, for platform validation states `validated` and `captive_portal`, exactly one HTTP request and one dial are made and no other contact.
  - REQ-NET-03-0019: Through the `TestMNet03Minimization` subtest, for each canary case (including a portal-style redirect, a content mismatch and a platform-reported captive portal) no redirect Location, response body, Set-Cookie value, authentication state, platform portal URL or interface canary appears in the result, pending store, bundle or errors, a redirect is recorded with `followed_count` 0, and the result decoder rejects injected URL, body, cookie, authentication or portal fields. No gate denial of a captive-portal flow is asserted, since the library has no such flow; whether a native path can trigger one is a native row (`RV-M-NET-03-PRECONDITIONS`), see docs/measurement-requirements.md.
  - REQ-NET-03-0020: The shared portable gates assert, each with zero contacts: a task transport condition the network does not meet, a task excluding VPN on a VPN-detected network and a local charging-only device restriction defer the run at `StartRun` (`run_not_due`, no result); a transport condition lost after reservation gives `not_run` with `constraints_unsatisfied`; and a proxy-path task with no proxy configured is either rejected at admission with `task_schema_invalid` or ends `not_run` with `proxy_not_configured` (the test accepts either); a deployment that withholds the target registration rejects the task with `target_not_registered`, and a registration that lapses after admission gives a result with reason `local_policy_rejected`.

<a id="M-NET-03-EXECUTION"></a>

### M-NET-03-EXECUTION

- Case: `M-NET-03-EXECUTION`
- Result: passed
- Tests: `TestMNet03Execution` in `net03_conformance_test.go` (package `root`); `TestMNet03PlatformReadFailure` in `net03_closeout_test.go` (package `root`); `TestMNet03Termination` in `net03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet03Execution$' ./.`
- Scope: Record NET-01 stability check, platform validation capture, one request without credentials/cookies or redirect following, bounded comparison, and close. Test HTTP and HTTPS independently, with TLS only for HTTPS; keep a destination canary behind every redirect.
- Requirement aspects exercised:

  - REQ-NET-03-0005: For https and http, the platform state is read and retained as platform_validation and then one bounded GET is compared with the registered status (response_class expected); a failing platform read is recorded as unknown while the request still decides, and a generation change during the platform read ends the attempt before any dial.
  - REQ-NET-03-0006: Each run makes three access reads (reservation check, attempt capture, NET-03's platform read), exactly one connection and one GET to the registered path with no body, User-Agent, Cookie, Authorization, If-None-Match or If-Modified-Since, follows no redirect, and leaves no server connection open after the result.

<a id="M-NET-03-NORMALIZATION"></a>

### M-NET-03-NORMALIZATION

- Case: `M-NET-03-NORMALIZATION`
- Result: passed
- Tests: `TestMNet03Normalization` in `net03_conformance_test.go` (package `root`); `TestMNet03ResponseClassBoundaries` in `net03_closeout_test.go` (package `root`); `TestMNet03Execution` in `net03_conformance_test.go` (package `root`); `TestMNet03Termination` in `net03_conformance_test.go` (package `root`); `TestMNet03Limits` in `net03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet03Normalization$' ./.`
- Scope: Exercise every failure/outcome row and active-response class with the platform state varied independently, including conflicting/stale/unknown state. Preserve both evidence sources and test expected, mismatch, redirect and unrecognized responses. Classify valid responses per REQ-NET-03-0041, with no response_class for a header block over the bound or an exchange without a final status, and select the outcome in the REQ-NET-03-0042 order: the response class first, platform captive_portal or not_validated only without a class, then the DNS, TCP or TLS failure, otherwise inconclusive.
- Requirement aspects exercised:

  - REQ-NET-03-0007: An expected response on an unchanged generation is validated under each of the validated, not_validated, captive_portal and unknown platform states; a generation change during the platform read is incomplete network_changed with no measurement.
  - REQ-NET-03-0008: Completed runs produce captive_portal_suspected (platform portal, redirect or content mismatch), not_validated, dns_failure, connection_failure, tls_failure, unexpected_response and inconclusive, and a generation change is the incomplete network_changed state without a measurement rather than an outcome.
  - REQ-NET-03-0027: Each outcome row is selected with its terminal evidence: expected is validated under any platform state; redirect and content_mismatch are captive_portal_suspected; unexpected is unexpected_response; a failed dns_query, tcp_connect or tls_handshake is the last observation of dns_failure, connection_failure or tls_failure; and an exchange without a valid response is inconclusive under validated or unknown platform state.
  - REQ-NET-03-0028: An https URL records tcp_connect, tls_handshake and http_exchange and an http URL records tcp_connect and http_exchange; an untrusted certificate is tls_failure; redirects are never followed.
  - REQ-NET-03-0041: Statuses 301, 302, 303, 307 and 308 are redirect; a matching status is expected without a digest or with a matching digest and content_mismatch with a different one; 204, 300, 304 and 503 against expected 200, gzip with a digest, a premature body end, malformed chunked framing and a body-read timeout are unexpected; gzip without a digest is expected; a response without a status line and a header block over 16,384 bytes have no response_class and a failed http_exchange.
  - REQ-NET-03-0042: For every response class and for no-response, TCP, TLS and DNS failures under each of the four platform states, the outcome follows the five-step order with platform_validation unchanged, and without a class or decisive platform state tcp_failure, dns_failure, tls_failure, timeout and protocol_error map to connection_failure, dns_failure, tls_failure, inconclusive and inconclusive.

<a id="M-NET-03-LIMITS"></a>

### M-NET-03-LIMITS

- Case: `M-NET-03-LIMITS`
- Result: passed
- Tests: `TestMNet03Limits` in `net03_conformance_test.go` (package `root`); `TestMNet03PrefixStopsReading` in `net03_closeout_test.go` (package `root`); `TestMNet03Termination` in `net03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet03Limits$' ./.`
- Scope: Test 16,384-byte response headers and each applicable registered response-prefix bound at below/equal/above boundaries, outer deadline, zero retries and one active run per generation. Count actual operations and buffered reads separately; delay or reject overlapping runs as specified. A digest undecidable within the prefix is response_class unexpected, not resource_limit (REQ-NET-03-0041).
- Requirement aspects exercised:

  - REQ-NET-03-0021: Bodies of 0, 1,023, 1,024, 1,025 and 4,096 bytes record bytes_processed of at most 1,024 with truncated only above it; header blocks of 16,383 and 16,384 bytes are classified and 16,385 is not; with a 16,384-byte body a received budget of exactly the head plus 1,024 bytes completes and one byte less is resource_limit; every run makes one request and one connection with no retry; a second run waits without contact while the first is active and proceeds after it; and the outer attempt deadline while awaiting the response is incomplete attempt_deadline_exceeded.
  - REQ-NET-03-0041: With a configured digest, a body over the 1,024-byte prefix is response_class unexpected and outcome unexpected_response rather than resource_limit, without a digest a truncated body stays expected, and a header block over 16,384 bytes has no class and selects inconclusive under unknown and captive_portal_suspected under captive_portal platform state.

<a id="M-NET-03-TERMINATION"></a>

### M-NET-03-TERMINATION

- Case: `M-NET-03-TERMINATION`
- Result: passed
- Tests: `TestMNet03Termination` in `net03_conformance_test.go` (package `root`); `TestNet03AllTerminationBoundaries` in `internal/execution/net03_termination_test.go` (package `internal/execution`); `TestMNet03Normalization` in `net03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet03Termination$' ./.`
- Scope: Drive NET-03 with a synthetic registered validation target, a fake platform validation source, a scripted DNS/TCP/TLS/HTTP transport and a deterministic clock, and inject each REQ-NET-03-0029 interruption at every Procedure point: during the NET-01 network-generation stability check; after it and before the platform validation/captive-portal state read; during that read; before and during the dns_query; before and during the tcp_connect; before and during the tls_handshake for an https URL only (no tls_handshake point exists for an http URL); after the request is sent and while response headers are read; while the bounded response prefix is read; and after the first response or failure but before the connection is closed. Expected states follow the GO-SPEC-ISSUE-004 disposition: the NET-01 network-generation stability check starts the test, so every point from it until the outcome is recorded is incomplete as asserted below; an interruption before it is not_run with no measurement and no network_context; and after the first response or failure but before the connection is closed the attempt stays completed with the recorded outcome. For each incomplete point iterate the five interruptions named by the requirement: a network-generation change (expect execution.state incomplete, reason network_changed), host cancellation (incomplete, cancelled_by_host), expiry of the outer attempt deadline as distinct from the bounded operation timeout (incomplete, attempt_deadline_exceeded), an outer non-time resource bound (incomplete, resource_limit) and platform suspension (incomplete, platform_suspended). Assert every such result has no measurement object at all (no outcome, observations, platform_validation or response_class, since NET-03 is not the NET-05 partial-result exception), carries exactly one reason naming the first termination condition, and that injecting two conditions keeps only the first. Assert that a bounded protocol timeout, DNS/TCP/TLS failure, redirect or content mismatch occurring without an outer interruption instead yields state completed with a measurement, that no retry, second request or redirect request is recorded after any interruption, that the connection and goroutines are released, and that no terminal protocol outcome or captured platform state is invented for an interrupted stage. Requirements: REQ-NET-03-0029.
- Requirement aspects exercised:

  - REQ-NET-03-0029: For https and http, an interruption before the test starts is not_run with no measurement or network_context; during and after the stability check, during and after the platform read, during and after DNS, during and after TCP connect, during TLS (https only), while awaiting headers and while reading the prefix, each of host cancellation, network_changed, attempt deadline, resource_limit and injected suspension is incomplete with only the first of two injected reasons, no measurement, fixed dial and lookup counts, closed connections, no timers and goroutines back to baseline; release after the outcome stays completed; through Client, a generation change in the platform read, host cancellation and the deadline are incomplete, while a bounded protocol timeout and DNS, TCP or TLS failures, redirects and content mismatches complete with a measurement.

<a id="M-NET-03-MINIMIZATION"></a>

### M-NET-03-MINIMIZATION

- Case: `M-NET-03-MINIMIZATION`
- Result: passed
- Tests: `TestMNet03Minimization` in `net03_conformance_test.go` (package `root`), `TestMNet03Safety` in `measurement_safety_conformance_test.go` (package `root`), `TestMNet03Preconditions` in `net03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet03Minimization$' ./.`
- Scope: Configure a synthetic NET-03 task and scripted transport that place a distinct canary in every item REQ-NET-03-0030 names: the validation URL (scheme, host, path and query canaries), the resolved validation address, a 302 Location redirect URL with host, path, query and fragment canaries, the response body, the body SHA-256 digest (including the expected_body_sha256 input value), HTTP header names and header values (custom response headers), a Set-Cookie value, an authentication-state canary, a platform portal URL and interface name/index from the fake platform validation source, and a raw transport/platform error string for each DNS, TCP, TLS and HTTP failure path. Iterate validated, redirect, content_mismatch, unexpected, dns_failure, connection_failure, tls_failure, not_validated and incomplete paths, serialize the complete result and every nested observation and error path, and assert no canary appears in any byte of the encoding. Assert the measurement contains only outcome, observations, platform_validation and conditional response_class; observations contain only type, status, duration_ms, normalized failure and the registered details (dns_query engine/query_type/responses, tcp_connect address_family, tls_handshake registered fields, http_exchange address_family, protocol, status_code, redirect with followed_count 0 and terminal_location_class only, body bytes_processed/truncated, checks kind/index/matched, content_encoding); and that unknown or extra fields are rejected. Raw errors map to registered failure values or inconclusive. No participant or live measurement data is used. Requirements: REQ-NET-03-0030.
- Requirement aspects exercised:

  - REQ-NET-03-0030: Unique canaries are placed in the validation URL host, path and query, the resolved address, the redirect Location host, path, query and fragment, the response body, the expected and observed body digests, a response header name and value, a Set-Cookie value, an Authentication-Info value, the certificate subject, issuer and serials, raw DNS, TCP, TLS and x509 errors and the status text; the platform portal URL and interface name and index canaries stay withheld beside the fake, whose validation port has no field for them. Over the `validated`, redirect and content-mismatch `captive_portal_suspected`, `unexpected_response`, `dns_failure`, `connection_failure`, `tls_failure`, `not_validated`, `inconclusive` and host-cancelled `incomplete` runs, no canary appears in any searched encoding in any result, the bundle, the pending store's bytes, any returned error or the task-validation rejection of a task whose unknown input member carries the canaries, every measurement path is within the registered NET-03 field set, a redirect keeps only `followed_count` 0 and `terminal_location_class`, and the strict decoder rejects `url`, `address`, `location`, `body`, `digest`, `headers`, `cookie`, `authentication`, `portal_url`, `interface` or `error` injected into the measurement, the last observation or its details.

<a id="M-NET-03-INTERPRETATION"></a>

### M-NET-03-INTERPRETATION

- Case: `M-NET-03-INTERPRETATION`
- Result: passed
- Tests: `TestMNet03Interpretation` in `net03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet03Interpretation$' ./.`
- Scope: Assert the serialized NET-03 measurement field set is exactly outcome, observations, platform_validation and conditional response_class, with outcome limited to validated, captive_portal_suspected, not_validated, dns_failure, connection_failure, tls_failure, unexpected_response and inconclusive, so no interpretation, verdict, censorship, operator or general-reachability field exists. Assert the Go field descriptions, documented enum descriptions and sanitized example and host-facing text for every outcome contain none of these prohibited claims: that validated proves arbitrary or general Internet destinations are reachable; that captive_portal_suspected, redirect or content_mismatch identifies the operator or purpose of an intercepted page; that not_validated or any unvalidated result is evidence of censorship; that a single result is definitive proof of censorship or a generic blocking, anomaly or success verdict; and that platform_validation overrides or confirms the active request. Exercise synthetic conflicting platform and active observations and assert outcome inconclusive rather than a stronger claim. Requirements: REQ-NET-03-0001, REQ-NET-03-0002, REQ-NET-03-0003, REQ-NET-03-0031, REQ-NET-03-0032, REQ-NET-03-0033. Judgement-based completion is the separate review case RV-M-NET-03-INTERPRETATION; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-NET-03-0001: No non-negated sentence in the comments and string literals of production Go files and sanitized examples and the shipped Markdown documentation (case plans, manifests, evidence, changelog and specification-issue records excluded) matches the "validated proves general reachability" claim pattern, and the NET-03 section of the interpretation guidance contains "does not establish that arbitrary or general internet destinations are reachable". This is a textual scan of shipped text; it does not prove how a reader interprets a result.
  - REQ-NET-03-0002: The same scan finds no non-negated sentence stating that `captive_portal_suspected`, a redirect or a content mismatch identifies an operator or purpose, and the guidance contains "does not identify the operator or the purpose of an intercepting page".
  - REQ-NET-03-0003: The same scan finds no non-negated sentence calling `not_validated` or an unvalidated result evidence or proof of censorship, the guidance contains "are not evidence of censorship", and no member name or string value of any run's measurement contains a verdict token (such as `verdict`, `censor`, `block`, `filter`, `anomal`, `confiden`) and the strict result decoder rejects `verdict`, `interpretation`, `censored`, `blocked`, `filtered`, `anomaly`, `confidence`, `reachable` and `authoritative` fields injected into a valid result.
  - REQ-NET-03-0031: The nine completed runs carry registered outcomes, the guidance contains "establishes whether the configured validation endpoint returned its expected response, appeared to be intercepted, or could not be validated on the observed network path" and "does not override or confirm it", and no sentence states that `platform_validation` overrides or confirms the active request; a platform `validated` state with no valid response yields `inconclusive` without `response_class`, and a platform `captive_portal` state with the expected response yields `validated`.
  - REQ-NET-03-0032: Not asserted: the test requires none of the listed evidence items in the guidance and checks none of them.
  - REQ-NET-03-0033: No non-negated sentence of the scanned text matches the common claims (a result is or provides definitive or conclusive proof of censorship or blocking; a result proves, demonstrates or confirms censorship or blocking; a censorship, blocking, anomaly or success verdict), and no member name or string value of any run's measurement contains a verdict token (such as `verdict`, `censor`, `block`, `filter`, `anomal`, `confiden`) and the strict result decoder rejects `verdict`, `interpretation`, `censored`, `blocked`, `filtered`, `anomaly`, `confidence`, `reachable` and `authoritative` fields injected into a valid result; the test does not require a definitive-proof qualifier in the NET-03 guidance.

<a id="M-NET-03-SAFETY"></a>

### M-NET-03-SAFETY

- Case: `M-NET-03-SAFETY`
- Result: passed
- Tests: `TestMNet03Safety` in `measurement_safety_conformance_test.go` (package `root`); `TestMNet03Minimization` in `net03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet03Safety$' ./.`
- Scope: Iterate the registered Baseline Risk rows (risk level low; requires user consent false, subject to host and local policy) and each Risk Factors row (network_observability low, traffic_distinctiveness low, target_sensitivity low, data_sensitivity low, legal_or_social_context low) with synthetic task safety declarations: assert effective risk is never below low, that a task declaring risk_level medium or high or any factor at medium or high raises the effective risk to at least that level, that an unknown factor keeps the overall level unknown unless trusted local assessment resolves it, that requires_user_consent true blocks execution until the fake host establishes consent, and that consent false adds no requirement while host or local policy may still decline. Exercise each normative safety action with synthetic targets and a recording transport: data minimization (no response body, full redirect URL, cookie or unrelated header in the result); traffic distinctiveness (exactly one validation request per run, zero retries, and no added repeated requests when the fake platform reports reliable state); user identification (the recorded request carries no identifiers, cache validators, cookies or application authentication); target risk (a target absent from the locally registered benign validation set is rejected and never contacted, and a redirect target is never contacted); and probe authority (fake local policy can delay, skip or reject execution, yielding not_run with local_policy_rejected when rejected). Requirements: REQ-NET-03-0034, REQ-NET-03-0035, REQ-NET-03-0036, REQ-NET-03-0037, REQ-NET-03-0038, REQ-NET-03-0039, REQ-NET-03-0040. Judgement-based completion is the separate review case RV-M-NET-03-SAFETY; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-NET-03-0034: The test asserts that the production registry records a `low` baseline for NET-03 and that a task declaring `low`, `medium`, `high` or `unknown` has effective risk exactly that level (admitted under issuer and local bounds permitting it, rejected `task_issuer_not_authorized` with no contact when either side permits only lower levels). For the consent baseline it asserts that `false` adds no requirement (with host consent absent the run starts, yields one result and never queries the consent capability), while consent required by the task, by an issuer or by the local policy blocks every result and contact unless the fake host reports valid consent (absent, ended and unavailable all block).
  - REQ-NET-03-0035: For each of the five factor types the test asserts that the registered baseline equals the most restrictive of the fixture's row levels (all five `low`), that declaring the factor at `low`, `medium` or `high` gives effective risk exactly the maximum of baseline, row level and declared level, that an `unknown` factor gives `unknown`, and that an unknown factor under a known overall level or a factor above the overall level is rejected `task_schema_invalid` and an unrecognized factor type `task_unsupported`, all with no contact. The row levels are fixture constants checked only against the registered baseline, and the rationale text is not asserted; NET-03 has no per-factor action rows.
  - REQ-NET-03-0036: The test asserts that no surface contains the validation URL host, path and query, resolved address, redirect `Location` parts, response body, body digests, response header name and value, `Set-Cookie` value, authentication state, platform portal URL, interface name and index, certificate fields or raw errors, and that the measurement holds only `outcome`, `platform_validation`, `response_class` and the closed primitive observation fields, with redirect details limited to `followed_count` 0 and the terminal location class. The decoder rejects injected URL, address, location, body, digest, headers, cookie, authentication, portal URL, interface and error fields.
  - REQ-NET-03-0037: For platform states `validated`, `not_validated` and `unknown` the test asserts exactly one validation request and one dial per run with no dial to any other address, and a single dial after a failure (no retry). It does not assert that the request is skipped when the platform reports validated state, and synchronized schedules are not asserted.
  - REQ-NET-03-0038: The test asserts that each validation request is a GET carrying none of `Cookie`, `Authorization`, `Proxy-Authorization`, `If-None-Match`, `If-Modified-Since`, `Referer`, `Origin`, `X-Requested-With`, `X-Client-Id` or `X-Device-Id` and no header value containing `safety-` or `barometer`.
  - REQ-NET-03-0039: The test asserts that an unregistered address, path or expected status and a loopback target are rejected at admission with no dial, and that a 302 redirect to another host is never contacted (one dial, none elsewhere). Repeated execution and the residual-risk statement are not asserted beyond the shared harness rows.
  - REQ-NET-03-0040: The test asserts that a local policy not permitting the task's network path rejects a valid task at admission with no contact, and that a local charging-only device constraint holds an admitted valid task (`run_not_due` twice, no result, no contact) until satisfied, after which the same task runs; a target registration that expires after admission ends the attempt `local_policy_rejected` with no contact. The no-lowering clause is asserted through the declared-level rows: no task declaration yields effective risk below the `low` baseline.

<a id="M-NET-03-OUTPUT"></a>

### M-NET-03-OUTPUT

- Case: `M-NET-03-OUTPUT`
- Result: passed
- Tests: `TestMNet03Output` in `net03_conformance_test.go` (package `root`); `TestMNet03Execution` in `net03_conformance_test.go` (package `root`); `TestMNet03Normalization` in `net03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet03Output$' ./.`
- Scope: Check required platform_validation, conditional response_class and its complete enum, minimized redirect metadata, and protocol observations actually begun. No redirect location is retained or contacted; platform and active evidence remain distinct.
- Requirement aspects exercised:

  - REQ-NET-03-0023: With platform captive_portal and an expected response the result keeps platform_validation captive_portal beside response_class expected and outcome validated, and across all classes and failures each platform state is retained unchanged.
  - REQ-NET-03-0024: platform_validation is present with each of validated, not_validated, captive_portal and unknown, the measurement has only outcome, observations, platform_validation and response_class, and response_class is present exactly when a valid response was received.
  - REQ-NET-03-0025: expected, redirect, content_mismatch and unexpected are each emitted, for a matching response, a redirect status, a digest mismatch and other responses.
  - REQ-NET-03-0026: Redirect results carry a redirect object with only followed_count 0 and terminal_location_class (unregistered_https, prohibited_host, non_https, credentials, invalid), no location canary appears in the result and the location is never contacted.

<a id="M-NET-03-VERSIONS"></a>

### M-NET-03-VERSIONS

- Case: `M-NET-03-VERSIONS`
- Result: passed
- Tests: `TestMNet03Versions` in `net03_conformance_test.go` (package `root`); `TestMNet03VersionShapes` in `net03_closeout_test.go` (package `root`); `TestMNet03Output` in `net03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMNet03Versions$' ./.`
- Scope: Assert the NET-03 registration emits exactly one combination: test.id NET-03, test.version 0.1.0, Measurement Result schema_version 0.1.0 and test.result_version 0.1.0. Reject before execution, as a task-validation failure that produces no Measurement Result, a NET-03 task naming test version 0.2.0, 1.0.0, 0.1, v0.1.0, 0.1.0 with surrounding whitespace, an empty string, a JSON number or a missing version. Reject in the result decoder, without best-effort interpretation or silent field dropping, results with schema_version or test.result_version set to each of the same unsupported or malformed values, results whose test.version and result_version are individually valid but differ from the registered combination, and cross-test results: test.id NET-03 with a NET-01 access-only Output or a NET-05 profile Output, and test.id NET-01 or NET-05 carrying the NET-03 Output (platform_validation and response_class). Assert that an unknown field inside a known 0.1.0 version is rejected, that no version is coerced, defaulted or added, and that the Draft status and registration metadata of tests/net-03.md do not by themselves cause any other version to be accepted. Requirements: REQ-NET-03-0022.
- Requirement aspects exercised:

  - REQ-NET-03-0022: Results carry schema_version, test.version and result_version 0.1.0 for NET-03 and decode; task versions 0.2.0 and 1.0.0 are task_unsupported and 0.1, v0.1.0, trailing whitespace, empty, numeric and missing versions are task_schema_invalid with no run, replay commit or contact; the result decoder rejects each unsupported or malformed schema_version, test.version and result_version, NET-01, NET-05 and HTTP-01 identities, whole NET-01 access-only and NET-05 profile measurements, added access or profile members, missing or invalid NET-03 fields, a contradicting class, foreign outcomes and an unknown field.

<a id="M-DNS-03-ADMISSION"></a>

### M-DNS-03-ADMISSION

- Case: `M-DNS-03-ADMISSION`
- Result: passed
- Tests: `TestMDns03Admission` in `dns03_conformance_test.go` (package `root`); `TestMDns03ExecutionAndOutput` in `dns03_conformance_test.go` (package `root`); `TestMDns03Normalization` in `dns03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns03Admission$' ./.`
- Scope: Vary resolver_url HTTPS/2,048-byte syntax, reviewed hostname/A-or-AAAA and optional canonical bootstrap address against the exact registered resolver. Test credentials/fragments/user parameters/local/interpolated/unknown targets. Bootstrap IP must not change certificate hostname authentication; redirect targets are never contacted.
- Requirement aspects exercised:

  - REQ-DNS-03-0004: Client admission accepts a DNS-03 task only when both the query name and the complete resolver profile are registered; a different resolver hostname or a different query name alone is target_not_registered with no lookup, dial or replay commit.
  - REQ-DNS-03-0009: resolver_url, hostname, record_type and bootstrap_ip are checked against their constraints: https URLs up to 2,048 bytes, 253-byte and IDNA names, A and AAAA, and canonical public IPv4 and IPv6 bootstrap addresses are admitted; non-https, credential-bearing, fragment, interpolated, IP-literal, local, special-use, malformed-port and relative URLs, IP-literal, single-label, local, interpolated and over-long names, other record types and non-canonical or mapped addresses are task_schema_invalid, and a 2,049-byte URL is task_limit_exceeded.
  - REQ-DNS-03-0010: A different path, an added query string, a different port or bootstrap address, or a bootstrap address added or omitted relative to the registered resolver profile is target_not_registered without lookup or dial.
  - REQ-DNS-03-0011: With bootstrap_ip the single ClientHello names the URL hostname and the handshake is trusted (TestMDns03ExecutionAndOutput, direct and via CONNECT to the bootstrap address); with bootstrap_ip a certificate for another name or one valid only for the bootstrap address is tls_failure with certificate_name_mismatch and validation name_mismatch (TestMDns03Normalization).
  - REQ-DNS-03-0012: A 307 response carrying a Location is http_failure with its status retained, exactly one dial and one request, and no contact with the redirect target (TestMDns03Normalization).
  - REQ-DNS-03-0013: Credential-bearing and interpolated resolver URLs, local and localhost resolver and query names, and loopback, private, CGNAT, link-local, unique-local and documentation bootstrap addresses are task_schema_invalid before any lookup or dial.
  - REQ-DNS-03-0014: Unknown input fields, including method and headers members, and missing required fields are task_schema_invalid.
  - REQ-DNS-03-0015: Every invalid input is rejected at admission with no handle, replay commit, lookup, dial or result record.

<a id="M-DNS-03-OPTIONS"></a>

### M-DNS-03-OPTIONS

- Case: `M-DNS-03-OPTIONS`
- Result: passed
- Test: `TestMDns03Options` in `dns03_conformance_test.go` (root package, Go package `barometer`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns03Options$' ./.`
- Scope: Exercise auto/ipv4/ipv6 and compatibility with bootstrap_ip, including missing/default/invalid/unknown cases. Reject GET, redirects, alternate endpoints, custom headers, credentials/cookies and task proxy configuration.
- Requirement aspects exercised:

  - REQ-DNS-03-0016: Omitted, auto, ipv4 and ipv6 address_family values are admitted when compatible with the bootstrap address or with system bootstrap; an incompatible family and null, empty, mis-cased, numeric or unknown values are task_schema_invalid without dial or replay commit.
  - REQ-DNS-03-0017: Options requesting GET, redirect following, alternate endpoints, custom headers, credentials, cookies, a proxy, connection reuse or retries, and any unknown option, are task_schema_invalid without dial or replay commit.

<a id="M-DNS-03-PRECONDITIONS"></a>

### M-DNS-03-PRECONDITIONS

- Case: `M-DNS-03-PRECONDITIONS`
- Result: passed
- Tests: `TestMDns03Preconditions` in `dns03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns03Preconditions$' ./.`
- Scope: Vary resolver/query technical/privacy/safety approval, local profile availability, explicit host enablement or approved anomaly-trigger policy and stable generation/deadline. Assert no configuration download or query before gates pass; deployment risk review requires separate evidence. Portable gate tests plus separately required native/host validation; no platform claim.
- Requirement aspects exercised:

  - REQ-DNS-03-0018: Tasks whose query name, resolver host or bootstrap IP differs from the local registration are rejected at admission with zero transport dials and system-resolver lookups. The technical, privacy and safety review itself is not asserted; host row, see docs/measurement-requirements.md.
  - REQ-DNS-03-0019: Asserted only indirectly: a resolver profile that differs from the local registration is rejected at admission, and every gate row records zero dials and lookups. No row checks for a configuration download as such.
  - REQ-DNS-03-0020: A deployment that lists only another test and no registrations rejects the task with `task_issuer_not_authorized` and no contact. No anomaly-triggered path is asserted; the deployment review is a host row, see docs/measurement-requirements.md.
  - REQ-DNS-03-0021: Through the shared gates, an unreadable network access state at reservation defers the run (`run_not_due`, no result), and a network change after reservation to a transport the task does not allow gives `not_run` with `constraints_unsatisfied`, both with zero contacts. Execution deadline availability is not asserted.
  - REQ-DNS-03-0022: The shared portable gates assert, each with zero contacts: a task transport condition the network does not meet, a task excluding VPN on a VPN-detected network and a local charging-only device restriction defer the run at `StartRun` (`run_not_due`, no result); a transport condition lost after reservation gives `not_run` with `constraints_unsatisfied`; and a proxy-path task with no proxy configured is either rejected at admission with `task_schema_invalid` or ends `not_run` with `proxy_not_configured` (the test accepts either); a deployment that withholds the target registration rejects the task with `target_not_registered`, and a registration that lapses after admission gives a result with reason `local_policy_rejected`; contacts count both transport dials and system-resolver lookups.

<a id="M-DNS-03-EXECUTION"></a>

### M-DNS-03-EXECUTION

- Case: `M-DNS-03-EXECUTION`
- Result: passed
- Tests: `TestMDns03ExecutionAndOutput` in `dns03_conformance_test.go` (package `root`); `TestDns03PlaintextAccounting` in `internal/execution/dns03_termination_test.go` (package `internal/execution`); `TestMDns03Normalization` in `dns03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns03ExecutionAndOutput$' ./.`
- Scope: Record optional system bootstrap, one selected resolver TCP connection, authenticated TLS, one RFC 8484 POST with application/dns-message and bounded response validation. Vary returned media type, DNS ID/name/type mismatch, malformed/oversized message and each failed stage; prohibit redirects, endpoint/transport fallback and application retries.
- Requirement aspects exercised:

  - REQ-DNS-03-0005: Through Client with the production registry, direct with and without system bootstrap and via CONNECT with a bootstrap address, a run makes at most one bootstrap lookup, one connection, one ClientHello naming the URL hostname with ALPN http/1.1, one trusted handshake and one POST to the registered path with application/dns-message Content-Type and Accept, identity encoding, a 31-byte query with ID 0 and no other request header; TestDns03PlaintextAccounting repeats the single request on the direct and proxy paths.
  - REQ-DNS-03-0006: Observations follow the Procedure order (system dns_query only without bootstrap_ip, then tcp_connect, tls_handshake, http_exchange and the doh dns_query); a system bootstrap returning two addresses dials only the selected one, whose refusal is tcp_failure with no second dial; each failed-stage row in TestMDns03Normalization stops with one dial and at most one request and no retry, fallback or redirect contact.

<a id="M-DNS-03-NORMALIZATION"></a>

### M-DNS-03-NORMALIZATION

- Case: `M-DNS-03-NORMALIZATION`
- Result: passed
- Tests: `TestMDns03Normalization` in `dns03_conformance_test.go` (package `root`); `TestPOSTStatusPrecedesMediaType` in `internal/protocol/http/exchange_test.go` (package `internal/protocol/http`); `TestMDns03Limits` in `dns03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns03Normalization$' ./.`
- Scope: Exercise all terminal outcome rows, including non-success HTTP, content type, DNS response-code failures, malformed/mismatching DNS, per-stage timeout and inconclusive errors. Attribute failure only to the actually observed stage and preserve ordered preceding observations.
- Requirement aspects exercised:

  - REQ-DNS-03-0007: A NOERROR message matching the query identifier, name and type is outcome response (with an answer, empty, or with media-type parameters); NXDOMAIN, SERVFAIL, REFUSED and an unregistered code are parsed DNS outcomes retaining their rcode; an identifier, name or type mismatch is query_mismatch.
  - REQ-DNS-03-0008: Each failure class is produced from loopback behavior or a wrapped platform error at its observed stage: bootstrap_failure (not found), tcp_failure (refused, reset, unreachable), tls_failure (reset, untrusted, name mismatch, alert, protocol error), http_failure (503, 404 with a DNS body and 307 without a DNS parse, unsupported encoding, premature body end, malformed response), unexpected_content_type (other or missing media type), dns_nxdomain, dns_servfail, dns_refused, malformed_dns_response (empty, invalid, oversized), query_mismatch, timeout (bootstrap and TCP here, TLS and HTTP in TestMDns03Limits) and inconclusive (unclassified TCP and TLS errors, unregistered rcode); network_changed is an incomplete execution state, not an outcome (M-DNS-03-TERMINATION).
  - REQ-DNS-03-0027: For each outcome the terminal observation type, status and failure match the row, prior observations are succeeded and ordered, http_failure retains the status code, unexpected_content_type retains no media-type value, tls_failure retains the validation class, and every result passes bundle validation; TestPOSTStatusPrecedesMediaType shows a non-2xx POST ends at its status before encoding and media-type checks without reading the body or classifying a Location, while a 2xx HTML response is unexpected_content_type (GO-SPEC-ISSUE-020).
  - REQ-DNS-03-0040: A 503 HTML response, a 404 carrying `application/dns-message` and a 307 are `http_failure` with the status retained and no DNS parse; `TestPOSTStatusPrecedesMediaType` shows the exchange checks no encoding or media type of a non-2xx response, reads no body and classifies no `Location`, while a 2xx HTML response is `unexpected_content_type`.

<a id="M-DNS-03-LIMITS"></a>

### M-DNS-03-LIMITS

- Case: `M-DNS-03-LIMITS`
- Result: passed
- Tests: `TestMDns03Limits` in `dns03_conformance_test.go` (package `root`); `TestMDns03TupleConcurrency` in `dns03_conformance_test.go` (package `root`); `TestDns03PlaintextAccounting` in `internal/execution/dns03_termination_test.go` (package `internal/execution`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns03Limits$' ./.`
- Scope: Test the 65,535-byte DNS message and 16 retained answers at below/equal/above boundaries under stricter shared DNS/HTTP/TLS/task limits. Assert one post-bootstrap resolver connection/request, no retry, full-chain deadline and deduplication by resolver/query tuple.
- Requirement aspects exercised:

  - REQ-DNS-03-0023: The TLS and HTTP operation bounds give completed timeout while the outer deadline is open and the outer deadline gives incomplete attempt_deadline_exceeded, each with one connection and the session released; DNS messages of 65,534 and 65,535 bytes are response, 65,536 and 200,000 bytes are malformed_dns_response, and a stricter received limit below the framed response is incomplete resource_limit; 15 and 16 answers are retained and 17 are malformed_dns_response; sent and received budgets equal to the measured request and response complete while one byte less ends the attempt with resource_limit, and metered usage equals the server-side plaintext on direct and proxy paths (TestDns03PlaintextAccounting); every row makes one connection and one request; an equivalent resolver and query tuple waits without contact and its cancellation is not_run (TestMDns03TupleConcurrency).

<a id="M-DNS-03-TERMINATION"></a>

### M-DNS-03-TERMINATION

- Case: `M-DNS-03-TERMINATION`
- Result: passed
- Tests: `TestMDns03Termination` in `dns03_conformance_test.go` (package `root`); `TestDns03TerminationPoints` in `internal/execution/dns03_termination_test.go` (package `internal/execution`); `TestMDns03Limits` in `dns03_conformance_test.go` (package `root`); `TestMDns03Normalization` in `dns03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns03Termination$' ./.`
- Scope: Drive DNS-03 with a synthetic locally registered resolver profile, a scripted system resolver, scripted TCP/TLS/HTTP transports, a deterministic clock and a fake platform, in two variants: without bootstrap_ip (system bootstrap dns_query required) and with an approved synthetic bootstrap_ip (first observation omitted). Interrupt at each Procedure point: after input validation and before bootstrap; during the system bootstrap dns_query; between bootstrap and TCP; during tcp_connect; between TCP and TLS; during tls_handshake (including resolver certificate authentication for the URL hostname); during the single HTTPS POST with media type application/dns-message (request send and response read); during HTTP response validation and bounded DNS message parsing; and after the final doh dns_query observation but before result emission. Expected states follow the GO-SPEC-ISSUE-004 disposition: after input validation and before bootstrap the attempt is not_run with that reason, no measurement and no network_context; after the final doh dns_query observation it stays completed with the recorded outcome; every other point is incomplete as asserted below. At every point inject each REQ-DNS-03-0028 condition separately: network generation change (execution.reason network_changed), host cancellation (cancelled_by_host), outer attempt deadline expiry (attempt_deadline_exceeded), proxy failure or reconfiguration where a host proxy path is in effect (proxy_unavailable, proxy_connect_failed, proxy_reconfigured), and platform suspension (platform_suspended). At the incomplete points, assert execution.state is incomplete with exactly that reason, that no measurement object is present (DNS-03 does not permit NET-05-style partial output), and that when two conditions occur the reason names only the first. Verify contrastively that an active-stage protocol timeout is a completed result with outcome timeout and the terminal DNS, TCP, TLS or HTTP observation failed with its registered timeout, distinct from attempt_deadline_exceeded. Assert via the recording transports that after interruption sockets and TLS/HTTP sessions are released, no further connection or request is made, no fallback to another address, endpoint, transport, GET or redirect target occurs, no application-layer retry occurs, and no unstarted stage or terminal DNS outcome is synthesized. Requirements: REQ-DNS-03-0028.
- Requirement aspects exercised:

  - REQ-DNS-03-0028: For host cancellation, network change, attempt deadline, injected platform suspension, proxy_unavailable, proxy_connect_failed and proxy_reconfigured, with a second cause injected after the first, in both the bootstrap_ip and system-bootstrap variants: before the test starts the attempt is not_run with no network_context; during bootstrap, after bootstrap before TCP, during the connect, after TCP before the ClientHello, during a stalled handshake, after the POST is read and during the response body read it is incomplete with the first reason, no measurement, no further dial or lookup and every connection closed; during close after the outcome is recorded it stays completed with response. Through Client a network change and cancellation after the POST and a refused CONNECT tunnel are incomplete without direct or resolver fallback; active-stage bootstrap, TCP, TLS and HTTP timeouts are completed timeout (TestMDns03Normalization, TestMDns03Limits).

<a id="M-DNS-03-MINIMIZATION"></a>

### M-DNS-03-MINIMIZATION

- Case: `M-DNS-03-MINIMIZATION`
- Result: passed
- Tests: `TestMDns03Minimization` in `dns03_conformance_test.go` (package `root`), `TestMDns03Safety` in `measurement_safety_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns03Minimization$' ./.`
- Scope: Configure a synthetic locally registered resolver profile and scripted system resolver, TCP, TLS and HTTP transports seeded with one unique canary per REQ-DNS-03-0029 item: resolver_url (including path canary), resolver hostname, resolver address (returned by the scripted bootstrap resolver and used for TCP), bootstrap_ip, query hostname, raw DNS message bytes (query and response, including a canary query identifier and canary authority and additional section records), HTTP header names and values (including a canary Content-Type value on an unexpected_content_type response), HTTP body bytes beyond the DNS message, certificate data (canary subject, issuer, serial and fingerprint in the synthetic chain), resolver cookies (Set-Cookie canary), and raw transport, TLS, HTTP, DNS and platform error strings. Iterate the success path and each failure class (bootstrap_failure, tcp_failure, tls_failure, http_failure, unexpected_content_type, dns_nxdomain, dns_servfail, dns_refused, malformed_dns_response, query_mismatch, timeout, inconclusive) plus incomplete terminations, with and without bootstrap_ip. Assert serialized result and error paths, including every nested observation, contain none of the canaries in any encoding and no field outside the allowed set: observation type, status, duration_ms, failure, and details limited to dns_query engine, address_family, query_type, responses (arrival_index, arrival_offset_ms, rcode, answers with type and value), response_limit_reached; tcp_connect address_family; tls_handshake address_family, tls_version, alpn, certificate_validation, alert; http_exchange address_family, protocol, status_code, body (bytes_processed, truncated), content_encoding. Verify retained answers in the doh dns_query are only bounded normalized A, AAAA or CNAME values, at most 16 per response, deduplicated, without TTLs. No participant or live measurement data is used. Requirements: REQ-DNS-03-0029.
- Requirement aspects exercised:

  - REQ-DNS-03-0029: Unique canaries cover the `resolver_url` path and hostname, the resolver address, `bootstrap_ip`, the query hostname and identifier, DNS authority, additional and raw-message patterns, the answer TTL, an HTTP header name and value, a Set-Cookie value, a Content-Type value, HTTP body bytes beyond the DNS message, the status text, the certificate subject, issuer and serials, and raw transport, x509 and TLS errors, plus hexadecimal and base64 forms of every raw DNS query and raw HTTP response the scripted server handled. Over `response` with and without `bootstrap_ip`, `bootstrap_failure`, `tcp_failure`, `tls_failure`, `http_failure`, `unexpected_content_type`, `dns_nxdomain`, `dns_servfail`, `dns_refused`, `malformed_dns_response`, `query_mismatch`, `timeout`, `inconclusive` and host-cancelled `incomplete`, no canary appears in any result, the bundle, the pending store's bytes, any returned error or the task-validation rejection of a task whose unknown input member carries the canaries and every measurement path is within the registered set; the `bootstrap_ip` response keeps answers `9.9.9.9` then `1.1.1.1` deduplicated, and the strict decoder rejects 12 injected members (including `resolver_url`, `bootstrap_ip`, `ttl`, `message`, `headers`, `cookie`, `certificate` and `error`) in the last two observations' details and in an answer. Certificate DER bytes are not searched for.

<a id="M-DNS-03-INTERPRETATION"></a>

### M-DNS-03-INTERPRETATION

- Case: `M-DNS-03-INTERPRETATION`
- Result: passed
- Tests: `TestMDns03Interpretation` in `dns03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns03Interpretation$' ./.`
- Scope: Generate synthetic results for every DNS-03 outcome (response, bootstrap_failure, tcp_failure, tls_failure, http_failure, unexpected_content_type, dns_nxdomain, dns_servfail, dns_refused, malformed_dns_response, query_mismatch, timeout, inconclusive) plus incomplete terminations, including a response whose answers differ from a synthetic system DNS answer set. Assert the serialized result schema has no interpretation, verdict, blocking, censorship, interference, authoritative, correctness or anomaly field, and that unknown fields of that kind are rejected. Assert generated Go field descriptions, error and log strings, and the sanitized DNS-03 example and host-facing text contain none of these prohibited claims: that the test configured DoH for the device or host application; that a DoH result is authoritative or correct; that a difference from system DNS establishes interference; that a bootstrap, TCP, TLS, HTTP or DNS failure is DoH blocking or censorship; and that answers carry on-device ASN, organization, geolocation, reputation or censorship-verdict enrichment. Verify each failure is attributed only to the directly observed stage via observation order and the terminal observation. Requirements: REQ-DNS-03-0001, REQ-DNS-03-0002, REQ-DNS-03-0003, REQ-DNS-03-0030, REQ-DNS-03-0031, REQ-DNS-03-0032. Judgement-based completion is the separate review case RV-M-DNS-03-INTERPRETATION; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-DNS-03-0001: No non-negated sentence in the comments and string literals of production Go files and sanitized examples and the shipped Markdown documentation (case plans, manifests, evidence, changelog and specification-issue records excluded) matches the "DNS-03 configures DoH for the device or host" claim pattern, and the DNS-03 section of the interpretation guidance contains "does not configure doh for the device or the host application"; the test does not inspect device or host configuration. This is a textual scan of shipped text; it does not prove how a reader interprets a result.
  - REQ-DNS-03-0002: The same scan finds no non-negated sentence stating that a DoH or DNS-03 answer, result or response is authoritative or correct, the guidance contains "a doh answer is not authoritative or correct", no run's measurement contains `authoritative` or `correct`, and the strict decoder rejects an injected `authoritative` field.
  - REQ-DNS-03-0003: The same scan finds no non-negated sentence stating that a difference from system DNS establishes, indicates, shows, proves or means interference, the guidance contains "a difference from system dns does not establish interference", and no run's measurement contains `interfer`; no run compares with system DNS.
  - REQ-DNS-03-0030: The guidance contains "establishes whether each observed bootstrap, tcp, tls, http and dns stage completed for the configured resolver and query", all fifteen runs carry registered outcomes or none when interrupted, and for each of the twelve failure outcomes the terminal observation has the expected stage type with every earlier observation succeeded.
  - REQ-DNS-03-0031: Not asserted: the test requires none of the listed evidence items in the guidance and checks none of them.
  - REQ-DNS-03-0032: Each failure's terminal observation is its stage (`bootstrap_failure` and the DNS outcomes `dns_query`, `tcp_failure` and `timeout` `tcp_connect`, `tls_failure` `tls_handshake`, `http_failure` and `unexpected_content_type` `http_exchange`) after only succeeded observations, the same scan finds no non-negated sentence stating that a stage failure is or means DoH blocking, blocking or censorship, and the guidance contains "is not summarized as doh blocking or censorship" and "attributed only to the directly observed stage".

<a id="M-DNS-03-SAFETY"></a>

### M-DNS-03-SAFETY

- Case: `M-DNS-03-SAFETY`
- Result: passed
- Tests: `TestMDns03Safety` in `measurement_safety_conformance_test.go` (package `root`); `TestMDns03Minimization` in `dns03_conformance_test.go` (package `root`); `TestMDns03Limits` in `dns03_conformance_test.go` (package `root`); `TestMDns03TupleConcurrency` in `dns03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns03Safety$' ./.`
- Scope: Using synthetic locally registered resolver endpoints and query names, a fake local policy and host, a deterministic clock and recording TCP/TLS/HTTP transports, iterate the Baseline Risk rows: assert the registered baseline risk level is medium and requires_user_consent is false, and verify a task declaring low is rejected or evaluated as at least medium, a task declaring high yields effective risk high, a task requiring consent without valid host consent is not executed, and host or local policy requiring consent is honored despite the false baseline. Iterate each Risk Factors row (network_observability medium, traffic_distinctiveness medium, target_sensitivity medium, data_sensitivity low, legal_or_social_context medium): assert the task overall risk is at least as restrictive as every factor, an unknown factor yields unknown unless resolved by trusted local assessment, and no factor lowers the baseline. Exercise each normative action: data minimization, asserting no response body, full headers, resolver cookies or address lists are retained; traffic distinctiveness, asserting the recording transport sees exactly one ordinary RFC 8484 POST with application/dns-message and no unusual headers, padding variation or synchronized repetition under the deterministic clock; user-identification risk, asserting the recorded request carries no cookies, authentication, installation identifiers or application headers; target risk, asserting an unregistered or unreviewed resolver endpoint or query name is rejected with zero recorded transport contacts; and probe authority, asserting local policy can delay, skip or reject execution (execution.reason local_policy_rejected per the public execution table) with no network contact. Requirements: REQ-DNS-03-0033, REQ-DNS-03-0034, REQ-DNS-03-0035, REQ-DNS-03-0036, REQ-DNS-03-0037, REQ-DNS-03-0038, REQ-DNS-03-0039. Judgement-based completion is the separate review case RV-M-DNS-03-SAFETY; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-DNS-03-0033: The test asserts that the production registry records a `medium` baseline for DNS-03 and that a task declaring `low` or `medium` has effective risk exactly `medium`, `high` exactly `high` and `unknown` exactly `unknown` (admitted under issuer and local bounds permitting that level, rejected `task_issuer_not_authorized` with no contact when either side permits only lower levels). For the consent baseline it asserts that `false` adds no requirement (with host consent absent the run starts, yields one result and never queries the consent capability), while consent required by the task, by an issuer or by the local policy blocks every result and contact unless the fake host reports valid consent (absent, ended and unavailable all block).
  - REQ-DNS-03-0034: For each of the five factor types the test asserts that the registered baseline equals the most restrictive of the fixture's row levels (`data_sensitivity` `low`, the other four `medium`), that declaring the factor at `low`, `medium` or `high` gives effective risk exactly the maximum of baseline, row level and declared level, that an `unknown` factor gives `unknown`, and that an unknown factor under a known overall level or a factor above the overall level is rejected `task_schema_invalid` and an unrecognized factor type `task_unsupported`, all with no contact. The row levels are fixture constants checked only against the registered baseline, and the rationale text is not asserted.
  - REQ-DNS-03-0035: The test asserts that no surface contains the resolver URL path and host, resolved and bootstrap addresses, query name and identifier, DNS authority and additional records, raw message bytes, TTL, response header name and value, `Set-Cookie` value, `Content-Type` value, body bytes beyond the DNS message, status text, certificate fields or raw errors, and that the measurement holds only the closed field set (DNS answer fields plus TLS, HTTP status, protocol, body counters and content encoding), with answers bounded, deduplicated and ordered. The decoder rejects injected resolver URL, headers, content type, cookie, certificate and error fields.
  - REQ-DNS-03-0036: The test asserts one connection and one request per run, the request being a POST to `/dns-query` with `Content-Type: application/dns-message`, a fixed 31-byte body and none of the listed identifying headers, and that under TLS and HTTP stalls, attempt deadlines and size bounds each run makes exactly one connection and, once a request is sent, one request (no retry). Synchronized repetition is covered only by the tuple serialization in the repeated-execution row; cadence is not asserted.
  - REQ-DNS-03-0037: The test asserts that requests carry none of `Cookie`, `Authorization`, `Proxy-Authorization`, `If-None-Match`, `If-Modified-Since`, `Referer`, `Origin`, `X-Requested-With`, `X-Client-Id` or `X-Device-Id` and no header value containing `safety-` or `barometer`, and that a server requesting a client certificate receives none.
  - REQ-DNS-03-0038: The test asserts that an unreviewed query name, resolver host, resolver path or bootstrap address (public or loopback) is rejected at admission with zero dials and lookups, and that loopback, private, link-local and IPv6 loopback bootstrap addresses are refused by the codec or policy or never contacted. For repeated execution it asserts that a second run on the same resolver tuple makes no connection while the first is in its handshake and can be cancelled `not_run`; the residual-risk statement is not asserted.
  - REQ-DNS-03-0039: The test asserts that a local policy not permitting the task's network path rejects a valid task at admission with no contact, that a local risk bound below `medium` rejects it, and that a local charging-only device constraint holds an admitted valid task (`run_not_due` twice, no result, no contact) until satisfied, after which the same task runs; a target registration that expires after admission ends the attempt `local_policy_rejected` with no contact. The no-lowering clause is asserted through the declared-level rows: no task declaration yields effective risk below the `medium` baseline.

<a id="M-DNS-03-OUTPUT"></a>

### M-DNS-03-OUTPUT

- Case: `M-DNS-03-OUTPUT`
- Result: passed
- Tests: `TestMDns03ExecutionAndOutput` in `dns03_conformance_test.go` (package `root`); `TestMDns03Normalization` in `dns03_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns03ExecutionAndOutput$' ./.`
- Scope: Verify bootstrap DNS only when actually used, then TCP/TLS/HTTP/doh DNS in execution order, stopping at the failed stage. Use shared DNS normalization and HTTP/TLS observations without importing HTTP-01 GET semantics or synthesizing unstarted stages.
- Requirement aspects exercised:

  - REQ-DNS-03-0025: Results contain the system dns_query only when bootstrap resolution ran, then tcp_connect, tls_handshake, http_exchange and the doh dns_query in execution order, stopping at the failed stage with no unstarted stage and no doh dns_query after a non-2xx or wrong-media-type response; only registered detail keys appear.
  - REQ-DNS-03-0026: With bootstrap_ip the first observation is omitted; the doh dns_query carries engine doh, the address family, query type and one response with rcode, arrival offset equal to the exchange duration and normalized answers without TTL.

<a id="M-DNS-03-VERSIONS"></a>

### M-DNS-03-VERSIONS

- Case: `M-DNS-03-VERSIONS`
- Result: passed
- Test: `TestMDns03Versions` in `dns03_conformance_test.go` (root package, Go package `barometer`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns03Versions$' ./.`
- Scope: Select test.id DNS-03 with test.version 0.1.0, Measurement Result schema_version 0.1.0 and test.result_version 0.1.0 and assert that exact registered combination validates and is the only combination emitted. Iterate rejections before any transport contact on a recording transport: unsupported test.version values (0.1.1, 0.2.0, 1.0.0), malformed versions (empty, 0.1, v0.1.0, 0.1.0-draft, non-string, whitespace-padded), unsupported schema_version and unsupported test.result_version values in the same forms, a mismatched pair where only one of test.version and result_version is 0.1.0, and cross-test combinations such as test.id DNS-03 paired with a DNS-01 or DNS-04 Output object, or test.id DNS-01 carrying a DNS-03 Output object. Assert each rejected task is a task-validation failure reported outside Measurement Results, produces no Measurement Result and no measurement object, and is never coerced, guessed, interpreted best-effort or mapped to another version. Verify a received result with an unsupported version is not parsed as the known 0.1.0 format and that unknown fields within 0.1.0 are rejected. Assert the Draft status and registration metadata alone do not cause any additional version to be accepted. Requirements: REQ-DNS-03-0024.
- Requirement aspects exercised:

  - REQ-DNS-03-0024: Task test_version 0.1.1, 0.2.0 and 1.0.0 are task_unsupported and empty, v-prefixed, partial, suffixed, padded, non-string, null and missing values are task_schema_invalid, all with no lookup, dial, replay commit or result; the emitted 0.1.0 result decodes, while schema_version, test.version or test.result_version set to unsupported or malformed values, DNS-04 and DNS-01 identities or bindings, a bootstrap binding without the system observation, foreign outcomes, engines and query types, extra fields, a missing HTTP stage and reordered stages are rejected.

<a id="M-DNS-04-ADMISSION"></a>

### M-DNS-04-ADMISSION

- Case: `M-DNS-04-ADMISSION`
- Result: passed
- Tests: `TestMDns04Admission` in `dns04_conformance_test.go` (package `root`); `TestMDns04NormalizedAdmission` in `dns04_closeout_test.go` (package `root`); `TestMDns04ExecutionAndOutput` in `dns04_conformance_test.go` (package `root`); `TestMDns04Normalization` in `dns04_conformance_test.go` (package `root`); `TestMDns04ExchangeFaults` in `dns04_closeout_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns04Admission$' ./.`
- Scope: Vary registered resolver_name/port/bootstrap tuple, normalized 253-byte resolver and query names, A/AAAA and canonical public bootstrap syntax. Require reviewed support for nonstandard ports, reject unknown/local/private/interpolated targets, and assert TLS always authenticates resolver_name even with bootstrap_ip; no cleartext DNS is sent on the DoT connection.
- Requirement aspects exercised:

  - REQ-DNS-04-0004: Through Client admission against the production target rule, a registered resolver profile on port 853 and on a non-standard port 8853 registered in the matching profile is admitted with its query name, while port 53 or 8853 against a profile registered on 853, another resolver name and another query name are target_not_registered with no lookup, dial or replay commit.
  - REQ-DNS-04-0009: Admission accepts A and AAAA, an IPv4 and an IPv6 bootstrap address, a case-folded resolver and query name and a Unicode query name matched against canonical registrations (TestMDns04NormalizedAdmission), and 253-byte names, and rejects as task_schema_invalid each missing required field, a non-integer, zero or 65536 port, other record types and 254-byte, IP-literal, single-label, trailing-dot, wildcard, whitespace-padded and empty names.
  - REQ-DNS-04-0010: A task whose port, bootstrap address, resolver name or query name differs from the registration, or that adds or omits the bootstrap address relative to the registered profile, is target_not_registered before any lookup or contact; case-folded and Unicode spellings match the canonical registration.
  - REQ-DNS-04-0011: With a bootstrap address on the direct and proxy paths the server sees one ClientHello whose server name is the resolver name (TestMDns04ExecutionAndOutput), and a certificate for another name with a bootstrap address is tls_failure with certificate_name_mismatch and name_mismatch (TestMDns04Normalization).
  - REQ-DNS-04-0012: The server receives only the 33-byte framed query as TLS plaintext after a completed handshake and nothing after it, a TLS failure delivers no plaintext, and faulted exchanges make one dial with no second connection or other destination (TestMDns04ExecutionAndOutput, TestMDns04ExchangeFaults).
  - REQ-DNS-04-0013: Local, loopback, link-local, private, shared, documentation, unspecified, IPv4-mapped, zoned and non-canonical bootstrap addresses and local, special-use and interpolated names are task_schema_invalid, and unregistered targets are target_not_registered, all without lookup or contact.
  - REQ-DNS-04-0014: Unknown input fields, including transport and cleartext_fallback, and non-object inputs are task_schema_invalid.
  - REQ-DNS-04-0015: Every invalid input returns task_schema_invalid with no handle, task handle, replay commit, lookup, dial or output record.

<a id="M-DNS-04-OPTIONS"></a>

### M-DNS-04-OPTIONS

- Case: `M-DNS-04-OPTIONS`
- Result: passed
- Test: `TestMDns04Options` in `dns04_conformance_test.go` (root package, Go package `barometer`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns04Options$' ./.`
- Scope: Exercise auto/ipv4/ipv6 with bootstrap compatibility, omitted default and invalid/unknown cases. Assert no opportunistic authentication, cleartext fallback, cross-run connection reuse or alternate resolver fallback.
- Requirement aspects exercised:

  - REQ-DNS-04-0016: Omitted, auto, ipv4 and ipv6 address families compatible with the bootstrap address, and ipv4 or ipv6 with system bootstrap, are admitted without contact; ipv6 with an IPv4 bootstrap address, ipv4 with an IPv6 one, and null, empty, mistyped, mis-cased, numeric and unknown values or option objects are task_schema_invalid.
  - REQ-DNS-04-0017: Options requesting opportunistic authentication, cleartext fallback, a fallback resolver, connection reuse, ALPN, skipped verification or retries, and any unknown option, are task_schema_invalid with no handle, replay commit or contact.

<a id="M-DNS-04-PRECONDITIONS"></a>

### M-DNS-04-PRECONDITIONS

- Case: `M-DNS-04-PRECONDITIONS`
- Result: passed
- Tests: `TestMDns04Preconditions` in `dns04_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns04Preconditions$' ./.`
- Scope: Vary technical/privacy/safety review, local registered profile, explicit host/anomaly policy, stable generation and deadline. No resolver contact before all applicable gates pass; native/deployment policy still needs scoped review. Portable gate tests plus separately required native/host validation; no platform claim.
- Requirement aspects exercised:

  - REQ-DNS-04-0018: Tasks whose query name, resolver host or port differs from the local registration are rejected at admission with zero transport dials and system-resolver lookups. The technical, privacy and safety review itself is not asserted; host row, see docs/measurement-requirements.md.
  - REQ-DNS-04-0019: Asserted only indirectly: a resolver profile that differs from the local registration is rejected at admission, and every gate row records zero dials and lookups. No row checks for a configuration download as such.
  - REQ-DNS-04-0020: A deployment that lists only another test and no registrations rejects the task with `task_issuer_not_authorized` and no contact. No anomaly-triggered path is asserted; the deployment review is a host row, see docs/measurement-requirements.md.
  - REQ-DNS-04-0021: Through the shared gates, an unreadable network access state at reservation defers the run (`run_not_due`, no result), and a network change after reservation to a transport the task does not allow gives `not_run` with `constraints_unsatisfied`, both with zero contacts. Execution deadline availability is not asserted.
  - REQ-DNS-04-0022: The shared portable gates assert, each with zero contacts: a task transport condition the network does not meet, a task excluding VPN on a VPN-detected network and a local charging-only device restriction defer the run at `StartRun` (`run_not_due`, no result); a transport condition lost after reservation gives `not_run` with `constraints_unsatisfied`; and a proxy-path task with no proxy configured is either rejected at admission with `task_schema_invalid` or ends `not_run` with `proxy_not_configured` (the test accepts either); a deployment that withholds the target registration rejects the task with `target_not_registered`, and a registration that lapses after admission gives a result with reason `local_policy_rejected`; contacts count both transport dials and system-resolver lookups.

<a id="M-DNS-04-EXECUTION"></a>

### M-DNS-04-EXECUTION

- Case: `M-DNS-04-EXECUTION`
- Result: passed
- Tests: `TestMDns04ExecutionAndOutput` in `dns04_conformance_test.go` (package `root`); `TestMDns04ExchangeFaults` in `dns04_closeout_test.go` (package `root`); `TestMDns04Normalization` in `dns04_conformance_test.go` (package `root`); `TestDns04PlaintextAccounting` in `internal/execution/dns04_termination_test.go` (package `internal/execution`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns04ExecutionAndOutput$' ./.`
- Scope: Record optional system bootstrap, one selected TCP connection, authenticated TLS, one length-prefixed DNS query and bounded matching response, then close. Fault the prefix, message length, query match and each stage; require no cleartext or resolver fallback and no application retry.
- Requirement aspects exercised:

  - REQ-DNS-04-0005: Through Client with the production registry on the direct path with and without system bootstrap and on the proxy path, the attempt makes at most one lookup, one dial or one CONNECT to the bootstrap address, one handshake naming the resolver with no ALPN, sends one 33-byte framed query and receives one response; under prefix, message, mismatch and TLS faults it still makes one dial with no second connection or other destination (TestMDns04ExchangeFaults); the meter counts one connection (TestDns04PlaintextAccounting).
  - REQ-DNS-04-0006: System bootstrap makes one IPv4 lookup of the resolver name and a bootstrap address makes none; one connection follows, then TLS, then one framed query and response, and the server sees the connection closed; a failed bootstrap makes no dial (TestMDns04Normalization) and each faulted exchange ends without a further write, lookup or connection (TestMDns04ExchangeFaults).

<a id="M-DNS-04-NORMALIZATION"></a>

### M-DNS-04-NORMALIZATION

- Case: `M-DNS-04-NORMALIZATION`
- Result: passed
- Tests: `TestMDns04Normalization` in `dns04_conformance_test.go` (package `root`); `TestMDns04Limits` in `dns04_conformance_test.go` (package `root`); `TestMDns04ExchangeFaults` in `dns04_closeout_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns04Normalization$' ./.`
- Scope: Exercise every outcome row, including TCP reset before TLS versus reset during TLS, authenticated negative DNS codes, malformed prefix/message, mismatch, post-connect timeout and inconclusive. Keep terminal failure at the observed stage with prior observations ordered.
- Requirement aspects exercised:

  - REQ-DNS-04-0007: A matching NOERROR response with one answer and with no answers is response with rcode NOERROR, and authenticated NXDOMAIN, SERVFAIL and REFUSED responses keep their rcode in a completed result.
  - REQ-DNS-04-0008: Loopback behavior and wrapped platform errors yield bootstrap_failure (timeout, resolver_error), tcp_timeout, tcp_refused, tcp_reset before and during TLS, tls_failure for untrusted, name-mismatched, alerting and garbled servers, dns_nxdomain, dns_servfail, dns_refused, malformed_dns_response for a short message and for a stream closed inside the prefix, inside the message or after the query or a prefix beyond the stream (TestMDns04ExchangeFaults), query_mismatch, and post-connect timeout at the TLS and DNS stages (TestMDns04Limits).
  - REQ-DNS-04-0027: Each outcome's terminal observation carries the registered failure (timeout, connection_refused, connection_reset, certificate_untrusted with untrusted, certificate_name_mismatch with name_mismatch, alert, protocol_error, nxdomain, servfail, refused, malformed_response, query_mismatch, inconclusive) after succeeded prior stages in order; an unclassified TCP failure, an unregistered NOTIMP code and a reset after the query (TestMDns04ExchangeFaults) are inconclusive at their stage; every result passes bundle validation.

<a id="M-DNS-04-LIMITS"></a>

### M-DNS-04-LIMITS

- Case: `M-DNS-04-LIMITS`
- Result: passed
- Tests: `TestMDns04Limits` in `dns04_conformance_test.go` (package `root`); `TestMDns04TupleConcurrency` in `dns04_conformance_test.go` (package `root`); `TestDns04PlaintextAccounting` in `internal/execution/dns04_termination_test.go` (package `internal/execution`); `TestMDns04ExecutionAndOutput` in `dns04_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns04Limits$' ./.`
- Scope: Test the 65,535-byte DNS message and 16-answer limit at below/equal/above boundaries under stricter common limits. Assert one resolver connection/DNS exchange after optional bootstrap, bounded chain deadline, no retries and deduplication by resolver/query tuple.
- Requirement aspects exercised:

  - REQ-DNS-04-0023: The 15-second TLS and DNS operation bounds give completed timeout while a 60-second deadline is open and a 1,000 ms deadline gives incomplete attempt_deadline_exceeded, each with one dial and the session released; 65,534- and 65,535-byte messages are accepted, a received limit one byte below the framed 65,535-byte message is incomplete resource_limit, 16 answers are retained and 17 are malformed_dns_response; max_connections 1 with a sent limit equal to the framed query succeeds with one connection; the meter counts the message and its two-byte prefix in both directions on the direct and proxy paths, and a sent limit one byte below the framed query is resource_limit before any plaintext leaves (TestDns04PlaintextAccounting); system bootstrap makes exactly one lookup (TestMDns04ExecutionAndOutput); an equivalent resolver and query tuple waits without contact and is not_run when cancelled (TestMDns04TupleConcurrency).

<a id="M-DNS-04-TERMINATION"></a>

### M-DNS-04-TERMINATION

- Case: `M-DNS-04-TERMINATION`
- Result: passed
- Tests: `TestMDns04Termination` in `dns04_conformance_test.go` (package `root`); `TestDns04TerminationPoints` in `internal/execution/dns04_termination_test.go` (package `internal/execution`); `TestMDns04Limits` in `dns04_conformance_test.go` (package `root`); `TestMDns04ExchangeFaults` in `dns04_closeout_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns04Termination$' ./.`
- Scope: Drive DNS-04 with a synthetic registered resolver profile, a scripted system resolver, a scripted TCP/TLS/DoT transport, a fake platform and a deterministic clock, and inject each interruption at each Procedure point: before input validation completes, during system bootstrap resolution of the resolver name (omitted when an approved bootstrap_ip is supplied), during the TCP connect to the one selected address on the configured port, during TLS establishment and resolver-name certificate validation, after the length-prefixed DNS query is written but before the bounded response is fully read, and while closing the session. Expected states follow the GO-SPEC-ISSUE-004 disposition: before input validation completes the attempt is not_run with that reason, no measurement and no network_context; while closing the session after the outcome is recorded it stays completed with the recorded outcome; every other point is incomplete as asserted below. At each point inject in turn a network generation change, host cancellation, expiry of the outer attempt deadline, platform suspension, and, only where a proxy path is configured, proxy loss and host proxy reconfiguration. At the incomplete points, assert execution.state incomplete with exactly the first applicable reason from the common table (network_changed, cancelled_by_host, attempt_deadline_exceeded, platform_suspended, proxy_unavailable, proxy_connect_failed or proxy_reconfigured), no measurement object, no outcome and no observations, and that no failure value such as timeout, tcp_timeout or network_changed is emitted as a measurement outcome for these interruptions. Verify the contrasting controls: a per-stage operation bound reached while the outer deadline remains open yields execution.state completed with outcome tcp_timeout or timeout, and an oversized length prefix yields completed malformed_dns_response rather than an incomplete state. Assert the recording transport shows the socket closed, no further write after the interruption, no cleartext DNS, no alternate resolver, no second connection and no application-layer retry, and that no partial NET-05 style measurement is retained. Requirements: REQ-DNS-04-0028.
- Requirement aspects exercised:

  - REQ-DNS-04-0028: For host cancellation, network_changed, the attempt deadline, platform suspension and the three proxy reasons injected before start, during bootstrap, TCP connect, TLS and after the framed query is written, the attempt is not_run before start (no measurement or network context) and otherwise incomplete with exactly that reason, no measurement, the expected dial and lookup counts, every connection closed and no second connection; an interruption while closing after the outcome is recorded stays completed response (TestDns04TerminationPoints); through Client, a generation change or cancellation after the query and a refused proxy tunnel are incomplete with no further plaintext and no direct or resolver fallback; operation bounds with an open deadline (TestMDns04Limits) and a prefix beyond the stream (TestMDns04ExchangeFaults) stay completed outcomes.

<a id="M-DNS-04-MINIMIZATION"></a>

### M-DNS-04-MINIMIZATION

- Case: `M-DNS-04-MINIMIZATION`
- Result: passed
- Tests: `TestMDns04Minimization` in `dns04_conformance_test.go` (package `root`), `TestMDns04Safety` in `measurement_safety_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns04Minimization$' ./.`
- Scope: Exercise DNS-04 against a scripted DoT transport using one unique synthetic canary per prohibited item: a canary resolver_name, a canary resolver address returned by the scripted bootstrap, a canary non-standard port value accepted by a synthetic reviewed profile, a canary bootstrap_ip, a canary query hostname distinct from every scripted answer value, a canary marker embedded in the raw DNS response message bytes (query identifier, flags, authority and additional sections), canary TLS certificate data (subject, issuer, serial, fingerprint), canary query framing bytes in the two-byte length prefix, and a canary raw error string from each scripted stage failure. Iterate success, bootstrap_failure, tcp_refused, tls_failure, dns_refused, malformed_dns_response and query_mismatch runs, and assert serialized output of the full Measurement Result, every nested observation and every error path contains none of the canaries. Assert the measurement contains only outcome and ordered observations whose fields are limited to type, status, duration_ms, failure and details, with details limited to address_family for tcp_connect; address_family, tls_version, certificate_validation and an optional alert for tls_handshake; and engine (system or dot), address_family, query_type and responses with arrival_index, arrival_offset_ms, rcode and answers of type and value for dns_query, with at most 16 normalized answers and no TTL. Reject any extra field. Requirements: REQ-DNS-04-0029.
- Requirement aspects exercised:

  - REQ-DNS-04-0029: Unique canaries cover `resolver_name`, the resolver address, the non-standard port, `bootstrap_ip`, the query hostname and identifier, DNS authority, additional and raw-message patterns, the answer TTL, the certificate subject, issuer and serials and raw stage, x509 and TLS errors, plus hexadecimal and base64 forms of each length-prefixed DNS query and response the scripted server handled. Over `response` with and without `bootstrap_ip`, `bootstrap_failure`, `tcp_timeout`, `tcp_refused`, `tcp_reset`, `tls_failure`, `dns_nxdomain`, `dns_servfail`, `dns_refused`, `malformed_dns_response`, `query_mismatch` and `inconclusive`, no canary appears in any result, the bundle, the pending store's bytes, any returned error or the task-validation rejection of a task whose unknown input member carries the canaries, every measurement path is within the registered set, the `bootstrap_ip` response keeps two answers starting with `9.9.9.9`, and the strict decoder rejects 12 injected members (including `resolver_name`, `port`, `ttl`, `message`, `length_prefix`, `certificate` and `error`) in the last two observations' details and in an answer. No incomplete or `timeout` path is run, and certificate DER bytes and fingerprints are not searched for.

<a id="M-DNS-04-INTERPRETATION"></a>

### M-DNS-04-INTERPRETATION

- Case: `M-DNS-04-INTERPRETATION`
- Result: passed
- Tests: `TestMDns04Interpretation` in `dns04_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns04Interpretation$' ./.`
- Scope: Assert serialized DNS-04 results for every registered outcome (response, bootstrap_failure, tcp_timeout, tcp_refused, tcp_reset, tls_failure, dns_nxdomain, dns_servfail, dns_refused, malformed_dns_response, query_mismatch, timeout, inconclusive) contain no interpretation, verdict, censorship, blocking, anomaly or intent field and no field outside the registered Output schema. Assert Go field documentation, outcome descriptions and the sanitized example and host-facing text contain none of these prohibited claims: that the test configures DoT for the device or host application; that a response answer is authoritative or correct; that a failure on the standard DoT port 853 proves intentional blocking; that a transport difference between DNS-04 and DNS-01 or DNS-03 establishes censor intent; that a single execution establishes more than whether each observed bootstrap, TCP, TLS and DNS stage completed; or that dns_refused, tls_failure, tcp_reset or tcp_timeout by itself excludes resolver, authoritative DNS, certificate or routing outage, bootstrap failure, resolver rate limiting or policy refusal, port-specific firewall behavior, VPN, captive portal, enterprise policy or local security software, or IPv4/IPv6 path differences. Requirements: REQ-DNS-04-0001, REQ-DNS-04-0002, REQ-DNS-04-0003, REQ-DNS-04-0030, REQ-DNS-04-0031, REQ-DNS-04-0032. Judgement-based completion is the separate review case RV-M-DNS-04-INTERPRETATION; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-DNS-04-0001: No non-negated sentence in the comments and string literals of production Go files and sanitized examples and the shipped Markdown documentation (case plans, manifests, evidence, changelog and specification-issue records excluded) matches the "DNS-04 configures DoT for the device or host" claim pattern, and the DNS-04 section of the interpretation guidance contains "does not configure dot for the device or the host application"; the test does not inspect device or host configuration. This is a textual scan of shipped text; it does not prove how a reader interprets a result.
  - REQ-DNS-04-0002: The same scan finds no non-negated sentence stating that a DoT or DNS-04 answer, result or response is authoritative or correct, the guidance contains "a response answer is not authoritative or correct", no run's measurement contains `authoritative` or `correct`, and the strict decoder rejects an injected `authoritative` field.
  - REQ-DNS-04-0003: The same scan finds no non-negated sentence stating that a port 853 or standard DoT port failure establishes, proves, shows or means intentional blocking, the guidance contains "does not establish intentional blocking", and no member name or string value of any run's measurement contains a verdict token (such as `verdict`, `censor`, `block`, `filter`, `anomal`, `confiden`) and the strict result decoder rejects `verdict`, `interpretation`, `censored`, `blocked`, `filtered`, `anomaly`, `confidence`, `reachable` and `authoritative` fields injected into a valid result; the runs use a non-standard registered port, so no standard-port failure is executed.
  - REQ-DNS-04-0030: The guidance must contain only the fragment "and nothing more" of its single-execution statement, and the thirteen runs carry registered outcomes; unlike DNS-03, no stage attribution is checked.
  - REQ-DNS-04-0031: Not asserted: the test requires none of the listed evidence items in the guidance and checks none of them.
  - REQ-DNS-04-0032: The same scan finds no non-negated sentence stating that a transport difference establishes, proves, shows or means intent or censorship, and the guidance contains "not a conclusion about censor intent"; no cross-transport comparison is run.

<a id="M-DNS-04-SAFETY"></a>

### M-DNS-04-SAFETY

- Case: `M-DNS-04-SAFETY`
- Result: passed
- Tests: `TestMDns04Safety` in `measurement_safety_conformance_test.go` (package `root`); `TestMDns04Minimization` in `dns04_conformance_test.go` (package `root`); `TestMDns04Limits` in `dns04_conformance_test.go` (package `root`); `TestMDns04TupleConcurrency` in `dns04_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns04Safety$' ./.`
- Scope: Exercise DNS-04 with synthetic registered resolver profiles and query names, a recording TCP/TLS/DoT transport and a fake host policy. Iterate the Baseline Risk rows: risk level medium, where a task declaring low is rejected or evaluated as at least medium and tasks declaring medium, high or unknown keep effective risk at least medium; and requires user consent false subject to host and local policy, where a fake host or local policy requiring consent without valid consent blocks execution. Iterate each Risk Factors row (network_observability medium, traffic_distinctiveness medium, target_sensitivity medium, data_sensitivity low, legal_or_social_context medium) and assert a task factor declared below its registered level does not lower effective risk. Exercise each Safety Analysis action: data minimization, asserting only stage outcomes, coarse quantized timing and minimal DNS properties are retained; traffic distinctiveness, asserting one standard exchange with no port-range probing, TLS variant probing or fallback; user-identification risk, asserting the recorded bytes contain no identifier, authentication or unrelated application data canary; target risk, rejecting unregistered or unreviewed resolver endpoints, query names, non-standard ports without review, and local, loopback, link-local, private, interpolated or unknown targets before any traffic; and probe authority, verifying local policy can delay, skip or reject execution with not_run local_policy_rejected and no traffic. Requirements: REQ-DNS-04-0033, REQ-DNS-04-0034, REQ-DNS-04-0035, REQ-DNS-04-0036, REQ-DNS-04-0037, REQ-DNS-04-0038, REQ-DNS-04-0039. Judgement-based completion is the separate review case RV-M-DNS-04-SAFETY; this portable case cannot complete the family alone.
- Requirement aspects exercised:

  - REQ-DNS-04-0033: The test asserts that the production registry records a `medium` baseline for DNS-04 and that a task declaring `low` or `medium` has effective risk exactly `medium`, `high` exactly `high` and `unknown` exactly `unknown` (admitted under issuer and local bounds permitting that level, rejected `task_issuer_not_authorized` with no contact when either side permits only lower levels). For the consent baseline it asserts that `false` adds no requirement (with host consent absent the run starts, yields one result and never queries the consent capability), while consent required by the task, by an issuer or by the local policy blocks every result and contact unless the fake host reports valid consent (absent, ended and unavailable all block).
  - REQ-DNS-04-0034: For each of the five factor types the test asserts that the registered baseline equals the most restrictive of the fixture's row levels (`data_sensitivity` `low`, the other four `medium`), that declaring the factor at `low`, `medium` or `high` gives effective risk exactly the maximum of baseline, row level and declared level, that an `unknown` factor gives `unknown`, and that an unknown factor under a known overall level or a factor above the overall level is rejected `task_schema_invalid` and an unrecognized factor type `task_unsupported`, all with no contact. The row levels are fixture constants checked only against the registered baseline, and the rationale text is not asserted.
  - REQ-DNS-04-0035: The test asserts that no surface contains the resolver name and address, a non-standard port, the bootstrap address, query name and identifier, DNS authority and additional records, raw message bytes, TTL, certificate fields or raw stage errors, and that the measurement holds only stage outcomes, observation type, status, `duration_ms`, failure and minimal DNS and TLS properties, with answers bounded and deduplicated. The decoder rejects injected resolver name, address, port, query identifier, TTL, message, length prefix, certificate, ALPN and error fields; quantization of the timing value is not asserted.
  - REQ-DNS-04-0036: The test asserts one dial, one accepted connection and one ClientHello carrying the registered resolver name, and that under TLS and DNS stalls, attempt deadlines and size bounds each run makes exactly one connection (no retry or fallback). Port-range probing is covered by the target row rejecting a non-standard port; TLS variants are not asserted beyond the single ClientHello.
  - REQ-DNS-04-0037: The test asserts that a server requesting a client certificate receives none and reruns the closed-result minimization row; in the traffic row a run limited to `max_bytes_sent` of 33 (length prefix plus the 31-byte query) still completes. The sent bytes are not scanned for identifier or application-data canaries.
  - REQ-DNS-04-0038: The test asserts that an unreviewed query name, resolver name, non-standard port (8853) or bootstrap address (public or loopback) is rejected at admission with zero dials and lookups, and that loopback, private, link-local, documentation and IPv6 loopback bootstrap addresses are refused by the codec or policy or never contacted. For repeated execution it asserts that a second run on the same resolver tuple makes no connection while the first is in its handshake; the residual-risk statement is not asserted.
  - REQ-DNS-04-0039: The test asserts that a local policy not permitting the task's network path rejects a valid task at admission with no contact, that a local risk bound below `medium` rejects it, and that a local charging-only device constraint holds an admitted valid task (`run_not_due` twice, no result, no contact) until satisfied, after which the same task runs; a target registration that expires after admission ends the attempt `local_policy_rejected` with no contact. The no-lowering clause is asserted through the declared-level rows: no task declaration yields effective risk below the `medium` baseline.

<a id="M-DNS-04-OUTPUT"></a>

### M-DNS-04-OUTPUT

- Case: `M-DNS-04-OUTPUT`
- Result: passed
- Tests: `TestMDns04ExecutionAndOutput` in `dns04_conformance_test.go` (package `root`); `TestMDns04Normalization` in `dns04_conformance_test.go` (package `root`); `TestMDns04Limits` in `dns04_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns04ExecutionAndOutput$' ./.`
- Scope: Verify optional bootstrap DNS, TCP, TLS, then dot DNS; omit bootstrap when an approved bootstrap_ip is supplied and omit unstarted stages after failure. Reuse the common DNS answer normalization and bounds without retaining framing bytes.
- Requirement aspects exercised:

  - REQ-DNS-04-0025: Results carry a system dns_query (engine system, no responses) only with system bootstrap, then tcp_connect, tls_handshake and a dot dns_query, in that order, with only the stages that began after a failure and 250-ms quantized durations.
  - REQ-DNS-04-0026: With a bootstrap address the first observation is the tcp_connect; the dot dns_query keeps query_type A and one response with arrival_index, arrival_offset_ms, rcode and answers only, at most 16 answers (TestMDns04Limits), no ttl, and no observation detail outside the registered sets; results pass bundle validation.

<a id="M-DNS-04-VERSIONS"></a>

### M-DNS-04-VERSIONS

- Case: `M-DNS-04-VERSIONS`
- Result: passed
- Tests: `TestMDns04Versions` in `dns04_conformance_test.go` (package `root`); `TestMDns04CrossTestOutputs` in `dns04_closeout_test.go` (package `root`); `TestProductionRegistryContainsOnlyDeliveredMeasurements` in `registry_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestMDns04Versions$' ./.`
- Scope: Exercise task acceptance and result emission for DNS-04 with the exact registered combination test_version 0.1.0 (test.version 0.1.0), Measurement Result schema_version 0.1.0 and test.result_version 0.1.0, and assert the emitted result carries exactly test.id DNS-04 with that one combination. Iterate unsupported values (0.1.1, 0.2.0, 1.0.0), malformed values (empty string, 0.1, v0.1.0, 0.1.0 with trailing whitespace, numeric 0.1, null) for test_version, schema_version and result_version independently, and mixed combinations where only one of the three is unsupported. Reject each before execution as a task-validation failure, assert the recording transport saw no bootstrap, TCP, TLS or DNS activity and that no Measurement Result is produced. Inject cross-test combinations: a DNS-04 measurement object validated under test.id DNS-01 or DNS-03, and a DNS-03 or DNS-01 output validated as DNS-04 result_version 0.1.0, and assert rejection rather than best-effort interpretation, coercion or emission of a second combination. Verify registry Draft status and version metadata alone do not mark any other version supported. Requirements: REQ-DNS-04-0024.
- Requirement aspects exercised:

  - REQ-DNS-04-0024: Task test_version 0.1.1, 0.2.0 and 1.0.0 are task_unsupported and empty, v-prefixed, two-part, padded, numeric, null and missing values are task_schema_invalid, with no handle, lookup, contact or result; the emitted result decodes under the DNS-04 0.1.0 binding, which rejects each altered schema_version, test.version or test.result_version, DNS-03 and DNS-01 identities, foreign outcomes, engines and query types, extra fields and missing or reordered stages, and the bytes are rejected under DNS-03, DNS-01 and bootstrap bindings; results produced by the DNS-03 and DNS-01 runners are rejected under the DNS-04 binding with their own and a relabelled test.id (TestMDns04CrossTestOutputs); the production registry holds only one DNS-04 entry (TestProductionRegistryContainsOnlyDeliveredMeasurements).

## Measurement Unit-Test Cases

<a id="UT.internal.result.TestDNSCollectionAndHTTPEvidence"></a>

### UT.internal.result.TestDNSCollectionAndHTTPEvidence

- Case: `UT.internal.result.TestDNSCollectionAndHTTPEvidence`
- Result: passed
- Test: `TestDNSCollectionAndHTTPEvidence` in `internal/result/payload_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestDNSCollectionAndHTTPEvidence$' ./internal/result`
- Scope: Existing synthetic Go unit test TestDNSCollectionAndHTTPEvidence in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-01-0038: Decodes a UDP DNS-01 result with NOERROR and NXDOMAIN responses as conflicting_responses, rejects the same evidence labelled resolved, and with response_limit_reached:true accepts outcome response_limit_reached while rejecting conflicting_responses.
  - REQ-HTTP-01-0036: Rejects an HTTP-01 expected_response result when the binding configures one body check that is absent from the http_exchange checks.
  - REQ-NET-03-0027: For HTTP NET-03 accepts validated with response_class expected, and rejects validated with response_class missing or redirect, and connection_failure carrying a response_class.

<a id="UT.internal.result.TestDoHNonSuccessResponseAndRedirectBounds"></a>

### UT.internal.result.TestDoHNonSuccessResponseAndRedirectBounds

- Case: `UT.internal.result.TestDoHNonSuccessResponseAndRedirectBounds`
- Result: passed
- Test: `TestDoHNonSuccessResponseAndRedirectBounds` in `internal/result/payload_test.go` (package `internal/result`)
- Reproduce one case: `go test -count=1 -run '^TestDoHNonSuccessResponseAndRedirectBounds$' ./internal/result`
- Scope: Existing synthetic Go unit test TestDoHNonSuccessResponseAndRedirectBounds in package internal/result, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-03-0027: Accepts a DNS-03 http_failure result whose terminal http_exchange succeeded with status 503 and no DNS observation, and rejects the same result when the status is 200.
  - REQ-DNS-03-0040: A non-2xx `http_failure` result carries no `doh` DNS observation, and a 2xx status cannot be reported as `http_failure`.

<a id="UT.root.TestMTcp01AddressFamilies"></a>

### UT.root.TestMTcp01AddressFamilies

- Case: `UT.root.TestMTcp01AddressFamilies`
- Result: passed
- Test: `TestMTcp01AddressFamilies` in `tcp01_conformance_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMTcp01AddressFamilies$' ./.`
- Scope: Existing synthetic Go unit test TestMTcp01AddressFamilies in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TCP-01-0017: With the selected family available the run connects and records that family; with it unavailable or unknown the attempt is incomplete with constraints_unsatisfied and no dial.
  - REQ-TCP-01-0019: Before contact the runner confirms the recorded network generation and family through the execution check under the attempt deadline; an unavailable or unknown family stops the attempt with constraints_unsatisfied and no dial.

<a id="UT.root.TestMTcp01ExecutionAndOutput"></a>

### UT.root.TestMTcp01ExecutionAndOutput

- Case: `UT.root.TestMTcp01ExecutionAndOutput`
- Result: passed
- Test: `TestMTcp01ExecutionAndOutput` in `tcp01_conformance_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMTcp01ExecutionAndOutput$' ./.`
- Scope: Existing synthetic Go unit test TestMTcp01ExecutionAndOutput in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TCP-01-0020: The proxy row runs only through the configured scripted CONNECT proxy, with no direct dial.
  - REQ-TCP-01-0027: The serialized result contains neither the destination literal nor port, local address, payload or raw error canaries.

<a id="UT.root.TestMDns01ExecutionAndOutput"></a>

### UT.root.TestMDns01ExecutionAndOutput

- Case: `UT.root.TestMDns01ExecutionAndOutput`
- Result: passed
- Test: `TestMDns01ExecutionAndOutput` in `dns01_conformance_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMDns01ExecutionAndOutput$' ./.`
- Scope: Existing synthetic Go unit test TestMDns01ExecutionAndOutput in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-01-0040: The serialized result contains none of the hostname, resolver address, raw error, ttl, query_id or port canaries.

<a id="UT.internal.execution.TestTcp01AllTerminationBoundaries"></a>

### UT.internal.execution.TestTcp01AllTerminationBoundaries

- Case: `UT.internal.execution.TestTcp01AllTerminationBoundaries`
- Result: passed
- Test: `TestTcp01AllTerminationBoundaries` in `internal/execution/tcp01_termination_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestTcp01AllTerminationBoundaries$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestTcp01AllTerminationBoundaries in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TCP-01-0026: At the engine and production-runner boundary, seven causes at four stages give not_run before start, incomplete at the generation check and connect, and completed during close, with no dial before connect and no leaked timer or watch.
  - REQ-RESULT-0051: For every common cause, an interruption during context capture is not_run and one at the generation/family check or connect is incomplete with exactly that reason and no measurement.
  - REQ-RESULT-0052: An interruption at close after the connected outcome is recorded leaves the attempt completed with no reason.
  - REQ-RESULT-0053: A not_run attempt carries no network_context.

<a id="UT.internal.execution.TestTcp01ConnectionMeter"></a>

### UT.internal.execution.TestTcp01ConnectionMeter

- Case: `UT.internal.execution.TestTcp01ConnectionMeter`
- Result: passed
- Test: `TestTcp01ConnectionMeter` in `internal/execution/tcp01_termination_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestTcp01ConnectionMeter$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestTcp01ConnectionMeter in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TCP-01-0022: max_connections 0 refuses the dial before the transport with resource_limit and zero usage; at 1 and 2 the usage is exactly one connection and no bytes.
  - REQ-LIMIT-0006: With max_connections 0 the TCP-01 dial is refused before the transport and the attempt ends incomplete with resource_limit; at 1 and 2 the metered usage is exactly one connection.

<a id="UT.internal.execution.TestTcp01ReleasesSocketOnEveryPath"></a>

### UT.internal.execution.TestTcp01ReleasesSocketOnEveryPath

- Case: `UT.internal.execution.TestTcp01ReleasesSocketOnEveryPath`
- Result: passed
- Test: `TestTcp01ReleasesSocketOnEveryPath` in `internal/execution/tcp01_termination_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestTcp01ReleasesSocketOnEveryPath$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestTcp01ReleasesSocketOnEveryPath in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TCP-01-0022: A socket returned after host cancellation, after the attempt deadline or after the operation bound, and a connected socket, is closed exactly once; a refusal is not retried.

<a id="UT.internal.execution.TestDns01AllTerminationBoundaries"></a>

### UT.internal.execution.TestDns01AllTerminationBoundaries

- Case: `UT.internal.execution.TestDns01AllTerminationBoundaries`
- Result: passed
- Test: `TestDns01AllTerminationBoundaries` in `internal/execution/dns01_termination_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestDns01AllTerminationBoundaries$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestDns01AllTerminationBoundaries in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-01-0039: At the engine and production-runner boundary, seven causes at every UDP, TCP and system stage give the GO-SPEC-ISSUE-004 states, with no transport operation after the interruption.
  - REQ-RESULT-0051: For UDP and TCP and seven injected causes, an interruption during context capture is not_run, and one at any stage from the generation/family check through response windows is incomplete with exactly the first reason and no measurement.
  - REQ-RESULT-0052: An interruption while the connection closes after the resolved outcome is recorded leaves the attempt completed with no reason and the resolved measurement.
  - REQ-RESULT-0053: A not_run attempt carries no network_context.

<a id="UT.internal.execution.TestDnsUDPSourceAndTruncation"></a>

### UT.internal.execution.TestDnsUDPSourceAndTruncation

- Case: `UT.internal.execution.TestDnsUDPSourceAndTruncation`
- Result: passed
- Test: `TestDnsUDPSourceAndTruncation` in `internal/execution/dns01_transport_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestDnsUDPSourceAndTruncation$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestDnsUDPSourceAndTruncation in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-01-0010: The production UDP datagram reader reports a reply from another source as non-matching and an oversized datagram as truncated.

<a id="UT.internal.execution.TestDefaultAddressPolicy"></a>

### UT.internal.execution.TestDefaultAddressPolicy

- Case: `UT.internal.execution.TestDefaultAddressPolicy`
- Result: passed
- Test: `TestDefaultAddressPolicy` in `internal/execution/targets_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestDefaultAddressPolicy$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestDefaultAddressPolicy in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TCP-01-0012: The default execution address policy refuses every special-purpose class and admits public controls.

<a id="UT.internal.measurement.dns01.TestCodecFieldTypesAndModeRelations"></a>

### UT.internal.measurement.dns01.TestCodecFieldTypesAndModeRelations

- Case: `UT.internal.measurement.dns01.TestCodecFieldTypesAndModeRelations`
- Result: passed
- Test: `TestCodecFieldTypesAndModeRelations` in `internal/measurement/dns01/codec_test.go` (package `internal/measurement/dns01`)
- Reproduce one case: `go test -count=1 -run '^TestCodecFieldTypesAndModeRelations$' ./internal/measurement/dns01`
- Scope: Existing synthetic Go unit test TestCodecFieldTypesAndModeRelations in package internal/measurement/dns01, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-01-0014: The codec rejects null, boolean, numeric, array and object values for each input member.
  - REQ-DNS-01-0015: The codec rejects wrong-type resolver addresses and ports and defaults the port to 53.
  - REQ-DNS-01-0025: The codec applies the system, UDP and no-collection defaults and rejects direct-only options in system mode.
  - REQ-DNS-01-0026: The codec rejects null, array and object option values.

<a id="UT.internal.measurement.dns01.TestCodecNormalizedNameBoundaries"></a>

### UT.internal.measurement.dns01.TestCodecNormalizedNameBoundaries

- Case: `UT.internal.measurement.dns01.TestCodecNormalizedNameBoundaries`
- Result: passed
- Test: `TestCodecNormalizedNameBoundaries` in `internal/measurement/dns01/codec_test.go` (package `internal/measurement/dns01`)
- Reproduce one case: `go test -count=1 -run '^TestCodecNormalizedNameBoundaries$' ./internal/measurement/dns01`
- Scope: Existing synthetic Go unit test TestCodecNormalizedNameBoundaries in package internal/measurement/dns01, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-01-0016: The codec admits 252- and 253-byte normalized names, rejects 254 bytes, and maps Unicode, punycode and upper-case spellings to one A-label.

<a id="UT.internal.protocol.dns.TestQueryCorrelationAndCollectionPrecedence"></a>

### UT.internal.protocol.dns.TestQueryCorrelationAndCollectionPrecedence

- Case: `UT.internal.protocol.dns.TestQueryCorrelationAndCollectionPrecedence`
- Result: passed
- Test: `TestQueryCorrelationAndCollectionPrecedence` in `internal/protocol/dns/query_test.go` (package `internal/protocol/dns`)
- Reproduce one case: `go test -count=1 -run '^TestQueryCorrelationAndCollectionPrecedence$' ./internal/protocol/dns`
- Scope: Existing synthetic Go unit test TestQueryCorrelationAndCollectionPrecedence in package internal/protocol/dns, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-01-0010: The query primitive discards mismatched, foreign-question and wrong-source datagrams and accepts a later usable reply, with one dial and one write.
  - REQ-DNS-01-0013: It yields query_mismatch, unexpected_source, malformed_response and timeout for the matching scripted packets.
  - REQ-DNS-01-0038: A malformed datagram after a usable reply does not override it, identical replies stay resolved and the count bound wins.

<a id="UT.internal.protocol.dns.TestWireMessageBoundaries"></a>

### UT.internal.protocol.dns.TestWireMessageBoundaries

- Case: `UT.internal.protocol.dns.TestWireMessageBoundaries`
- Result: passed
- Test: `TestWireMessageBoundaries` in `internal/protocol/dns/wire_test.go` (package `internal/protocol/dns`)
- Reproduce one case: `go test -count=1 -run '^TestWireMessageBoundaries$' ./internal/protocol/dns`
- Scope: Existing synthetic Go unit test TestWireMessageBoundaries in package internal/protocol/dns, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-01-0033: The parser accepts UDP 4,096 and TCP 65,535-byte messages at and below the limit and rejects UDP above it.
  - REQ-LIMIT-0014: DNS messages one byte below and at the 4,096-byte UDP and 65,535-byte TCP bounds parse and one byte above the UDP bound is rejected.

<a id="UT.internal.protocol.dns.TestWireRetainsOnlyRelevantPublicAnswers"></a>

### UT.internal.protocol.dns.TestWireRetainsOnlyRelevantPublicAnswers

- Case: `UT.internal.protocol.dns.TestWireRetainsOnlyRelevantPublicAnswers`
- Result: passed
- Test: `TestWireRetainsOnlyRelevantPublicAnswers` in `internal/protocol/dns/wire_test.go` (package `internal/protocol/dns`)
- Reproduce one case: `go test -count=1 -run '^TestWireRetainsOnlyRelevantPublicAnswers$' ./internal/protocol/dns`
- Scope: Existing synthetic Go unit test TestWireRetainsOnlyRelevantPublicAnswers in package internal/protocol/dns, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-01-0022: A CNAME chain is retained as data only.
  - REQ-DNS-01-0036: Unrelated owners and non-public answers are dropped, keeping the CNAME and requested-type answer in order.
  - REQ-OBS-0013: Retained answers keep first-occurrence order (CNAME then A) and an exact duplicate A record is removed.

<a id="UT.root.TestMTls01Normalization"></a>

### UT.root.TestMTls01Normalization

- Case: `UT.root.TestMTls01Normalization`
- Result: passed
- Test: `TestMTls01Normalization` in `tls01_conformance_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMTls01Normalization$' ./.`
- Scope: Existing synthetic Go unit test TestMTls01Normalization in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TLS-01-0028: Across every outcome row the serialized result contains no destination address, port, server name, other name, certificate canary or serial, raw alert code, raw library error text or loopback address.
  - REQ-OBS-0019: Real loopback behaviour yields TLS failures timeout, connection_reset, alert, protocol_error, certificate_untrusted, certificate_name_mismatch, certificate_time_invalid and alpn_no_overlap.
  - REQ-OBS-0043: Successful handshakes retain address family, version and validation, with alpn present only when negotiated.
  - REQ-OBS-0048: A negotiated TLS 1.3 handshake retains TLS1.3.
  - REQ-OBS-0049: A negotiated h2 is retained as h2 and a server offering only dot yields alpn_no_overlap.
  - REQ-OBS-0050: Peer alerts 40, 42, 70, 49, 80, 200 and 120 (without an offer) map to handshake_failure, certificate_error, protocol_error, access_denied, internal_error, other and handshake_failure.
  - REQ-OBS-0052: A successful handshake negotiating `h2` retains `alpn` `h2`, one without ALPN retains no `alpn` value, and every failure row, including `alpn_no_overlap` and certificate failures against a server offering `h2`, retains no `alpn` value.

<a id="UT.root.TestMTls01ExecutionAndOutput"></a>

### UT.root.TestMTls01ExecutionAndOutput

- Case: `UT.root.TestMTls01ExecutionAndOutput`
- Result: passed
- Test: `TestMTls01ExecutionAndOutput` in `tls01_conformance_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMTls01ExecutionAndOutput$' ./.`
- Scope: Existing synthetic Go unit test TestMTls01ExecutionAndOutput in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TLS-01-0035: The server observes one standard ClientHello with the configured server name and ALPN offer and no second or comparative handshake, on the direct and proxy paths.
  - REQ-TLS-01-0036: The server receives no plaintext application data after the handshake.

<a id="UT.root.TestMTls01NoImplicitDNS"></a>

### UT.root.TestMTls01NoImplicitDNS

- Case: `UT.root.TestMTls01NoImplicitDNS`
- Result: passed
- Test: `TestMTls01NoImplicitDNS` in `tls01_execution_closeout_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMTls01NoImplicitDNS$' ./.`
- Scope: Existing synthetic Go unit test TestMTls01NoImplicitDNS in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TLS-01-0005: The one connection uses the explicit IP literal; the transport refuses any other authority.
  - REQ-TLS-01-0006: For a successful handshake and a refused connect, the installed system resolver is never called and exactly one dial is made.

<a id="UT.root.TestMTls01NoSessionResumption"></a>

### UT.root.TestMTls01NoSessionResumption

- Case: `UT.root.TestMTls01NoSessionResumption`
- Result: passed
- Test: `TestMTls01NoSessionResumption` in `tls01_execution_closeout_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMTls01NoSessionResumption$' ./.`
- Scope: Existing synthetic Go unit test TestMTls01NoSessionResumption in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TLS-01-0005: Two consecutive TLS-01 runs against a TLS 1.2 session-ticket endpoint both complete full trusted handshakes with ALPN h2.
  - REQ-TLS-01-0017: Neither run resumes a session, while a control client with a session cache does resume on the same kind of endpoint, so the assertion is not vacuous.

<a id="UT.internal.execution.TestTls01PlaintextAccounting"></a>

### UT.internal.execution.TestTls01PlaintextAccounting

- Case: `UT.internal.execution.TestTls01PlaintextAccounting`
- Result: passed
- Test: `TestTls01PlaintextAccounting` in `internal/execution/tls01_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestTls01PlaintextAccounting$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestTls01PlaintextAccounting in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TLS-01-0023: On the direct and proxy paths the attempt uses exactly one connection and zero metered bytes with a connection limit of 1.
  - REQ-TLS-01-0025: The completed measurement has two observations with ALPN h2, TLS1.3 and certificate_validation trusted.
  - REQ-LIMIT-0034: A completed TLS-01 handshake meters one connection and zero sent and received bytes on direct and proxy paths, while the server and tunnel saw non-zero TLS bytes, so handshake and record bytes are not counted.

<a id="UT.internal.execution.TestTls01AllTerminationBoundaries"></a>

### UT.internal.execution.TestTls01AllTerminationBoundaries

- Case: `UT.internal.execution.TestTls01AllTerminationBoundaries`
- Result: passed
- Test: `TestTls01AllTerminationBoundaries` in `internal/execution/tls01_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestTls01AllTerminationBoundaries$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestTls01AllTerminationBoundaries in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TLS-01-0027: For seven common causes at six procedure points, the attempt is not_run before the start, incomplete with the first cause and no measurement from the network check through the stalled handshake, and completed with handshake_succeeded when interrupted during close.
  - REQ-RESULT-0051: For every common cause, an interruption before the network check is not_run and one at the generation/family check, connect, before hello or during the handshake is incomplete with exactly that reason and no measurement.
  - REQ-RESULT-0052: An interruption while the session closes after handshake_succeeded is recorded leaves the attempt completed with no reason.
  - REQ-RESULT-0053: A not_run attempt carries no network_context.

<a id="UT.internal.measurement.tls01.TestMTls01ALPNBoundsIsolated"></a>

### UT.internal.measurement.tls01.TestMTls01ALPNBoundsIsolated

- Case: `UT.internal.measurement.tls01.TestMTls01ALPNBoundsIsolated`
- Result: passed
- Test: `TestMTls01ALPNBoundsIsolated` in `internal/measurement/tls01/tls01_options_test.go` (package `internal/measurement/tls01`)
- Reproduce one case: `go test -count=1 -run '^TestMTls01ALPNBoundsIsolated$' ./internal/measurement/tls01`
- Scope: Existing synthetic Go unit test TestMTls01ALPNBoundsIsolated in package internal/measurement/tls01, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TLS-01-0016: The options descriptor accepts eight distinct identifiers and rejects nine; normalization rejects unregistered 1-, 32- and 33-byte identifiers, duplicates and an empty identifier, and keeps registered offers in order.
  - REQ-OBS-0049: Only http/1.1, h2 and dot (singly or together, in order) normalize as offers; unregistered identifiers of 1, 32 and 33 bytes, eight distinct unregistered identifiers, duplicates and the empty string are rejected, and the options descriptor rejects nine entries.

<a id="UT.root.TestMHttp01CheckSemantics"></a>

### UT.root.TestMHttp01CheckSemantics

- Case: `UT.root.TestMHttp01CheckSemantics`
- Result: passed
- Test: `TestMHttp01CheckSemantics` in `http01_closeout_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMHttp01CheckSemantics$' ./.`
- Scope: Existing synthetic Go unit test TestMHttp01CheckSemantics in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-HTTP-01-0012: contains, prefix and suffix body checks are case-sensitive and every body check must match.
  - REQ-HTTP-01-0013: Header names in equals, present and contains checks match case-insensitively.
  - REQ-HTTP-01-0014: Multivalue headers match when any value matches, fail when none does, and every header check must match.
  - REQ-HTTP-01-0037: Indexed results retain decided checks only: truncation omits undecidable body checks, and a decided status mismatch keeps the later decided header and body results.

<a id="UT.root.TestMHttp01EquivalentSpellingSerialized"></a>

### UT.root.TestMHttp01EquivalentSpellingSerialized

- Case: `UT.root.TestMHttp01EquivalentSpellingSerialized`
- Result: passed
- Test: `TestMHttp01EquivalentSpellingSerialized` in `http01_closeout_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMHttp01EquivalentSpellingSerialized$' ./.`
- Scope: Existing synthetic Go unit test TestMHttp01EquivalentSpellingSerialized in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-HTTP-01-0032: HTTPS://8.8.8.8:443/check normalizes to the active run's URL and waits without a dial or request; cancelled, it is not_run and nothing is contacted after the active run is cancelled.

<a id="UT.root.TestMHttp01InconclusiveResultShape"></a>

### UT.root.TestMHttp01InconclusiveResultShape

- Case: `UT.root.TestMHttp01InconclusiveResultShape`
- Result: passed
- Test: `TestMHttp01InconclusiveResultShape` in `http01_closeout_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMHttp01InconclusiveResultShape$' ./.`
- Scope: Existing synthetic Go unit test TestMHttp01InconclusiveResultShape in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-HTTP-01-0036: The bound decoder accepts an inconclusive outcome with an inconclusive http_exchange failure, empty body metadata and only the status check, and rejects the same shape with a timeout failure.

<a id="UT.root.TestMHttp01OptionsAddressFamilyAndDefaults"></a>

### UT.root.TestMHttp01OptionsAddressFamilyAndDefaults

- Case: `UT.root.TestMHttp01OptionsAddressFamilyAndDefaults`
- Result: passed
- Test: `TestMHttp01OptionsAddressFamilyAndDefaults` in `http01_closeout_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMHttp01OptionsAddressFamilyAndDefaults$' ./.`
- Scope: Existing synthetic Go unit test TestMHttp01OptionsAddressFamilyAndDefaults in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-HTTP-01-0024: ipv6 is rejected for an IPv4 literal and ipv4 for an IPv6 literal; an IPv6 literal completes with ipv6 observations under ipv6, auto and the default; the default body bound is 4,096 bytes at normalization and at run time, and is rejected against max_bytes_received 4,095 with task_limit_exceeded and no contact.

<a id="UT.root.TestMHttp01OptionsOverridesRejected"></a>

### UT.root.TestMHttp01OptionsOverridesRejected

- Case: `UT.root.TestMHttp01OptionsOverridesRejected`
- Result: passed
- Test: `TestMHttp01OptionsOverridesRejected` in `http01_closeout_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMHttp01OptionsOverridesRejected$' ./.`
- Scope: Existing synthetic Go unit test TestMHttp01OptionsOverridesRejected in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-HTTP-01-0024: Null follow_redirects, max_redirects and max_body_bytes values are rejected.
  - REQ-HTTP-01-0025: Method, header, user-agent, body, cookie, credential, authorization, proxy, path, compression, encoding, TLS, server-name, ALPN and insecure-verification overrides are task_schema_invalid with no handle, replay commit or contact.

<a id="UT.root.TestMHttp01OuterBoundsBoundary"></a>

### UT.root.TestMHttp01OuterBoundsBoundary

- Case: `UT.root.TestMHttp01OuterBoundsBoundary`
- Result: passed
- Test: `TestMHttp01OuterBoundsBoundary` in `http01_closeout_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMHttp01OuterBoundsBoundary$' ./.`
- Scope: Existing synthetic Go unit test TestMHttp01OuterBoundsBoundary in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-HTTP-01-0032: On the direct and CONNECT paths, sent, received and connection limits at measured usage +1 and equal complete and at -1 are incomplete resource_limit, with the expected number of accepted connections and no direct fallback; the CONNECT exchange is not charged.
  - REQ-LIMIT-0034: On direct and CONNECT proxy paths, max_bytes_sent equal to the server-observed plaintext request bytes and max_bytes_received equal to the raw response length complete while one byte less ends incomplete with resource_limit; equality on the proxy path shows the CONNECT exchange is not charged.

<a id="UT.root.TestMHttp01ResultVersionShapes"></a>

### UT.root.TestMHttp01ResultVersionShapes

- Case: `UT.root.TestMHttp01ResultVersionShapes`
- Result: passed
- Test: `TestMHttp01ResultVersionShapes` in `http01_closeout_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMHttp01ResultVersionShapes$' ./.`
- Scope: Existing synthetic Go unit test TestMHttp01ResultVersionShapes in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-HTTP-01-0033: The decoder rejects missing schema_version, test.version, result_version and test.id, test.id NET-03, NET-01 access fields and NET-05 profile, steps and network_validation fields.

<a id="UT.root.TestMHttp01TaskVersionShapes"></a>

### UT.root.TestMHttp01TaskVersionShapes

- Case: `UT.root.TestMHttp01TaskVersionShapes`
- Result: passed
- Test: `TestMHttp01TaskVersionShapes` in `http01_closeout_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMHttp01TaskVersionShapes$' ./.`
- Scope: Existing synthetic Go unit test TestMHttp01TaskVersionShapes in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-HTTP-01-0033: Task test_version 0.1, null, 1, a tab-padded value and a missing member are task_schema_invalid, and test_id http-01, NET-03, TLS-01, NET-01 and NET-05 are rejected, all with no handle, replay commit or contact.

<a id="UT.root.TestMHttp01Output"></a>

### UT.root.TestMHttp01Output

- Case: `UT.root.TestMHttp01Output`
- Result: passed
- Test: `TestMHttp01Output` in `http01_conformance_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMHttp01Output$' ./.`
- Scope: Existing synthetic Go unit test TestMHttp01Output in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-HTTP-01-0039: A result with expected status, body and header checks contains no URL, path, address, header name or value, body or certificate canary.

<a id="UT.root.TestMHttp01ResultVersionsAndClosedOutput"></a>

### UT.root.TestMHttp01ResultVersionsAndClosedOutput

- Case: `UT.root.TestMHttp01ResultVersionsAndClosedOutput`
- Result: passed
- Test: `TestMHttp01ResultVersionsAndClosedOutput` in `http01_conformance_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMHttp01ResultVersionsAndClosedOutput$' ./.`
- Scope: Existing synthetic Go unit test TestMHttp01ResultVersionsAndClosedOutput in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-HTTP-01-0039: The decoder rejects body_content, expected_digest, url, header_value, debug_log and censorship details added to the terminal observation.

<a id="UT.internal.execution.TestHttp01RedirectAndHopTerminationPoints"></a>

### UT.internal.execution.TestHttp01RedirectAndHopTerminationPoints

- Case: `UT.internal.execution.TestHttp01RedirectAndHopTerminationPoints`
- Result: passed
- Test: `TestHttp01RedirectAndHopTerminationPoints` in `internal/execution/http01_termination_points_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestHttp01RedirectAndHopTerminationPoints$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestHttp01RedirectAndHopTerminationPoints in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-HTTP-01-0038: For eight common causes after DNS before TCP, at a redirect response with following disabled and enabled, and at each redirect hop's DNS, TCP, TLS and header stage for max_redirects 1 and 2, the attempt is incomplete with exactly that reason and no measurement, with the expected dials, lookups and requests and every connection closed.
  - REQ-RESULT-0051: Interruptions after DNS, at a redirect response before next-URL validation, and at each redirect hop's DNS, TCP, TLS and HTTP stage are incomplete with exactly the injected reason and no measurement.
  - REQ-RESULT-0052: Only the pre-outcome side: interruptions before the terminal outcome is recorded, including after a redirect response is received, never yield completed; post-outcome events are covered by TestHttp01AllTerminationBoundaries.

<a id="UT.internal.execution.TestHttp01AllTerminationBoundaries"></a>

### UT.internal.execution.TestHttp01AllTerminationBoundaries

- Case: `UT.internal.execution.TestHttp01AllTerminationBoundaries`
- Result: passed
- Test: `TestHttp01AllTerminationBoundaries` in `internal/execution/http01_termination_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestHttp01AllTerminationBoundaries$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestHttp01AllTerminationBoundaries in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-HTTP-01-0038: For eight common causes, an interruption before the start is not_run with no network_context, during DNS, connect, TLS, headers and body it is incomplete with that reason and no partial output, and during cleanup it stays completed with expected_response.
  - REQ-RESULT-0051: For every common cause, an interruption during context capture is not_run and one at the DNS, connect, TLS, header or body stage is incomplete with exactly that reason and no measurement.
  - REQ-RESULT-0052: An interruption during cleanup after the expected_response outcome is recorded leaves the attempt completed with that outcome and no reason.
  - REQ-RESULT-0053: A not_run attempt carries no network_context.

<a id="UT.internal.execution.TestHTTPApprovedRedirectChain"></a>

### UT.internal.execution.TestHTTPApprovedRedirectChain

- Case: `UT.internal.execution.TestHTTPApprovedRedirectChain`
- Result: passed
- Test: `TestHTTPApprovedRedirectChain` in `internal/execution/http_primitive_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestHTTPApprovedRedirectChain$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestHTTPApprovedRedirectChain in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-HTTP-01-0018: A two-hop chain of registered redirect targets is followed on the direct and proxy paths with the followed count per hop and the attempt meter equal to the plaintext bytes and three connections.
  - REQ-OBS-0026: A two-hop approved redirect chain records followed_count 1, 2 and 2 on the three HTTP observations.
  - REQ-LIMIT-0034: Metered usage over the chain equals the sum of plaintext bytes all three servers read and wrote and three connections, on direct and proxy paths, so followed redirects count.

<a id="UT.internal.execution.TestHTTPRedirectUsesProductionDestinationGate"></a>

### UT.internal.execution.TestHTTPRedirectUsesProductionDestinationGate

- Case: `UT.internal.execution.TestHTTPRedirectUsesProductionDestinationGate`
- Result: passed
- Test: `TestHTTPRedirectUsesProductionDestinationGate` in `internal/execution/http_primitive_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestHTTPRedirectUsesProductionDestinationGate$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestHTTPRedirectUsesProductionDestinationGate in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-HTTP-01-0018: On the direct and proxy paths, loopback and unregistered redirect targets are redirect_rejected without a second connection.
  - REQ-HTTP-01-0051: On the proxy path a registered hostname redirect is redirect_rejected before any resolver call.

<a id="UT.internal.measurement.http01.TestCodecBoundsAndNormalization"></a>

### UT.internal.measurement.http01.TestCodecBoundsAndNormalization

- Case: `UT.internal.measurement.http01.TestCodecBoundsAndNormalization`
- Result: passed
- Test: `TestCodecBoundsAndNormalization` in `internal/measurement/http01/codec_test.go` (package `internal/measurement/http01`)
- Reproduce one case: `go test -count=1 -run '^TestCodecBoundsAndNormalization$' ./internal/measurement/http01`
- Scope: Existing synthetic Go unit test TestCodecBoundsAndNormalization in package internal/measurement/http01, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-HTTP-01-0010: URLs of 2,047 and 2,048 bytes are accepted and 2,049 rejected; uppercase scheme and host and the default port are normalized away; ports 0, 65536 and empty, an empty fragment, an interpolated host and an IPv4-mapped literal are rejected.
  - REQ-HTTP-01-0011: A 64-character lowercase sha256 body check is accepted.
  - REQ-HTTP-01-0013: Header names are normalized to lower case and the returned tuple cannot mutate the stored checks.
  - REQ-HTTP-01-0020: 0, 1 and 8 checks with 1- and 256-byte values are accepted; 9 checks and 257-byte values are rejected for body and header checks.

<a id="UT.internal.protocol.http.TestMHttp01NormalizationInconclusive"></a>

### UT.internal.protocol.http.TestMHttp01NormalizationInconclusive

- Case: `UT.internal.protocol.http.TestMHttp01NormalizationInconclusive`
- Result: passed
- Test: `TestMHttp01NormalizationInconclusive` in `internal/protocol/http/http01_inconclusive_test.go` (package `internal/protocol/http`)
- Reproduce one case: `go test -count=1 -run '^TestMHttp01NormalizationInconclusive$' ./internal/protocol/http`
- Scope: Existing synthetic Go unit test TestMHttp01NormalizationInconclusive in package internal/protocol/http, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-HTTP-01-0036: A body read that makes no progress and returns no error yields outcome inconclusive with one http_exchange failed inconclusive and empty body metadata.
  - REQ-HTTP-01-0037: Only the decided status and header results are retained and no body check is invented.
  - REQ-HTTP-01-0052: The inconclusive body read wins over the already failed status check, with no further read or connection and the stream closed.

<a id="UT.internal.execution.TestMNet05ConnectionCeiling"></a>

### UT.internal.execution.TestMNet05ConnectionCeiling

- Case: `UT.internal.execution.TestMNet05ConnectionCeiling`
- Result: passed
- Test: `TestMNet05ConnectionCeiling` in `internal/execution/net05_closeout_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestMNet05ConnectionCeiling$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestMNet05ConnectionCeiling in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0037: Four HTTP-01 steps that each follow two registered redirects need 12 connections: with a limit of 11 the fourth step is interrupted, at 12 the profile completes, and a following TCP-01 step needing a 13th connection is refused by the aggregate meter and becomes the single final incomplete step with resource_limit, the earlier steps kept.
  - REQ-LIMIT-0019: A profile needing 12 connections completes with an aggregate limit of 12, ends incomplete with resource_limit at 11, and a step needing a 13th connection ends as an incomplete final step, with usage never above 12.

<a id="UT.internal.execution.TestMNet05DurationCeiling"></a>

### UT.internal.execution.TestMNet05DurationCeiling

- Case: `UT.internal.execution.TestMNet05DurationCeiling`
- Result: passed
- Test: `TestMNet05DurationCeiling` in `internal/execution/net05_closeout_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestMNet05DurationCeiling$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestMNet05DurationCeiling in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0037: With the plan deadline at 60,000 ms, elapsed time one nanosecond below the ceiling completes the profile and reaching the ceiling gives incomplete attempt_deadline_exceeded with no further dial and no leaked timer.

<a id="UT.internal.execution.TestMNet05ReceivedCeiling"></a>

### UT.internal.execution.TestMNet05ReceivedCeiling

- Case: `UT.internal.execution.TestMNet05ReceivedCeiling`
- Result: passed
- Test: `TestMNet05ReceivedCeiling` in `internal/execution/net05_closeout_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestMNet05ReceivedCeiling$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestMNet05ReceivedCeiling in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0037: Four HTTP-01 steps receiving 262,144 application bytes in total complete at the ceiling, one byte below completes, and one byte above interrupts the fourth step with resource_limit while the earlier steps keep their results.
  - REQ-LIMIT-0034: Four HTTP-01 steps receiving 65,535 or 65,536 application bytes each complete at the 262,144-byte limit with usage equal to their sum, and 65,537 interrupts the fourth step with resource_limit while earlier steps keep their results.

<a id="UT.internal.execution.TestMNet05ReceiverRejectsVersions"></a>

### UT.internal.execution.TestMNet05ReceiverRejectsVersions

- Case: `UT.internal.execution.TestMNet05ReceiverRejectsVersions`
- Result: passed
- Test: `TestMNet05ReceiverRejectsVersions` in `internal/execution/net05_closeout_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestMNet05ReceiverRejectsVersions$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestMNet05ReceiverRejectsVersions in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0039: The emitted result validates with schema_version, NET-05 test and result version 0.1.0; unknown, malformed or missing outer and step versions and other test IDs are rejected.
  - REQ-NET-05-0043: network_validation state validated with source platform is emitted; other or missing states, not_validated on a completed run and another source are rejected.
  - REQ-NET-05-0066: The decoder accepts source platform and rejects another source value.
  - REQ-NET-05-0044: Each step carries step_id s0 to s2 and the TCP-01 0.1.0 tuple; an extra step test member, a step common envelope and a step Network Context are rejected.

<a id="UT.internal.execution.TestMNet05StepNegativeOutcomesStayCompleted"></a>

### UT.internal.execution.TestMNet05StepNegativeOutcomesStayCompleted

- Case: `UT.internal.execution.TestMNet05StepNegativeOutcomesStayCompleted`
- Result: passed
- Test: `TestMNet05StepNegativeOutcomesStayCompleted` in `internal/execution/net05_closeout_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestMNet05StepNegativeOutcomesStayCompleted$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestMNet05StepNegativeOutcomesStayCompleted in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0046: Steps with TCP timeout, TLS timeout, untrusted certificate, refusal and success all complete and the composite outcome is completed.
  - REQ-NET-05-0047: Primitive operation timeouts below the outer deadline are step outcomes, not execution reasons.
  - REQ-NET-05-0061: The profile is retained with execution.state completed, five dials, five connections, released sockets and timers, and one envelope, execution object and Network Context.

<a id="UT.internal.execution.TestCcNet05PathContextAndProxyGeneration"></a>

### UT.internal.execution.TestCcNet05PathContextAndProxyGeneration

- Case: `UT.internal.execution.TestCcNet05PathContextAndProxyGeneration`
- Result: passed
- Test: `TestCcNet05PathContextAndProxyGeneration` in `internal/execution/net05_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcNet05PathContextAndProxyGeneration$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestCcNet05PathContextAndProxyGeneration in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0040: Direct DNS-01, TCP-01, TLS-01 and HTTP-01 steps and proxy TCP-01, TLS-01 and HTTP-01 steps complete in order with one network_context and an unchanged proxy path.
  - REQ-CONTEXT-0025: A completed NET-05 composite on direct and proxy paths carries exactly one network_context for the whole attempt; deadline, cancellation and network-change handling during acquisition are not exercised here.

<a id="UT.internal.execution.TestMNet05Termination"></a>

### UT.internal.execution.TestMNet05Termination

- Case: `UT.internal.execution.TestMNet05Termination`
- Result: passed
- Test: `TestMNet05Termination` in `internal/execution/net05_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestMNet05Termination$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestMNet05Termination in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0045: Across eight causes and interruptions before, during and after each of three steps, at most one final step is incomplete, it has no measurement, unstarted steps are absent and stopped_before_step names the next step.
  - REQ-RESULT-0051: For eight causes at before, during and after each step, the composite attempt ends incomplete with exactly the first reason and a partial profile.
  - REQ-RESULT-0052: An interruption while the final step's connection closes after the composite outcome is recorded leaves the attempt completed.

<a id="UT.internal.execution.TestNet05StepTargetAndRedirectIsolation"></a>

### UT.internal.execution.TestNet05StepTargetAndRedirectIsolation

- Case: `UT.internal.execution.TestNet05StepTargetAndRedirectIsolation`
- Result: passed
- Test: `TestNet05StepTargetAndRedirectIsolation` in `internal/execution/net05_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestNet05StepTargetAndRedirectIsolation$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestNet05StepTargetAndRedirectIsolation in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-HTTP-01-0051: A step can follow only the redirect registration linked from its own initial target.
  - REQ-NET-05-0019: Each step gate grants only its own registered tuple: another step's redirect link, a redirect target as an initial target, a tuple with different checks and a cross-test grant are refused.

<a id="UT.root.TestMNet05RegisteredCeilingAdmission"></a>

### UT.root.TestMNet05RegisteredCeilingAdmission

- Case: `UT.root.TestMNet05RegisteredCeilingAdmission`
- Result: passed
- Test: `TestMNet05RegisteredCeilingAdmission` in `net05_conformance_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMNet05RegisteredCeilingAdmission$' ./.`
- Scope: Existing synthetic Go unit test TestMNet05RegisteredCeilingAdmission in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0037: max_duration_ms 60,000, max_connections 12, max_bytes_sent 65,536 and max_bytes_received 262,144 are admitted at the ceiling and one below, and one above is task_limit_exceeded with no handle, replay commit or contact.
  - REQ-NET-05-0038: Each registered ceiling is enforced at admission before execution, independently of the others.
  - REQ-NET-05-0026: The four primitive steps, which declare 4 connections and 86,017 received bytes, are admitted at max_connections 4 and max_bytes_received 86,017 and rejected with task_limit_exceeded at 3 and 86,016; four HTTP-01 steps following two redirects are admitted at 12 connections, and with a TCP-01 step their 13 declared connections are rejected; no rejection leaves a handle, replay record or run or makes contact.
  - REQ-LIMIT-0020: The rejections above happen before network activity for the whole profile (REQ-LIMIT-0035, GO-SPEC-ISSUE-021).
  - REQ-LIMIT-0035: Over-budget declared sums are rejected at admission with `task_limit_exceeded` before contact, through `Client`, with no handle or replay record.

<a id="UT.root.TestCcNet05ProxyLifecycleAndRenewal"></a>

### UT.root.TestCcNet05ProxyLifecycleAndRenewal

- Case: `UT.root.TestCcNet05ProxyLifecycleAndRenewal`
- Result: passed
- Test: `TestCcNet05ProxyLifecycleAndRenewal` in `net05_runner_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestCcNet05ProxyLifecycleAndRenewal$' ./.`
- Scope: Existing synthetic Go unit test TestCcNet05ProxyLifecycleAndRenewal in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0061: Replacing or clearing the proxy during the HTTP-01 step gives incomplete proxy_reconfigured with the completed TCP-01 step, the interrupted HTTP-01 step and stopped_before_step last, and no new proxy, direct dial or later step.

<a id="UT.root.TestMNet05ExecutionAndOutput"></a>

### UT.root.TestMNet05ExecutionAndOutput

- Case: `UT.root.TestMNet05ExecutionAndOutput`
- Result: passed
- Test: `TestMNet05ExecutionAndOutput` in `net05_runner_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMNet05ExecutionAndOutput$' ./.`
- Scope: Existing synthetic Go unit test TestMNet05ExecutionAndOutput in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0006: Through Client admission and StartRun, DNS-01, TCP-01, TLS-01 and HTTP-01 steps run in profile order with three dials and their own outcomes resolved, connected, alpn_no_overlap and expected_response.
  - REQ-NET-05-0064: Validation context is the platform's validated state; only the declared steps' traffic is sent, three dials and one HTTP request to the HTTP-01 step's target, with no `NET-03` request.
  - REQ-NET-05-0065: After admission of the whole profile, the run reads the platform validation state, obtains or reuses no `NET-03` observation, and the production runner dispatches each step in order through the shared protocol operations.
  - REQ-NET-05-0066: network_validation.source is platform, and no fresh or cached value is emitted.
  - REQ-NET-05-0039: The emitted outer and step test objects carry the registered 0.1.0 versions.
  - REQ-NET-05-0048: The serialized result contains no target URL, address, header, body or certificate canary and exactly one network_context.

<a id="UT.root.TestMNet05PlatformValidationGate"></a>

### UT.root.TestMNet05PlatformValidationGate

- Case: `UT.root.TestMNet05PlatformValidationGate`
- Result: passed
- Test: `TestMNet05PlatformValidationGate` in `net05_runner_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMNet05PlatformValidationGate$' ./.`
- Scope: Existing synthetic Go unit test TestMNet05PlatformValidationGate in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0061: Not-validated and unknown platform validation give completed network_not_validated and inconclusive with no step and no contact; a task-level require_validated constraint is refused at StartRun with run_not_due and produces no measurement.

<a id="UT.root.TestNet05StepCountExecution"></a>

### UT.root.TestNet05StepCountExecution

- Case: `UT.root.TestNet05StepCountExecution`
- Result: passed
- Test: `TestNet05StepCountExecution` in `net05_runner_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestNet05StepCountExecution$' ./.`
- Scope: Existing synthetic Go unit test TestNet05StepCountExecution in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0006: Profiles of one and eight TCP-01 steps execute every step once, in order.
  - REQ-NET-05-0037: The eight-step profile completes within the registered step-count and connection bounds with one dial per step.

<a id="UT.root.TestMNet05Admission"></a>

### UT.root.TestMNet05Admission

- Case: `UT.root.TestMNet05Admission`
- Result: passed
- Test: `TestMNet05Admission` in `net05_conformance_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMNet05Admission$' ./.`
- Scope: Existing synthetic Go unit test TestMNet05Admission in package root, executed with go test at the implementation baseline. Shared by both manifests; partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0005: Profiles of 1, 4 and 8 primitive steps, each with its own target, are admitted with one replay commit, no contact and the NET-05 runner installed.
  - REQ-NET-05-0014: Empty input, null input and a null profile are rejected.
  - REQ-NET-05-0016: Zero and nine steps are rejected; step_id lengths 1 and 64 are accepted and 0, 65, non-ASCII, control and interpolated identifiers rejected; the normalized step container is immutable.
  - REQ-NET-05-0017: Only an inline profile object is accepted; null and missing profiles are rejected.
  - REQ-NET-05-0018: A profile whose final step has a credential URL or whose second step targets a loopback address is rejected with no handle, run or contact.
  - REQ-NET-05-0019: A loopback TCP-01 target and a credential HTTP-01 URL are rejected by their primitive validation inside the profile.
  - REQ-NET-05-0020: A profile repeating a step_id is rejected.
  - REQ-NET-05-0022: Steps containing schedule, execution, safety, labels, script or conditions members are rejected.
  - REQ-NET-05-0023: Credential URLs, loopback targets, interpolated identifiers and more than eight steps are rejected.
  - REQ-NET-05-0024: An unknown profile member and a duplicated profile_version member are rejected.
  - REQ-NET-05-0025: Through the Client a 49,140-byte profile is admitted with one replay commit and no contact, while profiles of 49,141, 49,151, 49,152 and 49,153 bytes are task_limit_exceeded because the common input bound includes the 12-byte wrapper; the codec alone accepts encoded profiles up to 49,152 bytes (GO-SPEC-ISSUE-018).
  - REQ-NET-05-0027: Every rejection returns no handle and leaves no task handle, replay commit, run or transport contact.
  - REQ-NET-05-0062: profile_id lengths 1 and 64 are accepted and 0 and 65, non-ASCII, control and interpolated identifiers rejected; profile_version 0, -1, 1.5, a string and null are rejected.
  - REQ-LIMIT-0019: The 49,152-byte profile bound applies in addition to the equal common input bound and the stricter decides: through the Client a 49,140-byte profile is admitted and 49,141 bytes and more are task_limit_exceeded, the common input-bound class, while the codec alone accepts up to 49,152 bytes (GO-SPEC-ISSUE-018).

<a id="UT.root.TestMNet05ProfileVersionWithoutHistory"></a>

### UT.root.TestMNet05ProfileVersionWithoutHistory

- Case: `UT.root.TestMNet05ProfileVersionWithoutHistory`
- Result: passed
- Test: `TestMNet05ProfileVersionWithoutHistory` in `net05_conformance_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMNet05ProfileVersionWithoutHistory$' ./.`
- Scope: Existing synthetic Go unit test TestMNet05ProfileVersionWithoutHistory in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0062: Positive profile_version values 5, 3, 5, 1, the signed 64-bit maximum and 2 are admitted, and the normalized profile retains each value unchanged with profile_id baseline.
  - REQ-NET-05-0063: Equal and lower versions for one profile_id are admitted in one client instance, giving six task handles and six replay commits with no transport contact; no version is compared with another task's.

<a id="UT.root.TestMNet05CodecBoundsAndComposition"></a>

### UT.root.TestMNet05CodecBoundsAndComposition

- Case: `UT.root.TestMNet05CodecBoundsAndComposition`
- Result: passed
- Test: `TestMNet05CodecBoundsAndComposition` in `net05_conformance_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMNet05CodecBoundsAndComposition$' ./.`
- Scope: Existing synthetic Go unit test TestMNet05CodecBoundsAndComposition in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0016: A step missing test_version, input or options is rejected, and nested step options keep the common 8,192-byte cap (8,193 bytes is task_limit_exceeded); zero or foreign parameters have no targets and are not proxy-compatible.
  - REQ-NET-05-0025: The codec rejects a profile over its own 49,152-byte cap.

<a id="UT.root.TestMNet05PrimitiveRiskAndMissingCapability"></a>

### UT.root.TestMNet05PrimitiveRiskAndMissingCapability

- Case: `UT.root.TestMNet05PrimitiveRiskAndMissingCapability`
- Result: passed
- Test: `TestMNet05PrimitiveRiskAndMissingCapability` in `net05_conformance_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestMNet05PrimitiveRiskAndMissingCapability$' ./.`
- Scope: Existing synthetic Go unit test TestMNet05PrimitiveRiskAndMissingCapability in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0065: A primitive with a high baseline risk makes the profile task_issuer_not_authorized under a medium-risk policy, and a missing primitive registry entry makes it task_unsupported, both with no replay commit.

<a id="UT.internal.execution.TestMNet05Limits"></a>

### UT.internal.execution.TestMNet05Limits

- Case: `UT.internal.execution.TestMNet05Limits`
- Result: passed
- Test: `TestMNet05Limits` in `internal/execution/net05_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestMNet05Limits$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestMNet05Limits in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0037: With three TCP-01 steps, connection limits 0, 1 and 2 stop the profile with resource_limit after that many dials and limits 3 and 4 complete with three, every connection closed.
  - REQ-LIMIT-0019: Aggregate connection limits 0..4 against a three-TCP-step profile allow min(limit, 3) dials and end incomplete with resource_limit below 3.
  - REQ-NET-05-0026: Per-primitive declared maxima and declared sums of connections, received bytes and DNS-01 window durations are checked below, at and above the composite ceilings and stricter task bounds; undeclared sent bytes add nothing.
  - REQ-LIMIT-0020: A profile whose declared sums exceed the outer limits is rejected by Parameters.ValidateLimits before any plan runs.
  - REQ-LIMIT-0035: Each primitive-and-mode row declares its registered maxima (DNS-01 windows and received bytes, one connection for direct TCP DNS-01, TCP-01, TLS-01 and each permitted HTTP-01 request, HTTP-01 header bytes plus `max_body_bytes`); the per-resource sums are checked against the smaller of the composite limit and the task's `execution.limits`, and undeclared sent bytes add nothing.

<a id="UT.internal.execution.TestMNet05AggregateBytes"></a>

### UT.internal.execution.TestMNet05AggregateBytes

- Case: `UT.internal.execution.TestMNet05AggregateBytes`
- Result: passed
- Test: `TestMNet05AggregateBytes` in `internal/execution/net05_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestMNet05AggregateBytes$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestMNet05AggregateBytes in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0037: On the direct and real CONNECT paths, two HTTP-01 steps accumulate plaintext bytes in one meter; sent or received limits at measured usage -1 interrupt the second step with resource_limit, at equal or +1 complete, usage never exceeds the limit, and no timer leaks.
  - REQ-LIMIT-0034: HTTP plaintext bytes accumulate across two composite steps on direct and proxy paths; aggregate sent or received limits equal to the usage complete and one byte less ends with resource_limit after the first step.

<a id="UT.internal.execution.TestMNet05AggregateDeadline"></a>

### UT.internal.execution.TestMNet05AggregateDeadline

- Case: `UT.internal.execution.TestMNet05AggregateDeadline`
- Result: passed
- Test: `TestMNet05AggregateDeadline` in `internal/execution/net05_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestMNet05AggregateDeadline$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestMNet05AggregateDeadline in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0037: Stricter task deadlines of 1 ms, 30 s and 60 s reached during the second step give attempt_deadline_exceeded with two dials, two steps and no leaked timer.

<a id="UT.internal.execution.TestNet05WholeProfileSerialization"></a>

### UT.internal.execution.TestNet05WholeProfileSerialization

- Case: `UT.internal.execution.TestNet05WholeProfileSerialization`
- Result: passed
- Test: `TestNet05WholeProfileSerialization` in `internal/execution/net05_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestNet05WholeProfileSerialization$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestNet05WholeProfileSerialization in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0006: Primitive operations of one profile never overlap.
  - REQ-NET-05-0037: A second profile on the same runner waits without contact and is not_run when cancelled; after the first completes the runner is reused with six dials in total.

<a id="UT.root.TestNet05DirectDNSPrimitives"></a>

### UT.root.TestNet05DirectDNSPrimitives

- Case: `UT.root.TestNet05DirectDNSPrimitives`
- Result: passed
- Test: `TestNet05DirectDNSPrimitives` in `net05_runner_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestNet05DirectDNSPrimitives$' ./.`
- Scope: Existing synthetic Go unit test TestNet05DirectDNSPrimitives in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0006: A direct DNS-01 step over UDP and over TCP contacts only the registered resolver once and resolves.

<a id="UT.root.TestNet05TLSPrimitiveSuccess"></a>

### UT.root.TestNet05TLSPrimitiveSuccess

- Case: `UT.root.TestNet05TLSPrimitiveSuccess`
- Result: passed
- Test: `TestNet05TLSPrimitiveSuccess` in `net05_runner_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestNet05TLSPrimitiveSuccess$' ./.`
- Scope: Existing synthetic Go unit test TestNet05TLSPrimitiveSuccess in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-05-0006: A TLS-01 step without an ALPN offer completes handshake_succeeded with one dial.

<a id="UT.internal.execution.TestNet01GenerationChange"></a>

### UT.internal.execution.TestNet01GenerationChange

- Case: `UT.internal.execution.TestNet01GenerationChange`
- Result: passed
- Test: `TestNet01GenerationChange` in `internal/execution/net01_termination_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestNet01GenerationChange$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestNet01GenerationChange in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-01-0021: A platform generation change landing after the read returns is not merged into a snapshot, and no dialer is requested.
  - REQ-NET-01-0027: A generation change after the read and before the outcome is recorded gives incomplete network_changed with no measurement, a change after the outcome is recorded leaves the completed snapshot with the first generation's values, and the watch is released in both cases.

<a id="UT.internal.execution.TestNet01DeadlineReleasesRead"></a>

### UT.internal.execution.TestNet01DeadlineReleasesRead

- Case: `UT.internal.execution.TestNet01DeadlineReleasesRead`
- Result: passed
- Test: `TestNet01DeadlineReleasesRead` in `internal/execution/net01_termination_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestNet01DeadlineReleasesRead$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestNet01DeadlineReleasesRead in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-01-0021: Advancing the fake clock to the 30-second outer deadline ends a platform read that never answers as incomplete attempt_deadline_exceeded, with no timer, network watch or dialer request left.

<a id="UT.root.TestMNet03AdmissionTypesAndStatuses"></a>

### UT.root.TestMNet03AdmissionTypesAndStatuses

- Case: `UT.root.TestMNet03AdmissionTypesAndStatuses`
- Result: passed
- Test: `TestMNet03AdmissionTypesAndStatuses` in `net03_closeout_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMNet03AdmissionTypesAndStatuses$' .`
- Scope: Existing synthetic Go unit test TestMNet03AdmissionTypesAndStatuses in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-03-0009: A non-string url, null, boolean or array expected_status, numeric, 65-character or non-hex digest and a relative URL are task_schema_invalid with no contact.
  - REQ-NET-03-0010: A URL with an unregistered user-specific query is target_not_registered with no contact.
  - REQ-NET-03-0011: fe80::1, 0.0.0.0, IPv4-mapped IPv6 and 192.168.1.1 URLs are task_schema_invalid with no contact.
  - REQ-NET-03-0043: expected_status 200, 300, 304, 305, 306, 404 and 599 are admitted with no contact.

<a id="UT.root.TestMNet03OptionsDefaultAndUnsupportedFamily"></a>

### UT.root.TestMNet03OptionsDefaultAndUnsupportedFamily

- Case: `UT.root.TestMNet03OptionsDefaultAndUnsupportedFamily`
- Result: passed
- Test: `TestMNet03OptionsDefaultAndUnsupportedFamily` in `net03_closeout_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMNet03OptionsDefaultAndUnsupportedFamily$' .`
- Scope: Existing synthetic Go unit test TestMNet03OptionsDefaultAndUnsupportedFamily in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-03-0015: The omitted family normalizes to auto, and ipv6 on a network without IPv6 ends incomplete with constraints_unsatisfied, no measurement and no contact.

<a id="UT.root.TestMNet03ResponseClassBoundaries"></a>

### UT.root.TestMNet03ResponseClassBoundaries

- Case: `UT.root.TestMNet03ResponseClassBoundaries`
- Result: passed
- Test: `TestMNet03ResponseClassBoundaries` in `net03_closeout_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMNet03ResponseClassBoundaries$' .`
- Scope: Existing synthetic Go unit test TestMNet03ResponseClassBoundaries in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-03-0025: redirect, unexpected and expected are emitted for the boundary responses listed under REQ-NET-03-0041.
  - REQ-NET-03-0041: 301, 303 and 308 are redirect; 300 and 304 against expected 200, malformed chunked framing with a digest and a body-read timeout are unexpected; 304 against expected 304 and gzip without a digest are expected; no status line and a header block over 16,384 bytes have no class and a failed http_exchange; each run makes one dial and one request.
  - REQ-NET-03-0042: With unknown platform state, redirects select captive_portal_suspected, unexpected selects unexpected_response, expected selects validated and no class selects inconclusive; a body-read timeout under captive_portal still selects unexpected_response.

<a id="UT.root.TestMNet03PlatformReadFailure"></a>

### UT.root.TestMNet03PlatformReadFailure

- Case: `UT.root.TestMNet03PlatformReadFailure`
- Result: passed
- Test: `TestMNet03PlatformReadFailure` in `net03_closeout_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMNet03PlatformReadFailure$' .`
- Scope: Existing synthetic Go unit test TestMNet03PlatformReadFailure in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-03-0005: A failing platform read is recorded as platform_validation unknown and the expected response selects validated.
  - REQ-NET-03-0006: After a failing platform read the run still sends exactly one request.

<a id="UT.root.TestMNet03VersionShapes"></a>

### UT.root.TestMNet03VersionShapes

- Case: `UT.root.TestMNet03VersionShapes`
- Result: passed
- Test: `TestMNet03VersionShapes` in `net03_closeout_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMNet03VersionShapes$' .`
- Scope: Existing synthetic Go unit test TestMNet03VersionShapes in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-03-0022: Task versions 0.2.0 and 1.0.0 are task_unsupported and malformed or missing versions task_schema_invalid with no run or contact; the emitted result decodes and whole NET-01 access-only and NET-05 profile measurements are rejected.

<a id="UT.root.TestMNet03PrefixStopsReading"></a>

### UT.root.TestMNet03PrefixStopsReading

- Case: `UT.root.TestMNet03PrefixStopsReading`
- Result: passed
- Test: `TestMNet03PrefixStopsReading` in `net03_closeout_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMNet03PrefixStopsReading$' .`
- Scope: Existing synthetic Go unit test TestMNet03PrefixStopsReading in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-03-0021: With a 16,384-byte http body, a received budget of exactly the head plus 1,024 bytes completes with one request and one connection, and one byte less is incomplete resource_limit.
  - REQ-NET-03-0041: The completed run classifies the truncated body without a digest as expected with bytes_processed 1,024 and truncated true.

<a id="UT.root.TestMNet03Output"></a>

### UT.root.TestMNet03Output

- Case: `UT.root.TestMNet03Output`
- Result: passed
- Test: `TestMNet03Output` in `net03_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMNet03Output$' .`
- Scope: Existing synthetic Go unit test TestMNet03Output in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-03-0030: An expected response and five redirect location classes produce results containing none of the URL, address, location, body, digest, header name and value, cookie, raw-error and certificate canaries, with http_exchange details limited to registered keys.
  - REQ-NET-03-0036: The results retain no response body, redirect URL, cookie or custom header canary, and the redirect object has only followed_count and terminal_location_class.

<a id="UT.root.TestMNet03Execution"></a>

### UT.root.TestMNet03Execution

- Case: `UT.root.TestMNet03Execution`
- Result: passed
- Test: `TestMNet03Execution` in `net03_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMNet03Execution$' .`
- Scope: Existing synthetic Go unit test TestMNet03Execution in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-03-0038: The single validation GET carries no User-Agent, Cookie, Authorization, If-None-Match or If-Modified-Since header and no body.

<a id="UT.root.TestMNet03Options"></a>

### UT.root.TestMNet03Options

- Case: `UT.root.TestMNet03Options`
- Result: passed
- Test: `TestMNet03Options` in `net03_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMNet03Options$' .`
- Scope: Existing synthetic Go unit test TestMNet03Options in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-03-0040: A proxy network_path is task_schema_invalid and a max_bytes_received below the 1,024-byte prefix is rejected at admission rather than clamped.

<a id="UT.root.TestMDns03Normalization"></a>

### UT.root.TestMDns03Normalization

- Case: `UT.root.TestMDns03Normalization`
- Result: passed
- Test: `TestMDns03Normalization` in `dns03_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMDns03Normalization$' .`
- Scope: Existing synthetic Go unit test TestMDns03Normalization in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-03-0029: For every outcome row the serialized result contains none of the resolver, query, address, certificate, cookie, header, body, media-type or raw-error canaries.
  - REQ-DNS-03-0032: Each failure is reported with the terminal observation of the stage that failed, after ordered succeeded prior observations, with an outcome naming that stage's class.

<a id="UT.root.TestMDns03Admission"></a>

### UT.root.TestMDns03Admission

- Case: `UT.root.TestMDns03Admission`
- Result: passed
- Test: `TestMDns03Admission` in `dns03_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMDns03Admission$' .`
- Scope: Existing synthetic Go unit test TestMDns03Admission in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-03-0038: An unregistered resolver endpoint or query name is target_not_registered with no lookup or dial.

<a id="UT.root.TestMDns04Admission"></a>

### UT.root.TestMDns04Admission

- Case: `UT.root.TestMDns04Admission`
- Result: passed
- Test: `TestMDns04Admission` in `dns04_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMDns04Admission$' .`
- Scope: Existing synthetic Go unit test TestMDns04Admission in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-04-0038: Resolver endpoints and query names outside the local registration, including another port, bootstrap address, resolver or query name, are target_not_registered, and local, private and interpolated targets are task_schema_invalid, all before any lookup or contact.

<a id="UT.root.TestMDns04Normalization"></a>

### UT.root.TestMDns04Normalization

- Case: `UT.root.TestMDns04Normalization`
- Result: passed
- Test: `TestMDns04Normalization` in `dns04_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMDns04Normalization$' .`
- Scope: Existing synthetic Go unit test TestMDns04Normalization in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-04-0029: Results for every outcome row contain none of the resolver name, query name, bootstrap address, port, field names, certificate canary, raw-error canary carried by injected errors, loopback address, alpn or ttl.
  - REQ-DNS-04-0035: Results for every outcome row keep only the outcome and ordered observations with 250-ms quantized durations and registered details, and pass bundle validation.

<a id="UT.root.TestMDns04NormalizedAdmission"></a>

### UT.root.TestMDns04NormalizedAdmission

- Case: `UT.root.TestMDns04NormalizedAdmission`
- Result: passed
- Test: `TestMDns04NormalizedAdmission` in `dns04_closeout_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMDns04NormalizedAdmission$' .`
- Scope: Existing synthetic Go unit test TestMDns04NormalizedAdmission in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-04-0004: A task naming the registered resolver and query name in other letter case is admitted against the canonical registration without lookup or contact.
  - REQ-DNS-04-0009: Case-folded resolver and query names and a Unicode query name normalize to the registered lowercase and A-label forms.
  - REQ-DNS-04-0010: The normalized resolver profile and query name match the registered profile exactly, so admission succeeds with no contact.

<a id="UT.root.TestMDns04ExchangeFaults"></a>

### UT.root.TestMDns04ExchangeFaults

- Case: `UT.root.TestMDns04ExchangeFaults`
- Result: passed
- Test: `TestMDns04ExchangeFaults` in `dns04_closeout_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMDns04ExchangeFaults$' .`
- Scope: Existing synthetic Go unit test TestMDns04ExchangeFaults in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-04-0005: Under prefix, message, query-mismatch, reset and TLS faults the attempt makes one dial and one accepted connection to the bootstrap address, with no lookup and no other destination.
  - REQ-DNS-04-0006: After each fault the server has received at most the 33-byte framed query and nothing further, so the session closes without retry or cleartext fallback.
  - REQ-DNS-04-0008: A stream closed inside the length prefix, inside the message or right after the query, a prefix announcing more than is sent, and an 11-byte message are malformed_dns_response; a mismatched ID is query_mismatch; a reset after the query is inconclusive; an untrusted certificate is tls_failure.
  - REQ-DNS-04-0028: A length prefix beyond the stream yields a completed malformed_dns_response with a measurement rather than an incomplete state.

<a id="UT.root.TestMDns04CrossTestOutputs"></a>

### UT.root.TestMDns04CrossTestOutputs

- Case: `UT.root.TestMDns04CrossTestOutputs`
- Result: passed
- Test: `TestMDns04CrossTestOutputs` in `dns04_closeout_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMDns04CrossTestOutputs$' .`
- Scope: Existing synthetic Go unit test TestMDns04CrossTestOutputs in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-04-0024: Results produced by the DNS-03 and DNS-01 runners decode under their own 0.1.0 bindings and are rejected under the DNS-04 0.1.0 binding, both with their own test.id and relabelled as DNS-04.

<a id="UT.internal.protocol.http.TestPOSTStatusPrecedesMediaType"></a>

### UT.internal.protocol.http.TestPOSTStatusPrecedesMediaType

- Case: `UT.internal.protocol.http.TestPOSTStatusPrecedesMediaType`
- Result: passed
- Test: `TestPOSTStatusPrecedesMediaType` in `internal/protocol/http/exchange_test.go` (package `internal/protocol/http`)
- Reproduce one case: `go test -count=1 -run '^TestPOSTStatusPrecedesMediaType$' ./internal/protocol/http`
- Scope: Existing synthetic Go unit test TestPOSTStatusPrecedesMediaType in package internal/protocol/http, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-03-0040: For a DNS POST over one scripted connection, a 503 HTML response, a 404 declaring `application/dns-message`, a 302 with `Location` and a 500 with `Content-Encoding: gzip` end at their status: the status is retained, no content encoding or media type is checked, no body is read, no `Location` is classified and no second connection opens. A 200 HTML response is `unexpected_content_type`. The DNS-03 outcome mapping is exercised by M-DNS-03-NORMALIZATION.

## Cross-Cutting Conformance Family Cases

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
  - REQ-TRUST-0032: Asserts that policy-shaped content in permitted task labels is admitted as opaque text and that policy-shaped unknown task members are rejected as `task_schema_invalid`, with the active generation, issuers and registrations unchanged in both cases. Proxy and context endpoint configuration exist but are not exercised by this case; collector and target paths do not exist.

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

<a id="CC-SCHEDULE-ENVELOPE"></a>

### CC-SCHEDULE-ENVELOPE

- Case: `CC-SCHEDULE-ENVELOPE`
- Result: passed
- Tests: `TestCcScheduleEnvelope` in `runtime_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestCcScheduleEnvelope$' ./.`
- Scope: Synthetic core contract cases; Drive the scheduler with a deterministic clock and a fake platform that can report the application as suspended. Admit a NET-01 task with `starts_at` T and `not_after` T+24h. Keep the fake platform suspended from T-1h to T+2h and assert no attempt starts during suspension and the run is not marked failed or missed at T. On resume at T+2h assert the run becomes eligible and its attempt starts after T+2h, and assert the result `started_at` reflects the actual start, not `starts_at`. Repeat with suspension lasting until after `not_after` and assert no run or attempt starts after resume. Assert no API return value or status describes `starts_at` as a guaranteed or exact execution time. Requirements: REQ-SCHEDULE-0001.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0001: With the synthetic platform reporting suspension from T-1h to T+2h, `StartRun` returns `run_not_due` at T, T+1s and T+1h+1s with no attempt, no result and no counted run, and after resume at T+2h the run starts, completes with one counted run, `RunState` reports `terminal`, and the result `started_at` is the actual start (T+2h, minute precision) rather than `starts_at`; a suspension lasting past `not_after` leaves `run_not_due` and no attempt after resume, an attempt in flight at suspension ends `incomplete` with `platform_suspended`, and an unknown suspension state does not block the run. The vehicle is the synthetic TCP-01 runner rather than NET-01, and the test does not scan return values for `starts_at` beyond checking these fixed codes and run states.

<a id="CC-SCHEDULE-FIELDS"></a>

### CC-SCHEDULE-FIELDS

- Case: `CC-SCHEDULE-FIELDS`
- Result: passed
- Test: `TestCcScheduleFields` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcScheduleFields$' .`
- Scope: Synthetic core contract cases; Exercise omitted/present and conditional schedule fields, defaults, wrong types and unknown properties.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0002: Asserts through `Client.AdmitTask` and `Client.StartRun` on the injected clock that a schedule with only `not_after` is admitted with one run and one attempt, that every optional field can be present, that omitting `not_after` is rejected for an embedded task with two runs, a remote and a repeated task, and that wrong types and unknown members are `task_schema_invalid` with no replay record.

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
  - REQ-SCHEDULE-0011: Asserts for signed remote packages that `not_after` at acceptance is `task_expired`, that one week after the earlier of issuance and acceptance is admitted and one second more is `task_limit_exceeded`, and that an embedded one-time task without `not_after` is admitted through the facade under the exception of REQ-LIMIT-0013, runs once, yields no second run on redelivery or an hour later, and keeps a replay record that expires 604,800 seconds and is retained 691,200 seconds after acceptance, while an embedded task with two runs, a repeating embedded task and a remote task without `not_after` are `task_schema_invalid` and `validation.ScheduleAt` honors the exception only for one run of an embedded task.

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

<a id="CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-EQUAL-ENDPOINTS"></a>

### CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-EQUAL-ENDPOINTS

- Case: `CC-SCHEDULE-SCHEDULING-MODEL-TIME-WINDOW-EQUAL-ENDPOINTS`
- Result: passed
- Test: `TestCcScheduleSchedulingModelTimeWindowEqualEndpoints` in `internal/execution/cross_cutting_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcScheduleSchedulingModelTimeWindowEqualEndpoints$' ./internal/execution`
- Scope: Synthetic core contract cases; Admit a time_window whose only range has equal start and end, such as 12:00 to 12:00, and drive a deterministic clock across a full day: no run start or retry is permitted at any minute. Add a nonempty range beside the empty one and assert that exactly its minutes permit starts. Submit two identical empty ranges and assert rejection as duplicates (REQ-SCHEDULE-0044).
- Requirement aspects exercised:

  - REQ-SCHEDULE-0044: Asserts through `validation.Common` that a window whose only range is 12:00 to 12:00 is valid, and through `validation.RunEligible` and `validation.RetryEligible` at every minute of a day that it permits no run start and no retry, while without a window every minute permits both; that beside a 14:00 to 16:00 range exactly that range's minutes permit a run start and a retry; and that two identical empty ranges are rejected as duplicates while two distinct empty ranges are valid.

<a id="CC-SCHEDULE-SCHEDULING-MODEL-REPEAT"></a>

### CC-SCHEDULE-SCHEDULING-MODEL-REPEAT

- Case: `CC-SCHEDULE-SCHEDULING-MODEL-REPEAT`
- Result: passed
- Test: `TestCcScheduleSchedulingModelRepeat` in `runtime_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcScheduleSchedulingModelRepeat$' .`
- Scope: Synthetic core contract cases; Distinguish none and interval, disallow interval-only fields for none and generate the bounded number of runs.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0018: Asserts through `Client.AdmitTask` and `Client.StartRun` on the injected clock that omitted and `none` repeat create one run over six host calls, that `interval` creates exactly `max_runs` runs (1, 2 and 5) over ten calls, each writing one result, with distinct run identities among the results the pending store still retains, and that interval fields with `none`, an unregistered or missing mode and unknown repeat members are `task_schema_invalid`.

<a id="CC-SCHEDULE-SCHEDULING-MODEL-REPEAT-INTERVAL-FIELDS"></a>

### CC-SCHEDULE-SCHEDULING-MODEL-REPEAT-INTERVAL-FIELDS

- Case: `CC-SCHEDULE-SCHEDULING-MODEL-REPEAT-INTERVAL-FIELDS`
- Result: passed
- Test: `TestCcScheduleSchedulingModelRepeatIntervalFields` in `internal/execution/cross_cutting_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcScheduleSchedulingModelRepeatIntervalFields$' ./internal/execution`
- Scope: Synthetic core contract cases; Exercise interval/jitter bounds and delayed due runs with injected random offsets; offsets are locally generated per run and stay inside start/expiry/window constraints. Due times follow the anchored grid of REQ-SCHEDULE-0045: occurrence 0 at the anchor without jitter, a window-delayed occurrence current until the next, skipped occurrences that consume no run, no start at or after not_after, and one run per occurrence.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0019: Asserts through `validation.RunEligible` that interval 21,600 and 604,800 seconds with jitter 300 and 1,800 seconds are valid, and that interval 21,599 or 604,801 seconds and jitter 299 or 1,801 seconds make the schedule invalid, so no run is eligible.
  - REQ-SCHEDULE-0020: Asserts that a repeated run becomes due on the interval grid and may start later: an occurrence delayed by a time window starts when the window opens before the next occurrence becomes current.
  - REQ-SCHEDULE-0021: Asserts with injected offsets of 0, 150 seconds, 300 seconds and one hour that occurrence 1 is due at the anchor plus the interval plus the offset capped at the task's 300-second `jitter`, not one second before.
  - REQ-SCHEDULE-0034: Asserts that a due occurrence starts only while the start envelope holds: not before `starts_at`, not outside the time window, and not at or after `not_after`.
  - REQ-SCHEDULE-0045: Asserts the anchored grid: occurrence 0 starts at `starts_at` without jitter even with a five-minute jitter source; a second start in the same occurrence is refused; an occurrence the window keeps from starting before the next becomes current is skipped, consuming no run, and the next starts instead; an occurrence due at `not_after` never starts; `max_runs` 3 yields exactly three runs across five occurrences; and an occurrence started one second before the next becomes current leaves only the next occurrence's injected 10-second offset between the two runs.

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
- Scope: Synthetic core contract cases; Go-owned rows only. Seed a storage scope with pending bundles, run accounting, replay records and a synthetic linkage key, and fault-inject each clearing step before and after it takes effect. Assert an explicit incomplete error for every failure, a durable clearing marker that persists across a simulated restart while any deletion step may have run, idempotent recovery clearing from every partial state that leaves every record class and the key deleted, and that clearing derives no replacement key and has no transport. Cancellation of active attempts with bounded cleanup is asserted by the root UT cases. Host user-interface and operating-system reset wiring, native keystores, reinstall sentinels and host-owned copies are planned in CC-LINKAGE-CLEAR-DATA-CONTRACT-HOST. Requirements: REQ-LINKAGE-0023, REQ-LINKAGE-0024.
- Requirement aspects exercised:

  - REQ-LINKAGE-0023: Asserts with `persistence.Clear`, over a scope seeded with a replay record, its run record and a pending record and with a synthetic linkage key, that each of the 13 record-port steps (the marker write; the listing, deletion and confirming listing of pending, run and replay records; and the listing, deletion and confirming listing of the marker class), failed before and after it takes effect, returns `ErrClearIncomplete` and never success (item 7); that after a simulated restart `Clearing` reports the marker exactly while it is stored and no partial clear leaves the scope unmarked; that a later clear deletes the pending, run and replay records and the linkage key (items 3 to 5); and that a failed key deletion keeps the scope marked until a retry deletes the key. Active attempts, caches, epochs, sentinels and in-memory linkage material are not exercised by this case.
  - REQ-LINKAGE-0024: Asserts that after every injected failure two consecutive clears both complete and leave the scope empty, and that `Clear` takes only the record port and erasers, so no key derivation or transport is reachable from it; the absence of network packages in its dependency closure is shown by `TestClearingHasNoNetworkDependency`, not by observed traffic.

<a id="CC-LINKAGE-STORAGE-FAILURE"></a>

### CC-LINKAGE-STORAGE-FAILURE

- Case: `CC-LINKAGE-STORAGE-FAILURE`
- Result: passed
- Tests: `TestCcLinkageStorageFailure` in `linkage_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestCcLinkageStorageFailure$' ./.`
- Scope: Synthetic core contract cases; Inject unavailable/locked/inconsistent/invalid storage; omit linkage with permitted bundle grouping, preserve deletion, forbid weak persistence/host-visible/sticky-token fallback and sanitize diagnostics.
- Requirement aspects exercised:

  - REQ-LINKAGE-0037: For unavailable, locked, inconsistent and invalid key stores, each bundle after the failure omits `linkage`, carries exactly three top-level members and one result, does not contain the earlier token, and successive bundles have distinct `bundle_id` values; once storage recovers, linkage returns.
  - REQ-LINKAGE-0038: The record-port dump contains neither the earlier token nor any key-store canary, no bundle member replaces the omitted linkage, the earlier token is never reused, and with a key store that also fails deletion `ClearMeasurementData` returns an error yet leaves no pending result. Asking Flutter or the host to persist the secret is not exercised in this test.
  - REQ-LINKAGE-0039: No `linkage_unavailable` diagnostic exists in this build; the test asserts that the key-alias, platform-account and failed-derivation-byte canaries from key-store errors appear in no `Client` value or error and not in the record port.

<a id="CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS"></a>

### CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS

- Case: `CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS`
- Result: passed
- Tests: `TestCcLinkagePrivacyAndSafetyRequirements` in `linkage_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestCcLinkagePrivacyAndSafetyRequirements$' ./.`
- Scope: Synthetic core contract cases; Use a fake LinkageKeyStore holding a synthetic secret canary, a deterministic clock, and recording fakes for logs, diagnostics, crash/analytics hooks, the Network Context endpoints, measurement targets and the local proxy. Assert initialization without result_linkage emits no linkage and creates no key, and that tasks carrying linkage-enabling fields cannot enable it. With rotating_installation enabled, assert the secret canary never appears in any API return value, error, log, diagnostic or hook output, and the token appears only in bundle linkage.token. Assert Network Context requests contain no bundle_id, scope, epoch or token, and target and proxy traffic (including CONNECT requests) contains no linkage material. Assert task admission and scheduling decisions are identical for linkage disabled and for two different tokens. Assert the previous epoch's token is not retained in memory after rotation and that clearMeasurementData removes cached tokens so the next bundle in the same epoch carries a new token. Requirements: REQ-LINKAGE-0043, REQ-LINKAGE-0044, REQ-LINKAGE-0045, REQ-LINKAGE-0046, REQ-LINKAGE-0047, REQ-LINKAGE-0048, REQ-LINKAGE-0049, REQ-LINKAGE-0050, REQ-LINKAGE-0051, REQ-LINKAGE-0052.
- Requirement aspects exercised:

  - REQ-LINKAGE-0043: With an empty configuration, the bundle carries no linkage, no key is held or created and no derivation occurs.
  - REQ-LINKAGE-0044: A task carrying a `result_linkage` member is rejected at admission, and with linkage absent from the local configuration no linkage is emitted.
  - REQ-LINKAGE-0045: Not asserted: the test does not check how the token is named or described.
  - REQ-LINKAGE-0046: With a key store holding a secret canary, no form of the secret appears in any `Client` value or error, bundle, stored record, context request, `CONNECT` request, target byte, runner input or captured standard `log` output, the token appears only once per bundle as `linkage.token` and in no `Client` value or error, and a static scan of production Go files finds no logging, printing, crash-output or diagnostic sink.
  - REQ-LINKAGE-0047: Admission codes, start codes, run states and result states for valid, redelivered, unsupported, high-risk and repeating tasks are identical with linkage disabled and under two different tokens, and admission performs no derivation. Other prohibited uses are not asserted.
  - REQ-LINKAGE-0048: The raw requests received by the primary and fallback Network Context endpoints contain no deployment scope, epoch, `bundle_id`, token or secret.
  - REQ-LINKAGE-0049: The requests recorded by the proxy, including `CONNECT` requests, the bytes received by the loopback measurement target and the runner inputs contain no deployment scope, epoch, `bundle_id`, token or secret.
  - REQ-LINKAGE-0050: Partially: `linkage.Source` has no field type that can hold a token, and after rotation the previous epoch's token is absent from stored records; no bound tied to a documented research need is checked.
  - REQ-LINKAGE-0051: Not asserted: the test does not check combination of the token with timestamps, addresses, ASN and region data or host account data.
  - REQ-LINKAGE-0052: Only the library side: after `ClearMeasurementData` and reinitialization the next bundle in the same epoch carries a new token; the host's obligation to provide a reset path is not asserted.

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

<a id="CC-EXECUTION-FIELDS-NETWORK-FIELDS"></a>

### CC-EXECUTION-FIELDS-NETWORK-FIELDS

- Case: `CC-EXECUTION-FIELDS-NETWORK-FIELDS`
- Result: passed
- Tests: `TestCcExecutionFieldsNetworkFields` in `internal/execution/cross_cutting_conformance_test.go` (package `internal/execution`); `TestCcExecutionFieldsNetworkFieldsNetworkContextRequirements` in `internal/execution/egress_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcExecutionFieldsNetworkFields$' ./internal/execution`
- Scope: Synthetic core contract cases; Iterate allowed_transports [wifi], [cellular] and [wifi, cellular] against fake-platform network_type wifi, cellular, ethernet, other and unknown, and assert the task runs only when the reported type is in the list. Reject a task whose allowed_transports contains a value outside the initial list (wifi, cellular). Iterate vpn_policy any, exclude and require (and the omitted default any) against vpn_state detected, not_detected and unknown and assert run or do-not-run exactly per the public VPN eligibility table. Exercise require_validated true with validated yes, no and unknown, and assert execution only on the platform's validated yes, with no library connectivity check sent for any value (REQ-EXECUTION-0033). Exercise context required_fields and allowed_country_codes with present, missing and non-matching values. Assert every do-not-run case issues no network operation. Requirements: REQ-EXECUTION-0033, REQ-EXECUTION-0017.
- Requirement aspects exercised:

  - REQ-EXECUTION-0033: Asserts on the engine with the test-only fake platform network that `allowed_transports` lists [wifi], [cellular] and [wifi, cellular] run only when the reported network type is listed, across wifi, cellular, ethernet, other and unknown; that every `vpn_policy` value and the omitted default run exactly per the public VPN table; and that `require_validated: true` runs only when the platform reports validated yes and does not run for no or unknown. A run records `completed` with one runner call and one dialer issued; every do-not-run case records `not_run` with `constraints_unsatisfied`, no runner call and no dialer issued, so no library connectivity check or other network operation occurs. Its Network Context rows, required fields and the country allowlist on the attempt's own path, run in `TestCcExecutionFieldsNetworkFieldsNetworkContextRequirements`.
  - REQ-EXECUTION-0017: Asserts that `allowed_transports` values outside the initial list (`ethernet`, `other`, `unknown`, `WIFI` and a list containing `bluetooth`) are rejected at task decoding and validation, and that `wifi` and `cellular` are the only transports that can satisfy the condition.

<a id="CC-EXECUTION-FIELDS-NETWORK-FIELDS-VPN-POLICY"></a>

### CC-EXECUTION-FIELDS-NETWORK-FIELDS-VPN-POLICY

- Case: `CC-EXECUTION-FIELDS-NETWORK-FIELDS-VPN-POLICY`
- Result: passed
- Test: `TestCcExecutionFieldsNetworkFieldsVpnPolicy` in `internal/execution/cross_cutting_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcExecutionFieldsNetworkFieldsVpnPolicy$' ./internal/execution`
- Scope: Synthetic core contract cases; Use a fake platform that reports VPN state as detected, not_detected or unknown and evaluate each vpn_policy value (any, exclude, require) for all nine combinations. Expected: any runs for detected, not_detected and unknown; exclude runs only for not_detected and does not run for detected or unknown; require runs only for detected and does not run for not_detected or unknown. Assert omitted vpn_policy behaves as any. Assert the unknown state is carried through evaluation as unknown and is never mapped to not_detected, so exclude with unknown does not run. Reject an unregistered vpn_policy value (for example prefer) at validation. Requirements: REQ-EXECUTION-0018, REQ-EXECUTION-0019, REQ-EXECUTION-0020.
- Requirement aspects exercised:

  - REQ-EXECUTION-0018: Asserts that the registered values `any`, `exclude` and `require` decode through task decoding, that an omitted `vpn_policy` and an empty `network` object decode as `any`, and that the unregistered values `prefer`, `ANY`, the empty string and `none` are rejected at validation.
  - REQ-EXECUTION-0019: Asserts that the platform's unknown VPN state normalizes to `unknown`, never `not_detected`, and that `exclude` with `unknown` is unsatisfied in `execution.NetworkConditions` and does not run on the engine.
  - REQ-EXECUTION-0020: Asserts on the engine, for all nine combinations of `any`, `exclude` and `require` with `detected`, `not_detected` and `unknown` reported by the fake platform, that the attempt runs (`completed`, one dialer issued) exactly where the public table says run, and otherwise records `not_run` with `constraints_unsatisfied` with no dialer issued and no runner call.

<a id="CC-EXECUTION-FIELDS-NETWORK-FIELDS-NETWORK-CONTEXT-REQUIREMENTS"></a>

### CC-EXECUTION-FIELDS-NETWORK-FIELDS-NETWORK-CONTEXT-REQUIREMENTS

- Case: `CC-EXECUTION-FIELDS-NETWORK-FIELDS-NETWORK-CONTEXT-REQUIREMENTS`
- Result: passed
- Tests: `TestCcExecutionFieldsNetworkFieldsNetworkContextRequirements` in `internal/execution/egress_test.go` (package `internal/execution`); `TestCcContextExecutionIntegration` in `internal/execution/egress_test.go` (package `internal/execution`); `TestLocalPolicyContextEligibility` in `local_restrictions_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestCcExecutionFieldsNetworkFieldsNetworkContextRequirements$' ./internal/execution`
- Scope: Synthetic core contract cases; Exercise required_fields/allowlists, missing/invalid/unknown context, optional context failure and stricter policy. Count requests to prove no endpoint/path override by tasks.
- Requirement aspects exercised:

  - REQ-EXECUTION-0021: The task decoder accepts `network.context` with `required_fields` `country_code`, `region_code` and `asn` and `allowed_country_codes` `NL` and `DE`, and rejects an unregistered required field (`city`) and a lowercase country code (`nl`).
  - REQ-EXECUTION-0022: The task schema rejects a task-supplied `primary_endpoint`, and an attempt whose task sets `required_fields` and `allowed_country_codes` dials only the two configured endpoint names, each once, primary then fallback. In `TestLocalPolicyContextEligibility`, a proxy-path task under a local policy that permits context acquisition only on the direct path records egress `unavailable` with `network_path` `proxy` and sends no `CONNECT` to the context endpoint.
  - REQ-EXECUTION-0023: `ContextConditions` returns `network_context_unavailable` for a required field missing from a fresh or cached context, for an unregistered required field, and for a country allowlist against an unavailable context, and `constraints_unsatisfied` for a country outside the allowlist; the schema rejects region, ASN, provider and operator allowlists. End to end, a required field with both endpoints answering 503 and an allowlist against a classification without a country end `not_run` with `network_context_unavailable` without running the test, and in `TestLocalPolicyContextEligibility` a local policy that forbids acquisition or narrows it to a path or endpoints that cannot answer turns required-field and allowlist tasks into `network_context_unavailable`.
  - REQ-EXECUTION-0024: `ContextConditions` imposes no condition when the task has no context requirements and egress is unavailable, end to end a task without requirements completes with the runner called and `egress.status` `unavailable` when both endpoints answer 503, and in `TestLocalPolicyContextEligibility` it completes with `egress.status` `unavailable` when local policy forbids or narrows acquisition.

<a id="CC-EXECUTION-FIELDS-DEVICE-FIELDS"></a>

### CC-EXECUTION-FIELDS-DEVICE-FIELDS

- Case: `CC-EXECUTION-FIELDS-DEVICE-FIELDS`
- Result: passed
- Tests: `TestCcExecutionFieldsDeviceFields` in `local_restrictions_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestCcExecutionFieldsDeviceFields$' ./.`
- Scope: Synthetic core contract cases; Exercise battery/charging constraints using synthetic platform state, with local policy stricter than task values.
- Requirement aspects exercised:

  - REQ-EXECUTION-0025: Through `Client.StartRun` with synthetic device state, a task `minimum_battery_percent` of 20 runs at 20% and returns `run_not_due` at 19% with no attempt or result, and a task `allow_charging_only` runs while charging; a task without device constraints never queries the device capability.
  - REQ-EXECUTION-0026: Local device constraints are enforced in addition to the task's: a local minimum applies without a task value, the higher of the local and task minimums wins, either side's charging requirement applies, unknown, unavailable or absent device state never satisfies a local constraint, and a local minimum that becomes unmet after a retryable first attempt prevents the retry.

<a id="CC-CONTEXT-CONTEXT-LAYERS-ACCESS-CONTEXT"></a>

### CC-CONTEXT-CONTEXT-LAYERS-ACCESS-CONTEXT

- Case: `CC-CONTEXT-CONTEXT-LAYERS-ACCESS-CONTEXT`
- Result: passed
- Test: `TestCcContextContextLayersAccessContext` in `internal/networkcontext/cross_cutting_conformance_test.go` (package `internal/networkcontext`)
- Reproduce one case: `go test -count=1 -run '^TestCcContextContextLayersAccessContext$' ./internal/networkcontext`
- Scope: Synthetic core contract cases; Feed the Go core access context through a fake platform that also offers canary values for every forbidden item: interface name, local IPv4 and IPv6 addresses, DNS server address, SSID, BSSID, cell identifier, SIM information, subscriber identifier, carrier account information and physical location. Iterate `network_type` over `wifi`, `cellular`, `ethernet`, `other` and `unknown`, `vpn_state` over `detected`, `not_detected` and `unknown`, `ipv4_available`, `ipv6_available` and `validated` over `yes`, `no` and `unknown`, and `network_changed` over `true` and `false`. Assert the serialized access object contains exactly these six fields with the supplied normalized values. Assert serialized output of the access object, the Measurement Result, logs, cache entries and egress-endpoint requests contains none of the canary values. Reject an access input carrying any additional field. Requirements: REQ-CONTEXT-0002, REQ-CONTEXT-0003.
- Requirement aspects exercised:

  - REQ-CONTEXT-0002: Asserts that `platform.AccessState` has only the network type, VPN state, IPv4 and IPv6 availability, validation state and generation fields plus the captive-portal flag, which only NET-03's platform validation reads, and that for every value of each field, with `network_changed` false and true, the serialized access object of a captured direct-path snapshot holds exactly the six registered fields with the normalized values the fake platform reported.
  - REQ-CONTEXT-0003: Asserts that none of the synthetic canaries the fake platform holds beside its reading (interface name, local IPv4 and IPv6 addresses, DNS server address, SSID, BSSID, cell identifier, SIM information, subscriber identifier, carrier account and location) appears in the encoded access object or in a Measurement Result built with it, and that access input carrying an interface, local address, SSID or location field is rejected. The core writes no log. This case does not inspect the egress cache or the endpoint request: the cache holds only classifications keyed by generations and the path name, and `CC-CONTEXT-ENDPOINT-CONTRACT` asserts that the request carries only its request line, `Host` and `Accept`.

<a id="CC-CONTEXT-FORMAT-ACCESS-OBJECT"></a>

### CC-CONTEXT-FORMAT-ACCESS-OBJECT

- Case: `CC-CONTEXT-FORMAT-ACCESS-OBJECT`
- Result: passed
- Test: `TestCcContextFormatAccessObject` in `internal/networkcontext/cross_cutting_conformance_test.go` (package `internal/networkcontext`)
- Reproduce one case: `go test -count=1 -run '^TestCcContextFormatAccessObject$' ./internal/networkcontext`
- Scope: Synthetic core contract cases; Iterate every registered access value and accept each: network_type wifi, cellular, ethernet, other, unknown; vpn_state detected, not_detected, unknown; ipv4_available, ipv6_available and validated yes, no, unknown; network_changed true and false. Reject case variants (WIFI, Yes), unregistered strings, network_changed as the string false or as null, and objects missing each required field in turn. Configure the fake platform to expose canaries for interface names, addresses, Wi-Fi identifiers, DNS server addresses, cell information, SIM information, carrier names, subscriber identifiers and location, and assert none appears in the encoded access object or result; reject decoding of an access object carrying any extra field for those items. Requirements: REQ-CONTEXT-FORMAT-0004, REQ-CONTEXT-FORMAT-0005.
- Requirement aspects exercised:

  - REQ-CONTEXT-FORMAT-0004: Asserts that every registered value of `network_type`, `vpn_state`, `ipv4_available`, `ipv6_available`, `validated` and `network_changed` is accepted; that case variants, unregistered strings, wrong types, `network_changed` as the string `false`, as null or as 0, and an object missing each field in turn are rejected; and that platform values outside the registered enumerations normalize to `unknown`.
  - REQ-CONTEXT-FORMAT-0005: Asserts that an access object carrying an interface name, address, SSID, BSSID, DNS server, cell identifier, SIM, carrier, subscriber identifier or location field is rejected, and that no fake-platform canary and no generation number appears in a captured access object or in a Measurement Result built with it.

<a id="CC-EXECUTION-ENVELOPE"></a>

### CC-EXECUTION-ENVELOPE

- Case: `CC-EXECUTION-ENVELOPE`
- Result: passed
- Test: `TestCcExecutionEnvelope` in `internal/execution/cross_cutting_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcExecutionEnvelope$' ./internal/execution`
- Scope: Synthetic core contract cases; Recheck constraints immediately before attempts and preserve selected registered protocol semantics through direct/proxy transport.
- Requirement aspects exercised:

  - REQ-EXECUTION-0001: Asserts on the engine that the preconditions and the proxy path check are each evaluated immediately before every attempt: twice for a run whose first attempt ends `proxy_unavailable` and is retried after its 35-second backoff.
  - REQ-EXECUTION-0002: Asserts that the same synthetic runner against the same loopback target records an identical `measurement` object and identical metered usage on the direct path and through the scripted loopback CONNECT proxy, and that the direct result records `network_path: direct` and the proxied one `network_path: proxy` with its `connection_code`. Other result fields are not compared.

<a id="CC-PROXY-HTTP-CONNECT-BEHAVIOR"></a>

### CC-PROXY-HTTP-CONNECT-BEHAVIOR

- Case: `CC-PROXY-HTTP-CONNECT-BEHAVIOR`
- Result: passed
- Test: `TestCcProxyHttpConnectBehavior` in `internal/execution/cross_cutting_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcProxyHttpConnectBehavior$' ./internal/execution`
- Scope: Synthetic core contract cases; Capture synthetic CONNECT exchanges for explicit IP authorities, 2xx/non-2xx/malformed/timeouts, redirects and cancellation. Assert exact authority, no credentials/cleartext target GET, fresh tunnels and unchanged TLS semantics.
- Requirement aspects exercised:

  - REQ-PROXY-0017: Asserts that each tunnel to the scripted loopback proxy carries exactly one CONNECT request, with no `Authorization`, `Cookie` or `User-Agent` header and no `GET` or `POST` request; that HTTP/1.1 200, 204 and 299 responses establish a tunnel while a 302, a status line without a space, HTTP/2, a folded header, bare LF line ends and no response do not; that TLS inside the tunnel to a loopback TLS server sends the expected SNI and validates the certificate, rejecting a mismatched name; and, through its `TestLateConnect` subtest, that a CONNECT not answered within the attempt deadline is `proxy_connect_failed`.
  - REQ-PROXY-0018: Asserts that no cleartext target request is sent through the proxy and, through its `TestEveryStreamOpensItsOwnTunnel` subtest, that a second stream of one attempt opens its own CONNECT on the same endpoint.
  - REQ-PROXY-0043: Asserts that every tunnel of the case, to an explicit loopback IP address and port, sends exactly that address and port as the CONNECT authority. The test no longer dials a hostname authority; refusal of a hostname authority before proxy contact is asserted by `UT.internal.execution.TestConnectAuthority` and `UT.internal.execution.TestTargetGateProxyRefusesHostnameBeforeStartOrResolution`.
  - REQ-PROXY-0020: Asserts through `TestEveryStreamOpensItsOwnTunnel` that a later attempt never reuses an earlier tunnel, and that an HTTP/2 response does not establish a tunnel.

<a id="CC-PROXY-TASK-SELECTION"></a>

### CC-PROXY-TASK-SELECTION

- Case: `CC-PROXY-TASK-SELECTION`
- Result: passed
- Test: `TestCcProxyTaskSelection` in `internal/execution/cross_cutting_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcProxyTaskSelection$' ./internal/execution`
- Scope: Synthetic core contract cases; Exercise direct default and proxy preconditions; task cannot carry endpoint, category or proxy identity.
- Requirement aspects exercised:

  - REQ-PROXY-0021: Asserts that the strict task decoder rejects a task that carries a `proxy` object, `proxy_id`, `endpoint`, `connection_code`, `proxy_authorization` or `fallback` member beside `network_path: proxy`.
  - REQ-PROXY-0022: Asserts that an omitted `network_path` decodes as `direct`, that only `direct` and `proxy` are accepted (`socks`, `PROXY`, empty and null are rejected), and that a proxy attempt whose pre-attempt check reports `proxy_not_configured`, `proxy_not_permitted` or `proxy_path_unsupported` records `not_run` with that reason, with no context capture, proxy or target contact or runner call. The check is scripted here; the root package derives it from the configuration, local policy and registry rule.

<a id="CC-PROXY-RESULT-METADATA"></a>

### CC-PROXY-RESULT-METADATA

- Case: `CC-PROXY-RESULT-METADATA`
- Result: passed
- Test: `TestCcProxyResultMetadata` in `internal/execution/cross_cutting_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcProxyResultMetadata$' ./internal/execution`
- Scope: Synthetic core contract cases; Use a scripted transport and a fake local HTTP CONNECT proxy configured with synthetic canary values (host "127.0.0.1", a distinctive canary port, connection_code "outline_mobileproxy") to produce four attempts: direct success, direct failure, proxied success after a scripted 2xx CONNECT response, and proxied failures recorded as proxy_connect_failed (scripted non-2xx CONNECT response carrying canary response headers and a canary body) and proxy_unavailable (closed listener port). Assert every proxied result records network_context.egress.network_path "proxy" and connection_code "outline_mobileproxy", including when no network_context endpoints are configured and egress status is unavailable, and every direct result records network_path "direct" with connection_code absent. Assert the serialized results and bundle contain no proxy host or port field or value, no proxy generation value, no CONNECT request line or authority echo, none of the canary proxy response headers or body, and no Outline transport configuration canary passed to the host fake; assert no CONNECT status class field is present unless the pinned result schema defines one. Requirements: REQ-PROXY-0031, REQ-PROXY-0032, REQ-PROXY-0033.
- Requirement aspects exercised:

  - REQ-PROXY-0031: Asserts that a completed proxied attempt, a `proxy_connect_failed` attempt and a `proxy_unavailable` attempt each record egress `network_path: proxy` with the configured `connection_code`, with egress unavailable because no context endpoint is configured.
  - REQ-PROXY-0032: Asserts that a completed and a `socket_error` direct attempt record egress with exactly `status`, `lookup_source` and `network_path: direct`, and no `connection_code`.
  - REQ-PROXY-0033: Asserts that no proxied result contains the target address, the proxy port, generation or CONNECT material, a response header name, value or body canary of a refusing proxy, its `403` status or a `status_class` member.

<a id="CC-PROXY-PROXY-CONFIGURATION"></a>

### CC-PROXY-PROXY-CONFIGURATION

- Case: `CC-PROXY-PROXY-CONFIGURATION`
- Result: passed
- Test: `TestCcProxyProxyConfiguration` in `internal/execution/cross_cutting_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcProxyProxyConfiguration$' ./internal/execution`
- Scope: Synthetic core contract cases; Exercise loopback/localhost, port and connection_code boundaries and forbidden fields/transports; reject non-loopback endpoints.
- Requirement aspects exercised:

  - REQ-PROXY-0005: Asserts that the configuration decoder accepts `127.0.0.1`, `127.255.255.254`, `::1` and `localhost`, ports 1 and 65535 and approved connection codes of up to 64 lowercase characters including `.`, `-` and `_`, and rejects port 0, 65536, a string or fractional port, a host with a port or zone, and an empty, uppercase, over-long, unapproved or `@`-bearing code; and that the production dialer accepts the same endpoints.
  - REQ-PROXY-0006: Asserts that `username`, `password`, `type`, `tls`, `id`, `display_name`, `pac` and `chain` members and a list of proxies are rejected.
  - REQ-PROXY-0007: Asserts that `192.0.2.1`, `0.0.0.0`, `::`, `LOCALHOST`, a hostname and an empty host are rejected by the decoder and, through `TestProxyEndpointIsLoopback`, by `NewProxyDialers`, and that an invalid replacement through `Decoder.WithProxy` leaves the previous configuration unchanged.

<a id="CC-PROXY-CONFIGURATION-VALIDATION"></a>

### CC-PROXY-CONFIGURATION-VALIDATION

- Case: `CC-PROXY-CONFIGURATION-VALIDATION`
- Result: passed
- Test: `TestCcProxyConfigurationValidation` in `proxy_lifecycle_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcProxyConfigurationValidation$' .`
- Scope: Synthetic core contract cases; Attempt malformed replacement while an old proxy is active; preserve old state. Test category stability/privacy and internally derived compatibility.
- Requirement aspects exercised:

  - REQ-PROXY-0008: Asserts through `Client.SetProxy`, while a proxy run is active, that 18 invalid replacements (empty, `null`, `{}`, `[]`, trailing data, input above the configuration size bound, a non-loopback host, ports 0, 65536 and `1.0`, an uppercase, unapproved, 65-character or empty code, a password, a tests list, a duplicate host member and an egress endpoint member) each return `proxy_rejected`, leave the active configuration and generation unchanged and do not cancel the run; and that a valid replacement then installs a higher generation and cancels the run with `proxy_reconfigured`.
  - REQ-PROXY-0009: Asserts that an uppercase, unapproved, 65-character or empty `connection_code` is rejected, and that the approved category is unchanged across renewal to a new port.
  - REQ-PROXY-0010: Asserts that a replacement carrying a tests list is rejected, and that registry-derived compatibility still admits a proxy task after renewal, so the host declares no supported tests.

<a id="CC-PROXY-EXECUTION-STATES"></a>

### CC-PROXY-EXECUTION-STATES

- Case: `CC-PROXY-EXECUTION-STATES`
- Result: passed
- Test: `TestCcProxyExecutionStates` in `internal/execution/cross_cutting_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcProxyExecutionStates$' ./internal/execution`
- Scope: Synthetic core contract cases; Trigger each path failure and assert normalized state, no censorship attribution, no direct retry, and any new generation needs a separately scheduled attempt.
- Requirement aspects exercised:

  - REQ-PROXY-0028: Asserts `proxy_not_configured`, `proxy_not_permitted` and `proxy_path_unsupported` as `not_run` from a scripted pre-attempt check with no context capture or proxy contact; `proxy_unavailable` for a closed endpoint port; `proxy_connect_failed` for a non-2xx, HTTP/1.0, four-digit, SOCKS, malformed, oversized or absent CONNECT response and for a CONNECT not answered within the attempt deadline; and `proxy_reconfigured` when the run is cancelled with the reconfiguration cause during its CONNECT.
  - REQ-PROXY-0029: Asserts that each failure is recorded as the attempt's execution state and reason, and that the runner's `socket_error` outcome for the failed dial is not retained, so no failure is reported as a protocol observation.
  - REQ-PROXY-0030: Asserts that no failure dials the loopback target outside the proxy, that a retry after `proxy_unavailable` uses the same proxy snapshot and never the direct path, and that after `proxy_reconfigured` the cancelled run does not retry, while a separately started run uses the new endpoint.
  - REQ-PROXY-0046: Asserts that `proxy_path_unsupported` from the pre-attempt check records `not_run` with no context capture or proxy contact. The check is scripted: admission applies the same registered rules, so an admitted task does not fail it in this build. The per-dial gate refusal, `not_run` before the test starts and `incomplete` after, is asserted by `UT.internal.execution.TestCcProxyProductionTargetRefusal`, and admission rejection by `UT.internal.admission.TestGateProxyCompatibility`.

<a id="CC-EXECUTION-FIELDS-NETWORK"></a>

### CC-EXECUTION-FIELDS-NETWORK

- Case: `CC-EXECUTION-FIELDS-NETWORK`
- Result: passed
- Test: `TestCcExecutionFieldsNetwork` in `internal/execution/cross_cutting_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcExecutionFieldsNetwork$' ./internal/execution`
- Scope: Synthetic core contract cases; Configure the fake platform so required network information is unavailable: network_type unknown with allowed_transports [wifi, cellular], vpn_state unknown under vpn_policy exclude and require, and context required_fields [country_code] with the egress classification unavailable. Assert for each case that no network operation is issued on the recording transport and the attempt does not run (not_run with constraints_unsatisfied for each undeterminable network condition, network_context_unavailable only for the missing required context field), and that unknown is never treated as satisfying a constraint. Assert the same task runs when the fake platform reports the required values. Requirements: REQ-EXECUTION-0031.
- Requirement aspects exercised:

  - REQ-EXECUTION-0031: Asserts that `network_type` unknown with `allowed_transports`, `vpn_state` unknown under `exclude` and `require`, `validated` unknown with `require_validated`, and a failed access capability each keep the attempt from running, and that the same task runs once the fake platform reports the values; and that a required `country_code` with no lookup configured records `not_run` with `network_context_unavailable`, with no runner call and no dial.

<a id="CC-CONTEXT-FORMAT-USE-IN-EXECUTION"></a>

### CC-CONTEXT-FORMAT-USE-IN-EXECUTION

- Case: `CC-CONTEXT-FORMAT-USE-IN-EXECUTION`
- Result: passed
- Test: `TestCcExecutionFieldsNetworkFieldsNetworkContextRequirements` in `internal/execution/egress_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcExecutionFieldsNetworkFieldsNetworkContextRequirements$' ./internal/execution`
- Scope: Synthetic core contract cases; Use normalized required fields and country eligibility only; reject region/ASN/provider/operator task allowlists.
- Requirement aspects exercised:

  - REQ-CONTEXT-FORMAT-0013: Asserts that `required_fields` of `country_code`, `region_code` and `asn` and a country allowlist decode under their normalized names and are evaluated by `execution.ContextConditions` against the normalized egress object, and that an unknown field name (`city`) and a lowercase country code are rejected.
  - REQ-CONTEXT-FORMAT-0014: Asserts that `allowed_region_codes`, `allowed_asns`, `allowed_providers` and `allowed_operators` are rejected by the strict task decoder. Stricter local deployment eligibility is not provided by the local policy model and is not tested.

<a id="CC-CONTEXT-OVERVIEW"></a>

### CC-CONTEXT-OVERVIEW

- Case: `CC-CONTEXT-OVERVIEW`
- Result: passed
- Test: `TestCcContextOverview` in `internal/execution/egress_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcContextOverview$' ./internal/execution`
- Scope: Synthetic core contract cases; Run multiple tests on one generation/path; acquire shared bounded context rather than independently repeating per test.
- Requirement aspects exercised:

  - REQ-CONTEXT-0001: Asserts that three runs on one access-network generation and the direct path make one request to the primary synthetic HTTPS endpoint and record `lookup_source` `primary`, then `cache`, then `cache`.

<a id="CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT"></a>

### CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT

- Case: `CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT`
- Result: passed
- Test: `TestCcContextContextLayersEgressContext` in `internal/execution/egress_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcContextContextLayersEgressContext$' ./internal/execution`
- Scope: Synthetic core contract cases; Run synthetic attempts against a local synthetic classification endpoint that records the observed source address and returns `{"country_code": "NL", "region_code": "NL-NH", "asn": 64500, "classification_version": "0.1.0"}`. For a `direct` attempt, assert the egress object contains `status`, `lookup_source`, `network_path: direct`, the classification fields and `classification_version`, and no `connection_code`. For a `proxy` attempt through a scripted loopback CONNECT proxy with `connection_code` `outline_mobileproxy`, assert the lookup goes through the proxy and the egress object includes that `connection_code`. Configure the endpoint to additionally return a field carrying the observed source address and assert the response is rejected and egress reports `status: unavailable`. Assert the recorded source address never appears in serialized results, cache keys, logs or status output. Requirements: REQ-CONTEXT-0004, REQ-CONTEXT-0005, REQ-CONTEXT-0006.
- Requirement aspects exercised:

  - REQ-CONTEXT-0004: Asserts that the egress object of a direct and of a proxied attempt carries `status: fresh`, country `NL`, region `NL-NH`, ASN 64500 and classification version `0.1.0`, with `connection_code` only on the proxy path.
  - REQ-CONTEXT-0005: Asserts that a response carrying an `ip` member is rejected, leaving egress unavailable and the address absent from the result, and that the source address and port the synthetic endpoint observed for each request appear in no result.
  - REQ-CONTEXT-0006: Asserts that results carry only the coarse classification fields and no endpoint name, path, address, port or generation. That the values describe an IP classification rather than the user's location is an interpretation obligation that this case cannot establish; the review case `RV-CC-CONTEXT-CONTEXT-LAYERS-EGRESS-CONTEXT` stays open and non-gating.

<a id="CC-CONTEXT-HOST-INITIALIZATION"></a>

### CC-CONTEXT-HOST-INITIALIZATION

- Case: `CC-CONTEXT-HOST-INITIALIZATION`
- Result: passed
- Test: `TestCcContextHostInitialization` in `network_context_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcContextHostInitialization$' .`
- Scope: Synthetic core contract cases; Validate endpoint URLs/fields/independent approvals, absence fallback to access-only, immutable configuration and rejection of redirects/discovery/credentials/per-user URLs.
- Requirement aspects exercised:

  - REQ-CONTEXT-0007: Asserts through the root facade that `Initialize` with an approved primary endpoint yields fresh primary egress, and that an unapproved, `http`, credential-bearing, fragment-bearing or loopback endpoint and a fallback without a primary fail initialization.
  - REQ-CONTEXT-0008: Asserts that without `network_context` the result has egress unavailable and no endpoint request is made, and that a task requiring `country_code` then records `not_run` with `network_context_unavailable`, again with no request.
  - REQ-CONTEXT-0009: Asserts that no exported `Client` method name refers to context or endpoints, so an initialized client has no call that changes the configuration. Reinitialization with a fresh cache is covered by `UT.root.TestEgressCacheLifecycle`.
  - REQ-CONTEXT-0010: Asserts that an unknown `discovery` member, a fallback list and a second fallback endpoint fail initialization, and that a 302 redirect from the primary is not followed: the redirect target receives no request and egress is unavailable.

<a id="CC-CONTEXT-ENDPOINT-CONTRACT"></a>

### CC-CONTEXT-ENDPOINT-CONTRACT

- Case: `CC-CONTEXT-ENDPOINT-CONTRACT`
- Result: passed
- Tests: `TestCcContextEndpointContract` in `internal/execution/egress_test.go` (package `internal/execution`) and `TestCcContextEndpointContract` in `internal/networkcontext/egress_test.go` (package `internal/networkcontext`)
- Reproduce one case: `go test -count=1 -run '^TestCcContextEndpointContract$' ./internal/execution` and `go test -count=1 -run '^TestCcContextEndpointContract$' ./internal/networkcontext`
- Scope: Synthetic core contract cases; Capture one fixed HTTPS GET without identifier/target/category/cookie/auth data. Validate exact response schema, 4096-byte/5000ms limits, forbidden fields and successful unknown classifications.
- Requirement aspects exercised:

  - REQ-CONTEXT-0011: Asserts on the decrypted request bytes at the synthetic HTTPS endpoint that one attempt sends exactly one `GET /network-context HTTP/1.1` with `Host: 8.8.8.8` and `Accept: application/json` as its only header, and no request to the fallback.
  - REQ-CONTEXT-0012: Asserts that the request has no query and no `User-Agent`, `Accept-Encoding`, `Cookie` or `Authorization` header, and carries no task ID, test ID, `connection_code`, loopback address, or run or attempt marker.
  - REQ-CONTEXT-0013: Asserts in `internal/networkcontext` that ASN 0, above 32 bits, as a string or spelled `-0`, `64500.0` or `6.45e4`, an empty or three-letter country, a region inconsistent with the country, an empty version, a non-object and an empty body make the response invalid and egress unavailable, and that a response with only `classification_version` is valid.
  - REQ-CONTEXT-0014: Asserts in `internal/networkcontext` that responses carrying `ip`, `latitude`, `city`, `postal_code`, `isp` or `request_id` members, an escaped unknown name, or an equal or escaped-equivalent duplicate member, are invalid.
  - REQ-CONTEXT-0015: Asserts in `internal/networkcontext` that a 4,096-byte response is accepted and a 4,097-byte response rejected, and that on the engine clock an answer after 4,999 ms is accepted while one after 5,001 ms is rejected because the operation is stopped at 5,000 ms, with no operation timer left armed after either answer.
  - REQ-CONTEXT-0016: Asserts in `internal/networkcontext` that a version-only response yields fresh primary egress with no classification values. That such a response does not trigger the fallback is asserted by `CC-CONTEXT-ACQUISITION-AND-FALLBACK`.

<a id="CC-CONTEXT-ACQUISITION-AND-FALLBACK"></a>

### CC-CONTEXT-ACQUISITION-AND-FALLBACK

- Case: `CC-CONTEXT-ACQUISITION-AND-FALLBACK`
- Result: passed
- Test: `TestCcContextAcquisitionAndFallback` in `internal/networkcontext/egress_test.go` (package `internal/networkcontext`)
- Reproduce one case: `go test -count=1 -run '^TestCcContextAcquisitionAndFallback$' ./internal/networkcontext`
- Scope: Synthetic core contract cases; Count primary/fallback calls, test valid-but-empty primary, malformed/TLS/HTTP/timeout/size failures, exact cache-key hits/misses, freshness and every invalidation event. Verify no persistent cache.
- Requirement aspects exercised:

  - REQ-CONTEXT-0043: Asserts with a scripted fetch and eligible explicit-IP endpoints (`8.8.8.8` primary, `9.9.9.9` fallback) the bounded order: a fresh cache entry; otherwise one primary request; otherwise one fallback request after an unreachable, malformed, version-less, address-bearing, unapproved-country, unapproved-version, region-without-country or oversized primary; otherwise unavailable, with exactly one request per endpoint; and that without a fallback a failing primary makes one request. Endpoint eligibility before contact on the proxy path is asserted by `UT.internal.networkcontext.TestProxyContextHostnameEligibility`.
  - REQ-CONTEXT-0021: Asserts that a valid primary with no classification values is fresh primary egress without contacting the fallback, that no endpoint receives a second request within one acquisition, and that a failed acquisition is not cached, so the next acquisition asks again.
  - REQ-CONTEXT-0022: Asserts on the injected clock that an entry is reused at exactly 300 seconds and not one second later. Configurable freshness up to 900 seconds does not apply: this build's freshness is fixed at the default.
  - REQ-CONTEXT-0023: Asserts that each proxy generation and the direct path are separate keys; that `ForgetProxy` drops proxy entries and keeps the direct entry; that a new access-network generation is a new key and drops the older generation's entries; that `Clear` drops every entry and stops a lookup created before it from writing; and that closing the lookup drops its entries.
  - REQ-CONTEXT-0024: Asserts that entries are separate per access-network generation, path and proxy generation and are dropped by `Clear`, `ForgetProxy`, a newer access-network generation and closing the lookup. That the cache is in memory only, with no persistence port and no address or identifier in its key, follows from the `networkcontext.Cache` and `cacheKey` types and is not asserted by this case. Shutdown, reinitialization, policy replacement and clearing through the facade are asserted by `UT.root.TestEgressCacheLifecycle`.

<a id="CC-CONTEXT-EXECUTION-INTEGRATION"></a>

### CC-CONTEXT-EXECUTION-INTEGRATION

- Case: `CC-CONTEXT-EXECUTION-INTEGRATION`
- Result: passed
- Tests: `TestCcContextExecutionIntegration` in `internal/execution/egress_test.go` (package `internal/execution`)
- Reproduce the primary test: `go test -count=1 -run '^TestCcContextExecutionIntegration$' ./internal/execution`
- Scope: Synthetic core contract cases; Inject cancellation/deadline/suspension/network changes during acquisition; distinguish missing required context, allowlist mismatch and optional egress failure. Requirements: REQ-CONTEXT-0025, REQ-CONTEXT-0026, REQ-CONTEXT-0027, REQ-CONTEXT-0028.
- Requirement aspects exercised:

  - REQ-CONTEXT-0025: While the primary endpoint holds the context request, host cancellation, the attempt deadline, a network change and a platform suspension each end the attempt `not_run` with `cancelled_by_host`, `attempt_deadline_exceeded`, `network_changed` or `platform_suspended`, with the runner never called and the fallback endpoint not contacted; background and local safety policy are not exercised separately.
  - REQ-CONTEXT-0026: A required field absent from the classification, or both endpoints answering 503, ends the attempt `not_run` with `network_context_unavailable` without running the test, and when the schedule lists that reason the retry acquires again (a second primary request after one wait) and completes once the field is available.
  - REQ-CONTEXT-0027: With no context requirement and both endpoints answering 503, the attempt completes, the runner is called and `egress.status` is `unavailable`.
  - REQ-CONTEXT-0028: Missing required fields, failing endpoints and an allowlist against a classification without a country record `network_context_unavailable`, while an available country outside the allowlist ends the attempt `not_run` with `constraints_unsatisfied` and no test run; a mismatch on a retry ends the run with only the first attempt's result.

<a id="CC-CONTEXT-FORMAT-EGRESS-OBJECT"></a>

### CC-CONTEXT-FORMAT-EGRESS-OBJECT

- Case: `CC-CONTEXT-FORMAT-EGRESS-OBJECT`
- Result: passed
- Test: `TestCcContextFormatEgressObject` in `internal/networkcontext/egress_test.go` (package `internal/networkcontext`)
- Reproduce one case: `go test -count=1 -run '^TestCcContextFormatEgressObject$' ./internal/networkcontext`
- Scope: Synthetic core contract cases; Test all status/source/path/connection-code/classification combinations, missing optional classification, and no guessed values.
- Requirement aspects exercised:

  - REQ-CONTEXT-FORMAT-0006: Asserts that on the direct and the proxy path the serialized egress object holds exactly `status`, `lookup_source` and `network_path`, the classification fields the response supplied, and `connection_code` on the proxy path, and nothing else.
  - REQ-CONTEXT-FORMAT-0007: Asserts that when both endpoints fail the egress object is `unavailable` with `lookup_source: none`, no classification fields and the requested `network_path`, with `connection_code` still present on the proxy path.
  - REQ-CONTEXT-FORMAT-0008: Asserts `fresh` with `primary` and with `fallback`, and `cached` with `cache`.
  - REQ-CONTEXT-FORMAT-0009: Asserts that a fallback response with only `classification_version` yields fresh egress with no country, region or ASN.

<a id="CC-CONTEXT-FORMAT-USE-IN-RESULTS"></a>

### CC-CONTEXT-FORMAT-USE-IN-RESULTS

- Case: `CC-CONTEXT-FORMAT-USE-IN-RESULTS`
- Result: passed
- Test: `TestCcContextContextLayersEgressContext` in `internal/execution/egress_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcContextContextLayersEgressContext$' ./internal/execution`
- Scope: Synthetic core contract cases; Test schema-permitted context carriage/deduplication and reject persistent cross-batch identity or listed private details.
- Requirement aspects exercised:

  - REQ-CONTEXT-FORMAT-0015: Asserts that each Measurement Result written by a direct and by a proxied attempt carries its own `network_context` egress object. Deduplication and cross-batch identifiers are not asserted; the results are written by separate runs.
  - REQ-CONTEXT-FORMAT-0016: Asserts that no result contains the endpoint names, the endpoint path, `127.0.0.1`, the observed source address or port, the proxy port, the proxy generation or the word `generation`. Geolocation and provider names cannot reach a result because the endpoint decoder rejects them (`CC-CONTEXT-ENDPOINT-CONTRACT`).

<a id="CC-PROXY-NETWORK-CONTEXT"></a>

### CC-PROXY-NETWORK-CONTEXT

- Case: `CC-PROXY-NETWORK-CONTEXT`
- Result: passed
- Test: `TestCcProxyNetworkContext` in `internal/execution/egress_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcProxyNetworkContext$' ./internal/execution`
- Scope: Synthetic core contract cases; Fail primary context endpoint and verify same-generation proxied fallback, exact cache keys and required-context gating with zero direct requests.
- Requirement aspects exercised:

  - REQ-PROXY-0045: Asserts that with a failing primary a proxied attempt sends CONNECT for the explicit-IP primary `8.8.8.8:443` and then for the explicit-IP fallback `9.9.9.9:443` to the same scripted proxy, which routes both to loopback endpoints, and records fresh fallback egress on the proxy path with its `connection_code`, while the direct routes see no dial; and that a proxy refusing CONNECT leaves egress unavailable without a direct request. Hostname endpoints on the proxy path are asserted by `UT.internal.execution.TestProxyHostnameContextUnavailable` and `UT.internal.networkcontext.TestProxyContextHostnameEligibility`, and the local address policy check by `UT.internal.execution.TestContextFetchUsesLocalAddressPolicy`.
  - REQ-PROXY-0024: Asserts that a proxy attempt does not reuse the direct entry, that a second attempt on the same proxy generation reuses its entry, and that the next proxy generation asks again.
  - REQ-PROXY-0025: Asserts that a proxy refusing CONNECT with 403 leaves egress unavailable without a direct request, and that a task requiring `country_code` then records `not_run` with `network_context_unavailable` without running the test.

<a id="CC-LINKAGE-OVERVIEW"></a>

### CC-LINKAGE-OVERVIEW

- Case: `CC-LINKAGE-OVERVIEW`
- Result: passed
- Tests: `TestCcLinkageOverview` in `linkage_test.go` (package `root`); `TestDisabledByDefault` in `internal/persistence/linkage/linkage_test.go` (package `internal/persistence/linkage`)
- Reproduce the primary test: `go test -count=1 -run '^(TestCcLinkageOverview)$' ./. && go test -count=1 -run '^(TestDisabledByDefault)$' ./internal/persistence/linkage`
- Scope: Synthetic core contract cases; Initialize once without result_linkage and once with mode "none", then build at least three bundles across two rotation epochs using a deterministic clock and a fake LinkageKeyStore. Assert no bundle contains a linkage object, each bundle_id is distinct and single-use, no bundle-level value other than schema constants and task-defined fields repeats across bundles, and the fake key store records zero DeriveHMACSHA256 calls and no secret creation. Enable rotating_installation with synthetic scope "org.example.app.test" and assert the linkage object contains only mode, scope, epoch and token, with no person, account, device or access-network field. Requirements: REQ-LINKAGE-0001.
- Requirement aspects exercised:

  - REQ-LINKAGE-0001: Without result_linkage and with mode none, three bundles across two 86,400-second epochs carry no linkage, distinct bundle_ids and no repeated bundle-level value other than schema_version and results, with zero key-store derivations and no secret created; with rotating_installation and scope org.example.app.test the linkage object has exactly the members mode, scope, epoch and token. Pseudonymous-not-anonymous wording is RV-CC-LINKAGE-OVERVIEW.

<a id="CC-LINKAGE-GOALS-AND-NON-GOALS"></a>

### CC-LINKAGE-GOALS-AND-NON-GOALS

- Case: `CC-LINKAGE-GOALS-AND-NON-GOALS`
- Result: passed
- Tests: `TestCcLinkageGoalsAndNonGoals` in `linkage_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestCcLinkageGoalsAndNonGoals$' ./.`
- Scope: Synthetic core contract cases; Attempt cross-app/deployment/reinstall/reset/account restoration and targeted delivery; no stable identity continuity is permitted.
- Requirement aspects exercised:

  - REQ-LINKAGE-0002: Two deployment scopes under one secret and two secrets under one scope give different tokens in the same epoch, a fresh key store, `ClearMeasurementData` and the next epoch each give a new token, and only the same key store continues a token; `Initialize` rejects configurations carrying a token, epoch, secret, account, device or restore value, a stable mode or a zero rotation period without using the key store, and `AdmitTask` rejects tasks naming an installation, token or linkage object. Counting people and platform-level reinstall, backup and cross-application separation are not asserted here.

<a id="CC-LINKAGE-INITIALIZATION-CONFIGURATION"></a>

### CC-LINKAGE-INITIALIZATION-CONFIGURATION

- Case: `CC-LINKAGE-INITIALIZATION-CONFIGURATION`
- Result: passed
- Tests: `TestCcLinkageInitializationConfiguration` in `linkage_test.go` (package `root`); `TestCcLinkageRotationLimits` in `linkage_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^(TestCcLinkageInitializationConfiguration|TestCcLinkageRotationLimits)$' ./.`
- Scope: Synthetic core contract cases; Iterate result_linkage initialization values: mode "none" and "rotating_installation" accepted; other strings and wrong case rejected. For rotating_installation, missing scope or missing rotation_period_seconds rejected; scope of 1 and 128 allowed characters accepted, empty and 129 characters rejected, uppercase letters, spaces, "/" and non-ASCII rejected; rotation_period_seconds 86,399 rejected, 86,400 and 604,800 accepted, 604,801 rejected. For mode "none", scope or rotation_period_seconds present is rejected. Absent result_linkage behaves as mode "none" with no linkage object and no key derivation. Assert tasks, scripted server responses and runtime task configuration carrying result_linkage, scope or rotation fields cannot enable linkage, change scope or lengthen rotation; that no runtime setter changes linkage configuration and a changed configuration takes effect only after reinitialization; and that trusted local safety policy disabling linkage yields no linkage object even when initialization requests rotating_installation. Requirements: REQ-LINKAGE-0003, REQ-LINKAGE-0004, REQ-LINKAGE-0005, REQ-LINKAGE-0006.
- Requirement aspects exercised:

  - REQ-LINKAGE-0003: Initialize accepts mode none, scopes of 1 and 128 allowed characters and rotation 86,400 and 604,800, and rejects other or wrong-case modes, missing mode, scope or rotation, empty, 129-character, uppercase, space, '/' and non-ASCII scopes and rotation 86,399 and 604,801, each before any key-store call.
  - REQ-LINKAGE-0004: Absent result_linkage and mode none produce bundles without linkage and no key-store call; mode none with scope or rotation_period_seconds is rejected.
  - REQ-LINKAGE-0005: Only a scope on the trusted approved list is accepted (an unreviewed, syntactically valid scope is rejected); whether the scope is user- or installation-derived is left to RV-CC-LINKAGE-INITIALIZATION-CONFIGURATION.
  - REQ-LINKAGE-0006: Tasks carrying result_linkage at top level or in options are not admitted and SetProxy rejects a linkage object, the next bundle keeping the initialized scope and 86,400-second epoch; Client has no linkage method; Initialize while initialized is refused and a changed period applies only after clearing and reinitialization; local policy disallowing linkage yields no linkage and no key call for a rotating_installation request.

<a id="CC-LINKAGE-INSTALLATION-SECRET"></a>

### CC-LINKAGE-INSTALLATION-SECRET

- Case: `CC-LINKAGE-INSTALLATION-SECRET`
- Result: passed
- Tests: `TestCcLinkageInstallationSecret` in `linkage_conformance_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^TestCcLinkageInstallationSecret$' ./.`
- Scope: Synthetic core contract cases; Use fake protected-key capabilities and controlled randomness to check fresh 256-bit key creation, no host-derived secret, protected ownership and no export/backup/network/log/analytics sinks.
- Requirement aspects exercised:

  - REQ-LINKAGE-0007: With a key store drawing from a controlled random source, no secret is drawn at `Initialize`, at admission or while linkage is disabled; the first linked bundle draws exactly one 32-byte secret that later tokens derive from, and after `ClearMeasurementData` the next linked bundle draws a fresh, different secret. The randomness quality of the source is not asserted.
  - REQ-LINKAGE-0008: The key-store contract exposes only `Delete` and `DeriveHMACSHA256` (no import or export), clearing deletes the held secret without creating a replacement, and no configuration can supply a secret. Non-return through APIs, network and log sinks is left to other tests, and non-exportability and backup exclusion are not asserted here.
  - REQ-LINKAGE-0009: Every derivation message equals the message built from the purpose, deployment scope and epoch, every token matches the HMAC under the drawn secret, and configurations supplying a secret, key, seed, top-level installation secret, vendor ID, null secret or `linkage_keys` are rejected without drawing or using a key.

<a id="CC-LINKAGE-TOKEN-DERIVATION"></a>

### CC-LINKAGE-TOKEN-DERIVATION

- Case: `CC-LINKAGE-TOKEN-DERIVATION`
- Result: passed
- Tests: `TestDerivation` in `internal/persistence/linkage/linkage_test.go` (package `internal/persistence/linkage`); `TestFailuresOmitLinkage` in `internal/persistence/linkage/linkage_test.go` (package `internal/persistence/linkage`); `TestEpochFloor` in `internal/persistence/linkage/linkage_test.go` (package `internal/persistence/linkage`); `TestCcLinkageOverview` in `linkage_test.go` (package `root`); `TestBundleLinkage` in `internal/persistence/pending/linkage_test.go` (package `internal/persistence/pending`); `TestLinkedBundleRestart` in `internal/persistence/pending/linkage_test.go` (package `internal/persistence/pending`)
- Reproduce the primary test: `go test -count=1 -run '^(TestDerivation|TestFailuresOmitLinkage|TestEpochFloor)$' ./internal/persistence/linkage && go test -count=1 -run '^(TestCcLinkageOverview)$' ./. && go test -count=1 -run '^(TestBundleLinkage|TestLinkedBundleRestart)$' ./internal/persistence/pending`
- Scope: Synthetic core contract cases; Use independent HMAC known-answer vectors for exact purpose/separator/scope/decimal epoch message, floor boundaries, base64url and memory cleanup; suppress linkage on invalid/rollback clock until nondecreasing validity.
- Requirement aspects exercised:

  - REQ-LINKAGE-0010: The key store receives exactly "barometer-result-linkage-v1\x00<scope>\x00<decimal epoch>" and the projection equals two HMAC-SHA-256 known-answer vectors; at the facade each bundle token equals an HMAC computed in the test.
  - REQ-LINKAGE-0011: The epoch is floor(unix/period): 2964 at the last second before 2965*604800 and 2965 at it, and the bundle's encoded epoch is the integer in the HMAC message.
  - REQ-LINKAGE-0012: Different epochs and a different scope under the same secret derive different tokens.
  - REQ-LINKAGE-0013: The bundle token is unpadded base64url of the 32-byte HMAC (43 characters, matching the independently computed value), and stored bundles with padded, short or non-canonical tokens are rejected; memory clearing is a recorded disposition, not asserted.
  - REQ-LINKAGE-0014: An unavailable clock, a pre-1970 time and a cancelled context omit linkage without a key call; a clock moved back into an earlier epoch omits linkage, including after reinitialization of the same client, until the floor epoch returns.

<a id="CC-LINKAGE-ROTATION-SEMANTICS"></a>

### CC-LINKAGE-ROTATION-SEMANTICS

- Case: `CC-LINKAGE-ROTATION-SEMANTICS`
- Result: passed
- Tests: `TestCcLinkageRotationSemantics` in `internal/persistence/linkage/linkage_test.go` (package `internal/persistence/linkage`); `TestDerivation` in `internal/persistence/linkage/linkage_test.go` (package `internal/persistence/linkage`); `TestCcLinkageOverview` in `linkage_test.go` (package `root`); `TestBundleLinkage` in `internal/persistence/pending/linkage_test.go` (package `internal/persistence/pending`); `TestCcLinkageRotationLimits` in `linkage_test.go` (package `root`); `TestCcLinkageInitializationConfiguration` in `linkage_test.go` (package `root`)
- Reproduce the primary test: `go test -count=1 -run '^(TestCcLinkageRotationSemantics|TestDerivation)$' ./internal/persistence/linkage && go test -count=1 -run '^(TestCcLinkageOverview|TestCcLinkageRotationLimits|TestCcLinkageInitializationConfiguration)$' ./. && go test -count=1 -run '^(TestBundleLinkage)$' ./internal/persistence/pending`
- Scope: Synthetic core contract cases; Compare same/different scopes/epochs across boundaries; no old-new token mapping or continuity request, and no period extension beyond reviewed bounds.
- Requirement aspects exercised:

  - REQ-LINKAGE-0015: Two sources with one secret agree on the token of a scope and epoch up to the boundary second and rotate together at it; at the facade two bundles of one epoch share a token and a bundle in the next epoch carries a new one; a pending bundle keeps the linkage derived when it was created.
  - REQ-LINKAGE-0016: The projection holds only scope, epoch and token and the bundle linkage object only mode, scope, epoch and token (stored extra members rejected), and after rotation the previous epoch's token is not derived again; the collector-side continuity prohibition is a recorded disposition.
  - REQ-LINKAGE-0017: Rotation 86,399 and 604,801 are rejected and 86,400 and 604,800 accepted with epochs computed from exactly the configured period, and tasks or the proxy setter cannot change the period; the shortest-needed-period choice is a deployment decision.

<a id="CC-BUNDLE-LINKAGE-OBJECT"></a>

### CC-BUNDLE-LINKAGE-OBJECT

- Case: `CC-BUNDLE-LINKAGE-OBJECT`
- Result: passed
- Tests: `TestCcBundleLinkageObject` in `internal/persistence/pending/linkage_test.go` (package `internal/persistence/pending`); `TestLinkedBundleRestart` in `internal/persistence/pending/linkage_test.go` (package `internal/persistence/pending`); `TestBundleLinkage` in `internal/persistence/pending/linkage_test.go` (package `internal/persistence/pending`)
- Reproduce the primary test: `go test -count=1 -run '^(TestCcBundleLinkageObject|TestLinkedBundleRestart|TestBundleLinkage)$' ./internal/persistence/pending`
- Scope: Synthetic core contract cases; Validate the exact complete rotating-installation object and local scope, encoded token/epoch, no stable mode or user_id representation.
- Requirement aspects exercised:

  - REQ-BUNDLE-0010: A linked bundle contains exactly one linkage member, byte-for-byte {"mode":"rotating_installation","scope":<configured scope>,"epoch":<decimal>,"token":<43-character unpadded base64url of 32 bytes>}; a projection in another scope or with a negative epoch is omitted.
  - REQ-BUNDLE-0011: Stored bundles with mode stable_installation, an extra, renamed or reordered linkage member, or a non-canonical scope, epoch or token are rejected on open.
  - REQ-BUNDLE-0012: Produced bundles contain no user_id, and stored bundles with a user_id member in or beside linkage, or the token renamed to user_id, are rejected; collector labeling is outside the library.

## Cross-Cutting Unit-Test Cases

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

<a id="UT.internal.admission.TestGateReleaseReviewedOneTimeException"></a>

### UT.internal.admission.TestGateReleaseReviewedOneTimeException

- Case: `UT.internal.admission.TestGateReleaseReviewedOneTimeException`
- Result: passed
- Test: `TestGateReleaseReviewedOneTimeException` in `internal/admission/release_review_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateReleaseReviewedOneTimeException$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateReleaseReviewedOneTimeException in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0013: An embedded one-time task without `not_after` is admitted only when the active local policy lists the embedded issuer with the SHA-256 of the task's exact bytes; with no listing, another task listed, re-spaced bytes, another issuer, `max_runs` 2, a repeat, or as a remote task it is rejected with `task_schema_invalid` before the codec runs, and a task with `not_after` needs no listing. Non-persistence across application updates and the limits applied to repeating embedded tasks are not asserted.

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
  - REQ-TRUST-0036: Asserts that a registered and enabled NET-05 proxy task whose synthetic composite codec supplies no normalized bounded step list is `task_unsupported`.
  - REQ-TRUST-0039: Asserts that the NET-05 registry rule of that synthetic entry is not consulted, so a composite without normalized steps is not classified by its rule. Rejection of a production NET-05 profile with an incompatible step as `task_schema_invalid` is asserted by `UT.root.TestMNet05AtomicTargetAndProxyRules`.

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
  - REQ-TRUST-0011: Asserts that issuer and deployment ceilings bound attempt duration, bytes sent and received, connections, max_runs, retry max_attempts, validity lifetime, repeat interval and jitter minima, and retry initial delay, maximum delay and jitter; an embedded one-time task without `not_after` has the registered maximum lifetime of 604,800 seconds, so it is admitted under a validity ceiling of 604,800 seconds and is task_limit_exceeded under 604,799 or 3,600 seconds.
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

  - REQ-PROXY-0044: Asserts with synthetic rules that a proxy task of a test whose registered rule is never compatible (`registry.NoProxy`) is rejected and no direct task is, and that the rule's per-parameter verdict, combined by `registry.ProxyCompatible` with the common destination check, decides acceptance. The production per-test rows are not exercised here; see `UT.internal.registry.TestCommonProxyDestinationCompatibility`, `UT.root.TestCcProxyOverviewDNSRow` and `UT.root.TestMNet05AtomicTargetAndProxyRules`.
  - REQ-PROXY-0027: Asserts that a proxy task whose synthetic rule reports its parameters incompatible is rejected atomically at admission, for embedded and remote origin, with no call or open connection on the unused recording transport; no compatible subset is admitted.
  - REQ-PROXY-0046: Asserts that an incompatible proxy task is rejected at admission as `task_schema_invalid` rather than returned as a gated task that could end in `proxy_path_unsupported`. The pre-attempt and per-dial execution checks are not exercised by this test.
  - REQ-TRUST-0018: Asserts that every compatibility rejection and every acceptance causes no call or open connection on the unused recording transport.
  - REQ-TRUST-0039: Asserts that a `proxy` task whose synthetic rule reports its parameters incompatible is `task_schema_invalid`, that a compatible one is accepted with its effective path narrowed to proxy, that a `direct` task is not checked, that a never-compatible rule (`registry.NoProxy`) rejects every proxy task and no direct task, and that the check runs after input and options and before issuer and local path authority, target registration and the step-9 risk bound. NET-05 profiles and `task_unsupported` for a permitted but unprovided combination are not tested.
  - REQ-TRUST-0047: Asserts that invalid input, invalid options, a codec failure, a test not permitted by the issuer and policy rollback keep their classes ahead of the compatibility check and leave the rule unconsulted, that an egress context requirement does not bypass the check, and that a compatible task still receives `task_issuer_not_authorized` or `target_not_registered` from the later step-8 checks.

<a id="UT.internal.admission.TestGateRejectsCompositeProfiles"></a>

### UT.internal.admission.TestGateRejectsCompositeProfiles

- Case: `UT.internal.admission.TestGateRejectsCompositeProfiles`
- Result: passed
- Test: `TestGateRejectsCompositeProfiles` in `internal/admission/capability_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateRejectsCompositeProfiles$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateRejectsCompositeProfiles in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0018: Asserts that every NET-05 rejection, unregistered, without composite enablement or registered and enabled, causes no call or open connection on the unused recording transport.
  - REQ-TRUST-0036: Asserts that NET-05 absent from the registry, at the registered or another version, is task_unsupported even under a policy that enables composite tasks, and that a registered NET-05 composite enabled by both issuer and deployment is task_unsupported because its synthetic codec supplies no normalized step list, while SYN-01 is admitted under the same policy. Composite-step matching of the production NET-05 codec is not exercised by this test.
  - REQ-TRUST-0047: Asserts that policy rollback and missing composite enablement in the issuer or in the deployment (`task_issuer_not_authorized`) precede the step-8 composite check, that invalid input precedes it, that an egress context requirement does not change its `task_unsupported` class, and that it precedes path authority, target registration and risk bounds.

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
  - REQ-TRUST-0014: Asserts that the matched registration's direct or proxy path permission binds, that a redirect not listed by the matched registration is rejected, and that a proxy-only duplicate tuple does not hide a direct registration. The redirect path-permission row now asserts that the HTTP-01 proxy task, whose URL host is a hostname, is rejected earlier as `task_schema_invalid` by proxy compatibility (REQ-PROXY-0044), so a redirect registration's path permission is no longer exercised by this case.

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

  - REQ-TRUST-0014: Changes one component at a time (IP address and family, port, scheme, added hostname, server name, each ALPN change including order, URL host, path, port and query, expected status, body and header checks, redirect target, an extra redirect, queried name, resolver name, bootstrap address including order, and direct versus proxy path) and asserts target_not_registered for each, except the HTTP-01 proxy-path variant: its hostname URL is now rejected earlier as `task_schema_invalid` by proxy compatibility (REQ-PROXY-0044).
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

  - REQ-SCHEDULE-0020: Asserts under the anchored due-time grid of REQ-SCHEDULE-0045 that occurrence 0 is due at the anchor, not one second before, and that after its reservation occurrence 1 is due at the anchor plus the 21,600-second interval plus jitter: one second before that it is not due, at and after it it is.
  - REQ-SCHEDULE-0021: Asserts that the jitter of a due run comes from the injected local jitter source and is capped at the task's `jitter`: a source offering one hour yields 300 seconds.
  - REQ-SCHEDULE-0045: Asserts that occurrence 0 is due at the anchor without jitter and that occurrence 1 is due at the anchor plus the interval plus a jitter drawn when the occurrence becomes current.

<a id="UT.internal.execution.TestMissedOccurrencesAreSkipped"></a>

### UT.internal.execution.TestMissedOccurrencesAreSkipped

- Case: `UT.internal.execution.TestMissedOccurrencesAreSkipped`
- Result: passed
- Test: `TestMissedOccurrencesAreSkipped` in `internal/execution/occurrence_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestMissedOccurrencesAreSkipped$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestMissedOccurrencesAreSkipped in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SCHEDULE-0020: Asserts under the anchored due-time grid of REQ-SCHEDULE-0045 that when occurrence 1 is missed, occurrence 2 becomes the due run without consuming a run for the skipped one, and that after the last occurrence of `max_runs` 4 no run is due.
  - REQ-SCHEDULE-0021: Asserts that the jitter of a due run is drawn once: advancing within the current occurrence with a different jitter source leaves its due time unchanged.
  - REQ-SCHEDULE-0045: Asserts that a missed occurrence is skipped without consuming a run, that the jitter of the current occurrence is not redrawn, and that the schedule ends once the current occurrence reaches `max_runs`.

<a id="UT.internal.execution.TestNetworkChangeDuringAttempt"></a>

### UT.internal.execution.TestNetworkChangeDuringAttempt

- Case: `UT.internal.execution.TestNetworkChangeDuringAttempt`
- Result: passed
- Test: `TestNetworkChangeDuringAttempt` in `internal/execution/network_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestNetworkChangeDuringAttempt$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestNetworkChangeDuringAttempt in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-CONTEXT-0025: Asserts that a generation change the fake platform reports after the attempt's capture ends the attempt with `network_changed`, including a change between capture and watch, which leaves no runner call and no dialer issued, and that every watch is stopped when the attempt ends.
  - REQ-RESULT-0051: Asserts that a generation change before the test starts records `not_run` with `network_changed` and no Network Context, and after it starts `incomplete` with `network_changed` and a Network Context whose `network_changed` is true and whose egress is unavailable on the direct path.
  - REQ-RESULT-0052: Asserts that a generation change after the runner returned its outcome leaves the attempt `completed` with its measurement and a Network Context whose `network_changed` is false.

<a id="UT.internal.execution.TestNetworkConditionsBeforeEachAttempt"></a>

### UT.internal.execution.TestNetworkConditionsBeforeEachAttempt

- Case: `UT.internal.execution.TestNetworkConditionsBeforeEachAttempt`
- Result: passed
- Test: `TestNetworkConditionsBeforeEachAttempt` in `internal/execution/network_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestNetworkConditionsBeforeEachAttempt$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestNetworkConditionsBeforeEachAttempt in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0031: Asserts that `require_validated` with validation unknown at the first attempt records `not_run` with `constraints_unsatisfied`; that when validation becomes unknown after a retryable first attempt, the retry is not attempted and writes no result; and that without a network condition a failed access capture records `not_run` with `internal_error` rather than `network_context_unavailable`.

<a id="UT.internal.execution.TestNetworkConditionsFailClosed"></a>

### UT.internal.execution.TestNetworkConditionsFailClosed

- Case: `UT.internal.execution.TestNetworkConditionsFailClosed`
- Result: passed
- Test: `TestNetworkConditionsFailClosed` in `internal/execution/gates_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestNetworkConditionsFailClosed$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestNetworkConditionsFailClosed in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0028: Asserts through `execution.NetworkSatisfiable` and `execution.Gates` that an unregistered or unset path fails the start check while the direct and proxy paths without conditions pass; that an allowed-transport list, a `vpn_policy` other than `any` and `require_validated` pass the static gate and are reported as requiring access context; that required context fields and a country allowlist pass the static gate without requiring access context, for evaluation against each attempt's egress; and that an unset or unregistered `vpn_policy` never satisfies `execution.NetworkConditions`.

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

  - REQ-LIMIT-0007: Asserts that with zero byte and connection limits, as for NET-01, a dial is refused before the path dialer is called, the server accepts nothing, nothing is metered and, the test not having started, the attempt is recorded as `not_run` with `resource_limit`.

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
  - REQ-OBS-0052: `NegotiatedALPN` maps a negotiated `http/1.1`, `h2` or `dot` to itself and `h3` or another identifier to `unknown`, and returns no value when nothing was negotiated or the negotiated identifier is empty; `ValidateALPNOffer` accepts only offers drawn from the three registered identifiers.

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
  - REQ-TRUST-0060: A trusted key supplied to `NewLocalPolicy` is resolved only by its exact key ID, later changes to the input or to a returned copy do not alter it, unknown or case-variant IDs and the zero policy resolve no key, and four keys per issuer and 32 per policy are accepted. That delivered content cannot add keys is not asserted.

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
  - REQ-TRUST-0060: Each malformed or ambiguous key record (invalid or duplicate key ID, unknown or empty issuer, zero, identity or non-canonical public key, invalid validity window, a fifth key for one issuer or a 33rd key) makes `NewLocalPolicy` return the fixed trust-model error with an invalid, generation-0 policy that resolves no key.

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
  - REQ-BUNDLE-0037: Asserts on the injected clock that a bundle's retention runs from the persisting of its first result: a result added an hour later and a simulated restart do not move it, a returned bundle still expires at that time, and the bundle is deleted and never returned once 86,400 seconds have passed.

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

  - REQ-TRUST-0018: Asserts with `go list -deps`, for the host build, that the transitive dependency closure of `internal/admission` contains no `net` package other than `net/netip` and `net/url` and no `crypto/tls`, and that the closure of the root package contains no `os/exec`, `plugin` or `internal/testkit` package and reaches package `net` and `crypto/tls` only through `internal/execution` (the path dialers and the Network Context endpoint request) and `internal/protocol/tls` (the TLS-01 handshake over a stream those dialers opened), and `net/http` additionally through `internal/protocol/http`. Admission links no code that can open a network connection, and admission precedes every run. This is structural evidence: it does not observe runtime behavior, cover mobile builds or constrain a future trusted adapter seam.

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

<a id="UT.root.TestRegisteredSchemasStrictJSON"></a>

### UT.root.TestRegisteredSchemasStrictJSON

- Case: `UT.root.TestRegisteredSchemasStrictJSON`
- Result: passed
- Test: `TestRegisteredSchemasStrictJSON` in `strict_schemas_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestRegisteredSchemasStrictJSON$' .`
- Scope: Existing synthetic Go unit test TestRegisteredSchemasStrictJSON in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0003: Asserts through `Client.AdmitTask`, for embedded and signed remote delivery of TLS-01, HTTP-01, NET-03, NET-05, DNS-03 and DNS-04 tasks with the production codecs, that an equal repetition and an escaped-equivalent repetition (first name character as a `\u` escape) of an `input` member, and of an `options` member for every schema except NET-05, whose options are empty, are `task_schema_invalid`; for NET-05 the repetitions are also placed inside a step's `input`, which only that step's registered schema validates.
  - REQ-JSON-0004: Asserts for all seven schemas, including NET-01's empty `input` and `options`, that an escaped unknown name `\u0078` in `input`, in `options` and inside a NET-05 step's `input` is `task_schema_invalid`, and that escaped spellings of a known `input` or `options` name are admitted.
  - REQ-JSON-0005: Asserts that `null`, an array, an object and a boolean in place of a required string `input` member are `task_schema_invalid`.
  - REQ-JSON-0006: Asserts that `-0`, decimal (`.0`), exponent (`e0`, `E+0`) and quoted spellings of the TLS-01 `port`, HTTP-01 and NET-03 `expected_status`, NET-05 `profile_version` and DNS-04 `port` input members, and of HTTP-01 `max_body_bytes` in `options`, are `task_schema_invalid`.
  - REQ-JSON-0010: Asserts that every rejection returns no task handle and leaves no replay commit, run or result record, and that no case, accepted or rejected, calls the test's dialer, resolver or platform network-state reader during admission.

<a id="UT.root.TestRegisteredSchemasStrictJSONBounds"></a>

### UT.root.TestRegisteredSchemasStrictJSONBounds

- Case: `UT.root.TestRegisteredSchemasStrictJSONBounds`
- Result: passed
- Test: `TestRegisteredSchemasStrictJSONBounds` in `strict_schemas_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestRegisteredSchemasStrictJSONBounds$' .`
- Scope: Existing synthetic Go unit test TestRegisteredSchemasStrictJSONBounds in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0009: Asserts for TLS-01, HTTP-01, NET-01, NET-03, NET-05, DNS-03 and DNS-04, with each baseline value padded with whitespace, that `input` of 49,151 and 49,152 original bytes and `options` of 8,191 and 8,192 bytes are admitted, and that one byte over either bound is `task_limit_exceeded`.
  - REQ-JSON-0010: Asserts that each over-bound rejection returns no task handle and leaves no replay commit, run or result record, and that no case calls the test's dialer, resolver or platform network-state reader during admission.

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

<a id="UT.internal.execution.TestContextEndpointResponses"></a>

### UT.internal.execution.TestContextEndpointResponses

- Case: `UT.internal.execution.TestContextEndpointResponses`
- Result: passed
- Test: `TestContextEndpointResponses` in `internal/execution/egress_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestContextEndpointResponses$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestContextEndpointResponses in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-CONTEXT-0010: Asserts that a 302 primary response with a `Location` header is not followed: egress comes from the one fallback request.
  - REQ-CONTEXT-0015: Asserts against synthetic HTTPS endpoints that a primary body of exactly 4,096 bytes is accepted and a larger one is invalid, and that a stalled primary is stopped at 5,000 ms on the engine clock, not 1 ms earlier, after which the fallback answers.
  - REQ-CONTEXT-0043: Asserts against synthetic HTTPS endpoints at explicit-IP names on the direct path that a 500, redirect, non-JSON, gzip-encoded, empty, malformed or oversized response, an unknown certificate authority, a wrong certificate name and an unreachable primary are each followed by exactly one fallback request through the same dialer, which supplies fresh fallback egress.
  - REQ-CONTEXT-0021: Asserts that the primary is never requested more than once in one acquisition.

<a id="UT.internal.execution.TestCcContextEndpointContract"></a>

### UT.internal.execution.TestCcContextEndpointContract

- Case: `UT.internal.execution.TestCcContextEndpointContract`
- Result: passed
- Test: `TestCcContextEndpointContract` in `internal/execution/egress_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcContextEndpointContract$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestCcContextEndpointContract in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-CONTEXT-0036: Asserts on the decrypted request bytes that the endpoint request carries no `Cookie` or `Authorization` header, no task ID or test ID, no `connection_code` and no loopback target address.

<a id="UT.internal.networkcontext.TestCcContextEndpointContract"></a>

### UT.internal.networkcontext.TestCcContextEndpointContract

- Case: `UT.internal.networkcontext.TestCcContextEndpointContract`
- Result: passed
- Test: `TestCcContextEndpointContract` in `internal/networkcontext/egress_test.go` (package `internal/networkcontext`)
- Reproduce one case: `go test -count=1 -run '^TestCcContextEndpointContract$' ./internal/networkcontext`
- Scope: Existing synthetic Go unit test TestCcContextEndpointContract in package internal/networkcontext, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-CONTEXT-0013: Asserts that ASN 0, above 32 bits, as a string or spelled `-0`, `64500.0` or `6.45e4`, an empty or three-letter country, a region inconsistent with the country, an empty version, a non-object and an empty body make the response invalid and egress unavailable, and that a response with only `classification_version` is valid.
  - REQ-CONTEXT-0014: Asserts that responses carrying `ip`, `latitude`, `city`, `postal_code`, `isp` or `request_id` members, an escaped unknown name, or an equal or escaped-equivalent duplicate member, are invalid.
  - REQ-CONTEXT-0015: Asserts that 4,096 bytes are accepted and 4,097 rejected, and that on the engine clock an answer after 4,999 ms is accepted while one after 5,001 ms is rejected, with no operation timer left armed after either answer.
  - REQ-CONTEXT-0016: Asserts that a version-only response yields fresh primary egress with no classification values.
  - REQ-JSON-0003: Asserts that a response repeating `country_code` with equal values, directly or as `\u0063ountry_code`, makes the response invalid and egress unavailable.
  - REQ-JSON-0004: Asserts that an escaped unknown name `\u0078` makes the response invalid, and that an escaped spelling `\u0061sn` of the known name `asn` is accepted as fresh primary egress.
  - REQ-JSON-0006: Asserts that `asn` spelled `-0`, `64500.0` or `6.45e4` makes the response invalid and egress unavailable.

<a id="UT.internal.execution.TestProxyAttemptRunsThroughTunnel"></a>

### UT.internal.execution.TestProxyAttemptRunsThroughTunnel

- Case: `UT.internal.execution.TestProxyAttemptRunsThroughTunnel`
- Result: passed
- Test: `TestProxyAttemptRunsThroughTunnel` in `internal/execution/proxy_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestProxyAttemptRunsThroughTunnel$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestProxyAttemptRunsThroughTunnel in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0033: Asserts that metered usage equals the payload bytes the loopback target read and wrote through the tunnel and one connection, while the CONNECT request and response that the scripted proxy counted are not metered.
  - REQ-PROXY-0017: Asserts that one proxied attempt opens one connection to the scripted loopback proxy, which reads exactly one CONNECT for the target's address and relays one tunnel, and that the target is reached only through it.
  - REQ-PROXY-0031: Asserts that the result records the proxy path with its `connection_code` and no proxy endpoint, generation or CONNECT material.

<a id="UT.internal.execution.TestConnectExchangeIsNotMetered"></a>

### UT.internal.execution.TestConnectExchangeIsNotMetered

- Case: `UT.internal.execution.TestConnectExchangeIsNotMetered`
- Result: passed
- Test: `TestConnectExchangeIsNotMetered` in `internal/execution/proxy_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestConnectExchangeIsNotMetered$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestConnectExchangeIsNotMetered in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0033: Asserts that byte limits equal to the 16 request and 24 reply bytes through the tunnel let the attempt complete, although the CONNECT request alone is larger than the sent limit.
  - REQ-LIMIT-0034: Asserts that the tunnel counts toward `max_connections`: with a connection limit of zero the attempt is recorded as `not_run` with `resource_limit` and the proxy accepts no connection.

<a id="UT.internal.execution.TestProxyFailuresAreExecutionStates"></a>

### UT.internal.execution.TestProxyFailuresAreExecutionStates

- Case: `UT.internal.execution.TestProxyFailuresAreExecutionStates`
- Result: passed
- Test: `TestProxyFailuresAreExecutionStates` in `internal/execution/proxy_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestProxyFailuresAreExecutionStates$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestProxyFailuresAreExecutionStates in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0010: Asserts that an absent, refusing, malformed or closing proxy is an explicit `incomplete` execution state and never a direct dial.
  - REQ-PROXY-0028: Asserts `proxy_unavailable` for a closed endpoint port and `proxy_connect_failed` for a 403 with header and body canaries, HTTP/1.0, a four-digit status, a SOCKS reply, a header line without a colon, a head of more than 9,000 bytes and a close without a reply.
  - REQ-PROXY-0029: Asserts that each result keeps the execution state with no runner outcome, and that its proxy egress carries no response material.
  - REQ-PROXY-0030: Asserts that the loopback target is never dialed outside the proxy and that the failing proxy relays nothing.

<a id="UT.internal.execution.TestProxyRetryNeverFallsBackToDirect"></a>

### UT.internal.execution.TestProxyRetryNeverFallsBackToDirect

- Case: `UT.internal.execution.TestProxyRetryNeverFallsBackToDirect`
- Result: passed
- Test: `TestProxyRetryNeverFallsBackToDirect` in `internal/execution/proxy_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestProxyRetryNeverFallsBackToDirect$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestProxyRetryNeverFallsBackToDirect in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0030: Asserts that two attempts ending `proxy_unavailable`, separated by the 35-second backoff, never reach the target directly.

<a id="UT.internal.execution.TestEveryStreamOpensItsOwnTunnel"></a>

### UT.internal.execution.TestEveryStreamOpensItsOwnTunnel

- Case: `UT.internal.execution.TestEveryStreamOpensItsOwnTunnel`
- Result: passed
- Test: `TestEveryStreamOpensItsOwnTunnel` in `internal/execution/proxy_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestEveryStreamOpensItsOwnTunnel$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestEveryStreamOpensItsOwnTunnel in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0018: Asserts that an attempt with two connections opens two CONNECT tunnels on the same endpoint.
  - REQ-PROXY-0020: Asserts that a second attempt opens two more tunnels, four connections and four relays in all, so no tunnel is reused across attempts.

<a id="UT.internal.execution.TestLateConnect"></a>

### UT.internal.execution.TestLateConnect

- Case: `UT.internal.execution.TestLateConnect`
- Result: passed
- Test: `TestLateConnect` in `internal/execution/proxy_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestLateConnect$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestLateConnect in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0017: Asserts that a CONNECT the proxy holds unanswered past the attempt deadline is abandoned and its socket released.
  - REQ-PROXY-0028: Asserts `incomplete` with `proxy_connect_failed` for the deadline and with `cancelled_by_host` for a host cancellation during the exchange, with no direct dial in either case.

<a id="UT.internal.execution.TestPathCheckMakesNoContact"></a>

### UT.internal.execution.TestPathCheckMakesNoContact

- Case: `UT.internal.execution.TestPathCheckMakesNoContact`
- Result: passed
- Test: `TestPathCheckMakesNoContact` in `internal/execution/proxy_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestPathCheckMakesNoContact$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestPathCheckMakesNoContact in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0008: Asserts that a failing pre-attempt check records `not_run` with `proxy_not_configured`, `proxy_not_permitted` or `proxy_path_unsupported` before any context capture or proxy contact and ends the run without a retry, although the schedule lists retryable reasons, and that a passing check lets the attempt complete.
  - REQ-EXECUTION-0010: Asserts that each denied precondition is an explicit `not_run` state with no contact.
  - REQ-PROXY-0046: Asserts that an admitted task whose pre-attempt check reports `proxy_path_unsupported` records `not_run` and makes no contact. The check is scripted; admission applies the same registered rules in this build. The per-dial gate is not exercised by this test.

<a id="UT.internal.execution.TestConnectAuthority"></a>

### UT.internal.execution.TestConnectAuthority

- Case: `UT.internal.execution.TestConnectAuthority`
- Result: passed
- Test: `TestConnectAuthority` in `internal/execution/proxy_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestConnectAuthority$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestConnectAuthority in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0043: Asserts that an explicit loopback IPv4 address and port is sent unchanged as the CONNECT authority and relays the payload, that an IPv6 literal keeps its brackets, and that a mixed-case hostname authority is refused, like the other invalid authorities, without contacting the proxy endpoint. Whether a hostname is resolved locally is not observed by this test.
  - REQ-PROXY-0020: Asserts that `udp`, `tcp4` and `tcp6` dials and authorities without a valid port or hostname, with a zone, with an injected header line or over-long, and a hostname authority, are refused without contacting the endpoint.

<a id="UT.internal.execution.TestProxyEndpointIsLoopback"></a>

### UT.internal.execution.TestProxyEndpointIsLoopback

- Case: `UT.internal.execution.TestProxyEndpointIsLoopback`
- Result: passed
- Test: `TestProxyEndpointIsLoopback` in `internal/execution/proxy_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestProxyEndpointIsLoopback$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestProxyEndpointIsLoopback in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0005: Asserts that `NewProxyDialers` accepts `127.0.0.1:1`, `127.0.0.2:65535`, `::1` and `localhost` and rejects ports 0 and 65536 and an empty host.
  - REQ-PROXY-0007: Asserts that `192.0.2.1`, `10.0.0.1`, an IPv4-mapped IPv6 address, a link-local address with a zone and a hostname are rejected, and that zero-value dialers refuse to dial.

<a id="UT.root.TestEgressCacheLifecycle"></a>

### UT.root.TestEgressCacheLifecycle

- Case: `UT.root.TestEgressCacheLifecycle`
- Result: passed
- Test: `TestEgressCacheLifecycle` in `network_context_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestEgressCacheLifecycle$' .`
- Scope: Existing synthetic Go unit test TestEgressCacheLifecycle in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-CONTEXT-0023: Asserts through the facade that a second run on the same generation and path reuses the entry, and that `ClearMeasurementData`, reinitialization, policy replacement and `Close` each empty the cache, so the next run asks the primary again (three requests in all) and the new policy generation caches again.
  - REQ-CONTEXT-0024: Asserts that the cache is empty after clearing and after `Close`: it lives in memory with the initialized client and survives neither.

<a id="UT.root.TestEgressCacheProxyChange"></a>

### UT.root.TestEgressCacheProxyChange

- Case: `UT.root.TestEgressCacheProxyChange`
- Result: passed
- Test: `TestEgressCacheProxyChange` in `network_context_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestEgressCacheProxyChange$' .`
- Scope: Existing synthetic Go unit test TestEgressCacheProxyChange in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0016: Asserts through the facade that `SetProxy` and `ClearProxy` drop the previous generation's proxy entry and keep the direct entry, and that the next proxy run asks again through the new proxy with CONNECT to the explicit-IP endpoint `8.8.8.8:443`.

<a id="UT.root.TestCcProxyConfigurationValidation"></a>

### UT.root.TestCcProxyConfigurationValidation

- Case: `UT.root.TestCcProxyConfigurationValidation`
- Result: passed
- Test: `TestCcProxyConfigurationValidation` in `proxy_lifecycle_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcProxyConfigurationValidation$' .`
- Scope: Existing synthetic Go unit test TestCcProxyConfigurationValidation in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0005: Asserts through `Client.SetProxy` that a replacement with port 0, 65536 or `1.0`, or with an uppercase, unapproved, 65-character or empty `connection_code`, returns `proxy_rejected` and leaves the active configuration unchanged.
  - REQ-PROXY-0006: Asserts that a replacement carrying a `password` member returns `proxy_rejected`.
  - REQ-PROXY-0007: Asserts that a replacement with the non-loopback host `198.51.100.1` returns `proxy_rejected`, leaving the loopback endpoint active and the run uncancelled.

<a id="UT.root.TestProxyChangesOnlyCancelProxyRuns"></a>

### UT.root.TestProxyChangesOnlyCancelProxyRuns

- Case: `UT.root.TestProxyChangesOnlyCancelProxyRuns`
- Result: passed
- Test: `TestProxyChangesOnlyCancelProxyRuns` in `proxy_lifecycle_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestProxyChangesOnlyCancelProxyRuns$' .`
- Scope: Existing synthetic Go unit test TestProxyChangesOnlyCancelProxyRuns in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0011: Asserts that after `ClearProxy`, called twice, no proxy remains configured and a proxy task records `not_run` with `proxy_not_configured`, and that `SetProxy` then installs the single replacement used by the next proxy run.
  - REQ-PROXY-0012: Asserts that a direct run active during `SetProxy` or `ClearProxy` stays active and completes, while only the proxy run is cancelled.
  - REQ-PROXY-0013: Asserts that `SetProxy` and `ClearProxy` return only after the running proxy attempt of the previous generation is terminal with an `incomplete` `proxy_reconfigured` result.
  - REQ-PROXY-0015: Asserts that the cancelled attempt records `proxy_reconfigured` and is not retried although its schedule allows a retry, and that no result contains the proxy generation, ports, address or `CONNECT`.
  - REQ-PROXY-0030: Asserts that a cleared proxy is not replaced by the direct path: the proxy task records `proxy_not_configured` without a runner call, and only a separately started run uses the new proxy.

<a id="UT.root.TestProxyCleanupIncomplete"></a>

### UT.root.TestProxyCleanupIncomplete

- Case: `UT.root.TestProxyCleanupIncomplete`
- Result: passed
- Test: `TestProxyCleanupIncomplete` in `proxy_lifecycle_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestProxyCleanupIncomplete$' .`
- Scope: Existing synthetic Go unit test TestProxyCleanupIncomplete in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0008: Asserts on the injected clock that when a proxy attempt's cleanup does not finish, `SetProxy` and `ClearProxy` wait the full cleanup bound, not 1 ms less, then return `cleanup_incomplete`, leave the client blocked with the previous configuration and no replacement installed and refuse a further change, and that the attempt's result is `proxy_reconfigured`.

<a id="UT.root.TestProxyHostConfigurationStaysSeparate"></a>

### UT.root.TestProxyHostConfigurationStaysSeparate

- Case: `UT.root.TestProxyHostConfigurationStaysSeparate`
- Result: passed
- Test: `TestProxyHostConfigurationStaysSeparate` in `proxy_lifecycle_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestProxyHostConfigurationStaysSeparate$' .`
- Scope: Existing synthetic Go unit test TestProxyHostConfigurationStaysSeparate in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0002: Asserts that the Go facade takes the proxy as JSON at `Initialize` and through `SetProxy` and `ClearProxy`. Binding exposure is not tested.
  - REQ-PROXY-0003: Asserts that initialization, replacement and clearing succeed in sequence through the Go facade. Equivalent binding semantics are not tested.
  - REQ-PROXY-0004: Asserts that the egress endpoints configured beside the proxy at `Initialize` are unchanged after `SetProxy` and `ClearProxy`.

<a id="UT.root.TestProxyLoopbackRenewal"></a>

### UT.root.TestProxyLoopbackRenewal

- Case: `UT.root.TestProxyLoopbackRenewal`
- Result: passed
- Test: `TestProxyLoopbackRenewal` in `proxy_lifecycle_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestProxyLoopbackRenewal$' .`
- Scope: Existing synthetic Go unit test TestProxyLoopbackRenewal in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0013: Asserts with real loopback sockets that `SetProxy` and `ClearProxy`, while a CONNECT is held on the old endpoint, cancel that exchange: the old endpoint then sees its connection closed and the run ends with `proxy_reconfigured`.
  - REQ-PROXY-0014: Asserts that, with the old endpoint kept listening until the call returns as the host must, renewal does not use it again: it accepts exactly one connection in all.
  - REQ-PROXY-0015: Asserts that the affected attempt records `proxy_reconfigured` and that neither the new endpoint nor the target directly receives a connection from it, and that only a separately admitted run uses the new endpoint, after `SetProxy` when the proxy was cleared.

<a id="UT.root.TestProxyReconfiguredBeforeTestStarts"></a>

### UT.root.TestProxyReconfiguredBeforeTestStarts

- Case: `UT.root.TestProxyReconfiguredBeforeTestStarts`
- Result: passed
- Test: `TestProxyReconfiguredBeforeTestStarts` in `proxy_lifecycle_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestProxyReconfiguredBeforeTestStarts$' .`
- Scope: Existing synthetic Go unit test TestProxyReconfiguredBeforeTestStarts in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0015: Asserts that `ClearProxy` during the context capture of a proxy attempt records `proxy_reconfigured` with no runner call.
  - REQ-RESULT-0051: Asserts that this attempt is `not_run`, because the test had not started.
  - REQ-RESULT-0053: Asserts that the `not_run` result carries no `network_context`.

<a id="UT.root.TestProxyTaskCannotReplaceHostConfiguration"></a>

### UT.root.TestProxyTaskCannotReplaceHostConfiguration

- Case: `UT.root.TestProxyTaskCannotReplaceHostConfiguration`
- Result: passed
- Test: `TestProxyTaskCannotReplaceHostConfiguration` in `proxy_lifecycle_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestProxyTaskCannotReplaceHostConfiguration$' .`
- Scope: Existing synthetic Go unit test TestProxyTaskCannotReplaceHostConfiguration in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0034: Asserts that a task adding `proxy`, `host`, `port`, `endpoint` or `connection_code` at top level or inside `input`, `options` or `execution` is rejected as `task_schema_invalid`, leaves the active proxy unchanged and reaches neither the path nor the runner.
  - REQ-PROXY-0038: Asserts that a task cannot supply `connection_code`, which comes only from approved host configuration (`CC-PROXY-CONFIGURATION-VALIDATION`).

<a id="UT.root.TestStartRunOnTheProxyPath"></a>

### UT.root.TestStartRunOnTheProxyPath

- Case: `UT.root.TestStartRunOnTheProxyPath`
- Result: passed
- Test: `TestStartRunOnTheProxyPath` in `proxy_path_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestStartRunOnTheProxyPath$' .`
- Scope: Existing synthetic Go unit test TestStartRunOnTheProxyPath in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0012: Asserts that a direct task on a client configured with a proxy records `network_path: direct` and makes no proxy connection.
  - REQ-PROXY-0022: Asserts that a proxy task runs through the proxy configured at `Initialize`, with one CONNECT for the runner's target and the target reached once through the tunnel.
  - REQ-PROXY-0031: Asserts that the proxied result and its bundle record `network_path: proxy` and `connection_code: outline_mobileproxy`, with egress unavailable and four egress members.
  - REQ-PROXY-0032: Asserts that the direct result has three egress members and `network_path: direct`.
  - REQ-PROXY-0033: Asserts that neither the result nor the bundle contains the proxy port, `127.0.0.1`, `CONNECT` or `HTTP/1.1`.

<a id="UT.root.TestProxyPathFailuresAreRecorded"></a>

### UT.root.TestProxyPathFailuresAreRecorded

- Case: `UT.root.TestProxyPathFailuresAreRecorded`
- Result: passed
- Test: `TestProxyPathFailuresAreRecorded` in `proxy_path_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestProxyPathFailuresAreRecorded$' .`
- Scope: Existing synthetic Go unit test TestProxyPathFailuresAreRecorded in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0008: Asserts through the facade that a proxy task without a configured proxy is started and records `not_run` with `proxy_not_configured`, without `network_context`, runner call or target contact.
  - REQ-EXECUTION-0010: Asserts that neither failure uses the direct path: the target accepts no connection.
  - REQ-PROXY-0028: Asserts `proxy_not_configured` for an unconfigured client and `incomplete` with `proxy_unavailable` and proxy egress for a configured endpoint with no listener.
  - REQ-PROXY-0030: Asserts that no direct contact follows either failure.

<a id="UT.internal.admission.TestGateAdmitsEgressContextRequirements"></a>

### UT.internal.admission.TestGateAdmitsEgressContextRequirements

- Case: `UT.internal.admission.TestGateAdmitsEgressContextRequirements`
- Result: passed
- Test: `TestGateAdmitsEgressContextRequirements` in `internal/admission/capability_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateAdmitsEgressContextRequirements$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateAdmitsEgressContextRequirements in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0018: Asserts that admitting tasks with egress context requirements, for embedded and remote origin, performs no context lookup: the harness's recording transport sees no call.
  - REQ-TRUST-0036: Asserts that valid tasks requiring `country_code`, `region_code` or `asn` individually or together, a country allowlist, or both, are admitted for embedded and remote origin rather than rejected as `task_unsupported`, while an unknown or duplicate field, an empty allowlist, a lowercase code, an unknown context member and region, ASN, provider and operator allowlists are `task_schema_invalid`.

<a id="UT.internal.admission.TestGateEgressContextKeepsOtherSteps"></a>

### UT.internal.admission.TestGateEgressContextKeepsOtherSteps

- Case: `UT.internal.admission.TestGateEgressContextKeepsOtherSteps`
- Result: passed
- Test: `TestGateEgressContextKeepsOtherSteps` in `internal/admission/capability_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateEgressContextKeepsOtherSteps$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateEgressContextKeepsOtherSteps in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0047: Asserts that for an unknown field, an over-long `task_id`, expiry, a value above a registered limit, invalid input, a test not permitted, policy rollback, a label namespace, a path not permitted, an unregistered target and a risk bound, the rejection class is the same with and without a `country_code` requirement.

<a id="UT.internal.admission.TestGateInputOptionsClasses"></a>

### UT.internal.admission.TestGateInputOptionsClasses

- Case: `UT.internal.admission.TestGateInputOptionsClasses`
- Result: passed
- Test: `TestGateInputOptionsClasses` in `internal/admission/input_options_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateInputOptionsClasses$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateInputOptionsClasses in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0056: Asserts through the gate with a synthetic test-scoped codec, for embedded origin and for two remote rows, that an unknown input or option member, an option outside its field range or of the wrong type and an input the codec cannot normalize are `task_schema_invalid`; that options over a stricter test byte cap, invalid options over that cap, an option value over its own stricter cap and a value of the wrong type over that cap are `task_limit_exceeded`, the cap being checked before the content it bounds; and that options exactly at the cap are admitted. The codec is not reached for a schema or limit failure. No registered Draft `0.1.0` test defines a stricter byte cap, so these rows are synthetic.
  - REQ-OPTIONS-0033: Asserts that invalid options are rejected at admission with the class REQ-TRUST-0056 names, so no such task is admitted to be skipped or to fail at execution. Options local policy does not permit and the `task_unsupported` row are not exercised.
  - REQ-TRUST-0047: Asserts that each rejected task receives exactly one class, and that invalid input decides the class ahead of over-limit options. That order follows REQ-TRUST-0059.
  - REQ-LIMIT-0002: Asserts that an `input` or option value above a synthetic registered numeric limit is rejected as `task_limit_exceeded` and not clamped, and that the value at the limit is admitted. The class follows REQ-TRUST-0058.
  - REQ-TRUST-0058: Asserts that an `input` or option value of the permitted type above a synthetic registered maximum is `task_limit_exceeded` and the value at the maximum is admitted, while a value outside a range only its field table defines, including an `input` value below its minimum, or of the wrong type is `task_schema_invalid`; the over-limit and out-of-range option rows also run for remote origin.
  - REQ-TRUST-0059: Asserts the public order: invalid `input` decides ahead of over-limit `options`, and over-limit `input` ahead of invalid `options`; over-limit or invalid `options` decide ahead of an `input` the codec cannot normalize; within `input` and within `options` the first failing member in presented order decides, whether over the limit, out of range or unknown; a missing required member is checked after the present members, so an over-limit present member decides and a missing member alone is `task_schema_invalid`; and a byte cap is checked before the content it bounds.

<a id="UT.internal.admission.TestGateCarriesAuthorizedRedirectSnapshot"></a>

### UT.internal.admission.TestGateCarriesAuthorizedRedirectSnapshot

- Case: `UT.internal.admission.TestGateCarriesAuthorizedRedirectSnapshot`
- Result: passed
- Test: `TestGateCarriesAuthorizedRedirectSnapshot` in `internal/admission/target_test.go` (package `internal/admission`)
- Reproduce one case: `go test -count=1 -run '^TestGateCarriesAuthorizedRedirectSnapshot$' ./internal/admission`
- Scope: Existing synthetic Go unit test TestGateCarriesAuthorizedRedirectSnapshot in package internal/admission, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0016: Asserts that an admitted HTTP-01 task carries into execution only its matched initial registration and the redirect registration that registration links and that issuer and local policy both permit, as an owned snapshot that a caller mutating the returned copy cannot change; a redirect registration the issuer does not list is not carried.
  - REQ-SAFETY-0029: Asserts that the baseline risk of a reachable, authorized redirect registration counts toward the effective risk, so a task bounded to low risk is rejected as `task_issuer_not_authorized` when its high-risk redirect is reachable rather than admitted with that risk ignored.

<a id="UT.internal.execution.TestTargetGateResolutionAndRegistration"></a>

### UT.internal.execution.TestTargetGateResolutionAndRegistration

- Case: `UT.internal.execution.TestTargetGateResolutionAndRegistration`
- Result: passed
- Test: `TestTargetGateResolutionAndRegistration` in `internal/execution/targets_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestTargetGateResolutionAndRegistration$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestTargetGateResolutionAndRegistration in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0016: Asserts on the production gate `Targets.Prepare` with a scripted resolver that a registered direct-path hostname is resolved once and pinned to the first checked public address, that a loopback-only, a mixed public and private, and an empty answer set are refused, that a registration of another host, an expired registration, a proxy-only registration and another test version are refused without resolution, that a registration expiring during resolution is refused, and that cancellation reaches the resolver and prevents contact.
  - REQ-PROXY-0037: Asserts on the direct path that the gate refuses loopback and private answers under the default address policy. The same refusal through the proxy path is asserted for a prohibited literal redirect by `UT.internal.execution.TestTargetGateRedirectFullTupleAndSharedScope`; the whole family row of identical rejection on both paths is not exercised.

<a id="UT.internal.execution.TestTargetGateProxyRefusesHostnameBeforeStartOrResolution"></a>

### UT.internal.execution.TestTargetGateProxyRefusesHostnameBeforeStartOrResolution

- Case: `UT.internal.execution.TestTargetGateProxyRefusesHostnameBeforeStartOrResolution`
- Result: passed
- Test: `TestTargetGateProxyRefusesHostnameBeforeStartOrResolution` in `internal/execution/targets_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestTargetGateProxyRefusesHostnameBeforeStartOrResolution$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestTargetGateProxyRefusesHostnameBeforeStartOrResolution in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0043: Asserts that on the proxy path the production gate refuses a registered hostname authority before the test starts and without calling the resolver, so no address is substituted.
  - REQ-PROXY-0046: Asserts that the refusal is `proxy_path_unsupported`, that a runner dialing the hostname through the attempt scope reaches no transport and does not start the test, and that the refusal terminates the attempt with `proxy_path_unsupported` even if the runner would report a terminal outcome.

<a id="UT.internal.execution.TestTargetGateRedirectFullTupleAndSharedScope"></a>

### UT.internal.execution.TestTargetGateRedirectFullTupleAndSharedScope

- Case: `UT.internal.execution.TestTargetGateRedirectFullTupleAndSharedScope`
- Result: passed
- Test: `TestTargetGateRedirectFullTupleAndSharedScope` in `internal/execution/targets_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestTargetGateRedirectFullTupleAndSharedScope$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestTargetGateRedirectFullTupleAndSharedScope in package internal/execution, executed with go test at the implementation baseline. Shared by both manifests; partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0016: Asserts on the direct and proxy paths that a redirect is authorized only by its complete registered tuple and its parent's link, that the root dialer cannot contact a redirect registration directly, that a redirect dialer cannot contact its parent's authority, that two hops share the attempt's connection and byte meter, that a registered hostname redirect is refused on the proxy path and that a prohibited literal redirect is refused locally before contact.
  - REQ-HTTP-01-0051: On the direct and proxy paths a redirect is authorized only for the complete registered tuple linked from the current target, the redirect dialer cannot reach another authority, hops share one meter and byte budget, and registered hostname redirects are refused on the proxy path.

<a id="UT.internal.execution.TestTargetScopeResolutionCannotBypassBounds"></a>

### UT.internal.execution.TestTargetScopeResolutionCannotBypassBounds

- Case: `UT.internal.execution.TestTargetScopeResolutionCannotBypassBounds`
- Result: passed
- Test: `TestTargetScopeResolutionCannotBypassBounds` in `internal/execution/targets_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestTargetScopeResolutionCannotBypassBounds$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestTargetScopeResolutionCannotBypassBounds in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0016: Asserts that a mixed public and loopback answer set is refused through the attempt scope with no transport call and no reserved connection.
  - REQ-LIMIT-0007: Asserts that with a zero connection limit the scope refuses the dial without calling the resolver or the transport and without starting the test.

<a id="UT.internal.execution.TestTargetSnapshotOwnershipAndAncestorExpiry"></a>

### UT.internal.execution.TestTargetSnapshotOwnershipAndAncestorExpiry

- Case: `UT.internal.execution.TestTargetSnapshotOwnershipAndAncestorExpiry`
- Result: passed
- Test: `TestTargetSnapshotOwnershipAndAncestorExpiry` in `internal/execution/targets_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestTargetSnapshotOwnershipAndAncestorExpiry$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestTargetSnapshotOwnershipAndAncestorExpiry in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0016: Asserts that mutating the caller's registrations after constructing the gate cannot change its owned snapshot, and that a redirect hop whose ancestor registration expired while it was pending is refused.

<a id="UT.internal.execution.TestContextFetchUsesLocalAddressPolicy"></a>

### UT.internal.execution.TestContextFetchUsesLocalAddressPolicy

- Case: `UT.internal.execution.TestContextFetchUsesLocalAddressPolicy`
- Result: passed
- Test: `TestContextFetchUsesLocalAddressPolicy` in `internal/execution/targets_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestContextFetchUsesLocalAddressPolicy$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestContextFetchUsesLocalAddressPolicy in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0045: Asserts that the context fetch checks an approved explicit-IP endpoint against the local address policy before the transport is called: under a policy that permits no address the fetch fails with no dial.

<a id="UT.internal.execution.TestCcProxyProductionTargetRefusal"></a>

### UT.internal.execution.TestCcProxyProductionTargetRefusal

- Case: `UT.internal.execution.TestCcProxyProductionTargetRefusal`
- Result: passed
- Test: `TestCcProxyProductionTargetRefusal` in `internal/execution/target_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcProxyProductionTargetRefusal$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestCcProxyProductionTargetRefusal in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0043: Asserts that a runner dialing a registered hostname authority on the proxy path through the production gate is refused with no connection to the scripted proxy endpoint.
  - REQ-PROXY-0046: Asserts that the per-dial refusal ends the attempt with `proxy_path_unsupported` and no measurement, whether the runner returns an error or swallows the refusal and reports a connected outcome, and that it is not recorded as a TCP or TLS failure.
  - REQ-RESULT-0051: Asserts that the refused attempt is `not_run` when the runner had not started the test and `incomplete` when it had.

<a id="UT.internal.execution.TestCcProxyProductionAllowedTarget"></a>

### UT.internal.execution.TestCcProxyProductionAllowedTarget

- Case: `UT.internal.execution.TestCcProxyProductionAllowedTarget`
- Result: passed
- Test: `TestCcProxyProductionAllowedTarget` in `internal/execution/target_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcProxyProductionAllowedTarget$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestCcProxyProductionAllowedTarget in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0016: Asserts that a registered explicit public address and port passes the production gate under the default address policy immediately before contact on the proxy path, with no loopback exemption: the scripted proxy routes it to a loopback echo server and the attempt completes.
  - REQ-PROXY-0043: Asserts that the registered explicit IP address and port is the CONNECT authority of exactly one CONNECT request.

<a id="UT.internal.execution.TestCcDirectProductionAllowedTarget"></a>

### UT.internal.execution.TestCcDirectProductionAllowedTarget

- Case: `UT.internal.execution.TestCcDirectProductionAllowedTarget`
- Result: passed
- Test: `TestCcDirectProductionAllowedTarget` in `internal/execution/target_conformance_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcDirectProductionAllowedTarget$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestCcDirectProductionAllowedTarget in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0016: Asserts that a registered explicit public address and port passes the production gate on the direct path under the default address policy, with no loopback exemption: a synthetic route delivers it to a loopback echo server and the attempt completes.

<a id="UT.internal.execution.TestProxyHostnameContextUnavailable"></a>

### UT.internal.execution.TestProxyHostnameContextUnavailable

- Case: `UT.internal.execution.TestProxyHostnameContextUnavailable`
- Result: passed
- Test: `TestProxyHostnameContextUnavailable` in `internal/execution/egress_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestProxyHostnameContextUnavailable$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestProxyHostnameContextUnavailable in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0045: Asserts that with hostname primary and fallback context endpoints a proxy attempt makes no connection to the proxy endpoint, no direct dial and no endpoint request; that egress is then unavailable, so a task without context requirements completes; and that a task requiring `country_code` records `not_run` with `network_context_unavailable` and no measurement.
  - REQ-CONTEXT-0043: Asserts that when neither endpoint is eligible on the proxy path, egress status is `unavailable` with no request to either endpoint.

<a id="UT.internal.networkcontext.TestProxyContextHostnameEligibility"></a>

### UT.internal.networkcontext.TestProxyContextHostnameEligibility

- Case: `UT.internal.networkcontext.TestProxyContextHostnameEligibility`
- Result: passed
- Test: `TestProxyContextHostnameEligibility` in `internal/networkcontext/egress_test.go` (package `internal/networkcontext`)
- Reproduce one case: `go test -count=1 -run '^TestProxyContextHostnameEligibility$' ./internal/networkcontext`
- Scope: Existing synthetic Go unit test TestProxyContextHostnameEligibility in package internal/networkcontext, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0045: Asserts that a hostname endpoint configuration stays valid, that on the proxy path a hostname primary is never passed to the transport, that an independently approved explicit-IP fallback is then requested once and supplies fresh fallback egress, that a hostname or loopback fallback is also ineligible, leaving egress unavailable with nothing cached, and that the same hostname primary is still requested on the direct path.
  - REQ-CONTEXT-0043: Asserts that an ineligible primary receives no request and is followed by at most one request to an eligible fallback, and that with no eligible endpoint egress is `unavailable` with no request.

<a id="UT.internal.registry.TestCommonProxyDestinationCompatibility"></a>

### UT.internal.registry.TestCommonProxyDestinationCompatibility

- Case: `UT.internal.registry.TestCommonProxyDestinationCompatibility`
- Result: passed
- Test: `TestCommonProxyDestinationCompatibility` in `internal/registry/registry_test.go` (package `internal/registry`)
- Reproduce one case: `go test -count=1 -run '^TestCommonProxyDestinationCompatibility$' ./internal/registry`
- Scope: Existing synthetic Go unit test TestCommonProxyDestinationCompatibility in package internal/registry, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0044: Asserts that `registry.ProxyCompatible`, with a rule that otherwise permits the parameters, rejects a measurement target whose URL host is a hostname, even when the tuple also carries an explicit address, and a hostname stream target, and accepts a literal-IP URL, a DNS query name, a resolver hostname with an explicit bootstrap address, and a hostname redirect registration, which is checked only when a hop is followed.

<a id="UT.root.TestMNet05AtomicTargetAndProxyRules"></a>

### UT.root.TestMNet05AtomicTargetAndProxyRules

- Case: `UT.root.TestMNet05AtomicTargetAndProxyRules`
- Result: passed
- Test: `TestMNet05AtomicTargetAndProxyRules` in `net05_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMNet05AtomicTargetAndProxyRules$' .`
- Scope: Existing synthetic Go unit test TestMNet05AtomicTargetAndProxyRules in the root package, executed with go test at the implementation baseline. Shared by both manifests; partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0039: Asserts through `Client.AdmitTask` with the production NET-05 codec that a proxy profile whose last step is DNS-01 in system mode or direct UDP, or which contains an HTTP-01 step with a hostname URL, is rejected as `task_schema_invalid`, also when path authority and target registration would refuse it, with no handle, replay write, run or transport call.
  - REQ-PROXY-0044: Asserts the NET-05 row: a profile with any incompatible step is rejected as a whole, and a profile of compatible DNS-01 direct TCP, TCP-01, TLS-01 and explicit-IP HTTP-01 steps is admitted on the proxy path with no transport call.
  - REQ-NET-05-0018: A changed final-step registration, a registration for NET-05 instead of the primitive, a proxy-only path, an expired registration or an issuer missing one step's registration rejects the whole profile as target_not_registered with no contact.
  - REQ-NET-05-0019: Each step is matched against its own primitive registration and authority; a missing primitive permission or composite permission is task_issuer_not_authorized, a system or UDP DNS-01 step or a hostname HTTP-01 step makes a proxy profile task_schema_invalid, and a high-risk target cannot be hidden inside a medium-risk composite.

<a id="UT.root.TestCcProxyOverviewDNSRow"></a>

### UT.root.TestCcProxyOverviewDNSRow

- Case: `UT.root.TestCcProxyOverviewDNSRow`
- Result: passed
- Test: `TestCcProxyOverviewDNSRow` in `dns01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcProxyOverviewDNSRow$' .`
- Scope: Existing synthetic Go unit test TestCcProxyOverviewDNSRow in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-PROXY-0044: Asserts the DNS-01 row through the facade with the production DNS-01 codec: proxy tasks in system mode and with direct UDP are rejected at admission as `task_schema_invalid` with no transport call and no connection to the proxy endpoint, while direct TCP is admitted and resolves through exactly one CONNECT tunnel with no direct transport call.
  - REQ-LIMIT-0034: Asserts that with `max_bytes_sent` 31 and `max_bytes_received` 47 the DNS-over-TCP exchange completes while the scripted proxy observes exactly 31 tunnel bytes in and 47 out, so the DNS messages with both two-byte length prefixes fit the limits exactly and the CONNECT exchange is not counted.

<a id="UT.root.TestClientProductionRegistryRejectsUnregisteredTasks"></a>

### UT.root.TestClientProductionRegistryRejectsUnregisteredTasks

- Case: `UT.root.TestClientProductionRegistryRejectsUnregisteredTasks`
- Result: passed
- Test: `TestClientProductionRegistryRejectsUnregisteredTasks` in `registry_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestClientProductionRegistryRejectsUnregisteredTasks$' .`
- Scope: Existing synthetic Go unit test TestClientProductionRegistryRejectsUnregisteredTasks in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0036: Asserts that a client initialized without an injected registry binds the production registry for admission and execution, and that a well-formed synthetic SYN-01 task, which the production registry does not register, is `task_unsupported` with no handle, replay write or run, while the same task is admitted under an injected synthetic registry.

<a id="UT.internal.execution.TestNet01AllTerminationBoundaries"></a>

### UT.internal.execution.TestNet01AllTerminationBoundaries

- Case: `UT.internal.execution.TestNet01AllTerminationBoundaries`
- Result: passed
- Test: `TestNet01AllTerminationBoundaries` in `internal/execution/net01_termination_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestNet01AllTerminationBoundaries$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestNet01AllTerminationBoundaries in package internal/execution, executed with go test at the implementation baseline. Shared by both manifests; partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-01-0027: At the engine boundary with the production runner, each of host cancellation, generation change, outer deadline and injected suspension, followed by a second condition, gives not_run before the read, incomplete with only the first reason and no measurement during or after the read, and completed snapshot after the outcome is recorded, with no top-level network_context, zero usage and dialer requests and no timer or watch left.
  - REQ-RESULT-0051: For every common cause, an interruption before NET-01's collection is not_run and one during or after the access read is incomplete with exactly that reason and no measurement.
  - REQ-RESULT-0052: An interruption at release after the snapshot outcome is recorded leaves the attempt completed with the snapshot measurement and no reason.
  - REQ-RESULT-0053: No NET-01 result at any stage carries network_context, matching its registered access-only representation.

<a id="UT.internal.execution.TestNet03AllTerminationBoundaries"></a>

### UT.internal.execution.TestNet03AllTerminationBoundaries

- Case: `UT.internal.execution.TestNet03AllTerminationBoundaries`
- Result: passed
- Test: `TestNet03AllTerminationBoundaries` in `internal/execution/net03_termination_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestNet03AllTerminationBoundaries$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestNet03AllTerminationBoundaries in package internal/execution, executed with go test at the implementation baseline. Shared by both manifests; partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-NET-03-0029: At the engine and production-runner boundary, for https and http, each of five causes before the start is not_run and at every later Procedure point up to the prefix read is incomplete with only the first reason and no measurement, with connections closed, no timers and goroutines back to baseline; release after the outcome stays completed.
  - REQ-RESULT-0051: For every common cause, an interruption before the NET-01 stability check is not_run, and from that check until the outcome is recorded it is incomplete with exactly the first reason and no measurement.
  - REQ-RESULT-0052: An interruption at release after the validated outcome is recorded leaves the attempt completed with that outcome and no reason.
  - REQ-RESULT-0053: The not_run attempt carries no network_context.

<a id="UT.root.TestMDns03ExecutionAndOutput"></a>

### UT.root.TestMDns03ExecutionAndOutput

- Case: `UT.root.TestMDns03ExecutionAndOutput`
- Result: passed
- Test: `TestMDns03ExecutionAndOutput` in `dns03_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMDns03ExecutionAndOutput$' .`
- Scope: Existing synthetic Go unit test TestMDns03ExecutionAndOutput in the root package, executed with go test at the implementation baseline. Shared by both manifests; partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-03-0029: The serialized result contains none of the resolver URL or path, resolver hostname, query name, bootstrap or loopback address, certificate, cookie, header, body, media-type or raw-error canaries.
  - REQ-DNS-03-0035: The result retains no response body, HTTP header, cookie, or resolver or bootstrap address canary.
  - REQ-DNS-03-0036: The recorded request is one POST with only Content-Type, Accept, Accept-Encoding, Content-Length and Connection headers and a DNS ID 0 query.
  - REQ-DNS-03-0037: The recorded request carries no cookie, authorization, user agent or other application header.
  - REQ-PROXY-0044: On the proxy path DNS-03 with an approved bootstrap address opens one CONNECT to the bootstrap authority with no direct dial and completes; without a bootstrap address the proxy-path task is rejected at admission as task_schema_invalid with no dial, lookup or proxy contact.

<a id="UT.internal.execution.TestDns03PlaintextAccounting"></a>

### UT.internal.execution.TestDns03PlaintextAccounting

- Case: `UT.internal.execution.TestDns03PlaintextAccounting`
- Result: passed
- Test: `TestDns03PlaintextAccounting` in `internal/execution/dns03_termination_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestDns03PlaintextAccounting$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestDns03PlaintextAccounting in package internal/execution, executed with go test at the implementation baseline. Shared by both manifests; partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-03-0005: Direct and proxied runs with a bootstrap address make one connection and one 31-byte DNS POST and complete with outcome response.
  - REQ-DNS-03-0023: Sent and received budgets equal to the measured request and response complete; one byte less ends the attempt with resource_limit and no measurement, before the request for the sent budget and after it for the received budget.
  - REQ-LIMIT-0034: On direct and proxy paths the metered sent and received bytes equal the HTTP plaintext the DoH server read and wrote (request line, headers and DNS body), TLS wire bytes and CONNECT are below the counted boundary, a sent budget equal to the request completes and one byte less ends incomplete with resource_limit before the request is complete.

<a id="UT.internal.execution.TestDns03TerminationPoints"></a>

### UT.internal.execution.TestDns03TerminationPoints

- Case: `UT.internal.execution.TestDns03TerminationPoints`
- Result: passed
- Test: `TestDns03TerminationPoints` in `internal/execution/dns03_termination_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestDns03TerminationPoints$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestDns03TerminationPoints in package internal/execution, executed with go test at the implementation baseline. Shared by both manifests; partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-03-0028: For every common cause at every Procedure point in both bootstrap variants, interruption before start is not_run without network_context, every point from bootstrap to the response body read is incomplete with the first reason, no measurement, no further dial or lookup and connections closed, and close after the outcome stays completed.
  - REQ-RESULT-0051: For seven causes, an interruption during context capture is not_run with no context, and during bootstrap, TCP connect, TLS or after the POST is read it is incomplete with exactly that reason and no measurement.
  - REQ-RESULT-0052: An interruption while closing after the response outcome is recorded leaves the attempt completed with that outcome and no reason.

<a id="UT.root.TestMDns04ExecutionAndOutput"></a>

### UT.root.TestMDns04ExecutionAndOutput

- Case: `UT.root.TestMDns04ExecutionAndOutput`
- Result: passed
- Test: `TestMDns04ExecutionAndOutput` in `dns04_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMDns04ExecutionAndOutput$' .`
- Scope: Existing synthetic Go unit test TestMDns04ExecutionAndOutput in the root package, executed with go test at the implementation baseline. Shared by both manifests; partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-04-0029: Successful results on the direct, system-bootstrap and proxy paths contain none of the resolver name, query name, bootstrap address, port, field names, certificate canary, raw-error canary, loopback address, alpn or ttl.
  - REQ-DNS-04-0036: Each run makes one connection and one standard ClientHello with the resolver name and no ALPN offer, and sends one framed query; no second handshake, port or resolver is contacted.
  - REQ-DNS-04-0037: The server receives exactly the 33-byte framed query as application plaintext; the test does not inspect its contents for identifier canaries.
  - REQ-PROXY-0044: On the proxy path DNS-04 with an approved bootstrap address opens one CONNECT to the bootstrap authority while the ClientHello names the resolver, with no direct dial; without a bootstrap address the proxy-path task is rejected at admission as task_schema_invalid with no contact.

<a id="UT.internal.execution.TestDns04PlaintextAccounting"></a>

### UT.internal.execution.TestDns04PlaintextAccounting

- Case: `UT.internal.execution.TestDns04PlaintextAccounting`
- Result: passed
- Test: `TestDns04PlaintextAccounting` in `internal/execution/dns04_termination_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestDns04PlaintextAccounting$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestDns04PlaintextAccounting in package internal/execution, executed with go test at the implementation baseline. Shared by both manifests; partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-04-0005: On the direct and proxy paths one connection completes a handshake naming the resolver with no ALPN, sends one 31-byte query with its prefix and reads one response, giving outcome response with three observations.
  - REQ-DNS-04-0023: Usage is one connection, 2 plus 31 bytes sent and 2 plus the response length received; a sent limit equal to the framed query succeeds and one byte below it is incomplete resource_limit with no plaintext delivered.
  - REQ-LIMIT-0034: On direct and proxy paths the meter records the DoT query and response with their two-byte length prefixes and excludes TLS bytes; a sent budget equal to the framed query completes and one byte less ends incomplete with resource_limit with no plaintext sent.

<a id="UT.internal.execution.TestDns04TerminationPoints"></a>

### UT.internal.execution.TestDns04TerminationPoints

- Case: `UT.internal.execution.TestDns04TerminationPoints`
- Result: passed
- Test: `TestDns04TerminationPoints` in `internal/execution/dns04_termination_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestDns04TerminationPoints$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestDns04TerminationPoints in package internal/execution, executed with go test at the implementation baseline. Shared by both manifests; partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-DNS-04-0028: Each of seven interruption causes at six Procedure points gives not_run before start without network context, incomplete with exactly that reason and no measurement during bootstrap, connect, TLS and after the query, and completed response while closing, with expected dial and lookup counts, every connection closed and no timer left.
  - REQ-RESULT-0051: For seven causes, an interruption during context capture is not_run with no context, and during bootstrap, connect, TLS or after the query is written it is incomplete with exactly that reason and no measurement.
  - REQ-RESULT-0052: An interruption while closing after the response outcome is recorded leaves the attempt completed with that outcome and no reason.

<a id="UT.root.TestCcLinkageRotationLimits"></a>

### UT.root.TestCcLinkageRotationLimits

- Case: `UT.root.TestCcLinkageRotationLimits`
- Result: passed
- Test: `TestCcLinkageRotationLimits` in `linkage_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestCcLinkageRotationLimits$' .`
- Scope: Existing synthetic Go unit test TestCcLinkageRotationLimits in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0023: Linkage rotation periods of 86,399 and 604,801 seconds are rejected at Initialize before any key operation, and 86,400 and 604,800 are accepted; absent configuration yields no linkage.
  - REQ-LINKAGE-0003: rotation_period_seconds bounds 86,399/604,801 rejected and 86,400/604,800 accepted.
  - REQ-LINKAGE-0004: Absent result_linkage and mode none yield bundles without linkage and no key-store call even with a key store and permissive policy bound.
  - REQ-LINKAGE-0011: Bundle epoch equals the trusted-clock Unix time divided by each accepted period.
  - REQ-LINKAGE-0017: Periods outside 86,400..604,800 are rejected.
  - REQ-LINKAGE-0043: With no or none configuration, and with policy disallowing linkage, no bundle carries linkage and the key store receives no derivation.

<a id="UT.root.TestClearRemovesLinkage"></a>

### UT.root.TestClearRemovesLinkage

- Case: `UT.root.TestClearRemovesLinkage`
- Result: passed
- Test: `TestClearRemovesLinkage` in `linkage_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestClearRemovesLinkage$' .`
- Scope: Existing synthetic Go unit test TestClearRemovesLinkage in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LINKAGE-0023: ClearMeasurementData deletes the key once and creates no replacement, reinitialization creates no secret, and the next bundle in the same epoch carries a different token derived under a key created only when linkage is next required.
  - REQ-LINKAGE-0024: Clearing with linkage enabled returns success and leaves the client reinitializable; repeat safety and absence of transport are not asserted here (they are in CC-LINKAGE-CLEAR-DATA-CONTRACT).

<a id="UT.root.TestLinkageStorageFailure"></a>

### UT.root.TestLinkageStorageFailure

- Case: `UT.root.TestLinkageStorageFailure`
- Result: passed
- Test: `TestLinkageStorageFailure` in `linkage_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestLinkageStorageFailure$' .`
- Scope: Existing synthetic Go unit test TestLinkageStorageFailure in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LINKAGE-0037: With the key store failing, initialization succeeds and the run's result is delivered in a bundle with a bundle_id and no linkage; linkage returns after recovery.
  - REQ-LINKAGE-0038: A token derived before a key-store failure is not reused in the next bundle (no retained constant token), and clearing with a failing key store reports an explicit error while pending results are deleted.

<a id="UT.root.TestLinkageSecretNotExposed"></a>

### UT.root.TestLinkageSecretNotExposed

- Case: `UT.root.TestLinkageSecretNotExposed`
- Result: passed
- Test: `TestLinkageSecretNotExposed` in `linkage_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestLinkageSecretNotExposed$' .`
- Scope: Existing synthetic Go unit test TestLinkageSecretNotExposed in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LINKAGE-0008: A synthetic secret in the test key store appears in no raw, hex, base64 or base64url form in any string or error returned by Initialize, AdmitTask, StartRun, RunState, SetProxy, NextBundle, AcknowledgeBundle, ClearMeasurementData or Close, nor in the library-owned record store.
  - REQ-LINKAGE-0019: Every exported Client method takes and returns only strings and errors, so no key store, key handle or storage object can be passed in or out, and platform.LinkageKeyStore has two methods whose derivation returns a 32-byte HMAC output, not the secret.
  - REQ-LINKAGE-0046: The secret is absent from every returned value and error and the token appears only once, as linkage.token in the bundle; logs, crash reports and analytics hooks are not exercised (the root package has none to record).

<a id="UT.internal.persistence.linkage.TestDisabledByDefault"></a>

### UT.internal.persistence.linkage.TestDisabledByDefault

- Case: `UT.internal.persistence.linkage.TestDisabledByDefault`
- Result: passed
- Test: `TestDisabledByDefault` in `internal/persistence/linkage/linkage_test.go` (package `internal/persistence/linkage`)
- Reproduce one case: `go test -count=1 -run '^TestDisabledByDefault$' ./internal/persistence/linkage`
- Scope: Existing synthetic Go unit test TestDisabledByDefault in package internal/persistence/linkage, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LINKAGE-0004: Absent configuration and mode none give no derivation.
  - REQ-LINKAGE-0043: Absent, none, policy-disabled and no-key-store initializations, a nil source and a nil initialization give no derivation, and the key store receives no message.
  - REQ-LINKAGE-0044: A rotating_installation request yields no derivation when trusted approvals disallow linkage or no protected-key capability is bound.

<a id="UT.internal.persistence.linkage.TestFailuresOmitLinkage"></a>

### UT.internal.persistence.linkage.TestFailuresOmitLinkage

- Case: `UT.internal.persistence.linkage.TestFailuresOmitLinkage`
- Result: passed
- Test: `TestFailuresOmitLinkage` in `internal/persistence/linkage/linkage_test.go` (package `internal/persistence/linkage`)
- Reproduce one case: `go test -count=1 -run '^TestFailuresOmitLinkage$' ./internal/persistence/linkage`
- Scope: Existing synthetic Go unit test TestFailuresOmitLinkage in package internal/persistence/linkage, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LINKAGE-0014: An unavailable clock and a pre-1970 time omit linkage without a key call.
  - REQ-LINKAGE-0037: A key-store failure and a cancelled context omit linkage, and linkage is available again after recovery.
  - REQ-LINKAGE-0038: Failures return no linkage rather than a fallback value.

<a id="UT.internal.persistence.linkage.TestEpochFloor"></a>

### UT.internal.persistence.linkage.TestEpochFloor

- Case: `UT.internal.persistence.linkage.TestEpochFloor`
- Result: passed
- Test: `TestEpochFloor` in `internal/persistence/linkage/linkage_test.go` (package `internal/persistence/linkage`)
- Reproduce one case: `go test -count=1 -run '^TestEpochFloor$' ./internal/persistence/linkage`
- Scope: Existing synthetic Go unit test TestEpochFloor in package internal/persistence/linkage, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LINKAGE-0014: A clock moved back one epoch omits linkage, including through a later initialization of the same client, until the floor epoch returns; another period has its own floor.
  - REQ-LINKAGE-0023: Source.Erase drops the epoch floor so a later derivation may use an earlier epoch.

<a id="UT.internal.persistence.linkage.TestEraseDuringDerivation"></a>

### UT.internal.persistence.linkage.TestEraseDuringDerivation

- Case: `UT.internal.persistence.linkage.TestEraseDuringDerivation`
- Result: passed
- Test: `TestEraseDuringDerivation` in `internal/persistence/linkage/linkage_test.go` (package `internal/persistence/linkage`)
- Reproduce one case: `go test -count=1 -run '^TestEraseDuringDerivation$' ./internal/persistence/linkage`
- Scope: Existing synthetic Go unit test TestEraseDuringDerivation in package internal/persistence/linkage, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LINKAGE-0023: A clear (Erase plus key Delete) during a derivation discards that token, and the next derivation uses a newly created key and differs from the old token.

<a id="UT.internal.persistence.pending.TestBundleLinkage"></a>

### UT.internal.persistence.pending.TestBundleLinkage

- Case: `UT.internal.persistence.pending.TestBundleLinkage`
- Result: passed
- Test: `TestBundleLinkage` in `internal/persistence/pending/linkage_test.go` (package `internal/persistence/pending`)
- Reproduce one case: `go test -count=1 -run '^TestBundleLinkage$' ./internal/persistence/pending`
- Scope: Existing synthetic Go unit test TestBundleLinkage in package internal/persistence/pending, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LINKAGE-0015: Results added to an open bundle keep the epoch and token derived when it was created; only a new bundle derives again and carries the new epoch.
  - REQ-LINKAGE-0037: An unavailable derivation, or a projection the builder rejects, omits linkage while the result is still stored and returned in a bundle.

<a id="UT.internal.persistence.pending.TestLinkageReserve"></a>

### UT.internal.persistence.pending.TestLinkageReserve

- Case: `UT.internal.persistence.pending.TestLinkageReserve`
- Result: passed
- Test: `TestLinkageReserve` in `internal/persistence/pending/linkage_test.go` (package `internal/persistence/pending`)
- Reproduce one case: `go test -count=1 -run '^TestLinkageReserve$' ./internal/persistence/pending`
- Scope: Existing synthetic Go unit test TestLinkageReserve in package internal/persistence/pending, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0021: A linked pending store refuses a byte limit lacking room for one 65,536-byte result plus the bundle overhead and linkage reserve, and at the default aggregate pending-storage limit an empty linked store reports capacity 63.

<a id="UT.root.TestDns01RechecksBothRegistrationExpiries"></a>

### UT.root.TestDns01RechecksBothRegistrationExpiries

- Case: `UT.root.TestDns01RechecksBothRegistrationExpiries`
- Result: passed
- Test: `TestDns01RechecksBothRegistrationExpiries` in `dns01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestDns01RechecksBothRegistrationExpiries$' .`
- Scope: Existing synthetic Go unit test TestDns01RechecksBothRegistrationExpiries in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TRUST-0016: When either the hostname registration or the resolver registration expires between admission and StartRun, the run ends incomplete with local_policy_rejected and the resolver is never contacted (UDP).

<a id="UT.root.TestDns01StrictSchemaBothOrigins"></a>

### UT.root.TestDns01StrictSchemaBothOrigins

- Case: `UT.root.TestDns01StrictSchemaBothOrigins`
- Result: passed
- Test: `TestDns01StrictSchemaBothOrigins` in `dns01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestDns01StrictSchemaBothOrigins$' .`
- Scope: Existing synthetic Go unit test TestDns01StrictSchemaBothOrigins in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0002: Valid \u escapes in DNS-01 input member names (\u0068ostname, \u0070ort) are decoded before field matching for embedded and signed remote delivery; invalid UTF-8, malformed escapes and surrogates are not exercised here.
  - REQ-JSON-0003: Repeated input member names, including an escaped spelling of the same name and a repeat inside the nested resolver object, reject the task as task_schema_invalid with no handle, replay commit or contact, for both origins.
  - REQ-JSON-0004: An escaped spelling of a fixed field is accepted as that field and an unknown nested member (\u0078) is rejected rather than discarded, for both origins.

<a id="UT.root.TestMDns01ConnectionBudget"></a>

### UT.root.TestMDns01ConnectionBudget

- Case: `UT.root.TestMDns01ConnectionBudget`
- Result: passed
- Test: `TestMDns01ConnectionBudget` in `dns01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMDns01ConnectionBudget$' .`
- Scope: Existing synthetic Go unit test TestMDns01ConnectionBudget in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0006: With max_connections 1 and 2, a DNS-01 UDP or TCP query through the Client completes as resolved with exactly one transport dial.

<a id="UT.root.TestMDns01Options"></a>

### UT.root.TestMDns01Options

- Case: `UT.root.TestMDns01Options`
- Result: passed
- Test: `TestMDns01Options` in `dns01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMDns01Options$' .`
- Scope: Existing synthetic Go unit test TestMDns01Options in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0014: additional_response_window_ms of 1 and 1,000 is admitted and 1,001 or 10,001 is rejected as task_limit_exceeded with no contact or replay commit (the 1,000 ms observation-window row).

<a id="UT.root.TestMDns03Limits"></a>

### UT.root.TestMDns03Limits

- Case: `UT.root.TestMDns03Limits`
- Result: passed
- Test: `TestMDns03Limits` in `dns03_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMDns03Limits$' .`
- Scope: Existing synthetic Go unit test TestMDns03Limits in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0034: A 65,535-byte DoH message within the 65,536-byte DNS bound but with max_bytes_received 65,536 ends incomplete with resource_limit, showing the HTTP framing of the DoH exchange counts toward the received limit and the narrower message bound does not exempt it.

<a id="UT.root.TestMDns03Termination"></a>

### UT.root.TestMDns03Termination

- Case: `UT.root.TestMDns03Termination`
- Result: passed
- Test: `TestMDns03Termination` in `dns03_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMDns03Termination$' .`
- Scope: Existing synthetic Go unit test TestMDns03Termination in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0051: network_changed or host cancellation after the DoH POST is read, and a refused CONNECT tunnel, end the attempt incomplete with that reason (not not_run) and no measurement, with no direct or system-resolver fallback.

<a id="UT.root.TestMDns04Limits"></a>

### UT.root.TestMDns04Limits

- Case: `UT.root.TestMDns04Limits`
- Result: passed
- Test: `TestMDns04Limits` in `dns04_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMDns04Limits$' .`
- Scope: Existing synthetic Go unit test TestMDns04Limits in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0034: A 65,535-byte DoT message completes with max_bytes_received 65,537 and ends incomplete with resource_limit at 65,536, showing the two-byte length prefix counts as received; one connection completes with max_bytes_sent 33 (prefix plus query).

<a id="UT.root.TestMDns04Termination"></a>

### UT.root.TestMDns04Termination

- Case: `UT.root.TestMDns04Termination`
- Result: passed
- Test: `TestMDns04Termination` in `dns04_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMDns04Termination$' .`
- Scope: Existing synthetic Go unit test TestMDns04Termination in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0051: network_changed or host cancellation after the length-prefixed query is written, and a refused CONNECT tunnel, end the attempt incomplete with that reason (not not_run) and no measurement, with no fallback.

<a id="UT.root.TestMHttp01OuterBudgets"></a>

### UT.root.TestMHttp01OuterBudgets

- Case: `UT.root.TestMHttp01OuterBudgets`
- Result: passed
- Test: `TestMHttp01OuterBudgets` in `http01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMHttp01OuterBudgets$' .`
- Scope: Existing synthetic Go unit test TestMHttp01OuterBudgets in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0034: max_bytes_sent 1, max_bytes_received 4,096 against a larger response, and max_connections 1 with a followed redirect each end the HTTP-01 attempt incomplete with resource_limit and the server connection closed.

<a id="UT.root.TestMHttp01Termination"></a>

### UT.root.TestMHttp01Termination

- Case: `UT.root.TestMHttp01Termination`
- Result: passed
- Test: `TestMHttp01Termination` in `http01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMHttp01Termination$' .`
- Scope: Existing synthetic Go unit test TestMHttp01Termination in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0051: Host cancellation and outer-deadline expiry after the HTTP request was read end the attempt incomplete with that reason, not not_run.
  - REQ-RESULT-0052: A protocol operation timeout records the http_timeout outcome and the attempt is completed with no execution reason; later events after the outcome are not exercised here.

<a id="UT.internal.execution.TestCcLimitPurposeHTTPRows"></a>

### UT.internal.execution.TestCcLimitPurposeHTTPRows

- Case: `UT.internal.execution.TestCcLimitPurposeHTTPRows`
- Result: passed
- Test: `TestCcLimitPurposeHTTPRows` in `internal/execution/http_primitive_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestCcLimitPurposeHTTPRows$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestCcLimitPurposeHTTPRows in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0033: On direct and proxy paths the server's TLS wire bytes exceed the metered plaintext bytes and the proxy tunnel carries the wire bytes plus a CONNECT head, so byte limits are not on-wire limits.
  - REQ-LIMIT-0034: Metered sent bytes equal the request line, headers and optional POST body the server read and received bytes equal the full response; budgets equal to that usage complete with identical usage and one byte less in either direction ends the exchange with the budget cause.

<a id="UT.internal.execution.TestHTTPBodyCancellationAndDeadline"></a>

### UT.internal.execution.TestHTTPBodyCancellationAndDeadline

- Case: `UT.internal.execution.TestHTTPBodyCancellationAndDeadline`
- Result: passed
- Test: `TestHTTPBodyCancellationAndDeadline` in `internal/execution/http_primitive_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestHTTPBodyCancellationAndDeadline$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestHTTPBodyCancellationAndDeadline in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0008: A stalled body read ends after the 15,000 ms operation bound with the timeout outcome, and host cancellation or outer deadline unblocks it within a 2 s wait, on direct and proxy paths.
  - REQ-OBS-0023: The operation bound during the body yields the HTTP timeout outcome, while an outer interruption produces an error and no HTTP outcome.

<a id="UT.internal.execution.TestScopeSessionMetersPlaintext"></a>

### UT.internal.execution.TestScopeSessionMetersPlaintext

- Case: `UT.internal.execution.TestScopeSessionMetersPlaintext`
- Result: passed
- Test: `TestScopeSessionMetersPlaintext` in `internal/execution/tls01_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestScopeSessionMetersPlaintext$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestScopeSessionMetersPlaintext in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0034: After a session is taken, raw TLS-side bytes are not metered and plaintext-side writes are, and the plain stream refuses further I/O.
  - REQ-LIMIT-0027: A plaintext write exceeding the remaining sent budget is not issued, the attempt context is cancelled with the budget cause, and the raw stream becomes unusable.

<a id="UT.internal.persistence.pending.TestStrictReread"></a>

### UT.internal.persistence.pending.TestStrictReread

- Case: `UT.internal.persistence.pending.TestStrictReread`
- Result: passed
- Test: `TestStrictReread` in `internal/persistence/pending/pending_test.go` (package `internal/persistence/pending`)
- Reproduce one case: `go test -count=1 -run '^TestStrictReread$' ./internal/persistence/pending`
- Scope: Existing synthetic Go unit test TestStrictReread in package internal/persistence/pending, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0003: A stored bundle whose execution object repeats state with an equal value is refused on re-read: the store reports unavailable and the stored record is neither rewritten nor deleted.

<a id="UT.internal.protocol.dns.TestWireAnswerAndCNAMEBounds"></a>

### UT.internal.protocol.dns.TestWireAnswerAndCNAMEBounds

- Case: `UT.internal.protocol.dns.TestWireAnswerAndCNAMEBounds`
- Result: passed
- Test: `TestWireAnswerAndCNAMEBounds` in `internal/protocol/dns/wire_test.go` (package `internal/protocol/dns`)
- Reproduce one case: `go test -count=1 -run '^TestWireAnswerAndCNAMEBounds$' ./internal/protocol/dns`
- Scope: Existing synthetic Go unit test TestWireAnswerAndCNAMEBounds in package internal/protocol/dns, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0015: A response with 17 answers or 9 intervening CNAMEs (and a CNAME cycle) fails parsing as a whole rather than retaining a truncated answer list; 16 answers and 8 CNAMEs parse.

<a id="UT.internal.protocol.dns.TestWireEffectiveCodeVocabulary"></a>

### UT.internal.protocol.dns.TestWireEffectiveCodeVocabulary

- Case: `UT.internal.protocol.dns.TestWireEffectiveCodeVocabulary`
- Result: passed
- Test: `TestWireEffectiveCodeVocabulary` in `internal/protocol/dns/wire_test.go` (package `internal/protocol/dns`)
- Reproduce one case: `go test -count=1 -run '^TestWireEffectiveCodeVocabulary$' ./internal/protocol/dns`
- Scope: Existing synthetic Go unit test TestWireEffectiveCodeVocabulary in package internal/protocol/dns, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0046: Every effective response code 0..4095, including extended codes whose low header bits equal a retained code, maps to NOERROR, FORMERR, SERVFAIL, NXDOMAIN, NOTIMP, REFUSED for 0..5 and OTHER otherwise.

<a id="UT.internal.protocol.http.TestCcLimitHttpAndValidationLimits"></a>

### UT.internal.protocol.http.TestCcLimitHttpAndValidationLimits

- Case: `UT.internal.protocol.http.TestCcLimitHttpAndValidationLimits`
- Result: passed
- Test: `TestCcLimitHttpAndValidationLimits` in `internal/protocol/http/exchange_test.go` (package `internal/protocol/http`)
- Reproduce one case: `go test -count=1 -run '^TestCcLimitHttpAndValidationLimits$' ./internal/protocol/http`
- Scope: Existing synthetic Go unit test TestCcLimitHttpAndValidationLimits in package internal/protocol/http, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0016: A response header block at 16,384 bytes is accepted and one byte above is not; a 2,049-byte URL, body bound 0 or 65,537, 3 redirects, 9 header or body checks and a 257-byte matcher are rejected before any contact.
  - REQ-LIMIT-0017: An oversized header block does not produce expected_response.

<a id="UT.internal.protocol.http.TestCcObsHttpExchangeBody"></a>

### UT.internal.protocol.http.TestCcObsHttpExchangeBody

- Case: `UT.internal.protocol.http.TestCcObsHttpExchangeBody`
- Result: passed
- Test: `TestCcObsHttpExchangeBody` in `internal/protocol/http/exchange_test.go` (package `internal/protocol/http`)
- Reproduce one case: `go test -count=1 -run '^TestCcObsHttpExchangeBody$' ./internal/protocol/http`
- Scope: Existing synthetic Go unit test TestCcObsHttpExchangeBody in package internal/protocol/http, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0028: For 4,096- and 65,536-byte bounds, bytes_processed is min(length, bound) and truncated is true only above the bound, where the outcome is body_limit_reached.
  - REQ-OBS-0029: The serialized result contains neither the body digest nor the response-header canary.

<a id="UT.internal.protocol.http.TestCcObsHttpExchangeCheck"></a>

### UT.internal.protocol.http.TestCcObsHttpExchangeCheck

- Case: `UT.internal.protocol.http.TestCcObsHttpExchangeCheck`
- Result: passed
- Test: `TestCcObsHttpExchangeCheck` in `internal/protocol/http/exchange_test.go` (package `internal/protocol/http`)
- Reproduce one case: `go test -count=1 -run '^TestCcObsHttpExchangeCheck$' ./internal/protocol/http`
- Scope: Existing synthetic Go unit test TestCcObsHttpExchangeCheck in package internal/protocol/http, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0030: Status, header and body checks each produce one check entry with matched set, and failing checks select unexpected_status, header_check_failed or body_check_failed.

<a id="UT.internal.protocol.http.TestCcObsHttpExchangeDetails"></a>

### UT.internal.protocol.http.TestCcObsHttpExchangeDetails

- Case: `UT.internal.protocol.http.TestCcObsHttpExchangeDetails`
- Result: passed
- Test: `TestCcObsHttpExchangeDetails` in `internal/protocol/http/exchange_test.go` (package `internal/protocol/http`)
- Reproduce one case: `go test -count=1 -run '^TestCcObsHttpExchangeDetails$' ./internal/protocol/http`
- Scope: Existing synthetic Go unit test TestCcObsHttpExchangeDetails in package internal/protocol/http, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0025: HTTP details are constructible for address_family ipv4/ipv6/unknown and protocol http/1.1/h2/unknown, reject an empty or h3 protocol and a missing family, and carry 17 check entries at the maximum checks.

<a id="UT.internal.protocol.http.TestCcObsHttpExchangeFailures"></a>

### UT.internal.protocol.http.TestCcObsHttpExchangeFailures

- Case: `UT.internal.protocol.http.TestCcObsHttpExchangeFailures`
- Result: passed
- Test: `TestCcObsHttpExchangeFailures` in `internal/protocol/http/exchange_test.go` (package `internal/protocol/http`)
- Reproduce one case: `go test -count=1 -run '^TestCcObsHttpExchangeFailures$' ./internal/protocol/http`
- Scope: Existing synthetic Go unit test TestCcObsHttpExchangeFailures in package internal/protocol/http, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0023: Malformed status lines, invalid status, short bodies, truncated chunking and 101 map to protocol_error, and gzip content encoding maps to unsupported_content_encoding with no body summary.
  - REQ-OBS-0024: When the endpoint reports a DNS, TCP or TLS stage failure, the exchange emits no HTTP observation and returns that stage outcome.

<a id="UT.internal.protocol.http.TestCcObsHttpExchangeRedirect"></a>

### UT.internal.protocol.http.TestCcObsHttpExchangeRedirect

- Case: `UT.internal.protocol.http.TestCcObsHttpExchangeRedirect`
- Result: passed
- Test: `TestCcObsHttpExchangeRedirect` in `internal/protocol/http/exchange_test.go` (package `internal/protocol/http`)
- Reproduce one case: `go test -count=1 -run '^TestCcObsHttpExchangeRedirect$' ./internal/protocol/http`
- Scope: Existing synthetic Go unit test TestCcObsHttpExchangeRedirect in package internal/protocol/http, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0026: Each Location class (registered_https, unregistered_https, prohibited_host, non_https, credentials, invalid) is recorded as terminal_location_class, followed_count is 1 after one approved follow and 2 at the redirect limit.
  - REQ-OBS-0027: The serialized result never contains the redirect host or credentials, including for redirects that are classified but not contacted.
  - REQ-LIMIT-0018: With MaxRedirects 0 the exchange makes one contact, reports redirect_observed and retains only the minimized redirect class.

<a id="UT.internal.protocol.tls.TestHandshakeDoesNotInferPeerAlertFromErrors"></a>

### UT.internal.protocol.tls.TestHandshakeDoesNotInferPeerAlertFromErrors

- Case: `UT.internal.protocol.tls.TestHandshakeDoesNotInferPeerAlertFromErrors`
- Result: passed
- Test: `TestHandshakeDoesNotInferPeerAlertFromErrors` in `internal/protocol/tls/vocabularies_test.go` (package `internal/protocol/tls`)
- Reproduce one case: `go test -count=1 -run '^TestHandshakeDoesNotInferPeerAlertFromErrors$' ./internal/protocol/tls`
- Scope: Existing synthetic Go unit test TestHandshakeDoesNotInferPeerAlertFromErrors in package internal/protocol/tls, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0045: Raw error text, an opaque remote error and a local alert value become failure inconclusive with no alert, version or ALPN rather than being repaired into a vocabulary value.
  - REQ-OBS-0050: No details.alert is retained when there is no directly received peer alert, and no raw error text appears in the serialized observation.

<a id="UT.root.TestMNet03Limits"></a>

### UT.root.TestMNet03Limits

- Case: `UT.root.TestMNet03Limits`
- Result: passed
- Test: `TestMNet03Limits` in `net03_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMNet03Limits$' .`
- Scope: Existing synthetic Go unit test TestMNet03Limits in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0016: The NET-03 validation prefix is processed up to 1,024 bytes (bytes_processed and truncated) with a digest not decidable within it giving unexpected; a response header block at 16,384 bytes is valid and 16,385 yields no response class.
  - REQ-LIMIT-0034: A received budget equal to the response status line, headers and body completes and one byte less ends incomplete with resource_limit.

<a id="UT.root.TestMNet03Termination"></a>

### UT.root.TestMNet03Termination

- Case: `UT.root.TestMNet03Termination`
- Result: passed
- Test: `TestMNet03Termination` in `net03_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMNet03Termination$' .`
- Scope: Existing synthetic Go unit test TestMNet03Termination in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0051: A generation change during NET-03's own platform read, host cancellation and outer deadline after the request end incomplete with that reason and no measurement.
  - REQ-RESULT-0052: A bounded protocol timeout records not_validated and the attempt is completed with no reason; later events after the outcome are not exercised here.

<a id="UT.root.TestRuntimeDNSVocabularies"></a>

### UT.root.TestRuntimeDNSVocabularies

- Case: `UT.root.TestRuntimeDNSVocabularies`
- Result: passed
- Test: `TestRuntimeDNSVocabularies` in `runtime_vocabularies_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestRuntimeDNSVocabularies$' .`
- Scope: Existing synthetic Go unit test TestRuntimeDNSVocabularies in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0042: Each DNS-01 UDP/TCP result carries one response record with rcode and no answers that round-trips through the strict result decoder.
  - REQ-OBS-0045: The result decoder rejects rcode values BADVERS, noerror, "16", 16, PRIVATE-RCODE and null.
  - REQ-OBS-0046: Real responses with codes 0..6, 16, 18, 23, 3841 and 4095 are retained as the public table's label, OTHER beyond 5.
  - REQ-OBS-0047: FORMERR, NOTIMP and OTHER map to the inconclusive outcome and failure, never to no_data or success.

<a id="UT.root.TestRuntimeTLSNegotiatedVocabularies"></a>

### UT.root.TestRuntimeTLSNegotiatedVocabularies

- Case: `UT.root.TestRuntimeTLSNegotiatedVocabularies`
- Result: passed
- Test: `TestRuntimeTLSNegotiatedVocabularies` in `runtime_vocabularies_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestRuntimeTLSNegotiatedVocabularies$' .`
- Scope: Existing synthetic Go unit test TestRuntimeTLSNegotiatedVocabularies in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0043: Successful TLS-01 handshakes round-trip through the strict decoder with alert absent and alpn absent when nothing was offered.
  - REQ-OBS-0048: Negotiated TLS 1.2 and 1.3 are retained as TLS1.2 and TLS1.3, and the decoder rejects other tls_version values.
  - REQ-OBS-0049: A negotiated http/1.1, h2 or dot is retained exactly, and the decoder rejects H2, unregistered strings, numbers and null for alpn.
  - REQ-OBS-0052: For TLS 1.2 and TLS 1.3 handshakes against a loopback server, an offered and negotiated `http/1.1`, `h2` or `dot` is retained as exactly that `alpn`, the `alpn` member is absent when nothing is offered, and the decoder rejects a result whose `alpn` is replaced by a canary string, `H2`, `TLS1.0`, a number or null.

<a id="UT.root.TestRuntimeTLSPeerAlertVocabularies"></a>

### UT.root.TestRuntimeTLSPeerAlertVocabularies

- Case: `UT.root.TestRuntimeTLSPeerAlertVocabularies`
- Result: passed
- Test: `TestRuntimeTLSPeerAlertVocabularies` in `runtime_vocabularies_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestRuntimeTLSPeerAlertVocabularies$' .`
- Scope: Existing synthetic Go unit test TestRuntimeTLSPeerAlertVocabularies in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0043: An alert failure omits tls_version and alpn and round-trips through the strict decoder.
  - REQ-OBS-0050: Every listed peer description code received as a real TLS record maps to its table class, codes 1, 200 and 255 map to other, and the decoder rejects raw names, canaries, numeric codes and null.

<a id="UT.root.TestRuntimeTLSRejectsUnofferedALPN"></a>

### UT.root.TestRuntimeTLSRejectsUnofferedALPN

- Case: `UT.root.TestRuntimeTLSRejectsUnofferedALPN`
- Result: passed
- Test: `TestRuntimeTLSRejectsUnofferedALPN` in `runtime_vocabularies_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestRuntimeTLSRejectsUnofferedALPN$' .`
- Scope: Existing synthetic Go unit test TestRuntimeTLSRejectsUnofferedALPN in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0049: A peer selecting an unoffered ALPN yields an inconclusive TLS failure with no alpn, tls_version or alert and the canary identifier not copied.
  - REQ-OBS-0052: A server selecting an unoffered protocol (confirmed to fail `crypto/tls` ALPN validation) yields an `inconclusive` handshake with `inconclusive` certificate validation, no `alpn`, `tls_version` or `alert` member, and no copy of the server's identifier canary.

<a id="UT.root.TestRuntimeVocabularyReuse"></a>

### UT.root.TestRuntimeVocabularyReuse

- Case: `UT.root.TestRuntimeVocabularyReuse`
- Result: passed
- Test: `TestRuntimeVocabularyReuse` in `runtime_vocabularies_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestRuntimeVocabularyReuse$' .`
- Scope: Existing synthetic Go unit test TestRuntimeVocabularyReuse in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0046: In a NET-05 composite, a DNS-01 step's code-16 response is retained as OTHER.
  - REQ-OBS-0047: That OTHER response gives the DNS-01 step the inconclusive outcome while the composite still completes.
  - REQ-OBS-0050: HTTP-01 and NET-05 TLS steps receiving peer alert 200 retain details.alert other unchanged.

<a id="UT.root.TestMTcp01Admission"></a>

### UT.root.TestMTcp01Admission

- Case: `UT.root.TestMTcp01Admission`
- Result: passed
- Test: `TestMTcp01Admission` in `tcp01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMTcp01Admission$' .`
- Scope: Existing synthetic Go unit test TestMTcp01Admission in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0001: TCP-01 input null and [] are rejected as task_schema_invalid by the object schema, atomically and with no contact.
  - REQ-JSON-0002: Valid \u escapes in member names (\u0070ort, \u0078) are decoded before matching; invalid UTF-8, malformed escapes and surrogates are not exercised here.
  - REQ-JSON-0003: A repeated port, including an escaped spelling with an equal value, is rejected.
  - REQ-JSON-0004: An unknown member (\u0078) is rejected rather than discarded.
  - REQ-JSON-0005: A missing required port, port as a string "443" and port null are rejected rather than defaulted or coerced.

<a id="UT.root.TestMTcp01ConnectionBudget"></a>

### UT.root.TestMTcp01ConnectionBudget

- Case: `UT.root.TestMTcp01ConnectionBudget`
- Result: passed
- Test: `TestMTcp01ConnectionBudget` in `tcp01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMTcp01ConnectionBudget$' .`
- Scope: Existing synthetic Go unit test TestMTcp01ConnectionBudget in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0006: With max_connections 1 and 2, TCP-01 through the Client completes as connected with exactly one dial and one accepted connection.

<a id="UT.root.TestMTcp01Options"></a>

### UT.root.TestMTcp01Options

- Case: `UT.root.TestMTcp01Options`
- Result: passed
- Test: `TestMTcp01Options` in `tcp01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMTcp01Options$' .`
- Scope: Existing synthetic Go unit test TestMTcp01Options in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0004: TCP-01 input of 49,152 and options of 8,192 encoded bytes are admitted and one byte above is rejected as task_limit_exceeded.

<a id="UT.root.TestMTcp01Termination"></a>

### UT.root.TestMTcp01Termination

- Case: `UT.root.TestMTcp01Termination`
- Result: passed
- Test: `TestMTcp01Termination` in `tcp01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMTcp01Termination$' .`
- Scope: Existing synthetic Go unit test TestMTcp01Termination in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0051: network_changed or host cancellation during the dial, a family-check failure and a destination-gate refusal end the attempt incomplete with that reason and no measurement.
  - REQ-RESULT-0052: A network change while the connection closes after the connected outcome is recorded leaves the attempt completed with that outcome.

<a id="UT.root.TestTcp01AdmissionRiskAndTimezoneDomains"></a>

### UT.root.TestTcp01AdmissionRiskAndTimezoneDomains

- Case: `UT.root.TestTcp01AdmissionRiskAndTimezoneDomains`
- Result: passed
- Test: `TestTcp01AdmissionRiskAndTimezoneDomains` in `tcp01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestTcp01AdmissionRiskAndTimezoneDomains$' .`
- Scope: Existing synthetic Go unit test TestTcp01AdmissionRiskAndTimezoneDomains in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-SAFETY-0031: risk_level medium, high or unknown with risk_factors omitted or [] is rejected as task_schema_invalid with no replay commit; with one factor, and low in every form, it is admitted.
  - REQ-SCHEDULE-0043: timezone omitted or local is admitted and UTC, Europe/Berlin, empty, Local, LOCAL and 'local ' are rejected as task_schema_invalid.

<a id="UT.root.TestTcp01StrictSchemaBothOrigins"></a>

### UT.root.TestTcp01StrictSchemaBothOrigins

- Case: `UT.root.TestTcp01StrictSchemaBothOrigins`
- Result: passed
- Test: `TestTcp01StrictSchemaBothOrigins` in `tcp01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestTcp01StrictSchemaBothOrigins$' .`
- Scope: Existing synthetic Go unit test TestTcp01StrictSchemaBothOrigins in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0003: An escaped repeat of port is rejected for embedded and signed remote delivery.
  - REQ-JSON-0004: Escaped spellings of ip_address and port match their fixed fields and an unknown \u0078 member is rejected, for both origins.
  - REQ-JSON-0006: port 1e2 and -0 are rejected as task_schema_invalid for both origins.

<a id="UT.root.TestMTls01Limits"></a>

### UT.root.TestMTls01Limits

- Case: `UT.root.TestMTls01Limits`
- Result: passed
- Test: `TestMTls01Limits` in `tls01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMTls01Limits$' .`
- Scope: Existing synthetic Go unit test TestMTls01Limits in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0006: With max_connections 1 and 2, TLS-01 completes with exactly one dial and one accepted connection.
  - REQ-LIMIT-0008: A stalled handshake reaches the operation bound as tls_timeout (completed) while max_duration_ms 1,000 ends the attempt incomplete with attempt_deadline_exceeded, and the stalled session is released without retry.
  - REQ-LIMIT-0034: max_bytes_sent and max_bytes_received of 1 suffice for a successful handshake, so handshake and record bytes are not counted.

<a id="UT.root.TestMTls01Options"></a>

### UT.root.TestMTls01Options

- Case: `UT.root.TestMTls01Options`
- Result: passed
- Test: `TestMTls01Options` in `tls01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMTls01Options$' .`
- Scope: Existing synthetic Go unit test TestMTls01Options in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-OBS-0049: alpn_protocols with h2, dot, http/1.1 or all three is admitted, and unknown, h2c, H2, h3, spdy/3, empty, 33-byte, duplicate and nine-entry lists are rejected as task_schema_invalid with no contact.

<a id="UT.root.TestMTls01Termination"></a>

### UT.root.TestMTls01Termination

- Case: `UT.root.TestMTls01Termination`
- Result: passed
- Test: `TestMTls01Termination` in `tls01_conformance_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestMTls01Termination$' .`
- Scope: Existing synthetic Go unit test TestMTls01Termination in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-RESULT-0051: network_changed or host cancellation during the handshake, and a refused CONNECT tunnel, end the attempt incomplete with that reason and no measurement, with no direct fallback.

<a id="UT.root.TestNet05TLSPrimitiveUnclassifiedConnectFailure"></a>

### UT.root.TestNet05TLSPrimitiveUnclassifiedConnectFailure

- Case: `UT.root.TestNet05TLSPrimitiveUnclassifiedConnectFailure`
- Result: passed
- Test: `TestNet05TLSPrimitiveUnclassifiedConnectFailure` in `net05_runner_test.go` (root package, Go package `barometer`)
- Reproduce one case: `go test -count=1 -run '^TestNet05TLSPrimitiveUnclassifiedConnectFailure$' .`
- Scope: Existing synthetic Go unit test TestNet05TLSPrimitiveUnclassifiedConnectFailure in the root package, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-JSON-0010: A TCP failure without a TLS-01 class is published as failure inconclusive after strict re-decoding, without the unsupported failure value.

<a id="UT.internal.measurement.net05.TestDeclaredMaximaFailClosed"></a>

### UT.internal.measurement.net05.TestDeclaredMaximaFailClosed

- Case: `UT.internal.measurement.net05.TestDeclaredMaximaFailClosed`
- Result: passed
- Test: `TestDeclaredMaximaFailClosed` in `internal/measurement/net05/codec_test.go` (package `internal/measurement/net05`)
- Reproduce one case: `go test -count=1 -run '^TestDeclaredMaximaFailClosed$' ./internal/measurement/net05`
- Scope: Existing synthetic Go unit test TestDeclaredMaximaFailClosed in package internal/measurement/net05, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-LIMIT-0035: A profile with a step whose parameters report no declared maxima is not admitted by `Parameters.ValidateLimits`, even at the composite limits, so an undeclared primitive cannot pass the sum check.
<a id="UT.root.TestLocalPolicyContextEligibility"></a>

### UT.root.TestLocalPolicyContextEligibility

- Case: `UT.root.TestLocalPolicyContextEligibility`
- Result: passed
- Test: `TestLocalPolicyContextEligibility` in `local_restrictions_test.go` (package `root`)
- Reproduce one case: `go test -count=1 -run '^TestLocalPolicyContextEligibility$' ./.`
- Scope: Existing synthetic Go unit test TestLocalPolicyContextEligibility in package root, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-EXECUTION-0022: With local eligibility limited to the direct path, a proxy-path task records egress `unavailable` with `network_path` `proxy`, and `network_context_unavailable` when it requires `country_code`, with no `CONNECT` to the context endpoint, while the direct path still acquires a fresh context; across all rows only endpoints the host configured and local policy permits are contacted. The tasks supply no endpoint values in this test.
  - REQ-EXECUTION-0023: For the stricter-local-policy rule, a local policy that forbids acquisition, names no endpoint, permits only the proxy path, names only an unconfigured endpoint or permits only the failing fallback turns tasks with a `required_fields` or `allowed_country_codes` requirement into `network_context_unavailable`, while policies that permit the answering primary on the direct path let them complete.
  - REQ-EXECUTION-0024: In each of those restricted rows a task without context requirements still completes with `egress.status` `unavailable`.
