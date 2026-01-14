package controlador

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ifinu/ifinu-api-go/middleware"
	"github.com/ifinu/ifinu-api-go/servico"
	"github.com/ifinu/ifinu-api-go/util"
)

type AssinaturaControlador struct {
	assinaturaServico *servico.AssinaturaServico
}

func NovoAssinaturaControlador(assinaturaServico *servico.AssinaturaServico) *AssinaturaControlador {
	return &AssinaturaControlador{
		assinaturaServico: assinaturaServico,
	}
}

// Status retorna o status da assinatura do usuário
// GET /api/assinaturas/status
func (ctrl *AssinaturaControlador) Status(c *gin.Context) {
	email, exists := middleware.ObterEmailUsuario(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	status, err := ctrl.assinaturaServico.ObterStatus(email)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao obter status da assinatura", err)
		return
	}

	util.RespostaSucesso(c, "Status da assinatura obtido com sucesso", status)
}

// Cancelar cancela a assinatura do usuário
// POST /api/assinaturas/cancelar
func (ctrl *AssinaturaControlador) Cancelar(c *gin.Context) {
	email, exists := middleware.ObterEmailUsuario(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	err := ctrl.assinaturaServico.CancelarAssinatura(email)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao cancelar assinatura", err)
		return
	}

	util.RespostaSucesso(c, "Assinatura cancelada com sucesso", nil)
}
