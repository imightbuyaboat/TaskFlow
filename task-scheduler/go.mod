module github.com/imightbuyaboat/TaskFlow/task-scheduler

go 1.23.0

replace github.com/imightbuyaboat/TaskFlow/pkg => ../pkg

require (
	github.com/google/uuid v1.6.0
	github.com/imightbuyaboat/TaskFlow/pkg v0.0.0-00010101000000-000000000000
	github.com/jackc/pgx/v5 v5.7.5
	github.com/joho/godotenv v1.5.1
	go.uber.org/zap v1.27.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/text v0.24.0 // indirect
)
