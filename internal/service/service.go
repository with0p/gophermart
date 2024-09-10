package service

import (
	"context"

	"github.com/with0p/gophermart/internal/models"
)

type Service interface {
	RegisterUser(ctx context.Context, login string, password string) error
	AuthenticateUser(ctx context.Context, login string, password string) error
	AddOrder(ctx context.Context, login string, orderID string) error
	GetUserOrders(ctx context.Context, login string) ([]models.Order, error)
}
