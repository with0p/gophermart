package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/with0p/gophermart/internal/auth"
)

func (h *HandlerUserAPI) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
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

	withdrawals, err := h.service.GetUserWithdrawals(ctx, login)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	statusCode := http.StatusOK

	if len(withdrawals) == 0 {
		statusCode = http.StatusNoContent
	}

	response, err := json.Marshal(withdrawals)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}
