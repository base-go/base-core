package authorization

import (
	"base/core/logger"
	"base/core/router"
	"base/core/types"
	"fmt"
	"net/http"
	"strconv"
)

// AuthorizationController handles HTTP requests for authorization
type AuthorizationController struct {
	Service *AuthorizationService
	Logger  logger.Logger
}

// NewAuthorizationController creates a new authorization controller
func NewAuthorizationController(service *AuthorizationService, logger logger.Logger) *AuthorizationController {
	return &AuthorizationController{
		Service: service,
		Logger:  logger,
	}
}

// Routes registers routes for the authorization controller
func (c *AuthorizationController) Routes(router *router.RouterGroup) {
	c.Logger.Info("Setting up authorization routes")
	authzRoutes := router.Group("/authorization")
	{
		c.Logger.Info("Registering authorization role management routes")
		// Role management
		authzRoutes.GET("/roles", c.GetRoles)
		authzRoutes.GET("/roles/:id", c.GetRole)
		authzRoutes.POST("/roles", c.CreateRole)
		authzRoutes.PUT("/roles/:id", c.UpdateRole)
		authzRoutes.DELETE("/roles/:id", c.DeleteRole)

		// Role-permission management
		authzRoutes.GET("/roles/:id/permissions", c.GetRolePermissions)
		authzRoutes.POST("/roles/:id/permissions", c.AssignPermission)
		authzRoutes.DELETE("/roles/:id/permissions/:permissionId", c.RevokePermission)

		// Resource permissions
		authzRoutes.POST("/resource-permissions", c.CreateResourcePermission)
		authzRoutes.DELETE("/resource-permissions/:id", c.DeleteResourcePermission)

		// Permission checks
		authzRoutes.POST("/check", c.CheckPermission)

	}
	c.Logger.Info("Authorization routes registered successfully")
}

// GetRoles returns all roles for an organization
// @Summary Get all roles for an organization
// @Description Retrieves all roles associated with a specific organization via Base-Orgid header, you need to provide it in the header
// @Tags Core/Authorization
// @Security BearerAuth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {object} object{data=[]Role} "Successful operation"
// @Failure 400 {object} types.ErrorResponse "Bad request - Missing organization_id"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /authorization/roles [get]
func (c *AuthorizationController) GetRoles(ctx *router.Context) error {
	orgIdStr := ctx.GetHeader("Base-Orgid")
	var orgId uint64
	if orgIdStr != "" {
		parsedId, err := strconv.ParseUint(orgIdStr, 10, 64)
		if err == nil {
			// Successfully parsed the organization Id
			orgId = parsedId
			c.Logger.Info("Fetching roles for organization",
				logger.String("organization_id", fmt.Sprintf("%d", orgId)))
		} else {
			c.Logger.Warn("Invalid organization Id in header",
				logger.String("Base-Orgid", orgIdStr),
				logger.String("error", err.Error()))
		}
	} else {
		c.Logger.Info("No organization Id provided, fetching system roles only")
	}

	roles, err := c.Service.GetRoles(orgId)
	if err != nil {
		c.Logger.Error("Error getting roles",
			logger.String("error", err.Error()),
			logger.String("organization_id", fmt.Sprintf("%d", orgId)))

		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to retrieve roles",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"data": roles,
	})
}

// GetRole returns a specific role by Id
// @Summary Get role by Id
// @Description Retrieves a specific role by its Id
// @Tags Core/Authorization
// @Security BearerAuth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Role Id"
// @Success 200 {object} object{data=Role} "Successful operation"
// @Failure 404 {object} types.ErrorResponse "Role not found"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /authorization/roles/{id} [get]
func (c *AuthorizationController) GetRole(ctx *router.Context) error {
	roleId := ctx.Param("id")
	roleIdUint, err := strconv.ParseUint(roleId, 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid role Id: " + err.Error(),
		})
	}

	role, err := c.Service.GetRole(roleIdUint)
	if err != nil {
		if err == ErrRoleNotFound {
			return ctx.JSON(http.StatusNotFound, types.ErrorResponse{
				Error: "Role not found",
			})
		}

		c.Logger.Error("Error getting role",
			logger.String("error", err.Error()),
			logger.String("role_id", roleId))

		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to retrieve role",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"data": role,
	})
}

// CreateRole creates a new role
// @Summary Create a new role
// @Description Creates a new role with the provided information
// @Tags Core/Authorization
// @Security BearerAuth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param role body Role true "Role object to be created"
// @Success 201 {object} object{data=Role} "Role created successfully"
// @Failure 400 {object} types.ErrorResponse "Invalid role data"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /authorization/roles [post]
func (c *AuthorizationController) CreateRole(ctx *router.Context) error {
	var role Role
	if err := ctx.ShouldBindJSON(&role); err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid role data: " + err.Error(),
		})
	}

	if err := c.Service.CreateRole(&role); err != nil {
		c.Logger.Error("Error creating role",
			logger.String("error", err.Error()),
			logger.String("role_name", role.Name))

		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to create role: " + err.Error(),
		})
	}

	return ctx.JSON(http.StatusCreated, map[string]any{
		"data": role,
	})
}

// UpdateRole updates an existing role
// @Summary Update a role
// @Description Updates an existing role with the provided information
// @Tags Core/Authorization
// @Security BearerAuth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Role Id"
// @Param role body Role true "Updated role object"
// @Success 200 {object} object{data=Role} "Role updated successfully"
// @Failure 400 {object} types.ErrorResponse "Invalid role data"
// @Failure 403 {object} types.ErrorResponse "System role cannot be modified"
// @Failure 404 {object} types.ErrorResponse "Role not found"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /authorization/roles/{id} [put]
func (c *AuthorizationController) UpdateRole(ctx *router.Context) error {
	roleId := ctx.Param("id")
	roleIdInt, err := strconv.ParseUint(roleId, 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid role Id: " + err.Error(),
		})
	}

	var role Role
	if err := ctx.ShouldBindJSON(&role); err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid role data: " + err.Error(),
		})
	}

	role.Id = uint(roleIdInt)

	if err := c.Service.UpdateRole(&role); err != nil {
		switch err {
		case ErrRoleNotFound:
			return ctx.JSON(http.StatusNotFound, types.ErrorResponse{
				Error: "Role not found",
			})
		case ErrSystemRoleUnmodifiable:
			return ctx.JSON(http.StatusForbidden, types.ErrorResponse{
				Error: "System roles cannot be modified",
			})
		}

		c.Logger.Error("Error updating role",
			logger.String("error", err.Error()),
			logger.String("role_id", roleId))

		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to update role",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"data": role,
	})
}

// DeleteRole deletes a role
// @Summary Delete a role
// @Description Deletes a role by its Id
// @Tags Core/Authorization
// @Security BearerAuth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Role Id"
// @Success 200 {object} object{success=boolean} "Role deleted successfully"
// @Failure 403 {object} types.ErrorResponse "System role cannot be deleted"
// @Failure 404 {object} types.ErrorResponse "Role not found"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /authorization/roles/{id} [delete]
func (c *AuthorizationController) DeleteRole(ctx *router.Context) error {
	roleId := ctx.Param("id")
	roleIdUint, err := strconv.ParseUint(roleId, 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid role Id: " + err.Error(),
		})
	}

	if err := c.Service.DeleteRole(roleIdUint); err != nil {
		switch err {
		case ErrRoleNotFound:
			return ctx.JSON(http.StatusNotFound, types.ErrorResponse{
				Error: "Role not found",
			})
		case ErrSystemRoleUnmodifiable:
			return ctx.JSON(http.StatusForbidden, types.ErrorResponse{
				Error: "System roles cannot be deleted",
			})
		}

		c.Logger.Error("Error deleting role",
			logger.String("error", err.Error()),
			logger.String("role_id", roleId))

		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to delete role",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"success": true,
	})
}

// GetRolePermissions returns all permissions for a role
// @Summary Get permissions for a role
// @Description Retrieves all permissions associated with a specific role
// @Tags Core/Authorization
// @Security BearerAuth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Role Id"
// @Success 200 {object} object{data=[]Permission} "Successful operation"
// @Failure 404 {object} types.ErrorResponse "Role not found"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /authorization/roles/{id}/permissions [get]
func (c *AuthorizationController) GetRolePermissions(ctx *router.Context) error {
	roleId := ctx.Param("id")
	roleIdUint, err := strconv.ParseUint(roleId, 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid role Id: " + err.Error(),
		})
	}

	permissions, err := c.Service.GetRolePermissions(roleIdUint)
	if err != nil {
		if err == ErrRoleNotFound {
			return ctx.JSON(http.StatusNotFound, types.ErrorResponse{
				Error: "Role not found",
			})
		}

		c.Logger.Error("Error getting role permissions",
			logger.String("error", err.Error()),
			logger.String("role_id", roleId))

		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to retrieve permissions",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"data": permissions,
	})
}

// AssignPermission assigns a permission to a role
// @Summary Assign permission to role
// @Description Assigns a permission to a role
// @Tags Core/Authorization
// @Security BearerAuth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Role Id"
// @Param assignRequest body object{permission_id=string} true "Permission Id to assign"
// @Success 200 {object} object{success=boolean} "Permission assigned successfully"
// @Failure 400 {object} types.ErrorResponse "Invalid request data"
// @Failure 404 {object} types.ErrorResponse "Role or permission not found"
// @Failure 409 {object} types.ErrorResponse "Permission already assigned"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /authorization/roles/{id}/permissions [post]
func (c *AuthorizationController) AssignPermission(ctx *router.Context) error {
	roleId := ctx.Param("id")
	roleIdUint, err := strconv.ParseUint(roleId, 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid role Id: " + err.Error(),
		})
	}

	var request struct {
		PermissionId string `json:"permission_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid request: " + err.Error(),
		})
	}

	permissionIdUint, err := strconv.ParseUint(request.PermissionId, 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid permission Id: " + err.Error(),
		})
	}

	if err := c.Service.AssignPermissionToRole(roleIdUint, permissionIdUint); err != nil {
		switch err {
		case ErrRoleNotFound:
			return ctx.JSON(http.StatusNotFound, types.ErrorResponse{
				Error: "Role not found",
			})
		case ErrPermissionNotFound:
			return ctx.JSON(http.StatusNotFound, types.ErrorResponse{
				Error: "Permission not found",
			})
		case ErrDuplicatePermission:
			return ctx.JSON(http.StatusConflict, types.ErrorResponse{
				Error: "Permission already assigned to this role",
			})
		}

		c.Logger.Error("Error assigning permission",
			logger.String("error", err.Error()),
			logger.String("role_id", roleId),
			logger.String("permission_id", request.PermissionId))

		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to assign permission",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"success": true,
	})
}

// RevokePermission removes a permission from a role
// @Summary Revoke permission from role
// @Description Removes a permission from a role
// @Tags Core/Authorization
// @Security BearerAuth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Role Id"
// @Param permissionId path string true "Permission Id"
// @Success 200 {object} object{success=boolean} "Permission revoked successfully"
// @Failure 404 {object} types.ErrorResponse "Role or permission not found"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /authorization/roles/{id}/permissions/{permissionId} [delete]
func (c *AuthorizationController) RevokePermission(ctx *router.Context) error {
	roleId := ctx.Param("id")
	permissionId := ctx.Param("permissionId")

	roleIdUint, err := strconv.ParseUint(roleId, 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid role Id: " + err.Error(),
		})
	}

	permissionIdUint, err := strconv.ParseUint(permissionId, 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid permission Id: " + err.Error(),
		})
	}

	if err := c.Service.RevokePermissionFromRole(roleIdUint, permissionIdUint); err != nil {
		switch err {
		case ErrRoleNotFound:
			return ctx.JSON(http.StatusNotFound, types.ErrorResponse{
				Error: "Role not found",
			})
		case ErrPermissionNotFound:
			return ctx.JSON(http.StatusNotFound, types.ErrorResponse{
				Error: "Permission not found",
			})
		}

		c.Logger.Error("Error revoking permission",
			logger.String("error", err.Error()),
			logger.String("role_id", roleId),
			logger.String("permission_id", permissionId))

		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to revoke permission",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"success": true,
	})
}

// CreateResourcePermission creates a resource-specific permission
// @Summary Create resource permission
// @Description Creates a resource-specific permission override
// @Tags Core/Authorization
// @Security BearerAuth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param resourcePermission body ResourcePermission true "Resource permission to create"
// @Success 201 {object} object{data=ResourcePermission} "Resource permission created successfully"
// @Failure 400 {object} types.ErrorResponse "Invalid resource permission data"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /authorization/resource-permissions [post]
func (c *AuthorizationController) CreateResourcePermission(ctx *router.Context) error {
	var resourcePermission ResourcePermission
	if err := ctx.ShouldBindJSON(&resourcePermission); err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid resource permission data: " + err.Error(),
		})
	}

	if err := c.Service.CreateResourcePermission(&resourcePermission); err != nil {
		c.Logger.Error("Error creating resource permission",
			logger.String("error", err.Error()),
			logger.String("resource_type", resourcePermission.ResourceType),
			logger.String("resource_id", resourcePermission.ResourceId))

		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to create resource permission",
		})
	}

	return ctx.JSON(http.StatusCreated, map[string]any{
		"data": resourcePermission,
	})
}

// DeleteResourcePermission deletes a resource-specific permission
// @Summary Delete resource permission
// @Description Deletes a resource-specific permission override
// @Tags Core/Authorization
// @Security BearerAuth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Resource Permission Id"
// @Success 200 {object} object{success=boolean} "Resource permission deleted successfully"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /authorization/resource-permissions/{id} [delete]
func (c *AuthorizationController) DeleteResourcePermission(ctx *router.Context) error {
	id := ctx.Param("id")
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid resource permission Id: " + err.Error(),
		})
	}

	if err := c.Service.DeleteResourcePermission(idUint); err != nil {
		c.Logger.Error("Error deleting resource permission",
			logger.String("error", err.Error()),
			logger.String("id", id))

		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to delete resource permission",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"success": true,
	})
}

// CheckPermission checks if a user has a specific permission
// @Summary Check user permission
// @Description Checks if a user has permission to perform an action on a resource
// @Tags Core/Authorization
// @Security BearerAuth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param checkRequest body object{user_id=string,organization_id=string,resource_type=string,action=string,resource_id=string} true "Permission check request"
// @Success 200 {object} object{has_permission=boolean} "Permission check result"
// @Failure 400 {object} types.ErrorResponse "Invalid request data"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /authorization/check [post]
func (c *AuthorizationController) CheckPermission(ctx *router.Context) error {
	var request struct {
		UserId       uint64 `json:"user_id" binding:"required"`
		OrgId        uint64 `json:"organization_id" binding:"required"`
		ResourceType string `json:"resource_type" binding:"required"`
		Action       string `json:"action" binding:"required"`
		ResourceId   string `json:"resource_id"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		return ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid request: " + err.Error(),
		})
	}

	var hasPermission bool
	var err error

	if request.ResourceId != "" {
		hasPermission, err = c.Service.HasResourcePermission(
			request.UserId,
			request.OrgId,
			request.ResourceType,
			request.ResourceId,
			request.Action,
		)
	} else {
		hasPermission, err = c.Service.HasPermission(
			request.UserId,
			request.OrgId,
			request.ResourceType,
			request.Action,
		)
	}

	if err != nil {
		c.Logger.Error("Error checking permission",
			logger.String("error", err.Error()),
			logger.String("user_id", fmt.Sprintf("%d", request.UserId)),
			logger.String("organization_id", fmt.Sprintf("%d", request.OrgId)),
			logger.String("resource_type", request.ResourceType),
			logger.String("action", request.Action),
			logger.String("resource_id", request.ResourceId))

		return ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to check permission",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"has_permission": hasPermission,
	})
}
