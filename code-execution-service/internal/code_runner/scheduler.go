package coderunner

import (
	"context"
	"fmt"
	"time"

	"github.com/A7bari/RunWave/internal/services"
	"github.com/A7bari/RunWave/internal/taskqueue"
)

// SchedulerCallback is a callback function for the scheduler
// it takes the language and the task ID as arguments
type SchedulerCallback func(language, taskId string)

// SchedulerOpts is the options for the scheduler
type SchedulerOpts struct {
	// PodManager is the pod manager to manage pods
	PodManager PodManager

	// consumer is the consumer to consume the task
	Consumer taskqueue.QueueConsumer

	// TaskService is the task service to manage tasks
	TaskService services.ITaskService

	// event callback: called when a task is started
	OnStartExec SchedulerCallback

	// event callback: called when a task is completed
	OnEndExec SchedulerCallback

	// event callback: called when a task fails
	OnFailExec SchedulerCallback
}

type Scheduler struct {
	languages []string
	stop      chan struct{}
	ctx       context.Context
	SchedulerOpts
}

func NewScheduler(conf SchedulerOpts, Languages ...string) *Scheduler {
	if len(Languages) == 0 {
		panic("[Scheduler] No languages provided")
	}
	s := &Scheduler{
		SchedulerOpts: conf,
		stop:          make(chan struct{}),
		ctx:           context.Background(),
		languages:     Languages,
	}
	return s
}

// StartWorkers starts the workers
// the workers are the pods that are created to execute code
// for each language, a goroutine is started to consume tasks
func (r *Scheduler) StartWorkers() {
	for _, lang := range r.languages {
		go r.startWorkers(lang)
	}
}

// startWorkers starts the workers for a specific language
// the workers are the pods that are created to execute code to that language
// the worker consumes tasks from the task queue subscribed to the language
func (r *Scheduler) startWorkers(language string) {
	fmt.Println("[Scheduler] Starting workers for language:", language)
	tasks := make(chan taskqueue.QueueMsg)
	go r.Consumer.ConsumeTasks(language, tasks)

	for {
		select {
		// If the scheduler is stopped, close the worker channel and return
		case <-r.stop:
			close(r.stop)
			return
		case taskMsg := <-tasks:

			// decode the task message
			task, err := taskqueue.Decode(taskMsg.Body)

			// If a worker is available, run the task
			worker := r.PodManager.ConsumePod(language)
			if err != nil {
				continue
			}

			go func() {
				// Defer the task completion
				// If there is an error, mark the task as failed
				// and requeue the task
				err = nil
				defer func() {
					if err != nil {
						r.TaskService.MarkTaskAsFailed(task.TaskID)
						if r.OnFailExec != nil {
							r.OnFailExec(language, task.TaskID)
						}
					}
				}()

				// If the task is started, mark it as started
				r.TaskService.MarkTaskAsRunning(task.TaskID)
				if r.OnStartExec != nil {
					r.OnStartExec(language, task.TaskID)
				}

				command, err := FormatCommand(task.Language, task.Code)
				if err != nil {
					return
				}

				// Run the command in the worker
				res, err := r.PodManager.Run(command, worker, 5*time.Second)
				if err != nil {
					return
				}

				// If the task is completed, mark it as completed
				r.TaskService.MarkTaskAsSuccess(task.TaskID, res.Output, res.IsError)
				if r.OnEndExec != nil {
					r.OnEndExec(language, task.TaskID)
				}
			}()
		}
	}
}

// Stop stops the scheduler
func (r *Scheduler) Stop() {
	r.stop <- struct{}{}
}
