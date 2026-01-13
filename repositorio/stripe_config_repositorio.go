package repositorio

import (
	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dominio/entidades"
	"gorm.io/gorm"
)

type StripeConfigRepositorio struct {
	db *gorm.DB
}

func NovoStripeConfigRepositorio(db *gorm.DB) *StripeConfigRepositorio {
	return &StripeConfigRepositorio{db: db}
}

// BuscarPorUsuario encontra a configuração Stripe de um usuário
func (r *StripeConfigRepositorio) BuscarPorUsuario(usuarioID uuid.UUID) (*entidades.StripeConfig, error) {
	var config entidades.StripeConfig
	err := r.db.Where("usuario_id = ?", usuarioID).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Criar cria uma nova configuração Stripe
func (r *StripeConfigRepositorio) Criar(config *entidades.StripeConfig) error {
	return r.db.Create(config).Error
}

// Atualizar atualiza uma configuração existente
func (r *StripeConfigRepositorio) Atualizar(config *entidades.StripeConfig) error {
	return r.db.Save(config).Error
}

// Deletar remove uma configuração Stripe
func (r *StripeConfigRepositorio) Deletar(usuarioID uuid.UUID) error {
	return r.db.Where("usuario_id = ?", usuarioID).Delete(&entidades.StripeConfig{}).Error
}

// Existe verifica se já existe configuração para o usuário
func (r *StripeConfigRepositorio) Existe(usuarioID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&entidades.StripeConfig{}).
		Where("usuario_id = ?", usuarioID).
		Count(&count).Error
	return count > 0, err
}
