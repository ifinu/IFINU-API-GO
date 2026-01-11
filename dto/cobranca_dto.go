package dto

import (
	"time"

	"github.com/ifinu/ifinu-api-go/dominio/enums"
)

// CobrancaRequest representa a requisição de criação/atualização de cobrança
type CobrancaRequest struct {
	ClienteID       int64                    `json:"clienteId" binding:"required"`
	Valor           float64                  `json:"valor" binding:"required,gt=0"`
	Descricao       string                   `json:"descricao" binding:"required"`
	DataVencimento  time.Time                `json:"dataVencimento" binding:"required"`
	TipoRecorrencia enums.TipoRecorrencia    `json:"tipoRecorrencia"`
	EnviarWhatsApp  bool                     `json:"enviarWhatsapp"`
	EnviarEmail     bool                     `json:"enviarEmail"`
}

// CobrancaResponse representa a cobrança na resposta
type CobrancaResponse struct {
	ID                             int64                    `json:"id"`
	ClienteID                      int64                    `json:"clienteId"`
	Cliente                        *ClienteResponse         `json:"cliente,omitempty"`
	Valor                          float64                  `json:"valor"`
	Descricao                      string                   `json:"descricao"`
	Status                         enums.StatusCobranca     `json:"status"`
	DataVencimento                 time.Time                `json:"dataVencimento"`
	DataPagamento                  *time.Time               `json:"dataPagamento,omitempty"`
	TipoRecorrencia                enums.TipoRecorrencia    `json:"tipoRecorrencia"`
	LinkPagamento                  string                   `json:"linkPagamento,omitempty"`
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
	ClienteID     *int64                `form:"clienteId"`
	Pagina        int                   `form:"pagina" binding:"min=1"`
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
