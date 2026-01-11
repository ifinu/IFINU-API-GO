package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/config"
	"github.com/ifinu/ifinu-api-go/repositorio"
)

// AssinaturaMiddleware verifica se o usuário tem assinatura ativa ou trial válido
func AssinaturaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obter email do usuário do contexto (já validado pelo AuthMiddleware)
		email, exists := c.Get("emailUsuario")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Usuário não autenticado",
			})
			c.Abort()
			return
		}

		emailStr := email.(string)

		// Buscar usuário
		usuarioRepo := repositorio.NovoUsuarioRepositorio(config.DB)
		usuario, err := usuarioRepo.BuscarPorEmail(emailStr)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Usuário não encontrado",
			})
			c.Abort()
			return
		}

		// Verificar se tem trial ativo e não expirado
		if usuario.TrialAtivo && !usuario.IsTrialExpirado() {
			c.Set("usuarioID", usuario.ID)
			c.Next()
			return
		}

		// Verificar se tem assinatura ativa
		assinaturaRepo := repositorio.NovoAssinaturaRepositorio(config.DB)
		assinatura, err := assinaturaRepo.BuscarPorUsuario(usuario.ID)
		if err == nil && assinatura.Ativa && assinatura.IsAtiva() {
			c.Set("usuarioID", usuario.ID)
			c.Next()
			return
		}

		// Não tem trial nem assinatura válidos
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Assinatura expirada ou trial encerrado. Por favor, renove sua assinatura.",
			"code":    "ASSINATURA_EXPIRADA",
		})
		c.Abort()
	}
}

// ObterUsuarioID retorna o ID do usuário do contexto
func ObterUsuarioID(c *gin.Context) (uuid.UUID, bool) {
	usuarioID, exists := c.Get("usuarioID")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := usuarioID.(uuid.UUID)
	return id, ok
}
