package util

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RespostaPadrao estrutura de resposta padrão da API
type RespostaPadrao struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Details   interface{} `json:"details,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// RespostaSucesso retorna resposta de sucesso
func RespostaSucesso(c *gin.Context, mensagem string, dados interface{}) {
	c.JSON(http.StatusOK, RespostaPadrao{
		Success:   true,
		Message:   mensagem,
		Data:      dados,
		Timestamp: time.Now(),
	})
}

// RespostaErro retorna resposta de erro
func RespostaErro(c *gin.Context, status int, erro string, detalhes interface{}) {
	c.JSON(status, RespostaPadrao{
		Success:   false,
		Error:     erro,
		Details:   detalhes,
		Timestamp: time.Now(),
	})
}

// RespostaCriado retorna resposta 201 Created
func RespostaCriado(c *gin.Context, mensagem string, dados interface{}) {
	c.JSON(http.StatusCreated, RespostaPadrao{
		Success:   true,
		Message:   mensagem,
		Data:      dados,
		Timestamp: time.Now(),
	})
}

// RespostaNaoAutorizado retorna 401 Unauthorized
func RespostaNaoAutorizado(c *gin.Context, mensagem string) {
	c.JSON(http.StatusUnauthorized, RespostaPadrao{
		Success:   false,
		Error:     "Não autorizado",
		Details:   mensagem,
		Timestamp: time.Now(),
	})
}

// RespostaNaoEncontrado retorna 404 Not Found
func RespostaNaoEncontrado(c *gin.Context, mensagem string) {
	c.JSON(http.StatusNotFound, RespostaPadrao{
		Success:   false,
		Error:     "Não encontrado",
		Details:   mensagem,
		Timestamp: time.Now(),
	})
}

// RespostaForbidden retorna 403 Forbidden
func RespostaForbidden(c *gin.Context, mensagem string) {
	c.JSON(http.StatusForbidden, RespostaPadrao{
		Success:   false,
		Error:     "Acesso negado",
		Details:   mensagem,
		Timestamp: time.Now(),
	})
}
