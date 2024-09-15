package service

import (
	"context"

	"github.com/with0p/gophermart/internal/models"
)

type Service interface {
	RegisterUser(ctx context.Context, login string, password string) error
	AuthenticateUser(ctx context.Context, login string, password string) error
	AddOrder(ctx context.Context, login string, orderID models.OrderID) error
	GetUserOrders(ctx context.Context, login string) ([]models.Order, error)
	ProcessOrders(queue chan models.OrderID, accrualAddr string)
	FeedQueue(queue chan models.OrderID)
	MakeWithdrawal(ctx context.Context, login string, orderID models.OrderID, amount float32) error
	GetUserBalance(ctx context.Context, login string) (*models.Balance, error)
	GetUserWithdrawals(ctx context.Context, login string) ([]models.Withdrawal, error)
}
