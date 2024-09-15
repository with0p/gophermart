package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/with0p/gophermart/internal/auth"
	customerror "github.com/with0p/gophermart/internal/custom-error"
	"github.com/with0p/gophermart/internal/models"
)

type OrderWithdrawalData struct {
	OrderID models.OrderID `json:"order"`
	Sum     int            `json:"sum"`
}

func (h *HandlerUserAPI) MakeWithdrawal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Not a POST requests", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("content-type") != "application/json" {
		http.Error(w, "Not a \"application/json\" content-type", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	login, errLogin := auth.GetLoginFromRequestContext(ctx)
	if errLogin != nil {
		http.Error(w, errLogin.Error(), http.StatusInternalServerError)
		return
	}

	var orderWithdrawal OrderWithdrawalData
	err := json.NewDecoder(r.Body).Decode(&orderWithdrawal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	errW := h.service.MakeWithdrawal(ctx, login, orderWithdrawal.OrderID, orderWithdrawal.Sum)

	var statusCode int

	switch {
	case errW == nil:
		statusCode = http.StatusOK
	case errors.Is(errW, customerror.ErrWrongOrderFormat):
		statusCode = http.StatusUnprocessableEntity
	case errors.Is(errW, customerror.ErrInsufficientBalance):
		statusCode = http.StatusPaymentRequired
	default:
		statusCode = http.StatusInternalServerError
	}

	w.WriteHeader(statusCode)
}
