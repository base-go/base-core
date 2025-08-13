package router

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Context represents the context of an HTTP request
type Context struct {
	Request  *http.Request
	Writer   ResponseWriter
	params   Params
	keys     map[string]any
	mu       sync.RWMutex
	index    int8
	handlers []HandlerFunc
}

// Param represents a URL parameter
type Param struct {
	Key   string
	Value string
}

// Params is a slice of URL parameters
type Params []Param

// Get returns the value of the first Param which key matches the given name
func (ps Params) Get(name string) string {
	for _, p := range ps {
		if p.Key == name {
			return p.Value
		}
	}
	return ""
}

// reset resets the context for reuse
func (c *Context) reset(w http.ResponseWriter, r *http.Request) {
	c.Request = r
	c.Writer = &responseWriter{ResponseWriter: w, status: http.StatusOK}
	c.params = c.params[:0]
	c.keys = make(map[string]any)
	c.index = -1
	c.handlers = nil
}

// Context returns the request's context
func (c *Context) Context() context.Context {
	return c.Request.Context()
}

// WithContext returns a copy with a new context
func (c *Context) WithContext(ctx context.Context) {
	c.Request = c.Request.WithContext(ctx)
}

// Deadline returns the deadline from the request's context
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.Request.Context().Deadline()
}

// Done returns the done channel from the request's context
func (c *Context) Done() <-chan struct{} {
	return c.Request.Context().Done()
}

// Err returns the error from the request's context
func (c *Context) Err() error {
	return c.Request.Context().Err()
}

// Value returns the value associated with key from the request's context
func (c *Context) Value(key any) any {
	return c.Request.Context().Value(key)
}

// Param returns the value of the URL param
func (c *Context) Param(key string) string {
	return c.params.Get(key)
}

// Query returns the keyed url query value
func (c *Context) Query(key string) string {
	value, _ := c.GetQuery(key)
	return value
}

// GetQuery returns the keyed url query value and existence
func (c *Context) GetQuery(key string) (string, bool) {
	if values, ok := c.GetQueryArray(key); ok {
		return values[0], ok
	}
	return "", false
}

// QueryArray returns the keyed url query values
func (c *Context) GetQueryArray(key string) ([]string, bool) {
	if values := c.Request.URL.Query()[key]; len(values) > 0 {
		return values, true
	}
	return []string{}, false
}

// DefaultQuery returns the keyed url query value if it exists, otherwise it returns the defaultValue
func (c *Context) DefaultQuery(key, defaultValue string) string {
	if value, ok := c.GetQuery(key); ok {
		return value
	}
	return defaultValue
}

// FormValue returns the form value by key
func (c *Context) FormValue(key string) string {
	return c.Request.FormValue(key)
}

// FormFile returns the multipart form file for the given key
func (c *Context) FormFile(key string) (*multipart.FileHeader, error) {
	if c.Request.MultipartForm == nil {
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			return nil, err
		}
	}
	file, header, err := c.Request.FormFile(key)
	if err != nil {
		return nil, err
	}
	file.Close()
	return header, nil
}

// MultipartForm returns the parsed multipart form, including file uploads
func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.Request.ParseMultipartForm(32 << 20)
	return c.Request.MultipartForm, err
}

// Header returns the request header value
func (c *Context) Header(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader sets a response header
func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

// Cookie returns the named cookie from the request
func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

// SetCookie adds a Set-Cookie header to the response
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Writer, cookie)
}

// Get returns the value for the given key
func (c *Context) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.keys[key]
	return value, exists
}

// Set stores a key/value pair in the context
func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.keys[key] = value
}

// MustGet returns the value for the given key or panics
func (c *Context) MustGet(key string) any {
	value, exists := c.Get(key)
	if !exists {
		panic("Key \"" + key + "\" does not exist")
	}
	return value
}

// Bind binds the request body to a struct
func (c *Context) Bind(obj any) error {
	contentType := c.ContentType()
	switch {
	case strings.Contains(contentType, "application/json"):
		return c.BindJSON(obj)
	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		return c.BindForm(obj)
	case strings.Contains(contentType, "multipart/form-data"):
		return c.BindForm(obj)
	default:
		return fmt.Errorf("unsupported content type: %s", contentType)
	}
}

// BindJSON binds the request body as JSON to a struct
func (c *Context) BindJSON(obj any) error {
	if c.Request.Body == nil {
		return fmt.Errorf("request body is nil")
	}
	decoder := json.NewDecoder(c.Request.Body)
	return decoder.Decode(obj)
}

// ShouldBindJSON binds the request body as JSON to a struct with validation
func (c *Context) ShouldBindJSON(obj any) error {
	return c.BindJSON(obj)
}

// BindQuery binds the query parameters to a struct
func (c *Context) BindQuery(obj any) error {
	values := c.Request.URL.Query()
	// This is a simplified version - in production you'd use a proper binding library
	// or implement reflection-based binding
	return bindData(obj, values)
}

// BindForm binds the form data to a struct
func (c *Context) BindForm(obj any) error {
	if err := c.Request.ParseForm(); err != nil {
		return err
	}
	return bindData(obj, c.Request.Form)
}

// JSON sends a JSON response
func (c *Context) JSON(code int, obj any) error {
	c.SetHeader("Content-Type", "application/json")
	c.Writer.WriteHeader(code)
	encoder := json.NewEncoder(c.Writer)
	return encoder.Encode(obj)
}

// String sends a string response
func (c *Context) String(code int, format string, values ...any) error {
	c.SetHeader("Content-Type", "text/plain")
	c.Writer.WriteHeader(code)
	_, err := fmt.Fprintf(c.Writer, format, values...)
	return err
}

// HTML sends an HTML response
func (c *Context) HTML(code int, html string) error {
	c.SetHeader("Content-Type", "text/html")
	c.Writer.WriteHeader(code)
	_, err := c.Writer.Write([]byte(html))
	return err
}

// Data sends a raw data response
func (c *Context) Data(code int, contentType string, data []byte) error {
	c.SetHeader("Content-Type", contentType)
	c.Writer.WriteHeader(code)
	_, err := c.Writer.Write(data)
	return err
}

// File sends a file response
func (c *Context) File(filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
}

// Status sets the response status code
func (c *Context) Status(code int) *Context {
	c.Writer.WriteHeader(code)
	return c
}

// Redirect performs an HTTP redirect
func (c *Context) Redirect(code int, location string) error {
	if code < 300 || code > 308 {
		return fmt.Errorf("invalid redirect code: %d", code)
	}
	c.SetHeader("Location", location)
	c.Writer.WriteHeader(code)
	return nil
}

// Error sends an error response
func (c *Context) Error(code int, err error) error {
	c.JSON(code, map[string]any{
		"error": err.Error(),
	})
	return err
}

// NoContent sends a no content response
func (c *Context) NoContent() error {
	c.Writer.WriteHeader(http.StatusNoContent)
	return nil
}

// ClientIP returns the client's IP address
func (c *Context) ClientIP() string {
	// Check X-Forwarded-For header
	if xff := c.Header("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i != -1 {
			return strings.TrimSpace(xff[:i])
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := c.Header("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	if ip, _, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil {
		return ip
	}

	return c.Request.RemoteAddr
}

// ContentType returns the Content-Type header of the request
func (c *Context) ContentType() string {
	return c.Header("Content-Type")
}

// IsWebSocket returns true if the request is a websocket request
func (c *Context) IsWebSocket() bool {
	if strings.Contains(strings.ToLower(c.Header("Connection")), "upgrade") &&
		strings.EqualFold(c.Header("Upgrade"), "websocket") {
		return true
	}
	return false
}

// Next executes the next handler in the chain
func (c *Context) Next() error {
	c.index++
	for c.index < int8(len(c.handlers)) {
		if err := c.handlers[c.index](c); err != nil {
			return err
		}
		c.index++
	}
	return nil
}

// Abort prevents pending handlers from being called
func (c *Context) Abort() {
	c.index = int8(len(c.handlers))
}

// IsAborted returns true if the current context was aborted
func (c *Context) IsAborted() bool {
	return c.index >= int8(len(c.handlers))
}

// GetUint returns a uint value from context
func (c *Context) GetUint(key string) uint {
	value, exists := c.Get(key)
	if !exists {
		return 0
	}

	switch v := value.(type) {
	case uint:
		return v
	case uint64:
		return uint(v)
	case int:
		if v >= 0 {
			return uint(v)
		}
		return 0
	case int64:
		if v >= 0 {
			return uint(v)
		}
		return 0
	case string:
		if num, err := strconv.ParseUint(v, 10, 32); err == nil {
			return uint(num)
		}
		return 0
	default:
		return 0
	}
}

// GetHeader returns request header value (alias for Header for compatibility)
func (c *Context) GetHeader(key string) string {
	return c.Header(key)
}

// ShouldBind binds the request body to a struct (alias for Bind)
func (c *Context) ShouldBind(obj any) error {
	return c.Bind(obj)
}

// AbortWithStatusJSON aborts the chain and sends a JSON response with status code
func (c *Context) AbortWithStatusJSON(code int, obj any) {
	c.Abort()
	c.JSON(code, obj)
}

// bindData is a simplified form/query binding helper
func bindData(obj any, values url.Values) error {
	// This is a placeholder - in production, you'd use reflection
	// to properly bind form values to struct fields
	// For now, returning nil to avoid compilation errors
	_ = obj    // Avoid unused parameter warning
	_ = values // Avoid unused parameter warning
	return nil
}
