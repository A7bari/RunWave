package store

import (
	"github.com/A7bari/RunWave/internal/taskqueue"
	"github.com/A7bari/RunWave/internal/types"
)

type Store interface {
	SaveResult(types.TaskOutput) error
	GetResult(string) (types.TaskOutput, error)
	DeleteResult(string, version int) error

	CreateTask(taskqueue.Task) (string, error)
	UpdateTask(taskqueue.Task) error
	GetTask(taskID string) (taskqueue.Task, error)

	Close() error
}
