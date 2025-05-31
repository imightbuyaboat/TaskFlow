package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/imightbuyaboat/TaskFlow/pkg/postgres"
	"github.com/imightbuyaboat/TaskFlow/pkg/task"
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

func (db *PostgresDB) GetPostponedTasks() ([]task.Task, error) {
	query := `select * from tasks
	where (status = 'postponed' and now() >= run_at)
	or (status = 'queued' and extract(hour from (now() - run_at)) > 1)`

	rows, err := db.Query(db.ctx, query)
	if err != nil {
		if err == pgx.ErrNoRows {
			return []task.Task{}, nil
		}
		return nil, fmt.Errorf("failed to select postponed tasks: %v", err)
	}

	tasks := []task.Task{}
	for rows.Next() {
		t := task.Task{}
		err := rows.Scan(
			&t.ID, &t.UserID, &t.Type, &t.Payload,
			&t.Status, &t.Retries, &t.MaxRetries,
			&t.RunAt, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %v", err)
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (db *PostgresDB) UpdateStatusOfTask(taskID uuid.UUID, status string) error {
	query := "update tasks set status = @status where id = @task_id"
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
