package servico

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dominio/entidades"
	"github.com/ifinu/ifinu-api-go/dominio/enums"
	"github.com/ifinu/ifinu-api-go/dto"
	"github.com/ifinu/ifinu-api-go/repositorio"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/invoice"
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
// IMPORTANTE: Usa Stripe Connect - dinheiro vai para conta conectada do usuário
func (s *StripeServico) CriarCheckoutSession(usuarioID uuid.UUID, req *dto.CreateCheckoutRequest) (*dto.CreateCheckoutResponse, error) {
	// Buscar usuário e verificar se tem conta Stripe Connect
	usuario, err := s.usuarioRepo.BuscarPorID(usuarioID)
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado: %w", err)
	}

	// Verificar se usuário tem conta conectada
	if usuario.StripeAccountID == "" {
		return nil, fmt.Errorf("usuário não possui conta Stripe Connect configurada. Configure em Configurações > Pagamentos")
	}

	// Verificar se onboarding está completo
	if !usuario.StripeOnboardingCompleto {
		return nil, fmt.Errorf("conta Stripe Connect ainda não está pronta para receber pagamentos. Complete o cadastro no Stripe")
	}

	// Calcular valores
	taxaPercentual := 1.0 // 1% de taxa
	taxaPlataforma := req.Valor * taxaPercentual / 100
	valorUsuario := req.Valor - taxaPlataforma

	// Converter valor para centavos (Stripe usa centavos)
	valorCentavos := int64(req.Valor * 100)

	// Garantir que a moeda está em lowercase (requisito do Stripe)
	moeda := strings.ToLower(req.Moeda)

	// Criar parâmetros da sessão
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:         stripe.String(req.SuccessURL),
		CancelURL:          stripe.String(req.CancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(moeda),
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

	// CRÍTICO: Usar conta conectada do usuário (dinheiro vai para ele)
	params.SetStripeAccount(usuario.StripeAccountID)

	// Adicionar email do cliente se fornecido
	if req.ClienteEmail != "" {
		params.CustomerEmail = stripe.String(req.ClienteEmail)
	}

	// Adicionar metadata
	params.Metadata = map[string]string{
		"cobranca_id":   req.CobrancaID,
		"cliente_nome":  req.ClienteNome,
		"cliente_email": req.ClienteEmail,
		"usuario_id":    usuarioID.String(),
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

// CriarCheckoutAssinatura cria uma sessão de checkout Stripe para assinatura recorrente
func (s *StripeServico) CriarCheckoutAssinatura(usuarioID uuid.UUID, req dto.CheckoutAssinaturaRequest) (*dto.CheckoutAssinaturaResponse, error) {
	// Verificar se usuário existe
	usuario, err := s.usuarioRepo.BuscarPorID(usuarioID)
	if err != nil {
		return nil, fmt.Errorf("usuário não encontrado: %w", err)
	}

	// Obter Price ID do Stripe para o plano escolhido
	priceID := enums.ObterStripePriceID(req.PlanoAssinatura)
	if priceID == "" {
		return nil, fmt.Errorf("Price ID não configurado para o plano %s. Configure a variável de ambiente STRIPE_PRICE_ID_%s", req.PlanoAssinatura, req.PlanoAssinatura)
	}

	// Obter valor do plano (para retornar na resposta)
	valor := enums.ObterValorPlano(req.PlanoAssinatura)

	// Criar parâmetros da sessão de SUBSCRIPTION com trial de 14 dias
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		Mode:               stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL:         stripe.String(req.SuccessURL),
		CancelURL:          stripe.String(req.CancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			TrialPeriodDays: stripe.Int64(14),
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

// ProcessarCheckoutWebhook processa o webhook quando checkout é concluído
func (s *StripeServico) ProcessarCheckoutWebhook(sessionID, subscriptionID string, metadata map[string]string) error {
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

	agora := time.Now()
	intervaloMeses := enums.ObterIntervaloCobranca(plano)
	proximaCobranca := agora.AddDate(0, intervaloMeses, 0).Add(14 * 24 * time.Hour) // Adiciona os 14 dias de trial
	valor := enums.ObterValorPlano(plano)

	if err != nil {
		// Assinatura não existe, criar uma nova
		novaAssinatura := &entidades.AssinaturaUsuario{
			UsuarioID:            usuarioID,
			Status:               entidades.StatusPeriodoGratuito, // Começa em trial
			PlanoAssinatura:      plano,
			DataUltimaCobranca:   nil, // Trial não cobra ainda
			DataProximaCobranca:  &proximaCobranca,
			ValorMensal:          valor,
			Currency:             "BRL",
			Country:              "BR",
			UltimaTransacaoID:    sessionID,
			StripeSubscriptionID: subscriptionID,
		}

		return s.assinaturaRepo.Criar(novaAssinatura)
	}

	// Assinatura existe, atualizar
	assinatura.Status = entidades.StatusPeriodoGratuito
	assinatura.PlanoAssinatura = plano
	assinatura.UltimaTransacaoID = sessionID
	assinatura.DataProximaCobranca = &proximaCobranca
	assinatura.ValorMensal = valor
	assinatura.StripeSubscriptionID = subscriptionID

	return s.assinaturaRepo.Atualizar(assinatura)
}

// ProcessarSubscriptionCriada processa quando subscription é criada no Stripe
func (s *StripeServico) ProcessarSubscriptionCriada(subscriptionID, customerID, status string) error {
	// Buscar assinatura por subscription ID
	assinatura, err := s.assinaturaRepo.BuscarPorStripeSubscriptionID(subscriptionID)
	if err != nil {
		// Se não encontrou por subscription ID, buscar por customer ID
		assinatura, err = s.assinaturaRepo.BuscarPorStripeCustomerID(customerID)
		if err != nil {
			return fmt.Errorf("assinatura não encontrada: %w", err)
		}
	}

	// Atualizar com os IDs do Stripe
	assinatura.StripeSubscriptionID = subscriptionID
	assinatura.StripeCustomerID = customerID

	// Status no Stripe pode ser: trialing, active, past_due, canceled, etc
	if status == "trialing" {
		assinatura.Status = entidades.StatusPeriodoGratuito
	} else if status == "active" {
		assinatura.Status = entidades.StatusAtiva
	}

	return s.assinaturaRepo.Atualizar(assinatura)
}

// ProcessarSubscriptionAtualizada processa quando subscription é atualizada
func (s *StripeServico) ProcessarSubscriptionAtualizada(subscriptionID, status string) error {
	assinatura, err := s.assinaturaRepo.BuscarPorStripeSubscriptionID(subscriptionID)
	if err != nil {
		return fmt.Errorf("assinatura não encontrada: %w", err)
	}

	// Atualizar status baseado no status do Stripe
	switch status {
	case "active":
		assinatura.Status = entidades.StatusAtiva
		// Atualizar data de última cobrança
		agora := time.Now()
		assinatura.DataUltimaCobranca = &agora

		// Calcular próxima cobrança
		intervaloMeses := enums.ObterIntervaloCobranca(assinatura.PlanoAssinatura)
		proximaCobranca := agora.AddDate(0, intervaloMeses, 0)
		assinatura.DataProximaCobranca = &proximaCobranca

	case "past_due":
		assinatura.Status = entidades.StatusPendentePagamento

	case "canceled":
		assinatura.Status = entidades.StatusCancelada
		agora := time.Now()
		assinatura.DataCancelamento = &agora

	case "unpaid":
		assinatura.Status = entidades.StatusBloqueada
	}

	return s.assinaturaRepo.Atualizar(assinatura)
}

// ProcessarSubscriptionCancelada processa quando subscription é cancelada
func (s *StripeServico) ProcessarSubscriptionCancelada(subscriptionID string) error {
	assinatura, err := s.assinaturaRepo.BuscarPorStripeSubscriptionID(subscriptionID)
	if err != nil {
		return fmt.Errorf("assinatura não encontrada: %w", err)
	}

	assinatura.Status = entidades.StatusCancelada
	agora := time.Now()
	assinatura.DataCancelamento = &agora

	return s.assinaturaRepo.Atualizar(assinatura)
}

// ProcessarPagamentoWebhook processa o webhook de pagamento confirmado (método legado)
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

	agora := time.Now()
	intervaloMeses := enums.ObterIntervaloCobranca(plano)
	proximaCobranca := agora.AddDate(0, intervaloMeses, 0)
	valor := enums.ObterValorPlano(plano)

	if err != nil {
		// Assinatura não existe, criar uma nova
		novaAssinatura := &entidades.AssinaturaUsuario{
			UsuarioID:           usuarioID,
			Status:              entidades.StatusAtiva,
			PlanoAssinatura:     plano,
			DataUltimaCobranca:  &agora,
			DataProximaCobranca: &proximaCobranca,
			ValorMensal:         valor,
			Currency:            "BRL",
			Country:             "BR",
			UltimaTransacaoID:   sessionID,
		}

		return s.assinaturaRepo.Criar(novaAssinatura)
	}

	// Assinatura existe, atualizar
	assinatura.Status = entidades.StatusAtiva
	assinatura.PlanoAssinatura = plano
	assinatura.UltimaTransacaoID = sessionID
	assinatura.DataUltimaCobranca = &agora
	assinatura.DataProximaCobranca = &proximaCobranca
	assinatura.ValorMensal = valor

	// Salvar assinatura
	return s.assinaturaRepo.Atualizar(assinatura)
}

// BuscarHistoricoFaturas busca o histórico de faturas do Stripe
func (s *StripeServico) BuscarHistoricoFaturas(usuarioID uuid.UUID) (*dto.HistoricoFaturasResponse, error) {
	// Buscar assinatura do usuário
	assinatura, err := s.assinaturaRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		return &dto.HistoricoFaturasResponse{
			Faturas: []dto.FaturaInfo{},
		}, nil
	}

	// Se não tem customer ID do Stripe, retornar vazio
	if assinatura.StripeCustomerID == "" {
		return &dto.HistoricoFaturasResponse{
			Faturas: []dto.FaturaInfo{},
		}, nil
	}

	// Buscar invoices do Stripe
	params := &stripe.InvoiceListParams{
		Customer: stripe.String(assinatura.StripeCustomerID),
	}
	params.Limit = stripe.Int64(10)

	faturas := []dto.FaturaInfo{}
	i := invoice.List(params)

	for i.Next() {
		inv := i.Invoice()

		fatura := dto.FaturaInfo{
			ID:          inv.ID,
			Data:        time.Unix(inv.Created, 0).Format("02/01/2006"),
			Valor:       float64(inv.AmountPaid) / 100, // Converter de centavos para reais
			Status:      converterStatusFatura(inv.Status),
			URLPagamento: inv.HostedInvoiceURL,
			URLPDF:      inv.InvoicePDF,
		}

		faturas = append(faturas, fatura)
	}

	if err := i.Err(); err != nil {
		return nil, fmt.Errorf("erro ao listar faturas: %w", err)
	}

	return &dto.HistoricoFaturasResponse{
		Faturas: faturas,
	}, nil
}

// BuscarDetalhesAssinatura retorna detalhes completos da assinatura
func (s *StripeServico) BuscarDetalhesAssinatura(usuarioID uuid.UUID) (*dto.DetalhesAssinaturaResponse, error) {
	// Buscar assinatura do usuário
	assinatura, err := s.assinaturaRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		return nil, fmt.Errorf("assinatura não encontrada: %w", err)
	}

	// Montar resposta
	response := &dto.DetalhesAssinaturaResponse{
		PlanoNome:      converterNomePlano(assinatura.PlanoAssinatura),
		PlanoDescricao: "Sistema de cobrança automática",
		Status:         string(assinatura.Status),
		StatusBadge:    converterStatusBadge(assinatura.Status),
		ValorMensal:    assinatura.ValorMensal,
		Moeda:          assinatura.Currency,
	}

	// Adicionar datas se disponíveis
	if assinatura.DataProximaCobranca != nil {
		proximaCobranca := assinatura.DataProximaCobranca.Format("02/01/2006")
		response.ProximaCobranca = &proximaCobranca
	}

	if assinatura.DataUltimaCobranca != nil {
		ultimaCobranca := assinatura.DataUltimaCobranca.Format("02/01/2006")
		response.UltimaCobranca = &ultimaCobranca
	}

	// Calcular dias restantes de trial
	if assinatura.Status == entidades.StatusPeriodoGratuito && assinatura.DataProximaCobranca != nil {
		diasRestantes := int(time.Until(*assinatura.DataProximaCobranca).Hours() / 24)
		if diasRestantes < 0 {
			diasRestantes = 0
		}
		response.DiasRestantesTrial = &diasRestantes
	}

	return response, nil
}

// Funções auxiliares
func converterStatusFatura(status stripe.InvoiceStatus) string {
	switch status {
	case stripe.InvoiceStatusPaid:
		return "Pago"
	case stripe.InvoiceStatusOpen:
		return "Aberta"
	case stripe.InvoiceStatusVoid:
		return "Cancelada"
	case stripe.InvoiceStatusUncollectible:
		return "Não cobrável"
	default:
		return "Pendente"
	}
}

func converterNomePlano(plano enums.PlanoAssinatura) string {
	switch plano {
	case enums.PlanoMensal:
		return "Plano Mensal"
	case enums.PlanoTrimestral:
		return "Plano Trimestral"
	case enums.PlanoAnual:
		return "Plano Anual"
	default:
		return "Plano IFINU"
	}
}

func converterStatusBadge(status entidades.StatusAssinatura) string {
	switch status {
	case entidades.StatusAtiva:
		return "success"
	case entidades.StatusPeriodoGratuito:
		return "info"
	case entidades.StatusPendentePagamento:
		return "warning"
	case entidades.StatusBloqueada:
		return "error"
	case entidades.StatusCancelada:
		return "default"
	case entidades.StatusVitalicia:
		return "premium"
	default:
		return "default"
	}
}
