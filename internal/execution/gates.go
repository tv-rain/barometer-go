package execution

import (
	"context"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/platform"
)

// Preconditions is evaluated immediately before every attempt, including a
// retry, because consent and device state can change after admission and
// between attempts (REQ-SAFETY-0001). False before the first attempt fails the
// run closed with no result; false before a retry ends the run without it.
type Preconditions interface {
	Satisfied(ctx context.Context) bool
}

// Gates is the attempt-start check of one admitted task: its execution
// constraints and, when admission recorded the requirement, consent. Every
// missing capability, error or undetermined value is unsatisfied; nothing is
// guessed. Concurrency is the caller's, because it spans runs.
type Gates struct {
	Execution       task.Execution
	ConsentRequired bool
	Consent         platform.Consent // queried only when ConsentRequired
	Device          platform.Device  // queried only for a device requirement
}

// Satisfied applies the fail-closed rule to each condition.
// requirement-ref: REQ-EXECUTION-0028
// requirement-ref: REQ-SAFETY-0020
func (g Gates) Satisfied(ctx context.Context) bool {
	return ctx.Err() == nil && NetworkSatisfiable(g.Execution) && g.device(ctx) && g.consent(ctx)
}

// NetworkSatisfiable reports whether the task states no network condition
// this build cannot determine. Until Network paths and context (DEV-578)
// evaluates them, a proxy path, an allowed-transport list, a vpn_policy other
// than any, require_validated and Network Context requirements are
// unsatisfied, so such a task never runs.
func NetworkSatisfiable(e task.Execution) bool {
	n := e.Network
	return e.NetworkPath == "direct" && len(n.AllowedTransports) == 0 && n.VPNPolicy == "any" && !n.RequireValidated &&
		len(n.Context.RequiredFields) == 0 && len(n.Context.AllowedCountryCodes) == 0
}

func (g Gates) device(ctx context.Context) bool {
	d := g.Execution.Device
	if d.MinimumBatteryPercent == nil && !d.ChargingOnly {
		return true
	}
	if g.Device == nil {
		return false
	}
	s, err := g.Device.State(ctx)
	if err != nil {
		return false
	}
	if d.MinimumBatteryPercent != nil && (!s.BatteryKnown || s.BatteryPercent < *d.MinimumBatteryPercent) {
		return false
	}
	return !d.ChargingOnly || s.ChargingKnown && s.Charging
}

func (g Gates) consent(ctx context.Context) bool {
	if !g.ConsentRequired {
		return true
	}
	if g.Consent == nil {
		return false
	}
	s, err := g.Consent.Status(ctx)
	return err == nil && s == platform.ConsentValid
}
