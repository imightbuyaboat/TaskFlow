package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/imightbuyaboat/TaskFlow/pkg/postgres"
)

type PostgresDB struct {
	*pgxpool.Pool
	ctx context.Context
}

func NewPostgresDB(postgresURL string) (*PostgresDB, error) {
	ctx := context.Background()

	pool, err := postgres.NewPoolDB(ctx, postgresURL)
	if err != nil {
		return nil, err
	}

	return &PostgresDB{pool, ctx}, nil
}

func (db *PostgresDB) UpdateStatusOfTask(taskID uuid.UUID, status string) error {
	query := "update tasks set status = @status"
	if status == "in_progress" {
		query += ", retries = retries + 1"
	}
	query += " where id = @task_id"

	args := pgx.NamedArgs{
		"status":  status,
		"task_id": taskID,
	}

	_, err := db.Exec(db.ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to update task status: %v", err)
	}

	return nil
}
