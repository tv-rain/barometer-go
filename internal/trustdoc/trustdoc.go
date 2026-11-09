// Package trustdoc decodes the trusted local document that reviewed,
// Barometer-owned native code installs through the gomobile facade: the local
// policy with its issuer keys, target registrations and generation, the
// configuration approvals and the egress-classification approvals. The policy
// also carries the optional local restriction inputs: the release-reviewed
// embedded one-time tasks, local device constraints and local Network Context
// eligibility.
//
// The document is a local deployment input, packaged with the application and
// read by the native glue. It is never a remotely delivered format: no task,
// signed package, collector reply, endpoint response, host configuration or
// Dart call carries it. Decoding uses the shared strict JSON decoder with a
// closed schema, so duplicate, unknown or missing members reject the whole
// document; the model constructors then apply every structural invariant.
// Errors carry no detail.
package trustdoc

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"math"
	"net/netip"
	"strconv"
	"time"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/model/task"
)

// Format is the only accepted value of the document's format member.
const Format = "barometer-trust/1"

// MaxBytes bounds the encoded document. It is an implementation-local bound,
// not a registered limit.
const MaxBytes = 1 << 20

// ErrInvalid is returned for any document that does not decode to a complete
// valid policy and approvals.
var ErrInvalid = errors.New("trustdoc: invalid trusted local document")

// Document is the decoded trusted input. Policy is valid.
type Document struct {
	Policy    task.LocalPolicy
	Approvals config.Approvals
	Context   observation.ContextApprovals
}

func (Document) String() string   { return "trustdoc.Document" }
func (Document) GoString() string { return "trustdoc.Document" }

var limits = jsondecode.Limits{MaxBytes: MaxBytes, MaxDepth: 10, MaxNodes: 1 << 17}

func object(fields map[string]jsondecode.Field) *jsondecode.Schema {
	return &jsondecode.Schema{Kind: jsondecode.Object, Fields: fields}
}
func req(s *jsondecode.Schema) jsondecode.Field { return jsondecode.Field{Schema: s, Required: true} }
func opt(s *jsondecode.Schema) jsondecode.Field { return jsondecode.Field{Schema: s} }
func text() *jsondecode.Schema                  { return &jsondecode.Schema{Kind: jsondecode.String} }
func flag() *jsondecode.Schema                  { return &jsondecode.Schema{Kind: jsondecode.Boolean} }
func integer(min, max int64) *jsondecode.Schema {
	return &jsondecode.Schema{Kind: jsondecode.Integer, Min: strconv.FormatInt(min, 10), Max: strconv.FormatInt(max, 10)}
}
func list(element *jsondecode.Schema, max int) *jsondecode.Schema {
	return &jsondecode.Schema{Kind: jsondecode.Array, Element: element, MaxCount: max}
}

// Collection bounds are at or above the model bounds, which reject larger
// valid-looking input with the same ErrInvalid.
const (
	maxEntries   = 256
	maxApprovals = 64
	maxInt       = math.MaxInt32
)

var testSchema = object(map[string]jsondecode.Field{"id": req(text()), "version": req(text())})
var pathsSchema = object(map[string]jsondecode.Field{"direct": req(flag()), "proxy": req(flag())})

var limitsSchema = func() *jsondecode.Schema {
	f := map[string]jsondecode.Field{"attempt": req(object(map[string]jsondecode.Field{
		"duration_ms": req(integer(0, maxInt)), "bytes_sent": req(integer(0, maxInt)),
		"bytes_received": req(integer(0, maxInt)), "connections": req(integer(0, maxInt)),
	}))}
	for _, name := range limitNames {
		f[name] = req(integer(0, maxInt))
	}
	return object(f)
}()

var limitNames = []string{
	"max_runs", "max_attempts", "max_validity_seconds", "min_repeat_interval_seconds", "min_repeat_jitter_seconds",
	"min_retry_delay_ms", "max_retry_delay_ms", "max_retry_jitter_ms", "max_result_bytes", "max_bundle_bytes",
	"max_results_per_bundle", "max_pending_bundles", "max_pending_bytes", "max_retention_seconds",
}

var permissionsSchema = object(map[string]jsondecode.Field{
	"tests": req(list(testSchema, maxEntries)), "registration_ids": req(list(text(), maxEntries)),
	"paths": req(pathsSchema), "max_risk": req(text()),
	"require_consent": req(flag()), "allow_medium_risk": req(flag()), "allow_composite": req(flag()), "allow_unknown_risk": req(flag()),
	"label_namespaces": req(list(text(), maxEntries)), "limits": req(limitsSchema),
})

var tupleSchema = object(map[string]jsondecode.Field{
	"hostname": opt(text()), "address": opt(text()), "port": opt(integer(1, 65535)), "scheme": opt(text()), "url": opt(text()),
	"resolver_name": opt(text()), "tls_server_name": opt(text()),
	"bootstrap_addresses": opt(list(text(), 16)), "alpn": opt(list(text(), 8)), "expected_status": opt(integer(100, 599)),
	"body_checks":   opt(list(object(map[string]jsondecode.Field{"type": req(text()), "value": req(text())}), 8)),
	"header_checks": opt(list(object(map[string]jsondecode.Field{"name": req(text()), "type": req(text()), "value": req(text())}), 8)),
})

var targetSchema = object(map[string]jsondecode.Field{
	"id": req(text()), "test": req(testSchema), "role": req(text()), "tuple": req(tupleSchema), "paths": req(pathsSchema),
	"baseline_risk": req(text()), "owner": req(text()), "purpose": req(text()),
	"review_generation": req(integer(1, math.MaxInt64)), "expires_at": req(text()), "redirect_ids": req(list(text(), maxEntries)),
})

var keySchema = object(map[string]jsondecode.Field{
	"key_id": req(text()), "issuer_id": req(text()), "public_key": req(text()), "not_before": req(text()), "not_after": req(text()),
})

var documentSchema = object(map[string]jsondecode.Field{
	"format": req(text()),
	"policy": req(object(map[string]jsondecode.Field{
		"generation":  req(integer(1, math.MaxInt64)),
		"permissions": req(permissionsSchema),
		"issuers":     req(list(object(map[string]jsondecode.Field{"issuer_id": req(text()), "permissions": req(permissionsSchema)}), maxEntries)),
		"targets":     req(list(targetSchema, maxEntries)),
		"keys":        req(list(keySchema, task.MaxPolicyKeys)),
		// The local restriction inputs are optional; an absent member
		// restricts nothing and lists no release-reviewed task.
		"release_reviewed_tasks": opt(list(object(map[string]jsondecode.Field{"issuer_id": req(text()), "task_sha256": req(text())}), task.MaxReleaseReviewedTasks)),
		"device":                 opt(object(map[string]jsondecode.Field{"minimum_battery_percent": opt(integer(0, 100)), "charging_only": req(flag())})),
		"network_context":        opt(object(map[string]jsondecode.Field{"paths": req(pathsSchema), "endpoints": req(list(text(), task.MaxContextEndpoints))})),
	})),
	"configuration_approvals": req(object(map[string]jsondecode.Field{
		"endpoints": req(list(text(), maxApprovals)), "connection_codes": req(list(text(), maxApprovals)),
		"linkage_scopes": req(list(text(), maxApprovals)), "allow_linkage": req(flag()),
	})),
	"context_approvals": req(object(map[string]jsondecode.Field{
		"connection_codes": req(list(text(), maxEntries)), "classification_versions": req(list(text(), maxEntries)),
		"country_codes": req(list(text(), maxEntries)), "region_codes": req(list(text(), maxEntries)),
	})),
})

// Decode returns the complete document or ErrInvalid. It performs no I/O and
// reads no clock.
func Decode(b []byte) (doc Document, err error) {
	defer func() {
		if recover() != nil {
			doc, err = Document{}, ErrInvalid
		}
	}()
	v, err := jsondecode.Decode(b, documentSchema, limits)
	if err != nil {
		return Document{}, ErrInvalid
	}
	d := decoder{}
	if d.text(v, "format") != Format {
		return Document{}, ErrInvalid
	}
	p := d.member(v, "policy")
	in := task.PolicyInput{Generation: d.uint(p, "generation"), Permissions: d.permissions(d.member(p, "permissions"))}
	for _, e := range d.elements(p, "issuers") {
		in.Issuers = append(in.Issuers, task.IssuerPermissions{IssuerID: d.text(e, "issuer_id"), Permissions: d.permissions(d.member(e, "permissions"))})
	}
	for _, e := range d.elements(p, "targets") {
		in.Targets = append(in.Targets, d.target(e))
	}
	for _, e := range d.elements(p, "keys") {
		in.Keys = append(in.Keys, d.key(e))
	}
	for _, e := range d.elements(p, "release_reviewed_tasks") {
		in.ReleaseReviewed = append(in.ReleaseReviewed, task.ReleaseReviewedTask{IssuerID: d.text(e, "issuer_id"), TaskSHA256: d.digest(d.text(e, "task_sha256"))})
	}
	if dv, ok := p.Lookup("device"); ok {
		in.Device.ChargingOnly = d.flag(dv, "charging_only")
		if _, ok := dv.Lookup("minimum_battery_percent"); ok {
			n := d.int(dv, "minimum_battery_percent")
			in.Device.MinimumBatteryPercent = &n
		}
	}
	if nc, ok := p.Lookup("network_context"); ok {
		in.Context = &task.ContextEligibility{Paths: d.paths(d.member(nc, "paths")), Endpoints: d.texts(nc, "endpoints")}
	}
	a := d.member(v, "configuration_approvals")
	approvals := config.Approvals{Endpoints: d.texts(a, "endpoints"), ConnectionCodes: d.texts(a, "connection_codes"),
		LinkageScopes: d.texts(a, "linkage_scopes"), AllowLinkage: d.flag(a, "allow_linkage")}
	c := d.member(v, "context_approvals")
	context := observation.ContextApprovals{ConnectionCodes: d.texts(c, "connection_codes"), ClassificationVersions: d.texts(c, "classification_versions"),
		CountryCodes: d.texts(c, "country_codes"), RegionCodes: d.texts(c, "region_codes")}
	if d.bad {
		return Document{}, ErrInvalid
	}
	policy, err := task.NewLocalPolicy(in)
	if err != nil {
		return Document{}, ErrInvalid
	}
	// The configuration and context decoders validate the approvals they will
	// use, so an invalid document fails here rather than at Initialize.
	if _, err := config.NewDecoder(config.Contract, approvals); err != nil {
		return Document{}, ErrInvalid
	}
	if _, err := observation.NewContextDecoder(observation.Contract, context); err != nil {
		return Document{}, ErrInvalid
	}
	return Document{Policy: policy, Approvals: approvals, Context: context}, nil
}

// decoder reads schema-validated values. Any value the schema admits but the
// model cannot represent sets bad.
type decoder struct{ bad bool }

func (d *decoder) member(v *jsondecode.Value, name string) *jsondecode.Value {
	m, _ := v.Lookup(name)
	return m
}

func (d *decoder) elements(v *jsondecode.Value, name string) []*jsondecode.Value {
	m, ok := v.Lookup(name)
	if !ok {
		return nil
	}
	return m.Elements()
}

func (d *decoder) text(v *jsondecode.Value, name string) string {
	m, ok := v.Lookup(name)
	if !ok {
		return ""
	}
	return m.Text()
}

func (d *decoder) texts(v *jsondecode.Value, name string) []string {
	var out []string
	for _, e := range d.elements(v, name) {
		out = append(out, e.Text())
	}
	return out
}

func (d *decoder) flag(v *jsondecode.Value, name string) bool {
	m, ok := v.Lookup(name)
	return ok && m.Bool()
}

func (d *decoder) uint(v *jsondecode.Value, name string) uint64 {
	m, ok := v.Lookup(name)
	if !ok {
		return 0
	}
	n, err := strconv.ParseUint(m.Text(), 10, 64)
	if err != nil {
		d.bad = true
	}
	return n
}

func (d *decoder) int(v *jsondecode.Value, name string) int {
	m, ok := v.Lookup(name)
	if !ok {
		return 0
	}
	n, err := strconv.Atoi(m.Text())
	if err != nil {
		d.bad = true
	}
	return n
}

func (d *decoder) test(v *jsondecode.Value) task.TestVersion {
	return task.TestVersion{ID: d.text(v, "id"), Version: d.text(v, "version")}
}

func (d *decoder) paths(v *jsondecode.Value) task.Paths {
	return task.Paths{Direct: d.flag(v, "direct"), Proxy: d.flag(v, "proxy")}
}

func (d *decoder) risk(s string, unknown bool) task.Risk {
	switch s {
	case "low":
		return task.LowRisk
	case "medium":
		return task.MediumRisk
	case "high":
		return task.HighRisk
	case "unknown":
		if unknown {
			return task.UnknownRisk
		}
	}
	d.bad = true
	return 0
}

func (d *decoder) permissions(v *jsondecode.Value) task.Permissions {
	p := task.Permissions{RegistrationIDs: d.texts(v, "registration_ids"), Paths: d.paths(d.member(v, "paths")),
		MaxRisk: d.risk(d.text(v, "max_risk"), false), RequireConsent: d.flag(v, "require_consent"),
		AllowMediumRisk: d.flag(v, "allow_medium_risk"), AllowComposite: d.flag(v, "allow_composite"),
		AllowUnknownRisk: d.flag(v, "allow_unknown_risk"), LabelNamespaces: d.texts(v, "label_namespaces")}
	for _, e := range d.elements(v, "tests") {
		p.Tests = append(p.Tests, d.test(e))
	}
	l := d.member(v, "limits")
	a := d.member(l, "attempt")
	p.Limits = task.PolicyLimits{
		Attempt:                  task.AttemptLimits{DurationMS: d.int(a, "duration_ms"), BytesSent: d.int(a, "bytes_sent"), BytesReceived: d.int(a, "bytes_received"), Connections: d.int(a, "connections")},
		MaxRuns:                  d.int(l, "max_runs"),
		MaxAttempts:              d.int(l, "max_attempts"),
		MaxValiditySeconds:       d.int(l, "max_validity_seconds"),
		MinRepeatIntervalSeconds: d.int(l, "min_repeat_interval_seconds"),
		MinRepeatJitterSeconds:   d.int(l, "min_repeat_jitter_seconds"),
		MinRetryDelayMS:          d.int(l, "min_retry_delay_ms"),
		MaxRetryDelayMS:          d.int(l, "max_retry_delay_ms"),
		MaxRetryJitterMS:         d.int(l, "max_retry_jitter_ms"),
		MaxResultBytes:           d.int(l, "max_result_bytes"),
		MaxBundleBytes:           d.int(l, "max_bundle_bytes"),
		MaxResultsPerBundle:      d.int(l, "max_results_per_bundle"),
		MaxPendingBundles:        d.int(l, "max_pending_bundles"),
		MaxPendingBytes:          d.int(l, "max_pending_bytes"),
		MaxRetentionSeconds:      d.int(l, "max_retention_seconds"),
	}
	return p
}

func (d *decoder) role(s string) task.TargetRole {
	switch s {
	case "measurement":
		return task.MeasurementTarget
	case "resolver":
		return task.ResolverTarget
	case "validation":
		return task.ValidationTarget
	case "redirect":
		return task.RedirectTarget
	}
	d.bad = true
	return 0
}

// address accepts only the canonical text form of an address without a zone.
func (d *decoder) address(s string) netip.Addr {
	a, err := netip.ParseAddr(s)
	if err != nil || a.String() != s {
		d.bad = true
		return netip.Addr{}
	}
	return a
}

// timestamp accepts only whole UTC seconds in the form 2006-01-02T15:04:05Z.
func (d *decoder) timestamp(s string) task.Timestamp {
	const layout = "2006-01-02T15:04:05Z"
	t, err := time.Parse(layout, s)
	if err != nil || t.Format(layout) != s {
		d.bad = true
		return task.Timestamp{}
	}
	return task.Timestamp{Whole: t.UTC()}
}

func (d *decoder) target(v *jsondecode.Value) task.TargetRegistration {
	tv := d.member(v, "tuple")
	tuple := task.TargetTuple{Hostname: d.text(tv, "hostname"), Scheme: d.text(tv, "scheme"), URL: d.text(tv, "url"),
		ResolverName: d.text(tv, "resolver_name"), TLSServerName: d.text(tv, "tls_server_name"),
		ALPN: d.texts(tv, "alpn"), ExpectedStatus: d.int(tv, "expected_status")}
	if s, ok := tv.Lookup("address"); ok {
		tuple.Address = d.address(s.Text())
	}
	if _, ok := tv.Lookup("port"); ok {
		tuple.Port = uint16(d.int(tv, "port"))
	}
	for _, s := range d.texts(tv, "bootstrap_addresses") {
		tuple.BootstrapAddresses = append(tuple.BootstrapAddresses, d.address(s))
	}
	for _, e := range d.elements(tv, "body_checks") {
		tuple.BodyChecks = append(tuple.BodyChecks, task.BodyCheck{Type: d.text(e, "type"), Value: d.text(e, "value")})
	}
	for _, e := range d.elements(tv, "header_checks") {
		tuple.HeaderChecks = append(tuple.HeaderChecks, task.HeaderCheck{Name: d.text(e, "name"), Type: d.text(e, "type"), Value: d.text(e, "value")})
	}
	return task.TargetRegistration{ID: d.text(v, "id"), Test: d.test(d.member(v, "test")), Role: d.role(d.text(v, "role")),
		Tuple: tuple, Paths: d.paths(d.member(v, "paths")), BaselineRisk: d.risk(d.text(v, "baseline_risk"), true),
		Owner: d.text(v, "owner"), Purpose: d.text(v, "purpose"), ReviewGeneration: d.uint(v, "review_generation"),
		ExpiresAt: d.timestamp(d.text(v, "expires_at")), RedirectIDs: d.texts(v, "redirect_ids")}
}

// digest decodes a SHA-256 digest as exactly 64 lowercase hexadecimal digits.
func (d *decoder) digest(s string) [sha256.Size]byte {
	var out [sha256.Size]byte
	raw, err := hex.DecodeString(s)
	if err != nil || len(raw) != sha256.Size || hex.EncodeToString(raw) != s {
		d.bad = true
		return out
	}
	copy(out[:], raw)
	return out
}

// key decodes a public key as standard padded base64 of exactly 32 bytes.
func (d *decoder) key(v *jsondecode.Value) task.TrustedKey {
	k := task.TrustedKey{KeyID: d.text(v, "key_id"), IssuerID: d.text(v, "issuer_id"),
		NotBefore: d.timestamp(d.text(v, "not_before")), NotAfter: d.timestamp(d.text(v, "not_after"))}
	s := d.text(v, "public_key")
	raw, err := base64.StdEncoding.Strict().DecodeString(s)
	if err != nil || len(raw) != task.PublicKeySize || base64.StdEncoding.EncodeToString(raw) != s {
		d.bad = true
		return k
	}
	copy(k.PublicKey[:], raw)
	return k
}
