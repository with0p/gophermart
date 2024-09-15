package models

type Withdrawal struct {
	OrderID     OrderID `json:"order"`
	Sum         int     `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}
