package handlers

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/with0p/gophermart/internal/mock"
	"github.com/with0p/gophermart/internal/models"
)

func setup(t *testing.T) (*gomock.Controller, *mock.MockService, *HandlerUserAPI) {
	ctrl := gomock.NewController(t)
	mockService := mock.NewMockService(ctrl)
	h := &HandlerUserAPI{service: mockService}
	return ctrl, mockService, h
}

func setupWithQueue(t *testing.T) (*gomock.Controller, *mock.MockService, *HandlerUserAPI) {
	ctrl := gomock.NewController(t)
	mockService := mock.NewMockService(ctrl)
	queue := make(chan models.OrderID, 1)
	h := &HandlerUserAPI{service: mockService, queue: queue}
	return ctrl, mockService, h
}
