package language

// Language represents a supported language in the application
type Language struct {
	Code        string // ISO code like "en", "es", "sq"
	LocaleCode  string // Full locale code like "en_US", "es_ES", "sq_AL"
	Name        string // Display name in its own language
	EnglishName string // Display name in English
	IsDefault   bool   // Whether this is the default language
}

// GetDefaultLanguage returns the default language
func GetDefaultLanguage() Language {
	return Language{
		Code:        "en",
		LocaleCode:  "en_US",
		Name:        "English",
		EnglishName: "English",
		IsDefault:   true,
	}
}

// GetSupportedLanguages returns all supported languages
func GetSupportedLanguages() []Language {
	return []Language{
		GetDefaultLanguage(),
		{
			Code:        "es",
			LocaleCode:  "es_ES",
			Name:        "Español",
			EnglishName: "Spanish",
			IsDefault:   false,
		},
		{
			Code:        "sq",
			LocaleCode:  "sq_AL",
			Name:        "Shqip",
			EnglishName: "Albanian",
			IsDefault:   false,
		},
	}
}

// GetLanguageByCode returns a language by its code
func GetLanguageByCode(code string) (Language, bool) {
	for _, lang := range GetSupportedLanguages() {
		if lang.Code == code {
			return lang, true
		}
	}
	return GetDefaultLanguage(), false
}
