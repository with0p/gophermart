package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/with0p/gophermart/internal/auth"
	"github.com/with0p/gophermart/internal/custom-error"
	"github.com/with0p/gophermart/internal/models"
)

func (h *HandlerUserAPI) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Not a POST requests", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("content-type") != "application/json" {
		http.Error(w, "Not a \"application/json\" content-type", http.StatusBadRequest)
		return
	}

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	statusCode := http.StatusOK

	serviceErr := h.service.RegisterUser(r.Context(), user.Login, user.Password)
	if serviceErr != nil {
		if errors.Is(serviceErr, customerror.ErrUniqueKeyConstrantViolation) {
			statusCode = http.StatusConflict
		} else {
			http.Error(w, serviceErr.Error(), http.StatusBadRequest)
			return
		}
	}

	auth.SetAuth(r, w, user.Login)

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
}
