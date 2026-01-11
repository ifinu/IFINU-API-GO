package repositorio

import (
	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dominio/entidades"
	"gorm.io/gorm"
)

type UsuarioRepositorio struct {
	db *gorm.DB
}

func NovoUsuarioRepositorio(db *gorm.DB) *UsuarioRepositorio {
	return &UsuarioRepositorio{db: db}
}

// BuscarPorEmail encontra um usuário pelo email
func (r *UsuarioRepositorio) BuscarPorEmail(email string) (*entidades.Usuario, error) {
	var usuario entidades.Usuario
	err := r.db.Where("email = ?", email).First(&usuario).Error
	if err != nil {
		return nil, err
	}
	return &usuario, nil
}

// BuscarPorID encontra um usuário pelo ID
func (r *UsuarioRepositorio) BuscarPorID(id uuid.UUID) (*entidades.Usuario, error) {
	var usuario entidades.Usuario
	err := r.db.Where("id = ?", id).First(&usuario).Error
	if err != nil {
		return nil, err
	}
	return &usuario, nil
}

// Criar cria um novo usuário
func (r *UsuarioRepositorio) Criar(usuario *entidades.Usuario) error {
	return r.db.Create(usuario).Error
}

// Atualizar atualiza um usuário existente
func (r *UsuarioRepositorio) Atualizar(usuario *entidades.Usuario) error {
	return r.db.Save(usuario).Error
}

// Deletar remove um usuário
func (r *UsuarioRepositorio) Deletar(id uuid.UUID) error {
	return r.db.Delete(&entidades.Usuario{}, id).Error
}

// ExistePorEmail verifica se um email já está cadastrado
func (r *UsuarioRepositorio) ExistePorEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&entidades.Usuario{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// ListarTodos retorna todos os usuários (use com cuidado)
func (r *UsuarioRepositorio) ListarTodos() ([]entidades.Usuario, error) {
	var usuarios []entidades.Usuario
	err := r.db.Find(&usuarios).Error
	return usuarios, err
}

// BuscarUsuariosComTrialExpirado retorna usuários com trial expirado
func (r *UsuarioRepositorio) BuscarUsuariosComTrialExpirado() ([]entidades.Usuario, error) {
	var usuarios []entidades.Usuario
	err := r.db.Where("trial_ativo = ? AND data_trial_inicio < NOW() - INTERVAL '14 days'", true).Find(&usuarios).Error
	return usuarios, err
}

// BuscarUsuariosComAssinaturaAtiva retorna usuários com assinatura ativa
func (r *UsuarioRepositorio) BuscarUsuariosComAssinaturaAtiva() ([]entidades.Usuario, error) {
	var usuarios []entidades.Usuario
	err := r.db.Joins("JOIN assinaturas_usuarios ON assinaturas_usuarios.usuario_id = usuarios.id").
		Where("assinaturas_usuarios.ativa = ?", true).
		Find(&usuarios).Error
	return usuarios, err
}
