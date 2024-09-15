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

func TestGetUserBalance_AuthError(t *testing.T) {
	ctrl, _, handler := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.LoginKey, nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetUserBalance(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code %v, got %v", http.StatusInternalServerError, status)
	}
}

func TestGetUserBalance_MethodNotAllowed(t *testing.T) {
	ctrl, _, handler := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/api/user/balance", nil)
	rr := httptest.NewRecorder()

	handler.GetUserBalance(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %v, got %v", http.StatusMethodNotAllowed, status)
	}
}

func TestGetUserBalance_ServiceError(t *testing.T) {
	ctrl, mockService, handler := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.LoginKey, "user1")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	mockService.EXPECT().GetUserBalance(gomock.Any(), "user1").Return(nil, errors.New("service error"))

	handler.GetUserBalance(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code %v, got %v", http.StatusInternalServerError, status)
	}
}

func TestGetUserBalance_Success(t *testing.T) {
	ctrl, mockService, handler := setup(t)
	defer ctrl.Finish()

	balance := models.Balance{
		Current:   500,
		Withdrawn: 42,
	}
	bodyBytes, err := json.Marshal(balance)
	if err != nil {
		t.Fatalf("Failed to marshal balance: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.LoginKey, "user1")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	mockService.EXPECT().GetUserBalance(gomock.Any(), "user1").Return(&balance, nil)

	handler.GetUserBalance(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, status)
	}

	assert.JSONEq(t, string(bodyBytes), rr.Body.String())
}
