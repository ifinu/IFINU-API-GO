package servico

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dominio/entidades"
	"github.com/ifinu/ifinu-api-go/dto"
	"github.com/ifinu/ifinu-api-go/integracao"
	"github.com/ifinu/ifinu-api-go/repositorio"
	"github.com/ifinu/ifinu-api-go/util"
	"gorm.io/gorm"
)

type WhatsAppServico struct {
	whatsappRepo *repositorio.WhatsAppRepositorio
	usuarioRepo  *repositorio.UsuarioRepositorio
	evolutionAPI *integracao.EvolutionAPICliente
}

func NovoWhatsAppServico(
	whatsappRepo *repositorio.WhatsAppRepositorio,
	usuarioRepo *repositorio.UsuarioRepositorio,
	evolutionAPI *integracao.EvolutionAPICliente,
) *WhatsAppServico {
	return &WhatsAppServico{
		whatsappRepo: whatsappRepo,
		usuarioRepo:  usuarioRepo,
		evolutionAPI: evolutionAPI,
	}
}

// Conectar inicia o processo de conexão do WhatsApp
func (s *WhatsAppServico) Conectar(usuarioID uuid.UUID) (*dto.ConectarWhatsAppResponse, error) {
	// Buscar usuário
	usuario, err := s.usuarioRepo.BuscarPorID(usuarioID)
	if err != nil {
		return nil, err
	}

	// Verificar se já existe uma conexão
	conexaoExistente, err := s.whatsappRepo.BuscarPorUsuario(usuarioID)
	if err == nil && conexaoExistente.IsConectado() {
		return nil, errors.New("já existe uma conexão WhatsApp ativa")
	}

	// Gerar nome de instância único
	nomeInstancia := fmt.Sprintf("ifinu_%s_%d", usuario.Email, time.Now().Unix())

	// Criar instância no Evolution API
	resultado, err := s.evolutionAPI.CriarInstancia(nomeInstancia)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar instância: %w", err)
	}

	// Salvar ou atualizar conexão
	if conexaoExistente != nil {
		// Atualizar conexão existente
		conexaoExistente.InstanceName = nomeInstancia
		conexaoExistente.QRCode = resultado.Qrcode.Base64
		conexaoExistente.Status = entidades.StatusConexaoConectando
		err = s.whatsappRepo.Atualizar(conexaoExistente)
	} else {
		// Criar nova conexão
		conexao := &entidades.WhatsAppConexao{
			UsuarioID:    usuarioID,
			InstanceName: nomeInstancia,
			QRCode:       resultado.Qrcode.Base64,
			Status:       entidades.StatusConexaoConectando,
			DataCriacao:  time.Now(),
		}
		err = s.whatsappRepo.Criar(conexao)
	}

	if err != nil {
		return nil, err
	}

	return &dto.ConectarWhatsAppResponse{
		QRCode:        resultado.Qrcode.Base64,
		NomeInstancia: nomeInstancia,
	}, nil
}

// ObterStatus retorna o status da conexão WhatsApp
func (s *WhatsAppServico) ObterStatus(usuarioID uuid.UUID) (*dto.StatusWhatsAppResponse, error) {
	// Buscar conexão
	conexao, err := s.whatsappRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &dto.StatusWhatsAppResponse{
				Conectado: false,
			}, nil
		}
		return nil, err
	}

	// Verificar status no Evolution API
	status, err := s.evolutionAPI.ObterStatus(conexao.InstanceName)
	if err != nil {
		// Se der erro, retornar status da base de dados
		return &dto.StatusWhatsAppResponse{
			Conectado:     conexao.IsConectado(),
			NomeInstancia: conexao.InstanceName,
			DataConexao:   conexao.DataConexao,
		}, nil
	}

	// Atualizar status na base de dados se mudou
	statusConectado := status.Instance.State == "open"
	if statusConectado != conexao.IsConectado() {
		if statusConectado {
			conexao.Conectar("")
		} else {
			conexao.Desconectar()
		}
		s.whatsappRepo.Atualizar(conexao)
	}

	return &dto.StatusWhatsAppResponse{
		Conectado:     statusConectado,
		NomeInstancia: conexao.InstanceName,
		QRCode:        conexao.QRCode,
		DataConexao:   conexao.DataConexao,
	}, nil
}

// Desconectar desconecta o WhatsApp
func (s *WhatsAppServico) Desconectar(usuarioID uuid.UUID) error {
	// Buscar conexão
	conexao, err := s.whatsappRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("nenhuma conexão WhatsApp encontrada")
		}
		return err
	}

	// Desconectar no Evolution API
	err = s.evolutionAPI.Desconectar(conexao.InstanceName)
	if err != nil {
		return err
	}

	// Deletar conexão da base de dados
	return s.whatsappRepo.Deletar(usuarioID)
}

// EnviarMensagem envia uma mensagem via WhatsApp
func (s *WhatsAppServico) EnviarMensagem(usuarioID uuid.UUID, telefone, mensagem string) (*dto.EnviarMensagemResponse, error) {
	// Buscar conexão
	conexao, err := s.whatsappRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("WhatsApp não conectado")
		}
		return nil, err
	}

	if !conexao.IsConectado() {
		return nil, errors.New("WhatsApp não está conectado")
	}

	// Formatar telefone brasileiro (adicionar código do país 55 se necessário)
	telefoneFormatado := util.FormatarTelefoneBrasileiro(telefone)

	// Enviar mensagem via Evolution API
	resultado, err := s.evolutionAPI.EnviarMensagemTexto(conexao.InstanceName, telefoneFormatado, mensagem)
	if err != nil {
		// Incrementar contador de falhas
		conexao.MensagensEnviadas++
		conexao.MensagensFalha++
		s.whatsappRepo.Atualizar(conexao)

		return &dto.EnviarMensagemResponse{
			Sucesso:  false,
			Mensagem: fmt.Sprintf("Erro ao enviar mensagem: %v", err),
		}, nil
	}

	// Incrementar contadores de sucesso
	conexao.MensagensEnviadas++
	conexao.MensagensSucesso++
	now := time.Now()
	conexao.DataUltimaAtividade = &now
	s.whatsappRepo.Atualizar(conexao)

	return &dto.EnviarMensagemResponse{
		Sucesso:   true,
		Mensagem:  "Mensagem enviada com sucesso",
		MessageID: resultado.Key.ID,
	}, nil
}

// TestarConexao testa a conexão WhatsApp
func (s *WhatsAppServico) TestarConexao(usuarioID uuid.UUID) (*dto.TestarConexaoResponse, error) {
	// Buscar conexão
	conexao, err := s.whatsappRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &dto.TestarConexaoResponse{
				Sucesso:  false,
				Mensagem: "WhatsApp não conectado",
				Status:   "disconnected",
			}, nil
		}
		return nil, err
	}

	// Verificar status no Evolution API
	status, err := s.evolutionAPI.ObterStatus(conexao.InstanceName)
	if err != nil {
		return &dto.TestarConexaoResponse{
			Sucesso:  false,
			Mensagem: "Erro ao verificar status",
			Status:   "error",
		}, nil
	}

	if status.Instance.State == "open" {
		return &dto.TestarConexaoResponse{
			Sucesso:  true,
			Mensagem: "WhatsApp conectado",
			Status:   "connected",
		}, nil
	}

	return &dto.TestarConexaoResponse{
		Sucesso:  false,
		Mensagem: "WhatsApp não conectado",
		Status:   status.Instance.State,
	}, nil
}

// ObterQRCode retorna o QR code da conexão WhatsApp
func (s *WhatsAppServico) ObterQRCode(usuarioID uuid.UUID) (map[string]interface{}, error) {
	// Buscar conexão
	conexao, err := s.whatsappRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("nenhuma conexão WhatsApp encontrada")
		}
		return nil, err
	}

	// Se não está conectado e tem QR code, retornar
	if !conexao.IsConectado() && conexao.QRCode != "" {
		return map[string]interface{}{
			"qrcode":        conexao.QRCode,
			"nomeInstancia": conexao.InstanceName,
			"status":        conexao.Status,
		}, nil
	}

	// Se já está conectado
	if conexao.IsConectado() {
		return map[string]interface{}{
			"conectado":     true,
			"nomeInstancia": conexao.InstanceName,
			"dataConexao":   conexao.DataConexao,
		}, nil
	}

	return nil, errors.New("QR Code não disponível")
}

// LimparOrfaos remove conexões WhatsApp órfãs (sem instância válida na Evolution API)
func (s *WhatsAppServico) LimparOrfaos(usuarioID uuid.UUID) (map[string]interface{}, error) {
	// Buscar conexão do usuário
	conexao, err := s.whatsappRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return map[string]interface{}{
				"mensagem": "Nenhuma conexão encontrada",
				"removido": false,
			}, nil
		}
		return nil, err
	}

	// Verificar se a instância existe na Evolution API
	_, err = s.evolutionAPI.ObterStatus(conexao.InstanceName)
	if err != nil {
		// Se der erro ao buscar status, a instância provavelmente não existe mais
		// Remover conexão órfã
		err = s.whatsappRepo.Deletar(usuarioID)
		if err != nil {
			return nil, fmt.Errorf("erro ao remover conexão órfã: %w", err)
		}

		return map[string]interface{}{
			"mensagem": "Conexão órfã removida com sucesso",
			"removido": true,
		}, nil
	}

	return map[string]interface{}{
		"mensagem": "Conexão válida, nada a remover",
		"removido": false,
	}, nil
}

// ObterEstatisticas retorna estatísticas sobre o WhatsApp
func (s *WhatsAppServico) ObterEstatisticas(usuarioID uuid.UUID) (map[string]interface{}, error) {
	// Buscar conexão do usuário
	conexao, err := s.whatsappRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return map[string]interface{}{
				"conectado":      false,
				"totalConexoes":  0,
				"ultimaConexao":  nil,
			}, nil
		}
		return nil, err
	}

	estatisticas := map[string]interface{}{
		"conectado":       conexao.IsConectado(),
		"nomeInstancia":   conexao.InstanceName,
		"status":          conexao.Status,
		"dataCriacao":     conexao.DataCriacao,
		"dataConexao":     conexao.DataConexao,
		"dataAtualizacao": conexao.DataAtualizacao,
	}

	// Se conectado, buscar informações adicionais da Evolution API
	if conexao.IsConectado() {
		status, err := s.evolutionAPI.ObterStatus(conexao.InstanceName)
		if err == nil {
			estatisticas["instanceName"] = status.Instance.InstanceName
			estatisticas["state"] = status.Instance.State
		}
	}

	return estatisticas, nil
}
