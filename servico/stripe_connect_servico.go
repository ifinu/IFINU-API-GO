package servico

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dto"
	"github.com/ifinu/ifinu-api-go/repositorio"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/account"
	"github.com/stripe/stripe-go/v81/accountlink"
	"github.com/stripe/stripe-go/v81/loginlink"
)

type StripeConnectServico struct {
	usuarioRepo *repositorio.UsuarioRepositorio
}

func NovoStripeConnectServico(usuarioRepo *repositorio.UsuarioRepositorio) *StripeConnectServico {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	return &StripeConnectServico{
		usuarioRepo: usuarioRepo,
	}
}

// CriarContaConnect cria uma conta Express no Stripe Connect
func (s *StripeConnectServico) CriarContaConnect(usuarioID uuid.UUID, returnURL, refreshURL string) (*dto.CriarContaConnectResponse, error) {
	// Buscar usuário
	usuario, err := s.usuarioRepo.BuscarPorID(usuarioID)
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado: %w", err)
	}

	// Verificar se já tem conta conectada
	if usuario.StripeAccountID != "" {
		// Conta já existe, apenas gerar novo link de onboarding
		return s.GerarLinkOnboarding(usuarioID, returnURL, refreshURL)
	}

	// Criar conta Express no Stripe
	accountParams := &stripe.AccountParams{
		Type: stripe.String(string(stripe.AccountTypeExpress)),
		Country: stripe.String("BR"),
		Email: stripe.String(usuario.Email),
		Capabilities: &stripe.AccountCapabilitiesParams{
			CardPayments: &stripe.AccountCapabilitiesCardPaymentsParams{
				Requested: stripe.Bool(true),
			},
			Transfers: &stripe.AccountCapabilitiesTransfersParams{
				Requested: stripe.Bool(true),
			},
		},
		BusinessType: stripe.String("individual"), // ou "company" se tiver CNPJ
	}

	// Adicionar nome da empresa se existir
	if usuario.NomeEmpresa != "" {
		accountParams.BusinessProfile = &stripe.AccountBusinessProfileParams{
			Name: stripe.String(usuario.NomeEmpresa),
		}
	}

	// Criar conta
	acc, err := account.New(accountParams)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar conta Stripe Connect: %w", err)
	}

	// Salvar account ID no banco
	usuario.StripeAccountID = acc.ID
	if err := s.usuarioRepo.Atualizar(usuario); err != nil {
		return nil, fmt.Errorf("erro ao salvar Stripe Account ID: %w", err)
	}

	// Gerar link de onboarding
	linkParams := &stripe.AccountLinkParams{
		Account:    stripe.String(acc.ID),
		RefreshURL: stripe.String(refreshURL),
		ReturnURL:  stripe.String(returnURL),
		Type:       stripe.String("account_onboarding"),
	}

	link, err := accountlink.New(linkParams)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar link de onboarding: %w", err)
	}

	return &dto.CriarContaConnectResponse{
		AccountID:    acc.ID,
		OnboardingURL: link.URL,
		ExpiresAt:    link.ExpiresAt,
	}, nil
}

// GerarLinkOnboarding gera novo link de onboarding para conta existente
func (s *StripeConnectServico) GerarLinkOnboarding(usuarioID uuid.UUID, returnURL, refreshURL string) (*dto.CriarContaConnectResponse, error) {
	// Buscar usuário
	usuario, err := s.usuarioRepo.BuscarPorID(usuarioID)
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado: %w", err)
	}

	if usuario.StripeAccountID == "" {
		return nil, fmt.Errorf("usuário não possui conta Stripe Connect")
	}

	// Gerar novo link de onboarding
	linkParams := &stripe.AccountLinkParams{
		Account:    stripe.String(usuario.StripeAccountID),
		RefreshURL: stripe.String(refreshURL),
		ReturnURL:  stripe.String(returnURL),
		Type:       stripe.String("account_onboarding"),
	}

	link, err := accountlink.New(linkParams)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar link de onboarding: %w", err)
	}

	return &dto.CriarContaConnectResponse{
		AccountID:    usuario.StripeAccountID,
		OnboardingURL: link.URL,
		ExpiresAt:    link.ExpiresAt,
	}, nil
}

// ObterStatusConnect retorna status da conta conectada
func (s *StripeConnectServico) ObterStatusConnect(usuarioID uuid.UUID) (*dto.StatusStripeConnectResponse, error) {
	// Buscar usuário
	usuario, err := s.usuarioRepo.BuscarPorID(usuarioID)
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado: %w", err)
	}

	// Se não tem account ID, não está conectado
	if usuario.StripeAccountID == "" {
		return &dto.StatusStripeConnectResponse{
			Conectado:          false,
			OnboardingCompleto: false,
			ChargesHabilitado:  false,
			DetalhesSubmetidos: false,
			PrecisaAcao:        true,
		}, nil
	}

	// Buscar detalhes da conta no Stripe
	acc, err := account.GetByID(usuario.StripeAccountID, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar conta no Stripe: %w", err)
	}

	// Verificar status
	chargesHabilitado := acc.ChargesEnabled
	detalhesSubmetidos := acc.DetailsSubmitted
	onboardingCompleto := chargesHabilitado && detalhesSubmetidos

	// Atualizar banco de dados se status mudou
	atualizou := false
	if usuario.StripeOnboardingCompleto != onboardingCompleto {
		usuario.StripeOnboardingCompleto = onboardingCompleto
		if onboardingCompleto && usuario.StripeDataOnboarding == nil {
			agora := time.Now()
			usuario.StripeDataOnboarding = &agora
		}
		atualizou = true
	}
	if usuario.StripeChargesHabilitado != chargesHabilitado {
		usuario.StripeChargesHabilitado = chargesHabilitado
		atualizou = true
	}
	if usuario.StripeDetalhesSubmetidos != detalhesSubmetidos {
		usuario.StripeDetalhesSubmetidos = detalhesSubmetidos
		atualizou = true
	}

	if atualizou {
		if err := s.usuarioRepo.Atualizar(usuario); err != nil {
			// Log error mas não falhar
			fmt.Printf("Erro ao atualizar status Stripe do usuário: %v\n", err)
		}
	}

	return &dto.StatusStripeConnectResponse{
		Conectado:          true,
		AccountID:          usuario.StripeAccountID,
		OnboardingCompleto: onboardingCompleto,
		ChargesHabilitado:  chargesHabilitado,
		DetalhesSubmetidos: detalhesSubmetidos,
		PrecisaAcao:        !onboardingCompleto,
	}, nil
}

// GerarDashboardLink gera link para usuário acessar dashboard Stripe
func (s *StripeConnectServico) GerarDashboardLink(usuarioID uuid.UUID) (*dto.DashboardLinkResponse, error) {
	// Buscar usuário
	usuario, err := s.usuarioRepo.BuscarPorID(usuarioID)
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado: %w", err)
	}

	if usuario.StripeAccountID == "" {
		return nil, fmt.Errorf("usuário não possui conta Stripe Connect")
	}

	// Gerar login link
	params := &stripe.LoginLinkParams{
		Account: stripe.String(usuario.StripeAccountID),
	}

	link, err := loginlink.New(params)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar dashboard link: %w", err)
	}

	return &dto.DashboardLinkResponse{
		URL: link.URL,
	}, nil
}

// DesconectarConta remove conexão com Stripe Connect
func (s *StripeConnectServico) DesconectarConta(usuarioID uuid.UUID) error {
	// Buscar usuário
	usuario, err := s.usuarioRepo.BuscarPorID(usuarioID)
	if err != nil {
		return fmt.Errorf("usuário não encontrado: %w", err)
	}

	if usuario.StripeAccountID == "" {
		return fmt.Errorf("usuário não possui conta Stripe Connect")
	}

	// IMPORTANTE: Não deletar a conta no Stripe (ela continua existindo)
	// Apenas remover a conexão no nosso sistema

	usuario.StripeAccountID = ""
	usuario.StripeOnboardingCompleto = false
	usuario.StripeChargesHabilitado = false
	usuario.StripeDetalhesSubmetidos = false
	usuario.StripeDataOnboarding = nil

	if err := s.usuarioRepo.Atualizar(usuario); err != nil {
		return fmt.Errorf("erro ao desconectar conta: %w", err)
	}

	return nil
}

// ProcessarAccountWebhook processa webhooks de account.updated
func (s *StripeConnectServico) ProcessarAccountWebhook(accountID string, chargesEnabled, detailsSubmitted bool) error {
	// Buscar usuário pelo account ID
	usuario, err := s.usuarioRepo.BuscarPorStripeAccountID(accountID)
	if err != nil {
		return fmt.Errorf("usuário não encontrado para account ID %s: %w", accountID, err)
	}

	// Atualizar status
	onboardingCompleto := chargesEnabled && detailsSubmitted

	atualizou := false
	if usuario.StripeOnboardingCompleto != onboardingCompleto {
		usuario.StripeOnboardingCompleto = onboardingCompleto
		if onboardingCompleto && usuario.StripeDataOnboarding == nil {
			agora := time.Now()
			usuario.StripeDataOnboarding = &agora
		}
		atualizou = true
	}
	if usuario.StripeChargesHabilitado != chargesEnabled {
		usuario.StripeChargesHabilitado = chargesEnabled
		atualizou = true
	}
	if usuario.StripeDetalhesSubmetidos != detailsSubmitted {
		usuario.StripeDetalhesSubmetidos = detailsSubmitted
		atualizou = true
	}

	if atualizou {
		if err := s.usuarioRepo.Atualizar(usuario); err != nil {
			return fmt.Errorf("erro ao atualizar usuário: %w", err)
		}
	}

	return nil
}
