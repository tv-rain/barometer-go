package admission_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/admission"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/testkit/delivery"
)

// secondSigner is another trusted key of the remote issuer.
func secondSigner() delivery.Signer { return delivery.NewSigner("remote-key-2", "remote-issuer", 3) }

// keyedPolicy is the permissive policy with a second remote-issuer key, and
// the remote key's validity replaced when nb and na are set.
func keyedPolicy(t *testing.T, nb, na time.Time) task.LocalPolicy {
	t.Helper()
	in := permissive()
	if !nb.IsZero() {
		in.Keys[0] = signer().Key(nb, na)
	}
	in.Keys = append(in.Keys, secondSigner().Key(keyNotBefore, keyNotAfter))
	return mustPolicy(t, in)
}

// rawPackage encodes arbitrary member values, for malformed packages.
func rawPackage(members map[string]any) []byte {
	b, err := json.Marshal(members)
	if err != nil {
		panic(err)
	}
	return b
}

func members(p delivery.Package) map[string]any {
	return map[string]any{"issuer_id": p.IssuerID, "key_id": p.KeyID, "issued_at": p.IssuedAt,
		"task": base64.RawURLEncoding.EncodeToString(p.Task), "signature": base64.RawURLEncoding.EncodeToString(p.Signature)}
}

// gateRemote presents pkg on a fresh harness with the given policy and returns
// the class and whether the task codec was reached.
func gateRemote(t *testing.T, policy task.LocalPolicy, prepare func(*harness), pkg []byte) (task.AdmissionCode, bool) {
	t.Helper()
	h := newHarnessWith(t, policy)
	ref, err := h.store.InstallRemote(context.Background(), 2, pkg)
	if err != nil {
		t.Fatal("remote install rejected")
	}
	if prepare != nil {
		prepare(h) // after install, so clock calls count from resolution
	}
	_, err = h.admitter.Gate(context.Background(), ref, pkg)
	return code(err), h.normalize.Load() != 0
}

// Every row of the internal signed-package validation plan, each changing one
// property of an otherwise valid package and asserting the class of the first
// failing REQ-TRUST-0052 check. A rejected package never reaches the task
// codec, so its task bytes are not parsed.
// requirement-ref: REQ-TRUST-0043
// requirement-ref: REQ-TRUST-0046
// requirement-ref: REQ-TRUST-0047
// requirement-ref: REQ-TRUST-0048
// requirement-ref: REQ-TRUST-0049
// requirement-ref: REQ-TRUST-0050
// requirement-ref: REQ-TRUST-0052
// requirement-ref: REQ-LIMIT-0032
func TestCcTrustSignedRemotePackages(t *testing.T) {
	b := []byte(baseTask)
	policy := keyedPolicy(t, time.Time{}, time.Time{})
	s := signer()
	valid := s.Sign(epoch, b)
	resigned := func(edit func(*delivery.Package)) []byte {
		p := s.Sign(epoch, b)
		edit(&p)
		return s.Resign(p).Encode()
	}
	altered := func(edit func(*delivery.Package)) []byte {
		p := s.Sign(epoch, b)
		edit(&p)
		return p.Encode()
	}
	withMembers := func(edit func(map[string]any)) []byte {
		m := members(valid)
		edit(m)
		return rawPackage(m)
	}
	replaceIn := func(old, new string) []byte {
		return bytes.Replace(valid.Encode(), []byte(old), []byte(new), 1)
	}
	failClockAtCheck6 := func(h *harness) { h.clock.failFrom.Store(h.clock.calls.Load() + 2) }
	other := otherSigner()
	signedTask := base64.RawURLEncoding.EncodeToString(valid.Task)

	cases := []struct {
		name     string
		policy   task.LocalPolicy // the shared keyed policy when invalid
		noPolicy bool             // no policy installed
		prepare  func(*harness)
		pkg      []byte
		want     task.AdmissionCode
	}{
		{name: "valid signature", pkg: valid.Encode()},
		{name: "second key of the same issuer", pkg: secondSigner().Sign(epoch, b).Encode()},

		// Check 1: the signed-package byte limit, before anything is decoded.
		{name: "package one byte over the limit", pkg: append(bytes.Repeat([]byte(" "), admission.MaxPackageBytes-len(valid.Encode())+1), valid.Encode()...), want: task.LimitExceeded},
		{name: "valid package padded to exactly the limit", pkg: append(bytes.Repeat([]byte(" "), admission.MaxPackageBytes-len(valid.Encode())), valid.Encode()...)},
		{name: "oversized garbage", pkg: bytes.Repeat([]byte("!"), admission.MaxPackageBytes+1), want: task.LimitExceeded},

		// Check 2: trusted local policy.
		{name: "no policy installed", noPolicy: true, pkg: valid.Encode(), want: task.PolicyUnavailable},

		// Check 3: the package decodes and every member meets its constraints.
		{name: "not JSON", pkg: []byte("not a package"), want: task.SourceUnauthenticated},
		{name: "bare task", pkg: b, want: task.SourceUnauthenticated},
		{name: "array", pkg: []byte("[]"), want: task.SourceUnauthenticated},
		{name: "unknown member", pkg: withMembers(func(m map[string]any) { m["algorithm"] = "ed25519" }), want: task.SourceUnauthenticated},
		{name: "authentication claim member", pkg: withMembers(func(m map[string]any) { m["authenticated"] = true }), want: task.SourceUnauthenticated},
		{name: "duplicate member", pkg: replaceIn(`{`, `{"key_id":"remote-key-1",`), want: task.SourceUnauthenticated},
		{name: "escaped duplicate member", pkg: replaceIn(`{`, `{"key_`+"\\u0069"+`d":"remote-key-1",`), want: task.SourceUnauthenticated},
		{name: "missing issuer_id", pkg: withMembers(func(m map[string]any) { delete(m, "issuer_id") }), want: task.SourceUnauthenticated},
		{name: "missing key_id", pkg: withMembers(func(m map[string]any) { delete(m, "key_id") }), want: task.SourceUnauthenticated},
		{name: "missing issued_at", pkg: withMembers(func(m map[string]any) { delete(m, "issued_at") }), want: task.SourceUnauthenticated},
		{name: "missing task", pkg: withMembers(func(m map[string]any) { delete(m, "task") }), want: task.SourceUnauthenticated},
		{name: "missing signature", pkg: withMembers(func(m map[string]any) { delete(m, "signature") }), want: task.SourceUnauthenticated},
		{name: "member of wrong type", pkg: withMembers(func(m map[string]any) { m["issued_at"] = 1792411200 }), want: task.SourceUnauthenticated},
		{name: "task as an object", pkg: withMembers(func(m map[string]any) { m["task"] = json.RawMessage(b) }), want: task.SourceUnauthenticated},
		{name: "null member", pkg: withMembers(func(m map[string]any) { m["signature"] = nil }), want: task.SourceUnauthenticated},
		{name: "padded task base64url", pkg: withMembers(func(m map[string]any) { m["task"] = base64.URLEncoding.EncodeToString(b) }), want: task.SourceUnauthenticated},
		{name: "padded signature base64url", pkg: withMembers(func(m map[string]any) { m["signature"] = base64.URLEncoding.EncodeToString(valid.Signature) }), want: task.SourceUnauthenticated},
		{name: "standard base64 alphabet", pkg: withMembers(func(m map[string]any) {
			m["signature"] = base64.RawStdEncoding.EncodeToString(bytes.Repeat([]byte{0xfb}, 64))
		}), want: task.SourceUnauthenticated},
		{name: "newline inside task base64url", pkg: withMembers(func(m map[string]any) { m["task"] = signedTask[:8] + "\n" + signedTask[8:] }), want: task.SourceUnauthenticated},
		{name: "nonzero trailing bits", pkg: withMembers(func(m map[string]any) {
			sig := base64.RawURLEncoding.EncodeToString(valid.Signature)
			m["signature"] = sig[:len(sig)-1] + "B" // 64 bytes leave 4 unused bits; "A" is the only canonical zero
		}), want: task.SourceUnauthenticated},
		{name: "empty task", pkg: withMembers(func(m map[string]any) { m["task"] = "" }), want: task.SourceUnauthenticated},
		{name: "signature 63 bytes", pkg: withMembers(func(m map[string]any) { m["signature"] = base64.RawURLEncoding.EncodeToString(valid.Signature[:63]) }), want: task.SourceUnauthenticated},
		{name: "signature 65 bytes", pkg: withMembers(func(m map[string]any) {
			m["signature"] = base64.RawURLEncoding.EncodeToString(append(bytes.Clone(valid.Signature), 0))
		}), want: task.SourceUnauthenticated},
		{name: "fractional issued_at", pkg: resigned(func(p *delivery.Package) { p.IssuedAt = "2026-10-19T12:00:00.5Z" }), want: task.SourceUnauthenticated},
		{name: "zero fraction issued_at", pkg: resigned(func(p *delivery.Package) { p.IssuedAt = "2026-10-19T12:00:00.0Z" }), want: task.SourceUnauthenticated},
		{name: "offset issued_at", pkg: resigned(func(p *delivery.Package) { p.IssuedAt = "2026-10-19T12:00:00+00:00" }), want: task.SourceUnauthenticated},
		{name: "lowercase issued_at", pkg: resigned(func(p *delivery.Package) { p.IssuedAt = "2026-10-19t12:00:00z" }), want: task.SourceUnauthenticated},
		{name: "leap second issued_at", pkg: resigned(func(p *delivery.Package) { p.IssuedAt = "2026-12-31T23:59:60Z" }), want: task.SourceUnauthenticated},
		{name: "invalid date issued_at", pkg: resigned(func(p *delivery.Package) { p.IssuedAt = "2026-02-30T12:00:00Z" }), want: task.SourceUnauthenticated},
		{name: "uppercase key_id", pkg: resigned(func(p *delivery.Package) { p.KeyID = "Remote-Key-1" }), want: task.SourceUnauthenticated},
		{name: "key_id over 64 bytes", pkg: resigned(func(p *delivery.Package) { p.KeyID = strings.Repeat("k", 65) }), want: task.SourceUnauthenticated},
		{name: "issuer_id over 64 bytes", pkg: resigned(func(p *delivery.Package) { p.IssuerID = strings.Repeat("i", 65) }), want: task.SourceUnauthenticated},
		{name: "empty issuer_id", pkg: resigned(func(p *delivery.Package) { p.IssuerID = "" }), want: task.SourceUnauthenticated},

		// Check 4: the key is trusted by the active policy.
		{name: "unknown key_id", pkg: delivery.NewSigner("unknown-key", "remote-issuer", 9).Sign(epoch, b).Encode(), want: task.SourceUnauthenticated},
		{name: "unknown key before unavailable time", prepare: failClockAtCheck6, pkg: delivery.NewSigner("unknown-key", "remote-issuer", 9).Sign(epoch, b).Encode(), want: task.SourceUnauthenticated},

		// Check 5: the key belongs to the package issuer; a task cannot move
		// between issuers.
		{name: "signed by another issuer's key", pkg: other.Resign(delivery.Package{IssuerID: "remote-issuer", KeyID: other.KeyID, IssuedAt: valid.IssuedAt, Task: b}).Encode(), want: task.SourceUnauthenticated},
		{name: "valid package re-wrapped under another issuer", pkg: altered(func(p *delivery.Package) { p.IssuerID = "embedded-issuer" }), want: task.SourceUnauthenticated},
		{name: "re-wrapped with the other issuer's key ID", pkg: altered(func(p *delivery.Package) { p.IssuerID, p.KeyID = "embedded-issuer", other.KeyID }), want: task.SourceUnauthenticated},

		// Check 6: trusted time.
		{name: "trusted time unavailable", prepare: failClockAtCheck6, pkg: valid.Encode(), want: task.TrustUnavailable},

		// Check 7: both times within the key validity, both ends inclusive.
		{name: "acceptance one second after not_after", policy: keyedPolicy(t, epoch.Add(-48*time.Hour), epoch.Add(-time.Second)), pkg: s.Sign(epoch.Add(-time.Hour), b).Encode(), want: task.SourceUnauthenticated},
		{name: "acceptance exactly at not_after", policy: keyedPolicy(t, epoch.Add(-48*time.Hour), epoch), pkg: s.Sign(epoch.Add(-time.Hour), b).Encode()},
		{name: "issued_at one second before not_before", policy: keyedPolicy(t, epoch, epoch.Add(48*time.Hour)), pkg: s.Sign(epoch.Add(-time.Second), b).Encode(), want: task.SourceUnauthenticated},
		{name: "issued_at exactly at not_before", policy: keyedPolicy(t, epoch, epoch.Add(48*time.Hour)), pkg: s.Sign(epoch, b).Encode()},
		{name: "key not yet valid", policy: keyedPolicy(t, epoch.Add(time.Second), epoch.Add(48*time.Hour)), pkg: s.Sign(epoch, b).Encode(), want: task.SourceUnauthenticated},
		{name: "issued_at after not_after", policy: keyedPolicy(t, epoch.Add(-48*time.Hour), epoch.Add(time.Minute)), pkg: s.Sign(epoch.Add(time.Minute+time.Second), b).Encode(), want: task.SourceUnauthenticated},
		{name: "expired key before unavailable time is not reached", policy: keyedPolicy(t, epoch.Add(-48*time.Hour), epoch.Add(-time.Second)), prepare: failClockAtCheck6, pkg: s.Sign(epoch.Add(-time.Hour), b).Encode(), want: task.TrustUnavailable},

		// Check 8: issued_at at most 300 seconds after acceptance.
		{name: "issued_at 300 seconds ahead", pkg: s.Sign(epoch.Add(admission.MaxIssuedAhead), b).Encode()},
		{name: "issued_at 301 seconds ahead", pkg: s.Sign(epoch.Add(admission.MaxIssuedAhead+time.Second), b).Encode(), want: task.SourceUnauthenticated},

		// Check 9: the signature over M.
		{name: "one task byte altered", pkg: altered(func(p *delivery.Package) { p.Task[len(p.Task)-2] = ' ' }), want: task.SourceUnauthenticated},
		{name: "task byte appended", pkg: altered(func(p *delivery.Package) { p.Task = append(p.Task, ' ') }), want: task.SourceUnauthenticated},
		{name: "issued_at one second later", pkg: altered(func(p *delivery.Package) { p.IssuedAt = "2026-10-19T12:00:01Z" }), want: task.SourceUnauthenticated},
		{name: "issued_at one second earlier", pkg: altered(func(p *delivery.Package) { p.IssuedAt = "2026-10-19T11:59:59Z" }), want: task.SourceUnauthenticated},
		{name: "key_id changed to another key of the issuer", pkg: altered(func(p *delivery.Package) { p.KeyID = "remote-key-2" }), want: task.SourceUnauthenticated},
		{name: "signature bit flipped", pkg: altered(func(p *delivery.Package) { p.Signature[10] ^= 1 }), want: task.SourceUnauthenticated},
		{name: "signature over another context", pkg: altered(func(p *delivery.Package) {
			m := bytes.Replace(delivery.Message(p.KeyID, p.IssuerID, p.IssuedAt, p.Task), []byte("barometer-signed-task-v1"), []byte("barometer-signed-task-v2"), 1)
			p.Signature = ed25519.Sign(ed25519.NewKeyFromSeed(s.Seed[:]), m)
		}), want: task.SourceUnauthenticated},
		{name: "signature over the bare task", pkg: altered(func(p *delivery.Package) { p.Signature = ed25519.Sign(ed25519.NewKeyFromSeed(s.Seed[:]), p.Task) }), want: task.SourceUnauthenticated},
		{name: "all-zero signature", pkg: altered(func(p *delivery.Package) { p.Signature = make([]byte, 64) }), want: task.SourceUnauthenticated},

		// Several failures: exactly one class, from the first failing check.
		{name: "oversized before missing policy", noPolicy: true, pkg: bytes.Repeat([]byte("!"), admission.MaxPackageBytes+1), want: task.LimitExceeded},
		{name: "missing policy before malformed", noPolicy: true, pkg: []byte("x"), want: task.PolicyUnavailable},
		{name: "malformed before unavailable time", prepare: failClockAtCheck6, pkg: []byte("x"), want: task.SourceUnauthenticated},
		{name: "bad signature over an invalid task", pkg: altered(func(p *delivery.Package) { p.Task = []byte(`{"schema_version":`) }), want: task.SourceUnauthenticated},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := policy
			switch {
			case c.noPolicy:
				p = task.LocalPolicy{}
			case c.policy.Valid():
				p = c.policy
			}
			got, parsed := gateRemote(t, p, c.prepare, c.pkg)
			if got != c.want {
				t.Fatalf("got class %d, want %d", got, c.want)
			}
			if c.want != 0 && parsed {
				t.Fatal("task bytes were parsed before the package verified")
			}
		})
	}

	// A valid signature does not make a task valid: admission continues with
	// the decoded task, which fails its own step.
	if got, _ := gateRemote(t, policy, nil, s.Sign(epoch, []byte(`{"schema_version":`)).Encode()); got != task.SchemaInvalid {
		t.Fatalf("validly signed invalid task: got class %d", got)
	}
	if got, _ := gateRemote(t, policy, nil, s.Sign(epoch, padded([]byte(baseTask), admission.MaxTaskBytes+1)).Encode()); got != task.LimitExceeded {
		t.Fatalf("validly signed oversized task: got class %d", got)
	}
}

// After admission neither the package, its signature nor its key ID remains:
// the replay record holds the issuer, task ID and the digest of the exact
// task bytes, and the single-use delivery record is gone, whether the task
// was accepted or rejected.
// requirement-ref: REQ-TRUST-0053
func TestCcTrustSignedPackageNotRetained(t *testing.T) {
	ctx := context.Background()
	b := []byte(baseTask)
	pkg := signed(0, b)
	p := signer().Sign(epoch, b)
	sig := base64.RawURLEncoding.EncodeToString(p.Signature)
	pkgDigest := sha256.Sum256(pkg)

	h := newHarness(t)
	ref, err := h.store.InstallRemote(ctx, 2, pkg)
	if err != nil {
		t.Fatal(err)
	}
	a, err := h.admitter.Admit(ctx, ref, pkg)
	if err != nil {
		t.Fatal("valid package rejected:", code(err))
	}
	rec := admission.AdmittedRecord(a)
	if rec.Digest != sha256.Sum256(b) || rec.Digest == pkgDigest || rec.Key.IssuerID != "remote-issuer" || rec.Origin != task.Remote {
		t.Fatal("replay record does not hold exactly the task identity")
	}
	retained := fmt.Sprintf("%#v %s", rec, h.replay.Dump())
	for _, secret := range []string{"remote-key-1", sig, string(pkg), fmt.Sprintf("%x", pkgDigest), base64.StdEncoding.EncodeToString(pkgDigest[:]), base64.RawURLEncoding.EncodeToString(pkgDigest[:])} {
		if strings.Contains(retained, secret) {
			t.Fatal("retained state contains verification material")
		}
	}
	if _, err := h.store.Resolve(ctx, ref, pkg); code(err) != task.SourceUnauthenticated {
		t.Fatal("the remote delivery record outlived admission")
	}
	// An exact redelivery needs a new reference and resolves to the same task.
	ref, _ = h.store.InstallRemote(ctx, 2, signer().Sign(epoch.Add(time.Minute), b).Encode())
	again, err := h.admitter.Admit(ctx, ref, signer().Sign(epoch.Add(time.Minute), b).Encode())
	if err != nil || !again.Redelivery() || again.Identity() != a.Identity() {
		t.Fatal("re-signed exact task bytes are not the same admitted task")
	}

	// Rejected packages leave no replay or delivery state either.
	h = newHarness(t)
	bad := p
	bad.Signature[0] ^= 1
	ref, _ = h.store.InstallRemote(ctx, 2, bad.Encode())
	if _, err := h.admitter.Admit(ctx, ref, bad.Encode()); code(err) != task.SourceUnauthenticated {
		t.Fatal("bad signature admitted")
	}
	if h.replay.Len() != 0 || h.replay.Commits() != 0 || h.replay.Staged() != 0 {
		t.Fatal("rejected package wrote state")
	}
	if _, err := h.store.Resolve(ctx, ref, bad.Encode()); code(err) != task.SourceUnauthenticated {
		t.Fatal("the rejected remote delivery record survived")
	}
}

// Verification errors carry only the fixed class text: no package, key ID,
// key, signature, issuer, task bytes, which check failed, or library errors.
// requirement-ref: REQ-TRUST-0035
// requirement-ref: REQ-TRUST-0052
func TestCcTrustSignedPackageErrorsDiscloseNothing(t *testing.T) {
	ctx := context.Background()
	b := taskWith(map[string]any{"input.subject": "CANARY-PAYLOAD"})
	s := signer()
	p := s.Sign(epoch, b)
	pub := s.PublicKey()
	canaries := []string{"CANARY-PAYLOAD", "remote-key-1", "remote-issuer", "embedded-issuer", base64.RawURLEncoding.EncodeToString(p.Signature)[:16],
		base64.RawURLEncoding.EncodeToString(pub[:])[:16], fmt.Sprintf("%x", pub[:8]), "ed25519", "base64", "check", "key", "signature", "CANARY-CLOCK-ERROR"}
	bad := p
	bad.Signature[5] ^= 1
	moved := p
	moved.IssuerID = "embedded-issuer"
	texts := map[string]bool{}
	for name, pkg := range map[string][]byte{
		"bad signature": bad.Encode(), "moved issuer": moved.Encode(), "unknown key": delivery.NewSigner("unknown-key", "remote-issuer", 9).Sign(epoch, b).Encode(),
		"malformed": append(p.Encode(), '}'), "future": s.Sign(epoch.Add(time.Hour), b).Encode(), "oversized": bytes.Repeat([]byte(" "), admission.MaxPackageBytes+1),
	} {
		h := newHarness(t)
		ref, _ := h.store.InstallRemote(ctx, 2, pkg)
		_, err := h.admitter.Gate(ctx, ref, pkg)
		if err == nil {
			t.Fatalf("%s admitted", name)
		}
		text := fmt.Sprintf("%v %+v %#v %s", err, err, err, err.Error())
		for _, c := range canaries {
			if strings.Contains(text, c) {
				t.Fatalf("%s: error discloses %q", name, c)
			}
		}
		texts[err.Error()] = true
	}
	if len(texts) != 2 || !texts["task_source_unauthenticated"] || !texts["task_limit_exceeded"] {
		t.Fatal("verification failures are distinguishable beyond their class")
	}
}

// FuzzSignedPackage presents arbitrary bytes as a remote package, seeded with
// valid and single-mutation packages. Every outcome is one public class or an
// acceptance, and an acceptance happens only for bytes that decode to a
// package whose signature over M verifies with the trusted key, checked
// independently of the admission code.
func FuzzSignedPackage(f *testing.F) {
	b := []byte(baseTask)
	valid := signer().Sign(epoch, b)
	f.Add(valid.Encode())
	f.Add(secondSigner().Sign(epoch, b).Encode())
	f.Add(otherSigner().Sign(epoch, b).Encode())
	for _, edit := range []func(*delivery.Package){
		func(p *delivery.Package) { p.Task[3] ^= 1 },
		func(p *delivery.Package) { p.IssuerID = "embedded-issuer" },
		func(p *delivery.Package) { p.KeyID = "remote-key-2" },
		func(p *delivery.Package) { p.IssuedAt = "2026-10-19T12:00:01Z" },
		func(p *delivery.Package) { p.Signature[63] |= 0xe0 },
	} {
		p := signer().Sign(epoch, b)
		edit(&p)
		f.Add(p.Encode())
	}
	f.Add([]byte(`{"issuer_id":"remote-issuer","key_id":"remote-key-1","issued_at":"2026-10-19T12:00:00Z","task":"","signature":""}`))
	f.Add([]byte("{}"))
	f.Add(b)
	known := map[string]delivery.Signer{"remote-key-1": signer(), "remote-key-2": secondSigner(), "embedded-key-1": otherSigner()}
	f.Fuzz(func(t *testing.T, pkg []byte) {
		if len(pkg) == 0 {
			return
		}
		h := newHarnessWith(t, keyedPolicy(t, time.Time{}, time.Time{}))
		ref, err := h.store.InstallRemote(context.Background(), 2, pkg)
		if err != nil {
			if len(pkg) <= task.MaxDeliveryBytes {
				t.Fatal("remote install of bounded bytes rejected")
			}
			return
		}
		g, err := h.admitter.Gate(context.Background(), ref, pkg)
		if err != nil {
			if code(err) == 0 {
				t.Fatal("rejection without a public class")
			}
			return
		}
		var m map[string]string
		if json.Unmarshal(pkg, &m) != nil || len(m) != 5 {
			t.Fatal("accepted bytes are not a five-member package")
		}
		k, ok := known[m["key_id"]]
		exact, err1 := base64.RawURLEncoding.DecodeString(m["task"])
		sig, err2 := base64.RawURLEncoding.DecodeString(m["signature"])
		pub := k.PublicKey()
		if !ok || k.IssuerID != m["issuer_id"] || err1 != nil || err2 != nil || !ed25519.Verify(pub[:], delivery.Message(m["key_id"], m["issuer_id"], m["issued_at"], exact), sig) {
			t.Fatal("accepted a package that does not verify independently")
		}
		if s := admission.GatedSource(g); s.IssuerID() != m["issuer_id"] {
			t.Fatal("source issuer is not the package issuer")
		}
	})
}
