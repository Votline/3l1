package db

import (
	"context"
	"fmt"
	"os"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	gc "users/internal/graceful"
)

type Repo struct {
	log *zap.Logger
	db  *sqlx.DB
	bd  sq.StatementBuilderType
}

func NewRepo(log *zap.Logger) *Repo {
	r := &Repo{
		log: log,
		bd:  sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
	r.db = r.initDB()

	return r
}
func (r *Repo) initDB() *sqlx.DB {
	var db *sqlx.DB
	var err error

	connStr := os.Getenv("POSTGRES_URL")
	for i := 0; i < 10; i++ {
		db, err = sqlx.Connect("postgres", connStr)
		if err != nil {
			r.log.Error("Connect PQ error", zap.Error(err))
			time.Sleep(2 * time.Second)
			continue
		}

		db.SetMaxOpenConns(20)
		db.SetMaxIdleConns(10)
		db.SetConnMaxLifetime(time.Hour)
		db.SetConnMaxIdleTime(10 * time.Minute)

		r.log.Debug("Successfully connected")
		return db
	}
	r.log.Fatal("Couldn't connect to DB")
	return nil
}
func (r *Repo) Stop(ctx context.Context) error {
	return gc.Shutdown(r.db.Close, ctx)
}

type User struct {
	ID   string
	Role string
	Pswd string
}

func (r *Repo) AddUser(id, userName, role, pswd string) error {
	const op = "UserPostgresRepository.AddUser"

	query, args, err := r.bd.
		Insert("users").
		Columns("id", "user_name", "role", "pswd").
		Values(id, userName, role, pswd).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: create query: %w", op, err)
	}

	if _, err := r.db.Exec(query, args...); err != nil {
		return fmt.Errorf("%s: execute query: %w", op, err)
	}

	return nil
}

func (r *Repo) LogUser(userName, pswd string) (*User, error) {
	const op = "UserPostgresRepository.LogUser"

	query, args, err := r.bd.
		Select("id", "role", "pswd").
		From("users").
		Where(sq.Eq{"user_name": userName}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: create query: %w", op, err)
	}

	var data User
	if err := r.db.QueryRow(query, args...).Scan(
		&data.ID,
		&data.Role,
		&data.Pswd); err != nil {
		return nil, fmt.Errorf("%s: execute query: %w", op, err)
	}

	return &data, nil
}

func (r *Repo) getUserRole(userID string, tx *sqlx.Tx) (string, error) {
	const op = "UserPostgresRepository.getUserRole"

	query, args, err := r.bd.
		Select("role").
		From("users").
		Where(sq.Eq{"id": userID}).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("%s: create query: %w", op, err)
	}

	var role string
	if err := tx.Get(&role, query, args...); err != nil {
		return "", fmt.Errorf("%s: execute query: %w", op, err)
	}

	return role, nil
}

func (r *Repo) DelUser(userID, role, delUserID string) error {
	const op = "UserPostgresRepository.DelUser"

	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("%s: create transaction: %w", op, err)
	}

	q := r.bd.Delete("users").Where(sq.Eq{"id": delUserID})

	if userID != delUserID {
		if role != "admin" {
			return fmt.Errorf("%s: match id's: %s", op, "User ID's don't match")
		}
		delRole, err := r.getUserRole(delUserID, tx)
		if err != nil {
			return fmt.Errorf("%s: get user role: %s", op, "Couldn't find deleting user's role")
		}
		if delRole == "admin" {
			return fmt.Errorf("%s: check role: %s", op, "Admin cannot delete admin")
		}
	}

	query, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("%s: create tx query: %w", op, err)
	}

	if _, err := tx.Exec(query, args...); err != nil {
		return fmt.Errorf("%s: execute tx query: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: commit transaction: %w", op, err)
	}

	return nil
}
