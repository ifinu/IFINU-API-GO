package controlador

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ifinu/ifinu-api-go/dto"
	"github.com/ifinu/ifinu-api-go/middleware"
	"github.com/ifinu/ifinu-api-go/servico"
	"github.com/ifinu/ifinu-api-go/util"
)

type WhatsAppControlador struct {
	whatsappServico *servico.WhatsAppServico
}

func NovoWhatsAppControlador(whatsappServico *servico.WhatsAppServico) *WhatsAppControlador {
	return &WhatsAppControlador{
		whatsappServico: whatsappServico,
	}
}

// Conectar inicia o processo de conexão do WhatsApp
// POST /api/whatsapp/conectar
func (ctrl *WhatsAppControlador) Conectar(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	resultado, err := ctrl.whatsappServico.Conectar(usuarioID)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, http.StatusOK, "QR Code gerado com sucesso", resultado)
}

// ObterStatus retorna o status da conexão WhatsApp
// GET /api/whatsapp/status
func (ctrl *WhatsAppControlador) ObterStatus(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	resultado, err := ctrl.whatsappServico.ObterStatus(usuarioID)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao obter status", err)
		return
	}

	util.RespostaSucesso(c, http.StatusOK, "Status obtido com sucesso", resultado)
}

// Desconectar desconecta o WhatsApp
// POST /api/whatsapp/desconectar
func (ctrl *WhatsAppControlador) Desconectar(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	err := ctrl.whatsappServico.Desconectar(usuarioID)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, http.StatusOK, "WhatsApp desconectado com sucesso", nil)
}

// EnviarMensagem envia uma mensagem via WhatsApp
// POST /api/whatsapp/enviar
func (ctrl *WhatsAppControlador) EnviarMensagem(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	var req dto.EnviarMensagemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "Dados inválidos", err)
		return
	}

	resultado, err := ctrl.whatsappServico.EnviarMensagem(usuarioID, req.Telefone, req.Mensagem)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, http.StatusOK, resultado.Mensagem, resultado)
}

// TestarConexao testa a conexão WhatsApp
// POST /api/whatsapp/testar
func (ctrl *WhatsAppControlador) TestarConexao(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	resultado, err := ctrl.whatsappServico.TestarConexao(usuarioID)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao testar conexão", err)
		return
	}

	util.RespostaSucesso(c, http.StatusOK, resultado.Mensagem, resultado)
}
