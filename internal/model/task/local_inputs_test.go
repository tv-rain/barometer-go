package task_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

// Local restriction inputs are trusted policy: malformed values reject the
// whole policy atomically, valid values are copied in and out, and none of
// them can carry authority in JSON.
func TestPolicyLocalRestrictionInputs(t *testing.T) {
	ten, over, under := 10, 101, -1
	endpoints := make([]string, task.MaxContextEndpoints+1)
	for i := range endpoints {
		endpoints[i] = fmt.Sprintf("https://context%d.example.org/v1", i)
	}
	invalid := map[string]func(*task.PolicyInput){
		"battery above 100": func(p *task.PolicyInput) { p.Device.MinimumBatteryPercent = &over },
		"negative battery":  func(p *task.PolicyInput) { p.Device.MinimumBatteryPercent = &under },
		"http endpoint": func(p *task.PolicyInput) {
			p.Context = &task.ContextEligibility{Endpoints: []string{"http://context.example.org/v1"}}
		},
		"endpoint credentials": func(p *task.PolicyInput) {
			p.Context = &task.ContextEligibility{Endpoints: []string{"https://u:p@context.example.org/v1"}}
		},
		"endpoint fragment": func(p *task.PolicyInput) {
			p.Context = &task.ContextEligibility{Endpoints: []string{"https://context.example.org/v1#f"}}
		},
		"relative endpoint": func(p *task.PolicyInput) { p.Context = &task.ContextEligibility{Endpoints: []string{"/v1"}} },
		"empty endpoint":    func(p *task.PolicyInput) { p.Context = &task.ContextEligibility{Endpoints: []string{""}} },
		"duplicate endpoint": func(p *task.PolicyInput) {
			p.Context = &task.ContextEligibility{Endpoints: []string{endpoints[0], endpoints[0]}}
		},
		"too many endpoints": func(p *task.PolicyInput) { p.Context = &task.ContextEligibility{Endpoints: endpoints} },
		"release review no issuer": func(p *task.PolicyInput) {
			p.ReleaseReviewed = []task.ReleaseReviewedTask{{IssuerID: "other", TaskSHA256: [32]byte{1}}}
		},
	}
	for name, mutate := range invalid {
		p := policy()
		mutate(&p)
		if s, err := task.NewLocalPolicy(p); err != task.ErrTrustModel || s.Valid() {
			t.Errorf("%s: invalid policy escaped", name)
		}
	}

	p := policy()
	p.Device = task.DeviceConstraints{MinimumBatteryPercent: &ten, ChargingOnly: true}
	p.Context = &task.ContextEligibility{Paths: task.Paths{Proxy: true}, Endpoints: []string{endpoints[0], endpoints[1]}}
	p.ReleaseReviewed = []task.ReleaseReviewedTask{{IssuerID: "reviewed-issuer", TaskSHA256: [32]byte{7}}}
	s, err := task.NewLocalPolicy(p)
	if err != nil {
		t.Fatal("valid restrictions rejected")
	}
	// Mutating the input or a returned copy changes nothing.
	ten = 90
	p.Context.Endpoints[0] = "https://changed.example.org/v1"
	p.ReleaseReviewed[0].TaskSHA256 = [32]byte{8}
	d := s.Device()
	*d.MinimumBatteryPercent = 0
	c, ok := s.ContextEligibility()
	c.Endpoints[1] = "https://changed.example.org/v2"
	d, again := s.Device(), mustEligibility(t, s)
	if !ok || *d.MinimumBatteryPercent != 10 || !d.ChargingOnly || !d.Active() || again.Endpoints[0] != endpoints[0] || again.Endpoints[1] != endpoints[1] || !s.ReleaseReviewed("reviewed-issuer", [32]byte{7}) {
		t.Fatal("restriction inputs alias caller memory")
	}
	if again.Permits("direct") || !again.Permits("proxy") || again.Permits("") {
		t.Fatal("path eligibility")
	}
	// Narrowing keeps configured order and never adds an endpoint.
	if got := again.Narrow(endpoints[1], "https://unlisted.example.org/v1", endpoints[0], ""); len(got) != 2 || got[0] != endpoints[1] || got[1] != endpoints[0] {
		t.Fatalf("narrowed %q", got)
	}
	if got := (task.ContextEligibility{}).Narrow(endpoints[0]); len(got) != 0 {
		t.Fatal("an empty eligibility permits an endpoint")
	}
	// Absent inputs restrict nothing; forbidding is a valid explicit input.
	base := policy()
	plain, _ := task.NewLocalPolicy(base)
	if _, ok := plain.ContextEligibility(); ok || plain.Device().Active() {
		t.Fatal("absent inputs restrict")
	}
	base.Context = &task.ContextEligibility{}
	if forbid, err := task.NewLocalPolicy(base); err != nil || mustEligibility(t, forbid).Permits("direct") {
		t.Fatal("forbidding eligibility")
	}
	for _, v := range []any{task.DeviceConstraints{}, task.ContextEligibility{}, task.ReleaseReviewedTask{}} {
		if _, err := json.Marshal(v); err == nil {
			t.Errorf("%T has a JSON form", v)
		}
	}
}

func mustEligibility(t *testing.T, p task.LocalPolicy) task.ContextEligibility {
	t.Helper()
	c, ok := p.ContextEligibility()
	if !ok {
		t.Fatal("no eligibility")
	}
	return c
}
