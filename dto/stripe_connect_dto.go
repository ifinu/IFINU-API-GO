package dto

// CriarContaConnectRequest - Request para criar conta Stripe Connect
type CriarContaConnectRequest struct {
	TipoConta string `json:"tipoConta"` // "express" ou "standard"
}

// CriarContaConnectResponse - Response com link de onboarding
type CriarContaConnectResponse struct {
	AccountID    string `json:"accountId"`
	OnboardingURL string `json:"onboardingUrl"`
	ExpiresAt    int64  `json:"expiresAt"`
}

// StatusStripeConnectResponse - Status da conta conectada
type StatusStripeConnectResponse struct {
	Conectado            bool   `json:"conectado"`
	AccountID            string `json:"accountId,omitempty"`
	OnboardingCompleto   bool   `json:"onboardingCompleto"`
	ChargesHabilitado    bool   `json:"chargesHabilitado"`
	DetalhesSubmetidos   bool   `json:"detalhesSubmetidos"`
	PrecisaAcao          bool   `json:"precisaAcao"`
	DashboardURL         string `json:"dashboardUrl,omitempty"`
}

// RefreshOnboardingRequest - Request para gerar novo link de onboarding
type RefreshOnboardingRequest struct {
	ReturnURL  string `json:"returnUrl"`
	RefreshURL string `json:"refreshUrl"`
}

// DashboardLinkResponse - Link para dashboard da conta conectada
type DashboardLinkResponse struct {
	URL string `json:"url"`
}
