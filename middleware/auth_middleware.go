package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ifinu/ifinu-api-go/util"
)

// AutenticacaoMiddleware valida o token JWT
func AutenticacaoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Pegar o token do header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token não fornecido",
			})
			c.Abort()
			return
		}

		// Verificar se começa com "Bearer "
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Formato de token inválido",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validar o token
		claims, err := util.ValidarToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token inválido ou expirado",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// Adicionar email do usuário ao contexto
		c.Set("emailUsuario", claims.Email)
		c.Next()
	}
}

// ObterEmailUsuario retorna o email do usuário do contexto
func ObterEmailUsuario(c *gin.Context) (string, bool) {
	email, exists := c.Get("emailUsuario")
	if !exists {
		return "", false
	}
	emailStr, ok := email.(string)
	return emailStr, ok
}
