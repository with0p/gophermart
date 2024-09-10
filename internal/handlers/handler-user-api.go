package handlers

import (
	"github.com/go-chi/chi"
	"github.com/with0p/gophermart/internal/auth"
	"github.com/with0p/gophermart/internal/service"
)

type HandlerUserAPI struct {
	service service.Service
}

func NewHandlerUserAPI(currentService service.Service) *HandlerUserAPI {
	return &HandlerUserAPI{service: currentService}
}

func (h HandlerUserAPI) GetHandlerUserAPIRouter() *chi.Mux {
	mux := chi.NewRouter()
	mux.Post(`/api/user/register`, h.RegisterUser)
	mux.Post(`/api/user/login`, h.LoginUser)
	mux.Post(`/api/user/orders`, auth.UseValidateAuth(h.AddOrder))
	mux.Get(`/api/user/orders`, auth.UseValidateAuth(h.GetUserOrders))
	return mux
}
