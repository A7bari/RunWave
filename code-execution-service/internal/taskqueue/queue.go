package taskqueue

// TaskQueue is an interface for task queue
// Can be Implemented by any db or in-memory queue
type TaskQueue interface {
	// AddTask adds a task to the queue
	AddTask(task Task)

	// GetTask gets a task from the queue
	GetTask() Task
}

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
}

// Callback is a function type for task callbacks
type Callback func(task Task)

// TaskCallbacksOpts is a struct to hold task callbacks
type TaskCallbacksOpts struct {
	// Called when a task is created
	OnCreated Callback

	// Called when a task status is changed
	OnChanged Callback

	// Called when a task result is set
	OnResult Callback
}

type TaskResult struct {
	Output  string
	IsError bool
}

// Task is a struct to hold task details
type TaskImp struct {
	taskID   string
	language string
	code     string
	status   string
	result   TaskResult
	err      error
	TaskCallbacksOpts
}

// runtime check
var _ Task = &TaskImp{}

func NewTask(taskID, language, code string, callbacks TaskCallbacksOpts) Task {
	t := &TaskImp{
		taskID:            taskID,
		language:          language,
		code:              code,
		status:            "pending",
		TaskCallbacksOpts: callbacks,
		err:               nil,
	}

	if t.OnCreated != nil {
		t.OnCreated(t)
	}

	return t
}

// implement Task interface
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
}

func (t *TaskImp) SetResult(output string, isError bool) {
	t.result = TaskResult{
		Output:  output,
		IsError: isError,
	}

	t.Completed()

	if t.OnResult != nil {
		t.OnResult(t)
	}
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
	if t.OnChanged != nil {
		t.OnChanged(t)
	}
}
