package authorization

import (
	"fmt"
)

// CreateDefaultOrganizationRoles creates default roles and permissions for a new organization
func (s *AuthorizationService) CreateDefaultOrganizationRoles(organizationId uint) error {
	// Define organization-specific roles
	orgRoles := []Role{
		{
			Name:           "Owner",
			Description:    "Full access to all organization resources",
			OrganizationId: organizationId,
			IsSystem:       false,
		},
		{
			Name:           "Administrator",
			Description:    "Administrative access with some limitations",
			OrganizationId: organizationId,
			IsSystem:       false,
		},
		{
			Name:           "Member",
			Description:    "Standard member with limited access",
			OrganizationId: organizationId,
			IsSystem:       false,
		},
		{
			Name:           "Viewer",
			Description:    "Read-only access to resources",
			OrganizationId: organizationId,
			IsSystem:       false,
		},
	}

	// Start transaction
	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Create roles for this organization
	var createdRoles = make(map[string]uint)
	for _, role := range orgRoles {
		if err := tx.Create(&role).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create role %s: %w", role.Name, err)
		}
		createdRoles[role.Name] = role.Id
		// Role created successfully
	}

	// Get all permissions from the system
	var allPermissions []Permission
	if err := tx.Find(&allPermissions).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to fetch permissions: %w", err)
	}

	// Assign all permissions to Owner role
	ownerRoleId, ok := createdRoles["Owner"]
	if !ok {
		tx.Rollback()
		return fmt.Errorf("owner role not found for organization %d", organizationId)
	}

	for _, permission := range allPermissions {
		rolePermission := RolePermission{
			RoleId:       ownerRoleId,
			PermissionId: permission.Id,
		}
		if err := tx.Create(&rolePermission).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to assign permission %s to Owner role: %w", permission.Name, err)
		}
	}

	// Assign appropriate permissions to Admin role
	adminRoleId, ok := createdRoles["Administrator"]
	if ok {
		adminPermissions := []string{
			"project:create", "project:read", "project:update", "project:delete", "project:list", "project:manage_members",
			"client:create", "client:read", "client:update", "client:delete", "client:list",
			"employee:create", "employee:read", "employee:update", "employee:delete", "employee:list",
			"invitation:create", "invitation:read", "invitation:update", "invitation:delete", "invitation:list",
			"organization_member:read", "organization_member:list", "organization_member:create", "organization_member:update", "organization_member:delete",
			"organization:read", "organization:update",
			"idea:create", "idea:read", "idea:update", "idea:delete", "idea:list",
			"scope:create", "scope:read", "scope:update", "scope:delete", "scope:list",
			"scope_version:create", "scope_version:read", "scope_version:update", "scope_version:delete", "scope_version:list",
			"absence:create", "absence:read", "absence:update", "absence:delete", "absence:list",
			"idea_group:create", "idea_group:read", "idea_group:update", "idea_group:delete", "idea_group:list",
			"notification:create", "notification:read", "notification:update", "notification:delete", "notification:list",
			"user:read", "user:list", "user:update",
		}

		for _, permName := range adminPermissions {
			for _, permission := range allPermissions {
				permString := permission.ResourceType + ":" + permission.Action
				if permString == permName {
					rolePermission := RolePermission{
						RoleId:       adminRoleId,
						PermissionId: permission.Id,
					}
					if err := tx.Create(&rolePermission).Error; err != nil {
						tx.Rollback()
						return fmt.Errorf("failed to assign permission %s to Admin role: %w", permission.Name, err)
					}
					break
				}
			}
		}
	}

	// Assign appropriate permissions to Member role
	memberRoleId, ok := createdRoles["Member"]
	if ok {
		memberPermissions := []string{
			"project:read", "project:list",
			"client:read", "client:list",
			"employee:read", "employee:list",
			"organization_member:read", "organization_member:list",
			"organization:read",
			"idea:create", "idea:read", "idea:update", "idea:list",
			"scope:create", "scope:read", "scope:update", "scope:list",
			"scope_version:create", "scope_version:read", "scope_version:update", "scope_version:list",
			"absence:create", "absence:read", "absence:update", "absence:list",
			"idea_group:create", "idea_group:read", "idea_group:update", "idea_group:list",
			"notification:read", "notification:list",
			"user:read", "user:list",
		}

		for _, permName := range memberPermissions {
			for _, permission := range allPermissions {
				permString := permission.ResourceType + ":" + permission.Action
				if permString == permName {
					rolePermission := RolePermission{
						RoleId:       memberRoleId,
						PermissionId: permission.Id,
					}
					if err := tx.Create(&rolePermission).Error; err != nil {
						tx.Rollback()
						return fmt.Errorf("failed to assign permission %s to Member role: %w", permission.Name, err)
					}
					break
				}
			}
		}
	}

	// Assign appropriate permissions to Viewer role
	viewerRoleId, ok := createdRoles["Viewer"]
	if ok {
		viewerPermissions := []string{
			"project:read", "project:list",
			"client:read", "client:list",
			"employee:read", "employee:list",
			"organization_member:read", "organization_member:list",
			"organization:read",
			"idea:read", "idea:list",
			"scope:read", "scope:list",
			"scope_version:read", "scope_version:list",
			"absence:read", "absence:list",
			"idea_group:read", "idea_group:list",
			"notification:read", "notification:list",
			"user:read", "user:list",
		}

		for _, permName := range viewerPermissions {
			for _, permission := range allPermissions {
				permString := permission.ResourceType + ":" + permission.Action
				if permString == permName {
					rolePermission := RolePermission{
						RoleId:       viewerRoleId,
						PermissionId: permission.Id,
					}
					if err := tx.Create(&rolePermission).Error; err != nil {
						tx.Rollback()
						return fmt.Errorf("failed to assign permission %s to Viewer role: %w", permission.Name, err)
					}
					break
				}
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Successfully created default roles and permissions for organization
	return nil
}
