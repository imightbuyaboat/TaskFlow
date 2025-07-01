package worker

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type Queue interface {
	NewConsumerChannel() (*amqp.Channel, <-chan amqp.Delivery, error)
}
