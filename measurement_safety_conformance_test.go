package barometer

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/netip"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/measurement"
	http01 "github.com/tv-rain/barometer-go/internal/measurement/http01"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/testkit/loopback"
)

// The M-*-SAFETY families of the nine measurements run on this one shared
// table-driven harness. Each family supplies a safetyFixture: its registered
// Baseline Risk and Risk Factors rows, a production harness whose recording
// transport counts every contact, and its Safety Analysis action rows. The
// harness then exercises, through the production Client, codec, target rule,
// registry, admission and execution:
//
//   - the registered baseline in the production registry, and the consent
//     baseline of false, which adds no requirement;
//   - effective risk for every declared level and every Risk Factors row,
//     probed black-box against the issuer and local risk bounds: a policy that
//     permits exactly the expected level admits the task and one that permits
//     only the level below rejects it, so effective risk is exactly the most
//     restrictive input and never below the registered baseline;
//   - consent required by the task, the issuer or the deployment, which blocks
//     every attempt and contact until the host reports valid consent;
//   - the probe's final authority: local policy rejects, delays or skips a
//     valid task with no contact, and a registration withdrawn after admission
//     is refused at contact with local_policy_rejected.
//
// Synthetic adapters, policies and loopback endpoints only; this is portable
// evidence, not platform, deployment or field validation.

// riskLevels is the reviewed order; unknown is unordered and dominates.
var riskLevels = []task.Risk{task.LowRisk, task.MediumRisk, task.HighRisk, task.UnknownRisk}

func riskName(r task.Risk) string {
	switch r {
	case task.LowRisk:
		return "low"
	case task.MediumRisk:
		return "medium"
	case task.HighRisk:
		return "high"
	}
	return "unknown"
}

// expectedRisk is computed independently of the implementation.
func expectedRisk(levels ...task.Risk) task.Risk {
	out := task.LowRisk
	for _, l := range levels {
		if l == task.UnknownRisk {
			return task.UnknownRisk
		}
		out = max(out, l)
	}
	return out
}

// riskFactorTypes are the five registered Risk Factors rows of every test.
var riskFactorTypes = []string{"network_observability", "traffic_distinctiveness", "target_sensitivity", "data_sensitivity", "legal_or_social_context"}

type safetyAction struct {
	name string
	run  func(t *testing.T)
}

type safetyFixture struct {
	test     task.TestVersion
	baseline task.Risk
	// factors holds the registered level of each Risk Factors row.
	factors map[string]task.Risk
	// primitives are the referenced primitive tests of a composite.
	primitives []task.TestVersion
	// start returns a production harness and its recording transport's
	// contact count; a valid task with valid consent reaches the transport
	// unless traffic is false.
	start   func(t *testing.T) (*harness, func() int32)
	task    func(id string) string
	traffic bool
	actions []safetyAction
	// Protocol fixtures shared with the action rows.
	tls tlsFixture
	doh dohFixture
	dot dotFixture
}

// withSafety replaces the task's safety object.
func withSafety(t *testing.T, taskJSON, safety string) string {
	t.Helper()
	const base = `"safety":{"risk_level":"low","requires_user_consent":false}`
	if !strings.Contains(taskJSON, base) {
		t.Fatal("fixture task has no default safety object")
	}
	return strings.Replace(taskJSON, base, `"safety":`+safety, 1)
}

// declaring is a safety object with the overall level and one factor.
func declaring(level, factor, factorLevel string, consent bool) string {
	factors := `[]`
	if factor != "" {
		factors = fmt.Sprintf(`[{"type":%q,"level":%q}]`, factor, factorLevel)
	}
	return fmt.Sprintf(`{"risk_level":%q,"risk_factors":%s,"requires_user_consent":%t}`, level, factors, consent)
}

// editPolicy rebuilds the harness's active policy with edit applied. The
// fixtures build their policies from synPolicyInput, so the issuers, the
// registrations named by any permission and the two synthetic keys are the
// complete input; a fixture policy with any other member fails the rebuild.
func editPolicy(t *testing.T, h *harness, edit func(*task.PolicyInput)) {
	t.Helper()
	in := policyInputOf(t, h.adapters.policy)
	edit(&in)
	var err error
	if h.adapters.policy, err = task.NewLocalPolicy(in); err != nil {
		t.Fatal("edited policy rejected:", err)
	}
}

func policyInputOf(t *testing.T, p task.LocalPolicy) task.PolicyInput {
	t.Helper()
	in := task.PolicyInput{Generation: p.Generation(), Permissions: p.Permissions(), Device: p.Device()}
	if _, ok := p.ContextEligibility(); ok {
		t.Fatal("fixture policy carries context eligibility")
	}
	ids := slices.Clone(in.Permissions.RegistrationIDs)
	for _, id := range []string{"embedded-issuer", "remote-issuer"} {
		is, ok := p.Issuer(id)
		if !ok {
			t.Fatal("fixture policy lacks", id)
		}
		in.Issuers = append(in.Issuers, is)
		for _, r := range is.Permissions.RegistrationIDs {
			if !slices.Contains(ids, r) {
				ids = append(ids, r)
			}
		}
	}
	for _, id := range ids {
		r, ok := p.Target(id)
		if !ok {
			t.Fatal("fixture registration missing")
		}
		in.Targets = append(in.Targets, r)
	}
	for _, id := range []string{"remote-key-1", "embedded-key-1"} {
		if k, ok := p.TrustedKey(id); ok {
			in.Keys = append(in.Keys, k)
		}
	}
	return in
}

// bothSides applies edit to the deployment and to every issuer.
func bothSides(edit func(*task.Permissions)) func(*task.PolicyInput) {
	return func(in *task.PolicyInput) {
		edit(&in.Permissions)
		for i := range in.Issuers {
			edit(&in.Issuers[i].Permissions)
		}
	}
}

// riskBound permits exactly the given effective level on both sides.
func riskBound(level task.Risk) func(*task.Permissions) {
	return func(p *task.Permissions) {
		p.AllowUnknownRisk = level == task.UnknownRisk
		p.AllowMediumRisk = level == task.MediumRisk || level == task.HighRisk
		p.MaxRisk = level
		if level == task.UnknownRisk {
			p.MaxRisk = task.HighRisk
		}
	}
}

// belowBound permits every level below the given one, and nothing at it.
func belowBound(level task.Risk) func(*task.Permissions) {
	return func(p *task.Permissions) {
		switch level {
		case task.UnknownRisk:
			riskBound(task.HighRisk)(p)
		case task.HighRisk:
			riskBound(task.MediumRisk)(p)
		default:
			riskBound(task.LowRisk)(p)
		}
	}
}

// admitCode admits taskJSON on a fresh fixture harness after edit and
// returns the admission code and the contacts it made.
func (f safetyFixture) admitCode(t *testing.T, taskJSON string, edit func(*task.PolicyInput)) (string, int32) {
	t.Helper()
	h, contacts := f.start(t)
	if edit != nil {
		editPolicy(t, h, edit)
	}
	c := h.initialized(t)
	defer c.Close()
	_, code := h.tryAdmit(t, c, taskJSON)
	return code, contacts()
}

// probeRisk proves the effective risk of taskJSON is exactly want: admitted
// under a policy permitting exactly want, rejected under one permitting only
// the levels below it, and never contacting the transport at admission.
func (f safetyFixture) probeRisk(t *testing.T, name, taskJSON string, want task.Risk) {
	t.Helper()
	code, contacts := f.admitCode(t, taskJSON, bothSides(riskBound(want)))
	if code != "ok" || contacts != 0 {
		t.Fatalf("%s: admission under a %s bound = %s (contacts %d), want ok", name, riskName(want), code, contacts)
	}
	if want == task.LowRisk {
		return
	}
	code, contacts = f.admitCode(t, taskJSON, bothSides(belowBound(want)))
	if code != "task_issuer_not_authorized" || contacts != 0 {
		t.Fatalf("%s: admission below a %s bound = %s (contacts %d), want task_issuer_not_authorized", name, riskName(want), code, contacts)
	}
	// Either side alone holds the bound.
	for _, side := range []string{"issuer", "local"} {
		code, _ = f.admitCode(t, taskJSON, func(in *task.PolicyInput) {
			bothSides(riskBound(want))(in)
			if side == "local" {
				belowBound(want)(&in.Permissions)
				return
			}
			for i := range in.Issuers {
				belowBound(want)(&in.Issuers[i].Permissions)
			}
		})
		if code != "task_issuer_not_authorized" {
			t.Fatalf("%s: %s bound below %s not enforced: %s", name, side, riskName(want), code)
		}
	}
}

// runOnce admits taskJSON after edit and starts one run; it returns the
// StartRun code, the written results and the contacts.
func (f safetyFixture) runOnce(t *testing.T, taskJSON string, edit func(*task.PolicyInput), set func(*harness)) (string, [][]byte, int32) {
	t.Helper()
	h, contacts := f.start(t)
	if edit != nil {
		editPolicy(t, h, edit)
	}
	if set != nil {
		set(h)
	}
	c := h.initialized(t)
	defer c.Close()
	code := startCode(t, c, h.admit(t, c, taskJSON))
	return code, h.run.output.Records(), contacts()
}

func executionOf(t *testing.T, raw []byte) (string, string) {
	t.Helper()
	var r struct {
		Execution struct{ State, Reason string }
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatal(err)
	}
	return r.Execution.State, r.Execution.Reason
}

func runSafety(t *testing.T, f safetyFixture) {
	t.Helper()
	t.Run("baseline risk", func(t *testing.T) {
		reg, err := productionRegistry(nil, true)
		if err != nil {
			t.Fatal(err)
		}
		e, ok := reg.Lookup(f.test)
		if !ok || e.BaselineRisk != f.baseline {
			t.Fatalf("registered baseline of %s is not %s", f.test.ID, riskName(f.baseline))
		}
		// The registered baseline is the most restrictive Risk Factors row.
		levels := []task.Risk{}
		for _, typ := range riskFactorTypes {
			l, ok := f.factors[typ]
			if !ok {
				t.Fatal("unregistered factor row", typ)
			}
			levels = append(levels, l)
		}
		if expectedRisk(levels...) != f.baseline {
			t.Fatal("baseline below a registered factor")
		}
		for _, p := range f.primitives {
			pe, ok := reg.Lookup(p)
			if !ok || !pe.Step || pe.BaselineRisk != task.LowRisk {
				t.Fatalf("primitive %s is not a registered low-risk step", p.ID)
			}
		}
	})
	t.Run("baseline consent false", func(t *testing.T) {
		// A false consent baseline adds no requirement: with nothing else
		// requiring consent, the run starts while the host reports none, and
		// the consent capability is never queried.
		h, contacts := f.start(t)
		h.run.consent.Set(platform.ConsentAbsent)
		c := h.initialized(t)
		defer c.Close()
		th := h.admit(t, c, f.task("consent-baseline"))
		if admitted(t, c, th).ConsentRequired() {
			t.Fatal("consent required with a false baseline")
		}
		if code := startCode(t, c, th); code != "ok" || len(h.run.output.Records()) != 1 || h.run.consent.Calls() != 0 {
			t.Fatalf("run without a consent requirement: %s, consent calls %d", code, h.run.consent.Calls())
		}
		if f.traffic != (contacts() > 0) {
			t.Fatalf("contacts %d with traffic %t", contacts(), f.traffic)
		}
	})
	for _, declared := range riskLevels {
		t.Run("declared "+riskName(declared), func(t *testing.T) {
			level := riskName(declared)
			factor := ""
			if declared != task.LowRisk {
				factor = "target_sensitivity"
			}
			data := withSafety(t, f.task("declared-"+level), declaring(level, factor, level, false))
			f.probeRisk(t, level, data, expectedRisk(f.baseline, declared))
		})
	}
	for _, typ := range riskFactorTypes {
		t.Run("factor "+typ, func(t *testing.T) {
			registered := f.factors[typ]
			// At, below and above the registered level: the overall level is
			// the factor's, the effective risk never drops below the baseline.
			for _, l := range []task.Risk{task.LowRisk, task.MediumRisk, task.HighRisk} {
				name := typ + " " + riskName(l)
				data := withSafety(t, f.task("factor-"+riskName(l)), declaring(riskName(l), typ, riskName(l), false))
				f.probeRisk(t, name, data, expectedRisk(f.baseline, registered, l))
			}
			// An unknown factor keeps the overall level unknown: no trusted
			// local assessment resolves it, so it needs the reviewed exception.
			data := withSafety(t, f.task("factor-unknown"), declaring("unknown", typ, "unknown", false))
			f.probeRisk(t, typ+" unknown", data, task.UnknownRisk)
			// An unknown factor under a known level, or a factor above the
			// overall level, cannot be evaluated reliably and is rejected.
			for _, s := range []string{declaring("high", typ, "unknown", false), declaring("low", typ, "medium", false), declaring("medium", typ, "high", false)} {
				if code, contacts := f.admitCode(t, withSafety(t, f.task("factor-invalid"), s), nil); code != "task_schema_invalid" || contacts != 0 {
					t.Fatalf("%s: inconsistent factor admitted: %s", typ, code)
				}
			}
		})
	}
	t.Run("unrecognized factor type", func(t *testing.T) {
		for _, typ := range []string{"battery_drain", "Network_observability", "network-observability"} {
			if code, contacts := f.admitCode(t, withSafety(t, f.task("factor-type"), declaring("unknown", typ, "unknown", false)), bothSides(riskBound(task.UnknownRisk))); code != "task_unsupported" || contacts != 0 {
				t.Fatalf("unrecognized factor %q: %s", typ, code)
			}
		}
	})
	t.Run("consent required", func(t *testing.T) {
		sources := []struct {
			name        string
			task        bool
			edit        func(*task.PolicyInput)
			issuer, loc bool
		}{
			{"task", true, nil, false, false},
			{"issuer", false, func(in *task.PolicyInput) {
				for i := range in.Issuers {
					in.Issuers[i].Permissions.RequireConsent = true
				}
			}, true, false},
			{"local policy", false, func(in *task.PolicyInput) { in.Permissions.RequireConsent = true }, false, true},
		}
		statuses := []struct {
			name   string
			status platform.ConsentStatus
		}{{"absent", platform.ConsentAbsent}, {"ended", platform.ConsentEnded}, {"unavailable", platform.ConsentUnavailable}, {"valid", platform.ConsentValid}}
		for _, src := range sources {
			for _, st := range statuses {
				data := f.task("consent")
				if src.task {
					data = withSafety(t, data, declaring("low", "", "", true))
				}
				code, records, contacts := f.runOnce(t, data, src.edit, func(h *harness) { h.run.consent.Set(st.status) })
				runs := st.status == platform.ConsentValid
				if runs != (code == "ok" && len(records) == 1) {
					t.Fatalf("%s requires consent, host consent %s: start %s, %d results", src.name, st.name, code, len(records))
				}
				if !runs && contacts != 0 || runs && f.traffic != (contacts > 0) {
					t.Fatalf("%s requires consent, host consent %s: %d contacts", src.name, st.name, contacts)
				}
			}
		}
	})
	t.Run("probe authority", func(t *testing.T) {
		// Reject: a deployment that does not permit the task's network path,
		// or permits only a lower risk, rejects a valid task at admission
		// with no contact, whatever the issuer permits.
		if code, contacts := f.admitCode(t, f.task("authority-reject"), func(in *task.PolicyInput) {
			in.Permissions.Paths = task.Paths{Proxy: true}
		}); code != "task_issuer_not_authorized" || contacts != 0 {
			t.Fatalf("local rejection: %s, %d contacts", code, contacts)
		}
		if f.baseline > task.LowRisk {
			if code, _ := f.admitCode(t, f.task("authority-risk"), func(in *task.PolicyInput) { belowBound(f.baseline)(&in.Permissions) }); code != "task_issuer_not_authorized" {
				t.Fatalf("local risk bound: %s", code)
			}
		}
		// Delay and skip: a local device restriction holds every attempt of a
		// valid task while unsatisfied, with no attempt, result or contact,
		// and the same admitted task runs once it is satisfied.
		h, contacts := f.start(t)
		editPolicy(t, h, func(in *task.PolicyInput) { in.Device = task.DeviceConstraints{ChargingOnly: true} })
		h.run.device.Set(platform.DeviceState{ChargingKnown: true, BatteryPercent: 100, BatteryKnown: true})
		c := h.initialized(t)
		defer c.Close()
		th := h.admit(t, c, f.task("authority-delay"))
		for range 2 {
			if code := startCode(t, c, th); code != "run_not_due" || len(h.run.output.Records()) != 0 || contacts() != 0 {
				t.Fatalf("local device restriction not held: %s, contacts %d", code, contacts())
			}
		}
		h.run.device.Set(platform.DeviceState{Charging: true, ChargingKnown: true, BatteryPercent: 100, BatteryKnown: true})
		if code := startCode(t, c, th); code != "ok" || len(h.run.output.Records()) != 1 || f.traffic != (contacts() > 0) {
			t.Fatalf("delayed task did not run once permitted: %s", code)
		}
		if !f.traffic {
			return
		}
		// Withdrawal: registrations that lapse after admission are refused at
		// every contact; the attempt ends local_policy_rejected without one.
		h, contacts = f.start(t)
		editPolicy(t, h, func(in *task.PolicyInput) {
			for i := range in.Targets {
				in.Targets[i].ExpiresAt = task.TimestampFromTime(epoch.Add(time.Second))
			}
		})
		c2 := h.initialized(t)
		defer c2.Close()
		th = h.admit(t, c2, f.task("authority-withdrawn"))
		h.advance(2 * time.Second)
		if code := startCode(t, c2, th); code != "ok" {
			t.Fatal("withdrawn registration start:", code)
		}
		records := h.run.output.Records()
		if len(records) != 1 {
			t.Fatalf("%d results", len(records))
		}
		if state, reason := executionOf(t, records[0]); reason != "local_policy_rejected" || state == "completed" {
			t.Fatalf("withdrawn registration: %s %s", state, reason)
		}
		if contacts() != 0 {
			t.Fatalf("withdrawn registration contacted %d times", contacts())
		}
	})
	for _, a := range f.actions {
		t.Run(a.name, a.run)
	}
}

func lowFactors() map[string]task.Risk {
	m := map[string]task.Risk{}
	for _, typ := range riskFactorTypes {
		m[typ] = task.LowRisk
	}
	return m
}

// encryptedDNSFactors are the DNS-03 and DNS-04 rows; NET-05 differs only in
// target_sensitivity.
func mediumFactors(low string) map[string]task.Risk {
	m := map[string]task.Risk{}
	for _, typ := range riskFactorTypes {
		m[typ] = task.MediumRisk
	}
	m[low] = task.LowRisk
	return m
}

// safetyConn records every application byte written and read.
type safetyConn struct {
	tcpClosingConn
	wrote, read *int
}

func (c *safetyConn) Write(p []byte) (int, error) { *c.wrote += len(p); return len(p), nil }
func (c *safetyConn) Read(p []byte) (int, error)  { *c.read++; return 0, io.EOF }

// safetyTargetRefused registers a special-purpose, local or private tuple in
// place of the fixture's reviewed registration and requires that no contact
// follows: the policy, admission or the per-contact address gate refuses it.
func safetyTargetRefused(t *testing.T, start func(*testing.T) (*harness, func() int32), name string, tuple task.TargetTuple, taskJSON string) {
	t.Helper()
	h, contacts := start(t)
	in := policyInputOf(t, h.adapters.policy)
	in.Targets[0].Tuple = tuple
	policy, err := task.NewLocalPolicy(in)
	if err != nil {
		return // the trusted policy cannot register it
	}
	h.adapters.policy = policy
	c := h.initialized(t)
	defer c.Close()
	th, code := h.tryAdmit(t, c, taskJSON)
	if code == "ok" {
		if code = startCode(t, c, th); code == "ok" {
			for _, raw := range h.run.output.Records() {
				if state, _ := executionOf(t, raw); state == "completed" {
					t.Fatalf("%s: special-purpose target completed: %s", name, raw)
				}
			}
		}
	}
	if contacts() != 0 {
		t.Fatalf("%s: special-purpose target contacted", name)
	}
}

func tcpStart(t *testing.T) (*harness, func() int32) {
	h, d := tcpProductionHarness(t)
	return h, d.calls.Load
}

func tcp01SafetyFixture(*testing.T) safetyFixture {
	return safetyFixture{test: tcp01, baseline: task.LowRisk, factors: lowFactors(), start: tcpStart, task: tcpTask, traffic: true}
}

// M-TCP-01-SAFETY.
// requirement-ref: REQ-TCP-01-0031
// requirement-ref: REQ-TCP-01-0032
// requirement-ref: REQ-TCP-01-0033
// requirement-ref: REQ-TCP-01-0034
// requirement-ref: REQ-TCP-01-0035
// requirement-ref: REQ-TCP-01-0036
// requirement-ref: REQ-TCP-01-0037
func TestMTcp01Safety(t *testing.T) {
	runSafety(t, withActions(tcp01SafetyFixture(t), safetyFixture{
		actions: []safetyAction{
			{"data minimization", TestMTcp01Minimization},
			{"traffic distinctiveness", func(t *testing.T) {
				h, d := tcpProductionHarness(t)
				c := h.initialized(t)
				defer c.Close()
				tcpResult(t, tcpRaw(t, h, c, tcpTask("one-connect")), "completed", "", "connection_refused")
				if d.calls.Load() != 1 {
					t.Fatalf("%d connects in one run", d.calls.Load())
				}
				for _, input := range []string{`{"ip_address":"8.8.8.8","port":[443,444]}`, `{"ip_address":"8.8.8.8","port":"443-444"}`, `{"ip_address":"8.8.8.8","ports":[443]}`} {
					data := strings.Replace(tcpTask("ports"), `{"ip_address":"8.8.8.8","port":443}`, input, 1)
					if _, code := h.tryAdmit(t, c, data); code != "task_schema_invalid" || d.calls.Load() != 1 {
						t.Fatalf("port sequence %s: %s", input, code)
					}
				}
				t.Run("equivalent runs", TestMTcp01EquivalentRunsSerialize)
				t.Run("tuple wait", TestMTcp01TupleConcurrency)
			}},
			{"user-identification risk", func(t *testing.T) {
				h, d := tcpProductionHarness(t)
				var wrote, read int
				d.dial = func(context.Context) (measurement.Conn, error) {
					return &safetyConn{tcpClosingConn{close: func() {}}, &wrote, &read}, nil
				}
				c := h.initialized(t)
				defer c.Close()
				tcpResult(t, tcpRaw(t, h, c, tcpTask("no-bytes")), "completed", "", "connected")
				if wrote != 0 || read != 0 {
					t.Fatalf("application bytes: wrote %d, read %d", wrote, read)
				}
			}},
			{"target risk", func(t *testing.T) {
				for _, addr := range []string{"127.0.0.1", "0.0.0.0", "224.0.0.1", "169.254.1.1", "10.0.0.1", "192.168.1.1", "192.0.2.1", "198.18.0.1", "::1", "fe80::1", "2001:db8::1"} {
					safetyTargetRefused(t, tcpStart, addr, task.TargetTuple{Address: netip.MustParseAddr(addr), Port: 443, Scheme: "tcp"},
						strings.Replace(tcpTask("special"), `{"ip_address":"8.8.8.8","port":443}`, `{"ip_address":"`+addr+`","port":443}`, 1))
				}
				h, d := tcpProductionHarness(t)
				c := h.initialized(t)
				defer c.Close()
				for _, input := range []string{`{"ip_address":"8.8.4.4","port":443}`, `{"ip_address":"8.8.8.8","port":444}`, `{"ip_address":"example.org","port":443}`} {
					if _, code := h.tryAdmit(t, c, strings.Replace(tcpTask("unreviewed"), `{"ip_address":"8.8.8.8","port":443}`, input, 1)); code == "ok" || d.calls.Load() != 0 {
						t.Fatalf("unreviewed target %s admitted: %s", input, code)
					}
				}
			}},
			{"repeated-execution risk", func(t *testing.T) {
				h, d := tcpProductionHarness(t)
				c := h.initialized(t)
				defer c.Close()
				th := h.admit(t, c, tcpTask("once"))
				if code := startCode(t, c, th); code != "ok" {
					t.Fatal(code)
				}
				h.advance(48 * time.Hour)
				if d.calls.Load() != 1 || len(h.run.output.Records()) != 1 {
					t.Fatal("the test layer repeated a run")
				}
			}},
		}}))
}

// clientCertListener is a loopback TLS endpoint that requests, but does not
// require, a client certificate and records how many the client presented.
func clientCertListener(t *testing.T, cert tls.Certificate, alpn []string) (string, func() (int, int)) {
	t.Helper()
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}, ClientAuth: tls.RequestClientCert, NextProtos: alpn})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	var mu sync.Mutex
	handshakes, presented := 0, 0
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer c.Close()
				tc := c.(*tls.Conn)
				_ = tc.SetDeadline(time.Now().Add(5 * time.Second))
				if tc.Handshake() == nil {
					mu.Lock()
					handshakes++
					presented += len(tc.ConnectionState().PeerCertificates)
					mu.Unlock()
				}
			}()
		}
	}()
	return ln.Addr().String(), func() (int, int) {
		mu.Lock()
		defer mu.Unlock()
		return handshakes, presented
	}
}

func waitHandshakes(t *testing.T, counts func() (int, int), want int) int {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(time.Millisecond) {
		if n, presented := counts(); n >= want {
			return presented
		}
	}
	t.Fatal("no handshake")
	return 0
}

// identifyingHeaders must never appear on a measurement request.
var identifyingHeaders = []string{"Cookie", "Authorization", "Proxy-Authorization", "If-None-Match", "If-Modified-Since", "Referer", "Origin", "X-Requested-With", "X-Client-Id", "X-Device-Id"}

func plainRequest(t *testing.T, r loopback.HTTPRequest, method string) {
	t.Helper()
	if r.Method != method {
		t.Fatalf("method %s", r.Method)
	}
	for _, name := range identifyingHeaders {
		if r.Header.Get(name) != "" {
			t.Fatalf("request carries %s", name)
		}
	}
	for name := range r.Header {
		for _, v := range r.Header[name] {
			if strings.Contains(v, "safety-") || strings.Contains(strings.ToLower(v), "barometer") {
				t.Fatalf("request header %s carries a task or library identifier", name)
			}
		}
	}
}

// decodedParams decodes input with the production codec of test; false
// means the codec itself refuses the input before any admission step.
func decodedParams(t *testing.T, test task.TestVersion, input string) (task.Parameters, bool) {
	t.Helper()
	reg, err := productionRegistry(nil, true)
	if err != nil {
		t.Fatal(err)
	}
	e, _ := reg.Lookup(test)
	d := e.Codec.Definition()
	l := jsondecode.Limits{MaxBytes: 65536, MaxDepth: 32, MaxNodes: 2000}
	i, err := jsondecode.Decode([]byte(input), d.Input, l)
	if err != nil {
		return nil, false
	}
	o, err := jsondecode.Decode([]byte(`{}`), d.Options, l)
	if err != nil {
		t.Fatal(err)
	}
	p, err := e.Codec.Normalize(i, o)
	return p, err == nil
}

// withActions adds a family's Safety Analysis action rows to its fixture.
func withActions(f, a safetyFixture) safetyFixture {
	f.actions = a.actions
	return f
}

func tls01SafetyFixture(t *testing.T) safetyFixture {
	f := newTLSFixture(t)
	start := func(t *testing.T) (*harness, func() int32) {
		h, d := tlsProductionHarness(t, f)
		return h, d.calls.Load
	}
	return safetyFixture{test: tls01, baseline: task.LowRisk, factors: lowFactors(), start: start, task: tlsTask, traffic: true, tls: f}
}

func dns01SafetyFixture(*testing.T) safetyFixture {
	start := func(t *testing.T) (*harness, func() int32) {
		h, d := dnsHarness(t, "udp")
		return h, d.calls.Load
	}
	return safetyFixture{test: dnsTest, baseline: task.LowRisk, factors: lowFactors(), start: start, task: func(id string) string { return dnsTask(id, "udp") }, traffic: true}
}

func dns03SafetyFixture(t *testing.T) safetyFixture {
	f := newDohFixture(t)
	start := func(t *testing.T) (*harness, func() int32) {
		h, d := dohHarness(t, f, dohInput)
		return h, func() int32 { return d.calls.Load() + d.lookups.Load() }
	}
	return safetyFixture{test: dohTest, baseline: task.MediumRisk, factors: mediumFactors("data_sensitivity"), start: start,
		task: func(id string) string { return dohTask(id, dohInput, `{}`) }, traffic: true, doh: f}
}

func dns04SafetyFixture(t *testing.T) safetyFixture {
	f := newDotFixture(t)
	start := func(t *testing.T) (*harness, func() int32) {
		h, d := dotHarness(t, f, dotInput)
		return h, func() int32 { return d.calls.Load() + d.lookups.Load() }
	}
	return safetyFixture{test: dotTest, baseline: task.MediumRisk, factors: mediumFactors("data_sensitivity"), start: start,
		task: func(id string) string { return dotTask(id, dotInput, `{}`) }, traffic: true, dot: f}
}

func http01SafetyFixture(*testing.T) safetyFixture {
	start := func(t *testing.T) (*harness, func() int32) {
		h, d, _ := httpHarness(t, httpInput, responseHTTP("ok"), false)
		return h, d.calls.Load
	}
	return safetyFixture{test: httpTest, baseline: task.LowRisk, factors: lowFactors(), start: start,
		task: func(id string) string { return httpTask(id, httpInput, `{}`) }, traffic: true}
}

func net01SafetyFixture(*testing.T) safetyFixture {
	start := func(t *testing.T) (*harness, func() int32) {
		h, _, d := net01Harness(t)
		return h, d.calls.Load
	}
	return safetyFixture{test: net01Test, baseline: task.LowRisk, factors: lowFactors(), start: start, task: net01Task}
}

func net03SafetyFixture(*testing.T) safetyFixture {
	start := func(t *testing.T) (*harness, func() int32) {
		h, d, _, _, other := net03Harness(t, net03Input, responseHTTP("ok"), "unknown", false, false)
		return h, func() int32 { return d.calls.Load() + other.Load() }
	}
	return safetyFixture{test: net03Test, baseline: task.LowRisk, factors: lowFactors(), start: start,
		task: func(id string) string { return net03Task(id, net03Input, `{}`) }, traffic: true}
}

// net05Served is a validated NET-05 harness whose system resolver and
// transport both count contacts.
type net05Served struct {
	h       *harness
	d       *httpTransport
	srv     *loopback.HTTPServer
	lookups *atomic.Int32
}

func net05Serve(t *testing.T) net05Served {
	h, d, srv := net05HTTPHarness(t, responseHTTP("ok"), false)
	lookups := &atomic.Int32{}
	h.adapters.run.resolver = dnsSystemResolver{func(context.Context, string, string) ([]netip.Addr, error) {
		lookups.Add(1)
		return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
	}}
	return net05Served{h, d, srv, lookups}
}

// net05SafetyTask is the four-step profile with a system-resolver DNS step.
func net05SafetyTask(id string) string {
	steps := primitiveSteps()
	steps[0] = profileStep("dns", "DNS-01", `{"hostname":"example.com","record_type":"A"}`, `{}`)
	return net05Task(id, profileInput(steps...), `{}`, "")
}

func net05SafetyFixture(*testing.T) safetyFixture {
	start := func(t *testing.T) (*harness, func() int32) {
		s := net05Serve(t)
		return s.h, func() int32 { return s.d.calls.Load() + s.lookups.Load() }
	}
	return safetyFixture{test: net05Test, baseline: task.MediumRisk, factors: mediumFactors("target_sensitivity"), primitives: []task.TestVersion{dnsTest, tcp01, tls01, httpTest},
		start: start, task: net05SafetyTask, traffic: true}
}

// M-TLS-01-SAFETY.
// requirement-ref: REQ-TLS-01-0032
// requirement-ref: REQ-TLS-01-0033
// requirement-ref: REQ-TLS-01-0034
// requirement-ref: REQ-TLS-01-0035
// requirement-ref: REQ-TLS-01-0036
// requirement-ref: REQ-TLS-01-0037
// requirement-ref: REQ-TLS-01-0038
func TestMTls01Safety(t *testing.T) {
	fx := tls01SafetyFixture(t)
	f, start := fx.tls, fx.start
	runSafety(t, withActions(fx, safetyFixture{
		actions: []safetyAction{
			{"data minimization", TestMTls01Minimization},
			{"traffic distinctiveness", func(t *testing.T) {
				t.Run("one connection and one ClientHello", TestMTls01ExecutionAndOutput)
				t.Run("no implicit DNS", TestMTls01NoImplicitDNS)
				h, d := tlsProductionHarness(t, f)
				c := h.initialized(t)
				defer c.Close()
				tcpRaw(t, h, c, tlsTask("no-retry"))
				if d.calls.Load() != 1 {
					t.Fatalf("%d connections after a refused connect", d.calls.Load())
				}
			}},
			{"user-identification risk", func(t *testing.T) {
				h, d := tlsProductionHarness(t, f)
				addr, counts := clientCertListener(t, f.cert, []string{"h2"})
				d.dial = func(ctx context.Context) (measurement.Conn, error) {
					var nd net.Dialer
					return nd.DialContext(ctx, "tcp", addr)
				}
				c := h.initialized(t)
				defer c.Close()
				tcpRaw(t, h, c, tlsTask("no-client-cert"))
				if waitHandshakes(t, counts, 1) != 0 {
					t.Fatal("a client certificate was presented")
				}
				t.Run("no session resumption", TestMTls01NoSessionResumption)
			}},
			{"target risk", func(t *testing.T) {
				h, d := tlsProductionHarness(t, f)
				c := h.initialized(t)
				defer c.Close()
				for _, input := range []string{strings.Replace(tlsInput, tlsServerName, "other.example", 1), strings.Replace(tlsInput, "443", "8443", 1), strings.Replace(tlsInput, "8.8.8.8", "8.8.4.4", 1)} {
					data := strings.Replace(tlsTask("unreviewed"), tlsInput, input, 1)
					if _, code := h.tryAdmit(t, c, data); code == "ok" || d.calls.Load() != 0 {
						t.Fatalf("unreviewed tuple %s: %s", input, code)
					}
				}
				for _, addr := range []string{"127.0.0.1", "10.0.0.1", "169.254.1.1", "192.0.2.1", "::1"} {
					safetyTargetRefused(t, start, addr, task.TargetTuple{Address: netip.MustParseAddr(addr), Port: 443, Scheme: "tls", TLSServerName: tlsServerName, ALPN: []string{"h2", "http/1.1"}},
						strings.Replace(tlsTask("special"), `"8.8.8.8"`, `"`+addr+`"`, 1))
				}
			}},
			{"repeated-execution risk", TestMTls01TupleConcurrency},
		}}))
}

// M-DNS-01-SAFETY.
// requirement-ref: REQ-DNS-01-0045
// requirement-ref: REQ-DNS-01-0046
// requirement-ref: REQ-DNS-01-0047
// requirement-ref: REQ-DNS-01-0048
// requirement-ref: REQ-DNS-01-0049
// requirement-ref: REQ-DNS-01-0050
// requirement-ref: REQ-DNS-01-0051
// requirement-ref: REQ-DNS-01-0052
func TestMDns01Safety(t *testing.T) {
	fx := dns01SafetyFixture(t)
	start := fx.start
	systemRun := func(t *testing.T, id string) (*harness, *atomic.Int32, *[]string) {
		h, _ := dnsHarness(t, "system")
		var lookups atomic.Int32
		names := &[]string{}
		var mu sync.Mutex
		h.adapters.run.resolver = dnsSystemResolver{func(_ context.Context, family, name string) ([]netip.Addr, error) {
			lookups.Add(1)
			mu.Lock()
			*names = append(*names, family+" "+name)
			mu.Unlock()
			return []netip.Addr{netip.MustParseAddr("9.9.9.9")}, nil
		}}
		c := h.initialized(t)
		defer c.Close()
		tcpRaw(t, h, c, dnsTask(id, "system"))
		return h, &lookups, names
	}
	runSafety(t, withActions(fx, safetyFixture{
		actions: []safetyAction{
			{"data minimization", TestMDns01Minimization},
			{"traffic distinctiveness", func(t *testing.T) {
				// One ordinary query per run, through either engine, and no
				// test-layer retry after a failure.
				_, lookups, names := systemRun(t, "one-query")
				if lookups.Load() != 1 || (*names)[0] != "ip4 example.org." {
					t.Fatalf("system lookups %v", *names)
				}
				h, d := dnsHarness(t, "udp")
				c := h.initialized(t)
				defer c.Close()
				tcpRaw(t, h, c, dnsTask("no-retry", "udp"))
				if d.calls.Load() != 1 {
					t.Fatalf("%d resolver contacts after a failure", d.calls.Load())
				}
				for _, typ := range []string{"TXT", "ANY", "MX", "PTR"} {
					data := strings.Replace(dnsTask("type", "udp"), `"record_type":"A"`, `"record_type":"`+typ+`"`, 1)
					if _, code := h.tryAdmit(t, c, data); code == "ok" || d.calls.Load() != 1 {
						t.Fatalf("record type %s: %s", typ, code)
					}
				}
			}},
			{"user-identification risk", func(t *testing.T) {
				// The library holds no DNS cache and reads none: the only name
				// that reaches the resolver API is the admitted query name.
				_, _, names := systemRun(t, "no-cache")
				if len(*names) != 1 || strings.Contains(strings.Join(*names, " "), "safety") {
					t.Fatalf("resolver saw %v", *names)
				}
				t.Run("closed result", TestMDns01Minimization)
			}},
			{"target risk", func(t *testing.T) {
				h, d := dnsHarness(t, "udp")
				c := h.initialized(t)
				defer c.Close()
				for _, swap := range [][2]string{{"example.org", "unreviewed.example"}, {"8.8.8.8", "8.8.4.4"}, {"8.8.8.8", "10.0.0.1"}, {"8.8.8.8", "127.0.0.1"}} {
					data := strings.Replace(dnsTask("unreviewed", "udp"), swap[0], swap[1], 1)
					if _, code := h.tryAdmit(t, c, data); code == "ok" || d.calls.Load() != 0 {
						t.Fatalf("unreviewed %s: %s", swap[1], code)
					}
				}
				for _, addr := range []string{"127.0.0.1", "10.0.0.1", "169.254.1.1", "::1"} {
					in := task.TargetTuple{Address: netip.MustParseAddr(addr), Port: 53, Scheme: "udp"}
					h, contacts := start(t)
					p := policyInputOf(t, h.adapters.policy)
					p.Targets[1].Tuple = in
					policy, err := task.NewLocalPolicy(p)
					if err != nil {
						continue
					}
					h.adapters.policy = policy
					c := h.initialized(t)
					if th, code := h.tryAdmit(t, c, strings.Replace(dnsTask("special", "udp"), "8.8.8.8", addr, 1)); code == "ok" {
						startCode(t, c, th)
					}
					_ = c.Close()
					if contacts() != 0 {
						t.Fatalf("special-purpose resolver %s contacted", addr)
					}
				}
			}},
			{"repeated-execution risk", func(t *testing.T) {
				t.Run("equivalent runs", TestDns01EquivalentRunsSerialize)
				h, d := dnsHarness(t, "udp")
				c := h.initialized(t)
				defer c.Close()
				tcpRaw(t, h, c, dnsTask("once", "udp"))
				h.advance(48 * time.Hour)
				if d.calls.Load() != 1 || len(h.run.output.Records()) != 1 {
					t.Fatal("the test layer repeated a run")
				}
			}},
		}}))
}

// encryptedDNSRequest asserts the DoH request is one ordinary RFC 8484 POST.
func encryptedDNSRequest(t *testing.T, r loopback.HTTPRequest) {
	t.Helper()
	plainRequest(t, r, "POST")
	if r.Header.Get("Content-Type") != "application/dns-message" || r.Target != "/dns-query" || len(r.Body) != dohQueryBytes {
		t.Fatalf("not an ordinary DoH POST: %s %v %d", r.Target, r.Header, len(r.Body))
	}
}

// M-DNS-03-SAFETY.
// requirement-ref: REQ-DNS-03-0033
// requirement-ref: REQ-DNS-03-0034
// requirement-ref: REQ-DNS-03-0035
// requirement-ref: REQ-DNS-03-0036
// requirement-ref: REQ-DNS-03-0037
// requirement-ref: REQ-DNS-03-0038
// requirement-ref: REQ-DNS-03-0039
func TestMDns03Safety(t *testing.T) {
	fx := dns03SafetyFixture(t)
	f, start, task03 := fx.doh, fx.start, fx.task
	served := func(t *testing.T) (*harness, *dohTransport, *loopback.HTTPServer, *Client) {
		h, d := dohHarness(t, f, dohInput)
		srv := dohServe(t, d, f.reply(dohReply(0, loopback.DNSRecord{Type: "A", Value: "9.9.9.9"})))
		c := h.initialized(t)
		t.Cleanup(func() { _ = c.Close() })
		return h, d, srv, c
	}
	runSafety(t, withActions(fx, safetyFixture{
		actions: []safetyAction{
			{"data minimization", TestMDns03Minimization},
			{"traffic distinctiveness", func(t *testing.T) {
				h, d, srv, c := served(t)
				tcpRaw(t, h, c, task03("one-post"))
				reqs := srv.Counts().Requests
				if len(reqs) != 1 || d.calls.Load() != 1 {
					t.Fatalf("%d requests, %d connections", len(reqs), d.calls.Load())
				}
				encryptedDNSRequest(t, reqs[0])
				t.Run("one connection and request, no retry", TestMDns03Limits)
			}},
			{"user-identification risk", func(t *testing.T) {
				h, _, srv, c := served(t)
				tcpRaw(t, h, c, task03("no-identifiers"))
				for _, r := range srv.Counts().Requests {
					encryptedDNSRequest(t, r)
				}
				h, d := dohHarness(t, f, dohInput)
				addr, counts := clientCertListener(t, f.cert.Certificate, []string{"h2", "http/1.1"})
				dohRoute(d, addr)
				c = h.initialized(t)
				defer c.Close()
				tcpRaw(t, h, c, task03("no-client-cert"))
				if waitHandshakes(t, counts, 1) != 0 {
					t.Fatal("a client certificate was presented")
				}
			}},
			{"target risk", func(t *testing.T) {
				h, d := dohHarness(t, f, dohInput)
				c := h.initialized(t)
				defer c.Close()
				for _, swap := range [][2]string{{"query.example", "unreviewed.example"}, {"doh.resolver.example/dns-query", "other.resolver.example/dns-query"}, {"/dns-query", "/other"}, {`"bootstrap_ip":"8.8.8.8"`, `"bootstrap_ip":"8.8.4.4"`}, {`"bootstrap_ip":"8.8.8.8"`, `"bootstrap_ip":"127.0.0.1"`}} {
					data := strings.Replace(task03("unreviewed"), swap[0], swap[1], 1)
					if _, code := h.tryAdmit(t, c, data); code == "ok" || d.calls.Load()+d.lookups.Load() != 0 {
						t.Fatalf("unreviewed %s: %s", swap[1], code)
					}
				}
				for _, addr := range []string{"127.0.0.1", "10.0.0.1", "169.254.1.1", "::1"} {
					input := strings.Replace(dohInput, `"8.8.8.8"`, `"`+addr+`"`, 1)
					h, contacts := start(t)
					p := policyInputOf(t, h.adapters.policy)
					params, ok := decodedParams(t, dohTest, input)
					if !ok {
						continue // the codec refuses the bootstrap address
					}
					targets, _ := dns03Targets{}.Targets(params)
					p.Targets[1].Tuple = targets[1].Tuple
					policy, err := task.NewLocalPolicy(p)
					if err != nil {
						continue
					}
					h.adapters.policy = policy
					c := h.initialized(t)
					if th, code := h.tryAdmit(t, c, dohTask("special", input, `{}`)); code == "ok" {
						startCode(t, c, th)
					}
					_ = c.Close()
					if contacts() != 0 {
						t.Fatalf("special-purpose bootstrap %s contacted", addr)
					}
				}
			}},
			{"repeated-execution risk", TestMDns03TupleConcurrency},
		}}))
}

// M-DNS-04-SAFETY.
// requirement-ref: REQ-DNS-04-0033
// requirement-ref: REQ-DNS-04-0034
// requirement-ref: REQ-DNS-04-0035
// requirement-ref: REQ-DNS-04-0036
// requirement-ref: REQ-DNS-04-0037
// requirement-ref: REQ-DNS-04-0038
// requirement-ref: REQ-DNS-04-0039
func TestMDns04Safety(t *testing.T) {
	fx := dns04SafetyFixture(t)
	f, start, task04 := fx.dot, fx.start, fx.task
	runSafety(t, withActions(fx, safetyFixture{
		actions: []safetyAction{
			{"data minimization", TestMDns04Minimization},
			{"traffic distinctiveness", func(t *testing.T) {
				h, d := dotHarness(t, f, dotInput)
				cfg := f.cert
				cfg.DoT = dotAnswer
				srv := dotServe(t, d, cfg)
				c := h.initialized(t)
				defer c.Close()
				tcpRaw(t, h, c, task04("one-exchange"))
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := srv.WaitIdle(ctx); err != nil {
					t.Fatal(err)
				}
				// One connection, one standard handshake on the registered
				// port, no ALPN or name variation and no fallback.
				if counts := srv.Counts(); d.calls.Load() != 1 || counts.Accepted != 1 || len(counts.Hellos) != 1 || counts.Hellos[0].ServerName != dotName {
					t.Fatalf("%d connections, server %+v", d.calls.Load(), counts)
				}
				t.Run("one connection, no retry", TestMDns04Limits)
			}},
			{"user-identification risk", func(t *testing.T) {
				h, d := dotHarness(t, f, dotInput)
				addr, counts := clientCertListener(t, f.cert.Certificate, nil)
				d.dial = func(ctx context.Context) (measurement.Conn, error) {
					var nd net.Dialer
					return nd.DialContext(ctx, "tcp", addr)
				}
				c := h.initialized(t)
				defer c.Close()
				tcpRaw(t, h, c, task04("no-client-cert"))
				if waitHandshakes(t, counts, 1) != 0 {
					t.Fatal("a client certificate was presented")
				}
				t.Run("closed result", TestMDns04Minimization)
			}},
			{"target risk", func(t *testing.T) {
				h, d := dotHarness(t, f, dotInput)
				c := h.initialized(t)
				defer c.Close()
				for _, swap := range [][2]string{{"query.example", "unreviewed.example"}, {"dot.resolver.example", "other.resolver.example"}, {`"port":853`, `"port":8853`}, {`"bootstrap_ip":"8.8.8.8"`, `"bootstrap_ip":"8.8.4.4"`}, {`"bootstrap_ip":"8.8.8.8"`, `"bootstrap_ip":"127.0.0.1"`}} {
					data := strings.Replace(task04("unreviewed"), swap[0], swap[1], 1)
					if _, code := h.tryAdmit(t, c, data); code == "ok" || d.calls.Load()+d.lookups.Load() != 0 {
						t.Fatalf("unreviewed %s: %s", swap[1], code)
					}
				}
				for _, addr := range []string{"127.0.0.1", "10.0.0.1", "169.254.1.1", "192.0.2.1", "::1"} {
					input := strings.Replace(dotInput, `"8.8.8.8"`, `"`+addr+`"`, 1)
					h, contacts := start(t)
					p := policyInputOf(t, h.adapters.policy)
					params, ok := decodedParams(t, dotTest, input)
					if !ok {
						continue // the codec refuses the bootstrap address
					}
					targets, _ := dns04Targets{}.Targets(params)
					p.Targets[1].Tuple = targets[1].Tuple
					policy, err := task.NewLocalPolicy(p)
					if err != nil {
						continue
					}
					h.adapters.policy = policy
					c := h.initialized(t)
					if th, code := h.tryAdmit(t, c, dotTask("special", input, `{}`)); code == "ok" {
						startCode(t, c, th)
					}
					_ = c.Close()
					if contacts() != 0 {
						t.Fatalf("special-purpose bootstrap %s contacted", addr)
					}
				}
			}},
			{"repeated-execution risk", TestMDns04TupleConcurrency},
		}}))
}

// M-HTTP-01-SAFETY.
// requirement-ref: REQ-HTTP-01-0044
// requirement-ref: REQ-HTTP-01-0045
// requirement-ref: REQ-HTTP-01-0046
// requirement-ref: REQ-HTTP-01-0047
// requirement-ref: REQ-HTTP-01-0048
// requirement-ref: REQ-HTTP-01-0049
// requirement-ref: REQ-HTTP-01-0050
func TestMHttp01Safety(t *testing.T) {
	fx := http01SafetyFixture(t)
	start, task01 := fx.start, fx.task
	served := func(t *testing.T, raw string) (*harness, *httpTransport, *loopback.HTTPServer, *Client) {
		h, d, srv := httpHarness(t, httpInput, raw, false)
		c := h.initialized(t)
		t.Cleanup(func() { _ = c.Close() })
		return h, d, srv, c
	}
	runSafety(t, withActions(fx, safetyFixture{
		actions: []safetyAction{
			{"network_observability", func(t *testing.T) {
				// Only the registered origin over authenticated TLS: a
				// plaintext URL is never admitted.
				h, d, srv, c := served(t, responseHTTP("ok"))
				for _, input := range []string{strings.Replace(httpInput, "https://", "http://", 1), strings.Replace(httpInput, "https://8.8.8.8/", "https://8.8.8.8:80/", 1)} {
					if _, code := h.tryAdmit(t, c, httpTask("plain", input, `{}`)); code == "ok" || d.calls.Load() != 0 {
						t.Fatalf("%s: %s", input, code)
					}
				}
				tcpRaw(t, h, c, task01("tls-only"))
				if counts := srv.Counts(); counts.Accepted != 1 || len(counts.Hellos) != 1 {
					t.Fatalf("not one authenticated TLS connection: %+v", counts)
				}
			}},
			{"traffic_distinctiveness", func(t *testing.T) {
				h, d, srv, c := served(t, responseHTTP("ok"))
				tcpRaw(t, h, c, task01("one-get"))
				reqs := srv.Counts().Requests
				if len(reqs) != 1 || d.calls.Load() != 1 || len(reqs[0].Body) != 0 || reqs[0].Target != "/check" {
					t.Fatalf("not one bodyless GET: %d requests", len(reqs))
				}
				plainRequest(t, reqs[0], "GET")
				if enc := reqs[0].Header.Get("Accept-Encoding"); enc != "" && enc != "identity" {
					t.Fatalf("content encoding %q", enc)
				}
				t.Run("equivalent runs", TestMHttp01EquivalentURLsSerialize)
				t.Run("equivalent spellings", TestMHttp01EquivalentSpellingSerialized)
			}},
			{"target_sensitivity", func(t *testing.T) {
				h, d, _, c := served(t, responseHTTP("ok"))
				for _, input := range []string{strings.Replace(httpInput, "8.8.8.8", "8.8.4.4", 1), strings.Replace(httpInput, "/check", "/other", 1), strings.Replace(httpInput, "8.8.8.8", "127.0.0.1", 1), strings.Replace(httpInput, "8.8.8.8", "10.0.0.1", 1), strings.Replace(httpInput, "8.8.8.8", "localhost", 1)} {
					if _, code := h.tryAdmit(t, c, httpTask("unregistered", input, `{}`)); code == "ok" || d.calls.Load() != 0 {
						t.Fatalf("%s: %s", input, code)
					}
				}
				for _, addr := range []string{"127.0.0.1", "10.0.0.1", "169.254.1.1"} {
					input := strings.Replace(httpInput, "8.8.8.8", addr, 1)
					params, ok := decodedParams(t, httpTest, input)
					if !ok {
						continue // the codec refuses the address
					}
					safetyTargetRefused(t, start, addr, params.(http01.Parameters).Target(), httpTask("special", input, `{}`))
				}
			}},
			{"data_sensitivity", TestMHttp01Limits},
			{"data minimization", TestMHttp01Minimization},
			{"user-identification risk", func(t *testing.T) {
				// A response that sets a cookie is not followed by a request
				// that returns it, and no credential, cache validator or
				// application identifier is attached.
				raw := "HTTP/1.1 200 OK\r\nSet-Cookie: session=cookie-canary\r\nContent-Length: 2\r\n\r\nok"
				h, _, srv, c := served(t, raw)
				tcpRaw(t, h, c, task01("first"))
				tcpRaw(t, h, c, task01("second"))
				reqs := srv.Counts().Requests
				if len(reqs) != 2 {
					t.Fatalf("%d requests", len(reqs))
				}
				for _, r := range reqs {
					plainRequest(t, r, "GET")
				}
				h, d, _ := httpHarness(t, httpInput, "", false)
				ca, err := loopback.NewTestCA()
				if err != nil {
					t.Fatal(err)
				}
				cert, err := ca.Issue("8.8.8.8", false)
				if err != nil {
					t.Fatal(err)
				}
				h.adapters.measurementRoots = ca.Pool()
				addr, counts := clientCertListener(t, cert, nil)
				d.dial = func(ctx context.Context, _, _ string) (measurement.Conn, error) {
					var nd net.Dialer
					return nd.DialContext(ctx, "tcp", addr)
				}
				c = h.initialized(t)
				defer c.Close()
				tcpRaw(t, h, c, task01("no-client-cert"))
				if waitHandshakes(t, counts, 1) != 0 {
					t.Fatal("a client certificate was presented")
				}
			}},
			{"target risk", TestMHttp01Redirects},
		}}))
}

// M-NET-01-SAFETY.
// requirement-ref: REQ-NET-01-0032
// requirement-ref: REQ-NET-01-0033
// requirement-ref: REQ-NET-01-0034
// requirement-ref: REQ-NET-01-0035
// requirement-ref: REQ-NET-01-0036
// requirement-ref: REQ-NET-01-0037
// requirement-ref: REQ-NET-01-0038
func TestMNet01Safety(t *testing.T) {
	runSafety(t, withActions(net01SafetyFixture(t), safetyFixture{
		actions: []safetyAction{
			{"target_sensitivity", func(t *testing.T) {
				h, _, d := net01Harness(t)
				c := h.initialized(t)
				defer c.Close()
				for _, input := range []string{`{"target":"example.org"}`, `{"ip_address":"8.8.8.8"}`, `{"url":"https://8.8.8.8/"}`} {
					if _, code := h.tryAdmit(t, c, strings.Replace(net01Task("target"), `"input":{}`, `"input":`+input, 1)); code != "task_schema_invalid" || d.calls.Load() != 0 {
						t.Fatalf("remote target %s: %s", input, code)
					}
				}
			}},
			{"data minimization and data_sensitivity", TestMNet01Minimization},
			{"repeated-execution risk", func(t *testing.T) {
				h, _, d := net01Harness(t)
				c := h.initialized(t)
				defer c.Close()
				for i := range 5 {
					tcpRaw(t, h, c, net01Task(fmt.Sprint("repeat-", i)))
				}
				if d.calls.Load() != 0 || len(h.run.output.Records()) != 5 {
					t.Fatal("repeated snapshots used the network")
				}
			}},
			{"residual risk", TestMNet01ExecutionAndOutput},
		}}))
}

// M-NET-03-SAFETY.
// requirement-ref: REQ-NET-03-0034
// requirement-ref: REQ-NET-03-0035
// requirement-ref: REQ-NET-03-0036
// requirement-ref: REQ-NET-03-0037
// requirement-ref: REQ-NET-03-0038
// requirement-ref: REQ-NET-03-0039
// requirement-ref: REQ-NET-03-0040
func TestMNet03Safety(t *testing.T) {
	fx := net03SafetyFixture(t)
	task03 := fx.task
	runSafety(t, withActions(fx, safetyFixture{
		actions: []safetyAction{
			{"data minimization", TestMNet03Minimization},
			{"traffic distinctiveness and user identification", func(t *testing.T) {
				for _, state := range []string{"validated", "not_validated", "unknown"} {
					h, d, srv, _, other := net03Harness(t, net03Input, responseHTTP("ok"), state, false, false)
					c := h.initialized(t)
					tcpRaw(t, h, c, task03("one-request"))
					_ = c.Close()
					reqs := srv.Counts().Requests
					if len(reqs) != 1 || d.calls.Load() != 1 || other.Load() != 0 {
						t.Fatalf("platform %s: %d requests", state, len(reqs))
					}
					plainRequest(t, reqs[0], "GET")
				}
				h, d, _, _, _ := net03Harness(t, net03Input, "", "unknown", false, false)
				d.dial = nil
				c := h.initialized(t)
				defer c.Close()
				tcpRaw(t, h, c, task03("no-retry"))
				if d.calls.Load() != 1 {
					t.Fatalf("%d requests after a failure", d.calls.Load())
				}
			}},
			{"target risk", func(t *testing.T) {
				h, d, _, _, other := net03Harness(t, net03Input, responseHTTP("ok"), "unknown", false, false)
				c := h.initialized(t)
				defer c.Close()
				for _, input := range []string{strings.Replace(net03Input, "8.8.8.8", "8.8.4.4", 1), strings.Replace(net03Input, "/check", "/other", 1), strings.Replace(net03Input, "200", "204", 1), strings.Replace(net03Input, "8.8.8.8", "127.0.0.1", 1)} {
					if _, code := h.tryAdmit(t, c, net03Task("unregistered", input, `{}`)); code == "ok" || d.calls.Load()+other.Load() != 0 {
						t.Fatalf("%s: %s", input, code)
					}
				}
				// A redirect target is never contacted.
				h, d, _, _, other = net03Harness(t, net03Input, "HTTP/1.1 302 Found\r\nLocation: https://portal.example/login\r\nContent-Length: 0\r\n\r\n", "unknown", false, false)
				c2 := h.initialized(t)
				defer c2.Close()
				tcpRaw(t, h, c2, task03("redirect"))
				if d.calls.Load() != 1 || other.Load() != 0 {
					t.Fatal("redirect target contacted")
				}
			}},
		}}))
}

// M-NET-05-SAFETY: NET-05 takes the most restrictive level of itself, every
// primitive it runs, the task declaration and local policy.
// requirement-ref: REQ-NET-05-0053
// requirement-ref: REQ-NET-05-0054
// requirement-ref: REQ-NET-05-0055
// requirement-ref: REQ-NET-05-0056
// requirement-ref: REQ-NET-05-0057
// requirement-ref: REQ-NET-05-0058
// requirement-ref: REQ-NET-05-0059
// requirement-ref: REQ-NET-05-0060
func TestMNet05Safety(t *testing.T) {
	fx := net05SafetyFixture(t)
	serve, start, task05, primitives := net05Serve, fx.start, fx.task, fx.primitives
	runSafety(t, withActions(fx, safetyFixture{
		actions: []safetyAction{
			{"most restrictive primitive", func(t *testing.T) {
				f := safetyFixture{start: start}
				// A primitive registration reviewed at high makes the whole
				// profile high, whatever the task declares.
				for _, level := range []task.Risk{task.HighRisk} {
					for i := range 4 {
						edit := func(in *task.PolicyInput) {
							for j := range in.Targets {
								if in.Targets[j].Test == primitives[i] {
									in.Targets[j].BaselineRisk = level
								}
							}
						}
						h, _ := start(t)
						editPolicy(t, h, edit)
						f.start = func(t *testing.T) (*harness, func() int32) {
							h, contacts := start(t)
							editPolicy(t, h, edit)
							return h, contacts
						}
						f.probeRisk(t, primitives[i].ID+" "+riskName(level), task05("primitive-risk"), level)
					}
				}
				t.Run("primitive registry baselines", TestMNet05PrimitiveRiskAndMissingCapability)
				// Composite execution needs its own enablement on both sides; a
				// policy cannot list NET-05 without it.
				disable := func(p *task.Permissions) {
					p.AllowComposite = false
					p.Tests = slices.DeleteFunc(p.Tests, func(v task.TestVersion) bool { return v == net05Test })
				}
				for _, side := range []string{"issuer", "local"} {
					code, contacts := f.admitCode(t, task05("composite-disabled"), func(in *task.PolicyInput) {
						if side == "local" {
							disable(&in.Permissions)
							return
						}
						for i := range in.Issuers {
							disable(&in.Issuers[i].Permissions)
						}
					})
					if code != "task_issuer_not_authorized" || contacts != 0 {
						t.Fatalf("%s composite disabled: %s", side, code)
					}
				}
			}},
			{"data minimization", TestMNet05Minimization},
			{"traffic distinctiveness", func(t *testing.T) {
				// Strictly sequential declared steps, no retry, no port
				// variation or target expansion; broad lists are rejected.
				t.Run("declared steps only", TestMNet05ExecutionAndOutput)
				t.Run("step count bounds", TestMNet05Admission)
				s := serve(t)
				var active, overlap atomic.Int32
				inner := s.d.dial
				s.d.dial = func(ctx context.Context, n, a string) (measurement.Conn, error) {
					if active.Add(1) > 1 {
						overlap.Add(1)
					}
					defer active.Add(-1)
					return inner(ctx, n, a)
				}
				c := s.h.initialized(t)
				defer c.Close()
				tcpRaw(t, s.h, c, task05("sequential"))
				if overlap.Load() != 0 || s.d.calls.Load() != 3 || s.lookups.Load() != 1 || len(s.srv.Counts().Requests) != 1 {
					t.Fatalf("steps overlapped or repeated: %d dials, %d lookups", s.d.calls.Load(), s.lookups.Load())
				}
			}},
			{"user-identification risk", func(t *testing.T) {
				s := serve(t)
				c := s.h.initialized(t)
				defer c.Close()
				tcpRaw(t, s.h, c, task05("no-identifiers"))
				for _, r := range s.srv.Counts().Requests {
					plainRequest(t, r, "GET")
				}
			}},
			{"target risk", TestMNet05AtomicTargetAndProxyRules},
			{"repeated-execution risk", func(t *testing.T) {
				// An overlapping second profile run waits for the first and is
				// cancelled before any contact; scheduling stays with the host.
				s := serve(t)
				entered := make(chan struct{}, 4)
				release := make(chan struct{})
				inner := s.d.dial
				s.d.dial = func(ctx context.Context, n, a string) (measurement.Conn, error) {
					entered <- struct{}{}
					select {
					case <-ctx.Done():
						return nil, ctx.Err()
					case <-release:
					}
					return inner(ctx, n, a)
				}
				c := s.h.initialized(t)
				defer c.Close()
				first, err := c.StartRun(s.h.admit(t, c, task05("overlap-first")))
				if err != nil {
					t.Fatal(codeOf(err))
				}
				select {
				case <-entered:
				case <-time.After(5 * time.Second):
					t.Fatal("first profile did not start")
				}
				calls, lookups := s.d.calls.Load(), s.lookups.Load()
				second, err := c.StartRun(s.h.admit(t, c, task05("overlap-second")))
				if err != nil {
					t.Fatal(codeOf(err))
				}
				if err := c.CancelRun(second); err != nil {
					t.Fatal(codeOf(err))
				}
				waitTerminal(t, c, second)
				if s.d.calls.Load() != calls || s.lookups.Load() != lookups {
					t.Fatal("an overlapping profile contacted a target")
				}
				close(release)
				waitTerminal(t, c, first)
			}},
		}}))
}
