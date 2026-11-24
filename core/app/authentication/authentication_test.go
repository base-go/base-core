package authentication_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"base/core/app/authentication"
	"base/core/app/profile"
	"base/core/emitter"
	"base/core/router"
	"base/core/storage"
	"base/internal/testutil"

	"gorm.io/gorm"
)

func newAuthRouter(t *testing.T) (*router.Router, *gorm.DB) {
	t.Helper()
	db := testutil.NewDB(t)
	if err := db.AutoMigrate(&storage.Attachment{}); err != nil {
		t.Fatalf("migrate attachments: %v", err)
	}
	mod := authentication.NewAuthenticationModule(db, nil, nil, testutil.Logger(), emitter.New())
	if err := mod.Migrate(); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	r := router.New()
	mod.Routes(r.Group("/api"))
	return r, db
}

func postJSON(r http.Handler, path, body string) (int, map[string]any) {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func createUserWithResetToken(t *testing.T, db *gorm.DB, token string, expiry time.Time) {
	t.Helper()
	u := authentication.AuthUser{
		User: profile.User{
			FirstName: "Ada", LastName: "Lovelace", Username: "ada", Phone: "1",
			Email: "ada@example.com", Password: "x",
		},
		ResetToken:       token,
		ResetTokenExpiry: &expiry,
	}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
}

func TestResetPasswordWithAWrongCodeAnswers400(t *testing.T) {
	r, db := newAuthRouter(t)
	createUserWithResetToken(t, db, "123456", time.Now().Add(10*time.Minute))

	code, body := postJSON(r, "/api/auth/reset-password",
		`{"email":"ada@example.com","token":"999999","new_password":"newpassword"}`)
	if code != http.StatusBadRequest {
		t.Fatalf("wrong code: got %d %v, want 400", code, body)
	}
}

func TestResetPasswordWithAnExpiredCodeAnswers400(t *testing.T) {
	r, db := newAuthRouter(t)
	createUserWithResetToken(t, db, "123456", time.Now().Add(-time.Minute))

	code, body := postJSON(r, "/api/auth/reset-password",
		`{"email":"ada@example.com","token":"123456","new_password":"newpassword"}`)
	if code != http.StatusBadRequest {
		t.Fatalf("expired code: got %d %v, want 400", code, body)
	}
}

func TestResetPasswordForAnUnknownEmailAnswers404(t *testing.T) {
	r, _ := newAuthRouter(t)
	code, body := postJSON(r, "/api/auth/reset-password",
		`{"email":"nobody@example.com","token":"123456","new_password":"newpassword"}`)
	if code != http.StatusNotFound {
		t.Fatalf("unknown email: got %d %v, want 404", code, body)
	}
}
