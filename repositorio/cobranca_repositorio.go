package repositorio

import (
	"time"

	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dominio/entidades"
	"github.com/ifinu/ifinu-api-go/dominio/enums"
	"gorm.io/gorm"
)

type CobrancaRepositorio struct {
	db *gorm.DB
}

func NovoCobrancaRepositorio(db *gorm.DB) *CobrancaRepositorio {
	return &CobrancaRepositorio{db: db}
}

// BuscarPorID encontra uma cobrança pelo ID (com validação de usuário)
func (r *CobrancaRepositorio) BuscarPorID(id uuid.UUID, usuarioID uuid.UUID) (*entidades.Cobranca, error) {
	var cobranca entidades.Cobranca
	err := r.db.Preload("Cliente").
		Where("cobrancas.id = ? AND cobrancas.usuario_id = ?", id, usuarioID).
		First(&cobranca).Error
	if err != nil {
		return nil, err
	}
	return &cobranca, nil
}

// BuscarPorUsuario retorna todas as cobranças de um usuário
func (r *CobrancaRepositorio) BuscarPorUsuario(usuarioID uuid.UUID) ([]entidades.Cobranca, error) {
	var cobrancas []entidades.Cobranca
	err := r.db.Preload("Cliente").
		Where("usuario_id = ?", usuarioID).
		Order("data_vencimento DESC").
		Find(&cobrancas).Error
	return cobrancas, err
}

// BuscarPorCliente retorna todas as cobranças de um cliente específico
func (r *CobrancaRepositorio) BuscarPorCliente(clienteID uuid.UUID, usuarioID uuid.UUID) ([]entidades.Cobranca, error) {
	var cobrancas []entidades.Cobranca
	err := r.db.Preload("Cliente").
		Where("cliente_id = ? AND usuario_id = ?", clienteID, usuarioID).
		Order("data_vencimento DESC").
		Find(&cobrancas).Error
	return cobrancas, err
}

// BuscarPorStatus retorna cobranças por status
func (r *CobrancaRepositorio) BuscarPorStatus(status enums.StatusCobranca, usuarioID uuid.UUID) ([]entidades.Cobranca, error) {
	var cobrancas []entidades.Cobranca
	err := r.db.Preload("Cliente").
		Where("status = ? AND usuario_id = ?", status, usuarioID).
		Order("data_vencimento DESC").
		Find(&cobrancas).Error
	return cobrancas, err
}

// Criar cria uma nova cobrança
func (r *CobrancaRepositorio) Criar(cobranca *entidades.Cobranca) error {
	return r.db.Create(cobranca).Error
}

// Atualizar atualiza uma cobrança existente
func (r *CobrancaRepositorio) Atualizar(cobranca *entidades.Cobranca) error {
	return r.db.Save(cobranca).Error
}

// Deletar remove uma cobrança (com validação de usuário)
func (r *CobrancaRepositorio) Deletar(id uuid.UUID, usuarioID uuid.UUID) error {
	return r.db.Where("id = ? AND usuario_id = ?", id, usuarioID).Delete(&entidades.Cobranca{}).Error
}

// Buscar com filtros
func (r *CobrancaRepositorio) Buscar(
	usuarioID uuid.UUID,
	status *enums.StatusCobranca,
	dataInicio *time.Time,
	dataFim *time.Time,
	pagina int,
	tamanhoPagina int,
) ([]entidades.Cobranca, int64, error) {
	var cobrancas []entidades.Cobranca
	var total int64

	query := r.db.Model(&entidades.Cobranca{}).Where("usuario_id = ?", usuarioID)

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if dataInicio != nil {
		query = query.Where("data_vencimento >= ?", *dataInicio)
	}

	if dataFim != nil {
		query = query.Where("data_vencimento <= ?", *dataFim)
	}

	// Contar total
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Buscar com paginação
	offset := (pagina - 1) * tamanhoPagina
	err = query.Preload("Cliente").
		Order("data_vencimento DESC").
		Offset(offset).
		Limit(tamanhoPagina).
		Find(&cobrancas).Error

	return cobrancas, total, err
}

// BuscarCobrancasVencendoHoje retorna cobranças que vencem hoje (para notificações)
func (r *CobrancaRepositorio) BuscarCobrancasVencendoHoje() ([]entidades.Cobranca, error) {
	hoje := time.Now().Truncate(24 * time.Hour)
	amanha := hoje.Add(24 * time.Hour)

	var cobrancas []entidades.Cobranca
	err := r.db.Preload("Cliente").Preload("Usuario").
		Where("status = ? AND data_vencimento >= ? AND data_vencimento < ? AND notificacao_vencimento_enviada = ?",
			enums.StatusCobrancaPendente, hoje, amanha, false).
		Find(&cobrancas).Error

	return cobrancas, err
}

// BuscarCobrancasParaLembrete retorna cobranças que precisam de lembrete (3 dias antes)
func (r *CobrancaRepositorio) BuscarCobrancasParaLembrete() ([]entidades.Cobranca, error) {
	tresDiasDepois := time.Now().Add(3 * 24 * time.Hour).Truncate(24 * time.Hour)
	quatroDiasDepois := tresDiasDepois.Add(24 * time.Hour)

	var cobrancas []entidades.Cobranca
	err := r.db.Preload("Cliente").Preload("Usuario").
		Where("status = ? AND data_vencimento >= ? AND data_vencimento < ? AND notificacao_lembrete_enviada = ?",
			enums.StatusCobrancaPendente, tresDiasDepois, quatroDiasDepois, false).
		Find(&cobrancas).Error

	return cobrancas, err
}

// BuscarCobrancasVencidas retorna cobranças vencidas
func (r *CobrancaRepositorio) BuscarCobrancasVencidas() ([]entidades.Cobranca, error) {
	hoje := time.Now().Truncate(24 * time.Hour)

	var cobrancas []entidades.Cobranca
	err := r.db.Preload("Cliente").Preload("Usuario").
		Where("status = ? AND data_vencimento < ?",
			enums.StatusCobrancaPendente, hoje).
		Find(&cobrancas).Error

	return cobrancas, err
}

// ContarPorStatus retorna contagem de cobranças por status para um usuário
func (r *CobrancaRepositorio) ContarPorStatus(usuarioID uuid.UUID) (map[enums.StatusCobranca]int64, error) {
	type Resultado struct {
		Status enums.StatusCobranca
		Total  int64
	}

	var resultados []Resultado
	err := r.db.Model(&entidades.Cobranca{}).
		Select("status, COUNT(*) as total").
		Where("usuario_id = ?", usuarioID).
		Group("status").
		Find(&resultados).Error

	if err != nil {
		return nil, err
	}

	contagem := make(map[enums.StatusCobranca]int64)
	for _, r := range resultados {
		contagem[r.Status] = r.Total
	}

	return contagem, nil
}

// CalcularValorTotalPorStatus retorna o valor total de cobranças por status
func (r *CobrancaRepositorio) CalcularValorTotalPorStatus(usuarioID uuid.UUID, status enums.StatusCobranca) (float64, error) {
	var total float64
	err := r.db.Model(&entidades.Cobranca{}).
		Where("usuario_id = ? AND status = ?", usuarioID, status).
		Select("COALESCE(SUM(valor), 0)").
		Scan(&total).Error

	return total, err
}
