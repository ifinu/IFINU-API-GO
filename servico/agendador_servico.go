package servico

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ifinu/ifinu-api-go/dominio/entidades"
	"github.com/ifinu/ifinu-api-go/dominio/enums"
	"github.com/ifinu/ifinu-api-go/integracao"
	"github.com/ifinu/ifinu-api-go/repositorio"
	"github.com/robfig/cron/v3"
)

type AgendadorServico struct {
	cobrancaRepo *repositorio.CobrancaRepositorio
	whatsappRepo *repositorio.WhatsAppRepositorio
	evolutionAPI *integracao.EvolutionAPICliente
	resendAPI    *integracao.ResendCliente
	cron         *cron.Cron
}

func NovoAgendadorServico(
	cobrancaRepo *repositorio.CobrancaRepositorio,
	whatsappRepo *repositorio.WhatsAppRepositorio,
	evolutionAPI *integracao.EvolutionAPICliente,
	resendAPI *integracao.ResendCliente,
) *AgendadorServico {
	return &AgendadorServico{
		cobrancaRepo: cobrancaRepo,
		whatsappRepo: whatsappRepo,
		evolutionAPI: evolutionAPI,
		resendAPI:    resendAPI,
		cron:         cron.New(),
	}
}

// Iniciar inicia o agendador de tarefas
func (s *AgendadorServico) Iniciar() {
	log.Println("📅 Iniciando agendador de notificações...")

	// Enviar notificações de lembrete (3 dias antes) - executa todos os dias às 9h
	s.cron.AddFunc("0 9 * * *", func() {
		log.Println("⏰ Executando job: Notificações de lembrete")
		s.EnviarNotificacoesLembrete()
	})

	// Enviar notificações de vencimento (dia do vencimento) - executa todos os dias às 9h
	s.cron.AddFunc("0 9 * * *", func() {
		log.Println("⏰ Executando job: Notificações de vencimento")
		s.EnviarNotificacoesVencimento()
	})

	// Verificar cobranças vencidas - executa todos os dias às 23h
	s.cron.AddFunc("0 23 * * *", func() {
		log.Println("⏰ Executando job: Atualizar cobranças vencidas")
		s.AtualizarCobrancasVencidas()
	})

	s.cron.Start()
	log.Println("✅ Agendador iniciado com sucesso")
}

// Parar para o agendador
func (s *AgendadorServico) Parar() {
	log.Println("🛑 Parando agendador...")
	s.cron.Stop()
}

// EnviarNotificacoesLembrete envia notificações de lembrete (3 dias antes do vencimento)
func (s *AgendadorServico) EnviarNotificacoesLembrete() {
	cobrancas, err := s.cobrancaRepo.BuscarCobrancasParaLembrete()
	if err != nil {
		log.Printf("❌ Erro ao buscar cobranças para lembrete: %v", err)
		return
	}

	if len(cobrancas) == 0 {
		log.Println("📭 Nenhuma cobrança para enviar lembrete")
		return
	}

	log.Printf("📬 Enviando %d notificações de lembrete...", len(cobrancas))

	// Usar goroutines para enviar notificações em paralelo
	var wg sync.WaitGroup
	for _, cobranca := range cobrancas {
		wg.Add(1)
		go func(c *entidades.Cobranca) {
			defer wg.Done()
			s.enviarNotificacaoLembrete(c)
		}(&cobranca)
	}

	wg.Wait()
	log.Println("✅ Notificações de lembrete enviadas")
}

// EnviarNotificacoesVencimento envia notificações de vencimento (dia do vencimento)
func (s *AgendadorServico) EnviarNotificacoesVencimento() {
	cobrancas, err := s.cobrancaRepo.BuscarCobrancasVencendoHoje()
	if err != nil {
		log.Printf("❌ Erro ao buscar cobranças vencendo hoje: %v", err)
		return
	}

	if len(cobrancas) == 0 {
		log.Println("📭 Nenhuma cobrança vencendo hoje")
		return
	}

	log.Printf("📬 Enviando %d notificações de vencimento...", len(cobrancas))

	// Usar goroutines para enviar notificações em paralelo
	var wg sync.WaitGroup
	for _, cobranca := range cobrancas {
		wg.Add(1)
		go func(c *entidades.Cobranca) {
			defer wg.Done()
			s.enviarNotificacaoVencimento(c)
		}(&cobranca)
	}

	wg.Wait()
	log.Println("✅ Notificações de vencimento enviadas")
}

// enviarNotificacaoLembrete envia notificação de lembrete para uma cobrança
func (s *AgendadorServico) enviarNotificacaoLembrete(cobranca *entidades.Cobranca) {
	// Enviar WhatsApp
	conexao, err := s.whatsappRepo.BuscarPorUsuario(cobranca.UsuarioID)
	if err == nil && conexao.Conectado {
		mensagem := fmt.Sprintf(
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

		_, err := s.evolutionAPI.EnviarMensagemTexto(
			conexao.NomeInstancia,
			cobranca.Cliente.Telefone,
			mensagem,
		)
		if err != nil {
			log.Printf("❌ Erro ao enviar WhatsApp para %s: %v", cobranca.Cliente.Nome, err)
		} else {
			log.Printf("✅ WhatsApp enviado para %s", cobranca.Cliente.Nome)
		}
	}

	// Enviar Email
	err = s.resendAPI.EnviarEmailLembrete(
		cobranca.Cliente.Email,
		cobranca.Cliente.Nome,
		cobranca.Descricao,
		cobranca.Valor,
		cobranca.DataVencimento.Format("02/01/2006"),
	)
	if err != nil {
		log.Printf("❌ Erro ao enviar email para %s: %v", cobranca.Cliente.Nome, err)
	} else {
		log.Printf("✅ Email enviado para %s", cobranca.Cliente.Nome)
	}

	// Marcar notificação como enviada
	cobranca.NotificacaoLembreteEnviada = true
	s.cobrancaRepo.Atualizar(cobranca)
}

// enviarNotificacaoVencimento envia notificação de vencimento para uma cobrança
func (s *AgendadorServico) enviarNotificacaoVencimento(cobranca *entidades.Cobranca) {
	// Enviar WhatsApp
	conexao, err := s.whatsappRepo.BuscarPorUsuario(cobranca.UsuarioID)
	if err == nil && conexao.Conectado {
		mensagem := fmt.Sprintf(
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

		_, err := s.evolutionAPI.EnviarMensagemTexto(
			conexao.NomeInstancia,
			cobranca.Cliente.Telefone,
			mensagem,
		)
		if err != nil {
			log.Printf("❌ Erro ao enviar WhatsApp para %s: %v", cobranca.Cliente.Nome, err)
		} else {
			log.Printf("✅ WhatsApp enviado para %s", cobranca.Cliente.Nome)
		}
	}

	// Enviar Email
	err = s.resendAPI.EnviarEmailVencimento(
		cobranca.Cliente.Email,
		cobranca.Cliente.Nome,
		cobranca.Descricao,
		cobranca.Valor,
	)
	if err != nil {
		log.Printf("❌ Erro ao enviar email para %s: %v", cobranca.Cliente.Nome, err)
	} else {
		log.Printf("✅ Email enviado para %s", cobranca.Cliente.Nome)
	}

	// Marcar notificação como enviada
	cobranca.NotificacaoVencimentoEnviada = true
	s.cobrancaRepo.Atualizar(cobranca)
}

// AtualizarCobrancasVencidas atualiza o status de cobranças vencidas
func (s *AgendadorServico) AtualizarCobrancasVencidas() {
	cobrancas, err := s.cobrancaRepo.BuscarCobrancasVencidas()
	if err != nil {
		log.Printf("❌ Erro ao buscar cobranças vencidas: %v", err)
		return
	}

	if len(cobrancas) == 0 {
		log.Println("📭 Nenhuma cobrança vencida para atualizar")
		return
	}

	log.Printf("🔄 Atualizando %d cobranças vencidas...", len(cobrancas))

	for _, cobranca := range cobrancas {
		cobranca.Status = enums.StatusCobrancaVencida
		err := s.cobrancaRepo.Atualizar(&cobranca)
		if err != nil {
			log.Printf("❌ Erro ao atualizar cobrança %d: %v", cobranca.ID, err)
		}
	}

	log.Println("✅ Cobranças vencidas atualizadas")
}
