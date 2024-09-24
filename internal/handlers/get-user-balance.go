package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/with0p/gophermart/internal/auth"
)

func (h *HandlerUserAPI) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Not a GET requests", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	login, errLogin := auth.GetLoginFromRequestContext(ctx)
	if errLogin != nil {
		http.Error(w, errLogin.Error(), http.StatusInternalServerError)
		return
	}

	balance, err := h.service.GetUserBalance(ctx, login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
