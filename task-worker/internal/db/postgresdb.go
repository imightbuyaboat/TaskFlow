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
	if status == "processing" {
		query += ", retries = retries + 1"
	}
	query += " where id = @task_id"

	if status == "processing" {
		query += " and retries < max_retries returning id"
	}

	args := pgx.NamedArgs{
		"status":  status,
		"task_id": taskID,
	}

	if status == "processing" {
		var id uuid.UUID
		if err := db.QueryRow(db.ctx, query, args).Scan(&id); err != nil {
			if err == pgx.ErrNoRows {
				return ErrMaxRetriesReached
			}
			return fmt.Errorf("failed to update task status: %v", err)
		}
		return nil
	}

	_, err := db.Exec(db.ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to update task status: %v", err)
	}

	return nil
}
