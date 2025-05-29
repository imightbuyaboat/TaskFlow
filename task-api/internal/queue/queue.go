package queue

import (
	"encoding/json"
	"fmt"

	"github.com/imightbuyaboat/TaskFlow/pkg/task"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Queue struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	queue      *amqp.Queue
}

func NewQueue(amqpURL string) (*Queue, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	q, err := ch.QueueDeclare(
		"tasks",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %v", err)
	}

	return &Queue{
		connection: conn,
		channel:    ch,
		queue:      &q,
	}, nil
}

func (q *Queue) Close() error {
	if err := q.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %v", err)
	}
	if err := q.connection.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %v", err)
	}
	return nil
}

func (q *Queue) Publish(t *task.Task) error {
	body, err := json.Marshal(&t)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %v", err)
	}

	err = q.channel.Publish(
		"",
		q.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish task: %v", err)
	}

	return nil
}
