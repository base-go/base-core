package language

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
)

// TranslationService manages translations for the application
type TranslationService struct {
	mu           sync.RWMutex
	translations map[string]map[string]string // map[locale]map[key]translation
	currentLang  Language
}

// NewTranslationService creates a new translation service
func NewTranslationService() *TranslationService {
	return &TranslationService{
		translations: make(map[string]map[string]string),
		currentLang:  GetDefaultLanguage(),
	}
}

// LoadTranslationsFromFS loads translations from an embedded filesystem
func (s *TranslationService) LoadTranslationsFromFS(fsys fs.FS, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure translations map is initialized
	if s.translations == nil {
		s.translations = make(map[string]map[string]string)
	}

	// Find all .json files in the specified path
	files, err := fs.ReadDir(fsys, path)
	if err != nil {

		return err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Extract language code from filename (e.g., "en.json" -> "en")
		langCode := strings.TrimSuffix(file.Name(), ".json")

		// Read the file content
		content, err := fs.ReadFile(fsys, filepath.Join(path, file.Name()))
		if err != nil {

			return err
		}

		// Parse the JSON content
		var data map[string]interface{}
		if err := json.Unmarshal(content, &data); err != nil {

			return err
		}

		// Initialize the translations map for this language
		if _, ok := s.translations[langCode]; !ok {
			s.translations[langCode] = make(map[string]string)
		}

		// Flatten the nested structure and add translations
		s.flattenTranslations(data, "", langCode)

	}

	return nil
}

// SetLanguage sets the current language
func (s *TranslationService) SetLanguage(lang Language) {
	s.mu.Lock()
	s.currentLang = lang
	s.mu.Unlock()
}

// GetCurrentLanguage returns the current language
func (s *TranslationService) GetCurrentLanguage() Language {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentLang
}

// flattenTranslations recursively flattens a nested map structure into a single level map
// with keys joined by dots (e.g., "section.subsection.key")
func (s *TranslationService) flattenTranslations(data map[string]interface{}, prefix string, langCode string) {
	for k, v := range data {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch value := v.(type) {
		case string:
			// Add the translation with the flattened key
			s.translations[langCode][key] = value

		case map[string]interface{}:
			// Recursively process nested maps
			s.flattenTranslations(value, key, langCode)
		case map[string]string:
			// Handle direct string maps (less common)
			for subKey, subValue := range value {
				fullKey := key + "." + subKey
				s.translations[langCode][fullKey] = subValue

			}
		default:
			// Skip other types

		}
	}
}

// Translate returns the translation for a key in the current language
func (s *TranslationService) Translate(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Try current language
	if translations, ok := s.translations[s.currentLang.Code]; ok {

		if translation, ok := translations[key]; ok {

			return translation
		} else {

		}
	} else {

	}

	// Fall back to default language
	if s.currentLang.Code != "en" {
		if translations, ok := s.translations["en"]; ok {
			if translation, ok := translations[key]; ok {

				return translation
			}
		}
	}

	// Return the key as fallback

	return key
}

// AddTranslation adds a translation for a key in a specific locale
func (s *TranslationService) AddTranslation(locale, key, translation string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.translations[locale]; !ok {
		s.translations[locale] = make(map[string]string)
	}

	s.translations[locale][key] = translation
}

// HasTranslation checks if a translation exists for a key in the current language
func (s *TranslationService) HasTranslation(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if translations, ok := s.translations[s.currentLang.Code]; ok {
		_, exists := translations[key]
		return exists
	}

	return false
}
