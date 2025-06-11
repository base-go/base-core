# Changelog

## [Unreleased] - 2025-06-03

### Added
- Conditional display of user profile component or login/register buttons in the landing page navigation bar (`app/theme/default/landing.html`) based on authentication status.
- Authentication protection for post-related web routes (e.g., `/posts`, `/posts/new`, `/posts/:id/edit`) using `AuthMiddleware` in `app/posts/controller.go`. The `/posts.json` endpoint remains public.

### Changed
- Updated `AuthMiddleware` in `core/middleware/auth.go` to redirect to language-prefixed login URLs (e.g., `/en/auth/login`), resolving a 404 error during unauthenticated access to protected routes.

### Fixed
- Corrected import statements in `app/posts/controller.go` to resolve compilation and linting errors.
- Addressed server startup issues related to "address already in use" errors by ensuring previous server instances are properly terminated.

### Removed

