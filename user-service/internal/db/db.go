package db

import (
	"os"
	"time"

	"go.uber.org/zap"
	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
	sq "github.com/Masterminds/squirrel"
)

type Repo struct {
	log *zap.Logger
	db *sqlx.DB
	bd sq.StatementBuilderType
}
func NewRepo(log *zap.Logger) *Repo {
	r := &Repo{
		log: log,
		bd: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
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
			time.Sleep(1*time.Second)
			continue
		}
		r.log.Debug("Successfully connected")
		return db
	}
	r.log.Fatal("Couldn't connect to DB")
	return nil
}

func (r *Repo) AddUser(name, role, pswd string) error {
	query, args, err := r.bd.
		Insert("users").
		Columns("name", "role", "pswd").
		Values(name, role, pswd).
		ToSql()
	if err != nil {
		r.log.Error("Failed to create error", zap.Error(err))
		return err
	}

	if _, err := r.db.Exec(query, args...); err != nil {
		r.log.Error("Failed to execute query", zap.Error(err))
		return err
	}

	return nil
}
