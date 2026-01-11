package dto

import "time"

// LoginRequest representa a requisição de login
type LoginRequest struct {
	Email string `json:"email" binding:"required,email"`
	Senha string `json:"senha" binding:"required"`
}

// CadastroRequest representa a requisição de cadastro
type CadastroRequest struct {
	NomeCompleto string `json:"nomeCompleto" binding:"required,min=3"`
	Email        string `json:"email" binding:"required,email"`
	Senha        string `json:"senha" binding:"required,min=6"`
}

// RefreshTokenRequest representa a requisição de refresh do token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// JwtResponse representa a resposta com tokens JWT
type JwtResponse struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	TokenType    string    `json:"tokenType"`
}

// UsuarioResponse representa os dados do usuário na resposta
type UsuarioResponse struct {
	ID                string     `json:"id"`
	NomeCompleto      string     `json:"nomeCompleto"`
	Email             string     `json:"email"`
	TrialAtivo        bool       `json:"trialAtivo"`
	DataTrialInicio   *time.Time `json:"dataTrialInicio,omitempty"`
	TrialExpirado     bool       `json:"trialExpirado"`
	AssinaturaAtiva   bool       `json:"assinaturaAtiva"`
	DataCriacao       time.Time  `json:"dataCriacao"`
	TwoFactorHabilitado bool     `json:"twoFactorHabilitado"`
}

// LoginResponse representa a resposta completa de login
type LoginResponse struct {
	Usuario UsuarioResponse `json:"usuario"`
	Token   JwtResponse     `json:"token"`
}

// Ativar2FARequest representa a requisição para ativar 2FA
type Ativar2FARequest struct {
	Codigo string `json:"codigo" binding:"required,len=6"`
}

// Verificar2FARequest representa a verificação de código 2FA no login
type Verificar2FARequest struct {
	Email  string `json:"email" binding:"required,email"`
	Codigo string `json:"codigo" binding:"required,len=6"`
}

// GerarQRCode2FAResponse representa a resposta com QR code para 2FA
type GerarQRCode2FAResponse struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qrCodeUrl"`
}
