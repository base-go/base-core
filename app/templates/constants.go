package templates

// Template names as constants for type safety
const (
	// Layouts
	LayoutDefault = "app.html"
	LayoutLanding = "landing.html"
	LayoutAuth    = "auth.html"

	// Common pages
	PageError   = "error.html"
	PageLanding = "landing.html"
	// Auth pages
	PageRegister       = "auth/register.html"
	PageLogin          = "auth/login.html"
	PageLogout         = "auth/logout.html"
	PageForgotPassword = "auth/forgot-password.html"
	PageResetPassword  = "auth/reset-password.html"

	// Post templates
	PostIndex = "posts/index.html"
	PostShow  = "posts/show.html"
	PostEdit  = "posts/edit.html"
	PostNew   = "posts/new.html"
)
