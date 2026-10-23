package execution_test

import (
	"context"
	"testing"

	"github.com/tv-rain/barometer-go/internal/execution"
	"github.com/tv-rain/barometer-go/internal/model/observation"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/platform"
	"github.com/tv-rain/barometer-go/internal/testkit/capability"
)

func direct() task.Execution {
	return task.Execution{NetworkPath: "direct", Network: task.NetworkRequirements{VPNPolicy: "any"}}
}

// Network conditions this build cannot evaluate are unsatisfied: Network
// Context requirements (DEV-647) and an unregistered path. Either registered
// path passes; a proxy path that cannot be used is recorded by its attempt's
// path check. Transport, VPN and validation conditions pass the static gate
// and are evaluated against the access context.
func TestNetworkConditionsFailClosed(t *testing.T) {
	proxy := direct()
	proxy.NetworkPath = "proxy"
	if !execution.NetworkSatisfiable(direct()) || !execution.NetworkSatisfiable(proxy) {
		t.Fatal("registered path without conditions")
	}
	for name, edit := range map[string]func(*task.Execution){
		"unregistered path": func(e *task.Execution) { e.NetworkPath = "socks" },
		"unset path":        func(e *task.Execution) { e.NetworkPath = "" },
		"context fields":    func(e *task.Execution) { e.Network.Context.RequiredFields = []string{"country_code"} },
		"country allowlist": func(e *task.Execution) { e.Network.Context.AllowedCountryCodes = []string{"NL"} },
	} {
		e := direct()
		edit(&e)
		if execution.NetworkSatisfiable(e) || (execution.Gates{Execution: e}).Satisfied(context.Background()) {
			t.Fatal(name)
		}
	}
	for name, edit := range map[string]func(*task.Execution){
		"transports":        func(e *task.Execution) { e.Network.AllowedTransports = []string{"wifi"} },
		"vpn exclude":       func(e *task.Execution) { e.Network.VPNPolicy = "exclude" },
		"require validated": func(e *task.Execution) { e.Network.RequireValidated = true },
	} {
		e := direct()
		edit(&e)
		if !execution.NetworkSatisfiable(e) || !execution.AccessRequired(e.Network) {
			t.Fatal(name)
		}
	}
	if execution.AccessRequired(direct().Network) {
		t.Fatal("access required without a condition")
	}
	// An unset or unregistered vpn_policy never satisfies a condition.
	for _, policy := range []string{"", "prefer", "ANY"} {
		if execution.NetworkConditions(task.NetworkRequirements{VPNPolicy: policy}, observation.Access{VPNState: "not_detected"}) {
			t.Fatal(policy)
		}
	}
}

// Consent is queried only when required; only valid consent passes.
func TestConsentGate(t *testing.T) {
	ctx := context.Background()
	for status, want := range map[platform.ConsentStatus]bool{platform.ConsentValid: true, platform.ConsentAbsent: false, platform.ConsentEnded: false, platform.ConsentUnavailable: false} {
		c := capability.NewConsent(status)
		if (execution.Gates{Execution: direct(), ConsentRequired: true, Consent: c}).Satisfied(ctx) != want || c.Calls() != 1 {
			t.Fatal(status)
		}
	}
	failing := capability.NewConsent(platform.ConsentValid)
	failing.Fail()
	if (execution.Gates{Execution: direct(), ConsentRequired: true, Consent: failing}).Satisfied(ctx) {
		t.Fatal("consent error")
	}
	if (execution.Gates{Execution: direct(), ConsentRequired: true}).Satisfied(ctx) {
		t.Fatal("missing consent capability")
	}
	unused := capability.NewConsent(platform.ConsentAbsent)
	if !(execution.Gates{Execution: direct(), Consent: unused}).Satisfied(ctx) || unused.Calls() != 0 {
		t.Fatal("consent queried without a requirement")
	}
	done, cancel := context.WithCancel(ctx)
	cancel()
	if (execution.Gates{Execution: direct()}).Satisfied(done) {
		t.Fatal("cancelled context satisfied")
	}
}

// Battery and charging requirements need a known, sufficient state.
func TestDeviceGate(t *testing.T) {
	ctx := context.Background()
	twenty := 20
	battery := direct()
	battery.Device.MinimumBatteryPercent = &twenty
	charging := direct()
	charging.Device.ChargingOnly = true
	cases := []struct {
		name  string
		e     task.Execution
		state platform.DeviceState
		want  bool
	}{
		{"unknown battery", battery, platform.DeviceState{BatteryPercent: 99}, false},
		{"limit -1", battery, platform.DeviceState{BatteryPercent: 19, BatteryKnown: true}, false},
		{"limit", battery, platform.DeviceState{BatteryPercent: 20, BatteryKnown: true}, true},
		{"limit +1", battery, platform.DeviceState{BatteryPercent: 21, BatteryKnown: true}, true},
		{"unknown charging", charging, platform.DeviceState{Charging: true}, false},
		{"not charging", charging, platform.DeviceState{ChargingKnown: true}, false},
		{"charging", charging, platform.DeviceState{Charging: true, ChargingKnown: true}, true},
	}
	for _, c := range cases {
		if got := (execution.Gates{Execution: c.e, Device: capability.NewDevice(c.state)}).Satisfied(ctx); got != c.want {
			t.Fatal(c.name)
		}
	}
	failing := capability.NewDevice(platform.DeviceState{BatteryPercent: 99, BatteryKnown: true})
	failing.Fail()
	if (execution.Gates{Execution: battery, Device: failing}).Satisfied(ctx) || (execution.Gates{Execution: battery}).Satisfied(ctx) {
		t.Fatal("unavailable device state")
	}
	unused := capability.NewDevice(platform.DeviceState{})
	if !(execution.Gates{Execution: direct(), Device: unused}).Satisfied(ctx) || unused.Calls() != 0 {
		t.Fatal("device queried without a requirement")
	}
}
