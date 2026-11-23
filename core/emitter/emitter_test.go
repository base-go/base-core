package emitter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestZeroValueEmitterOnThenEmit(t *testing.T) {
	var e Emitter
	var got any
	e.On("user.registered", func(data any) { got = data })
	e.Emit("user.registered", "alice")
	if got != "alice" {
		t.Fatalf("listener got %v, want alice", got)
	}
}

func TestNewEmitterOnThenEmit(t *testing.T) {
	e := New()
	var calls atomic.Int32
	e.On("a", func(any) { calls.Add(1) })
	e.On("a", func(any) { calls.Add(1) })
	e.Emit("a", nil)
	if calls.Load() != 2 {
		t.Fatalf("got %d calls, want 2", calls.Load())
	}
}

func TestEmitterConcurrentOnAndEmit(t *testing.T) {
	e := &Emitter{}
	var calls atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			e.On("tick", func(any) { calls.Add(1) })
		}()
		go func() {
			defer wg.Done()
			e.Emit("tick", nil)
			e.EmitAsync("tick", nil)
			_ = e.ListenerCount("tick")
		}()
	}
	wg.Wait()
	if n := e.ListenerCount("tick"); n != 50 {
		t.Fatalf("got %d listeners, want 50", n)
	}
}

func TestEmitterRecoversPanickingListener(t *testing.T) {
	e := &Emitter{}
	var ran atomic.Bool
	e.On("boom", func(any) { panic("listener failed") })
	e.On("boom", func(any) { ran.Store(true) })

	e.Emit("boom", nil) // must not panic
	if !ran.Load() {
		t.Fatal("the second listener did not run after the first panicked")
	}
}

func TestListenerMayRegisterDuringEmit(t *testing.T) {
	e := &Emitter{}
	e.On("first", func(any) { e.On("second", func(any) {}) })

	done := make(chan struct{})
	go func() {
		e.Emit("first", nil)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Emit deadlocked when a listener called On")
	}
	if e.ListenerCount("second") != 1 {
		t.Fatal("listener registered during Emit was lost")
	}
}
