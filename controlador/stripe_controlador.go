package controlador

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
