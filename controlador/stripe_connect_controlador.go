package controlador

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ifinu/ifinu-api-go/dto"
	"github.com/ifinu/ifinu-api-go/middleware"
	"github.com/ifinu/ifinu-api-go/servico"
	"github.com/ifinu/ifinu-api-go/util"
)

type StripeConnectControlador struct {
	stripeConnectServico *servico.StripeConnectServico
}

func NovoStripeConnectControlador(stripeConnectServico *servico.StripeConnectServico) *StripeConnectControlador {
	return &StripeConnectControlador{
		stripeConnectServico: stripeConnectServico,
	}
}

// CriarContaConnect cria conta Stripe Connect e retorna link de onboarding
// POST /api/stripe-connect/criar-conta
func (ctrl *StripeConnectControlador) CriarContaConnect(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	// URLs de retorno
	baseURL := "https://app.ifinu.io" // TODO: Pegar de variável de ambiente
	returnURL := baseURL + "/painel/configuracoes?stripe-connect=sucesso"
	refreshURL := baseURL + "/painel/configuracoes?stripe-connect=refresh"

	resultado, err := ctrl.stripeConnectServico.CriarContaConnect(usuarioID, returnURL, refreshURL)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao criar conta Stripe Connect", err)
		return
	}

	util.RespostaSucesso(c, "Conta criada com sucesso", resultado)
}

// ObterStatus retorna status da conta conectada
// GET /api/stripe-connect/status
func (ctrl *StripeConnectControlador) ObterStatus(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	resultado, err := ctrl.stripeConnectServico.ObterStatusConnect(usuarioID)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao obter status", err)
		return
	}

	util.RespostaSucesso(c, "Status obtido com sucesso", resultado)
}

// RefreshOnboarding gera novo link de onboarding
// POST /api/stripe-connect/refresh-onboarding
func (ctrl *StripeConnectControlador) RefreshOnboarding(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	// URLs de retorno
	baseURL := "https://app.ifinu.io" // TODO: Pegar de variável de ambiente
	returnURL := baseURL + "/painel/configuracoes?stripe-connect=sucesso"
	refreshURL := baseURL + "/painel/configuracoes?stripe-connect=refresh"

	resultado, err := ctrl.stripeConnectServico.GerarLinkOnboarding(usuarioID, returnURL, refreshURL)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao gerar link de onboarding", err)
		return
	}

	util.RespostaSucesso(c, "Link gerado com sucesso", resultado)
}

// GerarDashboardLink gera link para dashboard Stripe
// GET /api/stripe-connect/dashboard-link
func (ctrl *StripeConnectControlador) GerarDashboardLink(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	resultado, err := ctrl.stripeConnectServico.GerarDashboardLink(usuarioID)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao gerar dashboard link", err)
		return
	}

	util.RespostaSucesso(c, "Dashboard link gerado", resultado)
}

// Desconectar remove conexão com Stripe Connect
// DELETE /api/stripe-connect/desconectar
func (ctrl *StripeConnectControlador) Desconectar(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	err := ctrl.stripeConnectServico.DesconectarConta(usuarioID)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao desconectar", err)
		return
	}

	util.RespostaSucesso(c, "Conta desconectada com sucesso", nil)
}

// WebhookAccountUpdated processa webhook de account.updated
// POST /api/stripe-connect/webhook
func (ctrl *StripeConnectControlador) WebhookAccountUpdated(c *gin.Context) {
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

	// Processar apenas eventos de account
	if eventType != "account.updated" {
		// Ignorar outros eventos
		util.RespostaSucesso(c, "Evento ignorado", nil)
		return
	}

	// Extrair dados
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

	accountID, _ := object["id"].(string)
	chargesEnabled, _ := object["charges_enabled"].(bool)
	detailsSubmitted, _ := object["details_submitted"].(bool)

	// Processar
	err := ctrl.stripeConnectServico.ProcessarAccountWebhook(accountID, chargesEnabled, detailsSubmitted)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao processar webhook", err)
		return
	}

	util.RespostaSucesso(c, "Webhook processado com sucesso", nil)
}
