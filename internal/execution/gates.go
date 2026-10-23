package execution

import (
	"context"
	"slices"

	"github.com/tv-rain/barometer-go/internal/model/observation"
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
// guessed. Concurrency is the caller's, because it spans runs. Conditions on
// the access context are not part of Gates: they need the attempt's own
// access snapshot, so run reservation and the engine evaluate them with
// NetworkConditions.
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
// this build cannot evaluate. Until the proxy path (DEV-645) and egress
// lookup (DEV-647) exist, a proxy path and Network Context requirements are
// unsatisfied, so such a task never runs. Transport, VPN and validation
// conditions depend on the access context and are evaluated against it by
// NetworkConditions.
func NetworkSatisfiable(e task.Execution) bool {
	n := e.Network
	return e.NetworkPath == "direct" && len(n.Context.RequiredFields) == 0 && len(n.Context.AllowedCountryCodes) == 0
}

// AccessRequired reports whether the task states a condition on the access
// context: an allowed-transport list, a vpn_policy other than any, or
// require_validated.
func AccessRequired(n task.NetworkRequirements) bool {
	return len(n.AllowedTransports) > 0 || n.VPNPolicy != "any" || n.RequireValidated
}

// NetworkConditions evaluates the access-context conditions against one
// normalized access reading. A value that is unknown, or outside the
// registered vocabulary, never satisfies a condition (REQ-EXECUTION-0031):
// unknown is not not_detected (REQ-EXECUTION-0019), and only the platform's
// own validated yes satisfies require_validated, with no connectivity check
// of the library's own (GO-SPEC-ISSUE-013).
// requirement-ref: REQ-EXECUTION-0016
// requirement-ref: REQ-EXECUTION-0017
// requirement-ref: REQ-EXECUTION-0018
// requirement-ref: REQ-EXECUTION-0019
// requirement-ref: REQ-EXECUTION-0020
// requirement-ref: REQ-EXECUTION-0031
func NetworkConditions(n task.NetworkRequirements, a observation.Access) bool {
	if len(n.AllowedTransports) > 0 && (a.NetworkType != "wifi" && a.NetworkType != "cellular" || !slices.Contains(n.AllowedTransports, a.NetworkType)) {
		return false
	}
	switch n.VPNPolicy {
	case "any":
	case "exclude":
		if a.VPNState != "not_detected" {
			return false
		}
	case "require":
		if a.VPNState != "detected" {
			return false
		}
	default:
		return false
	}
	return !n.RequireValidated || a.Validated == "yes"
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
