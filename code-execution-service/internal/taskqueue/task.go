package taskqueue

// Task is an interface for task
type ITask interface {
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

	// Retry Return if the task should be retried
	Retry() bool
}

type TaskResult struct {
	Output  string `json:"output"`
	IsError bool   `json:"is_error"`
}

// Task is a struct to hold task details
type Task struct {
	TaskID   string     `json:"task_id"`
	Language string     `json:"lang"`
	Code     string     `json:"code"`
	Status   string     `json:"status"`
	Result   TaskResult `json:"result"`
	Err      error      `json:"error"`
	MaxRetry int        `json:"max_retry"`
	RetryCnt int        `json:"retry_cnt"`
}

// runtime check
var _ ITask = (*Task)(nil)

// NewTask creates a new task
// a private function to create a new task with default values
// tobe used in the builder
func NewTask(taskId, code, language string) *Task {
	return &Task{
		TaskID:   taskId,
		Code:     code,
		Language: language,
		Status:   "pending",
		MaxRetry: 1,
		RetryCnt: 0,
		Result:   TaskResult{},
		Err:      nil,
	}
}

func newTaskBase() *Task {
	return &Task{
		Status:   "pending",
		MaxRetry: 1,
		RetryCnt: 0,
	}
}

// implement Task interface
// SetResult sets the task result and calls the OnSuccess callback
func (t *Task) SetResult(output string, isError bool) {
	t.Result = TaskResult{
		Output:  output,
		IsError: isError,
	}

	t.Completed()
}

// implement Task interface
// Retry Return if the task should be retried
// and calls the OnRetry callback
func (t *Task) Retry() bool {
	return t.RetryCnt < t.MaxRetry
}

func (t *Task) Running() {
	t.setStatus("running")
}

func (t *Task) Completed() {
	t.setStatus("completed")
}

func (t *Task) Pending() {
	t.setStatus("pending")
}

func (t *Task) Failed(err error) {
	t.Err = err
	t.setStatus("failed")
}

func (t *Task) setStatus(status string) {
	t.Status = status
}
