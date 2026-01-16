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
	assinaturaRepo := repositorio.NovoAssinaturaRepositorio(config.DB)
	stripeConfigRepo := repositorio.NovoStripeConfigRepositorio(config.DB)

	// Inicializar integrações
	evolutionAPI := integracao.NovoEvolutionAPICliente()

	// Inicializar integrações adicionais
	resendAPI := integracao.NovoResendCliente()

	// Inicializar services
	autenticacaoServico := servico.NovoAutenticacaoServico(usuarioRepo)
	clienteServico := servico.NovoClienteServico(clienteRepo)
	cobrancaServico := servico.NovoCobrancaServico(cobrancaRepo, clienteRepo)
	whatsappServico := servico.NovoWhatsAppServico(whatsappRepo, usuarioRepo, evolutionAPI)
	assinaturaServico := servico.NovoAssinaturaServico(assinaturaRepo, usuarioRepo)
	relatorioServico := servico.NovoRelatorioServico(clienteRepo, cobrancaRepo)
	stripeServico := servico.NovoStripeServico(usuarioRepo, assinaturaRepo)
	stripeConfigServico := servico.NovoStripeConfigServico(stripeConfigRepo)
	stripeConnectServico := servico.NovoStripeConnectServico(usuarioRepo)

	// Inicializar e iniciar agendador
	redisAddr := viper.GetString("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Fallback para desenvolvimento
	}
	agendadorServico := servico.NovoAgendadorServico(cobrancaRepo, whatsappRepo, usuarioRepo, assinaturaRepo, evolutionAPI, resendAPI, whatsappServico, redisAddr)
	agendadorServico.Iniciar()

	// Inicializar controllers
	autenticacaoController := controlador.NovoAutenticacaoControlador(autenticacaoServico)
	clienteController := controlador.NovoClienteControlador(clienteServico)
	cobrancaController := controlador.NovoCobrancaControlador(cobrancaServico)
	whatsappController := controlador.NovoWhatsAppControlador(whatsappServico)
	assinaturaController := controlador.NovoAssinaturaControlador(assinaturaServico)
	relatorioController := controlador.NovoRelatorioControlador(relatorioServico)
	stripeController := controlador.NovoStripeControlador(stripeServico)
	stripeConfigController := controlador.NovoStripeConfigControlador(stripeConfigServico)
	stripeConnectController := controlador.NovoStripeConnectControlador(stripeConnectServico)

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

	// Rotas de autenticação SEM /api (compatibilidade com frontend)
	authLegacy := r.Group("/auth")
	{
		authLegacy.POST("/login", autenticacaoController.Login)
		authLegacy.POST("/cadastro", autenticacaoController.Cadastro)
		authLegacy.POST("/refresh", autenticacaoController.RefreshToken)
		authLegacy.POST("/2fa/verificar", autenticacaoController.Verificar2FA)

		authLegacyProtegido := authLegacy.Group("")
		authLegacyProtegido.Use(middleware.AutenticacaoMiddleware())
		{
			authLegacyProtegido.GET("/me", autenticacaoController.Me)
			authLegacyProtegido.GET("/status-trial", autenticacaoController.StatusTrial)
			authLegacyProtegido.POST("/2fa/gerar", autenticacaoController.Gerar2FA)
			authLegacyProtegido.POST("/2fa/ativar", autenticacaoController.Ativar2FA)
		}
	}

	// Rotas de WhatsApp SEM /api (compatibilidade com frontend)
	whatsappLegacy := r.Group("/whatsapp")
	whatsappLegacy.Use(middleware.AutenticacaoMiddleware())
	whatsappLegacy.Use(middleware.AssinaturaMiddleware())
	{
		whatsappLegacy.POST("/conectar", whatsappController.Conectar)
		whatsappLegacy.GET("/status", whatsappController.ObterStatus)
		whatsappLegacy.GET("/qrcode", whatsappController.ObterQRCode)
		whatsappLegacy.POST("/desconectar", whatsappController.Desconectar)
		whatsappLegacy.POST("/enviar", whatsappController.EnviarMensagem)
		whatsappLegacy.POST("/testar", whatsappController.TestarConexao)
		whatsappLegacy.POST("/limpar-orfaos", whatsappController.LimparOrfaos)
		whatsappLegacy.GET("/estatisticas", whatsappController.ObterEstatisticas)
	}

	// Rotas de Stripe SEM /api (compatibilidade com frontend)
	stripeLegacy := r.Group("/stripe-trial")
	stripeLegacy.Use(middleware.AutenticacaoMiddleware())
	{
		stripeLegacy.POST("/create-checkout", stripeController.CreateCheckout)
	}

	// Grupo de rotas API
	api := r.Group("/api")
	{
		// Webhooks Stripe (público - sem autenticação)
		api.POST("/stripe/webhook", stripeController.WebhookStripe)
		api.POST("/stripe-connect/webhook", stripeConnectController.WebhookAccountUpdated)

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

		// Rotas protegidas apenas com autenticação (sem verificar assinatura)
		autenticado := api.Group("")
		autenticado.Use(middleware.AutenticacaoMiddleware())
		{
			// Rota de perfil (sem exigir assinatura ativa)
			autenticado.GET("/perfil", autenticacaoController.Me)

			// Rotas de segurança
			seguranca := autenticado.Group("/seguranca")
			{
				seguranca.POST("/alterar-senha", autenticacaoController.AlterarSenha)
			}

			// Rotas de configuração Stripe (sem exigir assinatura ativa)
			stripeConfig := autenticado.Group("/stripe")
			{
				stripeConfig.GET("/config", stripeConfigController.BuscarConfiguracao)
				stripeConfig.POST("/config", stripeConfigController.SalvarConfiguracao)
				stripeConfig.DELETE("/config", stripeConfigController.DeletarConfiguracao)
				stripeConfig.POST("/test-connection", stripeConfigController.TestarConexao)
			}

			// Rotas de Stripe Connect (sem exigir assinatura ativa)
			stripeConnect := autenticado.Group("/stripe-connect")
			{
				stripeConnect.POST("/criar-conta", stripeConnectController.CriarContaConnect)
				stripeConnect.GET("/status", stripeConnectController.ObterStatus)
				stripeConnect.GET("/account-status", stripeConnectController.ObterStatus)
				stripeConnect.POST("/refresh-onboarding", stripeConnectController.RefreshOnboarding)
				stripeConnect.POST("/create-onboarding-link", stripeConnectController.RefreshOnboarding)
				stripeConnect.GET("/dashboard-link", stripeConnectController.GerarDashboardLink)
				stripeConnect.DELETE("/desconectar", stripeConnectController.Desconectar)
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
				whatsapp.GET("/qrcode", whatsappController.ObterQRCode)
				whatsapp.POST("/desconectar", whatsappController.Desconectar)
				whatsapp.POST("/enviar", whatsappController.EnviarMensagem)
				whatsapp.POST("/testar", whatsappController.TestarConexao)
				whatsapp.POST("/limpar-orfaos", whatsappController.LimparOrfaos)
				whatsapp.GET("/estatisticas", whatsappController.ObterEstatisticas)
			}

			// Rotas de assinaturas
			assinaturas := protegido.Group("/assinaturas")
			{
				assinaturas.GET("/status", assinaturaController.Status)
				assinaturas.GET("/planos", stripeController.ListarPlanos)
				assinaturas.POST("/checkout", stripeController.CriarCheckoutAssinatura)
				assinaturas.POST("/cancelar", assinaturaController.Cancelar)
				assinaturas.GET("/historico", stripeController.BuscarHistoricoFaturas)
				assinaturas.GET("/detalhes", stripeController.BuscarDetalhesAssinatura)
			}

			// Rotas de relatórios
			relatorios := protegido.Group("/relatorios")
			{
				relatorios.GET("/dashboard", relatorioController.Dashboard)
			}

			// Rotas de Stripe Connect
			stripeConnect := protegido.Group("/stripe-connect")
			{
				stripeConnect.POST("/create-checkout-session", stripeController.CreateCheckoutSession)
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
