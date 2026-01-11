package entidades

import (
	"time"

	"github.com/google/uuid"
)

type Cliente struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UsuarioID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"usuarioId" validate:"required"`
	Nome            string     `gorm:"type:varchar(255);not null" json:"nome" validate:"required"`
	Email           string     `gorm:"type:varchar(255)" json:"email" validate:"omitempty,email"`
	Telefone        string     `gorm:"type:varchar(20)" json:"telefone"`
	CPF             string     `gorm:"type:varchar(14)" json:"cpf"`
	CNPJ            string     `gorm:"type:varchar(18)" json:"cnpj"`
	Endereco        string     `gorm:"type:varchar(255)" json:"endereco"`
	Cidade          string     `gorm:"type:varchar(100)" json:"cidade"`
	Estado          string     `gorm:"type:varchar(2)" json:"estado"`
	CEP             string     `gorm:"type:varchar(10)" json:"cep"`
	Observacoes     string     `gorm:"type:text" json:"observacoes"`
	DataCriacao     time.Time  `gorm:"autoCreateTime" json:"dataCriacao"`
	DataAtualizacao time.Time  `gorm:"autoUpdateTime" json:"dataAtualizacao"`

	// Relacionamentos
	Usuario   Usuario    `gorm:"foreignKey:UsuarioID" json:"-"`
	Cobrancas []Cobranca `gorm:"foreignKey:ClienteID" json:"-"`
}

// TableName sobrescreve o nome da tabela
func (Cliente) TableName() string {
	return "clientes"
}

// FormatarTelefone formata o telefone para envio WhatsApp
func (c *Cliente) FormatarTelefone() string {
	// Remove caracteres especiais e deixa só números
	telefone := c.Telefone

	// Lógica de formatação pode ser adicionada aqui
	// Por exemplo: +5511999999999

	return telefone
}
