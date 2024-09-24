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

func TestGetUserWithdrawals_AuthError(t *testing.T) {
	ctrl, _, handler := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.LoginKey, nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetUserWithdrawals(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code %v, got %v", http.StatusInternalServerError, status)
	}
}

func TestGetUserWithdrawals_NoContent(t *testing.T) {
	ctrl, mockService, handler := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.LoginKey, "user1")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	mockService.EXPECT().GetUserWithdrawals(gomock.Any(), "user1").Return(make([]models.Withdrawal, 0), nil)

	handler.GetUserWithdrawals(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("Expected status code %v, got %v", http.StatusNoContent, status)
	}
}

func TestGetUserWithdrawals_MethodNotAllowed(t *testing.T) {
	ctrl, _, handler := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/api/user/withdrawals", nil)
	rr := httptest.NewRecorder()

	handler.GetUserWithdrawals(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %v, got %v", http.StatusMethodNotAllowed, status)
	}
}

func TestGetUserWithdrawals_ServiceError(t *testing.T) {
	ctrl, mockService, handler := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.LoginKey, "user1")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	mockService.EXPECT().GetUserWithdrawals(gomock.Any(), "user1").Return(nil, errors.New("service error"))

	handler.GetUserWithdrawals(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code %v, got %v", http.StatusInternalServerError, status)
	}
}

func TestGetUserWithdrawals_Success(t *testing.T) {
	ctrl, mockService, handler := setup(t)
	defer ctrl.Finish()

	withdrawals := []models.Withdrawal{
		{OrderID: "2377225624", Sum: 500, ProcessedAt: "2020-12-09T16:09:57+03:00"},
	}
	bodyBytes, err := json.Marshal(withdrawals)
	if err != nil {
		t.Fatalf("Failed to marshal withdrawals: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.LoginKey, "user1")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	mockService.EXPECT().GetUserWithdrawals(gomock.Any(), "user1").Return(withdrawals, nil)

	handler.GetUserWithdrawals(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, status)
	}

	assert.JSONEq(t, string(bodyBytes), rr.Body.String())
}
