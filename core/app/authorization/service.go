package authorization

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// AuthorizationService handles business logic for authorization
type AuthorizationService struct {
	DB *gorm.DB
}

// NewAuthorizationService creates a new authorization service
func NewAuthorizationService(db *gorm.DB) *AuthorizationService {
	return &AuthorizationService{
		DB: db,
	}
}

// GetRoles returns all roles for an organization
func (s *AuthorizationService) GetRoles(organizationId uint64) ([]Role, error) {
	// If organizationId is not 0, fetch both system roles (organization_id=0) and org-specific roles
	// If organizationId is 0, just fetch system roles
	var roles []Role
	var result *gorm.DB

	if organizationId != 0 {
		// Fetch both system roles (organization_id=0) and organization-specific roles
		result = s.DB.Where("organization_id = ? OR organization_id = 0", organizationId).Find(&roles)
	} else {
		// Just fetch system roles
		result = s.DB.Where("organization_id = 0").Find(&roles)
	}

	if result.Error != nil {
		return nil, result.Error
	}

	// For each role, count its permissions
	for i := range roles {
		// Count permissions for this role
		var count int64
		if err := s.DB.Model(&RolePermission{}).
			Where("role_id = ?", roles[i].Id).
			Count(&count).Error; err != nil {
			// Log the error but continue
			fmt.Printf("Error counting permissions for role %d: %v\n", roles[i].Id, err)
		}

		// Set the permission count
		roles[i].PermissionCount = int(count)
	}
	return roles, nil
}

// GetRole returns a role by Id
func (s *AuthorizationService) GetRole(id uint64) (*Role, error) {
	var role Role
	result := s.DB.First(&role, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, result.Error
	}

	return &role, nil
}

// CreateRole creates a new role
func (s *AuthorizationService) CreateRole(role *Role) error {
	// Set creation time
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	result := s.DB.Create(role)
	return result.Error
}

// UpdateRole updates an existing role
func (s *AuthorizationService) UpdateRole(role *Role) error {
	var existingRole Role
	result := s.DB.First(&existingRole, "id = ?", role.Id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return result.Error
	}

	// Cannot modify system roles
	if existingRole.IsSystem {
		return ErrSystemRoleUnmodifiable
	}

	// Update fields
	existingRole.Name = role.Name
	existingRole.Description = role.Description
	existingRole.UpdatedAt = time.Now()

	result = s.DB.Save(&existingRole)
	if result.Error != nil {
		return result.Error
	}

	// Update the role object with saved data
	*role = existingRole

	return nil
}

// DeleteRole deletes a role
func (s *AuthorizationService) DeleteRole(id uint64) error {
	var existingRole Role
	result := s.DB.First(&existingRole, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return result.Error
	}

	// Cannot delete system roles
	if existingRole.IsSystem {
		return ErrSystemRoleUnmodifiable
	}

	// First delete associated role permissions
	if err := s.DB.Where("role_id = ?", id).Delete(&RolePermission{}).Error; err != nil {
		return err
	}

	// Then delete the role
	result = s.DB.Delete(&existingRole)
	return result.Error
}

// GetRolePermissions returns all permissions for a role
func (s *AuthorizationService) GetRolePermissions(roleId uint64) ([]Permission, error) {
	// Convert string Id to uint

	// Check if role exists
	var role Role
	result := s.DB.First(&role, "id = ?", roleId)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, result.Error
	}

	// Get permissions
	var permissions []Permission
	err := s.DB.Raw(`
		SELECT p.* FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = ?
	`, roleId).Scan(&permissions).Error

	if err != nil {
		return nil, err
	}

	return permissions, nil
}

// AssignPermissionToRole assigns a permission to a role
func (s *AuthorizationService) AssignPermissionToRole(roleId uint64, permissionId uint64) error {

	// Check if role exists
	var role Role
	result := s.DB.First(&role, "id = ?", roleId)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return result.Error
	}

	// Check if permission exists
	var permission Permission
	result = s.DB.First(&permission, "id = ?", permissionId)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrPermissionNotFound
		}
		return result.Error
	}

	// Check if permission is already assigned
	var count int64
	s.DB.Model(&RolePermission{}).
		Where("role_id = ? AND permission_id = ?", roleId, permissionId).
		Count(&count)

	if count > 0 {
		return ErrDuplicatePermission
	}

	// Create role permission
	rolePermission := RolePermission{
		RoleId:       uint(roleId),
		PermissionId: uint(permissionId),
		CreatedAt:    time.Now(),
	}

	result = s.DB.Create(&rolePermission)
	return result.Error
}

// RevokePermissionFromRole removes a permission from a role
func (s *AuthorizationService) RevokePermissionFromRole(roleId uint64, permissionId uint64) error {
	// Check if role exists
	var role Role
	result := s.DB.First(&role, "id = ?", roleId)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return result.Error
	}

	// Check if permission exists
	var permission Permission
	result = s.DB.First(&permission, "id = ?", permissionId)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrPermissionNotFound
		}
		return result.Error
	}

	// Delete role permission
	result = s.DB.Where("role_id = ? AND permission_id = ?", roleId, permissionId).
		Delete(&RolePermission{})

	return result.Error
}

// CreateResourcePermission creates a resource-specific permission
func (s *AuthorizationService) CreateResourcePermission(rp *ResourcePermission) error {
	// Set creation time
	rp.CreatedAt = time.Now()
	rp.UpdatedAt = time.Now()

	result := s.DB.Create(rp)
	return result.Error
}

// DeleteResourcePermission deletes a resource-specific permission
func (s *AuthorizationService) DeleteResourcePermission(id uint64) error {
	result := s.DB.Delete(&ResourcePermission{}, "id = ?", id)
	return result.Error
}

// HasPermission checks if a user has permission for a resource type
func (s *AuthorizationService) HasPermission(userId uint64, orgId uint64, resourceType, action string) (bool, error) {
	// Skip organization check if orgId is 0 (indicates a global endpoint)
	if orgId == 0 {
		return true, nil
	}

	// Get the organization member record
	var memberId uint
	var roleId string
	var isOwnerFlag bool
	var department string
	var membershipType string

	memberErr := s.DB.Raw(`
		SELECT id, role_id, is_owner, COALESCE(department, '') as department, 
		COALESCE(membership_type, 'Internal') as membership_type 
		FROM organization_members
		WHERE user_id = ? AND organization_id = ?
	`, userId, orgId).Row().Scan(&memberId, &roleId, &isOwnerFlag, &department, &membershipType)

	if memberErr != nil {
		return false, ErrUserNotAuthorized
	}

	// STEP 1: Check if the user is marked as owner in the organization_members table
	if isOwnerFlag {
		return true, nil
	}

	// STEP 2: Check if the user has the Owner role for this organization
	var isOwnerRole int64
	ownerErr := s.DB.Raw(`
		SELECT COUNT(*) FROM organization_members om
		JOIN roles r ON CAST(om.role_id AS UNSIGNED) = r.id
		WHERE om.user_id = ?
		AND om.organization_id = ?
		AND r.name = 'Owner'
	`, userId, orgId).Count(&isOwnerRole).Error

	if ownerErr != nil {
		return false, ownerErr
	}

	// If the user has an Owner role, automatically grant all permissions
	if isOwnerRole > 0 {
		return true, nil
	}

	// STEP 3: Check for specific ResourceAccess entries for this member
	var resourceAccessCount int64
	s.DB.Model(&ResourceAccess{}).
		Where("member_id = ? AND resource_type = ?", memberId, resourceType).
		Count(&resourceAccessCount)

	if resourceAccessCount > 0 {
		// Found specific access rules for this member, so we'll use them
		// Check if any of the resource access entries allow the requested action
		var actionAllowed int64
		resourceAccessErr := s.DB.Raw(`
			SELECT COUNT(*) FROM resource_access
			WHERE member_id = ?
			AND resource_type = ?
			AND access_type IN ('all', 'read_write')
		`, memberId, resourceType).Count(&actionAllowed).Error

		if resourceAccessErr != nil {
			return false, resourceAccessErr
		}

		if actionAllowed > 0 {
			return true, nil
		}

		// For more specific action checking
		if action == "read" {
			// Check if the user has any access type that allows reading
			var readAllowed int64
			s.DB.Raw(`
				SELECT COUNT(*) FROM resource_access
				WHERE member_id = ?
				AND resource_type = ?
				AND access_type IN ('read_only', 'all', 'read_write')
			`, memberId, resourceType).Count(&readAllowed)

			if readAllowed > 0 {
				return true, nil
			}
		}

		// If we get here, the specific resource access rules don't grant this permission
	}

	// STEP 4: Check for role-based resource permissions
	if roleId != "" {
		var rolePermCount int64
		s.DB.Model(&ResourcePermission{}).
			Where("role_id = ? AND resource_type = ? AND action = ?", roleId, resourceType, action).
			Count(&rolePermCount)

		if rolePermCount > 0 {
			return true, nil
		}
	}

	// STEP 5: Fall back to the legacy permission system
	var count int64
	err := s.DB.Raw(`
		SELECT COUNT(*) FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		JOIN organization_members om ON CAST(om.role_id AS UNSIGNED) = rp.role_id
		WHERE om.user_id = ?
		AND om.organization_id = ?
		AND p.resource_type = ?
		AND p.action = ?
	`, userId, orgId, resourceType, action).Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// HasResourcePermission checks if a user has permission for a specific resource
func (s *AuthorizationService) HasResourcePermission(userId uint64, orgId uint64, resourceType, resourceId, action string) (bool, error) {
	// Skip organization check if orgId is 0 (indicates a global endpoint)
	if orgId == 0 {
		return true, nil
	}

	// STEP 1: Check if the user has the Owner role for this organization
	var isOwner int64
	ownerErr := s.DB.Raw(`
		SELECT COUNT(*) FROM organization_members om
		JOIN roles r ON CAST(om.role_id AS UNSIGNED) = r.id
		WHERE om.user_id = ?
		AND om.organization_id = ?
		AND r.name = 'Owner'
	`, userId, orgId).Count(&isOwner).Error

	if ownerErr != nil {
		return false, ownerErr
	}

	// If the user is an Owner, automatically grant all permissions
	if isOwner > 0 {
		return true, nil
	}

	// STEP 2: Check if the user has general permission for this resource type
	hasGeneralPermission, err := s.HasPermission(userId, orgId, resourceType, action)
	if err != nil {
		return false, err
	}

	// If user has general permission, no need to check resource-specific permissions
	if hasGeneralPermission {
		return true, nil
	}

	// STEP 3: Check resource-specific permission
	var count int64
	err = s.DB.Raw(`
		SELECT COUNT(*) FROM resource_permissions rp 
		WHERE rp.user_id = ? 
		AND rp.organization_id = ? 
		AND rp.resource_type = ? 
		AND rp.resource_id = ? 
		AND rp.action = ?
	`, userId, orgId, resourceType, resourceId, action).Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetUserPermissions returns all permissions for a user across all organizations
func (s *AuthorizationService) GetUserPermissions(userId string) ([]Permission, error) {
	// Convert string Id to uint
	userIdUint, err := strconv.ParseUint(userId, 10, 32)
	if err != nil {
		fmt.Printf("GetUserPermissions: Invalid user Id format: %s, error: %v\n", userId, err)
		return nil, ErrInvalidId
	}

	fmt.Printf("GetUserPermissions: Getting permissions for user Id: %d\n", userIdUint)

	// Get permissions from role-based permissions
	var permissions []Permission
	err = s.DB.Raw(`
		SELECT DISTINCT p.* FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN organization_members om ON om.role_id = rp.role_id
		WHERE om.user_id = ?
	`, uint(userIdUint)).Scan(&permissions).Error

	if err != nil {
		fmt.Printf("GetUserPermissions: Error getting role-based permissions: %v\n", err)
		return nil, err
	}

	fmt.Printf("GetUserPermissions: Found %d role-based permissions\n", len(permissions))

	// Get permissions from resource-specific permissions
	var resourcePermissions []Permission
	err = s.DB.Raw(`
		SELECT DISTINCT p.* FROM permissions p
		JOIN resource_permissions rp ON p.id = rp.permission_id
		WHERE rp.user_id = ?
	`, uint(userIdUint)).Scan(&resourcePermissions).Error

	if err != nil {
		fmt.Printf("GetUserPermissions: Error getting resource-specific permissions: %v\n", err)
		return nil, err
	}

	fmt.Printf("GetUserPermissions: Found %d resource-specific permissions\n", len(resourcePermissions))

	// Merge the two sets of permissions
	// Create a map to avoid duplicates
	permMap := make(map[uint]Permission)
	for _, p := range permissions {
		permMap[p.Id] = p
	}

	for _, p := range resourcePermissions {
		permMap[p.Id] = p
	}

	// Convert map back to slice
	result := make([]Permission, 0, len(permMap))
	for _, p := range permMap {
		result = append(result, p)
	}

	fmt.Printf("GetUserPermissions: Returning %d total permissions\n", len(result))
	return result, nil
}

// SeedPermissions creates default permissions if they don't exist
func (s *AuthorizationService) SeedPermissions() error {
	// Define resource types and actions (aligned with module seeding) Only for system roles and core modules
	resourceTypes := []string{
		"user", "authorization", "media", "profile",
	}
	actions := []string{"create", "read", "update", "delete", "list"}

	// Create permissions for each resource type and action
	for _, resourceType := range resourceTypes {
		for _, action := range actions {
			var permission Permission

			// Check if permission already exists
			result := s.DB.Where("resource_type = ? AND action = ?", resourceType, action).First(&permission)
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// Create permission
				permission = Permission{
					Name:         action + " " + resourceType,
					Description:  "Permission to " + action + " " + resourceType,
					ResourceType: resourceType,
					Action:       action,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}

				if err := s.DB.Create(&permission).Error; err != nil {
					return err
				}
			} else if result.Error != nil {
				return result.Error
			}
		}
	}

	return nil
}

// SeedRoles creates default roles if they don't exist
func (s *AuthorizationService) SeedRoles() error {
	// Define default roles
	defaultRoles := []Role{
		{
			Name:        "Owner",
			Description: "Full access to all resources",
			IsSystem:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "Administrator",
			Description: "Administrative access with some limitations",
			IsSystem:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "Member",
			Description: "Standard member with limited access",
			IsSystem:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "External",
			Description: "External user with minimal access",
			IsSystem:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Create roles if they don't exist
	for _, role := range defaultRoles {
		var existingRole Role
		result := s.DB.First(&existingRole, "name = ? AND is_system = ?", role.Name, role.IsSystem)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if err := s.DB.Create(&role).Error; err != nil {
				return err
			}
		} else if result.Error != nil {
			return result.Error
		}
	}

	return nil
}

// SetupRolePermissions assigns default permissions to system roles
func (s *AuthorizationService) SetupRolePermissions() error {
	// First seed permissions
	if err := s.SeedPermissions(); err != nil {
		return err
	}

	// Then seed roles
	if err := s.SeedRoles(); err != nil {
		return err
	}

	// Get all permissions
	var permissions []Permission
	if err := s.DB.Find(&permissions).Error; err != nil {
		return err
	}

	// Get the owner role
	var ownerRole Role
	if err := s.DB.Where("name = ? AND is_system = ?", "Owner", true).First(&ownerRole).Error; err != nil {
		return err
	}

	// Assign all permissions to the owner role
	for _, permission := range permissions {
		// Skip organization:delete for non-system roles
		if permission.ResourceType == "organization" && permission.Action == "delete" {
			continue
		}

		// Check if permission is already assigned
		var count int64
		s.DB.Model(&RolePermission{}).
			Where("role_id = ? AND permission_id = ?", ownerRole.Id, permission.Id).
			Count(&count)

		if count == 0 {
			rolePermission := RolePermission{
				RoleId:       ownerRole.Id,
				PermissionId: permission.Id,
				CreatedAt:    time.Now(),
			}

			if err := s.DB.Create(&rolePermission).Error; err != nil {
				return err
			}
		}
	}

	// Get the admin role
	var adminRole Role
	if err := s.DB.Where("name = ? AND is_system = ?", "Administrator", true).First(&adminRole).Error; err != nil {
		return err
	}

	// Define admin permissions
	adminPermissionTypes := map[string][]string{
		"user":    {"create", "read", "update", "delete", "list"},
		"media":   {"create", "read", "update", "delete", "list"},
		"profile": {"create", "read", "update", "delete", "list"},
	}

	// Assign admin permissions
	for resourceType, actions := range adminPermissionTypes {
		for _, action := range actions {
			var permission Permission
			if err := s.DB.Where("resource_type = ? AND action = ?", resourceType, action).First(&permission).Error; err != nil {
				continue // Skip if permission not found
			}

			// Check if permission is already assigned
			var count int64
			s.DB.Model(&RolePermission{}).
				Where("role_id = ? AND permission_id = ?", adminRole.Id, permission.Id).
				Count(&count)

			if count == 0 {
				rolePermission := RolePermission{
					RoleId:       adminRole.Id,
					PermissionId: permission.Id,
					CreatedAt:    time.Now(),
				}

				if err := s.DB.Create(&rolePermission).Error; err != nil {
					return err
				}
			}
		}
	}

	// Get the member role
	var memberRole Role
	if err := s.DB.Where("name = ? AND is_system = ?", "Member", true).First(&memberRole).Error; err != nil {
		return err
	}

	// Define member permissions
	memberPermissionTypes := map[string][]string{
		"user":    {"read", "list"},
		"media":   {"read", "list"},
		"profile": {"read", "list"},
	}

	// Assign member permissions
	for resourceType, actions := range memberPermissionTypes {
		for _, action := range actions {
			var permission Permission
			if err := s.DB.Where("resource_type = ? AND action = ?", resourceType, action).First(&permission).Error; err != nil {
				continue // Skip if permission not found
			}

			// Check if permission is already assigned
			var count int64
			s.DB.Model(&RolePermission{}).
				Where("role_id = ? AND permission_id = ?", memberRole.Id, permission.Id).
				Count(&count)

			if count == 0 {
				rolePermission := RolePermission{
					RoleId:       memberRole.Id,
					PermissionId: permission.Id,
					CreatedAt:    time.Now(),
				}

				if err := s.DB.Create(&rolePermission).Error; err != nil {
					return err
				}
			}
		}
	}

	// Get the external role
	var externalRole Role
	if err := s.DB.Where("name = ? AND is_system = ?", "External", true).First(&externalRole).Error; err != nil {
		return err
	}

	// Define external permissions
	externalPermissionTypes := map[string][]string{
		"user":    {"read", "list"},
		"media":   {"read", "list"},
		"profile": {"read"},
	}

	// Assign external permissions
	for resourceType, actions := range externalPermissionTypes {
		for _, action := range actions {
			var permission Permission
			if err := s.DB.Where("resource_type = ? AND action = ?", resourceType, action).First(&permission).Error; err != nil {
				continue // Skip if permission not found
			}

			// Check if permission is already assigned
			var count int64
			s.DB.Model(&RolePermission{}).
				Where("role_id = ? AND permission_id = ?", externalRole.Id, permission.Id).
				Count(&count)

			if count == 0 {
				rolePermission := RolePermission{
					RoleId:       externalRole.Id,
					PermissionId: permission.Id,
					CreatedAt:    time.Now(),
				}

				if err := s.DB.Create(&rolePermission).Error; err != nil {
					return err
				}
			}
		}
	}

	return nil
}
