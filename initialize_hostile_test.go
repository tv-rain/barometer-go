package barometer

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/tv-rain/barometer-go/internal/config"
)

// Synthetic host configuration members, matching the approvals of
// initHarness; none is a deployment endpoint, scope or connection code.
const (
	initProxy   = `{"host":"127.0.0.1","port":54321,"connection_code":"outline_mobileproxy"}`
	initNetwork = `{"primary_endpoint":"https://context.example/network-context","fallback_endpoint":"https://backup.example/network-context"}`
	initLinkage = `{"mode":"rotating_installation","scope":"org.example.production","rotation_period_seconds":604800}`
	initConfig  = `{"proxy":` + initProxy + `,"network_context":` + initNetwork + `,"result_linkage":` + initLinkage + `}`
)

// countingKeys is a linkage key store that only counts calls, so a test can
// show that a rejected configuration never reaches the protected-key adapter.
type countingKeys struct{ calls atomic.Int64 }

func (k *countingKeys) DeriveHMACSHA256(context.Context, []byte) ([32]byte, error) {
	k.calls.Add(1)
	return [32]byte{}, nil
}
func (k *countingKeys) Delete(context.Context) error { k.calls.Add(1); return nil }

// initHarness is a harness whose trusted approvals admit every member of
// initConfig and whose linkage key store counts calls.
func initHarness(t *testing.T) (*harness, *countingKeys) {
	t.Helper()
	h := newHarness(t)
	keys := &countingKeys{}
	h.adapters.linkageKeys = keys
	h.adapters.approvals = config.Approvals{
		Endpoints:       []string{"https://context.example/network-context", "https://backup.example/network-context"},
		ConnectionCodes: []string{"outline_mobileproxy"},
		LinkageScopes:   []string{"org.example.production"}, AllowLinkage: true,
	}
	return h, keys
}

// wantNotInitialized checks that a rejected Initialize left the client exactly
// as it was before the call: uninitialized, with no runtime generation
// activated and no protected-key adapter call.
func wantNotInitialized(t *testing.T, c *Client, keys *countingKeys, err error) {
	t.Helper()
	if codeOf(err) != "not_initialized" {
		t.Fatalf("got %s, want not_initialized", codeOf(err))
	}
	c.mu.Lock()
	st, rt, gen := c.state, c.rt, c.generation
	c.mu.Unlock()
	if st != uninitialized || rt != nil || gen != 0 || keys.calls.Load() != 0 {
		t.Fatal("rejected configuration activated a generation or reached the key store")
	}
}

// The strict JSON rules apply to the host configuration container at the
// public boundary (GO-SPEC-ISSUE-001, DEV-566): names compare after decoding
// escapes, duplicate names are rejected in every object, integers are exact
// and the complete-input limit is counted on original bytes. Each rejection
// leaves the client uninitialized and a later valid Initialize succeeds.
func TestInitializeStrictJSON(t *testing.T) {
	replace := func(old, new string) string {
		if !strings.Contains(initConfig, old) {
			t.Fatal("fixture edit does not apply:", old)
		}
		return strings.Replace(initConfig, old, new, 1)
	}
	escapeAll := func(s string) string {
		var b strings.Builder
		for _, r := range s {
			b.WriteString(esc(r))
		}
		return b.String()
	}
	escapedScope := replace(`"scope":"org.example.production"`, `"scope":"`+escapeAll("org.example.production")+`"`)
	pad := func(cfg string, total int) string { return cfg + strings.Repeat(" ", total-len(cfg)) }
	port := func(v string) string { return replace(`"port":54321`, `"port":`+v) }
	rotation := func(v string) string {
		return replace(`"rotation_period_seconds":604800`, `"rotation_period_seconds":`+v)
	}

	rejected := []struct{ name, cfg string }{
		{"equal duplicate top-level member", initConfig[:len(initConfig)-1] + `,"proxy":` + initProxy + `}`},
		{"escaped duplicate top-level member", initConfig[:len(initConfig)-1] + `,"pro` + esc('x') + `y":` + initProxy + `}`},
		{"conflicting duplicate proxy member", replace(`"port":54321`, `"port":54321,"port":54322`)},
		{"escaped duplicate proxy member", replace(`"host":"127.0.0.1"`, `"host":"127.0.0.1","h`+esc('o')+`st":"127.0.0.1"`)},
		{"duplicate network_context member", replace(`"primary_endpoint"`, `"fallback_endpoint":"https://backup.example/network-context","primary_endpoint"`)},
		{"escaped duplicate result_linkage member", replace(`"mode":"rotating_installation"`, `"mode":"rotating_installation","m`+esc('o')+`de":"none"`)},
		{"escaped unknown top-level name", initConfig[:len(initConfig)-1] + `,"` + esc('p') + `olicy":{}}`},
		{"escaped unknown nested name", replace(`"host":"127.0.0.1"`, `"host":"127.0.0.1","`+esc('u')+`ser":"x"`)},
		{"case variant of a member name", replace(`"proxy"`, `"Proxy"`)},
		{"negative zero port", port(`-0`)},
		{"decimal port spelling", port(`54321.0`)},
		{"exponent port spelling", port(`5.4321e4`)},
		{"quoted port", port(`"54321"`)},
		{"port beyond the exact float range", port(`9007199254740993`)},
		{"decimal rotation period spelling", rotation(`604800.0`)},
		{"exponent rotation period spelling", rotation(`6.048e5`)},
		{"negative zero rotation period", rotation(`-0`)},
		{"trailing value", initConfig + ` {}`},
		{"byte-order mark", "\xef\xbb\xbf" + initConfig},
		{"invalid UTF-8 in a member name", replace(`"proxy"`, "\"pro\xffxy\"")},
		{"invalid UTF-8 in a member value", replace(`"127.0.0.1"`, "\"127.0.0.1\xff\"")},
		{"lone surrogate escape", replace(`"org.example.production"`, `"org.example.production\ud800"`)},
		{"nesting beyond the parse depth", initConfig[:len(initConfig)-1] + `,"x":` + strings.Repeat("[", 200) + strings.Repeat("]", 200) + `}`},
		{"escaped configuration one byte over the limit", pad(escapedScope, config.MaxBytes+1)},
	}
	for _, tc := range rejected {
		t.Run(tc.name, func(t *testing.T) {
			h, keys := initHarness(t)
			c := h.client(t)
			wantNotInitialized(t, c, keys, c.Initialize(tc.cfg))
			if err := c.Initialize(initConfig); err != nil {
				t.Fatal("valid initialization after a rejection:", codeOf(err))
			}
		})
	}

	accepted := []struct{ name, cfg string }{
		{"complete configuration", initConfig},
		{"escaped spelling of a member name", replace(`"proxy"`, `"pro`+esc('x')+`y"`)},
		{"escaped spelling of an approved value", escapedScope},
		{"escaped configuration padded to exactly the limit", pad(escapedScope, config.MaxBytes)},
	}
	for _, tc := range accepted {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.cfg) > config.MaxBytes {
				t.Fatal("positive control exceeds the configuration limit")
			}
			h, _ := initHarness(t)
			c := h.client(t)
			if err := c.Initialize(tc.cfg); err != nil || stateOf(c) != initialized {
				t.Fatal("equivalent configuration rejected:", codeOf(err))
			}
		})
	}
}

// FuzzInitialize drives Client.Initialize with arbitrary configuration text.
// Every call either activates exactly one generation or returns
// not_initialized and leaves the client uninitialized with no key-store call;
// the outcome is the same as the internal decoder's for the same bytes, and a
// later valid initialization still succeeds after a rejection.
func FuzzInitialize(f *testing.F) {
	for _, seed := range []string{
		initConfig, `{}`, ``, `{} {}`, `null`, `[]`, "\xef\xbb\xbf{}",
		strings.Replace(initConfig, `"port":54321`, `"port":5.4321e4`, 1),
		strings.Replace(initConfig, `"proxy"`, `"pro`+esc('x')+`y"`, 1),
		initConfig[:len(initConfig)-1] + `,"proxy":` + initProxy + `}`,
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, cfg string) {
		h, keys := initHarness(t)
		d, err := config.NewDecoder(config.Contract, h.adapters.approvals)
		if err != nil {
			t.Fatal(err)
		}
		_, decodeErr := d.Decode([]byte(cfg))
		c := h.client(t)
		err = c.Initialize(cfg)
		if (err == nil) != (decodeErr == nil) {
			t.Fatalf("facade and decoder disagree: %s vs %v", codeOf(err), decodeErr)
		}
		if err != nil {
			wantNotInitialized(t, c, keys, err)
			if err := c.Initialize(`{}`); err != nil {
				t.Fatal("valid initialization after a rejection:", codeOf(err))
			}
			return
		}
		if stateOf(c) != initialized || keys.calls.Load() != 0 {
			t.Fatal("acceptance must activate one generation without deriving a key")
		}
	})
}
