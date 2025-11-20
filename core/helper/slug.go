package helper

import (
	"fmt"

	"github.com/gosimple/slug"
)

// SlugHelper provides methods for generating and validating slugs
type SlugHelper struct{}

// NewSlugHelper creates a new instance of SlugHelper
func NewSlugHelper() *SlugHelper {
	return &SlugHelper{}
}

// Normalize ensures the slug is properly formatted
// If customSlug is provided, it will be used as the base for the slug
// Otherwise, name will be used to generate a slug
func (h *SlugHelper) Normalize(name string, customSlug string, lang string) string {
	if customSlug != "" {
		// If a custom slug is provided, use it but ensure it's properly formatted
		return slug.MakeLang(customSlug, lang)
	}

	// Generate slug from name
	return slug.MakeLang(name, lang)
}

// GenerateUniqueSlug generates a unique slug based on the given base slug and a function to check if a slug exists
// The existsFunc should return true if the slug already exists, and false otherwise
func (h *SlugHelper) GenerateUniqueSlug(baseSlug string, existsFunc func(string) (bool, error)) (string, error) {
	// First check if the base slug is available
	exists, err := existsFunc(baseSlug)
	if err != nil {
		return "", err
	}

	// If the base slug is available, use it
	if !exists {
		return baseSlug, nil
	}

	// Otherwise, find an available slug by adding a sequential number
	uniqueSlug := baseSlug
	for i := 2; ; i++ {
		uniqueSlug = fmt.Sprintf("%s-%d", baseSlug, i)
		exists, err = existsFunc(uniqueSlug)
		if err != nil {
			return "", err
		}
		if !exists {
			return uniqueSlug, nil
		}
	}
}
