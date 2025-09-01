package types

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// DateTime is a custom type for handling time values with flexible parsing
type DateTime struct {
	time.Time
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (dt *DateTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		dt.Time = time.Time{}
		return nil
	}

	// Try different time formats in order of preference
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		dt.Time = t
		return nil
	}

	// Try other formats if RFC3339 fails
	formats := []string{
		"2006-01-02T15:04:05-0700", // Without colon in timezone
		"2006-01-02T15:04:05Z",     // UTC
		"2006-01-02T15:04:05",      // No timezone
		"2006-01-02 15:04:05",      // MySQL format
		"2006-01-02",               // Just date
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			dt.Time = t
			return nil
		}
	}

	return fmt.Errorf("cannot parse time: %v. Expected format: RFC3339 (e.g., 2006-01-02T15:04:05Z07:00) or YYYY-MM-DD", err)
}

// MarshalJSON implements the json.Marshaler interface
func (dt DateTime) MarshalJSON() ([]byte, error) {
	if dt.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", dt.Time.Format(time.RFC3339))), nil
}

// JSONSchema returns the JSON schema for DateTime to be treated as a string
func (DateTime) JSONSchema() *JSONSchemaType {
	return &JSONSchemaType{
		Type:        "string",
		Format:      "date-time",
		Example:     "2024-12-23",
		Description: "DateTime field that accepts multiple formats like '2024-12-23' or '2024-12-23T15:04:05Z'",
	}
}

// JSONSchemaType represents the JSON schema structure
type JSONSchemaType struct {
	Type        string `json:"type"`
	Format      string `json:"format,omitempty"`
	Example     string `json:"example,omitempty"`
	Description string `json:"description,omitempty"`
}

// Value implements the driver.Valuer interface for database operations
func (dt DateTime) Value() (driver.Value, error) {
	if dt.Time.IsZero() {
		return nil, nil
	}
	return dt.Time, nil
}

// Scan implements the sql.Scanner interface for database operations
func (dt *DateTime) Scan(value any) error {
	if value == nil {
		dt.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		dt.Time = v
		return nil
	case []byte:
		formats := []string{
			"2006-01-02 15:04:05",     // MySQL datetime
			"2006-01-02 15:04:05.000", // MySQL datetime with milliseconds
			"2006-01-02",              // Date only
		}

		var lastErr error
		for _, format := range formats {
			if t, err := time.Parse(format, string(v)); err == nil {
				dt.Time = t
				return nil
			} else {
				lastErr = err
			}
		}
		return lastErr
	case string:
		if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			dt.Time = t
			return nil
		}
		return fmt.Errorf("cannot parse time string: %v", v)
	default:
		return fmt.Errorf("cannot scan type %T into DateTime", value)
	}
}

// String implements the Stringer interface
func (dt DateTime) String() string {
	if dt.Time.IsZero() {
		return ""
	}
	return dt.Time.Format(time.RFC3339)
}

// Now returns the current time as DateTime
func Now() DateTime {
	return DateTime{Time: time.Now()}
}

// IsZero reports whether the DateTime represents the zero time instant
func (dt DateTime) IsZero() bool {
	return dt.Time.IsZero()
}

// Format returns a textual representation of the time value formatted according to the layout
func (dt DateTime) Format(layout string) string {
	return dt.Time.Format(layout)
}

// Add returns the time t+d
func (dt DateTime) Add(d time.Duration) DateTime {
	return DateTime{Time: dt.Time.Add(d)}
}

// Sub returns the duration t-u
func (dt DateTime) Sub(u DateTime) time.Duration {
	return dt.Time.Sub(u.Time)
}

// Before reports whether the time instant t is before u
func (dt DateTime) Before(u DateTime) bool {
	return dt.Time.Before(u.Time)
}

// After reports whether the time instant t is after u
func (dt DateTime) After(u DateTime) bool {
	return dt.Time.After(u.Time)
}

// Equal reports whether t and u represent the same time instant
func (dt DateTime) Equal(u DateTime) bool {
	return dt.Time.Equal(u.Time)
}
