package language

import (
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

// CookieLanguageMiddleware sets the language based on cookie preference
func CookieLanguageMiddleware(service *TranslationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get language from query parameter first (for language switching)
		langCode := c.Query("lang")

		// If not found, try to get from cookie
		if langCode == "" {
			langCode, _ = c.Cookie(LanguageCookieName)
		}

		// If a valid language is found, set it as the current language
		var currentLang Language
		if langCode != "" {
			if lang, found := GetLanguageByCode(langCode); found {
				service.SetLanguage(lang)
				currentLang = lang

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
			} else {
				// Invalid language code, use default
				currentLang = GetDefaultLanguage()
				service.SetLanguage(currentLang)
			}
		} else {
			// No language preference found, use default
			currentLang = GetDefaultLanguage()
			service.SetLanguage(currentLang)
		}

		// Store the translation service and language in the context for templates
		c.Set("TranslationService", service)
		c.Set("CurrentLanguage", currentLang)

		c.Next()
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
