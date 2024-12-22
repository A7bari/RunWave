package db

import (
	"sync"

	"github.com/A7bari/RunWave/internal/taskqueue"
)

var (
	taskQueue     *TaskQueue
	taskQueueOnce sync.Once
)

// initializes the TaskQueue with a configurable size
// only initializes the TaskQueue once
func GetInMemTaskQueue(size int) *TaskQueue {
	taskQueueOnce.Do(func() {
		taskQueue = &TaskQueue{tasks: make(chan taskqueue.Task, size)}
	})
	return taskQueue
}

type TaskQueue struct {
	tasks chan taskqueue.Task
}

// implement TaskQueue interface
func (tq *TaskQueue) AddTask(task taskqueue.Task) {
	tq.tasks <- task
}

// implement TaskQueue interface
func (tq *TaskQueue) GetTask() taskqueue.Task {
	return <-tq.tasks
}
