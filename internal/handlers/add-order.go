package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/with0p/gophermart/internal/auth"
	customerror "github.com/with0p/gophermart/internal/custom-error"
	"github.com/with0p/gophermart/internal/logger"
	"github.com/with0p/gophermart/internal/models"
)

func (h *HandlerUserAPI) AddOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Not a POST requests", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("content-type") != "text/plain" {
		http.Error(w, "Not a \"text/plain\" content-type", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	login, errLogin := auth.GetLoginFromRequestContext(ctx)
	if errLogin != nil {
		http.Error(w, errLogin.Error(), http.StatusInternalServerError)
		return
	}

	body, errRead := io.ReadAll(r.Body)
	defer r.Body.Close()
	if errRead != nil {
		http.Error(w, errRead.Error(), http.StatusInternalServerError)
		return
	}

	orderID := models.OrderID(string(body))

	errOrder := h.service.AddOrder(ctx, login, orderID)
	var statusCode int

	switch {
	case errOrder == nil:
		statusCode = http.StatusAccepted
		h.queue <- orderID
	case errors.Is(errOrder, customerror.ErrAnotherUserOrder):
		statusCode = http.StatusConflict
	case errors.Is(errOrder, customerror.ErrWrongOrderFormat):
		statusCode = http.StatusUnprocessableEntity
	case errors.Is(errOrder, customerror.ErrAlreadyAdded):
		statusCode = http.StatusOK
	default:
		statusCode = http.StatusInternalServerError
		logger.Error(errOrder)
	}

	w.WriteHeader(statusCode)
}
