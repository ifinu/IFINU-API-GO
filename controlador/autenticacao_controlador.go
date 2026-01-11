package controlador

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ifinu/ifinu-api-go/dto"
	"github.com/ifinu/ifinu-api-go/middleware"
	"github.com/ifinu/ifinu-api-go/servico"
	"github.com/ifinu/ifinu-api-go/util"
)

type AutenticacaoControlador struct {
	autenticacaoServico *servico.AutenticacaoServico
}

func NovoAutenticacaoControlador(autenticacaoServico *servico.AutenticacaoServico) *AutenticacaoControlador {
	return &AutenticacaoControlador{
		autenticacaoServico: autenticacaoServico,
	}
}

// Login realiza o login do usuário
// POST /api/auth/login
func (ctrl *AutenticacaoControlador) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "Dados inválidos", err)
		return
	}

	resultado, err := ctrl.autenticacaoServico.Login(req)
	if err != nil {
		if err.Error() == "2FA_NECESSARIO" {
			util.RespostaErro(c, http.StatusForbidden, "Autenticação de dois fatores necessária", nil)
			return
		}
		util.RespostaErro(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, http.StatusOK, "Login realizado com sucesso", resultado)
}

// Cadastro registra um novo usuário
// POST /api/auth/cadastro
func (ctrl *AutenticacaoControlador) Cadastro(c *gin.Context) {
	var req dto.CadastroRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "Dados inválidos", err)
		return
	}

	resultado, err := ctrl.autenticacaoServico.Cadastro(req)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, http.StatusCreated, "Cadastro realizado com sucesso", resultado)
}

// RefreshToken renova o access token
// POST /api/auth/refresh
func (ctrl *AutenticacaoControlador) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "Dados inválidos", err)
		return
	}

	resultado, err := ctrl.autenticacaoServico.RefreshToken(req)
	if err != nil {
		util.RespostaErro(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, http.StatusOK, "Token renovado com sucesso", resultado)
}

// Me retorna os dados do usuário autenticado
// GET /api/auth/me
func (ctrl *AutenticacaoControlador) Me(c *gin.Context) {
	email, exists := middleware.ObterEmailUsuario(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	usuario, err := ctrl.autenticacaoServico.BuscarUsuarioPorEmail(email)
	if err != nil {
		util.RespostaErro(c, http.StatusNotFound, "Usuário não encontrado", nil)
		return
	}

	util.RespostaSucesso(c, http.StatusOK, "Dados do usuário", usuario)
}

// Gerar2FA gera o QR code para configurar 2FA
// POST /api/auth/2fa/gerar
func (ctrl *AutenticacaoControlador) Gerar2FA(c *gin.Context) {
	email, exists := middleware.ObterEmailUsuario(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	resultado, err := ctrl.autenticacaoServico.Gerar2FA(email)
	if err != nil {
		util.RespostaErro(c, http.StatusInternalServerError, "Erro ao gerar 2FA", err)
		return
	}

	util.RespostaSucesso(c, http.StatusOK, "QR Code gerado com sucesso", resultado)
}

// Ativar2FA ativa o 2FA após validar o código
// POST /api/auth/2fa/ativar
func (ctrl *AutenticacaoControlador) Ativar2FA(c *gin.Context) {
	email, exists := middleware.ObterEmailUsuario(c)
	if !exists {
		util.RespostaErro(c, http.StatusUnauthorized, "Usuário não autenticado", nil)
		return
	}

	var req dto.Ativar2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "Dados inválidos", err)
		return
	}

	err := ctrl.autenticacaoServico.Ativar2FA(email, req.Codigo)
	if err != nil {
		util.RespostaErro(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, http.StatusOK, "2FA ativado com sucesso", nil)
}

// Verificar2FA valida o código 2FA no login
// POST /api/auth/2fa/verificar
func (ctrl *AutenticacaoControlador) Verificar2FA(c *gin.Context) {
	var req dto.Verificar2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.RespostaErro(c, http.StatusBadRequest, "Dados inválidos", err)
		return
	}

	resultado, err := ctrl.autenticacaoServico.Verificar2FA(req)
	if err != nil {
		util.RespostaErro(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	util.RespostaSucesso(c, http.StatusOK, "Login com 2FA realizado com sucesso", resultado)
}
