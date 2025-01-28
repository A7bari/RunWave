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
		taskQueue = &TaskQueue{tasks: sync.Map{}}
	})
	return taskQueue
}

// runtime check TaskQueue implements QueuePublisher and QueueConsumer interfaces
var _ taskqueue.QueuePublisher = (*TaskQueue)(nil)
var _ taskqueue.QueueConsumer = (*TaskQueue)(nil)

type TaskQueue struct {
	tasks sync.Map
}

// implement QueuePublisher interface
func (tq *TaskQueue) PublishTask(task taskqueue.QueueMsg, lang string) error {
	taskCh, ok := tq.tasks.Load(lang)
	if !ok {
		taskCh := make(chan taskqueue.QueueMsg, 100)
		tq.tasks.Store(lang, taskCh)
	}
	taskCh.(chan taskqueue.QueueMsg) <- task
	return nil
}

// implement TaskQueue interface
func (tq *TaskQueue) ConsumeTasks(lang string, taskChannel chan<- taskqueue.QueueMsg) error {
	taskCh, ok := tq.tasks.Load(lang)
	if !ok {
		taskCh := make(chan taskqueue.QueueMsg, 100)
		tq.tasks.Store(lang, taskCh)
	}

	go func() {
		for task := range taskCh.(chan taskqueue.QueueMsg) {
			taskChannel <- task
		}
	}()
	return nil
}

func (tq *TaskQueue) Close() error {
	tq.tasks.Range(func(key, value interface{}) bool {
		close(value.(chan taskqueue.QueueMsg))
		return true
	})

	return nil
}
