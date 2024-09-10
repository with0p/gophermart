package service

import (
	"context"
	"strconv"
	"time"

	"github.com/theplant/luhn"
	"github.com/with0p/gophermart/internal/custom-error"
	"github.com/with0p/gophermart/internal/models"
	"github.com/with0p/gophermart/internal/storage"
	"github.com/with0p/gophermart/internal/utils"
)

type ServiceGophermart struct {
	storage storage.Storage
}

func NewServiceGophermart(currentStorage storage.Storage) ServiceGophermart {
	return ServiceGophermart{
		storage: currentStorage,
	}
}

func (s *ServiceGophermart) RegisterUser(ctx context.Context, login string, password string) error {
	return s.storage.CreateUser(ctx, login, utils.HashPassword(password))
}

func (s *ServiceGophermart) AuthenticateUser(ctx context.Context, login string, password string) error {
	return s.storage.ValidateUser(ctx, login, utils.HashPassword(password))
}

func (s *ServiceGophermart) AddOrder(ctx context.Context, login string, orderID string) error {
	orderIDInt, errInt := strconv.ParseInt(orderID, 10, 64)
	if errInt != nil || !luhn.Valid(int(orderIDInt)) {
		return customerror.ErrWrongOrderFormat
	}

	userID, errUser := s.storage.GetUserID(ctx, login)
	if errUser != nil {
		return errUser
	}

	order, errOrder := s.storage.GetOrder(ctx, orderID)
	if errOrder != nil {
		return errOrder
	}

	if order == nil {
		return s.storage.AddOrder(ctx, userID, models.StatusNew, orderID)
	}

	if order.UserID != userID {
		return customerror.ErrAnotherUserOrder
	}
	return customerror.ErrAreadyAdded
}

func (s *ServiceGophermart) GetUserOrders(ctx context.Context, login string) ([]models.Order, error) {
	userID, errUser := s.storage.GetUserID(ctx, login)
	if errUser != nil {
		return nil, errUser
	}

	orders, err := s.storage.GetUserOrders(ctx, userID)
	var ordersFormatted = make([]models.Order, len(orders))
	for i, order := range orders {
		orderUploadedParsed, err := time.Parse("2006-01-02 15:04:05.999999-07", order.UploadDate)
		if err != nil {
			continue
		}
		loc, errLoc := time.LoadLocation("Europe/Moscow")
		if errLoc != nil {
			continue
		}
		orderUploadedParsed = orderUploadedParsed.In(loc)
		ordersFormatted[i] = order
		ordersFormatted[i].UploadDate = orderUploadedParsed.Format(time.RFC3339)
	}

	return ordersFormatted, err
}
