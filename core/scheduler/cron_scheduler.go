package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"base/core/logger"

	"github.com/robfig/cron/v3"
)

// CronScheduler manages cron-based scheduled tasks
type CronScheduler struct {
	cron      *cron.Cron
	tasks     map[string]*CronTask
	mu        sync.RWMutex
	logger    logger.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	running   bool
}

// CronTask represents a task with cron scheduling
type CronTask struct {
	Name         string
	Description  string
	CronExpr     string
	Handler      TaskHandler
	Enabled      bool
	LastRun      *time.Time
	NextRun      *time.Time
	RunCount     int64
	ErrorCount   int64
	EntryID      cron.EntryID
}

// NewCronScheduler creates a new cron-based scheduler
func NewCronScheduler(log logger.Logger) *CronScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create cron with seconds precision
	c := cron.New(cron.WithSeconds())
	
	return &CronScheduler{
		cron:    c,
		tasks:   make(map[string]*CronTask),
		logger:  log,
		ctx:     ctx,
		cancel:  cancel,
		running: false,
	}
}

// Start starts the cron scheduler
func (cs *CronScheduler) Start() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	if cs.running {
		return fmt.Errorf("cron scheduler is already running")
	}
	
	cs.logger.Info("Starting cron scheduler")
	cs.cron.Start()
	cs.running = true
	
	return nil
}

// Stop stops the cron scheduler
func (cs *CronScheduler) Stop() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	if !cs.running {
		return
	}
	
	cs.logger.Info("Stopping cron scheduler")
	cs.cron.Stop()
	cs.running = false
	cs.cancel()
}

// RegisterTask registers a new cron task
func (cs *CronScheduler) RegisterTask(task *CronTask) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	if !task.Enabled {
		cs.tasks[task.Name] = task
		cs.logger.Info("Registered disabled cron task", 
			logger.String("name", task.Name),
			logger.String("expression", task.CronExpr))
		return nil
	}
	
	// Wrap handler to update task statistics
	wrappedHandler := func() {
		now := time.Now()
		cs.logger.Info("Executing cron task", 
			logger.String("name", task.Name),
			logger.String("description", task.Description))
		
		err := task.Handler(cs.ctx)
		
		cs.mu.Lock()
		task.LastRun = &now
		task.RunCount++
		if err != nil {
			task.ErrorCount++
		}
		// Update next run time
		cs.updateNextRunTime(task)
		cs.mu.Unlock()
		
		if err != nil {
			cs.logger.Error("Cron task execution failed",
				logger.String("name", task.Name),
				logger.String("error", err.Error()))
		} else {
			cs.logger.Info("Cron task completed successfully", 
				logger.String("name", task.Name))
		}
	}
	
	// Add job to cron
	entryID, err := cs.cron.AddFunc(task.CronExpr, wrappedHandler)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}
	
	task.EntryID = entryID
	cs.updateNextRunTime(task)
	cs.tasks[task.Name] = task
	
	nextRunStr := "unknown"
	if task.NextRun != nil {
		nextRunStr = task.NextRun.Format("2006-01-02 15:04:05")
	}
	
	cs.logger.Info("Registered cron task", 
		logger.String("name", task.Name),
		logger.String("description", task.Description),
		logger.String("expression", task.CronExpr),
		logger.String("next_run", nextRunStr))
	
	return nil
}

// updateNextRunTime updates the next run time for a task
func (cs *CronScheduler) updateNextRunTime(task *CronTask) {
	if task.EntryID == 0 || cs.cron == nil {
		return
	}
	
	entry := cs.cron.Entry(task.EntryID)
	if entry.ID != 0 && !entry.Next.IsZero() {
		task.NextRun = &entry.Next
	}
}

// UnregisterTask removes a task from the scheduler
func (cs *CronScheduler) UnregisterTask(name string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	task, exists := cs.tasks[name]
	if !exists {
		return fmt.Errorf("task not found: %s", name)
	}
	
	if task.EntryID != 0 {
		cs.cron.Remove(task.EntryID)
	}
	
	delete(cs.tasks, name)
	cs.logger.Info("Unregistered cron task", logger.String("name", name))
	
	return nil
}

// EnableTask enables a task
func (cs *CronScheduler) EnableTask(name string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	task, exists := cs.tasks[name]
	if !exists {
		return fmt.Errorf("task not found: %s", name)
	}
	
	if task.Enabled {
		return nil // Already enabled
	}
	
	// Remove old entry if exists
	if task.EntryID != 0 {
		cs.cron.Remove(task.EntryID)
	}
	
	// Add new entry
	return cs.registerTaskInternal(task)
}

// DisableTask disables a task
func (cs *CronScheduler) DisableTask(name string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	
	task, exists := cs.tasks[name]
	if !exists {
		return fmt.Errorf("task not found: %s", name)
	}
	
	if !task.Enabled {
		return nil // Already disabled
	}
	
	if task.EntryID != 0 {
		cs.cron.Remove(task.EntryID)
		task.EntryID = 0
	}
	
	task.Enabled = false
	task.NextRun = nil
	
	cs.logger.Info("Disabled cron task", logger.String("name", name))
	return nil
}

// RunTaskNow executes a task immediately
func (cs *CronScheduler) RunTaskNow(name string) error {
	cs.mu.RLock()
	task, exists := cs.tasks[name]
	cs.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("task not found: %s", name)
	}
	
	cs.logger.Info("Running cron task manually", logger.String("name", name))
	
	now := time.Now()
	err := task.Handler(cs.ctx)
	
	cs.mu.Lock()
	task.LastRun = &now
	task.RunCount++
	if err != nil {
		task.ErrorCount++
	}
	cs.mu.Unlock()
	
	return err
}

// GetTask returns a task by name
func (cs *CronScheduler) GetTask(name string) (*CronTask, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	task, exists := cs.tasks[name]
	return task, exists
}

// GetAllTasks returns all registered tasks
func (cs *CronScheduler) GetAllTasks() []*CronTask {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	tasks := make([]*CronTask, 0, len(cs.tasks))
	for _, task := range cs.tasks {
		// Update next run time before returning
		cs.updateNextRunTime(task)
		tasks = append(tasks, task)
	}
	
	return tasks
}

// GetStats returns scheduler statistics
func (cs *CronScheduler) GetStats() map[string]interface{} {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	
	enabledTasks := 0
	disabledTasks := 0
	
	tasks := make([]map[string]interface{}, 0, len(cs.tasks))
	for _, task := range cs.tasks {
		if task.Enabled {
			enabledTasks++
		} else {
			disabledTasks++
		}
		
		taskInfo := map[string]interface{}{
			"name":         task.Name,
			"description":  task.Description,
			"enabled":      task.Enabled,
			"cron_expr":    task.CronExpr,
			"run_count":    task.RunCount,
			"error_count":  task.ErrorCount,
		}
		
		if task.LastRun != nil {
			taskInfo["last_run"] = task.LastRun.Format("2006-01-02 15:04:05")
		}
		
		if task.NextRun != nil {
			taskInfo["next_run"] = task.NextRun.Format("2006-01-02 15:04:05")
		}
		
		tasks = append(tasks, taskInfo)
	}
	
	return map[string]interface{}{
		"running":        cs.running,
		"total_tasks":    len(cs.tasks),
		"enabled_tasks":  enabledTasks,
		"disabled_tasks": disabledTasks,
		"tasks":          tasks,
	}
}

// registerTaskInternal registers a task (internal method - assumes lock is held)
func (cs *CronScheduler) registerTaskInternal(task *CronTask) error {
	// Wrap handler to update task statistics
	wrappedHandler := func() {
		now := time.Now()
		cs.logger.Info("Executing cron task", 
			logger.String("name", task.Name),
			logger.String("description", task.Description))
		
		err := task.Handler(cs.ctx)
		
		cs.mu.Lock()
		task.LastRun = &now
		task.RunCount++
		if err != nil {
			task.ErrorCount++
		}
		cs.updateNextRunTime(task)
		cs.mu.Unlock()
		
		if err != nil {
			cs.logger.Error("Cron task execution failed",
				logger.String("name", task.Name),
				logger.String("error", err.Error()))
		} else {
			cs.logger.Info("Cron task completed successfully", 
				logger.String("name", task.Name))
		}
	}
	
	// Add job to cron
	entryID, err := cs.cron.AddFunc(task.CronExpr, wrappedHandler)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}
	
	task.EntryID = entryID
	task.Enabled = true
	cs.updateNextRunTime(task)
	
	cs.logger.Info("Enabled cron task", 
		logger.String("name", task.Name),
		logger.String("expression", task.CronExpr))
	
	return nil
}