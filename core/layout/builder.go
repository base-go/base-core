package layout

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// ViewBuilder provides a simple, fluent API for building views
type ViewBuilder struct {
	engine     *Engine
	template   string
	layoutName string
	data       gin.H
}

// NewViewBuilder creates a new view builder
func NewViewBuilder(engine *Engine, template string) *ViewBuilder {
	return &ViewBuilder{
		engine:     engine,
		template:   template,
		layoutName: "app.html", // Default layout
		data:       gin.H{},
	}
}

// Layout sets the layout to use
func (v *ViewBuilder) Layout(layout string) *ViewBuilder {
	v.layoutName = layout
	return v
}

// NoLayout removes layout (render template only)
func (v *ViewBuilder) NoLayout() *ViewBuilder {
	v.layoutName = ""
	return v
}

// With adds data to the view
func (v *ViewBuilder) With(key string, value any) *ViewBuilder {
	v.data[key] = value
	return v
}

// WithData merges data into the view
func (v *ViewBuilder) WithData(data gin.H) *ViewBuilder {
	for k, val := range data {
		v.data[k] = val
	}
	return v
}

// WithTitle sets the page title
func (v *ViewBuilder) WithTitle(title string) *ViewBuilder {
	v.data["title"] = title
	return v
}

// WithError adds an error message
func (v *ViewBuilder) WithError(message string) *ViewBuilder {
	v.data["error"] = message
	return v
}

// WithSuccess adds a success message
func (v *ViewBuilder) WithSuccess(message string) *ViewBuilder {
	v.data["success"] = message
	return v
}

// Render executes the view with the layout
func (v *ViewBuilder) Render(ctx *gin.Context) {
	// Add template context data from middleware if available
	if templateContext, exists := ctx.Get("templateContext"); exists {
		if contextMap, ok := templateContext.(map[string]interface{}); ok {
			for key, value := range contextMap {
				// Don't overwrite existing values set by controllers
				if _, exists := v.data[key]; !exists {
					v.data[key] = value
				}
			}
		}
	}

	// Also add session data directly for backward compatibility
	session := sessions.Default(ctx)
	// Create a map to hold session values for template access
	sessionData := make(map[string]any)

	// Add common session values
	if session.Get("logged_in") == true {
		sessionData["logged_in"] = true
		sessionData["user_id"] = session.Get("user_id")
		sessionData["username"] = session.Get("username")
		sessionData["email"] = session.Get("email")

		// Add the user object if available
		if user := session.Get("user"); user != nil {
			sessionData["user"] = user
		}
	} else {
		sessionData["logged_in"] = false
	}

	v.data["session"] = sessionData

	if v.layoutName != "" {
		v.engine.RenderWithLayout(ctx.Writer, v.template, v.layoutName, v.data, ctx)
	} else {
		v.engine.Render(ctx.Writer, v.template, v.data, ctx)
	}
}

// RedirectBuilder provides a simple API for redirects
type RedirectBuilder struct {
	url    string
	status int
}

// NewRedirectBuilder creates a new redirect builder
func NewRedirectBuilder(url string) *RedirectBuilder {
	return &RedirectBuilder{
		url:    url,
		status: http.StatusSeeOther,
	}
}

// To sets the redirect URL
func (r *RedirectBuilder) To(url string) *RedirectBuilder {
	r.url = url
	return r
}

// WithStatus sets the HTTP status code
func (r *RedirectBuilder) WithStatus(status int) *RedirectBuilder {
	r.status = status
	return r
}

// Execute performs the redirect
func (r *RedirectBuilder) Execute(ctx *gin.Context) {
	ctx.Redirect(r.status, r.url)
}

// Controller provides a simple base for controllers
type Controller struct {
	Engine     *Engine
	layoutName string
}

// NewController creates a new controller with the specified layout
func NewController(engine *Engine, layoutName string) *Controller {
	return &Controller{
		Engine:     engine,
		layoutName: layoutName,
	}
}

// View creates a ViewBuilder with the controller's default layout
func (c *Controller) View(template string) *ViewBuilder {
	return &ViewBuilder{
		engine:     c.Engine,
		template:   template,
		layoutName: c.layoutName,
		data:       gin.H{},
	}
}

// Redirect creates a RedirectBuilder
func (c *Controller) Redirect(url string) *RedirectBuilder {
	return NewRedirectBuilder(url)
}

// Layout constants for easy reference
const (
	AppLayout     = "layouts/app.html"
	AuthLayout    = "layouts/auth.html"
	LandingLayout = "layouts/landing.html"
)

// Helper functions for creating controllers with specific layouts
func NewAppController(engine *Engine) *Controller {
	return NewController(engine, AppLayout)
}

func NewAuthController(engine *Engine) *Controller {
	return NewController(engine, AuthLayout)
}

func NewLandingController(engine *Engine) *Controller {
	return NewController(engine, LandingLayout)
}
