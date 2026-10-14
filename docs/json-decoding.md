# Shared Strict JSON Decoding

## Authority and Scope

`internal/jsondecode` adopts public Draft `0.1.0` [Runtime JSON Decoding](https://github.com/tv-rain/barometer-spec/blob/05c1364af4f9a604b617fca45f060949b557d265/data-formats/json-decoding.md) at immutable `tv-rain/barometer-spec` revision `05c1364af4f9a604b617fca45f060949b557d265` (PR #5). That public contract remains authoritative. DEV-477 provides reusable internal infrastructure and synthetic tests, without implementing or declaring support for each public serialized model, admission, measurements, platform lifecycle, or host/binding behavior.

The package is a standard-library-only dependency leaf. It owns bounded JSON parsing and declarative representation validation; it performs no I/O, logging, defaulting, normalization, scheduling, persistence, trust decision, or execution. Production does not import `testkit`. No public Go API or third-party dependency is introduced.

## Consumer Contract

`Decode(input, schema, limits)` requires an explicitly chosen schema from trusted containing/integration context. `DecodeSelected(input, selection, limits)` selects exactly one registered ordered tuple of required string fields, supporting either a schema-version field or a test-id/test-version pair. It parses once, validates the selected schema once, and never attempts another version on failure. A malformed or unsupported tuple rejects the complete model. Versionless formats use trusted schema selection without adding serialized version fields. Callers are responsible for registering exact supported versions and validating their syntax, never a latest/compatible fallback.

Schemas declare exact JSON types, required fields, explicit nullability, closed fixed-field objects, array element schemas, and explicit dynamic-key value schemas. An `Additional` schema permits dynamic names only where the public model defines them; model owners must still validate label namespaces, key syntax and other domain rules. Integer fields require finite canonical decimal minimum/maximum endpoints. Missing, null, empty and zero remain distinct; optional fields are not defaulted. Integer text remains exact, with no float conversion, wrapping or clamping. There is no non-integer numeric schema because this public contract introduces none.

`Value` provides typed-kind inspection, exact scalar text, booleans, member lookup with presence, copied name/element containers, and original encoded-value size. Values are candidates for downstream domain validation, never admitted models. No value is returned with an error, including nested validation failure. Callers must check kind/presence before extracting fields and construct domain models only after all domain validation succeeds. No trusted capability or source context may be populated from these values.

Schemas and selection descriptors are trusted configuration and must remain immutable during calls. Invalid descriptors, cycles, malformed/unbounded integer ranges and conflicting selections fail with a fixed configuration error. Descriptor checking permits at most 4096 schema visits and depth 128 per schema, at most 16 selection fields, and at most 256 cases. Integer endpoints permit at most 4096 decimal digits, allowing finite domains beyond signed/unsigned machine widths. A field's complete domain must fit these endpoints; a model with an unbounded or larger domain needs reviewed representation work before support is claimed.

## Resource and Failure Policy

Every call requires positive finite complete-input byte, value-depth and value-node bounds supplied by the consuming model owner. The stack ceiling is 128 value levels, including the root; nodes count values rather than member names. Reviewed model-specific limits may be lower. No default input maximum enables a model implicitly. Models without registered complete-input limits require reviewed local limits. Tests use local synthetic limits, not normative defaults or deployment policy.

The complete byte envelope is checked before UTF-8 scanning or parsing, including whitespace and unknown fields; unknown versions cannot enlarge it. Parsing checks depth and node budget before allocating the next value. Strings, names, numeric tokens and containers are bounded by input bytes and nodes. Schema `MaxBytes` checks each original encoded value span (braces/brackets and internal whitespace included); root surrounding whitespace is counted by `Limits.MaxBytes`. Schema `MaxCount` bounds object members or array elements. Optional schema bounds of zero defer to the mandatory global envelope, so consumers must supply every registered nested limit. Nested/count/schema rejection can occur after the bounded parse, but never publishes a partially validated value.

Work and memory remain finite under the selected envelope: the parser does not scan beyond it, depth is capped, exact-number arithmetic is bounded by token/endpoint lengths, and schema validation visits bounded values/descriptors. CPU/memory thresholds on real mobile devices are not measured by these synthetic tests. Callers must review their envelopes against complete registered model domains and mobile constraints.

Failures use fixed sentinel classes: syntax, text, duplicate, field, type, required, integer, range, version, bounds, or configuration. They contain no offsets, received values/names, targets, raw parser errors, or secrets. Input traversal order and sorted required-field checks give stable precedence for the same schema/policy baseline; cross-language precedence is not specified. Rejection is an operational error and supplies no Measurement Result or censorship evidence.

The package retains a bounded parsed representation, not a mutable slice of the input. Callers separately retain exact received task bytes where the trust/replay contract requires them; reserialization is not equivalent integrity evidence.

## Synthetic Evidence Map

All cases use synthetic data, the baseline above, and `internal/jsondecode/decode_test.go`. No transport, deployment service, credentials, or participant data is used.

| Public obligation | Implementation | Cases |
|---|---|---|
| REQ-JSON-0001: complete framing | `parser.go`: `parse`, `value` | `TestRejectionCorpus`, `TestPositiveAndPresence`, `FuzzParse` |
| REQ-JSON-0002: UTF-8/scalar strings | `parse`, `string`, `hex4` | Negative UTF-8/escape/surrogate corpus and positive pair/literal-replacement controls |
| REQ-JSON-0003: unique decoded names | `value` object parsing | Equal/conflicting/escaped, nested-label and array-object duplicates; distinct-object and non-normalized Unicode controls |
| REQ-JSON-0004: exact/unknown fields | `decode.go`: `validate` | Case variants, unknown top-level/nested names, escaped known field and dynamic labels |
| REQ-JSON-0005: types/presence | `validate`, `Value.Lookup` | Wrong type, required/null failure, missing/null/empty/zero controls |
| REQ-JSON-0006: integer spelling | `integerToken`, numeric grammar | Negative zero, decimal/exponent, quoted/boolean values, sign/leading-zero rejection |
| REQ-JSON-0007: exact numbers | `validate`, finite schema ranges | Exact 9007199254740993, signed/unsigned 64-bit endpoints and overflow, 100-digit finite-domain control |
| REQ-JSON-0008: exact selection | `DecodeSelected` | Missing/wrong/unsupported tuples, duplicates/configuration, selected-schema failure without fallback |
| REQ-JSON-0009: original bounds | `parse`, `value`, `validate` | `TestBounds`: below/at/above bytes, nested bytes, counts, nodes and depth; oversized input before parsing/selection |
| REQ-JSON-0010: bounded atomic decoding | `Decode`, `DecodeSelected` | Corpus nil-candidate assertions, malformed/deep input and deterministic fuzz checks |
| REQ-JSON-0011: sanitized errors | Fixed sentinels only | `TestErrorsAndDeterminism`, fuzz error-size checks |

Run `make check` and the public requirement-reference checker against the explicit sibling checkout. Run each fuzz target separately, for example `go test ./internal/jsondecode -run='^$' -fuzz='^FuzzDecode$' -fuzztime=20s -parallel=2`, then `FuzzParse` with the same arguments. Passing synthetic cases validates this infrastructure against its chosen descriptors; individual-model conformance, binding agreement, device resource behavior, specialist review, and field support remain pending.

## Local Validation — October 14, 2026

The prepared DEV-477 changes passed `make check`: formatting, host build, unit tests, race detection, `go vet`, module consistency/verification, working-diff whitespace, and pure-Go Android ARM64/iOS ARM64 compilation. Requirement-reference structure passed against the explicit public checkout with both Go and internal-review consumers; catalog allocation history was not checked because this work changes no allocations. Both repositories passed `git diff --check`.

Two 20-second fuzz runs with two workers passed: `FuzzDecode` executed 1,206,684 cases, checking deterministic complete-model/schema selection and bounded atomic failures; `FuzzParse` executed 218,185 cases, independently checking accepted framing with the standard JSON validator. These finite fuzz runs are regression evidence, not exhaustive proof. No native artifact, device test, CI run, model-conformance claim, release, or downstream issue closure is established.

## Adoption and Merge Order

Public PR #5 is merged. DEV-477 first aligns the internal trust and resource validation plans; those records must merge before this Go adoption is released as dependent behavior. Each repository uses a separate branch/commit and human-created PR. The Go [finding addendum](specification-issues.md#adoption-addendum--october-14-2026) preserves the historical finding and downstream evidence gaps. No release, support promotion, public schema change, or broad source-inventory rewrite accompanies shared infrastructure.
