package entidades

import (
	"time"

	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dominio/enums"
)

type Cobranca struct {
	ID                             int64                      `gorm:"primaryKey;autoIncrement" json:"id"`
	ClienteID                      int64                      `gorm:"not null;index" json:"clienteId" validate:"required"`
	UsuarioID                      uuid.UUID                  `gorm:"type:uuid;not null;index" json:"usuarioId" validate:"required"`
	Valor                          float64                    `gorm:"type:numeric(10,2);not null" json:"valor" validate:"required,gt=0"`
	DataVencimento                 time.Time                  `gorm:"type:timestamp;not null" json:"dataVencimento" validate:"required"`
	DataPagamento                  *time.Time                 `gorm:"type:timestamp" json:"dataPagamento"`
	Status                         enums.StatusCobranca       `gorm:"type:varchar(20);not null;default:'PENDENTE'" json:"status"`
	Descricao                      string                     `gorm:"type:varchar(255)" json:"descricao"`
	Observacoes                    string                     `gorm:"type:text" json:"observacoes"`
	TipoRecorrencia                enums.TipoRecorrencia      `gorm:"type:varchar(20);default:'UNICA'" json:"tipoRecorrencia"`
	IntervaloPeriodo               int                        `gorm:"type:integer" json:"intervaloPeriodo"`
	UnidadeTempo                   string                     `gorm:"type:varchar(20)" json:"unidadeTempo"`
	ProximaCobranca                *time.Time                 `gorm:"type:timestamp" json:"proximaCobranca"`
	RecorrenciaAtiva               bool                       `gorm:"default:false" json:"recorrenciaAtiva"`
	NotificacaoEnviada             bool                       `gorm:"default:false" json:"notificacaoEnviada"`
	NotificacaoLembreteEnviada     bool                       `gorm:"default:false" json:"notificacaoLembreteEnviada"`
	NotificacaoVencimentoEnviada   bool                       `gorm:"default:false" json:"notificacaoVencimentoEnviada"`
	ConfirmacaoPagamentoEnviada    bool                       `gorm:"default:false" json:"confirmacaoPagamentoEnviada"`
	TentativasNotificacao          int                        `gorm:"type:integer;not null;default:0" json:"tentativasNotificacao"`
	AsaasPaymentID                 string                     `gorm:"type:varchar(255)" json:"asaasPaymentId"`
	AsaasStatus                    string                     `gorm:"type:varchar(50)" json:"asaasStatus"`
	AsaasPaymentURL                string                     `gorm:"type:text" json:"asaasPaymentUrl"`
	StripeSessionID                string                     `gorm:"type:varchar(255)" json:"stripeSessionId"`
	StripeCheckoutID               string                     `gorm:"type:varchar(255)" json:"stripeCheckoutId"`
	StripePaymentIntentID          string                     `gorm:"type:varchar(255)" json:"stripePaymentIntentId"`
	EmailCliente                   string                     `gorm:"type:varchar(255)" json:"emailCliente"`
	DataCriacao                    time.Time                  `gorm:"autoCreateTime" json:"dataCriacao"`
	DataAtualizacao                time.Time                  `gorm:"autoUpdateTime" json:"dataAtualizacao"`

	// Relacionamentos
	Cliente Cliente `gorm:"foreignKey:ClienteID" json:"cliente,omitempty"`
	Usuario Usuario `gorm:"foreignKey:UsuarioID" json:"-"`
}

// TableName sobrescreve o nome da tabela
func (Cobranca) TableName() string {
	return "cobrancas"
}

// IsPaga verifica se a cobrança foi paga
func (c *Cobranca) IsPaga() bool {
	return c.Status == enums.StatusCobrancaPago
}

// IsVencida verifica se a cobrança está vencida
func (c *Cobranca) IsVencida() bool {
	if c.IsPaga() {
		return false
	}
	return time.Now().After(c.DataVencimento)
}

// DiasParaVencimento retorna quantos dias faltam para o vencimento
func (c *Cobranca) DiasParaVencimento() int {
	if c.IsPaga() {
		return 0
	}

	diasRestantes := int(time.Until(c.DataVencimento).Hours() / 24)

	if diasRestantes < 0 {
		return 0
	}

	return diasRestantes
}

// MarcarNotificacaoEnviada marca notificação como enviada
func (c *Cobranca) MarcarNotificacaoEnviada() {
	c.NotificacaoEnviada = true
}

// MarcarNotificacaoLembreteEnviada marca lembrete como enviado
func (c *Cobranca) MarcarNotificacaoLembreteEnviada() {
	c.NotificacaoLembreteEnviada = true
	c.NotificacaoEnviada = true
}

// MarcarNotificacaoVencimentoEnviada marca vencimento como enviado
func (c *Cobranca) MarcarNotificacaoVencimentoEnviada() {
	c.NotificacaoVencimentoEnviada = true
	c.NotificacaoEnviada = true
}

// MarcarConfirmacaoPagamentoEnviada marca confirmação como enviada
func (c *Cobranca) MarcarConfirmacaoPagamentoEnviada() {
	c.ConfirmacaoPagamentoEnviada = true
}

// MarcarComoPaga marca a cobrança como paga
func (c *Cobranca) MarcarComoPaga() {
	c.Status = enums.StatusCobrancaPago
	agora := time.Now()
	c.DataPagamento = &agora
}

// MarcarComoVencida marca a cobrança como vencida
func (c *Cobranca) MarcarComoVencida() {
	if !c.IsPaga() && c.IsVencida() {
		c.Status = enums.StatusCobrancaVencido
	}
}

// CalcularProximaCobranca calcula a próxima data de cobrança para recorrência
func (c *Cobranca) CalcularProximaCobranca() *time.Time {
	if !c.RecorrenciaAtiva {
		return nil
	}

	var proximaData time.Time

	switch c.TipoRecorrencia {
	case enums.TipoRecorrenciaMensal:
		proximaData = c.DataVencimento.AddDate(0, 1, 0)
	case enums.TipoRecorrenciaTrimestral:
		proximaData = c.DataVencimento.AddDate(0, 3, 0)
	case enums.TipoRecorrenciaSemestral:
		proximaData = c.DataVencimento.AddDate(0, 6, 0)
	case enums.TipoRecorrenciaAnual:
		proximaData = c.DataVencimento.AddDate(1, 0, 0)
	case enums.TipoRecorrenciaPersonalizado:
		// Implementar lógica personalizada
		if c.UnidadeTempo == "DIAS" {
			proximaData = c.DataVencimento.AddDate(0, 0, c.IntervaloPeriodo)
		} else if c.UnidadeTempo == "MESES" {
			proximaData = c.DataVencimento.AddDate(0, c.IntervaloPeriodo, 0)
		} else if c.UnidadeTempo == "ANOS" {
			proximaData = c.DataVencimento.AddDate(c.IntervaloPeriodo, 0, 0)
		}
	default:
		return nil
	}

	return &proximaData
}
