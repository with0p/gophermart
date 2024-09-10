package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/with0p/gophermart/internal/auth"
	"github.com/with0p/gophermart/internal/models"
)

func TestGetUserOrders_MethodNotAllowed(t *testing.T) {
	ctrl, _, handler := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", nil)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.GetUserOrders(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %v, got %v", http.StatusMethodNotAllowed, status)
	}
}

func TestGetUserOrders_ServiceError(t *testing.T) {
	ctrl, mockService, handler := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.LoginKey, "user1")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	mockService.EXPECT().GetUserOrders(gomock.Any(), "user1").Return(nil, errors.New("error"))

	handler.GetUserOrders(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code %v, got %v", http.StatusInternalServerError, status)
	}
}

func TestGetUserOrders_Success(t *testing.T) {
	ctrl, mockService, handler := setup(t)
	defer ctrl.Finish()

	orders := []models.Order{
		{OrderID: 1, Status: "completed", Accrual: 100, UploadDate: "1234"},
	}
	bodyBytes, err := json.Marshal(orders)
	if err != nil {
		t.Fatalf("Failed to marshal orders: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.LoginKey, "user1")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	mockService.EXPECT().GetUserOrders(gomock.Any(), "user1").Return(orders, nil)

	handler.GetUserOrders(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, status)
	}
	assert.JSONEq(t, string(bodyBytes), rr.Body.String())
}
