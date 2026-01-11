package dto

// CreateCheckoutRequest representa a requisição para criar checkout Stripe
type CreateCheckoutRequest struct {
	PriceID string `json:"priceId"`
}

// CreateCheckoutResponse representa a resposta do checkout Stripe
type CreateCheckoutResponse struct {
	SessionID   string `json:"sessionId"`
	CheckoutURL string `json:"checkoutUrl"`
}
