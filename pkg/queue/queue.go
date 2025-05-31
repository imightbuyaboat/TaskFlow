package queue

import (
	"github.com/imightbuyaboat/TaskFlow/pkg/task"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Queue interface {
	Publish(t *task.Task) error
	NewConsumerChannel() (*amqp.Channel, <-chan amqp.Delivery, error)
}
