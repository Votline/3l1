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
	ID         string    `db:"id"`
	UserID     string    `db:"user_id"`
	Status     string    `db:"status"`
	TargetURL  string    `db:"target_url"`
	ServiceURL string    `db:"service_url"`
	OrderType  string    `db:"order_type"`
	Quantity   int32     `db:"quantity"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

func (r *Repo) AddOrder(order *Order) error {
	query, args, err := r.bd.
		Insert("orders").
		Columns("id", "user_id", "status", "service_url",
				"target_url", "order_type", "quantity").
		Values(order.ID, order.UserID, "processed", order.ServiceURL,
				order.TargetURL, order.OrderType, order.Quantity).
		ToSql()
	if err != nil {
		r.log.Error("Failed to create query", zap.Error(err))
		return err
	}

	if _, err := r.db.Exec(query, args...); err != nil {
		r.log.Error("Failed to execute query")
		return err
	}

	return nil
}

func (r *Repo) OrderInfo(id string) (*Order, error) {
	query, args, err := r.bd.
		Select("user_id", "status", "target_url", "service_url",
				"order_type", "created_at", "updated_at").
		From("orders").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		r.log.Error("Failed to create query")
		return nil, err
	}

	order := Order{}
	if err := r.db.QueryRowx(query, args...).StructScan(&order); err != nil {
		r.log.Error("Failed to execute query")
		return nil, err
	}

	return &order, err
}
