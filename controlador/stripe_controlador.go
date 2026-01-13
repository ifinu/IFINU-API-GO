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
