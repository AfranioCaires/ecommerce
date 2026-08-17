package dto

type CheckoutItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type CheckoutRequest struct {
	Items []CheckoutItemRequest `json:"items"`
}

type ChallengeCheckoutItemRequest struct {
	ProductID string `json:"produtoId"`
	Quantity  int    `json:"quantidade"`
}

type ChallengeCheckoutRequest struct {
	CustomerID string                         `json:"clienteId"`
	Items      []ChallengeCheckoutItemRequest `json:"itens"`
}
