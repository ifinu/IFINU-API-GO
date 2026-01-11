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
	if err == nil && conexaoExistente.Conectado {
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
		conexaoExistente.NomeInstancia = nomeInstancia
		conexaoExistente.QRCode = resultado.Qrcode.Base64
		conexaoExistente.Conectado = false
		err = s.whatsappRepo.Atualizar(conexaoExistente)
	} else {
		// Criar nova conexão
		conexao := &entidades.WhatsAppConexao{
			UsuarioID:     usuarioID,
			NomeInstancia: nomeInstancia,
			QRCode:        resultado.Qrcode.Base64,
			Conectado:     false,
			DataCriacao:   time.Now(),
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
	status, err := s.evolutionAPI.ObterStatus(conexao.NomeInstancia)
	if err != nil {
		// Se der erro, retornar status da base de dados
		return &dto.StatusWhatsAppResponse{
			Conectado:     conexao.Conectado,
			NomeInstancia: conexao.NomeInstancia,
			DataConexao:   conexao.DataConexao,
		}, nil
	}

	// Atualizar status na base de dados se mudou
	statusConectado := status.Instance.State == "open"
	if statusConectado != conexao.Conectado {
		conexao.Conectado = statusConectado
		if statusConectado {
			agora := time.Now()
			conexao.DataConexao = &agora
		}
		s.whatsappRepo.Atualizar(conexao)
	}

	return &dto.StatusWhatsAppResponse{
		Conectado:     statusConectado,
		NomeInstancia: conexao.NomeInstancia,
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
	err = s.evolutionAPI.Desconectar(conexao.NomeInstancia)
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

	if !conexao.Conectado {
		return nil, errors.New("WhatsApp não está conectado")
	}

	// Enviar mensagem via Evolution API
	resultado, err := s.evolutionAPI.EnviarMensagemTexto(conexao.NomeInstancia, telefone, mensagem)
	if err != nil {
		return &dto.EnviarMensagemResponse{
			Sucesso:  false,
			Mensagem: fmt.Sprintf("Erro ao enviar mensagem: %v", err),
		}, nil
	}

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
	status, err := s.evolutionAPI.ObterStatus(conexao.NomeInstancia)
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
