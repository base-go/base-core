package authorization_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"base/core/app/authorization"
	"base/core/app/profile"
	"base/core/cache"
	"base/core/router"
	"base/core/storage"
	"base/internal/testutil"

	"gorm.io/gorm"
)

type fixture struct {
	db     *gorm.DB
	router *router.Router
	users  map[string]uint // role name -> user id
}

// newFixture builds the authorization module the way the app does, seeds it,
// creates one user per seeded role, and mounts every check on a route.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	db := testutil.NewDB(t)
	mod := authorization.NewAuthorizationModule(db, nil, testutil.Logger(), cache.NewNoopCache())
	if err := mod.Migrate(); err != nil {
		t.Fatalf("migrate authorization: %v", err)
	}
	if err := db.AutoMigrate(&storage.Attachment{}, &profile.User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}

	f := &fixture{db: db, users: map[string]uint{}}
	for _, name := range []string{"Owner", "Administrator", "Member", "Viewer"} {
		var role authorization.Role
		if err := db.Where("name = ?", name).First(&role).Error; err != nil {
			t.Fatalf("seeded role %s: %v", name, err)
		}
		u := profile.User{
			FirstName: name, LastName: "User", Username: name, Email: name + "@example.com",
			Phone: name, RoleId: role.Id,
		}
		if err := db.Create(&u).Error; err != nil {
			t.Fatalf("create user %s: %v", name, err)
		}
		f.users[name] = u.Id
	}

	// signIn stands in for the auth middleware: it puts the user id where it does.
	signIn := func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			if id := c.Header("X-Test-User"); id != "" {
				n, _ := strconv.ParseUint(id, 10, 64)
				c.Set("user_id", uint(n))
			}
			return next(c)
		}
	}
	ok := func(c *router.Context) error { return c.JSON(http.StatusOK, map[string]string{"ok": "yes"}) }

	r := router.New()
	r.Use(signIn)
	r.GET("/can", ok, authorization.Can("Create", "Media"))
	r.GET("/read", ok, authorization.Can("read", "media"))
	r.GET("/access/:id", ok, authorization.CanAccess("update", "media", "id"))
	r.GET("/role/owner", ok, authorization.HasRole("Owner"))
	r.GET("/role/viewer", ok, authorization.HasRole("Viewer"))
	r.GET("/any", ok, authorization.CanAny([]string{"create:media", "delete:media"}))
	r.GET("/all", ok, authorization.CanAll([]string{"read:media", "create:media"}))
	r.GET("/authmw", ok, authorization.AuthMiddleware("media", "create"))
	r.GET("/require/owner", ok, authorization.RequireRole("Owner"))
	f.router = r
	return f
}

func (f *fixture) get(t *testing.T, path, role string) int {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if role != "" {
		req.Header.Set("X-Test-User", strconv.FormatUint(uint64(f.users[role]), 10))
	}
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	return w.Code
}

func TestChecksDenyAUserWhoseRoleLacksThePermission(t *testing.T) {
	f := newFixture(t)
	for _, path := range []string{"/can", "/access/7", "/role/owner", "/any", "/all", "/authmw", "/require/owner"} {
		if code := f.get(t, path, "Viewer"); code != http.StatusForbidden {
			t.Errorf("Viewer GET %s: got %d, want 403", path, code)
		}
	}
}

func TestChecksAllowAUserWhoseRoleHasThePermission(t *testing.T) {
	f := newFixture(t)
	for _, path := range []string{"/can", "/read", "/access/7", "/role/owner", "/any", "/all", "/authmw", "/require/owner"} {
		if code := f.get(t, path, "Owner"); code != http.StatusOK {
			t.Errorf("Owner GET %s: got %d, want 200", path, code)
		}
	}
	if code := f.get(t, "/read", "Viewer"); code != http.StatusOK {
		t.Errorf("Viewer GET /read: got %d, want 200", code)
	}
}

func TestHasRoleChecksTheRoleItIsGiven(t *testing.T) {
	f := newFixture(t)
	if code := f.get(t, "/role/viewer", "Viewer"); code != http.StatusOK {
		t.Errorf("Viewer GET /role/viewer: got %d, want 200", code)
	}
	if code := f.get(t, "/role/viewer", "Owner"); code != http.StatusForbidden {
		t.Errorf("Owner GET /role/viewer: got %d, want 403", code)
	}
}

func TestResourcePermissionGrantsOneResource(t *testing.T) {
	f := newFixture(t)
	grant := authorization.ResourcePermission{
		ResourceType: "media", ResourceId: "7", UserId: f.users["Viewer"], Action: "update",
	}
	if err := f.db.Create(&grant).Error; err != nil {
		t.Fatalf("create resource permission: %v", err)
	}
	if code := f.get(t, "/access/7", "Viewer"); code != http.StatusOK {
		t.Errorf("Viewer GET /access/7 with a grant: got %d, want 200", code)
	}
	if code := f.get(t, "/access/8", "Viewer"); code != http.StatusForbidden {
		t.Errorf("Viewer GET /access/8 without a grant: got %d, want 403", code)
	}
}

func TestNoCheckAnswers500ForASignedInUser(t *testing.T) {
	f := newFixture(t)
	paths := []string{"/can", "/read", "/access/7", "/role/owner", "/role/viewer", "/any", "/all", "/authmw", "/require/owner"}
	for role := range f.users {
		for _, path := range paths {
			if code := f.get(t, path, role); code >= 500 {
				t.Errorf("%s GET %s: got %d", role, path, code)
			}
		}
	}
}

func TestChecksAnswer401WithoutAUser(t *testing.T) {
	f := newFixture(t)
	if code := f.get(t, "/can", ""); code != http.StatusUnauthorized {
		t.Errorf("anonymous GET /can: got %d, want 401", code)
	}
}

func TestOrganizationIdIsReadFromTheCORSAllowedHeader(t *testing.T) {
	r := router.New()
	var got uint64
	r.GET("/org", func(c *router.Context) error {
		id, err := authorization.GetOrganizationIdFromContext(c)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		got = id
		return c.NoContent()
	})
	req := httptest.NewRequest(http.MethodGet, "/org", nil)
	req.Header.Set("Base-Orgid", "42")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent || got != 42 {
		t.Fatalf("Base-Orgid: 42 gave status %d and id %d, want 204 and 42", w.Code, got)
	}
}
