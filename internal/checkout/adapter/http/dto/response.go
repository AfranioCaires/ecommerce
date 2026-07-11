package dto

type CheckoutItemResponse struct {
	ProductID      string `json:"product_id"`
	ProductName    string `json:"product_name"`
	UnitPriceCents int64  `json:"unit_price_cents"`
	Quantity       int    `json:"quantity"`
	SubtotalCents  int64  `json:"subtotal_cents"`
}

type CheckoutResponse struct {
	OrderID          string                 `json:"order_id"`
	OrderStatus      string                 `json:"order_status"`
	PaymentID        string                 `json:"payment_id"`
	PaymentStatus    string                 `json:"payment_status"`
	TotalAmountCents int64                  `json:"total_amount_cents"`
	Items            []CheckoutItemResponse `json:"items"`
	CreatedAt        string                 `json:"created_at"`
}
