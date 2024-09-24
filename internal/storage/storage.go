package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/with0p/gophermart/internal/models"
)

type Storage interface {
	CreateUser(ctx context.Context, login, password string) error
	ValidateUser(ctx context.Context, login, password string) error
	GetUserID(ctx context.Context, login string) (uuid.UUID, error)
	GetOrder(ctx context.Context, orderID models.OrderID) (*models.Order, error)
	AddOrder(ctx context.Context, userID uuid.UUID, status models.OrderStatus, orderID models.OrderID) error
	UpdateOrder(ctx context.Context, orderID models.OrderID, status models.OrderStatus, accrual float32) error
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
	GetUnfinishedOrderIDs(ctx context.Context) ([]models.OrderID, error)
	GetUserAccrualBalance(ctx context.Context, userID uuid.UUID) (float32, error)
	AddWithdrawal(ctx context.Context, userID uuid.UUID, orderID models.OrderID, amount float32) error
	GetUserWithdrawalSum(ctx context.Context, userID uuid.UUID) (float32, error)
	GetUserWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.Withdrawal, error)
}
