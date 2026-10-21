package bundle_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/bundle"
	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/result"
)

// These capabilities must never be invoked by model construction.
type clock struct{}

func (clock) Now(context.Context) (time.Time, error) { panic("unexpected clock call") }

type keys struct{}

func (keys) DeriveHMACSHA256(context.Context, []byte) ([32]byte, error) { panic("unexpected key call") }
func (keys) Delete(context.Context) error                               { panic("unexpected delete") }

func initialization(t testing.TB, mode string, available bool) *config.Initialization {
	t.Helper()
	d, err := config.NewDecoder(config.Contract, config.Approvals{AllowLinkage: true, LinkageScopes: []string{"org.example.production"}})
	if err != nil {
		t.Fatal(err)
	}
	input := `{}`
	if mode != "" {
		input = `{"result_linkage":{"mode":"rotating_installation","scope":"org.example.production","rotation_period_seconds":86400}}`
	}
	s, err := d.Decode([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	var k platform.LinkageKeyStore
	if available {
		k = keys{}
	}
	i, err := config.Prepare(s, config.Capabilities{Clock: clock{}, LinkageKeys: k})
	if err != nil {
		t.Fatal(err)
	}
	return i
}
func builder(t testing.TB, enabled bool) *bundle.Builder {
	t.Helper()
	mode := ""
	if enabled {
		mode = "rotating_installation"
	}
	b, err := bundle.New(bundle.Contract, initialization(t, mode, true))
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func ref(id string) model.Test {
	return model.Test{ID: id, Version: result.Contract, ResultVersion: result.Contract}
}
func notRun(t testing.TB) result.Value {
	t.Helper()
	d, err := result.New(result.Binding{TaskID: "campaign", Output: result.Output{Test: ref("TCP-01")}, MaxAttempts: 1, MaxDurationMS: 1000, Implementation: model.Implementation{LibraryVersion: "0.1.0", Platform: "android"}})
	if err != nil {
		t.Fatal(err)
	}
	v, err := d.Build(model.Result{SchemaVersion: result.Contract, TaskID: "campaign", RunID: "AQEBAQEBAQEBAQEBAQEBAQ", Attempt: 1, Test: ref("TCP-01"), Implementation: model.Implementation{LibraryVersion: "0.1.0", Platform: "android"}, Time: model.Time{StartedAt: "2026-10-16T12:00:00Z"}, Execution: model.Execution{State: "not_run", Reason: "constraints_unsatisfied"}})
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func encoded(t testing.TB, v bundle.Value) []byte {
	t.Helper()
	b, err := v.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func rejected(t testing.TB, v bundle.Value, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("accepted invalid bundle")
	}
	if _, e := v.MarshalJSON(); e == nil || v.ID() != "" || len(v.View().Results) != 0 {
		t.Fatal("partial bundle escaped")
	}
}

// requirement-ref: REQ-BUNDLE-0004
// requirement-ref: REQ-BUNDLE-0017
// requirement-ref: REQ-BUNDLE-0021
func TestContractMembershipAndIdentity(t *testing.T) {
	b := builder(t, false)
	r := notRun(t)
	for _, version := range []string{"", "0.1", "0.2.0"} {
		if _, err := bundle.New(version, initialization(t, "", false)); !errors.Is(err, jsondecode.ErrVersion) {
			t.Fatal(err)
		}
	}
	if _, err := bundle.New(bundle.Contract, nil); !errors.Is(err, bundle.ErrBinding) {
		t.Fatal(err)
	}
	var zero bundle.Builder
	for _, invalid := range []*bundle.Builder{nil, &zero} {
		v, err := invalid.Build([]result.Value{r}, nil)
		rejected(t, v, err)
	}
	for _, count := range []int{0, 1, 16, 17} {
		items := make([]result.Value, count)
		for i := range items {
			items[i] = r
		}
		v, err := b.Build(items, nil)
		if count == 0 || count == 17 {
			rejected(t, v, err)
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if len(v.View().Results) != count {
			t.Fatal("membership changed")
		}
		data := encoded(t, v)
		var wire map[string]json.RawMessage
		if json.Unmarshal(data, &wire) != nil || len(wire) != 3 || string(wire["schema_version"]) != `"0.1.0"` {
			t.Fatal("unexpected envelope")
		}
		id, err := base64.RawURLEncoding.Strict().DecodeString(v.ID())
		if err != nil || len(id) != 16 || bytes.Equal(id, make([]byte, 16)) || base64.RawURLEncoding.EncodeToString(id) != v.ID() {
			t.Fatal("invalid ID")
		}
	}
	v, err := b.Build([]result.Value{r, {}}, nil)
	rejected(t, v, err)
	seen := map[string]bool{}
	for range 64 {
		v, err := b.Build([]result.Value{r}, nil)
		if err != nil || seen[v.ID()] {
			t.Fatal("identity reused", err)
		}
		seen[v.ID()] = true
	}
}

// requirement-ref: REQ-BUNDLE-0003
// requirement-ref: REQ-BUNDLE-0010
// requirement-ref: REQ-BUNDLE-0011
// requirement-ref: REQ-BUNDLE-0018
// requirement-ref: REQ-BUNDLE-0023
// requirement-ref: REQ-BUNDLE-0025
// requirement-ref: REQ-BUNDLE-0026
func TestTrustedOptionalLinkage(t *testing.T) {
	r := []result.Value{notRun(t)}
	l := model.RotatingLinkage{Scope: "org.example.production", Epoch: 0, Token: [32]byte{1, 2, 3}}
	for _, b := range []*bundle.Builder{builder(t, false), builder(t, true)} {
		v, err := b.Build(r, nil)
		if err != nil || bytes.Contains(encoded(t, v), []byte("linkage")) {
			t.Fatal("linkage present by default", err)
		}
	}
	missing, err := bundle.New(bundle.Contract, initialization(t, "rotating_installation", false))
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range []*bundle.Builder{builder(t, false), missing} {
		v, err := b.Build(r, &l)
		rejected(t, v, err)
	}
	b := builder(t, true)
	for _, epoch := range []int64{0, 2948, 9007199254740993, math.MaxInt64} {
		l.Epoch = epoch
		v, err := b.Build(r, &l)
		if err != nil {
			t.Fatal(err)
		}
		var wire struct {
			Linkage struct {
				Mode, Scope, Token string
				Epoch              int64
			}
		}
		if err := json.Unmarshal(encoded(t, v), &wire); err != nil {
			t.Fatal(err)
		}
		token, err := base64.RawURLEncoding.Strict().DecodeString(wire.Linkage.Token)
		if err != nil || !bytes.Equal(token, l.Token[:]) || wire.Linkage.Epoch != epoch || wire.Linkage.Scope != l.Scope || wire.Linkage.Mode != "rotating_installation" {
			t.Fatal("linkage projection")
		}
	}
	for _, invalid := range []model.RotatingLinkage{{Scope: l.Scope, Epoch: -1}, {Scope: "other"}, {Scope: ""}, {Scope: strings.Repeat("a", 129)}} {
		v, err := b.Build(r, &invalid)
		rejected(t, v, err)
	}
	// Tasks and collector payloads cannot import identity, enablement or mappings.
	for _, input := range []string{`{}`, `{"bundle_id":"AQEBAQEBAQEBAQEBAQEBAQ"}`, `{"linkage":{"mode":"stable","user_id":"account"}}`, `{"linkage":{"previous_token":"old","token":"new"}}`} {
		var v bundle.Value
		if json.Unmarshal([]byte(input), &v) == nil {
			t.Fatal("wire input enabled bundle")
		}
	}
	for _, v := range []any{model.Bundle{}, model.RotatingLinkage{}, bundle.Value{}} {
		if _, err := json.Marshal(v); err == nil {
			t.Fatal("unvalidated record serialized")
		}
	}
}

// requirement-ref: REQ-BUNDLE-0007
// requirement-ref: REQ-BUNDLE-0013
// requirement-ref: REQ-BUNDLE-0016
func TestImmutableRetryBytesAndViews(t *testing.T) {
	b := builder(t, true)
	l := model.RotatingLinkage{Scope: "org.example.production", Epoch: 2948, Token: [32]byte{3}}
	items := []result.Value{notRun(t)}
	v, err := b.Build(items, &l)
	if err != nil {
		t.Fatal(err)
	}
	want := encoded(t, v)
	l.Epoch++
	l.Token[0]++
	items[0] = result.Value{}
	view := v.View()
	view.BundleID = "replaced"
	view.SchemaVersion = "changed"
	view.Linkage.Scope = "changed"
	view.Linkage.Token[0]++
	view.Results[0].TaskID = "changed"
	view.Results = nil
	buffer := encoded(t, v)
	buffer[0] = '!'
	if !bytes.Equal(want, encoded(t, v)) {
		t.Fatal("mutable retry bytes")
	}
	if v.View().Results[0].Execution.State != "not_run" {
		t.Fatal("execution meaning changed")
	}
	copy := v
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 32 {
				if !bytes.Equal(want, encoded(t, copy)) {
					t.Error("concurrent retry changed")
				}
			}
		})
	}
	wg.Wait()
	for _, changed := range []*model.RotatingLinkage{&l, nil} {
		next, err := b.Build([]result.Value{notRun(t)}, changed)
		if err != nil || next.ID() == v.ID() {
			t.Fatal("changed content reused identity", err)
		}
	}
}

// requirement-ref: REQ-BUNDLE-0002
// requirement-ref: REQ-BUNDLE-0013
// requirement-ref: REQ-BUNDLE-0014
func TestMixedResultMembershipIsPreserved(t *testing.T) {
	// A negative protocol outcome remains completed, and an interrupted attempt
	// remains incomplete; bundling does not synthesize or merge their contexts.
	cd, err := observation.NewContextDecoder(observation.Contract, observation.ContextApprovals{})
	if err != nil {
		t.Fatal(err)
	}
	path, err := observation.NewPath(1, 0, "direct", "")
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := cd.Decode([]byte(`{"access":{"network_type":"wifi","vpn_state":"unknown","ipv4_available":"yes","ipv6_available":"unknown","validated":"unknown","network_changed":false},"egress":{"status":"unavailable","lookup_source":"none","network_path":"direct"}}`), path)
	if err != nil {
		t.Fatal(err)
	}
	d, err := result.New(result.Binding{TaskID: "campaign", Output: result.Output{Test: ref("TCP-01")}, MaxAttempts: 1, MaxDurationMS: 1000, Implementation: model.Implementation{LibraryVersion: "0.1.0", Platform: "android"}, ContextDecoder: cd, Path: path})
	if err != nil {
		t.Fatal(err)
	}
	base := notRun(t).View()
	base.NetworkContext = &ctx
	base.Execution = model.Execution{State: "incomplete", Reason: "cancelled_by_host"}
	incomplete, err := d.Build(base)
	if err != nil {
		t.Fatal(err)
	}
	o, err := observation.DecodeProtocol(observation.Contract, []byte(`{"type":"tcp_connect","status":"failed","failure":"timeout","duration_ms":0,"details":{"address_family":"ipv4"}}`), observation.ProtocolLimits{})
	if err != nil {
		t.Fatal(err)
	}
	base.Execution = model.Execution{State: "completed"}
	base.Measurement = &model.Measurement{Outcome: "timeout", Observations: []observation.Observation{o}}
	negative, err := d.Build(base)
	if err != nil {
		t.Fatal(err)
	}
	items := []result.Value{negative, incomplete, notRun(t)}
	v, err := builder(t, false).Build(items, nil)
	if err != nil {
		t.Fatal(err)
	}
	var wire struct{ Results []json.RawMessage }
	if err := json.Unmarshal(encoded(t, v), &wire); err != nil {
		t.Fatal(err)
	}
	for i, r := range items {
		want, err := r.MarshalJSON()
		if err != nil || !bytes.Equal(want, wire.Results[i]) {
			t.Fatal("result changed during bundling", err)
		}
	}
	view := v.View()
	view.Results[0].Measurement.Observations[0] = observation.Observation{}
	view.Results[0].Measurement.Outcome = "changed"
	if v.View().Results[0].Measurement.Outcome != "timeout" {
		t.Fatal("nested result alias")
	}
}

// The bundle has no decoder, so its output is checked with the strict shared
// parser (GO-SPEC-ISSUE-001 producer evidence, DEV-566): it is one complete
// value of valid UTF-8 with unique decoded names in every object, within the
// bundle byte bound, with and without linkage and at the largest result size.
func TestProducedBundleIsStrictJSON(t *testing.T) {
	limits := jsondecode.Limits{MaxBytes: bundle.MaxBytes, MaxDepth: 128, MaxNodes: bundle.MaxBytes}
	for _, linked := range []bool{false, true} {
		var linkage *model.RotatingLinkage
		if linked {
			linkage = &model.RotatingLinkage{Scope: "org.example.production", Epoch: 7}
		}
		for name, items := range map[string][]result.Value{
			"one result":           {notRun(t)},
			"sixteen results":      slices.Repeat([]result.Value{notRun(t)}, bundle.MaxResults),
			"large profile result": {largeResult(t, 0)},
		} {
			t.Run(fmt.Sprintf("%s linked=%v", name, linked), func(t *testing.T) {
				v, err := builder(t, linked).Build(items, linkage)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := jsondecode.Decode(encoded(t, v), &jsondecode.Schema{Kind: jsondecode.Any}, limits); err != nil {
					t.Fatal("produced bundle is not strict JSON:", err)
				}
			})
		}
	}
}

// largeResult builds an actually validated eight-step DNS profile. Increasing
// extra lengthens synthetic retained CNAMEs one byte at a time, permitting exact
// normalized bundle byte boundaries without whitespace or fabricated Values.
func largeResult(t testing.TB, extra int) result.Value {
	t.Helper()
	limits := observation.ProtocolLimits{RetainDNSAnswers: true, MaxDNSResponses: 3}
	cd, err := observation.NewContextDecoder(observation.Contract, observation.ContextApprovals{})
	if err != nil {
		t.Fatal(err)
	}
	path, err := observation.NewPath(1, 0, "direct", "")
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := cd.Decode([]byte(`{"access":{"network_type":"wifi","vpn_state":"unknown","ipv4_available":"yes","ipv6_available":"unknown","validated":"unknown","network_changed":false},"egress":{"status":"unavailable","lookup_source":"none","network_path":"direct"}}`), path)
	if err != nil {
		t.Fatal(err)
	}
	plan := &result.ProfilePlan{ID: "bounded", Version: 1}
	m := &model.Measurement{Outcome: "completed", Profile: &model.Profile{ID: "bounded", Version: 1}, NetworkValidation: &model.NetworkValidation{State: "unknown", Source: "platform"}}
	for step := range 8 {
		var responses []observation.DNSResponse
		for response := range 3 {
			var answers []observation.DNSAnswer
			for answer := range 8 {
				add := min(extra, 246)
				extra -= add
				name := strings.Repeat(string(rune('a'+answer)), 7+add)
				var labels []string
				for len(name) > 63 {
					labels = append(labels, name[:62])
					name = name[63:]
				}
				labels = append(labels, name)
				answers = append(answers, observation.DNSAnswer{Type: "CNAME", Value: strings.Join(labels, ".")})
			}
			responses = append(responses, observation.DNSResponse{ArrivalIndex: response + 1, ArrivalOffsetMS: 250, RCode: "NOERROR", Answers: answers})
		}
		o, err := observation.NewDNS(observation.Contract, "succeeded", 250, "", observation.DNSDetails{Engine: "udp", QueryType: "A", Responses: responses, ResponseLimitReached: true}, limits)
		if err != nil {
			t.Fatal(err)
		}
		id := fmt.Sprintf("step-%d", step)
		plan.Steps = append(plan.Steps, result.StepPlan{ID: id, Output: result.Output{Test: ref("DNS-01"), DNSMode: "udp", QueryType: "A", Limits: limits}})
		m.Steps = append(m.Steps, model.Step{ID: id, Test: ref("DNS-01"), State: "completed", Measurement: &model.Measurement{Outcome: "response_limit_reached", Observations: []observation.Observation{o}}})
	}
	if extra != 0 {
		t.Fatal("fixture capacity exceeded")
	}
	d, err := result.New(result.Binding{TaskID: "campaign", Output: result.Output{Test: ref("NET-05")}, MaxAttempts: 1, MaxDurationMS: 1000, Implementation: model.Implementation{LibraryVersion: "0.1.0", Platform: "android"}, Profile: plan, ContextDecoder: cd, Path: path})
	if err != nil {
		t.Fatal(err)
	}
	v, err := d.Build(model.Result{SchemaVersion: result.Contract, TaskID: "campaign", RunID: "AQEBAQEBAQEBAQEBAQEBAQ", Attempt: 1, Test: ref("NET-05"), Implementation: model.Implementation{LibraryVersion: "0.1.0", Platform: "android"}, Time: model.Time{StartedAt: "2026-10-16T12:00:00Z", DurationMS: 1000}, Execution: model.Execution{State: "completed"}, NetworkContext: &ctx, Measurement: m})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// requirement-ref: REQ-BUNDLE-0021
// requirement-ref: REQ-LIMIT-0021
// requirement-ref: REQ-LIMIT-0022
func TestExactEncodedBoundaries(t *testing.T) {
	for _, linked := range []bool{false, true} {
		b := builder(t, linked)
		var linkage *model.RotatingLinkage
		if linked {
			linkage = &model.RotatingLinkage{Scope: "org.example.production", Epoch: 2948}
		}
		base := largeResult(t, 0)
		items := []result.Value{base, base, base, base, base, base}
		v, err := b.Build(items, linkage)
		if err != nil {
			t.Fatal(err)
		}
		baseSize := len(encoded(t, v))
		for _, target := range []int{bundle.MaxBytes - 1, bundle.MaxBytes, bundle.MaxBytes + 1} {
			extra := target - baseSize
			for i := range items {
				n := min(extra, 8*3*8*246)
				extra -= n
				items[i] = largeResult(t, n)
			}
			if extra != 0 {
				t.Fatal("insufficient fixture capacity")
			}
			v, err := b.Build(items, linkage)
			if target > bundle.MaxBytes {
				rejected(t, v, err)
				if !errors.Is(err, jsondecode.ErrBounds) {
					t.Fatal(err)
				}
				continue
			}
			if err != nil {
				t.Fatal(err)
			}
			data := encoded(t, v)
			if len(data) != target || !json.Valid(data) {
				t.Fatalf("wanted %d bytes, got %d", target, len(data))
			}
		}
	}
}

func FuzzBundleProducer(f *testing.F) {
	r := notRun(f)
	b := builder(f, true)
	f.Add([]byte{1, 0, 0})
	f.Add([]byte{16, 1, 255})
	f.Add([]byte{17, 2, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 3 {
			return
		}
		count := int(data[0] % 19)
		items := make([]result.Value, count)
		for i := range items {
			items[i] = r
		}
		if data[1]&1 != 0 && count > 0 {
			items[count-1] = result.Value{}
		}
		var linkage *model.RotatingLinkage
		if data[1]&2 != 0 {
			linkage = &model.RotatingLinkage{Scope: "org.example.production", Epoch: int64(int8(data[2]))}
			copy(linkage.Token[:], data)
		}
		v, err := b.Build(items, linkage)
		if count == 0 || count > 16 || data[1]&1 != 0 || linkage != nil && linkage.Epoch < 0 {
			rejected(t, v, err)
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		wire := encoded(t, v)
		if !json.Valid(wire) || len(wire) > bundle.MaxBytes || len(v.View().Results) != count {
			t.Fatal("invalid output")
		}
		if !bytes.Equal(wire, encoded(t, v)) {
			t.Fatal("retry changed")
		}
	})
}
