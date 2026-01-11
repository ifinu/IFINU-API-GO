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

type ClienteControlador struct {
	clienteServico *servico.ClienteServico
}

func NovoClienteControlador(clienteServico *servico.ClienteServico) *ClienteControlador {
	return &ClienteControlador{
		clienteServico: clienteServico,
	}
}

// Criar cria um novo cliente
// POST /api/clientes
func (ctrl *ClienteControlador) Criar(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	var req dto.ClienteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "Dados inválidos", err)
		return
	}

	resultado, err := ctrl.clienteServico.Criar(usuarioID, req)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, http.StatusCreated, "Cliente criado com sucesso", resultado)
}

// BuscarPorID busca um cliente por ID
// GET /api/clientes/:id
func (ctrl *ClienteControlador) BuscarPorID(c *gin.Context) {
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

	resultado, err := ctrl.clienteServico.BuscarPorID(usuarioID, id)
	if err != nil {
		util.RespostaErro(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, http.StatusOK, "Cliente encontrado", resultado)
}

// Listar lista todos os clientes do usuário
// GET /api/clientes
func (ctrl *ClienteControlador) Listar(c *gin.Context) {
	usuarioID, exists := middleware.ObterUsuarioID(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	// Verificar se tem parâmetros de busca
	var req dto.BuscarClientesRequest
	if err := c.ShouldBindQuery(&req); err == nil && (req.Termo != "" || req.Pagina > 0) {
		// Busca com filtros
		resultado, err := ctrl.clienteServico.Buscar(usuarioID, req)
		if err != nil {
			util.RespostaErro(c, http.StatusInternalServerError, "Erro ao buscar clientes", err)
			return
		}
		util.RespostaSucesso(c, http.StatusOK, "Clientes encontrados", resultado)
		return
	}

	// Listagem simples
	resultado, err := ctrl.clienteServico.Listar(usuarioID)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao listar clientes", err)
		return
	}

	util.RespostaSucesso(c, http.StatusOK, "Clientes listados com sucesso", resultado)
}

// Atualizar atualiza um cliente
// PUT /api/clientes/:id
func (ctrl *ClienteControlador) Atualizar(c *gin.Context) {
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

	var req dto.ClienteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "Dados inválidos", err)
		return
	}

	resultado, err := ctrl.clienteServico.Atualizar(usuarioID, id, req)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, http.StatusOK, "Cliente atualizado com sucesso", resultado)
}

// Deletar remove um cliente
// DELETE /api/clientes/:id
func (ctrl *ClienteControlador) Deletar(c *gin.Context) {
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

	err = ctrl.clienteServico.Deletar(usuarioID, id)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, http.StatusOK, "Cliente deletado com sucesso", nil)
}
