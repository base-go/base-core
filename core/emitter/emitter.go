package emitter

import "sync"

type Emitter struct {
	listeners map[string][]func(interface{})
	mutex     sync.RWMutex
}

func New() *Emitter {
	return &Emitter{
		listeners: make(map[string][]func(interface{})),
	}
}

func (e *Emitter) On(event string, listener func(interface{})) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.listeners[event] = append(e.listeners[event], listener)
}

func (e *Emitter) Off(event string, listener func(interface{})) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	for i, l := range e.listeners[event] {
		if funcEqual(l, listener) {
			e.listeners[event] = append(e.listeners[event][:i], e.listeners[event][i+1:]...)
			break
		}
	}
}

func (e *Emitter) Emit(event string, data interface{}) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	for _, listener := range e.listeners[event] {
		listener(data)
	}
}

func (e *Emitter) Clear() {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.listeners = make(map[string][]func(interface{}))
}

func funcEqual(a, b interface{}) bool {
	return &a == &b
}
