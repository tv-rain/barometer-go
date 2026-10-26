# Go Unit-Test Evidence at 178bda0

DEV-664 records the milestone evidence for the Measurement reachability primitives sub-epic (DEV-656): executed synthetic Go tests for the production TCP-01 and DNS-01 codecs, runners and shared primitives (DEV-659 to DEV-663), plus the connection, collection, serialization, termination and version rows added under DEV-664. Each case section below is the conformance evidence for one measurement manifest case: 16 conformance family cases and 19 `UT.*` unit-test cases. The cross-cutting manifest keeps its earlier baseline and its [evidence at d788a80](go-unit-d788a80.md).

## Execution Record

- Implementation: `https://github.com/tv-rain/barometer-go` commit `178bda0c73bab64efaf0158f757edc316eeddd1a` on the DEV-664 task branch, exported with `git archive` into an empty directory outside any checkout.
- Specification baseline: `https://github.com/tv-rain/barometer-spec` commit `5d01879cfd1d47396c780befa2567f7e90439948` (Draft `0.1.0`, public PR #24 merge), the cross-cutting manifest baseline. Between the previous measurement baseline `59d49364d85e48cc8e69c762d130a7a86dc1fe65` and this revision the measurement documents add only REQ-NET-05-0061 and one unnumbered TLS-01 cross-reference; every other measurement requirement is unchanged. Later public revisions add REQ-HTTP-01-0051 with the explicit-IP-only proxy contract, which this baseline does not adopt.
- Environment: go1.26.9 darwin/arm64 (`GOTOOLCHAIN=local`, `GOFLAGS` empty, `CGO_ENABLED=1`), default test timeout, no race detector.
- Command, from the export root: `go test -count=1 -json ./...`
- Outcome: exit status 0; all 494 top-level tests passed, none failed or skipped. Each case below names its passing top-level tests; a top-level pass includes all of its subtests. Fuzz targets ran their seed corpora only.
- Inputs are the synthetic, in-repository fixtures at the implementation commit: public IP literals and synthetic names that the test transports map to loopback servers, the test-only scripted HTTP/1.1 CONNECT proxy and loopback DNS server of `internal/testkit/loopback`, a fake platform network, an injected system resolver and deterministic clocks. No public destination is contacted, and no deployment key, endpoint, proxy, participant, measurement or private data is involved.

## Carried-Forward Cases

The five `UT.*` cases of the [c5f4e92 report](go-unit-c5f4e92.md) were re-executed at this baseline and keep their recorded aspects; their sections are repeated below.

## New Cases

The eight automated families of TCP-01 and of DNS-01 named by the sub-epic's definition of done are executed: ADMISSION, OPTIONS, EXECUTION, NORMALIZATION, LIMITS, TERMINATION, OUTPUT and VERSIONS. A family passes only when every row of its case plan is exercised by the tests named in its section; three plan rows carry recorded dispositions (the DNS-01 option window maximum, the TCP-01 receiver quarantine row and the conservative DNS-01 serialization key), and the TERMINATION plans carry the GO-SPEC-ISSUE-004 states. PRECONDITIONS, MINIMIZATION, INTERPRETATION and SAFETY stay planned; the unit-test cases below link the existing tests that touch their requirements.

## Limits

Every case is partial evidence. A passing case shows only the stated aspects of each linked requirement; the manifest notes record the remaining obligations. Destinations are public literals mapped to loopback inside test transports, the system resolver is injected and suspension is an injected engine cause, so no native resolver, routing, suspension or proxy behavior is shown. Overlap checks for equivalent tuples wait a fixed 20 ms, so on a slow host they can only miss an overlap, never report a false one. No `RV-*` review case, platform or device validation or field measurement is involved. DNS and TCP observations are not evidence of censorship. These reports establish no implemented requirement, released API, platform or field support.

## Conformance Family Cases

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

## Unit-Test Cases

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

<a id="UT.internal.execution.TestTcp01ConnectionMeter"></a>

### UT.internal.execution.TestTcp01ConnectionMeter

- Case: `UT.internal.execution.TestTcp01ConnectionMeter`
- Result: passed
- Test: `TestTcp01ConnectionMeter` in `internal/execution/tcp01_termination_test.go` (package `internal/execution`)
- Reproduce one case: `go test -count=1 -run '^TestTcp01ConnectionMeter$' ./internal/execution`
- Scope: Existing synthetic Go unit test TestTcp01ConnectionMeter in package internal/execution, executed with go test at the implementation baseline. Partial evidence only for the narrow aspects of each linked requirement recorded in the evidence report; not a planned conformance family case and not runtime, platform or field conformance.
- Requirement aspects exercised:

  - REQ-TCP-01-0022: max_connections 0 refuses the dial before the transport with resource_limit and zero usage; at 1 and 2 the usage is exactly one connection and no bytes.

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
