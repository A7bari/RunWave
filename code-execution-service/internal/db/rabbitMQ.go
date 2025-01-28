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
	ExchangeName string
	consumerOnce sync.Once // Ensures only one consumer is created
	queues       map[string]string
}

// runtime check RabbitMQConn implements QueuePublisher and.QueueConsumer interfaces
var _ taskqueue.QueuePublisher = (*RabbitMQConn)(nil)
var _ taskqueue.QueueConsumer = (*RabbitMQConn)(nil)

// GetRabbitMQConn initializes the RabbitMQConn with a RabbitMQ connection
// only initializes the RabbitMQConn once
func GetRabbitMQConn() (*RabbitMQConn, error) {
	var err error
	rabbitMQConnOnce.Do(func() {
		fmt.Println("Setting up RabbitMQ connection")
		addr := config.GetConfig().QueueAdrr
		languages := config.GetConfig().Languages
		exchangeName := "Tasks_Direct"

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

		if qErr := sendChan.ExchangeDeclare(
			exchangeName, // name
			"direct",     // type
			true,         // durable
			false,        // auto-deleted
			false,        // internal
			false,        // no-wait
			nil,          // arguments
		); qErr != nil {
			sendChan.Close() // Rollback on error
			conn.Close()
			err = fmt.Errorf("failed to declare a queue: %w", qErr)
			return
		}

		rabbitMQConn = &RabbitMQConn{
			conn:         conn,
			sendChan:     sendChan,
			ExchangeName: exchangeName,
			queues:       make(map[string]string),
		}

		rabbitMQConn.setupQueues(languages)
	})
	return rabbitMQConn, err
}

// PublishTask sends a task message to the RabbitMQ queue
// implement QueuePublisher interface
func (r *RabbitMQConn) PublishTask(task taskqueue.QueueMsg, lang string) error {
	// Publish the task message
	err := r.sendChan.Publish(
		r.ExchangeName, // Exchange
		lang,           // Routing key
		false,          // Mandatory
		false,          // Immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // Make message persistent
			ContentType:  task.ContentType,
			Body:         task.Body,
		})

	if err != nil {
		return err
	}
	return nil
}

// ConsumeTasks receives a task message from the RabbitMQ queue
// implement.QueueConsumer interface
func (r *RabbitMQConn) ConsumeTasks(lang string, taskChannel chan<- taskqueue.QueueMsg) error {
	var err error

	// Create a channel to consume messages
	ch, err := r.conn.Channel()
	if err != nil {
		err = fmt.Errorf("failed to create channel: %w", err)
		return err
	}

	// // Declare and bind the queue
	// q, err := ch.QueueDeclare(
	// 	"", false, true, false, false, nil,
	// )
	// if err != nil {
	// 	err = fmt.Errorf("failed to declare queue: %w", err)
	// 	return err
	// }

	// err = ch.QueueBind(q.Name, lang, r.ExchangeName, false, nil)
	// if err != nil {
	// 	err = fmt.Errorf("failed to bind queue: %w", err)
	// 	return err
	// }

	qn, ok := r.queues[lang]
	if !ok {
		err = fmt.Errorf("queue for language %s not found", lang)
		return err
	}

	msgs, err := ch.Consume(qn, "", true, false, false, false, nil)
	if err != nil {
		err = fmt.Errorf("failed to consume messages: %w", err)
		return err
	}

	// Goroutine to handle incoming messages
	go func() {
		defer func() {
			close(taskChannel)
			ch.Close()
		}()
		for msg := range msgs {
			taskChannel <- taskqueue.QueueMsg{
				Body:        msg.Body,
				ContentType: msg.ContentType,
			}
		}
	}()

	fmt.Println("Started Consuming tasks for language:", lang)
	return nil
}

func (r *RabbitMQConn) Close() error {
	if err := r.sendChan.Close(); err != nil {
		return err
	}
	return r.conn.Close()
}

func (r *RabbitMQConn) setupQueues(langs []string) error {
	for _, lang := range langs {
		fmt.Println("Setting up queue for language:", lang)
		qn := lang + "_queue"
		_, err := r.sendChan.QueueDeclare(
			qn,    // name
			true,  // durable
			false, // autoDelete
			false, // exclusive
			false, // noWait
			nil,   // args
		)
		if err != nil {
			return err
		}

		if err := r.sendChan.QueueBind(
			qn,             // queue name
			lang,           // routing key
			r.ExchangeName, // exchange
			false,          // noWait
			nil,            // args
		); err != nil {
			return err
		}

		r.queues[lang] = qn
	}
	return nil
}
