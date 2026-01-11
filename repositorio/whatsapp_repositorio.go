package repositorio

import (
	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dominio/entidades"
	"gorm.io/gorm"
)

type WhatsAppRepositorio struct {
	db *gorm.DB
}

func NovoWhatsAppRepositorio(db *gorm.DB) *WhatsAppRepositorio {
	return &WhatsAppRepositorio{db: db}
}

// BuscarPorUsuario encontra a conexão WhatsApp de um usuário
func (r *WhatsAppRepositorio) BuscarPorUsuario(usuarioID uuid.UUID) (*entidades.WhatsAppConexao, error) {
	var conexao entidades.WhatsAppConexao
	err := r.db.Where("usuario_id = ?", usuarioID).First(&conexao).Error
	if err != nil {
		return nil, err
	}
	return &conexao, nil
}

// BuscarPorNomeInstancia encontra uma conexão pelo nome da instância
func (r *WhatsAppRepositorio) BuscarPorNomeInstancia(nomeInstancia string) (*entidades.WhatsAppConexao, error) {
	var conexao entidades.WhatsAppConexao
	err := r.db.Where("nome_instancia = ?", nomeInstancia).First(&conexao).Error
	if err != nil {
		return nil, err
	}
	return &conexao, nil
}

// Criar cria uma nova conexão WhatsApp
func (r *WhatsAppRepositorio) Criar(conexao *entidades.WhatsAppConexao) error {
	return r.db.Create(conexao).Error
}

// Atualizar atualiza uma conexão existente
func (r *WhatsAppRepositorio) Atualizar(conexao *entidades.WhatsAppConexao) error {
	return r.db.Save(conexao).Error
}

// Deletar remove uma conexão WhatsApp
func (r *WhatsAppRepositorio) Deletar(usuarioID uuid.UUID) error {
	return r.db.Where("usuario_id = ?", usuarioID).Delete(&entidades.WhatsAppConexao{}).Error
}

// BuscarConexoesAtivas retorna todas as conexões ativas
func (r *WhatsAppRepositorio) BuscarConexoesAtivas() ([]entidades.WhatsAppConexao, error) {
	var conexoes []entidades.WhatsAppConexao
	err := r.db.Preload("Usuario").
		Where("conectado = ?", true).
		Find(&conexoes).Error
	return conexoes, err
}

// ExistePorUsuario verifica se já existe uma conexão para o usuário
func (r *WhatsAppRepositorio) ExistePorUsuario(usuarioID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&entidades.WhatsAppConexao{}).
		Where("usuario_id = ?", usuarioID).
		Count(&count).Error
	return count > 0, err
}
