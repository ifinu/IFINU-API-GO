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

// FaturaInfo representa informações de uma fatura
type FaturaInfo struct {
	ID           string  `json:"id"`
	Data         string  `json:"data"`
	Valor        float64 `json:"valor"`
	Status       string  `json:"status"`
	URLPagamento string  `json:"urlPagamento,omitempty"`
	URLPDF       string  `json:"urlPdf,omitempty"`
}

// HistoricoFaturasResponse representa o histórico de faturas
type HistoricoFaturasResponse struct {
	Faturas []FaturaInfo `json:"faturas"`
}

// DetalhesAssinaturaResponse representa os detalhes completos da assinatura
type DetalhesAssinaturaResponse struct {
	PlanoNome          string   `json:"planoNome"`
	PlanoDescricao     string   `json:"planoDescricao"`
	Status             string   `json:"status"`
	StatusBadge        string   `json:"statusBadge"`
	ValorMensal        float64  `json:"valorMensal"`
	Moeda              string   `json:"moeda"`
	ProximaCobranca    *string  `json:"proximaCobranca,omitempty"`
	UltimaCobranca     *string  `json:"ultimaCobranca,omitempty"`
	DiasRestantesTrial *int     `json:"diasRestantesTrial,omitempty"`
}
