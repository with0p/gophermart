package handlers

import (
	"github.com/go-chi/chi"
	"github.com/with0p/gophermart/internal/auth"
	"github.com/with0p/gophermart/internal/models"
	"github.com/with0p/gophermart/internal/service"
)

type HandlerUserAPI struct {
	service service.Service
	queue   chan models.OrderID
}

func NewHandlerUserAPI(currentService service.Service, queue chan models.OrderID) *HandlerUserAPI {
	return &HandlerUserAPI{service: currentService, queue: queue}
}

func (h HandlerUserAPI) GetHandlerUserAPIRouter() *chi.Mux {
	mux := chi.NewRouter()
	mux.Post(`/api/user/register`, h.RegisterUser)
	mux.Post(`/api/user/login`, h.LoginUser)
	mux.Post(`/api/user/orders`, auth.UseValidateAuth(h.AddOrder))
	mux.Get(`/api/user/orders`, auth.UseValidateAuth(h.GetUserOrders))
	mux.Post(`/api/user/balance/withdraw`, auth.UseValidateAuth(h.MakeWithdrawal))
	mux.Get(`/api/user/balance`, auth.UseValidateAuth(h.GetUserBalance))
	mux.Get(`/api/user/withdrawals`, auth.UseValidateAuth(h.GetUserWithdrawals))
	return mux
}
