package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/with0p/gophermart/internal/auth"
	customerror "github.com/with0p/gophermart/internal/custom-error"
	"github.com/with0p/gophermart/internal/models"
)

func TestMakeWithdrawal_AuthError(t *testing.T) {
	ctrl, _, handler := setup(t)
	defer ctrl.Finish()

	withdrawalData := OrderWithdrawalData{
		OrderID: "2377225624",
		Sum:     500,
	}
	reqBody, err := json.Marshal(withdrawalData)
	if err != nil {
		t.Fatalf("Failed to marshal withdrawal data: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/user/withdrawals", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.LoginKey, nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.MakeWithdrawal(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code %v, got %v", http.StatusInternalServerError, status)
	}
}
func TestMakeWithdrawal_MethodNotAllowed(t *testing.T) {
	ctrl, _, handler := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	rr := httptest.NewRecorder()

	handler.MakeWithdrawal(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %v, got %v", http.StatusMethodNotAllowed, status)
	}
}

func TestMakeWithdrawal_ServiceError_InsufficientBalance(t *testing.T) {
	ctrl, mockService, handler := setup(t)
	defer ctrl.Finish()

	withdrawalData := OrderWithdrawalData{
		OrderID: "2377225624",
		Sum:     500,
	}
	reqBody, err := json.Marshal(withdrawalData)
	if err != nil {
		t.Fatalf("Failed to marshal withdrawal data: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/user/withdrawals", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.LoginKey, "user1")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	mockService.EXPECT().MakeWithdrawal(gomock.Any(), "user1", models.OrderID("2377225624"), float32(500.0)).Return(customerror.ErrInsufficientBalance)

	handler.MakeWithdrawal(rr, req)

	if status := rr.Code; status != http.StatusPaymentRequired {
		t.Errorf("Expected status code %v, got %v", http.StatusPaymentRequired, status)
	}
}

func TestMakeWithdrawal_ServiceError(t *testing.T) {
	ctrl, mockService, handler := setup(t)
	defer ctrl.Finish()

	withdrawalData := OrderWithdrawalData{
		OrderID: "2377225624",
		Sum:     500,
	}
	reqBody, err := json.Marshal(withdrawalData)
	if err != nil {
		t.Fatalf("Failed to marshal withdrawal data: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/user/withdrawals", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.LoginKey, "user1")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	mockService.EXPECT().MakeWithdrawal(gomock.Any(), "user1", models.OrderID("2377225624"), float32(500.0)).Return(customerror.ErrWrongOrderFormat)

	handler.MakeWithdrawal(rr, req)

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("Expected status code %v, got %v", http.StatusUnprocessableEntity, status)
	}
}
func TestMakeWithdrawal_Success(t *testing.T) {
	ctrl, mockService, handler := setup(t)
	defer ctrl.Finish()

	withdrawalData := OrderWithdrawalData{
		OrderID: "2377225624",
		Sum:     500,
	}
	reqBody, err := json.Marshal(withdrawalData)
	if err != nil {
		t.Fatalf("Failed to marshal withdrawal data: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/user/withdrawals", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.LoginKey, "user1")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	mockService.EXPECT().MakeWithdrawal(gomock.Any(), "user1", models.OrderID("2377225624"), float32(500.0)).Return(nil)

	handler.MakeWithdrawal(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, status)
	}
}
