package language

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// LanguageCookieName is the name of the cookie that stores the user's language preference
const LanguageCookieName = "language"

// GetLanguageFromPath extracts the language code from the URL path
func GetLanguageFromPath(path string) (string, string) {
	// Skip the leading slash and get the first segment
	parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 2)
	if len(parts) == 0 {
		return "", ""
	}

	// Check if the first part is a valid language code
	if lang, found := GetLanguageByCode(parts[0]); found {
		remainingPath := ""
		if len(parts) > 1 {
			remainingPath = "/" + parts[1]
		}
		return lang.Code, remainingPath
	}

	return "", ""
}

// LocalizedRoutingMiddleware extracts language from URL path and sets it in the context
func LocalizedRoutingMiddleware(service *TranslationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		fmt.Printf("=== Processing request for path: %s ===\n", path)
		fmt.Printf("=== Accept-Language header: %s ===\n", c.GetHeader("Accept-Language"))
		fmt.Printf("=== Language cookie: %s ===\n", c.Request.Header.Get("Cookie"))

		// Extract language code from URL path
		pathSegments := strings.Split(strings.TrimPrefix(path, "/"), "/")
		var langCode string
		if len(pathSegments) > 0 && len(pathSegments[0]) == 2 {
			langCode = pathSegments[0]
			fmt.Printf("=== Extracted language code from URL: %s ===\n", langCode)
		}

		// Check if language code is valid
		if langCode != "" {
			var lang *Language
			// Debug available languages
			fmt.Printf("=== Available languages: ===\n")
			supportedLanguages := GetSupportedLanguages()
			for _, l := range supportedLanguages {
				fmt.Printf("  - %s (%s)\n", l.Name, l.Code)
				if l.Code == langCode {
					lang = &l
				}
			}

			if lang != nil {
				// Set language in the service
				service.SetLanguage(*lang)
				fmt.Printf("=== Setting language to: %s (%s) ===\n", lang.Name, lang.Code)

				// Store language and translation service in context
				c.Set("language", *lang)
				c.Set("TranslationService", service)

				// Debug translations for this language
				transMap, exists := service.translations[lang.Code]
				if exists {
					fmt.Printf("=== Found %d translations for language %s ===\n", len(transMap), lang.Code)
					// Print a few sample translations
					sampleCount := 0
					fmt.Printf("=== Sample translations from %s: ===\n", lang.Code)
					for k, v := range transMap {
						if strings.HasPrefix(k, "posts.") && sampleCount < 5 {
							fmt.Printf("  %s: %s\n", k, v)
							sampleCount++
						}
					}
				} else {
					fmt.Printf("=== WARNING: No translations found for language %s! ===\n", lang.Code)
				}

				// Set language cookie
				c.SetCookie(
					LanguageCookieName,
					lang.Code,
					60*60*24*30, // 30 days expiration
					"/",         // path
					"",          // domain
					false,       // secure
					false,       // httpOnly
				)

				// Store language in context
				c.Set("language", lang)
				c.Set("translationService", service)
				c.Next()
				return
			} else {
				fmt.Printf("=== Language code %s not found in available languages ===\n", langCode)
			}
		}

		// If no language in path, get from cookie or use default
		// Redirect to default language if no language found in URL
		defaultLang := GetDefaultLanguage()
		service.SetLanguage(defaultLang)
		fmt.Printf("=== No valid language found, using default: %s ===\n", defaultLang.Code)

		// Check if this is an API request, skip redirect for API calls
		if strings.HasPrefix(path, "/api/") {
			c.Next()
			return
		}

		// Redirect to localized URL
		targetPath := "/" + defaultLang.Code
		if path != "/" {
			targetPath = targetPath + path
		}
		fmt.Printf("=== Redirecting to: %s ===\n", targetPath)
		c.Redirect(302, targetPath)
		c.Abort()

	}
}

// LanguageMiddleware sets the language for the current request
// This is kept for backward compatibility
func LanguageMiddleware(service *TranslationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get language from query parameter first
		langCode := c.Query("lang")

		// If not found, try to get from cookie
		if langCode == "" {
			langCode, _ = c.Cookie(LanguageCookieName)
		}

		// If a valid language is found, set it as the current language
		if langCode != "" {
			if lang, found := GetLanguageByCode(langCode); found {
				service.SetLanguage(lang)

				// Set/update the cookie for future requests
				c.SetCookie(
					LanguageCookieName,
					langCode,
					60*60*24*30, // 30 days expiration
					"/",         // path
					"",          // domain
					false,       // secure
					false,       // httpOnly
				)
			}
		}

		// Store the translation service in the context for templates
		c.Set("TranslationService", service)
		c.Set("CurrentLanguage", service.GetCurrentLanguage())

		c.Next()
	}
}
