package entidades

import (
	"time"

	"github.com/google/uuid"
)

type StatusConexao string

const (
	StatusConexaoDesconectado  StatusConexao = "DESCONECTADO"
	StatusConexaoConectado     StatusConexao = "CONECTADO"
	StatusConexaoConectando    StatusConexao = "CONECTANDO"
	StatusConexaoErro          StatusConexao = "ERRO"
)

type WhatsAppConexao struct {
	ID                  int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UsuarioID           uuid.UUID      `gorm:"type:uuid;not null;index" json:"usuarioId" validate:"required"`
	InstanceName        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"instanceName" validate:"required"`
	Status              StatusConexao  `gorm:"type:varchar(20);default:'DESCONECTADO'" json:"status"`
	QRCode              string         `gorm:"type:text" json:"qrCode"`
	NumeroConectado     string         `gorm:"type:varchar(20)" json:"numeroConectado"`
	DataConexao         *time.Time     `gorm:"type:timestamp" json:"dataConexao"`
	DataUltimaAtividade *time.Time     `gorm:"type:timestamp" json:"dataUltimaAtividade"`
	MensagensEnviadas   int            `gorm:"type:integer;default:0" json:"mensagensEnviadas"`
	MensagensSucesso    int            `gorm:"type:integer;default:0" json:"mensagensSucesso"`
	MensagensFalha      int            `gorm:"type:integer;default:0" json:"mensagensFalha"`
	DataCriacao         time.Time      `gorm:"autoCreateTime" json:"dataCriacao"`
	DataAtualizacao     time.Time      `gorm:"autoUpdateTime" json:"dataAtualizacao"`

	// Relacionamento
	Usuario Usuario `gorm:"foreignKey:UsuarioID" json:"-"`
}

// TableName sobrescreve o nome da tabela
func (WhatsAppConexao) TableName() string {
	return "whatsapp_conexoes"
}

// IsConectado verifica se está conectado
func (w *WhatsAppConexao) IsConectado() bool {
	return w.Status == StatusConexaoConectado
}

// Conectar marca como conectado
func (w *WhatsAppConexao) Conectar(numero string) {
	w.Status = StatusConexaoConectado
	w.NumeroConectado = numero
	agora := time.Now()
	w.DataConexao = &agora
	w.DataUltimaAtividade = &agora
	w.QRCode = "" // Limpa QR Code após conectar
}

// Desconectar marca como desconectado
func (w *WhatsAppConexao) Desconectar() {
	w.Status = StatusConexaoDesconectado
	w.QRCode = ""
}

// IncrementarMensagemEnviada incrementa contador de mensagem
func (w *WhatsAppConexao) IncrementarMensagemEnviada(sucesso bool) {
	w.MensagensEnviadas++

	if sucesso {
		w.MensagensSucesso++
	} else {
		w.MensagensFalha++
	}

	agora := time.Now()
	w.DataUltimaAtividade = &agora
}

// TaxaSucesso calcula taxa de sucesso de envios
func (w *WhatsAppConexao) TaxaSucesso() float64 {
	if w.MensagensEnviadas == 0 {
		return 0
	}

	return (float64(w.MensagensSucesso) / float64(w.MensagensEnviadas)) * 100
}
