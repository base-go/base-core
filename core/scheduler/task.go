package scheduler

import (
	"context"
	"time"
)

// Task represents a scheduled task
type Task struct {
	Name        string
	Description string
	Schedule    Schedule
	Handler     TaskHandler
	Enabled     bool
	LastRun     *time.Time
	NextRun     *time.Time
	RunCount    int64
	ErrorCount  int64
}

// TaskHandler is the function signature for task execution
type TaskHandler func(ctx context.Context) error

// Schedule defines when a task should run
type Schedule interface {
	// ShouldRun returns true if the task should run at the given time
	ShouldRun(now time.Time, lastRun *time.Time) bool
	// NextRunTime calculates the next time this task should run
	NextRunTime(now time.Time) time.Time
	// String returns a human-readable description of the schedule
	String() string
}

// DailySchedule runs a task daily at a specific time
type DailySchedule struct {
	Hour   int // 0-23
	Minute int // 0-59
}

func (d *DailySchedule) ShouldRun(now time.Time, lastRun *time.Time) bool {
	// Check if we're at the right time
	if now.Hour() != d.Hour || now.Minute() != d.Minute {
		return false
	}
	
	// If never run before, run now
	if lastRun == nil {
		return true
	}
	
	// Don't run if already ran today
	if lastRun.Year() == now.Year() && lastRun.YearDay() == now.YearDay() {
		return false
	}
	
	return true
}

func (d *DailySchedule) NextRunTime(now time.Time) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), d.Hour, d.Minute, 0, 0, now.Location())
	
	// If the time has passed today, schedule for tomorrow
	if next.Before(now) || next.Equal(now) {
		next = next.AddDate(0, 0, 1)
	}
	
	return next
}

func (d *DailySchedule) String() string {
	return time.Date(0, 1, 1, d.Hour, d.Minute, 0, 0, time.UTC).Format("15:04 daily")
}

// MonthlySchedule runs a task monthly on a specific day
type MonthlySchedule struct {
	Day    int // 1-31 (28-31 may not exist in all months)
	Hour   int // 0-23
	Minute int // 0-59
}

func (m *MonthlySchedule) ShouldRun(now time.Time, lastRun *time.Time) bool {
	// Check if we're on the right day (handle month end gracefully)
	targetDay := m.Day
	lastDayOfMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()
	if targetDay > lastDayOfMonth {
		targetDay = lastDayOfMonth // Use last day of month if target day doesn't exist
	}
	
	if now.Day() != targetDay {
		return false
	}
	
	// Check if we're within the execution time window (allow 2-minute window around target time)
	targetTime := time.Date(now.Year(), now.Month(), now.Day(), m.Hour, m.Minute, 0, 0, now.Location())
	timeDiff := now.Sub(targetTime)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	if timeDiff > 2*time.Minute {
		return false
	}
	
	// If never run before, run now
	if lastRun == nil {
		return true
	}
	
	// Don't run if already ran this month
	if lastRun.Year() == now.Year() && lastRun.Month() == now.Month() {
		return false
	}
	
	return true
}

func (m *MonthlySchedule) NextRunTime(now time.Time) time.Time {
	targetDay := m.Day
	
	// Try current month first
	lastDayOfCurrentMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()
	if targetDay > lastDayOfCurrentMonth {
		targetDay = lastDayOfCurrentMonth
	}
	
	next := time.Date(now.Year(), now.Month(), targetDay, m.Hour, m.Minute, 0, 0, now.Location())
	
	// If the time has passed this month, schedule for next month
	if next.Before(now) || next.Equal(now) {
		nextMonth := now.AddDate(0, 1, 0)
		lastDayOfNextMonth := time.Date(nextMonth.Year(), nextMonth.Month()+1, 0, 0, 0, 0, 0, nextMonth.Location()).Day()
		
		targetDay = m.Day
		if targetDay > lastDayOfNextMonth {
			targetDay = lastDayOfNextMonth
		}
		
		next = time.Date(nextMonth.Year(), nextMonth.Month(), targetDay, m.Hour, m.Minute, 0, 0, nextMonth.Location())
	}
	
	return next
}

func (m *MonthlySchedule) String() string {
	suffix := "th"
	switch m.Day % 10 {
	case 1:
		if m.Day != 11 {
			suffix = "st"
		}
	case 2:
		if m.Day != 12 {
			suffix = "nd"
		}
	case 3:
		if m.Day != 13 {
			suffix = "rd"
		}
	}
	
	timeStr := time.Date(0, 1, 1, m.Hour, m.Minute, 0, 0, time.UTC).Format("15:04")
	return timeStr + " on " + string(rune(m.Day)) + suffix + " of each month"
}

// IntervalSchedule runs a task at regular intervals
type IntervalSchedule struct {
	Interval time.Duration
}

func (i *IntervalSchedule) ShouldRun(now time.Time, lastRun *time.Time) bool {
	if lastRun == nil {
		return true
	}
	
	return now.Sub(*lastRun) >= i.Interval
}

func (i *IntervalSchedule) NextRunTime(now time.Time) time.Time {
	return now.Add(i.Interval)
}

func (i *IntervalSchedule) String() string {
	return "every " + i.Interval.String()
}

// CronSchedule uses cron-like expressions (simplified version)
type CronSchedule struct {
	Expression string
	// You could implement full cron parsing here or use a library
}

func (c *CronSchedule) ShouldRun(now time.Time, lastRun *time.Time) bool {
	// Simplified implementation - you could use a cron library here
	return false
}

func (c *CronSchedule) NextRunTime(now time.Time) time.Time {
	// Simplified implementation
	return now.Add(time.Hour)
}

func (c *CronSchedule) String() string {
	return "cron: " + c.Expression
}