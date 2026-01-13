package dto

// CreateCheckoutRequest representa a requisição para criar checkout Stripe
type CreateCheckoutRequest struct {
	PriceID      string  `json:"priceId"`
	CobrancaID   string  `json:"cobrancaId"`
	Valor        float64 `json:"valor"`
	Moeda        string  `json:"moeda"`
	Descricao    string  `json:"descricao"`
	ClienteNome  string  `json:"clienteNome"`
	ClienteEmail string  `json:"clienteEmail"`
	SuccessURL   string  `json:"successUrl"`
	CancelURL    string  `json:"cancelUrl"`
}

// CreateCheckoutResponse representa a resposta do checkout Stripe
type CreateCheckoutResponse struct {
	SessionID        string  `json:"sessionId"`
	CheckoutURL      string  `json:"checkoutUrl"`
	ValorTotal       float64 `json:"valorTotal"`
	TaxaPlataforma   float64 `json:"taxaPlataforma"`
	ValorUsuario     float64 `json:"valorUsuario"`
	Moeda            string  `json:"moeda"`
	StripeAccountID  string  `json:"stripeAccountId,omitempty"`
	Status           string  `json:"status"`
}
