package posts

import (
	"context"
	"time"

	"base/core/logger"
	"base/core/scheduler"
)

// PublishTask handles publish
type PublishTask struct {
	logger logger.Logger
}

// NewPublishTask creates a new PublishTask instance
func NewPublishTask(log logger.Logger) *PublishTask {
	return &PublishTask{
		logger: log,
	}
}

// RegisterTask registers the publish task with the scheduler
func (t *PublishTask) RegisterTask(s *scheduler.Scheduler) error {
	task := &scheduler.Task{
		Name:        "publish",
		Description: "publish task for posts module",
		Schedule:    &scheduler.DailySchedule{Hour: 2, Minute: 0}, // 2:00 AM daily
		Handler:     t.execute,
		Enabled:     true,
	}

	return s.RegisterTask(task)
}

// RegisterCronTask registers the publish task with cron scheduler (alternative)
func (t *PublishTask) RegisterCronTask(cs *scheduler.CronScheduler) error {
	task := &scheduler.CronTask{
		Name:        "publish",
		Description: "publish task for posts module",
		CronExpr:    "0 0 2 * * *", // 2:00 AM daily
		Handler:     t.execute,
		Enabled:     true,
	}

	return cs.RegisterTask(task)
}

// execute is the main task execution function
func (t *PublishTask) execute(ctx context.Context) error {
	t.logger.Info("Starting publish task")

	// Check for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// TODO: Implement your task logic here
	// Example:
	// - Clean up old records
	// - Send notifications
	// - Generate reports
	// - Backup data
	// - Process queued items

	// Simulate work (remove this in your implementation)
	time.Sleep(1 * time.Second)

	t.logger.Info("publish task completed successfully")
	return nil
}

// GetTaskInfo returns information about this task
func (t *PublishTask) GetTaskInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "publish",
		"description": "publish task for posts module",
		"module":      "posts",
		"type":        "scheduled_task",
	}
}
