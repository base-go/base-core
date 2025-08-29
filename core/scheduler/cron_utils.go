package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ConvertToCronExpression converts "DD HH:MM" format to cron expression
// Examples:
// "25 09:00" -> "0 0 9 25 * *" (25th day at 9:00 AM)
// "1 14:30" -> "0 30 14 1 * *" (1st day at 2:30 PM)
func ConvertToCronExpression(dateTimeStr string) (string, error) {
	parts := strings.Split(strings.TrimSpace(dateTimeStr), " ")
	if len(parts) != 2 {
		// Try legacy format (just day)
		if day, err := strconv.Atoi(strings.TrimSpace(dateTimeStr)); err == nil {
			if day < 1 || day > 31 {
				return "", fmt.Errorf("invalid day: %d (must be 1-31)", day)
			}
			// Default to 09:00 for legacy format
			return fmt.Sprintf("0 0 9 %d * *", day), nil
		}
		return "", fmt.Errorf("invalid format: expected 'DD HH:MM' or 'DD', got '%s'", dateTimeStr)
	}
	
	// Parse day
	day, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid day: %s", parts[0])
	}
	if day < 1 || day > 31 {
		return "", fmt.Errorf("invalid day: %d (must be 1-31)", day)
	}
	
	// Parse time
	timeParts := strings.Split(parts[1], ":")
	if len(timeParts) != 2 {
		return "", fmt.Errorf("invalid time format: expected 'HH:MM', got '%s'", parts[1])
	}
	
	hour, err := strconv.Atoi(timeParts[0])
	if err != nil {
		return "", fmt.Errorf("invalid hour: %s", timeParts[0])
	}
	if hour < 0 || hour > 23 {
		return "", fmt.Errorf("invalid hour: %d (must be 0-23)", hour)
	}
	
	minute, err := strconv.Atoi(timeParts[1])
	if err != nil {
		return "", fmt.Errorf("invalid minute: %s", timeParts[1])
	}
	if minute < 0 || minute > 59 {
		return "", fmt.Errorf("invalid minute: %d (must be 0-59)", minute)
	}
	
	// Convert to cron expression: "second minute hour day month dayofweek"
	// For monthly schedule: "0 minute hour day * *"
	return fmt.Sprintf("0 %d %d %d * *", minute, hour, day), nil
}

// ConvertToNextMinuteCron creates a cron expression for the next minute
func ConvertToNextMinuteCron() string {
	now := time.Now()
	nextMinute := now.Add(time.Minute)
	
	// Run at the start of the next minute: "0 minute hour day month *"
	return fmt.Sprintf("0 %d %d %d %d *", 
		nextMinute.Minute(), 
		nextMinute.Hour(), 
		nextMinute.Day(), 
		int(nextMinute.Month()))
}

// ValidateCronExpression validates a cron expression
func ValidateCronExpression(cronExpr string) error {
	parts := strings.Fields(cronExpr)
	if len(parts) != 6 {
		return fmt.Errorf("cron expression must have 6 fields (seconds minute hour day month dayofweek)")
	}
	return nil
}

// DescribeCronExpression provides a human-readable description of a cron expression
func DescribeCronExpression(cronExpr string) string {
	parts := strings.Fields(cronExpr)
	if len(parts) != 6 {
		return cronExpr
	}
	
	second := parts[0]
	minute := parts[1]
	hour := parts[2]
	day := parts[3]
	month := parts[4]
	// dayOfWeek := parts[5] // Not used for monthly schedules
	
	// Handle monthly schedules (day is specific, month and dayOfWeek are *)
	if month == "*" && day != "*" {
		timeStr := fmt.Sprintf("%s:%s", hour, minute)
		if second == "0" {
			return fmt.Sprintf("%s on %s of each month", timeStr, addOrdinalSuffix(day))
		}
		return fmt.Sprintf("%s:%s on %s of each month", timeStr, second, addOrdinalSuffix(day))
	}
	
	return cronExpr // Fallback to raw expression
}

// addOrdinalSuffix adds ordinal suffix to day numbers (1st, 2nd, 3rd, etc.)
func addOrdinalSuffix(dayStr string) string {
	day, err := strconv.Atoi(dayStr)
	if err != nil {
		return dayStr
	}
	
	suffix := "th"
	switch day % 10 {
	case 1:
		if day != 11 {
			suffix = "st"
		}
	case 2:
		if day != 12 {
			suffix = "nd"
		}
	case 3:
		if day != 13 {
			suffix = "rd"
		}
	}
	
	return fmt.Sprintf("%d%s", day, suffix)
}