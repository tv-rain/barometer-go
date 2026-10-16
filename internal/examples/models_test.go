// These non-normative examples exercise internal Foundation models only.
// See README.md for authoritative contracts and the pending runtime boundary.
package examples

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tv-rain/barometer-go/internal/config"
	"github.com/tv-rain/barometer-go/internal/execution"
	"github.com/tv-rain/barometer-go/internal/jsondecode"
	model "github.com/tv-rain/barometer-go/internal/model/result"
	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/taskdecode"
	"github.com/tv-rain/barometer-go/internal/validation"
)

// fixedClock is synthetic test capability input, never a production time source.
type fixedClock struct{}

func (fixedClock) Now(ctx context.Context) (time.Time, error) {
	return time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), ctx.Err()
}

func Example_configurationPreparation() {
	decoder, err := config.NewDecoder(config.Contract, config.Approvals{})
	if err != nil {
		fmt.Println("configuration decoder unavailable")
		return
	}
	// No proxy, context endpoint, linkage, or trusted policy is enabled here.
	snapshot, err := decoder.Decode([]byte(`{}`))
	if err != nil {
		fmt.Println("configuration rejected")
		return
	}
	_, err = config.Prepare(snapshot, config.Capabilities{})
	fmt.Println("missing clock rejected:", errors.Is(err, config.ErrUnavailable))
	prepared, err := config.Prepare(snapshot, config.Capabilities{Clock: fixedClock{}})
	if err != nil {
		fmt.Println("preparation unavailable")
		return
	}
	fmt.Println("prepared linkage mode:", prepared.Configuration().Linkage().Mode)
	// Preparation does not initialize a client or authorize measurements.
	_, err = decoder.Decode([]byte(`{"authenticated":true}`))
	fmt.Println("configuration cannot assert trust:", errors.Is(err, jsondecode.ErrField))
	// Output:
	// missing clock rejected: true
	// prepared linkage mode: none
	// configuration cannot assert trust: true
}

// SYN-01 is an invented empty schema for demonstrating common task decoding.
// It is not a registered measurement or a shortcut for a production test schema.
type syntheticParameters struct{}

func (syntheticParameters) TaskParameters() {}

type syntheticCodec struct{}

func (syntheticCodec) Definition() taskdecode.Definition {
	return taskdecode.Definition{
		TestID: "SYN-01", TestVersion: "0.1.0",
		Input:   &jsondecode.Schema{Kind: jsondecode.Object},
		Options: &jsondecode.Schema{Kind: jsondecode.Object},
	}
}

func (syntheticCodec) Normalize(_, _ *jsondecode.Value) (task.Parameters, error) {
	return syntheticParameters{}, nil
}

func Example_strictTaskDecoding() {
	decoder, err := taskdecode.New([]taskdecode.TestCodec{syntheticCodec{}})
	if err != nil {
		fmt.Println("task decoder unavailable")
		return
	}
	const input = `{"schema_version":"0.1.0","task_id":"synthetic-task","test_id":"SYN-01","test_version":"0.1.0","input":{},"options":{},"schedule":{"not_after":"2026-10-17T00:00:00Z"},"execution":{"limits":{}},"safety":{"risk_level":"low","requires_user_consent":false}}`
	// Fixed containing-owner timing input, never read from task JSON. Remote
	// origin here is a validation fixture, not proof of authenticated delivery.
	accepted := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	timing := validation.TimingContext{Origin: "remote", AcceptedAt: accepted, IssuedAt: &accepted}
	candidate, err := decoder.Decode([]byte(input), timing)
	if err != nil {
		fmt.Println("task rejected")
		return
	}
	fmt.Println("candidate only:", candidate.TestID)
	// Never print task bytes or raw errors from a containing integration.
	unknown := strings.Replace(input, `"input":{}`, `"input":{"unexpected":true}`, 1)
	candidate, err = decoder.Decode([]byte(unknown), timing)
	fmt.Println("unknown field rejected atomically:", candidate == nil && errors.Is(err, jsondecode.ErrField))
	duplicate := strings.Replace(input, `"options":{}`, `"options":{},"options":{}`, 1)
	candidate, err = decoder.Decode([]byte(duplicate), timing)
	fmt.Println("duplicate rejected atomically:", candidate == nil && err != nil)
	// Output:
	// candidate only: SYN-01
	// unknown field rejected atomically: true
	// duplicate rejected atomically: true
}

func Example_cancellationState() {
	running, err := execution.NewAttempt().Start()
	if err != nil {
		fmt.Println("attempt transition rejected")
		return
	}
	// This changes a pure value. It does not cancel network work or perform
	// cleanup; runtime cancellation remains a separate implementation gate.
	stopped, err := running.Stop(task.CancelledByHost)
	if err != nil {
		fmt.Println("stop transition rejected")
		return
	}
	fmt.Println("terminal:", stopped.Terminal())
	fmt.Println("reason:", stopped.Reason())
	_, retry := stopped.RetryReason()
	fmt.Println("retryable:", retry)
	unchanged, err := stopped.Complete()
	fmt.Println("completion cannot overwrite cancellation:", errors.Is(err, execution.ErrTransition) && unchanged == stopped)
	// Output:
	// terminal: true
	// reason: cancelled_by_host
	// retryable: false
	// completion cannot overwrite cancellation: true
}

func Example_resultSerialization() {
	// Entirely synthetic: no attempt, device acquisition, or measurement ran.
	test := model.Test{ID: "TCP-01", Version: "0.1.0", ResultVersion: "0.1.0"}
	// Platform is a schema fixture, not a report about this test's host.
	implementation := model.Implementation{LibraryVersion: "0.0.0-example", Platform: "android"}
	decoder, err := result.New(result.Binding{
		TaskID: "synthetic-task", Output: result.Output{Test: test},
		MaxAttempts: 1, MaxDurationMS: 1000, Implementation: implementation,
	})
	if err != nil {
		fmt.Println("result binding rejected")
		return
	}
	record := model.Result{
		SchemaVersion: "0.1.0", TaskID: "synthetic-task",
		RunID: "AQEBAQEBAQEBAQEBAQEBAQ", Attempt: 1, // fixed fixture ID; never generate runtime IDs this way
		Test: test, Implementation: implementation,
		Time:      model.Time{StartedAt: "2026-10-16T00:00:00Z", DurationMS: 0},
		Execution: model.Execution{State: "not_run", Reason: task.LocalPolicyRejected},
	}
	value, err := decoder.Build(record)
	if err != nil {
		fmt.Println("result rejected")
		return
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		fmt.Println("serialization rejected")
		return
	}
	decoded, err := decoder.Decode(encoded)
	if err != nil {
		fmt.Println("round trip rejected")
		return
	}
	again, err := json.Marshal(decoded)
	fmt.Println("canonical round trip:", err == nil && bytes.Equal(encoded, again))
	fmt.Println("execution:", decoded.View().Execution.State)
	fmt.Println("measurement absent:", decoded.View().Measurement == nil)
	_, err = json.Marshal(record)
	fmt.Println("detached record cannot bypass validation:", errors.Is(err, model.ErrEncoding))
	record.Attempt = 2
	_, err = decoder.Build(record)
	fmt.Println("attempt beyond local binding rejected:", err != nil)
	// Output:
	// canonical round trip: true
	// execution: not_run
	// measurement absent: true
	// detached record cannot bypass validation: true
	// attempt beyond local binding rejected: true
}
