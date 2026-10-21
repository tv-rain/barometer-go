package admission

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/binary"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/task"
)

// MaxPackageBytes is the registered signed remote task package limit
// (REQ-LIMIT-0032).
const MaxPackageBytes = 98304

// signedContext is the fixed context string of the only message construction
// (REQ-TRUST-0049).
const signedContext = "barometer-signed-task-v1"

// issuedLayout is the fixed package time format: whole UTC seconds, Z suffix.
const issuedLayout = "2006-01-02T15:04:05Z"

var packageString = &jsondecode.Schema{Kind: jsondecode.String}

// packageSchema is the closed package object of REQ-TRUST-0048. Duplicate and
// unknown members are rejected by the strict decoder.
var packageSchema = &jsondecode.Schema{Kind: jsondecode.Object, Fields: map[string]jsondecode.Field{
	"issuer_id": {Schema: packageString, Required: true},
	"key_id":    {Schema: packageString, Required: true},
	"issued_at": {Schema: packageString, Required: true},
	"task":      {Schema: packageString, Required: true},
	"signature": {Schema: packageString, Required: true},
}}

var packageLimits = jsondecode.Limits{MaxBytes: MaxPackageBytes, MaxDepth: 2, MaxNodes: 6}

// signedPackage is a decoded package whose members meet their constraints. It
// is not verified. It never leaves admission and has no serialized form.
type signedPackage struct {
	issuerID, keyID, issuedText string
	issued                      time.Time
	task                        []byte
	signature                   [ed25519.SignatureSize]byte
}

func (signedPackage) String() string   { return "admission.signedPackage" }
func (signedPackage) GoString() string { return "admission.signedPackage" }

// decodePackage decodes b under REQ-TRUST-0048. Every failure is the same
// single result, so callers cannot tell which member failed.
// requirement-ref: REQ-TRUST-0048
func decodePackage(b []byte) (signedPackage, bool) {
	v, err := jsondecode.Decode(b, packageSchema, packageLimits)
	if err != nil {
		return signedPackage{}, false
	}
	text := func(name string) string { m, _ := v.Lookup(name); return m.Text() }
	p := signedPackage{issuerID: text("issuer_id"), keyID: text("key_id"), issuedText: text("issued_at")}
	if !packageIdentifier(p.issuerID) || !packageIdentifier(p.keyID) {
		return signedPackage{}, false
	}
	var ok bool
	if p.issued, ok = packageTime(p.issuedText); !ok {
		return signedPackage{}, false
	}
	if p.task, ok = canonicalBase64URL(text("task")); !ok || len(p.task) == 0 {
		return signedPackage{}, false
	}
	sig, ok := canonicalBase64URL(text("signature"))
	if !ok || len(sig) != ed25519.SignatureSize {
		return signedPackage{}, false
	}
	copy(p.signature[:], sig)
	return p, true
}

// packageIdentifier is 1 to 64 lowercase ASCII letters, digits, '.', '_' or '-'.
func packageIdentifier(s string) bool {
	if len(s) == 0 || len(s) > 64 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if c := s[i]; !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '.' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

// packageTime accepts exactly YYYY-MM-DDTHH:MM:SSZ naming a valid UTC calendar
// date and time from year 0001, with no fraction, offset, lowercase letter or
// leap second. Formatting the parsed time must reproduce the input exactly.
func packageTime(s string) (time.Time, bool) {
	if len(s) != len(issuedLayout) {
		return time.Time{}, false
	}
	t, err := time.Parse(issuedLayout, s)
	if err != nil || t.Year() < 1 || t.Format(issuedLayout) != s {
		return time.Time{}, false
	}
	return t.UTC(), true
}

// canonicalBase64URL decodes unpadded base64url (RFC 4648 section 5) and
// accepts only the canonical encoding: the alphabet only, so no padding or
// whitespace (which Go's decoder would otherwise skip), and zero unused
// trailing bits.
func canonicalBase64URL(s string) ([]byte, bool) {
	for i := 0; i < len(s); i++ {
		if c := s[i]; !(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_') {
			return nil, false
		}
	}
	b, err := base64.RawURLEncoding.Strict().DecodeString(s)
	if err != nil || base64.RawURLEncoding.EncodeToString(b) != s {
		return nil, false
	}
	return b, true
}

// signedMessage is M of REQ-TRUST-0049: each part length-prefixed by a 4-byte
// big-endian length.
// requirement-ref: REQ-TRUST-0049
func signedMessage(p signedPackage) []byte {
	parts := [][]byte{[]byte(signedContext), []byte(p.keyID), []byte(p.issuerID), []byte(p.issuedText), p.task}
	n := 0
	for _, part := range parts {
		n += 4 + len(part)
	}
	m := make([]byte, 0, n)
	for _, part := range parts {
		m = binary.BigEndian.AppendUint32(m, uint32(len(part)))
		m = append(m, part...)
	}
	return m
}

// verifySignature is cofactorless RFC 8032 Ed25519 verification. Go's
// crypto/ed25519 rejects S not below the group order and, by comparing the
// recomputed R encoding with the signature bytes, a non-canonical R
// (REQ-TRUST-0050). Public keys were validated with the policy.
// requirement-ref: REQ-TRUST-0050
func verifySignature(key [task.PublicKeySize]byte, p signedPackage) bool {
	return ed25519.Verify(ed25519.PublicKey(key[:]), signedMessage(p), p.signature[:])
}

// verifyPackage performs REQ-TRUST-0052 checks 1 to 9 in order on the exact
// presented package and returns the class of the first failing check. On
// success it returns the remote source completed from the verified package,
// owned task bytes and the acceptance time read at check 6. The task bytes are
// not parsed or evaluated here. No result carries the package, key ID, key,
// signature, issuer or which check failed (REQ-TRUST-0035). Every presentation
// of a package is verified at its own acceptance time before replay handling,
// including an exact redelivery and a redelivery after policy replacement,
// reinitialization or clearing; nothing retained exempts it (REQ-TRUST-0055).
// requirement-ref: REQ-TRUST-0043
// requirement-ref: REQ-TRUST-0052
// requirement-ref: REQ-TRUST-0055
func (a *Admitter) verifyPackage(ctx context.Context, pending task.SourceContext, presented []byte) (task.SourceContext, []byte, time.Time, task.AdmissionCode) {
	fail := func(c task.AdmissionCode) (task.SourceContext, []byte, time.Time, task.AdmissionCode) {
		return task.SourceContext{}, nil, time.Time{}, c
	}
	// requirement-ref: REQ-LIMIT-0032
	if len(presented) > MaxPackageBytes { // 1
		return fail(task.LimitExceeded)
	}
	if !a.policy.Valid() { // 2
		return fail(task.PolicyUnavailable)
	}
	p, ok := decodePackage(presented) // 3
	if !ok {
		return fail(task.SourceUnauthenticated)
	}
	key, ok := a.policy.TrustedKey(p.keyID) // 4
	if !ok {
		return fail(task.SourceUnauthenticated)
	}
	if key.IssuerID != p.issuerID { // 5
		return fail(task.SourceUnauthenticated)
	}
	now, err := a.clock.Now(ctx) // 6
	if err != nil || now.IsZero() {
		return fail(task.TrustUnavailable)
	}
	if !key.Covers(now) || !key.Covers(p.issued) { // 7
		return fail(task.SourceUnauthenticated)
	}
	// requirement-ref: REQ-LIMIT-0010
	if p.issued.After(now.Add(MaxIssuedAhead)) { // 8
		return fail(task.SourceUnauthenticated)
	}
	if !verifySignature(key.PublicKey, p) { // 9
		return fail(task.SourceUnauthenticated)
	}
	source, err := pending.WithVerifiedPackage(p.issuerID, task.Timestamp{Whole: p.issued})
	if err != nil {
		return fail(task.SourceUnauthenticated)
	}
	return source, p.task, now, 0
}
