package barometer

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/platform"
)

// The portable gates shared by the M-*-PRECONDITIONS families. Each family's
// TestM<Id>Preconditions runs these on its safetyFixture and adds its own
// rows. Every gate rejects before contact: the recording transport and, where
// the test resolves, the system resolver see nothing. The native and host
// part of each row (platform APIs, routing policy, deployment review) is not
// exercised here; docs/measurement-requirements.md records where it is owned.

// gateOutcome admits taskJSON on a fresh fixture harness after edit and set,
// starts one run and returns the admission or start code, the execution
// state and reason of a written result, and the contacts made.
func (f safetyFixture) gateOutcome(t *testing.T, taskJSON string, edit func(*task.PolicyInput), set func(*harness)) (code, state, reason string, contacts int32) {
	t.Helper()
	h, count := f.start(t)
	if edit != nil {
		editPolicy(t, h, edit)
	}
	if set != nil {
		set(h)
	}
	c := h.initialized(t)
	defer c.Close()
	th, code := h.tryAdmit(t, c, taskJSON)
	if code == "ok" {
		code = startCode(t, c, th)
	}
	if records := h.run.output.Records(); len(records) > 0 {
		state, reason = executionOf(t, records[len(records)-1])
	}
	return code, state, reason, count()
}

// withExecution adds members to the task's execution object.
func withExecution(t *testing.T, taskJSON, members string) string {
	t.Helper()
	if !strings.Contains(taskJSON, `"limits":{}`) {
		t.Fatal("fixture task has no default limits")
	}
	return strings.Replace(taskJSON, `"limits":{}`, `"limits":{},`+members, 1)
}

type gateWant struct{ code, state, reason string }

func (f safetyFixture) wantGate(t *testing.T, name, taskJSON string, edit func(*task.PolicyInput), set func(*harness), want gateWant) {
	t.Helper()
	code, state, reason, contacts := f.gateOutcome(t, taskJSON, edit, set)
	if contacts != 0 {
		t.Fatalf("%s: %d contacts before the gate passed", name, contacts)
	}
	if code != want.code || state != want.state || reason != want.reason {
		t.Fatalf("%s: %s %s %s, want %s %s %s", name, code, state, reason, want.code, want.state, want.reason)
	}
}

// scriptAccess replaces the harness network with one that reports state
// from the given read on; read 1 is the reservation check in StartRun.
func scriptAccess(h *harness, from int, state platform.AccessState) {
	n := &net01Network{Network: h.run.network}
	n.script = func(_ context.Context, read int) (platform.AccessState, error, bool) {
		return state, nil, read >= from
	}
	h.adapters.run.network = n
}

// preconditionGates are the portable gates every family shares. A gate that
// is unsatisfied when StartRun reserves the run defers it: run_not_due, no
// attempt and no result, and the host may start it later while it is still
// eligible. One that becomes unsatisfied after reservation and before the
// first attempt yields not_run with the registered reason. Neither contacts.
func preconditionGates(t *testing.T, f safetyFixture) {
	t.Helper()
	deferred := gateWant{code: "run_not_due"}
	t.Run("network state unavailable", func(t *testing.T) {
		f.wantGate(t, "unreadable access", f.task("gate-network"), nil, func(h *harness) { h.run.network.Fail() }, deferred)
	})
	t.Run("stricter task constraint", func(t *testing.T) {
		f.wantGate(t, "transport at reservation", withExecution(t, f.task("gate-transport"), `"network":{"allowed_transports":["cellular"]}`), nil, nil, deferred)
		lost := defaultAccess
		lost.Generation, lost.Type = 2, platform.NetworkCellular
		f.wantGate(t, "transport lost before the attempt", withExecution(t, f.task("gate-lost"), `"network":{"allowed_transports":["wifi"]}`), nil,
			func(h *harness) { scriptAccess(h, 2, lost) }, gateWant{"ok", "not_run", "constraints_unsatisfied"})
	})
	t.Run("unintended path", func(t *testing.T) {
		vpn := defaultAccess
		vpn.VPN, vpn.Validated = platform.VPNDetected, platform.AvailabilityYes
		f.wantGate(t, "VPN route excluded by the task", withExecution(t, f.task("gate-vpn"), `"network":{"vpn_policy":"exclude"}`), nil,
			func(h *harness) { h.run.network.Set(vpn) }, deferred)
		// A proxy-path task never falls back to a direct contact: it is
		// rejected where the test has no proxy semantics, and otherwise ends
		// not_run while no proxy is configured.
		code, state, reason, contacts := f.gateOutcome(t, withExecution(t, f.task("gate-proxy"), `"network_path":"proxy"`), nil, nil)
		if contacts != 0 || !(code == "task_schema_invalid" && state == "" || code == "ok" && state == "not_run" && reason == "proxy_not_configured") {
			t.Fatalf("proxy path without a proxy: %s %s %s, %d contacts", code, state, reason, contacts)
		}
	})
	t.Run("stricter local policy", func(t *testing.T) {
		f.wantGate(t, "local device restriction", f.task("gate-device"), func(in *task.PolicyInput) { in.Device = task.DeviceConstraints{ChargingOnly: true} },
			func(h *harness) { h.run.device.Set(platform.DeviceState{ChargingKnown: true}) }, deferred)
		if !f.traffic {
			return
		}
		// The deployment withholds the target registrations the issuer
		// grants: admission rejects the task before any contact.
		code, _, _, contacts := f.gateOutcome(t, f.task("gate-unlisted"), func(in *task.PolicyInput) { in.Permissions.RegistrationIDs = nil }, nil)
		if code != "target_not_registered" || contacts != 0 {
			t.Fatalf("registration withheld by local policy: %s, %d contacts", code, contacts)
		}
		// A registration that is no longer current when the attempt starts is
		// refused at contact, before any traffic.
		h, count := f.start(t)
		editPolicy(t, h, func(in *task.PolicyInput) {
			for i := range in.Targets {
				in.Targets[i].ExpiresAt = task.TimestampFromTime(epoch.Add(time.Second))
			}
		})
		c := h.initialized(t)
		defer c.Close()
		th := h.admit(t, c, f.task("gate-lapsed"))
		h.advance(2 * time.Second)
		if code := startCode(t, c, th); code != "ok" || count() != 0 {
			t.Fatalf("lapsed registration: %s, %d contacts", code, count())
		}
		if _, reason := executionOf(t, h.run.output.Records()[0]); reason != "local_policy_rejected" {
			t.Fatal("lapsed registration:", reason)
		}
	})
}

// unregistered requires that each task is rejected at admission with no
// contact, under the fixture's otherwise valid policy.
func (f safetyFixture) unregistered(t *testing.T, tasks ...string) {
	t.Helper()
	for _, data := range tasks {
		if code, contacts := f.admitCode(t, data, nil); code == "ok" || contacts != 0 {
			t.Fatalf("unregistered task admitted: %s, %d contacts", code, contacts)
		}
	}
}

// notEnabled requires that a deployment which does not enable the test
// rejects it: the local permissions list only another test and no
// registration.
func (f safetyFixture) notEnabled(t *testing.T) {
	t.Helper()
	code, contacts := f.admitCode(t, f.task("not-enabled"), func(in *task.PolicyInput) {
		in.Permissions.Tests = []task.TestVersion{{ID: "SYN-01", Version: "0.1.0"}}
		in.Permissions.RegistrationIDs = nil
		in.Permissions.AllowComposite = false
	})
	if code != "task_issuer_not_authorized" || contacts != 0 {
		t.Fatalf("test not enabled by local policy: %s, %d contacts", code, contacts)
	}
}
