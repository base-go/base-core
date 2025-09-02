package app

import (
	"base/core/app/profile"
	"base/core/database"
	"base/core/module"
)

// AppModules implements module.AppModuleProvider interface
type AppModules struct{}

// GetAppModules returns the list of app modules to initialize
// This is the only function that needs to be updated when adding new app modules
func (am *AppModules) GetAppModules(deps module.Dependencies) map[string]module.Module {
	modules := make(map[string]module.Module)

	return modules
}

// NewAppModules creates a new AppModules provider
func NewAppModules() *AppModules {
	return &AppModules{}
}

/*
Extend function is called during authentication (login/register) to add custom data
to both the JWT token payload and the authentication response.

This function allows you to extend the user context with additional information
such as company IDs, roles, permissions, or any other data your application needs.

Usage from app/ directory:
1. Import required models and packages at the top of this file
2. Query the database to fetch related data for the user
3. Return a map[string]any with the additional context data

Common Use Cases:
- Multi-tenant applications: Include company_id or tenant_id
- Role-based access: Include user roles and permissions
- User preferences: Include settings or configuration data
- Organization data: Include department, team, or group information

Example Implementation:

	// Import your models at the top of the file
	// import "base/app/models"

	func Extend(user_id uint) any {
		// Basic fallback if database is not available
		if database.DB == nil {
			return map[string]any{
				"user_id": user_id,
			}
		}

		// Example: Multi-tenant application with company association
		// var company models.Company
		// if err := database.DB.Where("user_id = ?", user_id).First(&company).Error; err != nil {
		//     // If no company found, return just the user ID
		//     return map[string]any{
		//         "user_id": user_id,
		//     }
		// }
		//
		// return map[string]any{
		//     "user_id":    user_id,
		//     "company_id": company.Id,
		//     "company_name": company.Name,
		// }

		// Example: Role-based access control
		// var userRoles []models.UserRole
		// database.DB.Where("user_id = ?", user_id).Find(&userRoles)
		//
		// roles := make([]string, len(userRoles))
		// for i, role := range userRoles {
		//     roles[i] = role.RoleName
		// }
		//
		// return map[string]any{
		//     "user_id": user_id,
		//     "roles":   roles,
		// }

		// Default: Return minimal context
		return map[string]any{
			"user_id": user_id,
		}
	}

The returned data will be:
1. Added to the "extend" field in the authentication response
2. Embedded in the JWT token payload under the "extend" claim
3. Available in middleware and authorization checks without additional database queries
*/

func Extend(user_id uint) any {
	// Get database instance
	if database.DB == nil {
		return map[string]any{
			"user_id": user_id,
		}
	}

	// Get user's role from the database with role relationship preloaded
	var user profile.User
	if err := database.DB.Preload("Role").Where("id = ?", user_id).First(&user).Error; err != nil {
		// If user not found, return minimal context
		return map[string]any{
			"user_id": user_id,
		}
	}

	// Return user context with role information
	roleInfo := map[string]any{
		"id":   user.RoleId,
		"name": "",
	}

	if user.Role != nil {
		roleInfo["name"] = user.Role.Name
	}

	return map[string]any{
		"user_id": user_id,
		"role":    roleInfo,
	}
}
