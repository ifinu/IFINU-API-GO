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
	"github.com/ifinu/ifinu-api-go/util"
	"github.com/robfig/cron/v3"
)

type AgendadorServico struct {
	cobrancaRepo     *repositorio.CobrancaRepositorio
	whatsappRepo     *repositorio.WhatsAppRepositorio
	evolutionAPI     *integracao.EvolutionAPICliente
	resendAPI        *integracao.ResendCliente
	cron             *cron.Cron
	horarioComercial *util.HorarioComercial
	filaMensagem     *FilaMensagemServico
}

func NovoAgendadorServico(
	cobrancaRepo *repositorio.CobrancaRepositorio,
	whatsappRepo *repositorio.WhatsAppRepositorio,
	evolutionAPI *integracao.EvolutionAPICliente,
	resendAPI *integracao.ResendCliente,
	whatsappServico *WhatsAppServico,
	redisAddr string,
) *AgendadorServico {
	// Inicializar fila de mensagens
	filaMensagem := NovoFilaMensagemServico(redisAddr, whatsappServico, resendAPI)

	// Iniciar worker pool (10 workers processando em paralelo)
	if filaMensagem != nil {
		filaMensagem.IniciarWorkerPool(10)
	}

	return &AgendadorServico{
		cobrancaRepo:     cobrancaRepo,
		whatsappRepo:     whatsappRepo,
		evolutionAPI:     evolutionAPI,
		resendAPI:        resendAPI,
		cron:             cron.New(),
		horarioComercial: util.HorarioComercialPadrao(),
		filaMensagem:     filaMensagem,
	}
}

// Iniciar inicia o agendador de tarefas
func (s *AgendadorServico) Iniciar() {
	log.Println("📅 Iniciando agendador de notificações...")
	log.Printf("⏰ Horário comercial configurado: %dh às %dh (dias úteis)",
		s.horarioComercial.HoraInicio, s.horarioComercial.HoraFim)

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

	// Processar notificações pendentes - executa a cada hora durante horário comercial
	s.cron.AddFunc("0 * * * *", func() {
		agora := time.Now()
		if s.horarioComercial.EstaDentroHorarioComercial(agora) {
			log.Println("⏰ Executando job: Processar notificações pendentes (horário comercial)")
			s.ProcessarNotificacoesPendentes()
		}
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
	agora := time.Now()

	// Verificar se está dentro do horário comercial
	if !s.horarioComercial.EstaDentroHorarioComercial(agora) {
		proximoHorario := s.horarioComercial.FormatarProximoHorario(agora)
		log.Printf("⏸️  Fora do horário comercial. Próxima tentativa: %s", proximoHorario)
		return
	}

	cobrancas, err := s.cobrancaRepo.BuscarCobrancasParaLembrete()
	if err != nil {
		log.Printf("❌ Erro ao buscar cobranças para lembrete: %v", err)
		return
	}

	if len(cobrancas) == 0 {
		log.Println("📭 Nenhuma cobrança para enviar lembrete")
		return
	}

	log.Printf("📬 Enfileirando %d notificações de lembrete...", len(cobrancas))

	// Enfileirar todas as mensagens para processamento assíncrono
	enfileiradas := 0
	for _, cobranca := range cobrancas {
		// Se fila não disponível, enviar direto (fallback)
		if s.filaMensagem == nil {
			s.enviarNotificacaoLembrete(&cobranca)
			continue
		}

		// Enfileirar mensagem
		msg := &MensagemFila{
			ID:              fmt.Sprintf("lembrete_%d_%d", cobranca.ID, time.Now().Unix()),
			TipoNotificacao: "lembrete",
			Cobranca:        &cobranca,
			Tentativas:      0,
		}

		if err := s.filaMensagem.EnfileirarMensagem(msg); err != nil {
			log.Printf("❌ Erro ao enfileirar: %v. Enviando direto...", err)
			s.enviarNotificacaoLembrete(&cobranca)
		} else {
			enfileiradas++
		}

		// Marcar como processada (não enfileirar novamente)
		cobranca.NotificacaoLembreteEnviada = true
		s.cobrancaRepo.Atualizar(&cobranca)
	}

	if enfileiradas > 0 {
		log.Printf("✅ %d notificações de lembrete enfileiradas para processamento", enfileiradas)
	}
}

// EnviarNotificacoesVencimento envia notificações de vencimento (dia do vencimento)
func (s *AgendadorServico) EnviarNotificacoesVencimento() {
	agora := time.Now()

	// Verificar se está dentro do horário comercial
	if !s.horarioComercial.EstaDentroHorarioComercial(agora) {
		proximoHorario := s.horarioComercial.FormatarProximoHorario(agora)
		log.Printf("⏸️  Fora do horário comercial. Próxima tentativa: %s", proximoHorario)
		return
	}

	cobrancas, err := s.cobrancaRepo.BuscarCobrancasVencendoHoje()
	if err != nil {
		log.Printf("❌ Erro ao buscar cobranças vencendo hoje: %v", err)
		return
	}

	if len(cobrancas) == 0 {
		log.Println("📭 Nenhuma cobrança vencendo hoje")
		return
	}

	log.Printf("📬 Enfileirando %d notificações de vencimento...", len(cobrancas))

	// Enfileirar todas as mensagens para processamento assíncrono
	enfileiradas := 0
	for _, cobranca := range cobrancas {
		// Se fila não disponível, enviar direto (fallback)
		if s.filaMensagem == nil {
			s.enviarNotificacaoVencimento(&cobranca)
			continue
		}

		// Enfileirar mensagem
		msg := &MensagemFila{
			ID:              fmt.Sprintf("vencimento_%d_%d", cobranca.ID, time.Now().Unix()),
			TipoNotificacao: "vencimento",
			Cobranca:        &cobranca,
			Tentativas:      0,
		}

		if err := s.filaMensagem.EnfileirarMensagem(msg); err != nil {
			log.Printf("❌ Erro ao enfileirar: %v. Enviando direto...", err)
			s.enviarNotificacaoVencimento(&cobranca)
		} else {
			enfileiradas++
		}

		// Marcar como processada (não enfileirar novamente)
		cobranca.NotificacaoVencimentoEnviada = true
		s.cobrancaRepo.Atualizar(&cobranca)
	}

	if enfileiradas > 0 {
		log.Printf("✅ %d notificações de vencimento enfileiradas para processamento", enfileiradas)
	}
}

// ProcessarNotificacoesPendentes processa notificações que ficaram pendentes fora do horário comercial
func (s *AgendadorServico) ProcessarNotificacoesPendentes() {
	// Processar lembretes pendentes
	s.EnviarNotificacoesLembrete()

	// Processar vencimentos pendentes
	s.EnviarNotificacoesVencimento()
}

// enviarNotificacaoLembrete envia notificação de lembrete para uma cobrança
func (s *AgendadorServico) enviarNotificacaoLembrete(cobranca *entidades.Cobranca) {
	// Enviar WhatsApp
	conexao, err := s.whatsappRepo.BuscarPorUsuario(cobranca.UsuarioID)
	if err == nil && conexao.IsConectado() {
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
			conexao.InstanceName,
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
	if err == nil && conexao.IsConectado() {
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
			conexao.InstanceName,
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
		cobranca.Status = enums.StatusCobrancaVencido
		err := s.cobrancaRepo.Atualizar(&cobranca)
		if err != nil {
			log.Printf("❌ Erro ao atualizar cobrança %d: %v", cobranca.ID, err)
		}
	}

	log.Println("✅ Cobranças vencidas atualizadas")
}
