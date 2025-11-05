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
	for i := 0; i < 11; i++ {
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

type Order struct {
	ID string
	UserID string
	TargetURL string
	ServiceURL string
	OrderType string
	Quantity int32
}

func (r *Repo) AddOrder(order *Order) error {
	query, args, err := r.bd.
		Insert("orders").
		Columns("id", "user_id", "status", "smm_service_url",
				"target_url", "order_type", "quantity").
		Values(order.ID, order.UserID, "processed", order.ServiceURL,
				order.TargetURL, order.OrderType, order.Quantity).
		ToSql()
	if err != nil {
		r.log.Error("Failed to create query", zap.Error(err))
		return err
	}

	if _, err := r.db.Exec(query, args...); err != nil {
		r.log.Error("Faield to execute query")
		return err
	}

	return nil
}
