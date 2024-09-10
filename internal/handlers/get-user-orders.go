package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/with0p/gophermart/internal/auth"
)

func (h *HandlerUserAPI) GetUserOrders(w http.ResponseWriter, r *http.Request) {
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

	orders, errOrders := h.service.GetUserOrders(ctx, login)

	if errOrders != nil {
		http.Error(w, errOrders.Error(), http.StatusInternalServerError)
		return
	}

	statusCode := http.StatusOK

	if len(orders) == 0 {
		statusCode = http.StatusNoContent
	}

	response, err := json.Marshal(orders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}
