package models

type Withdrawal struct {
	OrderID     OrderID `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}
