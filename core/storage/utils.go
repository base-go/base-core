package storage

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Patterns for filename sanitization
var (
	illegalCharsPattern = regexp.MustCompile(`[^a-zA-Z0-9\-\.]`)
	multiDashPattern    = regexp.MustCompile(`-+`)
)

// slugify converts a string to a URL-friendly slug
func slugify(s string) string {
	// Convert to lowercase and replace illegal characters with dash
	s = strings.ToLower(s)
	s = illegalCharsPattern.ReplaceAllString(s, "-")
	// Replace multiple dashes with single dash
	s = multiDashPattern.ReplaceAllString(s, "-")
	// Trim dashes from ends
	return strings.Trim(s, "-")
}

// generateUniqueFilename generates a unique filename
func generateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, ext)
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d%s", slugify(name), timestamp, ext)
}
