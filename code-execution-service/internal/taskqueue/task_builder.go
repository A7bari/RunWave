package taskqueue

// BuildTask builds a task from the given parameters
type TaskBuilder interface {
	SetLang(lang string) TaskBuilder
	SetCode(code string) TaskBuilder
	SetID(id string) TaskBuilder
	SetMaxRetry(maxRetry int) TaskBuilder
	RegisterCallback(event TaskEvent, callback Callback) TaskBuilder
	Build() Task
}

// TaskBuilderImpl is an implementation of TaskBuilder
type TaskBuilderImpl struct {
	task *TaskImp
}

// NewTaskBuilder creates a new TaskBuilder
func NewTaskBuilder() TaskBuilder {
	return &TaskBuilderImpl{
		task: &TaskImp{},
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

func (b *TaskBuilderImpl) RegisterCallback(event TaskEvent, callback Callback) TaskBuilder {
	b.task.RegisterCallback(event, callback)
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

	b.task.call(OnCreated)
	return b.task
}
