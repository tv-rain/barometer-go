package output_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tv-rain/barometer-go/internal/result"
	"github.com/tv-rain/barometer-go/internal/testkit/output"
)

// A zero result.Value has no serialization, so these cases exercise the fault
// paths without constructing a result; the execution tests cover durable writes.
func TestFaultsApplyToTheSelectedCall(t *testing.T) {
	s := output.New()
	s.FailAt(1, output.BeforeWrite)
	if err := s.Put(context.Background(), result.Value{}); !errors.Is(err, output.ErrInjected) {
		t.Fatalf("Put() = %v", err)
	}
	if err := s.Put(context.Background(), result.Value{}); !errors.Is(err, result.ErrUnavailable) {
		t.Fatalf("unserializable value: Put() = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Put(ctx, result.Value{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled Put() = %v", err)
	}
	if s.Calls() != 3 || len(s.Records()) != 0 {
		t.Fatalf("calls = %d, records = %d", s.Calls(), len(s.Records()))
	}
}
