package controlador

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ifinu/ifinu-api-go/dto"
	"github.com/ifinu/ifinu-api-go/middleware"
	"github.com/ifinu/ifinu-api-go/servico"
	"github.com/ifinu/ifinu-api-go/util"
)

type StripeControlador struct {
	stripeServico *servico.StripeServico
}

func NovoStripeControlador(stripeServico *servico.StripeServico) *StripeControlador {
	return &StripeControlador{
		stripeServico: stripeServico,
	}
}

// CreateCheckout cria uma sessão de checkout Stripe
// POST /stripe-trial/create-checkout
func (ctrl *StripeControlador) CreateCheckout(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	resultado, err := ctrl.stripeServico.CriarCheckoutTrial(usuarioID)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao criar checkout", err)
		return
	}

	util.RespostaSucesso(c, "Checkout criado com sucesso", resultado)
}

// CreateCheckoutSession cria uma sessão de checkout Stripe completa
// POST /api/stripe-connect/create-checkout-session
func (ctrl *StripeControlador) CreateCheckoutSession(c *gin.Context) {
	_, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	var req dto.CreateCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "Dados inválidos", err)
		return
	}

	resultado, err := ctrl.stripeServico.CriarCheckoutSession(&req)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao criar sessão de checkout", err)
		return
	}

	util.RespostaSucesso(c, "Sessão de checkout criada com sucesso", resultado)
}

// ListarPlanos lista os planos de assinatura disponíveis
// GET /api/assinaturas/planos
func (ctrl *StripeControlador) ListarPlanos(c *gin.Context) {
	resultado := ctrl.stripeServico.ListarPlanos()
	util.RespostaSucesso(c, "Planos listados com sucesso", resultado)
}

// CriarCheckoutAssinatura cria uma sessão de checkout para assinatura
// POST /api/assinaturas/checkout
func (ctrl *StripeControlador) CriarCheckoutAssinatura(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	var req dto.CheckoutAssinaturaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "Dados inválidos", err)
		return
	}

	resultado, err := ctrl.stripeServico.CriarCheckoutAssinatura(usuarioID, req)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao criar checkout", err)
		return
	}

	util.RespostaSucesso(c, "Checkout criado com sucesso", resultado)
}

// WebhookStripe processa eventos do webhook Stripe
// POST /api/stripe/webhook
func (ctrl *StripeControlador) WebhookStripe(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "Payload inválido", err)
		return
	}

	// Extrair tipo de evento
	eventType, ok := payload["type"].(string)
	if !ok {
		util.RespostaErro(c, http.StatusBadRequest, "Tipo de evento não encontrado", nil)
		return
	}

	// Processar eventos do Stripe
	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		util.RespostaErro(c, http.StatusBadRequest, "Dados do evento inválidos", nil)
		return
	}

	object, ok := data["object"].(map[string]interface{})
	if !ok {
		util.RespostaErro(c, http.StatusBadRequest, "Objeto do evento inválido", nil)
		return
	}

	var err error

	switch eventType {
	case "checkout.session.completed":
		// Cliente completou o checkout e iniciou trial ou pagamento
		sessionID, _ := object["id"].(string)
		subscriptionID, _ := object["subscription"].(string)
		metadata, _ := object["metadata"].(map[string]interface{})

		// Converter metadata para map[string]string
		metadataStr := make(map[string]string)
		for k, v := range metadata {
			if str, ok := v.(string); ok {
				metadataStr[k] = str
			}
		}

		err = ctrl.stripeServico.ProcessarCheckoutWebhook(sessionID, subscriptionID, metadataStr)

	case "customer.subscription.created":
		// Subscription criada (trial começou)
		subscriptionID, _ := object["id"].(string)
		customerID, _ := object["customer"].(string)
		status, _ := object["status"].(string)

		err = ctrl.stripeServico.ProcessarSubscriptionCriada(subscriptionID, customerID, status)

	case "customer.subscription.updated":
		// Subscription atualizada (trial terminou, pagamento feito, etc)
		subscriptionID, _ := object["id"].(string)
		status, _ := object["status"].(string)

		err = ctrl.stripeServico.ProcessarSubscriptionAtualizada(subscriptionID, status)

	case "customer.subscription.deleted":
		// Subscription cancelada
		subscriptionID, _ := object["id"].(string)

		err = ctrl.stripeServico.ProcessarSubscriptionCancelada(subscriptionID)
	}

	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao processar webhook", err)
		return
	}

	util.RespostaSucesso(c, "Webhook processado com sucesso", nil)
}
