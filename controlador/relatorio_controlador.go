package controlador

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ifinu/ifinu-api-go/middleware"
	"github.com/ifinu/ifinu-api-go/servico"
	"github.com/ifinu/ifinu-api-go/util"
)

type RelatorioControlador struct {
	relatorioServico *servico.RelatorioServico
}

func NovoRelatorioControlador(relatorioServico *servico.RelatorioServico) *RelatorioControlador {
	return &RelatorioControlador{
		relatorioServico: relatorioServico,
	}
}

// Dashboard retorna estatísticas do dashboard
// GET /api/relatorios/dashboard
func (ctrl *RelatorioControlador) Dashboard(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	dashboard, err := ctrl.relatorioServico.ObterDashboard(usuarioID)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao obter dados do dashboard", err)
		return
	}

	util.RespostaSucesso(c, "Dashboard obtido com sucesso", dashboard)
}
