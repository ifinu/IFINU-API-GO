package dto

// SaveStripeConfigRequest representa a requisição para salvar configuração Stripe
type SaveStripeConfigRequest struct {
	PublishableKey string `json:"publishableKey" validate:"required"`
	SecretKey      string `json:"secretKey"` // Opcional na atualização
	TestMode       bool   `json:"testMode"`
}

// StripeConfigResponse representa a resposta com configuração Stripe
type StripeConfigResponse struct {
	Configured       bool   `json:"configured"`
	PublishableKey   string `json:"publishableKey,omitempty"`
	SecretKeyMasked  string `json:"secretKeyMasked,omitempty"`
	TestMode         bool   `json:"testMode"`
}

// TestConnectionResponse representa o resultado do teste de conexão
type TestConnectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
