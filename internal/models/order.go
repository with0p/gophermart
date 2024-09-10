package models

import (
	"github.com/google/uuid"
)

type Order struct {
	OrderID    string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    int       `json:"accrual"`
	UploadDate string    `json:"uploaded_at"`
	UserID     uuid.UUID `json:"-"`
}

type OrderStatus string

const (
	StatusNew        OrderStatus = "NEW"
	StatusProcessing OrderStatus = "PROCESSING"
	StatusInvalid    OrderStatus = "INVALID"
	StatusProcessed  OrderStatus = "PROCESSED"
)
