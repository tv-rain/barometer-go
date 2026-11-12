# Measurement Interpretation Limits

## Authority and Use

This document is the field and outcome guidance that ships with the Go implementation of the nine Draft `0.1.0` measurements. The public test documents under [`tests/`](https://github.com/tv-rain/barometer-spec/tree/main/tests) own the Supported Interpretation, Alternative Explanations and Evidence Needed for Stronger Interpretation of each test, and the cross-cutting [Measurement Limitations](https://github.com/tv-rain/barometer-spec/blob/main/measurement-limitations.md) apply to every result. This text restates those limits for readers of Go results; where it differs from the public text, the public text controls.

Every result is an observation of one execution. No result field, outcome or document in this repository is a verdict, and no result is proof of censorship. Results carry no interpretation, verdict, confidence, anomaly or blocking field, and the strict result decoder rejects such a field as unknown. The Go result types and their documentation describe what was observed; they never classify why.

The portable `M-*-INTERPRETATION` conformance cases check this document, the Go source comments and strings, the serialized results and the sanitized examples for the prohibited claims of each test. The judgement-based `RV-M-*-INTERPRETATION` review cases remain separate and are not automated coverage.

## TCP-01

One TCP-01 execution establishes only whether a TCP connection to the configured address and port completed or produced a particular observable socket outcome. The test performs no DNS resolution and sends no application data, so a result says nothing about DNS, TLS, HTTP or any other application protocol. A refusal, reset or timeout does not establish intentional filtering, and no outcome is a censorship, blocking, anomaly or success verdict.

Alternative explanations: the target service may be offline, overloaded or not listening; a firewall may reject or drop the connection for operational reasons; routing, peering, NAT, carrier-grade NAT, VPN or local network failure; IPv4/IPv6 path differences; transient packet loss or mobile-network handover; incorrect or stale target configuration.

Stronger interpretation needs repeated attempts under a stable network generation and a conservative schedule, successful connections to suitable control endpoints on the same port, DNS, TLS and HTTPS observations for related approved targets, and independent measurements from comparable times and network contexts. A single connection result is not definitive proof of censorship.

## TLS-01

One TLS-01 execution establishes whether TCP and normal authenticated TLS completed for the configured address, port, server name and ALPN offer. A handshake failure does not establish TLS or SNI filtering, censorship, blocking, interference or intent. A certificate outcome (`certificate_untrusted`, `certificate_name_mismatch` or `certificate_time_invalid`) does not establish interception or tampering; target, CDN, trust-store or certificate misconfiguration and incorrect device time remain explanations. `tcp_timeout`, `tcp_reset` and `tls_timeout` do not exclude routing, packet loss, mobile handover, VPN, firewall, proxy, rate limiting or overload. `alpn_no_overlap` and `protocol_error` do not exclude an unsupported TLS version, algorithm or ALPN configuration. `handshake_succeeded` does not establish HTTP or application-content availability. The test does not compare different SNI values, and a failure does not exclude a stale address-to-name mapping. Outcome `inconclusive` also covers a TCP connect failure the outcome table does not name (network or address unreachable, permission denied or another socket error), a known limitation of this version.

Alternative explanations: target, load balancer, CDN, certificate or trust-store misconfiguration; expired certificate or incorrect device time; routing, packet loss, mobile handover, VPN, firewall or proxy behavior; unsupported TLS version, algorithm or ALPN configuration; stale address-to-name mapping; server-side rate limiting or overload.

Stronger interpretation needs a TCP-01 result for the same address and port, repeated TLS observations under a stable network generation, successful control handshakes using comparable infrastructure, and HTTP-01 results and independent observations from comparable contexts.

## DNS-01

One DNS-01 execution establishes only the result exposed by the selected resolver path for the configured name, record type, time and network generation. In direct UDP mode it can also establish that several matching response datagrams arrived and whether their normalized contents differed; it cannot establish which response was injected or correct, and it does not label any response as spoofed, injected, legitimate or authoritative. Answers are not compared with authoritative data in the run, a returned address is not shown to be reachable, and an unexpected answer is not shown to have been introduced intentionally. A difference between resolvers is evidence of differing observations, not proof that either resolver is correct or that censorship occurred. NXDOMAIN or an empty answer is not a successful address resolution. Answers carry no ASN, organization, geolocation, reputation or verdict enrichment.

Alternative explanations: authoritative DNS or recursive resolver outage; DNSSEC validation failure, stale cache, negative caching, split-horizon DNS or DNS64 behavior; misconfigured, expired or recently changed records; captive portal, VPN, enterprise resolver or local filtering; platform-specific error collapsing; direct resolver outage, filtering, rate limiting or routing failure; a legitimate resolver retransmission or anycast change.

Stronger interpretation needs repeated system and direct results on a stable network, comparison with authoritative information under a separately reviewed method, DNS-03 and DNS-04 observations for the same safe query where permitted, TCP, TLS or HTTPS measurements to approved expected endpoints, and independent observations from comparable times and network contexts.

## DNS-03

One DNS-03 execution establishes whether each observed bootstrap, TCP, TLS, HTTP and DNS stage completed for the configured resolver and query. The test uses its own DNS-over-HTTPS client; it does not configure DoH for the device or the host application. A DoH answer is not authoritative or correct, and a difference from system DNS does not establish interference. A failure is attributed only to the directly observed stage, shown by the observation order and the terminal observation; a bootstrap, TCP, TLS, HTTP or DNS failure is not summarized as DoH blocking or censorship. Answers carry no on-device ASN, organization, geolocation, reputation or verdict enrichment.

Alternative explanations: resolver, authoritative DNS, CDN or certificate outage; bootstrap DNS failure unrelated to the DoH service; resolver rate limiting, policy refusal, maintenance or overload; VPN, proxy, captive portal, enterprise policy or local firewall behavior; IPv4/IPv6 routing differences or transient packet loss; stale or incorrect registered resolver configuration.

Stronger interpretation needs repeated DoH results on a stable network without aggressive retries, DNS-01 and DNS-04 results for the same safe query, layer-specific TCP, TLS and HTTPS checks against approved endpoints, and independent observations from comparable times and network contexts.

## DNS-04

One DNS-04 execution establishes whether each observed bootstrap, TCP, TLS and DNS stage completed for the configured resolver and query, and nothing more. The test uses its own DNS-over-TLS client; it does not configure DoT for the device or the host application. A response answer is not authoritative or correct. A failure on the standard DoT port 853 does not establish intentional blocking, and a difference between DNS-04 and DNS-01 or DNS-03 supports a transport-specific observation, not a conclusion about censor intent. `dns_refused`, `tls_failure`, `tcp_reset` or `tcp_timeout` alone does not exclude resolver, authoritative DNS, certificate or routing outage, bootstrap failure, resolver rate limiting or policy refusal, port-specific firewall behavior, VPN, captive portal, enterprise policy or local security software, or IPv4/IPv6 path differences.

Alternative explanations: resolver, authoritative DNS, certificate or routing outage; bootstrap DNS failure unrelated to the DoT service; resolver rate limiting, policy refusal, overload or maintenance; port-specific firewall behavior unrelated to censorship; VPN, captive portal, enterprise policy or local security software; IPv4/IPv6 path differences or stale resolver configuration.

Stronger interpretation needs repeated DoT results on a stable network without aggressive retries, DNS-01 and DNS-03 results for the same safe query, a TCP-01 and TLS-01 observation for the approved resolver endpoint, and independent observations from comparable times and network contexts.

## HTTP-01

One HTTP-01 execution establishes whether the configured endpoint or an approved terminal redirect target responded, and whether its bounded status, header and body checks matched the registered expectations, through the observed DNS, TCP, TLS and HTTP path. The test does not browse or render a page, does not authenticate, submit user data or execute scripts. An unexpected status, `redirect_observed`, `redirect_rejected`, `header_check_failed` or `body_check_failed` does not establish a block page, intentional manipulation or blocking; response differences and redirects need manual or separately reviewed classification. `body.bytes_processed` and `max_body_bytes` count bytes processed by the application and are not an exact on-wire limit, because the HTTP, TLS and socket layers may buffer more data. A result does not exclude origin, CDN, DNS, routing, certificate or deployment failure, intentional redirect or changed content, captive portal, enterprise proxy, VPN, antivirus or local filtering, server-side geolocation, rate limiting, load balancing or maintenance, IPv4/IPv6 or transient mobile-network differences, a stale expected status, body pattern, digest or header expectation, or legitimate redirect-chain changes.

Stronger interpretation needs DNS-01, TCP-01 and TLS-01 observations for related approved targets, a NET-03 result excluding obvious validation or captive-portal problems, repeated observations under a stable network generation, and suitable control endpoints and independent reference signals.

## NET-01

One NET-01 execution establishes only the normalized access-network properties the operating system exposed at that time and whether the network changed during the snapshot. The availability values (`ipv4_available`, `ipv6_available` and `validated`) are platform-reported and are not proven reachability: the test does not establish that the Internet is reachable, and dual-stack capability does not establish successful IPv4 or IPv6 Internet reachability. `vpn_state: detected` reports a VPN interface and does not establish that traffic is routed through a VPN. The test does not identify a network operator, subscriber, device or physical location, and country, region, ASN, provider, mobile operator and public IP address are not derived from NET-01; they require separate egress or consented context. A single result is not definitive proof of censorship.

Alternative explanations: platform APIs may expose incomplete or delayed state; a VPN may be present without carrying the measured traffic; operating-system validation state may be stale.

Stronger interpretation needs NET-03 results for Internet validation or captive-portal behavior, protocol measurements within the same stable network generation, and repeated snapshots after a reported network change.

## NET-03

One NET-03 execution establishes whether the configured validation endpoint returned its expected response, appeared to be intercepted, or could not be validated on the observed network path. `validated` does not establish that arbitrary or general Internet destinations are reachable. `captive_portal_suspected`, a redirect or a content mismatch does not identify the operator or the purpose of an intercepting page. `not_validated` and other unvalidated results are not evidence of censorship, and no result is definitive proof of censorship or a generic blocking, anomaly or success verdict. `platform_validation` is retained independently of the active request and does not override or confirm it; the outcome is selected from the active request first, and the platform state selects it only when the request produced no response class.

Alternative explanations: the validation endpoint, DNS provider, CDN or certificate may be misconfigured or unavailable; a subscription, quota or local access-control page may intercept traffic; the platform validation state may be stale; a proxy, antivirus product, enterprise gateway or VPN may alter the response; a transient network change may split the observations.

Stronger interpretation needs a repeat on the same stable network generation, successful and failed measurements against independently operated benign endpoints, DNS, TCP, TLS and HTTP layer results from the same time window, and independent reference signals from a comparable network context.

## NET-05

One complete NET-05 execution establishes only the ordered primitive results the inline profile produced on one network generation. The test does not prove that an allowlist policy exists or identify who operates it, does not classify the combined results as a broad restriction, allowlist or censorship event, and is not for sensitive, political, circumvention or broad public website lists. It does not replace outage, captive-portal, routing or independent reference analysis. Correlating the results with knowledge of the targets, repeated runs, controls and independent signals is outside NET-05 version `0.1.0`; even correlated observations cannot establish censor intent or policy with certainty, and external datasets are supporting reference signals, not ground truth.

Alternative explanations: simultaneous outage or misconfiguration affecting external profile targets; shared hidden CDN, DNS, routing or certificate dependencies among supposedly independent targets; international transit failure, routing isolation, congestion or packet loss; captive portal, subscription state, VPN, enterprise policy or local firewall; stale inline profile definitions or changed endpoint behavior; address-family-specific failure or mobile-network handover.

Stronger interpretation needs repeated complete profiles separated by a conservative schedule on stable network generations, consistent failure across independently operated targets and multiple protocol layers, continued reachability of suitable local or host-application controls, no captive-portal or obvious local-network explanation, and independent reference signals from comparable times and network or geographic contexts.
