package servico

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/ifinu/ifinu-api-go/dominio/entidades"
	"golang.org/x/time/rate"
)

const (
	FilaMensagensWhatsApp = "ifinu:fila:whatsapp"
	FilaMensagensEmail    = "ifinu:fila:email"
	MaxRetentativas       = 3
	TempoRetry            = 5 * time.Minute
)

type MensagemFila struct {
	ID              string                `json:"id"`
	TipoNotificacao string                `json:"tipo_notificacao"` // "lembrete", "vencimento", "pagamento"
	Cobranca        *entidades.Cobranca   `json:"cobranca"`
	Tentativas      int                   `json:"tentativas"`
	ProximaTentativa time.Time            `json:"proxima_tentativa"`
	CriadoEm        time.Time             `json:"criado_em"`
}

type FilaMensagemServico struct {
	redisClient  *redis.Client
	ctx          context.Context
	rateLimiter  *rate.Limiter
	whatsappSvc  *WhatsAppServico
	emailSvc     *ResendCliente
}

func NovoFilaMensagemServico(
	redisAddr string,
	whatsappSvc *WhatsAppServico,
	emailSvc *ResendCliente,
) *FilaMensagemServico {
	ctx := context.Background()

	redisClient := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     "",
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     100,
		MinIdleConns: 10,
	})

	// Testar conexão
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️  Redis não disponível: %v. Fila desabilitada.", err)
		return nil
	}

	// Rate Limiter: 50 mensagens/segundo
	limiter := rate.NewLimiter(rate.Limit(50), 100) // 50 req/s, burst de 100

	log.Println("✅ Fila de mensagens Redis conectada")

	return &FilaMensagemServico{
		redisClient: redisClient,
		ctx:         ctx,
		rateLimiter: limiter,
		whatsappSvc: whatsappSvc,
		emailSvc:    emailSvc,
	}
}

// EnfileirarMensagem adiciona mensagem na fila para processamento assíncrono
func (s *FilaMensagemServico) EnfileirarMensagem(msg *MensagemFila) error {
	if s == nil || s.redisClient == nil {
		return fmt.Errorf("fila não inicializada")
	}

	msg.CriadoEm = time.Now()
	msg.ProximaTentativa = time.Now()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %w", err)
	}

	// Adicionar na fila Redis (LPUSH = adiciona no início)
	err = s.redisClient.LPush(s.ctx, FilaMensagensWhatsApp, data).Err()
	if err != nil {
		return fmt.Errorf("erro ao enfileirar mensagem: %w", err)
	}

	log.Printf("📥 Mensagem enfileirada: %s (Cobrança ID: %d)", msg.TipoNotificacao, msg.Cobranca.ID)
	return nil
}

// IniciarWorkerPool inicia pool de workers para processar fila
func (s *FilaMensagemServico) IniciarWorkerPool(numWorkers int) {
	if s == nil || s.redisClient == nil {
		log.Println("⚠️  Fila não disponível. Worker pool desabilitado.")
		return
	}

	log.Printf("🚀 Iniciando %d workers para processar fila de mensagens", numWorkers)

	for i := 1; i <= numWorkers; i++ {
		go s.worker(i)
	}

	// Worker para limpar mensagens antigas
	go s.limparMensagensAntigas()
}

// worker processa mensagens da fila
func (s *FilaMensagemServico) worker(id int) {
	log.Printf("👷 Worker %d iniciado", id)

	for {
		// Aguardar rate limiter
		if err := s.rateLimiter.Wait(s.ctx); err != nil {
			log.Printf("❌ Worker %d: erro no rate limiter: %v", id, err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Buscar próxima mensagem da fila (BRPOP = bloqueante, aguarda até ter mensagem)
		result, err := s.redisClient.BRPop(s.ctx, 5*time.Second, FilaMensagensWhatsApp).Result()
		if err == redis.Nil {
			// Timeout, nenhuma mensagem disponível
			continue
		}
		if err != nil {
			log.Printf("❌ Worker %d: erro ao buscar mensagem: %v", id, err)
			time.Sleep(1 * time.Second)
			continue
		}

		// result[0] = nome da fila, result[1] = dados
		if len(result) < 2 {
			continue
		}

		var msg MensagemFila
		if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
			log.Printf("❌ Worker %d: erro ao deserializar mensagem: %v", id, err)
			continue
		}

		// Processar mensagem
		s.processarMensagem(id, &msg)
	}
}

// processarMensagem processa uma mensagem individual
func (s *FilaMensagemServico) processarMensagem(workerID int, msg *MensagemFila) {
	log.Printf("⚙️  Worker %d processando: %s (Tentativa %d/%d)",
		workerID, msg.TipoNotificacao, msg.Tentativas+1, MaxRetentativas)

	// Verificar se já passou o tempo de retry
	if time.Now().Before(msg.ProximaTentativa) {
		// Re-enfileirar para processar depois
		s.reenfileirarMensagem(msg)
		return
	}

	// Enviar mensagem WhatsApp
	sucesso := s.enviarWhatsApp(msg)

	if !sucesso {
		msg.Tentativas++
		if msg.Tentativas < MaxRetentativas {
			// Re-enfileirar com delay exponencial
			delay := time.Duration(msg.Tentativas*msg.Tentativas) * TempoRetry
			msg.ProximaTentativa = time.Now().Add(delay)
			s.reenfileirarMensagem(msg)
			log.Printf("🔄 Worker %d: Re-enfileirando mensagem. Próxima tentativa em %v",
				workerID, delay)
		} else {
			log.Printf("❌ Worker %d: Mensagem descartada após %d tentativas falhas",
				workerID, MaxRetentativas)
			// TODO: Salvar em DLQ (Dead Letter Queue) para análise
		}
	} else {
		log.Printf("✅ Worker %d: Mensagem processada com sucesso", workerID)
	}
}

// enviarWhatsApp envia mensagem via WhatsApp
func (s *FilaMensagemServico) enviarWhatsApp(msg *MensagemFila) bool {
	cobranca := msg.Cobranca

	// Montar mensagem baseada no tipo
	var textoMensagem string
	switch msg.TipoNotificacao {
	case "lembrete":
		textoMensagem = fmt.Sprintf(
			"🔔 *Lembrete de Cobrança*\n\n"+
				"Olá, %s!\n\n"+
				"Sua cobrança vence em 3 dias:\n"+
				"💰 Valor: R$ %.2f\n"+
				"📝 Descrição: %s\n"+
				"📅 Vencimento: %s\n\n"+
				"Atenciosamente,\nEquipe IFINU",
			cobranca.Cliente.Nome,
			cobranca.Valor,
			cobranca.Descricao,
			cobranca.DataVencimento.Format("02/01/2006"),
		)
	case "vencimento":
		textoMensagem = fmt.Sprintf(
			"⚠️ *Cobrança Vence Hoje*\n\n"+
				"Olá, %s!\n\n"+
				"Sua cobrança vence HOJE:\n"+
				"💰 Valor: R$ %.2f\n"+
				"📝 Descrição: %s\n\n"+
				"Atenciosamente,\nEquipe IFINU",
			cobranca.Cliente.Nome,
			cobranca.Valor,
			cobranca.Descricao,
		)
	default:
		log.Printf("⚠️  Tipo de notificação desconhecido: %s", msg.TipoNotificacao)
		return false
	}

	// Enviar via WhatsApp
	_, err := s.whatsappSvc.EnviarMensagem(
		cobranca.UsuarioID,
		cobranca.Cliente.Telefone,
		textoMensagem,
	)

	return err == nil
}

// reenfileirarMensagem coloca mensagem de volta na fila
func (s *FilaMensagemServico) reenfileirarMensagem(msg *MensagemFila) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("❌ Erro ao re-enfileirar mensagem: %v", err)
		return
	}

	// Adicionar no final da fila (LPUSH)
	s.redisClient.LPush(s.ctx, FilaMensagensWhatsApp, data)
}

// limparMensagensAntigas remove mensagens muito antigas da fila
func (s *FilaMensagemServico) limparMensagensAntigas() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		// Obter tamanho da fila
		tamanho, err := s.redisClient.LLen(s.ctx, FilaMensagensWhatsApp).Result()
		if err != nil {
			continue
		}

		if tamanho > 0 {
			log.Printf("📊 Fila de mensagens: %d pendentes", tamanho)
		}

		// TODO: Implementar limpeza de mensagens muito antigas (>24h)
	}
}

// ObterEstatisticas retorna estatísticas da fila
func (s *FilaMensagemServico) ObterEstatisticas() (map[string]interface{}, error) {
	if s == nil || s.redisClient == nil {
		return nil, fmt.Errorf("fila não inicializada")
	}

	tamanho, err := s.redisClient.LLen(s.ctx, FilaMensagensWhatsApp).Result()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"mensagens_pendentes": tamanho,
		"rate_limit":          "50 msg/s",
		"max_burst":           100,
	}, nil
}
