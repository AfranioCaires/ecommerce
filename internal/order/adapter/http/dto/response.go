package dto

type OrderItemResponse struct {
	ProductID      string `json:"product_id"`
	ProductName    string `json:"product_name"`
	UnitPriceCents int64  `json:"unit_price_cents"`
	Quantity       int    `json:"quantity"`
	SubtotalCents  int64  `json:"subtotal_cents"`
}

type OrderResponse struct {
	ID               string              `json:"id"`
	UserID           string              `json:"user_id"`
	Status           string              `json:"status"`
	TotalAmountCents int64               `json:"total_amount_cents"`
	Items            []OrderItemResponse `json:"items"`
	CreatedAt        string              `json:"created_at"`
	UpdatedAt        string              `json:"updated_at"`
}
