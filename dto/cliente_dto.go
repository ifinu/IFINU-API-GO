package dto

import (
	"time"

	"github.com/google/uuid"
)

// ClienteRequest representa a requisição de criação/atualização de cliente
type ClienteRequest struct {
	Nome     string  `json:"nome" binding:"required,min=2"`
	Email    string  `json:"email" binding:"required,email"`
	Telefone string  `json:"telefone" binding:"required"`
	Endereco string  `json:"endereco"`
	CPF      *string `json:"cpf"`
	CNPJ     *string `json:"cnpj"`
}

// ClienteResponse representa o cliente na resposta
type ClienteResponse struct {
	ID          uuid.UUID `json:"id"`
	Nome        string    `json:"nome"`
	Email       string    `json:"email"`
	Telefone    string    `json:"telefone"`
	Endereco    string    `json:"endereco,omitempty"`
	CPF         *string   `json:"cpf,omitempty"`
	CNPJ        *string   `json:"cnpj,omitempty"`
	DataCriacao time.Time `json:"dataCriacao"`
}

// ClienteListResponse representa a lista paginada de clientes
type ClienteListResponse struct {
	Clientes      []ClienteResponse `json:"clientes"`
	Total         int64             `json:"total"`
	Pagina        int               `json:"pagina"`
	TamanhoPagina int               `json:"tamanhoPagina"`
	TotalPaginas  int               `json:"totalPaginas"`
}

// BuscarClientesRequest representa os filtros de busca
type BuscarClientesRequest struct {
	Termo         string `form:"termo"`
	Pagina        int    `form:"pagina" binding:"min=1"`
	TamanhoPagina int    `form:"tamanhoPagina" binding:"min=1,max=100"`
}
