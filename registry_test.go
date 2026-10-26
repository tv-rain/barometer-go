package barometer

import (
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
)

func TestProductionRegistryContainsOnlyDeliveredMeasurements(t *testing.T) {
	first, err := productionRegistry(nil)
	if err != nil || first == nil {
		t.Fatalf("productionRegistry(nil) = %v, %v", first, err)
	}
	second, err := productionRegistry(nil)
	if err != nil || second == nil || first == second {
		t.Fatal("production registries must be instance-owned")
	}
	if len(first.Codecs()) != 3 || len(second.Codecs()) != 3 {
		t.Fatal("expected DNS-01, TCP-01 and TLS-01")
	}
	entry, ok := first.Lookup(task.TestVersion{ID: "TCP-01", Version: "0.1.0"})
	if !ok || entry.Runner == nil || entry.BaselineRisk != task.LowRisk || !entry.Step || entry.Composite {
		t.Fatal("TCP-01 composition missing or incorrect")
	}
	entry, ok = first.Lookup(tls01)
	if !ok || entry.Runner == nil || entry.BaselineRisk != task.LowRisk || !entry.Step || entry.Composite {
		t.Fatal("TLS-01 composition missing or incorrect")
	}
	for _, id := range []string{"SYN-01", "DNS-03", "NET-05"} {
		key := task.TestVersion{ID: id, Version: "0.1.0"}
		if _, ok := first.Lookup(key); ok {
			t.Fatalf("undelivered test %s is registered", id)
		}
		if _, ok := first.Runner(key); ok {
			t.Fatalf("undelivered runner %s is registered", id)
		}
	}
}

// The production default reaches actual admission, while synthetic injection
// remains available only through the internal seam used by tests.
// requirement-ref: REQ-TRUST-0036
func TestClientProductionRegistryRejectsUnregisteredTasks(t *testing.T) {
	for _, injected := range []bool{false, true} {
		name := "production default"
		if injected {
			name = "synthetic injection"
		}
		t.Run(name, func(t *testing.T) {
			h := admittingHarness(t)
			if !injected {
				h.adapters.registry = nil
			}
			c := h.initialized(t)
			defer c.Close()
			if c.bound.registry == nil {
				t.Fatal("bind did not install a registry")
			}
			if c.rt.run.registry != c.bound.registry {
				t.Fatal("execution did not receive the bound registry")
			}
			for _, id := range []string{"first", "second"} {
				data := synTask(id, time.Hour)
				handle, err := c.AdmitTask(data, h.embedded(t, data))
				if injected {
					if err != nil || handle == "" {
						t.Fatalf("synthetic registry admission: %s", codeOf(err))
					}
				} else if handle != "" || codeOf(err) != "task_unsupported" {
					t.Fatalf("production admission = %q, %s", handle, codeOf(err))
				}
			}
			if !injected && (taskHandles(c) != 0 || h.replay.Commits() != 0 || len(c.rt.runs.entries) != 0) {
				t.Fatal("unsupported tasks left handles, replay writes or runs")
			}
		})
	}
}
