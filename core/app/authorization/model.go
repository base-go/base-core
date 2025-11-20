package authorization

import (
	"errors"
	"time"
)

var (
	ErrRoleNotFound           = errors.New("role not found")
	ErrPermissionNotFound     = errors.New("permission not found")
	ErrInvalidPermission      = errors.New("invalid permission")
	ErrInvalidRole            = errors.New("invalid role")
	ErrUserNotAuthorized      = errors.New("user not authorized")
	ErrRolePermissionNotFound = errors.New("role permission not found")
	ErrInvalidId              = errors.New("invalid id")
	ErrInvalidRoleId          = errors.New("invalid role id")
	ErrSystemRoleUnmodifiable = errors.New("system role unmodifiable")
	ErrDuplicatePermission    = errors.New("duplicate permission")
)

// Role represents a set of permissions assigned to users within an organization
type Role struct {
	Id              uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name            string    `gorm:"not null" json:"name"`
	Description     string    `json:"description"`
	IsSystem        bool      `gorm:"default:false" json:"is_system"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	PermissionCount int       `json:"permission_count"` // New field
}

// ToResponse converts the role to a response object
func (r *Role) ToResponse() *RoleResponse {
	if r == nil {
		return nil
	}
	return &RoleResponse{
		Id:              r.Id,
		Name:            r.Name,
		Description:     r.Description,
		IsSystem:        r.IsSystem,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
		PermissionCount: r.PermissionCount,
	}
}

// RoleResponse represents the response structure for a role
type RoleResponse struct {
	Id              uint      `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	IsSystem        bool      `json:"is_system"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PermissionCount int       `json:"permission_count"` // New field
}

// CreateRoleRequest represents the payload for creating a role
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
}

// UpdateRoleRequest represents the payload for updating a role
type UpdateRoleRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// Permission defines an action that can be performed on a resource
type Permission struct {
	Id           uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Description  string    `json:"description"`
	ResourceType string    `gorm:"not null" json:"resource_type"`
	Action       string    `gorm:"not null" json:"action"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// ToResponse converts the permission to a response object
func (p *Permission) ToResponse() *PermissionResponse {
	if p == nil {
		return nil
	}
	return &PermissionResponse{
		Id:           p.Id,
		Name:         p.Name,
		Description:  p.Description,
		ResourceType: p.ResourceType,
		Action:       p.Action,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

// PermissionResponse represents the response structure for a permission
type PermissionResponse struct {
	Id           uint      `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ResourceType string    `json:"resource_type"`
	Action       string    `json:"action"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreatePermissionRequest represents the payload for creating a permission
type CreatePermissionRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	ResourceType string `json:"resource_type" binding:"required"`
	Action       string `json:"action" binding:"required"`
}

// UpdatePermissionRequest represents the payload for updating a permission
type UpdatePermissionRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// RolePermission associates permissions with roles
type RolePermission struct {
	Id           uint       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	RoleId       uint       `gorm:"column:role_id;not null;index" json:"role_id"`
	PermissionId uint       `gorm:"column:permission_id;not null;index" json:"permission_id"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	Role         Role       `gorm:"foreignKey:RoleId" json:"-"`
	Permission   Permission `gorm:"foreignKey:PermissionId" json:"-"`
}

// ToResponse converts the role permission to a response object
func (rp *RolePermission) ToResponse() *RolePermissionResponse {
	if rp == nil {
		return nil
	}
	return &RolePermissionResponse{
		Id:           rp.Id,
		RoleId:       rp.RoleId,
		PermissionId: rp.PermissionId,
		CreatedAt:    rp.CreatedAt,
	}
}

// RolePermissionResponse represents the response structure for a role permission
type RolePermissionResponse struct {
	Id           uint      `json:"id"`
	RoleId       uint      `json:"role_id"`
	PermissionId uint      `json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// ResourcePermission grants permissions on resource types or specific resources
type ResourcePermission struct {
	Id           uint       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ResourceType string     `gorm:"not null" json:"resource_type"` // Resource type (e.g., "project", "employee", etc.)
	ResourceId   string     `json:"resource_id"`                   // Optional: specific resource Id if applicable
	UserId       uint       `json:"user_id"`                       // Optional: specific user Id if applicable
	RoleId       string     `gorm:"index" json:"role_id"`          // Optional: role Id for role-based permissions
	Action       string     `json:"action"`                        // Action type (e.g., "create", "read", "update", "delete")
	DefaultScope string     `json:"default_scope"`                 // Default permission scope (e.g., "own", "team", "all")
	PermissionId uint       `gorm:"index" json:"permission_id"`    // Optional: legacy permission Id
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	Permission   Permission `gorm:"foreignKey:PermissionId" json:"-"`
	Role         Role       `gorm:"foreignKey:RoleId;references:Id" json:"-"` // Relationship to Role
}

// ToResponse converts the resource permission to a response object
func (rp *ResourcePermission) ToResponse() *ResourcePermissionResponse {
	if rp == nil {
		return nil
	}
	return &ResourcePermissionResponse{
		Id:           rp.Id,
		ResourceType: rp.ResourceType,
		ResourceId:   rp.ResourceId,
		UserId:       rp.UserId,
		RoleId:       rp.RoleId,
		Action:       rp.Action,
		DefaultScope: rp.DefaultScope,
		PermissionId: rp.PermissionId,
		CreatedAt:    rp.CreatedAt,
		UpdatedAt:    rp.UpdatedAt,
		// Role details will be added where needed
	}
}

// ResourcePermissionResponse represents the response structure for a resource permission
type ResourcePermissionResponse struct {
	Id           uint                `json:"id"`
	ResourceType string              `json:"resource_type"`
	ResourceId   string              `json:"resource_id,omitempty"`
	UserId       uint                `json:"user_id,omitempty"`
	RoleId       string              `json:"role_id,omitempty"`
	Action       string              `json:"action,omitempty"`
	DefaultScope string              `json:"default_scope,omitempty"`
	PermissionId uint                `json:"permission_id,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	Permission   *PermissionResponse `json:"permission,omitempty"`
	RoleDetails  *RoleResponse       `json:"role_details,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}

// Constants for actions
const (
	ActionCreate     = "create"
	ActionRead       = "read"
	ActionUpdate     = "update"
	ActionDelete     = "delete"
	ActionList       = "list"
	ActionAssign     = "assign"
	ActionManageRole = "manage_role"
)

// ResourceAccess defines fine-grained access control for specific resources
type ResourceAccess struct {
	Id           uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	RoleId       string    `gorm:"not null;index" json:"role_id"`
	MemberId     uint      `gorm:"not null;index" json:"member_id"`
	ResourceType string    `gorm:"not null" json:"resource_type"`
	ResourceId   string    `gorm:"not null" json:"resource_id"`
	AccessType   string    `gorm:"not null" json:"access_type"` // Permission scope (e.g., "own", "team", "all")
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// ToResponse converts the resource access to a response object
func (ra *ResourceAccess) ToResponse() *ResourceAccessResponse {
	if ra == nil {
		return nil
	}
	return &ResourceAccessResponse{
		Id:           ra.Id,
		RoleId:       ra.RoleId,
		MemberId:     ra.MemberId,
		ResourceType: ra.ResourceType,
		ResourceId:   ra.ResourceId,
		AccessType:   ra.AccessType,
		CreatedAt:    ra.CreatedAt,
		UpdatedAt:    ra.UpdatedAt,
	}
}

// ResourceAccessResponse represents the response structure for resource access
type ResourceAccessResponse struct {
	Id           uint      `json:"id"`
	RoleId       string    `json:"role_id"`
	MemberId     uint      `json:"member_id"`
	ResourceType string    `json:"resource_type"`
	ResourceId   string    `json:"resource_id"`
	AccessType   string    `json:"access_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateResourceAccessRequest represents the payload for creating resource access
type CreateResourceAccessRequest struct {
	RoleId       string `json:"role_id" binding:"required"`
	MemberId     uint   `json:"member_id" binding:"required"`
	ResourceType string `json:"resource_type" binding:"required"`
	ResourceId   string `json:"resource_id" binding:"required"`
	AccessType   string `json:"access_type" binding:"required"`
}

// UpdateResourceAccessRequest represents the payload for updating resource access
type UpdateResourceAccessRequest struct {
	RoleId       string `json:"role_id,omitempty"`
	MemberId     uint   `json:"member_id,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
	ResourceId   string `json:"resource_id,omitempty"`
	AccessType   string `json:"access_type,omitempty"`
}

// Constants for access types/scopes
const (
	AccessScopeOwn  = "own"
	AccessScopeTeam = "team"
	AccessScopeAll  = "all"
)

// Constants for table names
const (
	TableRoles               = "roles"
	TablePermissions         = "permissions"
	TableRolePermissions     = "role_permissions"
	TableResourcePermissions = "resource_permissions"
	TableResourceAccess      = "resource_access"
)

func (Permission) TableName() string {
	return "permissions"
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

func (ResourcePermission) TableName() string {
	return "resource_permissions"
}

func (ResourceAccess) TableName() string {
	return TableResourceAccess
}

// UserMembershipInfo represents user membership information
type UserMembershipInfo struct {
	UserId         uint64 `json:"user_id"`
	MemberId       uint64 `json:"member_id"`
	RoleId         uint64 `json:"role_id"`
	IsOwner        bool   `json:"is_owner"`
	Department     string `json:"department"`
	MembershipType string `json:"membership_type"`
}

func (Role) TableName() string {
	return "roles"
}
