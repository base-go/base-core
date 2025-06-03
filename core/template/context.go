package template

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ViewContext struct {
	*gin.Context
	Data      interface{}
	ViewName  string
	Layout    string
	CSRFToken string
}

func NewViewContext(ctx *gin.Context) *ViewContext {
	return &ViewContext{
		Context:   ctx,
		Data:      gin.H{},
		ViewName:  "",
		Layout:    "application",
		CSRFToken: generateCSRFToken(),
	}
}

func (vc *ViewContext) Render(status int, name string, data interface{}) {
	vc.ViewName = name
	vc.Data = data
	vc.HTML(status, name, data)
}

func (vc *ViewContext) RenderWithLayout(status int, name, layout string, data interface{}) {
	vc.ViewName = name
	vc.Layout = layout
	vc.Data = data
	
	// For now, we'll use the standard HTML method
	// This will be enhanced when we integrate with the template engine
	vc.HTML(status, name, data)
}

func (vc *ViewContext) Redirect(status int, url string) {
	vc.Context.Redirect(status, url)
}

func (vc *ViewContext) JSON(status int, data interface{}) {
	vc.Context.JSON(status, data)
}

func (vc *ViewContext) Error(status int, message string) {
	vc.Context.JSON(status, gin.H{"error": message})
}

func (vc *ViewContext) Success(status int, message string) {
	vc.Context.JSON(status, gin.H{"message": message})
}

func generateCSRFToken() string {
	// Simple CSRF token generation - should be enhanced for production
	return "csrf-token-placeholder"
}

// Helper methods for common HTTP responses
func (vc *ViewContext) Ok(data interface{}) {
	vc.JSON(http.StatusOK, data)
}

func (vc *ViewContext) Created(data interface{}) {
	vc.JSON(http.StatusCreated, data)
}

func (vc *ViewContext) BadRequest(message string) {
	vc.Error(http.StatusBadRequest, message)
}

func (vc *ViewContext) NotFound(message string) {
	vc.Error(http.StatusNotFound, message)
}

func (vc *ViewContext) InternalServerError(message string) {
	vc.Error(http.StatusInternalServerError, message)
}