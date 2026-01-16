package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dominio/enums"
)

// CobrancaRequest representa a requisição de criação/atualização de cobrança
type CobrancaRequest struct {
	ClienteID       uuid.UUID                `json:"clienteId" binding:"required"`
	Valor           float64                  `json:"valor" binding:"required,gt=0"`
	Descricao       string                   `json:"descricao" binding:"required"`
	DataVencimento  time.Time                `json:"dataVencimento" binding:"required"`
	TipoRecorrencia enums.TipoRecorrencia    `json:"tipoRecorrencia"`
	EnviarWhatsApp  bool                     `json:"enviarWhatsapp"`
	EnviarEmail     bool                     `json:"enviarEmail"`
}

// CobrancaResponse representa a cobrança na resposta
type CobrancaResponse struct {
	ID                             uuid.UUID                `json:"id"`
	ClienteID                      uuid.UUID                `json:"clienteId"`
	Cliente                        *ClienteResponse         `json:"cliente,omitempty"`
	ClienteNome                    string                   `json:"clienteNome"`
	ClienteEmail                   string                   `json:"clienteEmail,omitempty"`
	ClienteTelefone                string                   `json:"clienteTelefone"`
	Valor                          float64                  `json:"valor"`
	Descricao                      string                   `json:"descricao"`
	Status                         enums.StatusCobranca     `json:"status"`
	StatusDescricao                string                   `json:"statusDescricao"`
	DataVencimento                 time.Time                `json:"dataVencimento"`
	DataPagamento                  *time.Time               `json:"dataPagamento,omitempty"`
	TipoRecorrencia                enums.TipoRecorrencia    `json:"tipoRecorrencia"`
	TipoRecorrenciaDescricao       string                   `json:"tipoRecorrenciaDescricao"`
	Recorrente                     bool                     `json:"recorrente"`
	Vencida                        bool                     `json:"vencida"`
	DiasAtraso                     int                      `json:"diasAtraso"`
	LinkPagamento                  string                   `json:"linkPagamento,omitempty"`
	NotificacaoEnviada             bool                     `json:"notificacaoEnviada"`
	NotificacaoLembreteEnviada     bool                     `json:"notificacaoLembreteEnviada"`
	NotificacaoVencimentoEnviada   bool                     `json:"notificacaoVencimentoEnviada"`
	DataCriacao                    time.Time                `json:"dataCriacao"`
}

// CobrancaListResponse representa a lista paginada de cobranças
type CobrancaListResponse struct {
	Cobrancas     []CobrancaResponse `json:"cobrancas"`
	Total         int64              `json:"total"`
	Pagina        int                `json:"pagina"`
	TamanhoPagina int                `json:"tamanhoPagina"`
	TotalPaginas  int                `json:"totalPaginas"`
}

// BuscarCobrancasRequest representa os filtros de busca de cobranças
type BuscarCobrancasRequest struct {
	Status        *enums.StatusCobranca `form:"status"`
	DataInicio    *time.Time            `form:"dataInicio"`
	DataFim       *time.Time            `form:"dataFim"`
	ClienteID     *uuid.UUID            `form:"clienteId"`
	Pagina        int                   `form:"pagina" binding:"min=0"`
	TamanhoPagina int                   `form:"tamanhoPagina" binding:"min=1,max=100"`
}

// AtualizarStatusRequest representa a requisição de atualização de status
type AtualizarStatusRequest struct {
	Status enums.StatusCobranca `json:"status" binding:"required"`
}

// EstatisticasCobrancasResponse representa as estatísticas de cobranças
type EstatisticasCobrancasResponse struct {
	TotalPendente   int64   `json:"totalPendente"`
	TotalPaga       int64   `json:"totalPaga"`
	TotalVencida    int64   `json:"totalVencida"`
	TotalCancelada  int64   `json:"totalCancelada"`
	ValorPendente   float64 `json:"valorPendente"`
	ValorPago       float64 `json:"valorPago"`
	ValorVencido    float64 `json:"valorVencido"`
}
