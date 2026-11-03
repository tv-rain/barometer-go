package admission_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/tv-rain/barometer-go/internal/model/task"
	"github.com/tv-rain/barometer-go/internal/persistence"
)

// heldReplay blocks the first generation read, ignoring its context, until
// release is closed, as a stuck replay backend would.
type heldReplay struct {
	replayBackend
	once    sync.Once
	entered chan struct{}
	release chan struct{}
}

func (h *heldReplay) HighestGeneration(ctx context.Context, now time.Time) (uint64, error) {
	held := false
	h.once.Do(func() { held = true })
	if held {
		close(h.entered)
		<-h.release
	}
	return h.replayBackend.HighestGeneration(ctx, now)
}

var _ persistence.ReplayStore = (*heldReplay)(nil)

// An admission waiting for steps 10 and 11 while another holds them across a
// replay call that ignores its context fails closed when its own context
// ends, and the held admission completes once the call returns.
func TestAdmitWaiterHonorsContext(t *testing.T) {
	for _, kind := range replayBackends {
		t.Run(string(kind), func(t *testing.T) {
			h := newHarnessOn(t, kind, mustPolicy(t, permissive()))
			held := &heldReplay{replayBackend: h.replay, entered: make(chan struct{}), release: make(chan struct{})}
			h.replay = held
			h.restart(t, 2)
			first := taskWith(map[string]any{"task_id": "held-task"})
			second := taskWith(map[string]any{"task_id": "waiting-task"})
			firstRef, secondRef := mustEmbedded(t, h.store, first), mustEmbedded(t, h.store, second)

			done := make(chan error, 1)
			go func() {
				_, err := h.admitter.Admit(context.Background(), firstRef, first)
				done <- err
			}()
			<-held.entered
			guard := time.AfterFunc(10*time.Second, func() { close(held.release) })
			defer guard.Stop()

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			_, err := h.admitter.Admit(ctx, secondRef, second)
			cancel()
			expect(t, "waiting admission", err, task.TrustUnavailable)
			if !guard.Stop() {
				t.Fatal("the waiting admission outlived its deadline")
			}
			close(held.release)
			if err := <-done; err != nil {
				t.Fatal("held admission failed:", err)
			}
			if _, err := h.admitEmbedded(t, second); err != nil {
				t.Fatal("admission did not recover:", err)
			}
		})
	}
}
