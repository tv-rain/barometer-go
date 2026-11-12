package barometer

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"log"
	"net"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/measurement"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
	"github.com/tv-rain/barometer-go/internal/persistence/linkage"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/registry"
	"github.com/tv-rain/barometer-go/internal/testkit/capability"
	"github.com/tv-rain/barometer-go/internal/testkit/contextendpoint"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// The Go-owned result-linkage case families: CC-LINKAGE-GOALS-AND-NON-GOALS,
// CC-LINKAGE-INSTALLATION-SECRET, CC-LINKAGE-STORAGE-FAILURE and
// CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS. They run through the production
// Client, pending store and linkage source with a deterministic clock,
// test-scoped key stores holding synthetic secrets, and recording sinks: the
// synthetic Network Context endpoints, a scripted CONNECT proxy, a recording
// measurement target, the runner's inputs, the standard logger and every value
// and error the Client returns. The key stores are not protected storage; key
// creation, protection, backup and platform reset belong to the native
// adapters (docs/conformance/native-linkage.md). Synthetic evidence only, with
// no native key-store, deployment or support claim.

// linkageSecret is the synthetic installation secret canary.
func linkageSecret() [32]byte {
	var s [32]byte
	for i := range s {
		s[i] = byte(0xa5 ^ i*29)
	}
	return s
}

// secretForms are the spellings in which the secret canary is searched for;
// the canary harness adds the hexadecimal and base64 forms.
func secretForms(secret [32]byte) canarySet {
	return canarySet{{"installation secret", string(secret[:])}, {"installation secret (bytes)", fmt.Sprint(secret[:])}}
}

// derivationMessage is the REQ-LINKAGE-0010 message, built independently of
// internal/persistence/linkage.
func derivationMessage(scope string, epoch int64) []byte {
	return []byte("barometer-result-linkage-v1\x00" + scope + "\x00" + strconv.FormatInt(epoch, 10))
}

// linkageOfBundle decodes the linkage member of a bundle, or nil.
func linkageOfBundle(t *testing.T, b string) *bundleLinkage {
	t.Helper()
	var w struct {
		Linkage *bundleLinkage `json:"linkage"`
	}
	if err := json.Unmarshal([]byte(b), &w); err != nil {
		t.Fatal("not a bundle")
	}
	return w.Linkage
}

// nextAcknowledged returns the next pending bundle and acknowledges it.
func nextAcknowledged(t *testing.T, c *Client) string {
	t.Helper()
	b, err := c.NextBundle()
	if err != nil || b == "" {
		t.Fatal("no bundle:", codeOf(err))
	}
	if err := c.AcknowledgeBundle(bundleID(t, b)); err != nil {
		t.Fatal("acknowledge failed:", codeOf(err))
	}
	return b
}

// drawnKeys is a test-scoped key store whose key creation draws 32 bytes from
// a controlled random source, so a test knows every secret it creates. Like
// the platform contract it creates the key lazily on derivation and never
// returns it; Delete removes it.
type drawnKeys struct {
	mu       sync.Mutex
	random   io.Reader
	key      []byte
	drawn    [][32]byte
	messages [][]byte
}

func (k *drawnKeys) DeriveHMACSHA256(_ context.Context, message []byte) ([32]byte, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.messages = append(k.messages, append([]byte(nil), message...))
	if k.key == nil {
		var s [32]byte
		if _, err := io.ReadFull(k.random, s[:]); err != nil {
			return [32]byte{}, err
		}
		k.key, k.drawn = s[:], append(k.drawn, s)
	}
	m := hmac.New(sha256.New, k.key)
	m.Write(message)
	var out [32]byte
	copy(out[:], m.Sum(nil))
	return out, nil
}

func (k *drawnKeys) Delete(context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	clear(k.key)
	k.key = nil
	return nil
}

// countingReader is a deterministic random source: byte i is i*7+1.
type countingReader struct{ n byte }

func (r *countingReader) Read(p []byte) (int, error) {
	for i := range p {
		r.n++
		p[i] = r.n*7 + 1
	}
	return len(p), nil
}

// tokenFor is the expected token for a key, computed independently.
func tokenFor(key [32]byte, scope string, epoch int64) string {
	m := hmac.New(sha256.New, key[:])
	m.Write(derivationMessage(scope, epoch))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

// CC-LINKAGE-GOALS-AND-NON-GOALS: no stable identity continuity. Two
// deployment scopes under one installation secret give different tokens in
// the same epoch, two installations in one scope give different tokens, and
// the next epoch gives a new token. Continuity exists only through the
// protected key: a client bound to the same key store continues within the
// epoch, a fresh key store (a reinstall or data reset as the platform delivers
// it) and clearing both break it. No configuration, task or Client call can
// restore linkage from a token, secret, epoch or account value, select an
// installation for tasks, or request a stable mode. Reinstall, backup, restore
// and cross-application separation of the protected key itself are native
// rows (CC-LINKAGE-ANDROID-STORAGE, CC-LINKAGE-APPLE-PLATFORM-STORAGE).
// requirement-ref: REQ-LINKAGE-0002
func TestCcLinkageGoalsAndNonGoals(t *testing.T) {
	const otherScope = "org.example.other"
	secret := linkageSecret()
	run := func(t *testing.T, keys platform.LinkageKeyStore, scope string) (*harness, *Client, *bundleLinkage) {
		t.Helper()
		h := runningHarness(t)
		h.adapters.linkageKeys = keys
		h.adapters.approvals = config.Approvals{LinkageScopes: []string{linkageScope, otherScope}, AllowLinkage: true}
		c := h.client(t)
		cfg := fmt.Sprintf(`{"result_linkage":{"mode":"rotating_installation","scope":%q,"rotation_period_seconds":604800}}`, scope)
		if err := c.Initialize(cfg); err != nil {
			t.Fatal(codeOf(err))
		}
		completedRun(t, h, c, "goals")
		return h, c, linkageOfBundle(t, nextAcknowledged(t, c))
	}
	e := epoch.Unix() / 604800

	// Cross-deployment: one secret, two scopes.
	shared := capability.NewLinkageKeys(secret)
	_, _, a := run(t, shared, linkageScope)
	_, _, b := run(t, shared, otherScope)
	if a == nil || b == nil || a.Epoch != e || b.Epoch != e || a.Token == b.Token ||
		a.Token != tokenFor(secret, linkageScope, e) || b.Token != tokenFor(secret, otherScope, e) {
		t.Fatal("tokens of two deployment scopes are equal or not derived from the scope")
	}

	// Cross-installation: two secrets, one scope.
	_, _, other := run(t, capability.NewLinkageKeys([32]byte{1}), linkageScope)
	if other == nil || other.Token == a.Token {
		t.Fatal("two installations share a token")
	}

	// Continuity only through the key: the same key store continues within
	// the epoch, a fresh one does not, and the Go core carries no linkage
	// state into a new client.
	_, _, same := run(t, shared, linkageScope)
	_, _, fresh := run(t, &capability.LinkageKeys{}, linkageScope)
	if same == nil || same.Token != a.Token || fresh == nil || fresh.Token == a.Token {
		t.Fatal("continuity does not follow the protected key alone")
	}

	// Reset: clearing deletes the key, and the next bundle in the same epoch
	// carries a new token.
	h, c, before := run(t, capability.NewLinkageKeys(secret), linkageScope)
	if err := c.ClearMeasurementData(); err != nil {
		t.Fatal(codeOf(err))
	}
	if err := c.Initialize(rotatingConfig(604800)); err != nil {
		t.Fatal(codeOf(err))
	}
	completedRun(t, h, c, "after-reset")
	if after := linkageOfBundle(t, nextAcknowledged(t, c)); after == nil || after.Epoch != before.Epoch || after.Token == before.Token {
		t.Fatal("linkage survived a data reset")
	}

	// Rotation: the next epoch carries a new token. 2026-10-20T12:00Z plus 13
	// hours is in the next 86,400-second epoch.
	h, _ = linkedHarness(t, true)
	c = h.client(t)
	if err := c.Initialize(rotatingConfig(86400)); err != nil {
		t.Fatal(codeOf(err))
	}
	completedRun(t, h, c, "this-epoch")
	first := linkageOfBundle(t, nextAcknowledged(t, c))
	advanceBoth(t, h, 13*time.Hour)
	completedRun(t, h, c, "next-epoch")
	if next := linkageOfBundle(t, nextAcknowledged(t, c)); next == nil || next.Epoch != first.Epoch+1 || next.Token == first.Token {
		t.Fatal("token continued across epochs")
	}

	// No restoration and no stable mode: a configuration carrying a token,
	// secret, epoch, account, device or backup value, or asking for a stable
	// identifier, is rejected before any key operation.
	link := func(extra string) string {
		return `{"result_linkage":{"mode":"rotating_installation","scope":"` + linkageScope + `","rotation_period_seconds":604800` + extra + `}}`
	}
	for name, cfg := range map[string]string{
		"token":              link(`,"token":"` + a.Token + `"`),
		"epoch":              link(`,"epoch":` + strconv.FormatInt(e, 10)),
		"secret":             link(`,"installation_secret":"` + base64.RawURLEncoding.EncodeToString(secret[:]) + `"`),
		"account":            link(`,"account_id":"synthetic-account"`),
		"device":             link(`,"device_id":"synthetic-device"`),
		"restore":            link(`,"restore_from_backup":true`),
		"top-level token":    `{"result_linkage":{"mode":"none"},"linkage_token":"` + a.Token + `"}`,
		"stable mode":        `{"result_linkage":{"mode":"stable_installation","scope":"` + linkageScope + `"}}`,
		"permanent rotation": link(`,"rotation_period_seconds":0`),
	} {
		keys := capability.NewLinkageKeys(secret)
		h := runningHarness(t)
		h.adapters.linkageKeys = keys
		h.adapters.approvals = config.Approvals{LinkageScopes: []string{linkageScope}, AllowLinkage: true}
		if codeOf(h.client(t).Initialize(cfg)) != "not_initialized" || len(keys.Messages()) != 0 {
			t.Errorf("%s: configuration accepted or reached the key store", name)
		}
	}

	// No targeted delivery: a task naming an installation, token or linkage
	// object is rejected, and no Client method takes linkage input.
	base := runTask("targeted", "", "", false)
	for name, member := range map[string]string{
		"installation": `"target_installation":"` + a.Token + `"`,
		"token":        `"linkage_token":"` + a.Token + `"`,
		"linkage":      `"linkage":{"mode":"rotating_installation","scope":"` + linkageScope + `","epoch":` + strconv.FormatInt(e, 10) + `,"token":"` + a.Token + `"}`,
	} {
		h, keys := linkedHarness(t, true)
		c := h.client(t)
		if err := c.Initialize(rotatingConfig(604800)); err != nil {
			t.Fatal(codeOf(err))
		}
		for where, body := range map[string]string{
			"top level": base[:len(base)-1] + "," + member + "}",
			"options":   strings.Replace(base, `"options":{}`, `"options":{`+member+`}`, 1),
			"input":     strings.Replace(base, `"input":{}`, `"input":{`+member+`}`, 1),
		} {
			if _, err := c.AdmitTask(body, h.embedded(t, body)); err == nil {
				t.Errorf("task with %s in %s admitted", name, where)
			}
		}
		if len(keys.Messages()) != 0 {
			t.Fatal("task admission reached the key store")
		}
	}
}

// CC-LINKAGE-INSTALLATION-SECRET, Go-owned rows. The key store draws each
// secret from a controlled random source, so the test knows every secret it
// creates. No key is created at Initialize, while linkage is disabled or
// before a bundle needs one; the first linked bundle creates exactly one 256-bit
// secret, and every token is the HMAC under that drawn secret of the message
// carrying only the purpose, the deployment scope and the epoch, so no host,
// task or device value reaches the secret or the derivation. Clearing deletes
// the secret and creates none; the next linked bundle draws a fresh one. No
// configuration can supply a secret, and the key-store contract has no import
// or export method. The secret never reaches a Client value or error or
// library-owned storage (TestLinkageSecretNotExposed and
// TestCcLinkagePrivacyAndSafetyRequirements assert the network, log and
// storage sinks). Protected creation, non-exportability, backup exclusion and
// platform deletion are native rows (docs/conformance/native-linkage.md).
// requirement-ref: REQ-LINKAGE-0007
// requirement-ref: REQ-LINKAGE-0008
// requirement-ref: REQ-LINKAGE-0009
func TestCcLinkageInstallationSecret(t *testing.T) {
	setup := func(t *testing.T, cfg string) (*harness, *Client, *drawnKeys) {
		t.Helper()
		keys := &drawnKeys{random: &countingReader{}}
		h := runningHarness(t)
		h.adapters.linkageKeys = keys
		h.adapters.approvals = config.Approvals{LinkageScopes: []string{linkageScope}, AllowLinkage: true}
		c := h.client(t)
		if err := c.Initialize(cfg); err != nil {
			t.Fatal(codeOf(err))
		}
		return h, c, keys
	}

	// Disabled: no key is ever created.
	for name, cfg := range map[string]string{"absent": `{}`, "mode none": `{"result_linkage":{"mode":"none"}}`} {
		h, c, keys := setup(t, cfg)
		completedRun(t, h, c, "disabled")
		if linkageOfBundle(t, nextAcknowledged(t, c)) != nil || len(keys.drawn) != 0 || len(keys.messages) != 0 {
			t.Fatalf("%s: a key was created or used", name)
		}
	}

	// Enabled: lazy creation of one drawn 256-bit secret.
	h, c, keys := setup(t, rotatingConfig(604800))
	if len(keys.drawn) != 0 || len(keys.messages) != 0 {
		t.Fatal("Initialize created or used a key")
	}
	if _, err := c.AdmitTask(runTask("lazy", "", "", false), h.embedded(t, runTask("lazy", "", "", false))); err != nil || len(keys.drawn) != 0 {
		t.Fatal("admission created a key")
	}
	e := epoch.Unix() / 604800
	for i := range 2 {
		completedRun(t, h, c, fmt.Sprintf("linked-%d", i))
		l := linkageOfBundle(t, nextAcknowledged(t, c))
		if len(keys.drawn) != 1 || len(keys.drawn[0]) != 32 || l == nil || l.Token != tokenFor(keys.drawn[0], linkageScope, e) {
			t.Fatalf("bundle %d: not derived from exactly one drawn 256-bit secret", i)
		}
	}
	for _, m := range keys.messages {
		if !bytes.Equal(m, derivationMessage(linkageScope, e)) {
			t.Fatal("a derivation message carries more than the purpose, scope and epoch")
		}
	}

	// Clearing deletes the secret and creates none; the next linked bundle
	// draws a fresh secret, never the deleted one.
	first := keys.drawn[0]
	if err := c.ClearMeasurementData(); err != nil {
		t.Fatal(codeOf(err))
	}
	if keys.key != nil || len(keys.drawn) != 1 {
		t.Fatal("clearing kept the secret or created a replacement")
	}
	if err := c.Initialize(rotatingConfig(604800)); err != nil || len(keys.drawn) != 1 {
		t.Fatal("reinitialization created a secret")
	}
	completedRun(t, h, c, "after-clear")
	l := linkageOfBundle(t, nextAcknowledged(t, c))
	if len(keys.drawn) != 2 || keys.drawn[1] == first || l == nil || l.Token != tokenFor(keys.drawn[1], linkageScope, e) {
		t.Fatal("the secret after clearing is not freshly drawn")
	}

	// The host cannot supply a secret or seed, at any configuration level.
	for name, cfg := range map[string]string{
		"secret":       `{"result_linkage":{"mode":"rotating_installation","scope":"` + linkageScope + `","rotation_period_seconds":604800,"secret":"c2VjcmV0"}}`,
		"key":          `{"result_linkage":{"mode":"rotating_installation","scope":"` + linkageScope + `","rotation_period_seconds":604800,"key":[1,2,3]}}`,
		"seed":         `{"result_linkage":{"mode":"rotating_installation","scope":"` + linkageScope + `","rotation_period_seconds":604800,"seed":"advertising-id"}}`,
		"top level":    `{"installation_secret":"c2VjcmV0","result_linkage":{"mode":"rotating_installation","scope":"` + linkageScope + `","rotation_period_seconds":604800}}`,
		"host value":   `{"result_linkage":{"mode":"rotating_installation","scope":"` + linkageScope + `","rotation_period_seconds":604800,"vendor_id":"synthetic"}}`,
		"null secret":  `{"result_linkage":{"mode":"rotating_installation","scope":"` + linkageScope + `","rotation_period_seconds":604800,"secret":null}}`,
		"linkage keys": `{"linkage_keys":{},"result_linkage":{"mode":"none"}}`,
	} {
		keys := &drawnKeys{random: &countingReader{}}
		h := runningHarness(t)
		h.adapters.linkageKeys = keys
		h.adapters.approvals = config.Approvals{LinkageScopes: []string{linkageScope}, AllowLinkage: true}
		if codeOf(h.client(t).Initialize(cfg)) != "not_initialized" || len(keys.messages) != 0 || len(keys.drawn) != 0 {
			t.Errorf("%s accepted or reached the key store", name)
		}
	}

	// The key-store contract derives and deletes only: no method creates
	// from, imports or exports key material.
	kt := reflect.TypeOf((*platform.LinkageKeyStore)(nil)).Elem()
	var names []string
	for i := range kt.NumMethod() {
		names = append(names, kt.Method(i).Name)
	}
	if strings.Join(names, ",") != "Delete,DeriveHMACSHA256" {
		t.Fatalf("key-store contract methods %v", names)
	}
}

// faultyKeys is a test-scoped key store that fails in one of the storage
// failure modes. Its errors carry canaries for a key alias, platform account
// information and token material, and a failed derivation also returns
// canary bytes, so a test can show none of them reaches any surface.
type faultyKeys struct {
	mu       sync.Mutex
	mode     string // "", "unavailable", "locked", "inconsistent" or "invalid"
	key      [32]byte
	calls    int
	deletes  int
	failSoft bool // when set, Delete fails too
}

const (
	faultAlias   = "barometer-linkage-alias-q7xk2"
	faultAccount = "platform-account-zv81m"
)

// faultBytes is returned with a failed derivation.
var faultBytes = [32]byte{0xde, 0xad, 0x5e, 0xc7, 0xe7, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b,
	0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b}

func (k *faultyKeys) set(mode string) { k.mu.Lock(); k.mode = mode; k.mu.Unlock() }

func (k *faultyKeys) DeriveHMACSHA256(ctx context.Context, message []byte) ([32]byte, error) {
	k.mu.Lock()
	mode := k.mode
	k.calls++
	k.mu.Unlock()
	switch mode {
	case "unavailable":
		return [32]byte{}, errors.New("protected storage unavailable for " + faultAlias + " of " + faultAccount)
	case "locked":
		<-ctx.Done()
		return faultBytes, errors.New("keystore locked: " + faultAlias)
	case "inconsistent":
		return faultBytes, errors.New("state generation mismatch for " + faultAlias + " (" + faultAccount + ")")
	case "invalid":
		return faultBytes, errors.New("stored key failed validation: " + base64.RawURLEncoding.EncodeToString(faultBytes[:]))
	}
	m := hmac.New(sha256.New, k.key[:])
	m.Write(message)
	var out [32]byte
	copy(out[:], m.Sum(nil))
	return out, nil
}

func (k *faultyKeys) Delete(context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.failSoft {
		return errors.New("delete failed for " + faultAlias)
	}
	k.deletes++
	return nil
}

// CC-LINKAGE-STORAGE-FAILURE. Each failure mode of the key store (unavailable,
// locked beyond the bounded operation, inconsistent state and a stored key
// that fails validation) omits the linkage object while the result is still
// delivered with its own single-use bundle_id; a token derived before the
// failure is never reused; the library stores nothing new, so no secret,
// token or identifier falls back to the record port; the bundle carries no
// host-visible identifier in place of the linkage; the store's error text,
// key alias, account information and returned bytes reach no value, error,
// bundle or stored record; and a key store that fails deletion does not block
// result deletion while clearing reports an explicit error. No
// linkage_unavailable diagnostic exists; REQ-LINKAGE-0039 permits but does not
// require one. Asking Flutter or the host to persist the secret has no path:
// no Client method returns or takes key material (TestLinkageSecretNotExposed).
// requirement-ref: REQ-LINKAGE-0037
// requirement-ref: REQ-LINKAGE-0038
// requirement-ref: REQ-LINKAGE-0039
func TestCcLinkageStorageFailure(t *testing.T) {
	canaries := canarySet{{"key alias", faultAlias}, {"platform account", faultAccount}, {"failed-derivation bytes", string(faultBytes[:])}}
	for _, mode := range []string{"unavailable", "locked", "inconsistent", "invalid"} {
		t.Run(mode, func(t *testing.T) {
			keys := &faultyKeys{key: linkageSecret()}
			h := runningHarness(t)
			h.adapters.linkageKeys = keys
			h.adapters.approvals = config.Approvals{LinkageScopes: []string{linkageScope}, AllowLinkage: true}
			c := h.client(t)
			var out []string
			note := func(s string, err error) {
				out = append(out, s)
				if err != nil {
					out = append(out, err.Error())
				}
			}
			note("", c.Initialize(rotatingConfig(604800)))

			completedRun(t, h, c, "before")
			before := linkageOfBundle(t, nextAcknowledged(t, c))
			if before == nil {
				t.Fatal("no linkage while storage works")
			}
			keys.set(mode)
			var ids []string
			for i := range 2 {
				completedRun(t, h, c, fmt.Sprintf("failed-%d", i))
				b, err := c.NextBundle()
				note(b, err)
				var top map[string]json.RawMessage
				if err != nil || json.Unmarshal([]byte(b), &top) != nil {
					t.Fatal("no bundle while storage fails")
				}
				if _, linked := top["linkage"]; linked || len(top) != 3 || len(bundleResults(t, b)) != 1 || strings.Contains(b, before.Token) {
					t.Fatalf("failed storage: bundle members %d, linkage or earlier token present", len(top))
				}
				ids = append(ids, bundleID(t, b))
				note("", c.AcknowledgeBundle(bundleID(t, b)))
			}
			if ids[0] == ids[1] {
				t.Fatal("bundle_id reused while linkage is unavailable")
			}
			if h.run.output.Len(persistence.PendingClass) != 0 {
				t.Fatal("an acknowledged bundle is still stored")
			}
			dump := h.run.output.Dump()
			if bytes.Contains(dump, []byte(before.Token)) {
				t.Fatal("a token falls back to the record port")
			}
			canaries.absent(t, "the record port", dump)

			// Recovery derives under the same key again, in a new bundle.
			keys.set("")
			completedRun(t, h, c, "recovered")
			if after := linkageOfBundle(t, nextAcknowledged(t, c)); after == nil || after.Token != before.Token {
				t.Fatal("no linkage after storage recovered")
			}

			// Deletion is never blocked: a key store that cannot delete leaves
			// no pending result and clearing reports an explicit error.
			keys.set(mode)
			keys.mu.Lock()
			keys.failSoft = true
			keys.mu.Unlock()
			completedRun(t, h, c, "pending")
			if h.run.output.Len(persistence.PendingClass) == 0 {
				t.Fatal("no pending result to clear")
			}
			err := c.ClearMeasurementData()
			note("", err)
			if err == nil || h.run.output.Len(persistence.PendingClass) != 0 {
				t.Fatal("clearing with a failing key store reported success or kept results")
			}
			note("", c.Close())
			for _, s := range out {
				canaries.absent(t, "a Client value or error", []byte(s))
			}
		})
	}
}

// recordingTarget is a loopback measurement target that keeps every byte it
// receives and echoes it back.
type recordingTarget struct {
	ln net.Listener
	mu sync.Mutex
	in []byte
	wg sync.WaitGroup
}

func listenRecording(t *testing.T) *recordingTarget {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	r := &recordingTarget{ln: ln}
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			r.wg.Add(1)
			go func() {
				defer r.wg.Done()
				defer c.Close()
				buf := make([]byte, 4096)
				for {
					n, err := c.Read(buf)
					if n > 0 {
						r.mu.Lock()
						r.in = append(r.in, buf[:n]...)
						r.mu.Unlock()
						if _, werr := c.Write(buf[:n]); werr != nil {
							return
						}
					}
					if err != nil {
						return
					}
				}
			}()
		}
	}()
	t.Cleanup(func() { _ = ln.Close(); r.wg.Wait() })
	return r
}

func (r *recordingTarget) received() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]byte(nil), r.in...)
}

// tapedRunner is the echo runner that also records the task values each run
// receives.
type tapedRunner struct {
	*echoRunner
	mu     sync.Mutex
	inputs []string
}

func (r *tapedRunner) Run(ctx context.Context, in measurement.Input) (measurement.Output, error) {
	r.mu.Lock()
	r.inputs = append(r.inputs, fmt.Sprintf("%#v %#v", in.Parameters, in.Composite))
	r.mu.Unlock()
	return r.echoRunner.Run(ctx, in)
}

// productionSourceSinks returns, for every non-test Go file of the module, the
// logging, printing, crash-report and analytics sinks it uses.
func productionSourceSinks(t *testing.T) []string {
	t.Helper()
	forbiddenImports := map[string]bool{"log": true, "log/slog": true, "log/syslog": true, "expvar": true, "net/http/pprof": true, "runtime/trace": true}
	var found []string
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != "." && (strings.HasPrefix(d.Name(), ".") || d.Name() == "testdata" || d.Name() == "scripts") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			if p, _ := strconv.Unquote(imp.Path.Value); forbiddenImports[p] {
				found = append(found, path+" imports "+p)
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.SelectorExpr:
				if id, ok := x.X.(*ast.Ident); ok {
					switch {
					case id.Name == "os" && (x.Sel.Name == "Stdout" || x.Sel.Name == "Stderr"),
						id.Name == "fmt" && strings.HasPrefix(x.Sel.Name, "Print"),
						id.Name == "debug" && (x.Sel.Name == "PrintStack" || x.Sel.Name == "SetCrashOutput" || x.Sel.Name == "WriteHeapDump"):
						found = append(found, path+" uses "+id.Name+"."+x.Sel.Name)
					}
				}
			case *ast.CallExpr:
				if id, ok := x.Fun.(*ast.Ident); ok && (id.Name == "print" || id.Name == "println") {
					found = append(found, path+" calls "+id.Name)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return found
}

// CC-LINKAGE-PRIVACY-AND-SAFETY-REQUIREMENTS, Go-owned rows.
//
// Defaults: without result_linkage no linkage is emitted and no key is
// created, and a task carrying linkage-enabling fields cannot enable it.
//
// Sinks: with rotating linkage, a proxy and egress Network Context configured,
// a key store holding the secret canary runs a direct and a proxy task. The
// secret appears in no Client value or error, bundle, stored record, Network
// Context request, CONNECT request, target byte, runner input or standard log
// output; the token appears only as each bundle's linkage.token; and the
// context requests, the proxy and the target receive no bundle_id, scope,
// epoch or token. The library has no log, diagnostic, crash-report or
// analytics sink: no production file logs, prints or registers one.
//
// Use: task admission and scheduling decisions and results are identical with
// linkage disabled and under two different tokens, and no derivation happens
// before a result is stored.
//
// Retention: the linkage source holds no token between derivations, so the
// previous epoch's token is not retained after rotation, and clearing makes
// the next bundle of the same epoch carry a new token.
//
// Naming the token, deployment use, raw-token retention, combination with
// other data and the host reset path are documentation and deployment rows of
// the review case.
// requirement-ref: REQ-LINKAGE-0043
// requirement-ref: REQ-LINKAGE-0044
// requirement-ref: REQ-LINKAGE-0046
// requirement-ref: REQ-LINKAGE-0047
// requirement-ref: REQ-LINKAGE-0048
// requirement-ref: REQ-LINKAGE-0049
func TestCcLinkagePrivacyAndSafetyRequirements(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		keys := &capability.LinkageKeys{}
		h := runningHarness(t)
		h.adapters.linkageKeys = keys
		h.adapters.approvals = config.Approvals{LinkageScopes: []string{linkageScope}, AllowLinkage: true}
		c := h.client(t)
		if err := c.Initialize(`{}`); err != nil {
			t.Fatal(codeOf(err))
		}
		body := runTask("enable", "", "", false)
		body = body[:len(body)-1] + `,"result_linkage":{"mode":"rotating_installation","scope":"` + linkageScope + `","rotation_period_seconds":86400}}`
		if _, err := c.AdmitTask(body, h.embedded(t, body)); err == nil {
			t.Fatal("a task carrying result_linkage was admitted")
		}
		completedRun(t, h, c, "default")
		if held, created, _ := keys.Held(); linkageOfBundle(t, nextAcknowledged(t, c)) != nil || held || created != 0 || len(keys.Messages()) != 0 {
			t.Fatal("linkage emitted or a key created by default")
		}
	})

	t.Run("sinks", func(t *testing.T) {
		var logged bytes.Buffer
		prevOut, prevFlags := log.Writer(), log.Flags()
		log.SetOutput(&logged)
		t.Cleanup(func() { log.SetOutput(prevOut); log.SetFlags(prevFlags) })

		h, echo, e := contextHarness(t, contextendpoint.Response{Body: nlClassified})
		target := listenRecording(t)
		echo.address = target.ln.Addr().String()
		runner := &tapedRunner{echoRunner: echo}
		reg, err := registry.New([]registry.Entry{{Test: tcp01, Codec: tcpCodec{}, BaselineRisk: task.LowRisk, Targets: loopbackTargets(echo.address), Proxy: proxyCompatible{}, Runner: runner}})
		if err != nil {
			t.Fatal(err)
		}
		h.adapters.registry = reg
		in := synPolicyInput(2, "reg-a", func(p *task.Permissions) { p.Tests = []task.TestVersion{tcp01} })
		in.Targets[0].Test = tcp01
		in.Targets[0].Tuple = *loopbackTargets(echo.address).tuple
		if h.adapters.policy, err = task.NewLocalPolicy(in); err != nil {
			t.Fatal(err)
		}
		proxy, err := loopback.ListenProxy(loopback.ProxyConfig{Resolve: map[string]string{"8.8.8.8:443": e.primary.Address(), "9.9.9.9:443": e.fallback.Address()}})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = proxy.Close() })
		secret := linkageSecret()
		keys := capability.NewLinkageKeys(secret)
		h.adapters.linkageKeys = keys
		h.adapters.approvals.LinkageScopes, h.adapters.approvals.AllowLinkage = []string{linkageScope}, true

		c := h.client(t)
		var out []string
		note := func(s string, err error) {
			out = append(out, s)
			if err != nil {
				out = append(out, err.Error())
			}
		}
		cfg := `{"proxy":{"host":"127.0.0.1","port":` + strconv.Itoa(proxy.Port()) + `,"connection_code":"outline_mobileproxy"},` +
			`"network_context":{"primary_endpoint":"` + contextPrimary + `","fallback_endpoint":"` + contextFallback + `"},` +
			`"result_linkage":{"mode":"rotating_installation","scope":"` + linkageScope + `","rotation_period_seconds":604800}}`
		note("", c.Initialize(cfg))
		var bundles []string
		run := func(id, execution string) {
			taskJSON := runTask(id, "", execution, false)
			th, err := c.AdmitTask(taskJSON, h.embedded(t, taskJSON))
			note(th, err)
			rh, err := c.StartRun(th)
			note(rh, err)
			if err != nil {
				t.Fatal(codeOf(err))
			}
			waitTerminal(t, c, rh)
			note(c.RunState(rh))
			b, err := c.NextBundle()
			if err != nil || b == "" {
				t.Fatal("no bundle:", codeOf(err))
			}
			bundles = append(bundles, b)
			note("", c.AcknowledgeBundle(bundleID(t, b)))
		}
		run("direct", "")
		run("proxy", `"network_path":"proxy"`)
		if len(e.primary.Requests()) < 2 || len(proxy.Counts().Requests) < 2 || len(target.received()) == 0 || len(runner.inputs) != 2 {
			t.Fatalf("a sink saw no traffic: context %d, proxy %d, target %d bytes, runner %d", len(e.primary.Requests()), len(proxy.Counts().Requests), len(target.received()), len(runner.inputs))
		}
		stored := h.run.output.Dump()
		note("", c.ClearMeasurementData())
		note("", c.Close())

		material := canarySet{{"scope", linkageScope}, {"epoch", strconv.FormatInt(epoch.Unix()/604800, 10)}}
		var tokens []string
		for _, b := range bundles {
			l := linkageOfBundle(t, b)
			if l == nil || l.Token != tokenFor(secret, linkageScope, epoch.Unix()/604800) || strings.Count(b, l.Token) != 1 {
				t.Fatal("token missing from linkage.token or present elsewhere in the bundle")
			}
			tokens = append(tokens, l.Token)
			material = append(material, canary{"bundle_id " + bundleID(t, b), bundleID(t, b)})
		}
		material = append(material, canary{"token", tokens[0]})
		secretOnly := secretForms(secret)
		network := map[string][]byte{
			"the primary context endpoint":  e.primary.Raw(),
			"the fallback context endpoint": e.fallback.Raw(),
			"the proxy":                     []byte(strings.Join(proxy.Counts().Requests, "\n")),
			"the measurement target":        target.received(),
			"the runner input":              []byte(strings.Join(runner.inputs, "\n")),
			"the standard logger":           logged.Bytes(),
		}
		for surface, data := range network {
			material.absent(t, surface, data)
			secretOnly.absent(t, surface, data)
		}
		for _, s := range out {
			secretOnly.absent(t, "a Client value or error", []byte(s))
			canarySet{{"token", tokens[0]}}.absent(t, "a Client value or error", []byte(s))
		}
		for _, b := range bundles {
			secretOnly.absent(t, "a bundle", []byte(b))
		}
		secretOnly.absent(t, "the record port", stored)
		if sinks := productionSourceSinks(t); len(sinks) != 0 {
			t.Fatalf("production code has log or diagnostic sinks: %v", sinks)
		}
	})

	t.Run("equal treatment", func(t *testing.T) {
		transcript := func(keys *capability.LinkageKeys, cfg string) (string, *bundleLinkage) {
			h := runningHarness(t)
			h.adapters.linkageKeys = keys
			h.adapters.approvals = config.Approvals{LinkageScopes: []string{linkageScope}, AllowLinkage: true}
			c := h.client(t)
			if err := c.Initialize(cfg); err != nil {
				t.Fatal(codeOf(err))
			}
			var lines []string
			admit := func(name, taskJSON string) string {
				th, err := c.AdmitTask(taskJSON, h.embedded(t, taskJSON))
				lines = append(lines, name+" "+codeOf(err))
				return th
			}
			once := runTask("once", "", "", false)
			th := admit("valid", once)
			admit("redelivered", once)
			admit("unsupported", strings.Replace(runTask("unsupported", "", "", false), `"test_id":"TCP-01"`, `"test_id":"NET-01"`, 1))
			admit("high risk", strings.Replace(runTask("risky", "", "", false), `"risk_level":"low"`, `"risk_level":"high"`, 1))
			repeat := admit("repeating", runTask("repeat", `"repeat":{"mode":"interval","interval":21600,"jitter":300},"limits":{"max_runs":2}`, "", false))
			if len(keys.Messages()) != 0 {
				t.Fatal("admission derived linkage")
			}
			start := func(name, handle string) {
				rh, err := c.StartRun(handle)
				lines = append(lines, name+" "+codeOf(err))
				if err == nil {
					waitTerminal(t, c, rh)
					s, _ := c.RunState(rh)
					lines = append(lines, name+" state "+s)
				}
			}
			start("one-time", th)
			start("one-time again", th)
			start("repeating", repeat)
			start("repeating before due", repeat)
			h.advance(7 * time.Hour)
			start("repeating when due", repeat)
			start("repeating after max_runs", repeat)
			for _, r := range h.run.output.Records() {
				var res contextResult
				if json.Unmarshal(r, &res) != nil {
					t.Fatal("unreadable result")
				}
				lines = append(lines, "result "+res.Execution.State+" "+res.Execution.Reason)
			}
			b, err := c.NextBundle()
			if err != nil {
				t.Fatal(codeOf(err))
			}
			return strings.Join(lines, "\n"), linkageOfBundle(t, b)
		}
		off, none := transcript(capability.NewLinkageKeys([32]byte{1}), `{}`)
		a, la := transcript(capability.NewLinkageKeys([32]byte{1}), rotatingConfig(604800))
		b, lb := transcript(capability.NewLinkageKeys([32]byte{2}), rotatingConfig(604800))
		if none != nil || la == nil || lb == nil || la.Token == lb.Token {
			t.Fatal("the three configurations did not give no linkage and two different tokens")
		}
		if off != a || off != b {
			t.Fatalf("decisions differ with linkage:\ndisabled:\n%s\ntoken A:\n%s\ntoken B:\n%s", off, a, b)
		}
	})

	t.Run("retention", func(t *testing.T) {
		// The linkage source holds only the scope, the period, the epoch floor
		// and a clear counter: no field can keep a token between derivations.
		st := reflect.TypeOf(linkage.Source{})
		for i := range st.NumField() {
			switch f := st.Field(i); f.Type.Kind() {
			case reflect.String, reflect.Int64, reflect.Uint64:
			case reflect.Struct:
				if f.Type != reflect.TypeOf(sync.Mutex{}) {
					t.Fatalf("linkage.Source.%s may retain linkage material", f.Name)
				}
			default:
				t.Fatalf("linkage.Source.%s may retain linkage material", f.Name)
			}
		}

		h, _ := linkedHarness(t, true)
		c := h.client(t)
		if err := c.Initialize(rotatingConfig(86400)); err != nil {
			t.Fatal(codeOf(err))
		}
		completedRun(t, h, c, "epoch-1")
		old := linkageOfBundle(t, nextAcknowledged(t, c))
		advanceBoth(t, h, 13*time.Hour)
		completedRun(t, h, c, "epoch-2")
		current := linkageOfBundle(t, nextAcknowledged(t, c))
		if old == nil || current == nil || current.Epoch != old.Epoch+1 || current.Token == old.Token {
			t.Fatal("no rotation")
		}
		completedRun(t, h, c, "pending")
		if bytes.Contains(h.run.output.Dump(), []byte(old.Token)) {
			t.Fatal("the previous epoch's token is retained")
		}
		if err := c.ClearMeasurementData(); err != nil {
			t.Fatal(codeOf(err))
		}
		if err := c.Initialize(rotatingConfig(86400)); err != nil {
			t.Fatal(codeOf(err))
		}
		completedRun(t, h, c, "after-clear")
		if l := linkageOfBundle(t, nextAcknowledged(t, c)); l == nil || l.Epoch != current.Epoch || l.Token == current.Token {
			t.Fatal("a cached token survived clearing")
		}
	})
}
