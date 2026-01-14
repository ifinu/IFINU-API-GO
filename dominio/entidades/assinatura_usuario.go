package entidades

import (
	"time"

	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dominio/enums"
)

type StatusAssinatura string

const (
	StatusPeriodoGratuito     StatusAssinatura = "PERIODO_GRATUITO"
	StatusAtiva               StatusAssinatura = "ATIVA"
	StatusPendentePagamento   StatusAssinatura = "PENDENTE_PAGAMENTO"
	StatusBloqueada           StatusAssinatura = "BLOQUEADA"
	StatusCancelada           StatusAssinatura = "CANCELADA"
	StatusVitalicia           StatusAssinatura = "VITALICIA"
)

type AssinaturaUsuario struct {
	ID                         uuid.UUID                `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UsuarioID                  uuid.UUID                `gorm:"type:uuid;uniqueIndex;not null" json:"usuarioId" validate:"required"`
	Status                     StatusAssinatura         `gorm:"type:varchar(30);not null;default:'PERIODO_GRATUITO'" json:"status"`
	PlanoAssinatura            enums.PlanoAssinatura    `gorm:"type:varchar(20);default:'MENSAL'" json:"planoAssinatura"`
	DataInicioPeriodoGratuito  *time.Time               `gorm:"type:timestamp" json:"dataInicioPeriodoGratuito"`
	DataFimPeriodoGratuito     *time.Time               `gorm:"type:timestamp" json:"dataFimPeriodoGratuito"`
	DataUltimaCobranca         *time.Time               `gorm:"type:timestamp" json:"dataUltimaCobranca"`
	DataProximaCobranca        *time.Time               `gorm:"type:timestamp" json:"dataProximaCobranca"`
	DataCancelamento           *time.Time               `gorm:"type:timestamp" json:"dataCancelamento"`
	DataBloqueio               *time.Time               `gorm:"type:timestamp" json:"dataBloqueio"`
	ValorMensal                float64                  `gorm:"type:numeric(10,2);default:39.99" json:"valorMensal"`
	Currency                   string                   `gorm:"type:varchar(3);default:'BRL'" json:"currency"`
	Country                    string                   `gorm:"type:varchar(2);default:'BR'" json:"country"`
	AbacateCustomerID          string                   `gorm:"type:varchar(255)" json:"abacateCustomerId"`
	UltimaTransacaoID          string                   `gorm:"type:varchar(255)" json:"ultimaTransacaoId"`
	TentativasCobranca         int                      `gorm:"type:integer;default:0" json:"tentativasCobranca"`
	DiasTolerancia             int                      `gorm:"type:integer;default:3" json:"diasTolerancia"`
	Observacoes                string                   `gorm:"type:text" json:"observacoes"`
	DataCriacao                time.Time                `gorm:"autoCreateTime" json:"dataCriacao"`
	DataAtualizacao            time.Time                `gorm:"autoUpdateTime" json:"dataAtualizacao"`

	// Relacionamento
	Usuario Usuario `gorm:"foreignKey:UsuarioID" json:"-"`
}

// TableName sobrescreve o nome da tabela
func (AssinaturaUsuario) TableName() string {
	return "assinaturas_usuario"
}

// IsAtiva verifica se a assinatura está ativa
func (a *AssinaturaUsuario) IsAtiva() bool {
	return a.Status == StatusAtiva || a.Status == StatusPeriodoGratuito || a.Status == StatusVitalicia
}

// IsBloqueada verifica se a assinatura está bloqueada
func (a *AssinaturaUsuario) IsBloqueada() bool {
	return a.Status == StatusBloqueada
}

// Bloquear bloqueia a assinatura
func (a *AssinaturaUsuario) Bloquear(motivo string) {
	a.Status = StatusBloqueada
	agora := time.Now()
	a.DataBloqueio = &agora
	a.Observacoes = motivo
}

// Desbloquear desbloqueia a assinatura
func (a *AssinaturaUsuario) Desbloquear() {
	a.Status = StatusAtiva
	a.DataBloqueio = nil
}

// Cancelar cancela a assinatura
func (a *AssinaturaUsuario) Cancelar() {
	a.Status = StatusCancelada
	agora := time.Now()
	a.DataCancelamento = &agora
}
