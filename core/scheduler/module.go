package scheduler

import (
	"base/core/emitter"
	"base/core/logger"
	"base/core/module"
	"base/core/router"

	"gorm.io/gorm"
)

// Module represents the scheduler module
type Module struct {
	module.DefaultModule
	DB            *gorm.DB
	Scheduler     *Scheduler
	CronScheduler *CronScheduler // Simple cron scheduler
	Controller    *SchedulerController
	Logger        logger.Logger
}

// NewSchedulerModule creates a new scheduler module
func NewSchedulerModule(db *gorm.DB, routerGroup *router.RouterGroup, log logger.Logger, emitter *emitter.Emitter) module.Module {
	scheduler := NewScheduler(log)
	cronScheduler := NewCronScheduler(log)
	controller := NewSchedulerController(scheduler)

	m := &Module{
		DB:            db,
		Scheduler:     scheduler,
		CronScheduler: cronScheduler,
		Controller:    controller,
		Logger:        log,
	}

	return m
}

// Routes registers the scheduler routes
func (m *Module) Routes(router *router.RouterGroup) {
	schedulerGroup := router.Group("/scheduler")
	m.Controller.Routes(schedulerGroup)
}

// Start starts the scheduler
func (m *Module) Start() error {
	m.Logger.Info("Starting scheduler module")

	// Start both schedulers
	go m.Scheduler.Start()

	if err := m.CronScheduler.Start(); err != nil {
		return err
	}

	return nil
}

// Stop stops the scheduler
func (m *Module) Stop() error {
	m.Logger.Info("Stopping scheduler module")
	m.Scheduler.Stop()
	m.CronScheduler.Stop()
	return nil
}

// GetScheduler returns the scheduler instance
func (m *Module) GetScheduler() *Scheduler {
	return m.Scheduler
}

// GetCronScheduler returns the cron scheduler instance
func (m *Module) GetCronScheduler() *CronScheduler {
	return m.CronScheduler
}
