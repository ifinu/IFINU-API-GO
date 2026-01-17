package controlador

import (
	"fmt"
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

// HistoricoPagamentos retorna histórico completo de pagamentos
// GET /api/relatorios/pagamentos
func (ctrl *RelatorioControlador) HistoricoPagamentos(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	var pagina, tamanhoPagina int
	if p, ok := c.GetQuery("pagina"); ok {
		_, _ = fmt.Sscanf(p, "%d", &pagina)
	}
	if t, ok := c.GetQuery("tamanhoPagina"); ok {
		_, _ = fmt.Sscanf(t, "%d", &tamanhoPagina)
	}

	historico, err := ctrl.relatorioServico.ObterHistoricoPagamentos(usuarioID, pagina, tamanhoPagina)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao obter histórico de pagamentos", err)
		return
	}

	util.RespostaSucesso(c, "Histórico obtido com sucesso", historico)
}
