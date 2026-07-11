package dto

type StockResponse struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	UpdatedAt string `json:"updated_at"`
}
