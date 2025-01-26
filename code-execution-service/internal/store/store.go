package store

import (
	"context"
	"time"

	"github.com/A7bari/RunWave/internal/taskqueue"
)

type TaskFilter struct {
	Lang        *string
	Completed   *bool
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Limit       int
	Offset      int
}

type Store interface {
	SaveResult(ctx context.Context, taskID string, taskResult taskqueue.TaskResult) error
	CreateTask(ctx context.Context, task taskqueue.Task) (string, error)
	UpdateTask(ctx context.Context, task taskqueue.Task) error
	GetTaskById(ctx context.Context, taskID string) (taskqueue.Task, error)
	ListTasks(ctx context.Context, filter TaskFilter) ([]taskqueue.Task, error)
	UpdateTaskFields(ctx context.Context, id string, fields map[string]interface{}) error
	Close() error
}
