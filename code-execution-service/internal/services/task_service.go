package services

import (
	"context"

	"github.com/A7bari/RunWave/internal/store"
	"github.com/A7bari/RunWave/internal/taskqueue"
)

// ITaskService is an interface for task service
type ITaskService interface {
	// CreateTask creates a new task
	CreateTask(task taskqueue.Task) (string, error)

	// GetTask gets a task by ID
	GetTask(taskID string) (taskqueue.Task, error)

	// UpdateTask updates a task
	UpdateTask(task taskqueue.Task) error

	// SaveResult saves the result of a task
	SaveResult(taskId, output string, IsErr bool) error

	// get result of a task
	GetResult(taskId string) (string, bool, error)

	// change task's status to failed
	MarkTaskAsFailed(taskID string) error

	// change task's status to success
	MarkTaskAsSuccess(taskID string, output string, IsErr bool) error

	// change task's status to running
	MarkTaskAsRunning(taskID string) error

	// change task's status to pending
	MarkTaskAsPending(taskID string) error
}

// runtime check for TaskService struct
var _ ITaskService = (*TaskService)(nil)

// TaskService is a struct to hold task service details
type TaskService struct {
	db  store.Store
	ctx context.Context
}

// NewTaskService creates a new task service
func NewTaskService(store store.Store) *TaskService {
	return &TaskService{db: store, ctx: context.Background()}
}

// CreateTask creates a new task
func (ts *TaskService) CreateTask(task taskqueue.Task) (string, error) {
	return ts.db.CreateTask(ts.ctx, task)
}

// GetTask gets a task by ID
func (ts *TaskService) GetTask(taskID string) (taskqueue.Task, error) {
	return ts.db.GetTaskById(ts.ctx, taskID)
}

// UpdateTask updates a task
func (ts *TaskService) UpdateTask(task taskqueue.Task) error {
	return ts.db.UpdateTask(ts.ctx, task)
}

// SaveResult saves the result of a task
func (ts *TaskService) SaveResult(taskId, output string, IsErr bool) error {
	return ts.db.SaveResult(ts.ctx, taskId, taskqueue.TaskResult{Output: output, IsError: IsErr})
}

// get result of a task
func (ts *TaskService) GetResult(taskId string) (string, bool, error) {
	task, err := ts.GetTask(taskId)
	if err != nil {
		return "", false, err
	}
	return task.Result.Output, task.Result.IsError, nil
}

// change task's status to failed
func (ts *TaskService) MarkTaskAsFailed(taskID string) error {
	f := map[string]interface{}{"status": "failed"}
	return ts.db.UpdateTaskFields(ts.ctx, taskID, f)
}

// change task's status to success
func (ts *TaskService) MarkTaskAsSuccess(taskID string, output string, IsErr bool) error {
	err := ts.db.SaveResult(ts.ctx, taskID, taskqueue.TaskResult{Output: output, IsError: IsErr})
	if err != nil {
		return err
	}

	f := map[string]interface{}{"status": "completed"}
	return ts.db.UpdateTaskFields(ts.ctx, taskID, f)
}

// change task's status to running
func (ts *TaskService) MarkTaskAsRunning(taskID string) error {
	f := map[string]interface{}{"status": "running"}
	return ts.db.UpdateTaskFields(ts.ctx, taskID, f)
}

// change task's status to pending
func (ts *TaskService) MarkTaskAsPending(taskID string) error {
	f := map[string]interface{}{"status": "pending"}
	return ts.db.UpdateTaskFields(ts.ctx, taskID, f)
}
