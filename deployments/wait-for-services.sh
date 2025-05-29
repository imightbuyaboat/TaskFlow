#!/bin/sh
set -e

wait_for_service() {
  local name="$1"
  local host="$2"
  local port="$3"

  echo "Waiting for $name at $host:$port..."

  while ! nc -z "$host" "$port"; do
    echo "$name is unavailable - sleeping"
    sleep 1
  done

  echo "$name is up!"
}

wait_for_service "RabbitMQ" "rabbitmq" "${AMQP_PORT:-5672}"
wait_for_service "PostgreSQL" "db" "${POSTGRES_PORT:-5432}"

echo "All services are up - executing command"
exec "$@"