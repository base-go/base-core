package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"base/core/logger"
)

// Scheduler manages and executes scheduled tasks
type Scheduler struct {
	tasks       map[string]*Task
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	logger      logger.Logger
	running     bool
	checkInterval time.Duration
}

// NewScheduler creates a new scheduler instance
func NewScheduler(log logger.Logger) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Scheduler{
		tasks:         make(map[string]*Task),
		ctx:           ctx,
		cancel:        cancel,
		logger:        log,
		checkInterval: time.Minute, // Check every minute by default
	}
}

// SetCheckInterval sets how often the scheduler checks for tasks to run
func (s *Scheduler) SetCheckInterval(interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkInterval = interval
}

// RegisterTask adds a new task to the scheduler
func (s *Scheduler) RegisterTask(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if task.Name == "" {
		return fmt.Errorf("task name cannot be empty")
	}
	
	if task.Handler == nil {
		return fmt.Errorf("task handler cannot be nil")
	}
	
	if task.Schedule == nil {
		return fmt.Errorf("task schedule cannot be nil")
	}
	
	// Calculate initial next run time
	now := time.Now()
	nextRun := task.Schedule.NextRunTime(now)
	task.NextRun = &nextRun
	
	s.tasks[task.Name] = task
	
	s.logger.Info("Registered scheduled task",
		logger.String("name", task.Name),
		logger.String("description", task.Description),
		logger.String("schedule", task.Schedule.String()),
		logger.String("next_run", nextRun.Format("2006-01-02 15:04:05")),
		logger.String("enabled", fmt.Sprintf("%t", task.Enabled)),
	)
	
	return nil
}

// UnregisterTask removes a task from the scheduler
func (s *Scheduler) UnregisterTask(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.tasks[name]; exists {
		delete(s.tasks, name)
		s.logger.Info("Unregistered scheduled task", logger.String("name", name))
	}
}

// EnableTask enables a task
func (s *Scheduler) EnableTask(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	task, exists := s.tasks[name]
	if !exists {
		return fmt.Errorf("task %s not found", name)
	}
	
	task.Enabled = true
	s.logger.Info("Enabled scheduled task", logger.String("name", name))
	return nil
}

// DisableTask disables a task
func (s *Scheduler) DisableTask(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	task, exists := s.tasks[name]
	if !exists {
		return fmt.Errorf("task %s not found", name)
	}
	
	task.Enabled = false
	s.logger.Info("Disabled scheduled task", logger.String("name", name))
	return nil
}

// GetTask returns a task by name
func (s *Scheduler) GetTask(name string) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	task, exists := s.tasks[name]
	return task, exists
}

// GetAllTasks returns all registered tasks
func (s *Scheduler) GetAllTasks() map[string]*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Return a copy to prevent external modification
	tasks := make(map[string]*Task)
	for name, task := range s.tasks {
		tasks[name] = task
	}
	
	return tasks
}

// Start begins the scheduler loop
func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()
	
	s.logger.Info("Starting task scheduler", logger.String("check_interval", s.checkInterval.String()))
	
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Task scheduler stopped")
			return
		case <-ticker.C:
			s.checkAndRunTasks()
		}
	}
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()
	
	s.logger.Info("Stopping task scheduler")
	s.cancel()
}

// RunTaskNow executes a task immediately (bypassing schedule)
func (s *Scheduler) RunTaskNow(name string) error {
	s.mu.RLock()
	task, exists := s.tasks[name]
	s.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("task %s not found", name)
	}
	
	if !task.Enabled {
		return fmt.Errorf("task %s is disabled", name)
	}
	
	s.logger.Info("Running task manually", logger.String("name", name))
	return s.executeTask(task)
}

// checkAndRunTasks checks all tasks and runs those that are due
func (s *Scheduler) checkAndRunTasks() {
	now := time.Now()
	
	s.mu.RLock()
	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		if task.Enabled && task.Schedule.ShouldRun(now, task.LastRun) {
			tasks = append(tasks, task)
		}
	}
	s.mu.RUnlock()
	
	// Execute tasks outside of the read lock
	for _, task := range tasks {
		go func(t *Task) {
			if err := s.executeTask(t); err != nil {
				s.logger.Error("Task execution failed",
					logger.String("name", t.Name),
					logger.String("error", err.Error()),
				)
			}
		}(task)
	}
}

// executeTask runs a single task and updates its metadata
func (s *Scheduler) executeTask(task *Task) error {
	startTime := time.Now()
	
	s.logger.Info("Executing scheduled task",
		logger.String("name", task.Name),
		logger.String("description", task.Description),
	)
	
	// Create a context with timeout for the task
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Minute) // 30 minute timeout
	defer cancel()
	
	// Execute the task
	err := task.Handler(ctx)
	
	// Update task metadata
	s.mu.Lock()
	now := time.Now()
	task.LastRun = &now
	task.RunCount++
	
	if err != nil {
		task.ErrorCount++
	}
	
	// Calculate next run time
	nextRun := task.Schedule.NextRunTime(now)
	task.NextRun = &nextRun
	s.mu.Unlock()
	
	duration := time.Since(startTime)
	
	if err != nil {
		s.logger.Error("Scheduled task failed",
			logger.String("name", task.Name),
			logger.String("duration", duration.String()),
			logger.String("error", err.Error()),
		)
		return err
	}
	
	s.logger.Info("Scheduled task completed successfully",
		logger.String("name", task.Name),
		logger.String("duration", duration.String()),
		logger.String("next_run", nextRun.Format("2006-01-02 15:04:05")),
	)
	
	return nil
}

// GetStats returns scheduler statistics
func (s *Scheduler) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	stats := map[string]interface{}{
		"running":        s.running,
		"check_interval": s.checkInterval.String(),
		"total_tasks":    len(s.tasks),
		"enabled_tasks":  0,
		"disabled_tasks": 0,
		"tasks":          []map[string]interface{}{},
	}
	
	tasks := make([]map[string]interface{}, 0, len(s.tasks))
	enabledCount := 0
	
	for _, task := range s.tasks {
		if task.Enabled {
			enabledCount++
		}
		
		taskStats := map[string]interface{}{
			"name":        task.Name,
			"description": task.Description,
			"enabled":     task.Enabled,
			"schedule":    task.Schedule.String(),
			"run_count":   task.RunCount,
			"error_count": task.ErrorCount,
		}
		
		if task.LastRun != nil {
			taskStats["last_run"] = task.LastRun.Format("2006-01-02 15:04:05")
		}
		
		if task.NextRun != nil {
			taskStats["next_run"] = task.NextRun.Format("2006-01-02 15:04:05")
		}
		
		tasks = append(tasks, taskStats)
	}
	
	stats["enabled_tasks"] = enabledCount
	stats["disabled_tasks"] = len(s.tasks) - enabledCount
	stats["tasks"] = tasks
	
	return stats
}