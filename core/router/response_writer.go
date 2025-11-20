package router

import (
	"bufio"
	"net"
	"net/http"
)

// ResponseWriter wraps http.ResponseWriter with additional functionality
type responseWriter struct {
	http.ResponseWriter
	status  int
	size    int
	written bool
}

// WriteHeader sets the status code
func (w *responseWriter) WriteHeader(code int) {
	if w.written {
		return
	}
	w.status = code
	w.written = true
	w.ResponseWriter.WriteHeader(code)
}

// Write writes the data to the connection
func (w *responseWriter) Write(data []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	size, err := w.ResponseWriter.Write(data)
	w.size += size
	return size, err
}

// Status returns the response status code
func (w *responseWriter) Status() int {
	return w.status
}

// Size returns the size of the response body
func (w *responseWriter) Size() int {
	return w.size
}

// Written returns true if the response has been written
func (w *responseWriter) Written() bool {
	return w.written
}

// Hijack implements the http.Hijacker interface
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// Flush implements the http.Flusher interface
func (w *responseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Push implements the http.Pusher interface
func (w *responseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

// ResponseWriter interface extends http.ResponseWriter
type ResponseWriter interface {
	http.ResponseWriter
	Status() int
	Size() int
	Written() bool
	Hijack() (net.Conn, *bufio.ReadWriter, error)
	Flush()
	Push(target string, opts *http.PushOptions) error
}
