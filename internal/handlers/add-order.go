package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/with0p/gophermart/internal/auth"
	"github.com/with0p/gophermart/internal/custom-error"
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

	orderID := string(body)

	errOrder := h.service.AddOrder(ctx, login, orderID)
	var statusCode int

	switch {
	case errOrder == nil:
		statusCode = http.StatusAccepted
	case errors.Is(errOrder, customerror.ErrAnotherUserOrder):
		statusCode = http.StatusConflict
	case errors.Is(errOrder, customerror.ErrWrongOrderFormat):
		statusCode = http.StatusBadRequest
	case errors.Is(errOrder, customerror.ErrAreadyAdded):
		statusCode = http.StatusOK
	default:
		statusCode = http.StatusInternalServerError
	}

	w.WriteHeader(statusCode)
}
