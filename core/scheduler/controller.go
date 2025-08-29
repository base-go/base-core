package scheduler

import (
	"base/core/router"
	"net/http"
)

// SchedulerController provides HTTP endpoints for scheduler management
type SchedulerController struct {
	scheduler *Scheduler
}

// NewSchedulerController creates a new scheduler controller
func NewSchedulerController(scheduler *Scheduler) *SchedulerController {
	return &SchedulerController{
		scheduler: scheduler,
	}
}

// Routes registers scheduler endpoints
func (c *SchedulerController) Routes(router *router.RouterGroup) {
	// Routes are registered directly on the scheduler router group
	router.GET("/status", c.GetStatus)
	router.GET("/tasks", c.GetTasks)
	router.GET("/tasks/:name", c.GetTask)
	router.POST("/tasks/:name/run", c.RunTask)
	router.PUT("/tasks/:name/enable", c.EnableTask)
	router.PUT("/tasks/:name/disable", c.DisableTask)
	router.GET("/stats", c.GetStats)
}

// GetStatus returns scheduler status
func (c *SchedulerController) GetStatus(ctx *router.Context) error {
	stats := c.scheduler.GetStats()
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   stats,
	})
	return nil
}

// GetTasks returns all registered tasks
// @Summary Get all registered tasks
// @Tags Core/Scheduler
// @Description Returns a list of all registered tasks
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} []map[string]interface{}
// @Router /scheduler/tasks [get]
func (c *SchedulerController) GetTasks(ctx *router.Context) error {
	tasks := c.scheduler.GetAllTasks()

	taskList := make([]map[string]interface{}, 0, len(tasks))
	for _, task := range tasks {
		taskInfo := map[string]interface{}{
			"name":        task.Name,
			"description": task.Description,
			"enabled":     task.Enabled,
			"schedule":    task.Schedule.String(),
			"run_count":   task.RunCount,
			"error_count": task.ErrorCount,
		}

		if task.LastRun != nil {
			taskInfo["last_run"] = task.LastRun.Format("2006-01-02 15:04:05")
		}

		if task.NextRun != nil {
			taskInfo["next_run"] = task.NextRun.Format("2006-01-02 15:04:05")
		}

		taskList = append(taskList, taskInfo)
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   taskList,
	})
	return nil
}

// GetTask returns a specific task
// @Summary Get a specific task
// @Tags Core/Scheduler
// @Param name path string true "Task name"
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /scheduler/tasks/{name} [get]
func (c *SchedulerController) GetTask(ctx *router.Context) error {
	name := ctx.Param("name")

	task, exists := c.scheduler.GetTask(name)
	if !exists {
		ctx.JSON(http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"message": "Task not found",
		})
		return nil
	}

	taskInfo := map[string]interface{}{
		"name":        task.Name,
		"description": task.Description,
		"enabled":     task.Enabled,
		"schedule":    task.Schedule.String(),
		"run_count":   task.RunCount,
		"error_count": task.ErrorCount,
	}

	if task.LastRun != nil {
		taskInfo["last_run"] = task.LastRun.Format("2006-01-02 15:04:05")
	}

	if task.NextRun != nil {
		taskInfo["next_run"] = task.NextRun.Format("2006-01-02 15:04:05")
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   taskInfo,
	})
	return nil
}

// RunTask executes a task immediately
// @Summary Run a specific task immediately
// @Tags Core/Scheduler
// @Param name path string true "Task name"
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /scheduler/tasks/{name}/run [post]
func (c *SchedulerController) RunTask(ctx *router.Context) error {
	name := ctx.Param("name")

	err := c.scheduler.RunTaskNow(name)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Task executed successfully",
	})
	return nil
}

// EnableTask enables a task
// @Summary Enable a specific task
// @Tags Core/Scheduler
// @Param name path string true "Task name"
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /scheduler/tasks/{name}/enable [put]
func (c *SchedulerController) EnableTask(ctx *router.Context) error {
	name := ctx.Param("name")

	err := c.scheduler.EnableTask(name)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Task enabled successfully",
	})
	return nil
}

// DisableTask disables a task
// @Summary Disable a specific task
// @Tags Core/Scheduler
// @Param name path string true "Task name"
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /scheduler/tasks/{name}/disable [put]
func (c *SchedulerController) DisableTask(ctx *router.Context) error {
	name := ctx.Param("name")

	err := c.scheduler.DisableTask(name)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Task disabled successfully",
	})
	return nil
}

// GetStats returns detailed scheduler statistics
// @Summary Get scheduler statistics
// @Tags Core/Scheduler
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /scheduler/stats [get]
func (c *SchedulerController) GetStats(ctx *router.Context) error {
	stats := c.scheduler.GetStats()
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   stats,
	})
	return nil
}
