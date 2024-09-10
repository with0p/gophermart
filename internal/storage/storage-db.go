package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"

	"github.com/jackc/pgx/v5/pgconn"
	customerror "github.com/with0p/gophermart/internal/custom-error"
	"github.com/with0p/gophermart/internal/models"
)

type StorageDB struct {
	db *sql.DB
}

func NewStorageDB(ctx context.Context, db *sql.DB) (*StorageDB, error) {
	err := initTable(ctx, db)
	if err != nil {
		return nil, err
	}
	return &StorageDB{db: db}, nil
}

func initTable(ctx context.Context, db *sql.DB) error {
	tr, errTr := db.BeginTx(ctx, nil)
	if errTr != nil {
		return errTr
	}

	queryUserTable := `
    CREATE TABLE IF NOT EXISTS user_auth (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        login TEXT NOT NULL,
        password TEXT NOT NULL
    );`
	tr.ExecContext(ctx, queryUserTable)
	tr.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS user_login_index ON user_auth (login)`)

	queryOrderTable := `
	CREATE TABLE IF NOT EXISTS user_orders (
		order_id INTEGER PRIMARY KEY,
		status TEXT NOT NULL,
		accrual INTEGER DEFAULT -1,
		uploaded_at TIMESTAMPTZ,
		user_id UUID NOT NULL
	);`
	tr.ExecContext(ctx, queryOrderTable)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return tr.Commit()
	}
}

func (s *StorageDB) CreateUser(ctx context.Context, login string, password string) error {
	query := `
		INSERT INTO user_auth (login, password)
		VALUES ($1, $2)`

	_, errInsert := s.db.ExecContext(ctx, query, login, password)

	if errInsert != nil {
		var pgErr *pgconn.PgError
		if errors.As(errInsert, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			errInsert = customerror.ErrUniqueKeyConstrantViolation
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return errInsert
	}
}

func (s *StorageDB) ValidateUser(ctx context.Context, login string, password string) error {
	query := `SELECT id::text FROM user_auth WHERE login = $1 AND password = $2`

	var userID string
	err := s.db.QueryRowContext(ctx, query, login, password).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return customerror.ErrNoSuchUser
		}
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (s *StorageDB) GetUserID(ctx context.Context, login string) (uuid.UUID, error) {
	query := `SELECT id FROM user_auth WHERE login = $1`

	var userID uuid.UUID
	err := s.db.QueryRowContext(ctx, query, login).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, customerror.ErrNoSuchUser
		}
		return uuid.Nil, err
	}

	select {
	case <-ctx.Done():
		return uuid.Nil, ctx.Err()
	default:
		return userID, nil
	}
}

func (s *StorageDB) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	query := `SELECT order_id, status, accrual, user_id, uploaded_at::text FROM user_orders WHERE order_id = $1`
	var order models.Order
	err := s.db.QueryRowContext(ctx, query, orderID).Scan(&order.OrderID, &order.Status, &order.Accrual, &order.UserID, &order.UploadDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return &order, nil
	}
}

func (s *StorageDB) AddOrder(ctx context.Context, userID uuid.UUID, status models.OrderStatus, orderID string) error {
	query := `
	INSERT INTO user_orders (order_id, status, uploaded_at, user_id)
    VALUES ($1, $2, NOW(), $3);`
	_, err := s.db.ExecContext(ctx, query, orderID, status, userID)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return err
	}
}

func (s *StorageDB) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	query := `SELECT order_id, status, accrual, uploaded_at::text FROM user_orders WHERE user_id = $1 ORDER BY uploaded_at DESC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order

	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.OrderID, &order.Status, &order.Accrual, &order.UploadDate)
		if err != nil {
			continue
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return orders, nil
	}
}
