package db

import (
	"os"
	"time"
	"errors"
	"context"

	"go.uber.org/zap"
	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
	sq "github.com/Masterminds/squirrel"

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
	query, args, err := r.bd.
		Insert("users").
		Columns("id", "user_name", "role", "pswd").
		Values(id, userName, role, pswd).
		ToSql()
	if err != nil {
		r.log.Error("Failed to create insert query", zap.Error(err))
		return err
	}

	if _, err := r.db.Exec(query, args...); err != nil {
		r.log.Error("Failed to execute insert query", zap.Error(err))
		return err
	}

	return nil
}

func (r *Repo) LogUser(userName, pswd string) (*User, error) {
	query, args, err := r.bd.
		Select("id", "role", "pswd").
		From("users").
		Where(sq.Eq{"user_name": userName}).
		ToSql()
	if err != nil {
		r.log.Error("Failed to create select query", zap.Error(err))
		return nil, err
	}

	var data User
	if err := r.db.QueryRow(query, args...).Scan(
		&data.ID,
		&data.Role,
		&data.Pswd); err != nil {
		r.log.Error("Failed to execute select query", zap.Error(err))
		return nil, err
	}

	return &data, nil
}

func (r *Repo) getUserRole(userID string, tx *sqlx.Tx) (string, error) {
	query, args, err := r.bd.
		Select("role").
		From("users").
		Where(sq.Eq{"id": userID}).
		ToSql()
	if err != nil {
		r.log.Error("Failed to create select role query", zap.Error(err))
		return "", err
	}

	var role string
	if err := tx.Get(&role, query, args...); err != nil {
		r.log.Error("Failed to execute select role query", zap.Error(err))
		return "", err
	}

	return role, nil
}

func (r *Repo) DelUser(userID, role, delUserID string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		r.log.Error("Failed to create transaction", zap.Error(err))
		return err
	}

	q := r.bd.Delete("users").Where(sq.Eq{"id": delUserID})

	if userID != delUserID {
		if role != "admin" {
			r.log.Warn("User ID don't match",
				zap.String("uid", userID),
				zap.String("duig", delUserID))
			return errors.New("User ID's don't match")
		}
		delRole, err := r.getUserRole(delUserID, tx)
		if err != nil {
			r.log.Error("Failed to find deleting user's role", zap.Error(err))
			return errors.New("Couldn't find deleting user's role")
		}
		if delRole == "admin" {
			r.log.Warn("Admin cannot delete admin")
			return errors.New("Admin cannot delete admin")
		}
	}

	query, args, err := q.ToSql()
	if err != nil {
		r.log.Error("Failed to create delete query", zap.Error(err))
		return err
	}

	if _, err := tx.Exec(query, args...); err != nil {
		r.log.Error("Failed to execute delete query", zap.Error(err))
		return err
	}

	if err := tx.Commit(); err != nil {
		r.log.Error("Failed to commit transaction", zap.Error(err))
		return err
	}

	return nil
}
