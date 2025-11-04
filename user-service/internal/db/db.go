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
			time.Sleep(2*time.Second)
			continue
		}
		r.log.Debug("Successfully connected")
		return db
	}
	r.log.Fatal("Couldn't connect to DB")
	return nil
}

type User struct {
	ID   string
	Role string
}

func (r *Repo) AddUser(id, userName, role, pswd string) error {
	query, args, err := r.bd.
		Insert("users").
		Columns("id", "user_name", "role", "pswd").
		Values(id, userName, role, pswd).
		ToSql()
	if err != nil {
		r.log.Error("Failed to create query", zap.Error(err))
		return err
	}

	if _, err := r.db.Exec(query, args...); err != nil {
		r.log.Error("Failed to execute query", zap.Error(err))
		return err
	}

	return nil
}

func (r *Repo) LogUser(userName, pswd string) (*User, error){
	query, args, err := r.bd.
		Select("users").
		Columns("id", "role").
		Where(sq.Eq{"user_name":userName}).
		Where(sq.Eq{"password":pswd}).
		ToSql()
	if err != nil {
		r.log.Error("Failed to create query", zap.Error(err))
		return nil, err
	}

	var data User
	if err := r.db.QueryRow(query, args...).Scan(&data.ID, &data.Role); err != nil {
		r.log.Error("Failed to execute query", zap.Error(err))
		return nil, err
	}

	return &data, nil
}
