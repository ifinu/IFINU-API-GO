package controlador

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ifinu/ifinu-api-go/dto"
	"github.com/ifinu/ifinu-api-go/middleware"
	"github.com/ifinu/ifinu-api-go/servico"
	"github.com/ifinu/ifinu-api-go/util"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/account"
)

type StripeConfigControlador struct {
	stripeConfigServico *servico.StripeConfigServico
}

func NovoStripeConfigControlador(stripeConfigServico *servico.StripeConfigServico) *StripeConfigControlador {
	return &StripeConfigControlador{
		stripeConfigServico: stripeConfigServico,
	}
}

func (ctrl *StripeConfigControlador) BuscarConfiguracao(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	config, err := ctrl.stripeConfigServico.BuscarConfiguracao(usuarioID)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao buscar configuração", err)
		return
	}

	util.RespostaSucesso(c, "Configuração recuperada", config)
}

func (ctrl *StripeConfigControlador) SalvarConfiguracao(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	var req dto.SaveStripeConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "Dados inválidos", err)
		return
	}

	if req.PublishableKey == "" {
		util.RespostaErro(c, http.StatusBadRequest, "Chave pública é obrigatória", nil)
		return
	}

	err := ctrl.stripeConfigServico.SalvarConfiguracao(usuarioID, req)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, "Configuração salva com sucesso", nil)
}

func (ctrl *StripeConfigControlador) DeletarConfiguracao(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	err := ctrl.stripeConfigServico.DeletarConfiguracao(usuarioID)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, "Configuração removida com sucesso", nil)
}

func (ctrl *StripeConfigControlador) TestarConexao(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	chaveSecreta, err := ctrl.stripeConfigServico.ObterChaveSecreta(usuarioID)
	if err != nil {
		util.RespostaErro(c, http.StatusNotFound, "Configuração não encontrada", err)
		return
	}

	stripe.Key = chaveSecreta

	_, err = account.Get()
	if err != nil {
		util.RespostaSucesso(c, "", dto.TestConnectionResponse{
			Success: false,
			Message: "Falha ao conectar com Stripe. Verifique suas chaves.",
		})
		return
	}

	util.RespostaSucesso(c, "", dto.TestConnectionResponse{
		Success: true,
		Message: "Conexão com Stripe estabelecida com sucesso!",
	})
}
