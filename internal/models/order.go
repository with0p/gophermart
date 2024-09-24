package models

import (
	"github.com/google/uuid"
)

type OrderID string
type Order struct {
	OrderID    OrderID   `json:"number"`
	Status     string    `json:"status"`
	Accrual    float32   `json:"accrual"`
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

type OrderExternalData struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}
