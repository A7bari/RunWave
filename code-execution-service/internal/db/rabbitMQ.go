package db

import (
	"fmt"
	"sync"

	"github.com/A7bari/RunWave/internal/config"
	"github.com/A7bari/RunWave/internal/taskqueue"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	rabbitMQConn     *RabbitMQConn
	rabbitMQConnOnce sync.Once
)

type RabbitMQConn struct {
	conn         *amqp.Connection
	sendChan     *amqp.Channel
	receiveChan  *amqp.Channel
	queueName    string
	consumerOnce sync.Once // Ensures only one consumer is created
}

// runtime check RabbitMQConn implements DqueuePublisher and DqueueReceiver interfaces
var _ taskqueue.DqueuePublisher = (*RabbitMQConn)(nil)
var _ taskqueue.DqueueReceiver = (*RabbitMQConn)(nil)

// GetRabbitMQConn initializes the RabbitMQConn with a RabbitMQ connection
// only initializes the RabbitMQConn once
func GetRabbitMQConn() (*RabbitMQConn, error) {
	var err error
	rabbitMQConnOnce.Do(func() {
		addr := config.GetConfig().QueueAdrr
		queueName := config.GetConfig().QueueName

		conn, connErr := amqp.Dial(addr)
		if connErr != nil {
			err = fmt.Errorf("failed to connect to RabbitMQ: %w", connErr)
			return
		}

		sendChan, chErr := conn.Channel()
		if chErr != nil {
			conn.Close() // Rollback on error
			err = fmt.Errorf("failed to open a channel: %w", chErr)
			return
		}

		if _, qErr := sendChan.QueueDeclare(queueName, true, false, false, false, nil); qErr != nil {
			sendChan.Close() // Rollback on error
			conn.Close()
			err = fmt.Errorf("failed to declare a queue: %w", qErr)
			return
		}

		rabbitMQConn = &RabbitMQConn{
			conn:      conn,
			sendChan:  sendChan,
			queueName: queueName,
		}
	})
	return rabbitMQConn, err
}

// PublishTask sends a task message to the RabbitMQ queue
// implement DqueuePublisher interface
func (r *RabbitMQConn) PublishTask(task taskqueue.QueueMsg) error {
	// Publish the task message
	err := r.sendChan.Publish(
		"",          // Exchange
		r.queueName, // Routing key
		false,       // Mandatory
		false,       // Immediate
		amqp.Publishing{
			ContentType: task.ContentType,
			Body:        task.Body,
		})

	if err != nil {
		return err
	}
	return nil
}

// ReceiveTasks receives a task message from the RabbitMQ queue
// implement DqueueReceiver interface
func (r *RabbitMQConn) ReceiveTasks() (<-chan taskqueue.QueueMsg, error) {
	var err error
	taskChannel := make(chan taskqueue.QueueMsg)

	r.consumerOnce.Do(func() {
		ch, chErr := r.conn.Channel()
		if chErr != nil {
			err = chErr
			close(taskChannel) // Prevent potential hanging
			return
		}

		msgs, consumeErr := ch.Consume(
			r.queueName, "", true, false, false, false, nil,
		)
		if consumeErr != nil {
			err = consumeErr
			close(taskChannel)
			return
		}

		go func() {
			defer ch.Close() // Ensure the channel is closed when done
			for msg := range msgs {
				taskChannel <- taskqueue.QueueMsg{
					Body:        msg.Body,
					ContentType: msg.ContentType,
				}
			}
			close(taskChannel)
		}()
	})

	if err != nil {
		return nil, err
	}
	return taskChannel, nil
}

func (r *RabbitMQConn) Close() error {
	if err := r.sendChan.Close(); err != nil {
		return err
	}
	return r.conn.Close()
}
