package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/ifinu/ifinu-api-go/config"
	"github.com/spf13/viper"
)

func main() {
	// Banner
	fmt.Println(`
	╔═══════════════════════════════════════╗
	║      IFINU API GO - v1.0.0          ║
	║   Sistema de Cobrança Online         ║
	║   Migrado de Java para Go            ║
	╚═══════════════════════════════════════╝
	`)

	// Carregar configurações
	if err := config.CarregarConfiguracoes(); err != nil {
		log.Fatalf("❌ Erro ao carregar configurações: %v", err)
	}

	// Conectar ao banco de dados
	if err := config.ConectarBancoDados(); err != nil {
		log.Fatalf("❌ Erro ao conectar ao banco: %v", err)
	}

	// Configurar Gin
	if viper.GetString("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Middleware de CORS
	r.Use(corsMiddleware())

	// Rotas de saúde
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "IFINU API GO está rodando",
			"version": "1.0.0",
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Bem-vindo à IFINU API GO",
			"docs":    "/docs",
			"health":  "/health",
		})
	})

	// Grupo de rotas API
	api := r.Group("/api")
	{
		// Rotas de autenticação (públicas)
		auth := api.Group("/auth")
		{
			auth.GET("/status", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "Endpoint de autenticação funcionando",
				})
			})
			// TODO: Adicionar controllers de auth
		}

		// Rotas protegidas (requerem autenticação)
		// TODO: Adicionar middleware de autenticação
		clientes := api.Group("/clientes")
		{
			clientes.GET("/", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "Listagem de clientes",
					"data":    []string{},
				})
			})
		}

		cobrancas := api.Group("/cobrancas")
		{
			cobrancas.GET("/", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "Listagem de cobranças",
					"data":    []string{},
				})
			})
		}
	}

	// Iniciar servidor
	porta := viper.GetString("APP_PORT")
	log.Printf("🚀 Servidor iniciando na porta %s...", porta)

	if err := r.Run(":" + porta); err != nil {
		log.Fatalf("❌ Erro ao iniciar servidor: %v", err)
	}
}

// corsMiddleware adiciona headers CORS
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
