package repositorio

import (
	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dominio/entidades"
	"gorm.io/gorm"
)

type AssinaturaRepositorio struct {
	db *gorm.DB
}

func NovoAssinaturaRepositorio(db *gorm.DB) *AssinaturaRepositorio {
	return &AssinaturaRepositorio{db: db}
}

// BuscarPorUsuario encontra a assinatura de um usuário
func (r *AssinaturaRepositorio) BuscarPorUsuario(usuarioID uuid.UUID) (*entidades.AssinaturaUsuario, error) {
	var assinatura entidades.AssinaturaUsuario
	err := r.db.Where("usuario_id = ?", usuarioID).First(&assinatura).Error
	if err != nil {
		return nil, err
	}
	return &assinatura, nil
}

// BuscarPorID encontra uma assinatura pelo ID
func (r *AssinaturaRepositorio) BuscarPorID(id int64) (*entidades.AssinaturaUsuario, error) {
	var assinatura entidades.AssinaturaUsuario
	err := r.db.Where("id = ?", id).First(&assinatura).Error
	if err != nil {
		return nil, err
	}
	return &assinatura, nil
}

// BuscarPorStripeSubscriptionID encontra uma assinatura pelo Stripe Subscription ID
func (r *AssinaturaRepositorio) BuscarPorStripeSubscriptionID(subscriptionID string) (*entidades.AssinaturaUsuario, error) {
	var assinatura entidades.AssinaturaUsuario
	err := r.db.Where("stripe_subscription_id = ?", subscriptionID).First(&assinatura).Error
	if err != nil {
		return nil, err
	}
	return &assinatura, nil
}

// BuscarPorStripeCustomerID encontra uma assinatura pelo Stripe Customer ID
func (r *AssinaturaRepositorio) BuscarPorStripeCustomerID(customerID string) (*entidades.AssinaturaUsuario, error) {
	var assinatura entidades.AssinaturaUsuario
	err := r.db.Where("stripe_customer_id = ?", customerID).First(&assinatura).Error
	if err != nil {
		return nil, err
	}
	return &assinatura, nil
}

// Criar cria uma nova assinatura
func (r *AssinaturaRepositorio) Criar(assinatura *entidades.AssinaturaUsuario) error {
	return r.db.Create(assinatura).Error
}

// Atualizar atualiza uma assinatura existente
func (r *AssinaturaRepositorio) Atualizar(assinatura *entidades.AssinaturaUsuario) error {
	return r.db.Save(assinatura).Error
}

// DesativarAssinaturaAnterior desativa qualquer assinatura ativa anterior do usuário
func (r *AssinaturaRepositorio) DesativarAssinaturaAnterior(usuarioID uuid.UUID) error {
	return r.db.Model(&entidades.AssinaturaUsuario{}).
		Where("usuario_id = ? AND status IN ?", usuarioID, []string{"ATIVA", "PERIODO_GRATUITO"}).
		Update("status", entidades.StatusCancelada).Error
}

// BuscarAssinaturasVencidas retorna assinaturas que venceram
func (r *AssinaturaRepositorio) BuscarAssinaturasVencidas() ([]entidades.AssinaturaUsuario, error) {
	var assinaturas []entidades.AssinaturaUsuario
	err := r.db.Preload("Usuario").
		Where("status = ? AND data_proxima_cobranca < NOW()", entidades.StatusAtiva).
		Find(&assinaturas).Error
	return assinaturas, err
}

// BuscarTodasPorUsuario retorna todas as assinaturas de um usuário (histórico)
func (r *AssinaturaRepositorio) BuscarTodasPorUsuario(usuarioID uuid.UUID) ([]entidades.AssinaturaUsuario, error) {
	var assinaturas []entidades.AssinaturaUsuario
	err := r.db.Where("usuario_id = ?", usuarioID).
		Order("data_inicio DESC").
		Find(&assinaturas).Error
	return assinaturas, err
}
