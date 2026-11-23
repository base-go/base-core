package middleware

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"base/core/config"
	"base/core/router"
)

func rateLimitedRouter(requests int) *router.Router {
	r := router.New()
	ApplyConfigurableMiddleware(r, &config.MiddlewareConfig{
		RateLimitEnabled:  true,
		RateLimitRequests: requests,
		RateLimitWindow:   "1m",
	})
	r.GET("/api/things", func(c *router.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"ok": "yes"})
	})
	r.GET("/api/other", func(c *router.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"ok": "yes"})
	})
	return r
}

func get(r http.Handler, path, ip string) int {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.RemoteAddr = ip + ":1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestConditionalRateLimitAnswers429AfterLimit(t *testing.T) {
	r := rateLimitedRouter(3)

	for i := 1; i <= 3; i++ {
		if code := get(r, "/api/things", "10.0.0.1"); code != http.StatusOK {
			t.Fatalf("request %d: got %d, want 200", i, code)
		}
	}
	if code := get(r, "/api/things", "10.0.0.1"); code != http.StatusTooManyRequests {
		t.Fatalf("request 4: got %d, want 429", code)
	}
	// The limit is per client, shared across routes.
	if code := get(r, "/api/other", "10.0.0.1"); code != http.StatusTooManyRequests {
		t.Fatalf("request 5 on another route: got %d, want 429", code)
	}
	// Another client has its own bucket.
	if code := get(r, "/api/things", "10.0.0.2"); code != http.StatusOK {
		t.Fatalf("other client: got %d, want 200", code)
	}
}

func TestConditionalRateLimitDoesNotLeakGoroutines(t *testing.T) {
	r := rateLimitedRouter(1000)
	get(r, "/api/things", "10.0.0.1")
	before := runtime.NumGoroutine()

	for i := 0; i < 200; i++ {
		get(r, "/api/things", "10.0.0.1")
	}

	after := runtime.NumGoroutine()
	if after-before > 5 {
		t.Fatalf("goroutines grew from %d to %d over 200 requests", before, after)
	}
}

func TestTokenBucketStopEndsCleanupGoroutine(t *testing.T) {
	before := runtime.NumGoroutine()
	buckets := make([]*TokenBucket, 50)
	for i := range buckets {
		buckets[i] = NewTokenBucket(1, time.Minute, 1)
	}
	for _, b := range buckets {
		b.Stop()
		b.Stop() // Stop twice is safe.
	}

	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine()-before > 5 {
		if time.Now().After(deadline) {
			t.Fatalf("goroutines still %d after Stop, started at %d", runtime.NumGoroutine(), before)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
