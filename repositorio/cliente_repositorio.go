package repositorio

import (
	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dominio/entidades"
	"gorm.io/gorm"
)

type ClienteRepositorio struct {
	db *gorm.DB
}

func NovoClienteRepositorio(db *gorm.DB) *ClienteRepositorio {
	return &ClienteRepositorio{db: db}
}

// BuscarPorID encontra um cliente pelo ID (com validação de usuário)
func (r *ClienteRepositorio) BuscarPorID(id int64, usuarioID uuid.UUID) (*entidades.Cliente, error) {
	var cliente entidades.Cliente
	err := r.db.Where("id = ? AND usuario_id = ?", id, usuarioID).First(&cliente).Error
	if err != nil {
		return nil, err
	}
	return &cliente, nil
}

// BuscarPorUsuario retorna todos os clientes de um usuário
func (r *ClienteRepositorio) BuscarPorUsuario(usuarioID uuid.UUID) ([]entidades.Cliente, error) {
	var clientes []entidades.Cliente
	err := r.db.Where("usuario_id = ?", usuarioID).Order("nome ASC").Find(&clientes).Error
	return clientes, err
}

// BuscarPorEmail encontra um cliente pelo email (com validação de usuário)
func (r *ClienteRepositorio) BuscarPorEmail(email string, usuarioID uuid.UUID) (*entidades.Cliente, error) {
	var cliente entidades.Cliente
	err := r.db.Where("email = ? AND usuario_id = ?", email, usuarioID).First(&cliente).Error
	if err != nil {
		return nil, err
	}
	return &cliente, nil
}

// BuscarPorTelefone encontra um cliente pelo telefone (com validação de usuário)
func (r *ClienteRepositorio) BuscarPorTelefone(telefone string, usuarioID uuid.UUID) (*entidades.Cliente, error) {
	var cliente entidades.Cliente
	err := r.db.Where("telefone = ? AND usuario_id = ?", telefone, usuarioID).First(&cliente).Error
	if err != nil {
		return nil, err
	}
	return &cliente, nil
}

// Criar cria um novo cliente
func (r *ClienteRepositorio) Criar(cliente *entidades.Cliente) error {
	return r.db.Create(cliente).Error
}

// Atualizar atualiza um cliente existente
func (r *ClienteRepositorio) Atualizar(cliente *entidades.Cliente) error {
	return r.db.Save(cliente).Error
}

// Deletar remove um cliente (com validação de usuário)
func (r *ClienteRepositorio) Deletar(id int64, usuarioID uuid.UUID) error {
	return r.db.Where("id = ? AND usuario_id = ?", id, usuarioID).Delete(&entidades.Cliente{}).Error
}

// Buscar com filtros (nome, email, telefone)
func (r *ClienteRepositorio) Buscar(usuarioID uuid.UUID, termo string, pagina int, tamanhoPagina int) ([]entidades.Cliente, int64, error) {
	var clientes []entidades.Cliente
	var total int64

	query := r.db.Model(&entidades.Cliente{}).Where("usuario_id = ?", usuarioID)

	if termo != "" {
		query = query.Where("nome ILIKE ? OR email ILIKE ? OR telefone ILIKE ?",
			"%"+termo+"%", "%"+termo+"%", "%"+termo+"%")
	}

	// Contar total
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Buscar com paginação
	offset := (pagina - 1) * tamanhoPagina
	err = query.Order("nome ASC").Offset(offset).Limit(tamanhoPagina).Find(&clientes).Error

	return clientes, total, err
}

// ContarPorUsuario retorna o total de clientes de um usuário
func (r *ClienteRepositorio) ContarPorUsuario(usuarioID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&entidades.Cliente{}).Where("usuario_id = ?", usuarioID).Count(&count).Error
	return count, err
}

// ExistePorEmail verifica se um email já está cadastrado para o usuário
func (r *ClienteRepositorio) ExistePorEmail(email string, usuarioID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&entidades.Cliente{}).
		Where("email = ? AND usuario_id = ?", email, usuarioID).
		Count(&count).Error
	return count > 0, err
}
