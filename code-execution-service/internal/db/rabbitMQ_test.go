package db

import (
	"fmt"
	"testing"

	"github.com/A7bari/RunWave/internal/taskqueue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublisher(t *testing.T) {
	queue, err := GetRabbitMQConn()
	require.NoError(t, err)

	tMsg := taskqueue.NewTask("123", "python", "print('Hello, World!')")

	emsg, err := tMsg.Encode()
	require.NoError(t, err)

	queue.PublishTask(taskqueue.QueueMsg{
		Body: emsg,
	}, "python")

	// Consume the task
	consumer := make(chan taskqueue.QueueMsg)
	err = queue.ConsumeTasks("python", consumer)
	require.NoError(t, err)

	// Verify the task

	for {
		taskMsg := <-consumer
		task, err := taskqueue.Decode(taskMsg.Body)
		require.NoError(t, err)
		fmt.Println(task)
		if task.TaskID == tMsg.TaskID {
			assert.Equal(t, tMsg.TaskID, task.TaskID)
			assert.Equal(t, tMsg.Language, task.Language)
			assert.Equal(t, tMsg.Code, task.Code)
			break
		}
	}

	// Clean up
	return

}
