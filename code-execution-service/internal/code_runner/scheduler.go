package coderunner

import (
	"context"
	"time"

	"github.com/A7bari/RunWave/internal/taskqueue"
)

type SchedulerCallback func(taskqueue.Task)

// SchedulerOpts is the options for the scheduler
type SchedulerOpts struct {
	// PodManager is the pod manager to manage pods
	PodManager PodManager

	// TaskQueue is the task queue to get tasks from
	TaskQueue taskqueue.TaskQueue

	// event callback: called when a task is started
	OnStartExec SchedulerCallback

	// event callback: called when a task is completed
	OnEndExec SchedulerCallback

	// event callback: called when a task fails
	OnFailExec SchedulerCallback
}

type Scheduler struct {
	stop chan struct{}
	ctx  context.Context
	SchedulerOpts
}

func NewScheduler(conf SchedulerOpts) *Scheduler {
	s := &Scheduler{
		SchedulerOpts: conf,
		stop:          make(chan struct{}),
		ctx:           context.Background(),
	}
	return s
}

// StartWorkers starts the workers
// the workers are the pods that are created to execute code
// for each worker, a task is dequeued from the task queue and executed
func (r *Scheduler) StartWorkers() {
	for {
		select {
		// If the scheduler is stopped, close the worker channel and return
		case <-r.stop:
			close(r.stop)
			return
		default:
			// If a worker is available, run the task
			worker := r.PodManager.ConsumePod()
			task := r.TaskQueue.GetTask()
			go func() {
				// Run the task
				task.Running()
				command, err := FormatCommand(task)
				if err != nil {
					task.Failed(err)
					if r.OnFailExec != nil {
						r.OnFailExec(task)
					}
					return
				}

				if r.OnStartExec != nil {
					r.OnStartExec(task)
				}
				res, err := r.PodManager.Run(command, worker, 5*time.Second)

				// If there is an error, requeue the task
				if err != nil {
					task.Failed(err)
					if r.OnFailExec != nil {
						r.OnFailExec(task)
					}

					// Requeue the task
					r.TaskQueue.AddTask(task)
					return
				}

				// If the task is completed, mark it as completed
				task.Completed()
				task.SetResult(res.Output, res.IsError)
				if r.OnEndExec != nil {
					r.OnEndExec(task)
				}
			}()
		}
	}
}

// Stop stops the scheduler
func (r *Scheduler) Stop() {
	r.stop <- struct{}{}
}
