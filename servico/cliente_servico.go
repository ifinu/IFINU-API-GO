package servico

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dominio/entidades"
	"github.com/ifinu/ifinu-api-go/dto"
	"github.com/ifinu/ifinu-api-go/repositorio"
	"gorm.io/gorm"
)

type ClienteServico struct {
	clienteRepo *repositorio.ClienteRepositorio
}

func NovoClienteServico(clienteRepo *repositorio.ClienteRepositorio) *ClienteServico {
	return &ClienteServico{
		clienteRepo: clienteRepo,
	}
}

// Criar cria um novo cliente
func (s *ClienteServico) Criar(usuarioID uuid.UUID, req dto.ClienteRequest) (*dto.ClienteResponse, error) {
	// Verificar se email já existe para este usuário
	existe, err := s.clienteRepo.ExistePorEmail(req.Email, usuarioID)
	if err != nil {
		return nil, err
	}
	if existe {
		return nil, errors.New("já existe um cliente com este email")
	}

	// Criar cliente
	cliente := &entidades.Cliente{
		UsuarioID:   usuarioID,
		Nome:        req.Nome,
		Email:       req.Email,
		Telefone:    req.Telefone,
		Endereco:    req.Endereco,
		CPF:         req.CPF,
		CNPJ:        req.CNPJ,
		DataCriacao: time.Now(),
	}

	err = s.clienteRepo.Criar(cliente)
	if err != nil {
		return nil, err
	}

	return s.mapearParaDTO(cliente), nil
}

// BuscarPorID busca um cliente por ID
func (s *ClienteServico) BuscarPorID(usuarioID uuid.UUID, clienteID uuid.UUID) (*dto.ClienteResponse, error) {
	cliente, err := s.clienteRepo.BuscarPorID(clienteID, usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("cliente não encontrado")
		}
		return nil, err
	}

	return s.mapearParaDTO(cliente), nil
}

// Listar lista todos os clientes do usuário
func (s *ClienteServico) Listar(usuarioID uuid.UUID) ([]dto.ClienteResponse, error) {
	clientes, err := s.clienteRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		return nil, err
	}

	clientesDTO := make([]dto.ClienteResponse, len(clientes))
	for i, cliente := range clientes {
		clientesDTO[i] = *s.mapearParaDTO(&cliente)
	}

	return clientesDTO, nil
}

// Buscar busca clientes com filtros e paginação
func (s *ClienteServico) Buscar(usuarioID uuid.UUID, req dto.BuscarClientesRequest) (*dto.ClienteListResponse, error) {
	// Valores padrão
	if req.Pagina == 0 {
		req.Pagina = 1
	}
	if req.TamanhoPagina == 0 {
		req.TamanhoPagina = 20
	}

	clientes, total, err := s.clienteRepo.Buscar(usuarioID, req.Termo, req.Pagina, req.TamanhoPagina)
	if err != nil {
		return nil, err
	}

	clientesDTO := make([]dto.ClienteResponse, len(clientes))
	for i, cliente := range clientes {
		clientesDTO[i] = *s.mapearParaDTO(&cliente)
	}

	totalPaginas := int(math.Ceil(float64(total) / float64(req.TamanhoPagina)))

	return &dto.ClienteListResponse{
		Clientes:      clientesDTO,
		Total:         total,
		Pagina:        req.Pagina,
		TamanhoPagina: req.TamanhoPagina,
		TotalPaginas:  totalPaginas,
	}, nil
}

// Atualizar atualiza um cliente
func (s *ClienteServico) Atualizar(usuarioID uuid.UUID, clienteID uuid.UUID, req dto.ClienteRequest) (*dto.ClienteResponse, error) {
	// Buscar cliente
	cliente, err := s.clienteRepo.BuscarPorID(clienteID, usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("cliente não encontrado")
		}
		return nil, err
	}

	// Verificar se email já existe para outro cliente
	clienteExistente, err := s.clienteRepo.BuscarPorEmail(req.Email, usuarioID)
	if err == nil && clienteExistente.ID != clienteID {
		return nil, errors.New("já existe outro cliente com este email")
	}

	// Atualizar dados
	cliente.Nome = req.Nome
	cliente.Email = req.Email
	cliente.Telefone = req.Telefone
	cliente.Endereco = req.Endereco
	cliente.CPF = req.CPF
	cliente.CNPJ = req.CNPJ

	err = s.clienteRepo.Atualizar(cliente)
	if err != nil {
		return nil, err
	}

	return s.mapearParaDTO(cliente), nil
}

// Deletar remove um cliente
func (s *ClienteServico) Deletar(usuarioID uuid.UUID, clienteID uuid.UUID) error {
	// Verificar se cliente existe
	_, err := s.clienteRepo.BuscarPorID(clienteID, usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("cliente não encontrado")
		}
		return err
	}

	return s.clienteRepo.Deletar(clienteID, usuarioID)
}

// mapearParaDTO converte Cliente para ClienteResponse
func (s *ClienteServico) mapearParaDTO(cliente *entidades.Cliente) *dto.ClienteResponse {
	return &dto.ClienteResponse{
		ID:          cliente.ID,
		Nome:        cliente.Nome,
		Email:       cliente.Email,
		Telefone:    cliente.Telefone,
		Endereco:    cliente.Endereco,
		CPF:         cliente.CPF,
		CNPJ:        cliente.CNPJ,
		DataCriacao: cliente.DataCriacao,
	}
}
