package home

import (
	"base/core/language"
	"base/core/layout"
	"base/core/router"

	"github.com/gin-gonic/gin"
)

type HomeController struct {
	*layout.Controller
}

// getTranslation gets a translation string from the context
// It handles cases where the translation service might not be available
func getTranslation(ctx *gin.Context, key string, defaultValue string) string {
	translationService, exists := ctx.Get("TranslationService")
	if !exists {
		return defaultValue
	}

	translated := translationService.(*language.TranslationService).Translate(key)
	if translated == key {
		// Key not found, return default
		return defaultValue
	}

	return translated
}

func NewHomeController(layoutEngine *layout.Engine) *HomeController {
	return &HomeController{
		Controller: layout.NewLandingController(layoutEngine), // Use landing layout for home
	}
}

func (c *HomeController) Routes(ginRouter *gin.RouterGroup) {

	router.NewFromGroup(ginRouter).
		Get("/", c.Index).     // Home page
		Get("/about", c.About) // About page
}

// Index renders the home/landing page
func (c *HomeController) Index(ctx *gin.Context) {
	// Sample features data
	features := []map[string]string{
		{
			"title":       getTranslation(ctx, "home.feature1_title", "Fast & Efficient"),
			"description": getTranslation(ctx, "home.feature1_description", "Built with modern technology for optimal performance"),
		},
		{
			"title":       getTranslation(ctx, "home.feature2_title", "User Friendly"),
			"description": getTranslation(ctx, "home.feature2_description", "Intuitive interface designed for the best user experience"),
		},
		{
			"title":       getTranslation(ctx, "home.feature3_title", "Secure"),
			"description": getTranslation(ctx, "home.feature3_description", "Enterprise-grade security to protect your data"),
		},
	}

	c.View("home/index.html").
		WithTitle(getTranslation(ctx, "home.title", "Welcome")).
		WithData(map[string]interface{}{
			"hero_title":    getTranslation(ctx, "home.hero_title", "Welcome to Base"),
			"hero_subtitle": getTranslation(ctx, "home.hero_subtitle", "The modern platform for your next project"),
			"features":      features,
		}).
		Render(ctx)
}

// About renders the about page
func (c *HomeController) About(ctx *gin.Context) {
	c.View("home/about.html").
		WithTitle(getTranslation(ctx, "home.about_title", "About")).
		Render(ctx)
}
