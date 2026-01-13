package entidades

import (
	"time"

	"github.com/google/uuid"
)

type StripeConfig struct {
	ID                 int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UsuarioID          uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"usuarioId"`
	PublishableKey     string    `gorm:"type:varchar(255);not null" json:"publishableKey"`
	SecretKeyEncrypted string    `gorm:"type:text;not null" json:"-"` // Nunca retornar no JSON
	TestMode           bool      `gorm:"not null;default:true" json:"testMode"`
	DataCriacao        time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"dataCriacao"`
	DataAtualizacao    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"dataAtualizacao"`

	// Relacionamento
	Usuario Usuario `gorm:"foreignKey:UsuarioID" json:"-"`
}

// TableName sobrescreve o nome da tabela
func (StripeConfig) TableName() string {
	return "stripe_config"
}

// MaskSecretKey retorna a chave secreta mascarada
func (s *StripeConfig) MaskSecretKey() string {
	if len(s.SecretKeyEncrypted) < 10 {
		return "***"
	}
	// Mostrar apenas primeiros 7 caracteres e últimos 4
	prefix := s.SecretKeyEncrypted[:7]
	return prefix + "..." + s.SecretKeyEncrypted[len(s.SecretKeyEncrypted)-4:]
}
