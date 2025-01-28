package taskqueue

type QueueMsg struct {
	Body        []byte
	ContentType string // MIME content type
}

type QueuePublisher interface {
	// SendTask sends a task to the queue
	PublishTask(task QueueMsg, lang string) error

	// Close closes the connection to the queue
	Close() error
}

type QueueConsumer interface {
	// ConsumeTask receives a task from the queue
	ConsumeTasks(lang string, taskChannel chan<- QueueMsg) error

	// Close closes the connection to the queue
	Close() error
}
