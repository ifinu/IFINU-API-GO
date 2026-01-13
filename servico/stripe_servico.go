package servico

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dto"
	"github.com/ifinu/ifinu-api-go/repositorio"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
)

type StripeServico struct {
	usuarioRepo *repositorio.UsuarioRepositorio
}

func NovoStripeServico(usuarioRepo *repositorio.UsuarioRepositorio) *StripeServico {
	// Configurar chave API do Stripe
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	return &StripeServico{
		usuarioRepo: usuarioRepo,
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
