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
	task *TaskImp
}

// NewTaskBuilder creates a new TaskBuilder
func NewTaskBuilder() TaskBuilder {
	return &TaskBuilderImpl{
		task: newTaskBase(),
	}
}

func (b *TaskBuilderImpl) SetLang(lang string) TaskBuilder {
	b.task.language = lang
	return b
}

func (b *TaskBuilderImpl) SetCode(code string) TaskBuilder {
	b.task.code = code
	return b
}

func (b *TaskBuilderImpl) SetID(id string) TaskBuilder {
	b.task.taskID = id
	return b
}

func (b *TaskBuilderImpl) SetMaxRetry(maxRetry int) TaskBuilder {
	b.task.maxRetry = maxRetry
	return b
}

func (b *TaskBuilderImpl) SetRetryCnt(retryCnt int) TaskBuilder {
	b.task.retryCnt = retryCnt
	return b
}

func (b *TaskBuilderImpl) SetStatus(status string) TaskBuilder {
	b.task.status = status
	return b
}

func (b *TaskBuilderImpl) SetResult(result TaskResult) TaskBuilder {
	b.task.result = result
	return b
}

func (b *TaskBuilderImpl) SetError(err error) TaskBuilder {
	b.task.err = err
	return b
}

func (b *TaskBuilderImpl) Build() Task {
	// validate task
	if b.task.language == "" {
		panic("language is required")
	}

	if b.task.code == "" {
		panic("code is required")
	}

	if b.task.taskID == "" {
		panic("taskID is required")
	}

	if b.task.maxRetry < 0 {
		panic("maxRetry must be greater than or equal to 0")
	}
	return b.task
}
