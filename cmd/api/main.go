package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/ifinu/ifinu-api-go/config"
	"github.com/ifinu/ifinu-api-go/controlador"
	"github.com/ifinu/ifinu-api-go/integracao"
	"github.com/ifinu/ifinu-api-go/middleware"
	"github.com/ifinu/ifinu-api-go/repositorio"
	"github.com/ifinu/ifinu-api-go/servico"
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

	// Inicializar repositórios
	usuarioRepo := repositorio.NovoUsuarioRepositorio(config.DB)
	clienteRepo := repositorio.NovoClienteRepositorio(config.DB)
	cobrancaRepo := repositorio.NovoCobrancaRepositorio(config.DB)
	whatsappRepo := repositorio.NovoWhatsAppRepositorio(config.DB)

	// Inicializar integrações
	evolutionAPI := integracao.NovoEvolutionAPICliente()

	// Inicializar integrações adicionais
	resendAPI := integracao.NovoResendCliente()

	// Inicializar services
	autenticacaoServico := servico.NovoAutenticacaoServico(usuarioRepo)
	clienteServico := servico.NovoClienteServico(clienteRepo)
	cobrancaServico := servico.NovoCobrancaServico(cobrancaRepo, clienteRepo)
	whatsappServico := servico.NovoWhatsAppServico(whatsappRepo, usuarioRepo, evolutionAPI)

	// Inicializar e iniciar agendador
	agendadorServico := servico.NovoAgendadorServico(cobrancaRepo, whatsappRepo, evolutionAPI, resendAPI)
	agendadorServico.Iniciar()

	// Inicializar controllers
	autenticacaoController := controlador.NovoAutenticacaoControlador(autenticacaoServico)
	clienteController := controlador.NovoClienteControlador(clienteServico)
	cobrancaController := controlador.NovoCobrancaControlador(cobrancaServico)
	whatsappController := controlador.NovoWhatsAppControlador(whatsappServico)

	// Configurar Gin
	if viper.GetString("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Middleware de CORS
	r.Use(corsMiddleware())

	// Rotas públicas
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
			auth.POST("/login", autenticacaoController.Login)
			auth.POST("/cadastro", autenticacaoController.Cadastro)
			auth.POST("/refresh", autenticacaoController.RefreshToken)
			auth.POST("/2fa/verificar", autenticacaoController.Verificar2FA)

			// Rotas protegidas de autenticação
			authProtegido := auth.Group("")
			authProtegido.Use(middleware.AutenticacaoMiddleware())
			{
				authProtegido.GET("/me", autenticacaoController.Me)
				authProtegido.POST("/2fa/gerar", autenticacaoController.Gerar2FA)
				authProtegido.POST("/2fa/ativar", autenticacaoController.Ativar2FA)
			}
		}

		// Rotas protegidas (requerem autenticação e assinatura ativa)
		protegido := api.Group("")
		protegido.Use(middleware.AutenticacaoMiddleware())
		protegido.Use(middleware.AssinaturaMiddleware())
		{
			// Rotas de clientes
			clientes := protegido.Group("/clientes")
			{
				clientes.GET("", clienteController.Listar)
				clientes.POST("", clienteController.Criar)
				clientes.GET("/:id", clienteController.BuscarPorID)
				clientes.PUT("/:id", clienteController.Atualizar)
				clientes.DELETE("/:id", clienteController.Deletar)
			}

			// Rotas de cobranças
			cobrancas := protegido.Group("/cobrancas")
			{
				cobrancas.GET("", cobrancaController.Listar)
				cobrancas.POST("", cobrancaController.Criar)
				cobrancas.GET("/estatisticas", cobrancaController.ObterEstatisticas)
				cobrancas.GET("/:id", cobrancaController.BuscarPorID)
				cobrancas.PUT("/:id", cobrancaController.Atualizar)
				cobrancas.PATCH("/:id/status", cobrancaController.AtualizarStatus)
				cobrancas.DELETE("/:id", cobrancaController.Deletar)
			}

			// Rotas de WhatsApp
			whatsapp := protegido.Group("/whatsapp")
			{
				whatsapp.POST("/conectar", whatsappController.Conectar)
				whatsapp.GET("/status", whatsappController.ObterStatus)
				whatsapp.POST("/desconectar", whatsappController.Desconectar)
				whatsapp.POST("/enviar", whatsappController.EnviarMensagem)
				whatsapp.POST("/testar", whatsappController.TestarConexao)
			}
		}
	}

	// Iniciar servidor
	porta := viper.GetString("APP_PORT")
	if porta == "" {
		porta = "8080"
	}

	log.Printf("🚀 Servidor iniciando na porta %s...", porta)
	log.Printf("📚 Documentação disponível em: http://localhost:%s/", porta)
	log.Printf("💚 Health check: http://localhost:%s/health", porta)

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
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
