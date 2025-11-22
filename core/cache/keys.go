package cache

import "fmt"

// Key generation helpers for consistent cache key formatting

// UserKey generates a cache key for user data
func UserKey(userID uint) string {
	return fmt.Sprintf("user:%d", userID)
}

// UserEmailKey generates a cache key for user lookup by email
func UserEmailKey(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}

// UserPermissionsKey generates a cache key for user permissions
func UserPermissionsKey(userID uint) string {
	return fmt.Sprintf("user:permissions:%d", userID)
}

// RolePermissionsKey generates a cache key for role permissions
func RolePermissionsKey(roleID uint) string {
	return fmt.Sprintf("role:permissions:%d", roleID)
}

// UserResourcePermissionKey generates a cache key for resource-specific permission
func UserResourcePermissionKey(userID uint, resource, action string) string {
	return fmt.Sprintf("user:resource_permission:%d:%s:%s", userID, resource, action)
}

// JWTValidationKey generates a cache key for JWT validation results
func JWTValidationKey(tokenHash string) string {
	return fmt.Sprintf("jwt:valid:%s", tokenHash)
}

// UserPattern generates a pattern to match all user-related keys
func UserPattern(userID uint) string {
	return fmt.Sprintf("user:*:%d*", userID)
}

// RolePattern generates a pattern to match all role-related keys
func RolePattern(roleID uint) string {
	return fmt.Sprintf("role:*:%d*", roleID)
}

// HTTPResponseKey generates a cache key for HTTP response caching
func HTTPResponseKey(method, path, query string) string {
	if query != "" {
		return fmt.Sprintf("http:%s:%s?%s", method, path, query)
	}
	return fmt.Sprintf("http:%s:%s", method, path)
}
