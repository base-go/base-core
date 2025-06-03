package app

import (
	"embed"

	"base/core/language"
)

//go:embed translations/*.json
var embeddedTranslations embed.FS

// RegisterTranslations loads all translations from the embedded filesystem
// and registers them with the translation service
func RegisterTranslations(service *language.TranslationService) error {
	// Load translations from embedded filesystem
	return service.LoadTranslationsFromFS(embeddedTranslations, "translations")
}
