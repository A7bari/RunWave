package taskqueue

// TaskQueue is an interface for task queue
// Can be Implemented by any db or in-memory queue
type TaskQueue interface {
	// AddTask adds a task to the queue
	AddTask(task Task)

	// GetTask gets a task from the queue
	GetTask() Task
}
