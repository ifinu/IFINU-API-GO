package servico

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dominio/entidades"
	"github.com/ifinu/ifinu-api-go/dominio/enums"
	"github.com/ifinu/ifinu-api-go/dto"
	"github.com/ifinu/ifinu-api-go/repositorio"
	"gorm.io/gorm"
)

type CobrancaServico struct {
	cobrancaRepo *repositorio.CobrancaRepositorio
	clienteRepo  *repositorio.ClienteRepositorio
}

func NovoCobrancaServico(cobrancaRepo *repositorio.CobrancaRepositorio, clienteRepo *repositorio.ClienteRepositorio) *CobrancaServico {
	return &CobrancaServico{
		cobrancaRepo: cobrancaRepo,
		clienteRepo:  clienteRepo,
	}
}

// Criar cria uma nova cobrança
func (s *CobrancaServico) Criar(usuarioID uuid.UUID, req dto.CobrancaRequest) (*dto.CobrancaResponse, error) {
	// Verificar se cliente existe e pertence ao usuário
	cliente, err := s.clienteRepo.BuscarPorID(req.ClienteID, usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("cliente não encontrado")
		}
		return nil, err
	}

	// Criar cobrança
	cobranca := &entidades.Cobranca{
		UsuarioID:       usuarioID,
		ClienteID:       req.ClienteID,
		Valor:           req.Valor,
		Descricao:       req.Descricao,
		Status:          enums.StatusCobrancaPendente,
		DataVencimento:  req.DataVencimento,
		TipoRecorrencia: req.TipoRecorrencia,
		DataCriacao:     time.Now(),
	}

	err = s.cobrancaRepo.Criar(cobranca)
	if err != nil {
		return nil, err
	}

	// Carregar cliente para resposta
	cobranca.Cliente = *cliente

	// TODO: Enviar notificações se solicitado
	// if req.EnviarWhatsApp {
	// 	go enviarNotificacaoWhatsApp(cobranca)
	// }
	// if req.EnviarEmail {
	// 	go enviarNotificacaoEmail(cobranca)
	// }

	return s.mapearParaDTO(cobranca), nil
}

// BuscarPorID busca uma cobrança por ID
func (s *CobrancaServico) BuscarPorID(usuarioID uuid.UUID, cobrancaID int64) (*dto.CobrancaResponse, error) {
	cobranca, err := s.cobrancaRepo.BuscarPorID(cobrancaID, usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("cobrança não encontrada")
		}
		return nil, err
	}

	return s.mapearParaDTO(cobranca), nil
}

// Listar lista todas as cobranças do usuário
func (s *CobrancaServico) Listar(usuarioID uuid.UUID) ([]dto.CobrancaResponse, error) {
	cobrancas, err := s.cobrancaRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		return nil, err
	}

	cobrancasDTO := make([]dto.CobrancaResponse, len(cobrancas))
	for i, cobranca := range cobrancas {
		cobrancasDTO[i] = *s.mapearParaDTO(&cobranca)
	}

	return cobrancasDTO, nil
}

// Buscar busca cobranças com filtros e paginação
func (s *CobrancaServico) Buscar(usuarioID uuid.UUID, req dto.BuscarCobrancasRequest) (*dto.CobrancaListResponse, error) {
	// Valores padrão
	if req.Pagina == 0 {
		req.Pagina = 1
	}
	if req.TamanhoPagina == 0 {
		req.TamanhoPagina = 20
	}

	cobrancas, total, err := s.cobrancaRepo.Buscar(
		usuarioID,
		req.Status,
		req.DataInicio,
		req.DataFim,
		req.Pagina,
		req.TamanhoPagina,
	)
	if err != nil {
		return nil, err
	}

	cobrancasDTO := make([]dto.CobrancaResponse, len(cobrancas))
	for i, cobranca := range cobrancas {
		cobrancasDTO[i] = *s.mapearParaDTO(&cobranca)
	}

	totalPaginas := int(math.Ceil(float64(total) / float64(req.TamanhoPagina)))

	return &dto.CobrancaListResponse{
		Cobrancas:     cobrancasDTO,
		Total:         total,
		Pagina:        req.Pagina,
		TamanhoPagina: req.TamanhoPagina,
		TotalPaginas:  totalPaginas,
	}, nil
}

// Atualizar atualiza uma cobrança
func (s *CobrancaServico) Atualizar(usuarioID uuid.UUID, cobrancaID int64, req dto.CobrancaRequest) (*dto.CobrancaResponse, error) {
	// Buscar cobrança
	cobranca, err := s.cobrancaRepo.BuscarPorID(cobrancaID, usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("cobrança não encontrada")
		}
		return nil, err
	}

	// Verificar se cliente existe e pertence ao usuário
	_, err = s.clienteRepo.BuscarPorID(req.ClienteID, usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("cliente não encontrado")
		}
		return nil, err
	}

	// Atualizar dados
	cobranca.ClienteID = req.ClienteID
	cobranca.Valor = req.Valor
	cobranca.Descricao = req.Descricao
	cobranca.DataVencimento = req.DataVencimento
	cobranca.TipoRecorrencia = req.TipoRecorrencia

	err = s.cobrancaRepo.Atualizar(cobranca)
	if err != nil {
		return nil, err
	}

	return s.mapearParaDTO(cobranca), nil
}

// AtualizarStatus atualiza o status de uma cobrança
func (s *CobrancaServico) AtualizarStatus(usuarioID uuid.UUID, cobrancaID int64, novoStatus enums.StatusCobranca) (*dto.CobrancaResponse, error) {
	// Buscar cobrança
	cobranca, err := s.cobrancaRepo.BuscarPorID(cobrancaID, usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("cobrança não encontrada")
		}
		return nil, err
	}

	// Atualizar status
	cobranca.Status = novoStatus

	// Se marcada como paga, registrar data de pagamento
	if novoStatus == enums.StatusCobrancaPaga {
		agora := time.Now()
		cobranca.DataPagamento = &agora
	}

	err = s.cobrancaRepo.Atualizar(cobranca)
	if err != nil {
		return nil, err
	}

	return s.mapearParaDTO(cobranca), nil
}

// Deletar remove uma cobrança
func (s *CobrancaServico) Deletar(usuarioID uuid.UUID, cobrancaID int64) error {
	// Verificar se cobrança existe
	_, err := s.cobrancaRepo.BuscarPorID(cobrancaID, usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("cobrança não encontrada")
		}
		return err
	}

	return s.cobrancaRepo.Deletar(cobrancaID, usuarioID)
}

// ObterEstatisticas retorna estatísticas de cobranças do usuário
func (s *CobrancaServico) ObterEstatisticas(usuarioID uuid.UUID) (*dto.EstatisticasCobrancasResponse, error) {
	// Contar por status
	contagem, err := s.cobrancaRepo.ContarPorStatus(usuarioID)
	if err != nil {
		return nil, err
	}

	// Calcular valores
	valorPendente, _ := s.cobrancaRepo.CalcularValorTotalPorStatus(usuarioID, enums.StatusCobrancaPendente)
	valorPago, _ := s.cobrancaRepo.CalcularValorTotalPorStatus(usuarioID, enums.StatusCobrancaPaga)
	valorVencido, _ := s.cobrancaRepo.CalcularValorTotalPorStatus(usuarioID, enums.StatusCobrancaVencida)

	return &dto.EstatisticasCobrancasResponse{
		TotalPendente:  contagem[enums.StatusCobrancaPendente],
		TotalPaga:      contagem[enums.StatusCobrancaPaga],
		TotalVencida:   contagem[enums.StatusCobrancaVencida],
		TotalCancelada: contagem[enums.StatusCobrancaCancelada],
		ValorPendente:  valorPendente,
		ValorPago:      valorPago,
		ValorVencido:   valorVencido,
	}, nil
}

// mapearParaDTO converte Cobranca para CobrancaResponse
func (s *CobrancaServico) mapearParaDTO(cobranca *entidades.Cobranca) *dto.CobrancaResponse {
	response := &dto.CobrancaResponse{
		ID:                           cobranca.ID,
		ClienteID:                    cobranca.ClienteID,
		Valor:                        cobranca.Valor,
		Descricao:                    cobranca.Descricao,
		Status:                       cobranca.Status,
		DataVencimento:               cobranca.DataVencimento,
		DataPagamento:                cobranca.DataPagamento,
		TipoRecorrencia:              cobranca.TipoRecorrencia,
		LinkPagamento:                cobranca.LinkPagamento,
		NotificacaoLembreteEnviada:   cobranca.NotificacaoLembreteEnviada,
		NotificacaoVencimentoEnviada: cobranca.NotificacaoVencimentoEnviada,
		DataCriacao:                  cobranca.DataCriacao,
	}

	// Adicionar cliente se carregado
	if cobranca.Cliente.ID != 0 {
		response.Cliente = &dto.ClienteResponse{
			ID:          cobranca.Cliente.ID,
			Nome:        cobranca.Cliente.Nome,
			Email:       cobranca.Cliente.Email,
			Telefone:    cobranca.Cliente.Telefone,
			Endereco:    cobranca.Cliente.Endereco,
			CPF:         cobranca.Cliente.CPF,
			CNPJ:        cobranca.Cliente.CNPJ,
			DataCriacao: cobranca.Cliente.DataCriacao,
		}
	}

	return response
}
