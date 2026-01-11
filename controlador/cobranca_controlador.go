package controlador

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ifinu/ifinu-api-go/dto"
	"github.com/ifinu/ifinu-api-go/middleware"
	"github.com/ifinu/ifinu-api-go/servico"
	"github.com/ifinu/ifinu-api-go/util"
)

type CobrancaControlador struct {
	cobrancaServico *servico.CobrancaServico
}

func NovoCobrancaControlador(cobrancaServico *servico.CobrancaServico) *CobrancaControlador {
	return &CobrancaControlador{
		cobrancaServico: cobrancaServico,
	}
}

// Criar cria uma nova cobrança
// POST /api/cobrancas
func (ctrl *CobrancaControlador) Criar(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	var req dto.CobrancaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "Dados inválidos", err)
		return
	}

	resultado, err := ctrl.cobrancaServico.Criar(usuarioID, req)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	util.RespostaCriado(c, "Cobrança criada com sucesso", resultado)
}

// BuscarPorID busca uma cobrança por ID
// GET /api/cobrancas/:id
func (ctrl *CobrancaControlador) BuscarPorID(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "ID inválido", nil)
		return
	}

	resultado, err := ctrl.cobrancaServico.BuscarPorID(usuarioID, id)
	if err != nil {
		util.RespostaErro(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, "Cobrança encontrada", resultado)
}

// Listar lista todas as cobranças do usuário
// GET /api/cobrancas
func (ctrl *CobrancaControlador) Listar(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	// Verificar se tem parâmetros de busca
	var req dto.BuscarCobrancasRequest
	if err := c.ShouldBindQuery(&req); err == nil && req.Pagina > 0 {
		// Busca com filtros
		resultado, err := ctrl.cobrancaServico.Buscar(usuarioID, req)
		if err != nil {
			util.RespostaErro(c, http.StatusInternalServerError, "Erro ao buscar cobranças", err)
			return
		}
		util.RespostaSucesso(c, "Cobranças encontradas", resultado)
		return
	}

	// Listagem simples
	resultado, err := ctrl.cobrancaServico.Listar(usuarioID)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao listar cobranças", err)
		return
	}

	util.RespostaSucesso(c, "Cobranças listadas com sucesso", resultado)
}

// Atualizar atualiza uma cobrança
// PUT /api/cobrancas/:id
func (ctrl *CobrancaControlador) Atualizar(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "ID inválido", nil)
		return
	}

	var req dto.CobrancaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "Dados inválidos", err)
		return
	}

	resultado, err := ctrl.cobrancaServico.Atualizar(usuarioID, id, req)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, "Cobrança atualizada com sucesso", resultado)
}

// AtualizarStatus atualiza o status de uma cobrança
// PATCH /api/cobrancas/:id/status
func (ctrl *CobrancaControlador) AtualizarStatus(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "ID inválido", nil)
		return
	}

	var req dto.AtualizarStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "Dados inválidos", err)
		return
	}

	resultado, err := ctrl.cobrancaServico.AtualizarStatus(usuarioID, id, req.Status)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, "Status atualizado com sucesso", resultado)
}

// Deletar remove uma cobrança
// DELETE /api/cobrancas/:id
func (ctrl *CobrancaControlador) Deletar(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "ID inválido", nil)
		return
	}

	err = ctrl.cobrancaServico.Deletar(usuarioID, id)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, "Cobrança deletada com sucesso", nil)
}

// ObterEstatisticas retorna estatísticas de cobranças
// GET /api/cobrancas/estatisticas
func (ctrl *CobrancaControlador) ObterEstatisticas(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	resultado, err := ctrl.cobrancaServico.ObterEstatisticas(usuarioID)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao obter estatísticas", err)
		return
	}

	util.RespostaSucesso(c, "Estatísticas obtidas com sucesso", resultado)
}
