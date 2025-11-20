package http

import (
	"bufio"
	"context"
	"crypto/tls"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"time"
)

// Router defines the interface for HTTP routing
type Router interface {
	// Group creates a new route group with the given prefix
	Group(prefix string, middlewares ...Middleware) Router

	// HTTP methods
	GET(path string, handler Handler, middlewares ...Middleware)
	POST(path string, handler Handler, middlewares ...Middleware)
	PUT(path string, handler Handler, middlewares ...Middleware)
	DELETE(path string, handler Handler, middlewares ...Middleware)
	PATCH(path string, handler Handler, middlewares ...Middleware)
	HEAD(path string, handler Handler, middlewares ...Middleware)
	OPTIONS(path string, handler Handler, middlewares ...Middleware)

	// Handle registers a handler for the given method and path
	Handle(method, path string, handler Handler, middlewares ...Middleware)

	// Use adds middleware to this router
	Use(middlewares ...Middleware)

	// ServeHTTP implements the http.Handler interface
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// Context defines the interface for HTTP request context
type Context interface {
	// Request returns the underlying HTTP request
	Request() *http.Request

	// Response returns the response writer
	Response() ResponseWriter

	// Context returns the request context
	Context() context.Context

	// WithContext returns a copy with a new context
	WithContext(ctx context.Context) Context

	// Param returns a URL parameter by name
	Param(key string) string

	// Query returns a query parameter by name
	Query(key string) string

	// QueryDefault returns a query parameter with a default value
	QueryDefault(key, defaultValue string) string

	// Header returns a request header by name
	Header(key string) string

	// SetHeader sets a response header
	SetHeader(key, value string)

	// FormValue returns a form value by name
	FormValue(key string) string

	// FormFile returns a multipart form file
	FormFile(key string) (*multipart.FileHeader, error)

	// MultipartForm returns the multipart form
	MultipartForm() (*multipart.Form, error)

	// Body returns the request body
	Body() io.ReadCloser

	// Bind binds the request body to a struct
	Bind(v any) error

	// BindJSON binds JSON request body to a struct
	BindJSON(v any) error

	// BindQuery binds query parameters to a struct
	BindQuery(v any) error

	// JSON sends a JSON response
	JSON(code int, v any) error

	// String sends a string response
	String(code int, format string, values ...any) error

	// Data sends raw data response
	Data(code int, contentType string, data []byte) error

	// File sends a file response
	File(filepath string) error

	// Status sets the response status code
	Status(code int) Context

	// Redirect performs an HTTP redirect
	Redirect(code int, location string) error

	// Error sends an error response
	Error(code int, err error) error

	// NoContent sends a no content response
	NoContent() error

	// Get retrieves a value from the context
	Get(key string) (any, bool)

	// Set stores a value in the context
	Set(key string, value any)

	// MustGet retrieves a value from the context, panicking if not found
	MustGet(key string) any

	// Next calls the next handler in the chain
	Next() error

	// Abort stops the handler chain
	Abort()

	// IsAborted returns true if the handler chain was aborted
	IsAborted() bool

	// ClientIP returns the client IP address
	ClientIP() string

	// ContentType returns the request content type
	ContentType() string
}

// ResponseWriter extends http.ResponseWriter with additional methods
type ResponseWriter interface {
	http.ResponseWriter

	// Status returns the response status code
	Status() int

	// Written returns true if the response has been written
	Written() bool

	// Size returns the size of the response body
	Size() int

	// Hijack implements the http.Hijacker interface
	Hijack() (net.Conn, *bufio.ReadWriter, error)

	// Flush implements the http.Flusher interface
	Flush()

	// Push implements the http.Pusher interface
	Push(target string, opts *http.PushOptions) error
}

// Handler defines the handler function signature
type Handler func(ctx Context) error

// Middleware defines the middleware function signature
type Middleware func(next Handler) Handler

// ErrorHandler handles errors returned by handlers
type ErrorHandler func(ctx Context, err error)

// Server defines the HTTP server interface
type Server interface {
	// Router returns the router
	Router() Router

	// Start starts the server
	Start(addr string) error

	// StartTLS starts the server with TLS
	StartTLS(addr, certFile, keyFile string) error

	// Shutdown gracefully shuts down the server
	Shutdown(ctx context.Context) error

	// Use adds global middleware
	Use(middlewares ...Middleware)

	// SetErrorHandler sets the error handler
	SetErrorHandler(handler ErrorHandler)
}

// ServerConfig contains server configuration
type ServerConfig struct {
	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out writes of the response
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the next request
	IdleTimeout time.Duration

	// MaxHeaderBytes controls the maximum number of bytes the server will read
	MaxHeaderBytes int

	// TLSConfig optionally provides a TLS configuration
	TLSConfig *tls.Config

	// ErrorHandler handles errors returned by handlers
	ErrorHandler ErrorHandler
}

// NewServer creates a new HTTP server with the given configuration
type NewServer func(config *ServerConfig) Server
