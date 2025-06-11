package language

import (
	"strings"

	"base/core/layout"
)

// RegisterTemplateHelpers registers translation helper functions with the template engine
func RegisterTemplateHelpers(engine *layout.Engine, service *TranslationService) {

	// Add a helper function for translating text
	engine.AddHelper("t", func(key string, data ...interface{}) string {
		// Print debug info to help troubleshoot translation issues

		// Get the translation service from template data if available
		var translationService *TranslationService

		// Check if template data was passed (it should be the first argument)
		if len(data) > 0 {
			// Get template data (it should be a map)
			if templateData, ok := data[0].(map[string]interface{}); ok {
				// Try to get TranslationService directly from template data
				if ts, exists := templateData["_ts"]; exists {
					if ts, ok := ts.(*TranslationService); ok {
						translationService = ts
					}
				}
			}
		}

		// Fallback to the global service if not available in template data
		if translationService == nil {
			translationService = service
		}

		// First, try direct translation with the exact key
		translation := translationService.Translate(key)

		// If the key was returned (meaning no translation found), try prefixing with sections
		if translation == key {
			// Try common section keys
			prefixes := []string{"common.", "posts.", "navigation.", "errors.", "auth.", "landing."}

			for _, prefix := range prefixes {
				prefixedKey := prefix + key
				prefixedTranslation := translationService.Translate(prefixedKey)

				// If we found a translation with the prefix, use it
				if prefixedTranslation != prefixedKey {

					return prefixedTranslation
				}
			}
		}

		return translation
	})

	// Add a helper function for accessing nested translation keys
	engine.AddHelper("tr", func(key string, data ...interface{}) string {
		// This handles keys in the format "section.key" like "common.save"
		parts := strings.Split(key, ".")
		if len(parts) != 2 {
			return key
		}

		// Get the translation service from template data if available
		var translationService *TranslationService

		// Check if template data was passed (it should be the first argument)
		if len(data) > 0 {
			// Get template data (it should be a map)
			if templateData, ok := data[0].(map[string]interface{}); ok {
				// Try to get TranslationService directly from template data
				if ts, exists := templateData["_ts"]; exists {
					if ts, ok := ts.(*TranslationService); ok {
						translationService = ts
					}
				}
			}
		}

		// Fallback to the global service if not available in template data
		if translationService == nil {
			translationService = service
		}

		// Use the Translate method with the original key format
		return translationService.Translate(parts[0] + "." + parts[1])
	})

	// Add a helper function to get the current language
	engine.AddHelper("currentLanguage", func(data ...interface{}) Language {
		// First try to get language directly from template data
		if len(data) > 0 {
			if templateData, ok := data[0].(map[string]interface{}); ok {
				// Try to get language from template data
				if lang, exists := templateData["_lang"]; exists {
					if language, ok := lang.(Language); ok {
						return language
					}
				}

				// Try to get TranslationService from template data
				if ts, exists := templateData["_ts"]; exists {
					if translationService, ok := ts.(*TranslationService); ok {
						lang := translationService.GetCurrentLanguage()
						if lang.Code != "" {
							return lang
						}
					}
				}
			}
		}

		// Fallback to the global service
		var translationService *TranslationService = service

		lang := translationService.GetCurrentLanguage()
		// Fallback if no language is set
		if lang.Code == "" {
			return Language{Code: "en", Name: "English"}
		}
		return lang
	})

	// Add a helper function to get all supported languages
	engine.AddHelper("supportedLanguages", func() []Language {
		return GetSupportedLanguages()
	})

	// Add a helper to check if a language is the current one
	engine.AddHelper("isCurrentLanguage", func(code string, data ...interface{}) bool {
		// First try to get language directly from template data
		if len(data) > 0 {
			if templateData, ok := data[0].(map[string]interface{}); ok {
				// Try to get language from template data
				if lang, exists := templateData["_lang"]; exists {
					if language, ok := lang.(Language); ok {
						return language.Code == code
					}
				}

				// Try to get TranslationService from template data
				if ts, exists := templateData["_ts"]; exists {
					if translationService, ok := ts.(*TranslationService); ok {
						lang := translationService.GetCurrentLanguage()
						if lang.Code != "" {
							return lang.Code == code
						}
					}
				}
			}
		}

		// Fallback to the global service
		var translationService *TranslationService = service

		return translationService.GetCurrentLanguage().Code == code
	})

}
