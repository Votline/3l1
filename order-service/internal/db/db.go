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

	gc "orders/internal/graceful"
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
	for i := 0; i < 11; i++ {
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
type Order struct {
	ID         string    `db:"id"`
	UserID     string    `db:"user_id"`
	UserRl     string    `db:"user_role"`
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
		Columns("id", "user_id", "user_role", "status",
			"service_url", "target_url", "order_type", "quantity").
		Values(order.ID, order.UserID, order.UserRl, "processed",
			order.ServiceURL, order.TargetURL, order.OrderType,
			order.Quantity).
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

func (r *Repo) OrderInfo(id, userID string) (*Order, error) {
	query, args, err := r.bd.
		Select("user_id", "user_role", "status", "target_url",
			"service_url", "order_type", "created_at", "updated_at").
		From("orders").
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		r.log.Error("Failed to create select query", zap.Error(err))
		return nil, err
	}

	order := Order{}
	if err := r.db.QueryRowx(query, args...).StructScan(&order); err != nil {
		r.log.Error("Failed to execute select query", zap.Error(err))
		return nil, err
	}

	return &order, err
}

func (r *Repo) getUserInfo(orderID string, tx *sqlx.Tx) (string, string, error) {
	query, args, err := r.bd.
		Select("user_id", "user_role").
		From("orders").
		Where(sq.Eq{"id": orderID}).
		ToSql()
	if err != nil {
		r.log.Error("Failed to create select user info query", zap.Error(err))
		return "", "", err
	}

	var result struct {
		userID string `db:"user_id"`
		userRl string `db:"user_role"`
	}

	if err := tx.Get(&result, query, args...); err != nil {
		r.log.Error("Failed to execute select user info query", zap.Error(err))
		return "", "", err
	}
	return result.userID, result.userRl, err
}

func (r *Repo) DelOrder(id, userID, role string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		r.log.Error("Failed to begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	q := r.bd.Delete("orders").Where(sq.Eq{"id": id})

	if role != "admin" {
		q = q.Where(sq.Eq{"user_id": userID})
	} else {
		delID, delRole, err := r.getUserInfo(id, tx)
		if err != nil {
			r.log.Error("Failed to find deleting order info", zap.Error(err))
			return err
		}
		if userID != delID && role == delRole {
			r.log.Error("Admin cannot delete admin's order", zap.Error(err))
			return errors.New("Admin cannot delete admin's order")
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
		r.log.Error("Failed to commin transaction", zap.Error(err))
		return err
	}

	return nil
}
