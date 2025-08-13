package emitter

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Emitter struct {
	listeners map[string][]func(any)
	mutex     sync.RWMutex
}

func New() *Emitter {
	return &Emitter{
		listeners: make(map[string][]func(any)),
	}
}

func (e *Emitter) On(event string, listener func(any)) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.listeners[event] = append(e.listeners[event], listener)
}

func (e *Emitter) Emit(event string, data any) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	// Use a WaitGroup to wait for all listeners to finish
	var wg sync.WaitGroup
	for _, listener := range e.listeners[event] {
		wg.Add(1)
		go func(listener func(any)) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Recovered from panic in listener for event %s: %v\n", event, r)
				}
			}()
			listener(data)
		}(listener)
	}
	wg.Wait() // Block until all listeners complete
}

func (e *Emitter) Clear() {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.listeners = make(map[string][]func(any))
}

// EmitAsync emits an event asynchronously without blocking
func (e *Emitter) EmitAsync(event string, data any) {
	e.mutex.RLock()
	listeners := make([]func(any), len(e.listeners[event]))
	copy(listeners, e.listeners[event])
	e.mutex.RUnlock()

	// Fire and forget - don't wait for listeners
	for _, listener := range listeners {
		go func(listener func(any)) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Recovered from panic in async listener for event %s: %v\n", event, r)
				}
			}()
			listener(data)
		}(listener)
	}
}

// EmitWithContext emits an event with context support
func (e *Emitter) EmitWithContext(ctx context.Context, event string, data any) error {
	e.mutex.RLock()
	listeners := make([]func(any), len(e.listeners[event]))
	copy(listeners, e.listeners[event])
	e.mutex.RUnlock()

	// Create a channel to signal completion
	done := make(chan struct{})
	var wg sync.WaitGroup

	for _, listener := range listeners {
		wg.Add(1)
		go func(listener func(any)) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Recovered from panic in context listener for event %s: %v\n", event, r)
				}
			}()
			listener(data)
		}(listener)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// EmitWithTimeout emits an event with a timeout
func (e *Emitter) EmitWithTimeout(event string, data any, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return e.EmitWithContext(ctx, event, data)
}

// ListenerCount returns the number of listeners for an event
func (e *Emitter) ListenerCount(event string) int {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return len(e.listeners[event])
}

// EventNames returns all registered event names
func (e *Emitter) EventNames() []string {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	names := make([]string, 0, len(e.listeners))
	for name := range e.listeners {
		names = append(names, name)
	}
	return names
}
