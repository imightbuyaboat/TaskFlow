package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/imightbuyaboat/TaskFlow/pkg/postgres"
	"github.com/imightbuyaboat/TaskFlow/pkg/task"
	"github.com/imightbuyaboat/TaskFlow/task-api/internal/user"
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

func (db *PostgresDB) CreateTask(t *task.Task) (*task.Task, error) {
	query := "insert into tasks	(id, user_id, type, payload, max_retries"
	values := "values (@id, @user_id, @type, @payload, @max_retries"

	args := pgx.NamedArgs{
		"id":          t.ID,
		"user_id":     t.UserID,
		"type":        t.Type,
		"payload":     t.Payload,
		"max_retries": t.MaxRetries,
	}

	if t.RunAt != nil {
		args["run_at"] = t.RunAt
		args["status"] = "postponed"
		query += ", run_at, status"
		values += ", @run_at, @status"
	}

	query += ") " + values + ") returning *"

	var createdTask task.Task
	err := db.QueryRow(db.ctx, query, args).Scan(
		&createdTask.ID, &createdTask.UserID, &createdTask.Type, &createdTask.Payload,
		&createdTask.Status, &createdTask.Retries, &createdTask.MaxRetries,
		&createdTask.RunAt, &createdTask.CreatedAt, &createdTask.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert task into db: %v", err)
	}

	return &createdTask, nil
}

func (db *PostgresDB) GetTask(userID uint64, taskID uuid.UUID) (*task.Task, error) {
	query := "select * from tasks where id = @task_id and user_id = @user_id"
	args := pgx.NamedArgs{
		"task_id": taskID,
		"user_id": userID,
	}

	var t task.Task
	err := db.QueryRow(db.ctx, query, args).Scan(
		&t.ID, &t.UserID, &t.Type, &t.Payload,
		&t.Status, &t.Retries, &t.MaxRetries,
		&t.RunAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNoRows
		}
		return nil, fmt.Errorf("failed to select task from db: %v", err)
	}

	return &t, nil
}

func (db *PostgresDB) GetAllTasks(userID uint64) ([]task.Task, error) {
	query := "select * from tasks where user_id = @user_id"
	args := pgx.NamedArgs{
		"user_id": userID,
	}

	rows, err := db.Query(db.ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to select tasks from db: %v", err)
	}

	var tasks []task.Task
	for rows.Next() {
		t := task.Task{}
		err := rows.Scan(
			&t.ID, &t.UserID, &t.Type, &t.Payload,
			&t.Status, &t.Retries, &t.MaxRetries,
			&t.RunAt, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tasks from db: %v", err)
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (db *PostgresDB) CreateUser(u *user.User) (uint64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("failed to generate hash: %v", err)
	}

	query := "insert into users (email, password_hash) values (@email, @hash) returning id"
	args := pgx.NamedArgs{
		"email": u.Email,
		"hash":  hash,
	}

	var userID uint64
	err = db.QueryRow(db.ctx, query, args).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, ErrUserAlreadyExist
		}
		return 0, fmt.Errorf("failed to insert user into db: %v", err)
	}

	return userID, nil
}

func (db *PostgresDB) CheckUser(u *user.User) (uint64, error) {
	query := "select id, password_hash from users where email = @email"
	args := pgx.NamedArgs{
		"email": u.Email,
	}

	var id uint64
	var hash []byte
	err := db.QueryRow(db.ctx, query, args).Scan(&id, &hash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, ErrNoRows
		}
		return 0, fmt.Errorf("failed to select user from db: %v", err)
	}

	err = bcrypt.CompareHashAndPassword(hash, []byte(u.Password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return 0, ErrIncorrectPassword
		}
		return 0, fmt.Errorf("failed to compare hash and password: %v", err)
	}

	return id, nil
}
