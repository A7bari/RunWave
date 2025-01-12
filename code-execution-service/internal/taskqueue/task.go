package taskqueue

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
	taskID   string
	language string
	code     string
	status   string
	result   TaskResult
	err      error
	maxRetry int
	retryCnt int
}

// runtime check
var _ Task = (*TaskImp)(nil)

// NewTask creates a new task
// a private function to create a new task with default values
// tobe used in the builder
func newTask() *TaskImp {
	return &TaskImp{
		status:   "pending",
		maxRetry: 1,
		retryCnt: 0,
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
}

// implement Task interface
// Retry Return if the task should be retried
// and calls the OnRetry callback
func (t *TaskImp) Retry() bool {
	return t.retryCnt < t.maxRetry
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
}
