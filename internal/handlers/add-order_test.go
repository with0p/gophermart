package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/with0p/gophermart/internal/auth"
	customerror "github.com/with0p/gophermart/internal/custom-error"
	"github.com/with0p/gophermart/internal/models"
)

func TestAddOrder_MethodNotPost(t *testing.T) {
	ctrl, _, handler := setupWithQueue(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	rr := httptest.NewRecorder()

	handler.AddOrder(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %v, got %v", http.StatusMethodNotAllowed, status)
	}
}

func TestAddOrder_WrongContentType(t *testing.T) {
	ctrl, _, handler := setupWithQueue(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", nil)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.AddOrder(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, status)
	}
}

func TestAddOrder_AuthError(t *testing.T) {
	ctrl, _, handler := setupWithQueue(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903"))
	req.Header.Set("Content-Type", "text/plain")
	ctx := context.WithValue(req.Context(), auth.LoginKey, nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.AddOrder(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code %v, got %v", http.StatusInternalServerError, status)
	}
}

func TestAddOrder_ErrAnotherUserOrder(t *testing.T) {
	ctrl, mockService, handler := setupWithQueue(t)
	defer ctrl.Finish()

	mockService.EXPECT().AddOrder(gomock.Any(), "login", models.OrderID("1230")).Return(customerror.ErrAnotherUserOrder)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("1230"))
	req.Header.Set("Content-Type", "text/plain")
	ctx := context.WithValue(req.Context(), auth.LoginKey, "login")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.AddOrder(rr, req)

	if status := rr.Code; status != http.StatusConflict {
		t.Errorf("Expected status code %v, got %v", http.StatusConflict, status)
	}
}

func TestAddOrder_ErrWrongOrderFormat(t *testing.T) {
	ctrl, mockService, handler := setupWithQueue(t)
	defer ctrl.Finish()

	mockService.EXPECT().AddOrder(gomock.Any(), "login", models.OrderID("12345678903")).Return(customerror.ErrWrongOrderFormat)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903"))
	req.Header.Set("Content-Type", "text/plain")
	ctx := context.WithValue(req.Context(), auth.LoginKey, "login")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.AddOrder(rr, req)

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("Expected status code %v, got %v", http.StatusUnprocessableEntity, status)
	}
}

func TestAddOrder_ErrAlreadyAdded(t *testing.T) {
	ctrl, mockService, handler := setupWithQueue(t)
	defer ctrl.Finish()

	mockService.EXPECT().AddOrder(gomock.Any(), "login", models.OrderID("12345678903")).Return(customerror.ErrAlreadyAdded)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("12345678903"))
	req.Header.Set("Content-Type", "text/plain")
	ctx := context.WithValue(req.Context(), auth.LoginKey, "login")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.AddOrder(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, status)
	}
}

func TestAddOrder_Success(t *testing.T) {
	ctrl, mockService, handler := setupWithQueue(t)
	defer ctrl.Finish()

	mockService.EXPECT().AddOrder(gomock.Any(), "login", models.OrderID("1230")).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader("1230"))
	req.Header.Set("Content-Type", "text/plain")
	ctx := context.WithValue(req.Context(), auth.LoginKey, "login")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.AddOrder(rr, req)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("Expected status code %v, got %v", http.StatusAccepted, status)
	}
}
