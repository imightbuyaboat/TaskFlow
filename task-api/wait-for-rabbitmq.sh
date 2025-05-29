#!/bin/sh
set -e

host="rabbitmq"
port="${AMQP_PORT:-5672}"

echo "Waiting for RabbitMQ at $host:$port..."

while ! nc -z "$host" "$port"; do
  echo "RabbitMQ is unavailable - sleeping"
  sleep 1
done

echo "RabbitMQ is up - executing command"
exec "$@"