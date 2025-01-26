package taskqueue

// BuildTask builds a task from the given parameters
type TaskBuilder interface {
	SetLang(lang string) TaskBuilder
	SetCode(code string) TaskBuilder
	SetID(id string) TaskBuilder
	SetMaxRetry(maxRetry int) TaskBuilder
	SetRetryCnt(retryCnt int) TaskBuilder
	SetStatus(status string) TaskBuilder
	SetResult(result TaskResult) TaskBuilder
	SetError(err error) TaskBuilder
	Build() Task
}

// TaskBuilderImpl is an implementation of TaskBuilder
type TaskBuilderImpl struct {
	task *Task
}

// NewTaskBuilder creates a new TaskBuilder
func NewTaskBuilder() TaskBuilder {
	return &TaskBuilderImpl{
		task: newTaskBase(),
	}
}

func (b *TaskBuilderImpl) SetLang(lang string) TaskBuilder {
	b.task.Language = lang
	return b
}

func (b *TaskBuilderImpl) SetCode(code string) TaskBuilder {
	b.task.Code = code
	return b
}

func (b *TaskBuilderImpl) SetID(id string) TaskBuilder {
	b.task.TaskID = id
	return b
}

func (b *TaskBuilderImpl) SetMaxRetry(maxRetry int) TaskBuilder {
	b.task.MaxRetry = maxRetry
	return b
}

func (b *TaskBuilderImpl) SetRetryCnt(retryCnt int) TaskBuilder {
	b.task.RetryCnt = retryCnt
	return b
}

func (b *TaskBuilderImpl) SetStatus(status string) TaskBuilder {
	b.task.Status = status
	return b
}

func (b *TaskBuilderImpl) SetResult(result TaskResult) TaskBuilder {
	b.task.Result = result
	return b
}

func (b *TaskBuilderImpl) SetError(err error) TaskBuilder {
	b.task.Err = err
	return b
}

func (b *TaskBuilderImpl) Build() Task {
	// validate task
	if b.task.Language == "" {
		panic("language is required")
	}

	if b.task.Code == "" {
		panic("code is required")
	}

	if b.task.TaskID == "" {
		panic("taskID is required")
	}

	if b.task.MaxRetry < 0 {
		panic("maxRetry must be greater than or equal to 0")
	}
	return *b.task
}
