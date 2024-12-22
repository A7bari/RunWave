package taskqueue

import (
	"fmt"
)

// Task is an interface for task
type Task interface {
	// Running sets the task status to running
	Running()

	// Completed sets the task status to completed
	Completed()

	// Pending sets the task status to pending
	Pending()

	// Failed sets the task status to failed
	Failed(err error)

	// SetResult sets the task result
	SetResult(output string, isError bool)

	// GetResult returns the task result
	GetResult() (string, bool)

	// GetTaskID returns the task ID
	GetTaskID() string

	// GetLanguage returns the task language
	GetLanguage() string

	// GetCode returns the task code
	GetCode() string

	// GetStatus returns the task status
	GetStatus() string

	// GetError returns the task error
	GetError() error

	// RegisterCallback registers a callback for the task
	RegisterCallback(event TaskEvent, callback Callback) error

	// UnregisterCallback unregisters a callback for the task
	UnregisterCallback(event TaskEvent)

	// Retry Return if the task should be retried
	Retry() bool
}

// Callback is a function type for task callbacks
type Callback func(task Task)

// TaskCallbacksOpts is a struct to hold task callbacks
type TaskEvent int

const (
	OnCreated TaskEvent = iota
	OnChanged
	OnSuccess
	OnRetry
	OnFail
)

type TaskResult struct {
	Output  string
	IsError bool
}

// Task is a struct to hold task details
type TaskImp struct {
	taskID    string
	language  string
	code      string
	status    string
	result    TaskResult
	err       error
	callbacks map[TaskEvent]Callback
	maxRetry  int
	retryCnt  int
}

// runtime check
var _ Task = (*TaskImp)(nil)

// NewTask creates a new task
// a private function to create a new task with default values
// tobe used in the builder
func newTask() *TaskImp {
	return &TaskImp{
		callbacks: make(map[TaskEvent]Callback),
		status:    "pending",
		maxRetry:  1,
		retryCnt:  0,
	}
}

// implement Task interface
// SetResult sets the task result and calls the OnSuccess callback
func (t *TaskImp) SetResult(output string, isError bool) {
	t.result = TaskResult{
		Output:  output,
		IsError: isError,
	}

	t.Completed()

	t.call(OnSuccess)
}

// implement Task interface
// Retry Return if the task should be retried
// and calls the OnRetry callback
func (t *TaskImp) Retry() bool {
	if t.retryCnt >= t.maxRetry {
		return false
	}
	t.call(OnRetry)
	return true
}

// implement Task interface
// Register a callback for the task exepct for OnCreated
// onCreated event can be added only at the creation of the task
func (t *TaskImp) RegisterCallback(event TaskEvent, callback Callback) error {
	// onCreated event can be added only at the creation of the task
	if event == OnCreated {
		panic("task: onCreated event can be added only in NewTask function")
	}

	if _, ok := t.callbacks[event]; ok {
		return fmt.Errorf("task: callback for event %v already exists", event)
	}

	t.callbacks[event] = callback
	return nil
}

// implement Task interface
// Unregister a callback for the task
func (t *TaskImp) UnregisterCallback(event TaskEvent) {
	delete(t.callbacks, event)
}

func (t *TaskImp) Running() {
	t.setStatus("running")
}

func (t *TaskImp) Completed() {
	t.setStatus("completed")
}

func (t *TaskImp) Pending() {
	t.setStatus("pending")
}

func (t *TaskImp) Failed(err error) {
	t.err = err
	t.setStatus("failed")
	t.call(OnFail)
}

func (t *TaskImp) GetResult() (string, bool) {
	return t.result.Output, t.result.IsError
}

func (t *TaskImp) GetTaskID() string {
	return t.taskID
}

func (t *TaskImp) GetLanguage() string {
	return t.language
}

func (t *TaskImp) GetCode() string {
	return t.code
}

func (t *TaskImp) GetStatus() string {
	return t.status
}

func (t *TaskImp) GetError() error {
	return t.err
}

func (t *TaskImp) setStatus(status string) {
	t.status = status
	t.call(OnChanged)
}

func (t *TaskImp) call(event TaskEvent) {
	if cb, ok := t.callbacks[event]; ok {
		cb(t)
	}
}
