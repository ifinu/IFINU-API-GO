package servico

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dominio/entidades"
	"github.com/ifinu/ifinu-api-go/dominio/enums"
	"github.com/ifinu/ifinu-api-go/dto"
	"github.com/ifinu/ifinu-api-go/repositorio"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
)

type StripeServico struct {
	usuarioRepo     *repositorio.UsuarioRepositorio
	assinaturaRepo  *repositorio.AssinaturaRepositorio
}

func NovoStripeServico(usuarioRepo *repositorio.UsuarioRepositorio, assinaturaRepo *repositorio.AssinaturaRepositorio) *StripeServico {
	// Configurar chave API do Stripe
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	return &StripeServico{
		usuarioRepo:    usuarioRepo,
		assinaturaRepo: assinaturaRepo,
	}
}

// CriarCheckoutTrial cria uma sessão de checkout Stripe para trial
func (s *StripeServico) CriarCheckoutTrial(usuarioID uuid.UUID) (*dto.CreateCheckoutResponse, error) {
	// Verificar se usuário existe
	_, err := s.usuarioRepo.BuscarPorID(usuarioID)
	if err != nil {
		return nil, err
	}

	// Por enquanto, retornar uma URL mockada
	// TODO: Implementar integração real com Stripe
	sessionID := fmt.Sprintf("cs_test_%s", uuid.New().String()[:8])
	checkoutURL := fmt.Sprintf("https://checkout.stripe.com/pay/%s", sessionID)

	return &dto.CreateCheckoutResponse{
		SessionID:   sessionID,
		CheckoutURL: checkoutURL,
	}, nil
}

// CriarCheckoutSession cria uma sessão de checkout Stripe com dados completos
func (s *StripeServico) CriarCheckoutSession(req *dto.CreateCheckoutRequest) (*dto.CreateCheckoutResponse, error) {
	// Calcular valores
	taxaPercentual := 1.0 // 1% de taxa
	taxaPlataforma := req.Valor * taxaPercentual / 100
	valorUsuario := req.Valor - taxaPlataforma

	// Converter valor para centavos (Stripe usa centavos)
	valorCentavos := int64(req.Valor * 100)

	// Criar parâmetros da sessão
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:         stripe.String(req.SuccessURL),
		CancelURL:          stripe.String(req.CancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(req.Moeda),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(req.Descricao),
						Description: stripe.String(fmt.Sprintf("Cobrança #%s", req.CobrancaID)),
					},
					UnitAmount: stripe.Int64(valorCentavos),
				},
				Quantity: stripe.Int64(1),
			},
		},
	}

	// Adicionar email do cliente se fornecido
	if req.ClienteEmail != "" {
		params.CustomerEmail = stripe.String(req.ClienteEmail)
	}

	// Adicionar metadata
	params.Metadata = map[string]string{
		"cobranca_id":   req.CobrancaID,
		"cliente_nome":  req.ClienteNome,
		"cliente_email": req.ClienteEmail,
	}

	// Criar sessão no Stripe
	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar sessão Stripe: %w", err)
	}

	return &dto.CreateCheckoutResponse{
		SessionID:      sess.ID,
		CheckoutURL:    sess.URL,
		ValorTotal:     req.Valor,
		TaxaPlataforma: taxaPlataforma,
		ValorUsuario:   valorUsuario,
		Moeda:          req.Moeda,
		Status:         "created",
	}, nil
}

// CriarCheckoutAssinatura cria uma sessão de checkout Stripe para assinatura
func (s *StripeServico) CriarCheckoutAssinatura(usuarioID uuid.UUID, req dto.CheckoutAssinaturaRequest) (*dto.CheckoutAssinaturaResponse, error) {
	// Verificar se usuário existe
	usuario, err := s.usuarioRepo.BuscarPorID(usuarioID)
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado: %w", err)
	}

	// Obter valor do plano
	valor := enums.ObterValorPlano(req.PlanoAssinatura)
	valorCentavos := int64(valor * 100)

	// Criar parâmetros da sessão
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card", "boleto"}),
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:         stripe.String(req.SuccessURL),
		CancelURL:          stripe.String(req.CancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("brl"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(fmt.Sprintf("IFINU - Assinatura %s", req.PlanoAssinatura)),
						Description: stripe.String(enums.ObterDescricaoPlano(req.PlanoAssinatura)),
					},
					UnitAmount: stripe.Int64(valorCentavos),
				},
				Quantity: stripe.Int64(1),
			},
		},
	}

	// Adicionar email do usuário
	if usuario.Email != "" {
		params.CustomerEmail = stripe.String(usuario.Email)
	}

	// Adicionar metadata para webhook
	params.Metadata = map[string]string{
		"tipo":             "assinatura_ifinu",
		"usuario_id":       usuarioID.String(),
		"plano_assinatura": string(req.PlanoAssinatura),
		"valor":            fmt.Sprintf("%.2f", valor),
	}

	// Criar sessão no Stripe
	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar sessão Stripe: %w", err)
	}

	return &dto.CheckoutAssinaturaResponse{
		CheckoutURL: sess.URL,
		SessionID:   sess.ID,
		Valor:       valor,
		Plano:       string(req.PlanoAssinatura),
	}, nil
}

// ListarPlanos retorna informações sobre os planos disponíveis
func (s *StripeServico) ListarPlanos() dto.ListarPlanosResponse {
	planos := []dto.PlanoInfo{
		{
			Tipo:               enums.PlanoMensal,
			Nome:               "Mensal",
			Descricao:          "Pagamento mensal",
			Valor:              39.00,
			ValorMensal:        39.00,
			IntervaloMeses:     1,
			PercentualDesconto: 0,
			Recomendado:        false,
		},
		{
			Tipo:               enums.PlanoTrimestral,
			Nome:               "Trimestral",
			Descricao:          "Pagamento a cada 3 meses",
			Valor:              99.00,
			ValorMensal:        33.00,
			IntervaloMeses:     3,
			PercentualDesconto: 15,
			Recomendado:        true,
		},
		{
			Tipo:               enums.PlanoAnual,
			Nome:               "Anual",
			Descricao:          "Pagamento anual",
			Valor:              348.00,
			ValorMensal:        29.00,
			IntervaloMeses:     12,
			PercentualDesconto: 25,
			Recomendado:        false,
		},
	}

	return dto.ListarPlanosResponse{Planos: planos}
}

// ProcessarPagamentoWebhook processa o webhook de pagamento confirmado
func (s *StripeServico) ProcessarPagamentoWebhook(sessionID string, metadata map[string]string) error {
	// Extrair informações do metadata
	usuarioIDStr := metadata["usuario_id"]
	planoStr := metadata["plano_assinatura"]

	if usuarioIDStr == "" || planoStr == "" {
		return fmt.Errorf("metadata incompleto no webhook")
	}

	usuarioID, err := uuid.Parse(usuarioIDStr)
	if err != nil {
		return fmt.Errorf("usuario_id inválido: %w", err)
	}

	plano := enums.PlanoAssinatura(planoStr)

	// Buscar ou criar assinatura do usuário
	assinatura, err := s.assinaturaRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		// Assinatura não existe, vamos criar uma nova
		// Por enquanto, apenas logamos o sucesso
		// TODO: Criar nova assinatura
		return fmt.Errorf("assinatura não encontrada para usuário %s", usuarioID)
	}

	// Atualizar assinatura
	assinatura.Status = entidades.StatusAtiva
	assinatura.PlanoAssinatura = plano
	assinatura.UltimaTransacaoID = sessionID
	agora := time.Now()
	assinatura.DataUltimaCobranca = &agora

	// Calcular próxima cobrança baseado no plano
	intervaloMeses := enums.ObterIntervaloCobranca(plano)
	proximaCobranca := agora.AddDate(0, intervaloMeses, 0)
	assinatura.DataProximaCobranca = &proximaCobranca

	// Salvar assinatura
	return s.assinaturaRepo.Atualizar(assinatura)
}
