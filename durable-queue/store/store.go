package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"durqueue/job"

	_ "github.com/mattn/go-sqlite3"
)

var ErrNotFound = errors.New("job not found")

type Store interface {
	Insert(ctx context.Context, value job.Job) error
	Get(ctx context.Context, id string) (job.Job, error)
	Delete(ctx context.Context, id string) error
}

type SqliteStore struct {
	Db *sql.DB
}

func NewSqliteStore() *SqliteStore {
	db := MustConnect()

	MustRunMigrations(db)

	return &SqliteStore{
		Db: db,
	}
}

func MustConnect() *sql.DB {
	db, err := sql.Open("sqlite3", "../db.sqlite3")
	if err != nil {
		panic("error initializing sqlite store")
	}

	return db
}

func MustRunMigrations(db *sql.DB) {
	initQuery := `
 	 CREATE TABLE IF NOT EXISTS jobs (
	   id PRIMARY KEY 
	 )
	`
	_, err := db.Exec(initQuery)
	if err != nil {
		panic("migrations failed")
	}
}

func (str *SqliteStore) Insert(ctx context.Context, value job.Job) error {
	id := strings.TrimSpace(value.ID)
	if id == "" {
		return errors.New("id is required")
	}

	query := `
	   INSERT INTO jobs (id)
	   VALUES (?)
	`
	_, err := str.Db.ExecContext(
		ctx, query, id,
	)
	if err != nil {
		return fmt.Errorf("error insert job %s : %w", id, err)
	}

	return nil
}

func (str *SqliteStore) Get(ctx context.Context, id string) (job.Job, error) {
	var res job.Job

	query := `
	   SELECT id 
	   FROM jobs WHERE id = ?
	`

	err := str.Db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(&res.ID)

	if errors.Is(err, sql.ErrNoRows) {
		return res, ErrNotFound
	}

	if err != nil {
		return res, fmt.Errorf("error get job %s: %w", id, err)
	}

	return res, nil
}

func (str *SqliteStore) Update(ctx context.Context, value job.Job) error {
	id := strings.TrimSpace(value.ID)
	if id == "" {
		return errors.New("error: id is required")
	}

	query := `
		UPDATE jobs
		SET id = ?
		WHERE id = ?
	`

	res, err := str.Db.ExecContext(ctx, query, id, id)
	if err != nil {
		return fmt.Errorf("error update job %s: %w", id, err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error check update job %s: %w", id, err)
	}

	if n == 0 {
		return ErrNotFound
	}

	return nil
}

func (str *SqliteStore) Delete(ctx context.Context, id string) error {
	parsedId := strings.TrimSpace(id)
	if parsedId == "" {
		return errors.New("id is required")
	}

	query := `
	   DELETE FROM jobs WHERE id = ? 
	`

	res, err := str.Db.ExecContext(
		ctx, query, id,
	)
	if err != nil {
		return fmt.Errorf("error delete job: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error check delete job: %w", err)
	}

	if n == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *SqliteStore) Close() error {
	return s.Db.Close()
}
