package entidades

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Usuario struct {
	ID                     uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	NomeCompleto           string     `gorm:"type:varchar(255);not null" json:"nomeCompleto" validate:"required"`
	Email                  string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email" validate:"required,email"`
	Telefone               string     `gorm:"type:varchar(20)" json:"telefone"`
	NomeEmpresa            string     `gorm:"type:varchar(255)" json:"nomeEmpresa"`
	TipoEmpresa            string     `gorm:"type:varchar(50)" json:"tipoEmpresa"`
	CNPJ                   string     `gorm:"type:varchar(18)" json:"cnpj"`
	SenhaHash              string     `gorm:"type:varchar(255);not null" json:"-"`
	Ativo                  bool       `gorm:"default:true" json:"ativo"`
	EmailVerificado        bool       `gorm:"default:false" json:"emailVerificado"`
	Vitalicio              bool       `gorm:"default:false" json:"vitalicio"`
	StripeAccountID        string     `gorm:"type:varchar(255)" json:"stripeAccountId"`
	DuasEtapasAtivo        bool       `gorm:"default:false" json:"duasEtapasAtivo"`
	DuasEtapasSecret       string     `gorm:"type:varchar(255)" json:"-"`
	CodigosRecuperacao2FA  string     `gorm:"column:codigos_recuperacao_2fa;type:text" json:"-"`
	DataAtivacao2FA        *time.Time `gorm:"column:data_ativacao_2fa;type:timestamp" json:"dataAtivacao2FA"`
	DataTrialInicio        *time.Time `gorm:"type:timestamp" json:"dataTrialInicio"`
	TrialAtivo             bool       `gorm:"default:true" json:"trialAtivo"`
	DataCriacao            time.Time  `gorm:"autoCreateTime" json:"dataCriacao"`
	DataUltimoAcesso       *time.Time `gorm:"type:timestamp" json:"dataUltimoAcesso"`

	// Relacionamentos (não carrega automaticamente)
	Clientes             []Cliente           `gorm:"foreignKey:UsuarioID" json:"-"`
	Cobrancas            []Cobranca          `gorm:"foreignKey:UsuarioID" json:"-"`
	WhatsAppConexoes     []WhatsAppConexao   `gorm:"foreignKey:UsuarioID" json:"-"`
	AssinaturaUsuario    *AssinaturaUsuario  `gorm:"foreignKey:UsuarioID" json:"-"`
}

// TableName sobrescreve o nome da tabela
func (Usuario) TableName() string {
	return "usuarios"
}

// BeforeCreate hook do GORM - executa antes de criar
func (u *Usuario) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	// Define trial de 14 dias para novos usuários
	if u.DataTrialInicio == nil {
		agora := time.Now()
		u.DataTrialInicio = &agora
		u.TrialAtivo = true
	}

	return nil
}

// IsTrialExpirado verifica se o trial expirou
func (u *Usuario) IsTrialExpirado() bool {
	if !u.TrialAtivo || u.DataTrialInicio == nil {
		return true
	}

	dataExpiracao := u.DataTrialInicio.AddDate(0, 0, 14) // 14 dias
	return time.Now().After(dataExpiracao)
}

// DiasRestantesTrial retorna quantos dias faltam no trial
func (u *Usuario) DiasRestantesTrial() int {
	if u.DataTrialInicio == nil {
		return 0
	}

	dataExpiracao := u.DataTrialInicio.AddDate(0, 0, 14)
	diasRestantes := int(time.Until(dataExpiracao).Hours() / 24)

	if diasRestantes < 0 {
		return 0
	}

	return diasRestantes
}

// AtualizarUltimoAcesso atualiza data do último acesso
func (u *Usuario) AtualizarUltimoAcesso() {
	agora := time.Now()
	u.DataUltimoAcesso = &agora
}
