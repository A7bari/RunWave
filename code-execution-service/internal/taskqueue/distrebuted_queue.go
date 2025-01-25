package taskqueue

type QueueMsg struct {
	Body        []byte
	ContentType string // MIME content type
}

type DqueuePublisher interface {
	// SendTask sends a task to the queue
	PublishTask(task QueueMsg) error

	// Close closes the connection to the queue
	Close() error
}

type DqueueReceiver interface {
	// ReceiveTask receives a task from the queue
	ReceiveTasks() (<-chan QueueMsg, error)

	// Close closes the connection to the queue
	Close() error
}
