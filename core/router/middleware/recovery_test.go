package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"base/core/config"
	"base/core/router"
)

func TestRecoveryWithoutALoggerAnswers500AndServesTheNextRequest(t *testing.T) {
	r := router.New()
	ApplyConfigurableMiddleware(r, &config.MiddlewareConfig{RecoveryEnabled: true})
	r.GET("/boom", func(c *router.Context) error { panic("handler failed") })
	r.GET("/ok", func(c *router.Context) error { return c.NoContent() })

	srv := httptest.NewServer(r)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/boom")
	if err != nil {
		t.Fatalf("GET /boom: %v (the connection was dropped instead of answered)", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("GET /boom: got %d, want 500", resp.StatusCode)
	}

	resp, err = http.Get(srv.URL + "/ok")
	if err != nil {
		t.Fatalf("GET /ok after the panic: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("GET /ok after the panic: got %d, want 204", resp.StatusCode)
	}
}
