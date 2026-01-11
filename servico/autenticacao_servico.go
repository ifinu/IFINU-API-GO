package servico

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dominio/entidades"
	"github.com/ifinu/ifinu-api-go/dto"
	"github.com/ifinu/ifinu-api-go/repositorio"
	"github.com/ifinu/ifinu-api-go/util"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

type AutenticacaoServico struct {
	usuarioRepo *repositorio.UsuarioRepositorio
}

func NovoAutenticacaoServico(usuarioRepo *repositorio.UsuarioRepositorio) *AutenticacaoServico {
	return &AutenticacaoServico{
		usuarioRepo: usuarioRepo,
	}
}

// Login realiza o login do usuário
func (s *AutenticacaoServico) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	// Buscar usuário por email
	usuario, err := s.usuarioRepo.BuscarPorEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("email ou senha inválidos")
		}
		return nil, err
	}

	// Verificar senha
	if !util.VerificarSenha(req.Senha, usuario.SenhaHash) {
		return nil, errors.New("email ou senha inválidos")
	}

	// Se 2FA está habilitado, retornar erro específico
	if usuario.DuasEtapasAtivo {
		return nil, errors.New("2FA_NECESSARIO")
	}

	// Gerar tokens JWT
	accessToken, err := util.GerarToken(usuario.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := util.GerarRefreshToken(usuario.Email)
	if err != nil {
		return nil, err
	}

	// Atualizar último acesso
	agora := time.Now()
	usuario.DataUltimoAcesso = &agora
	s.usuarioRepo.Atualizar(usuario)

	// Montar resposta
	return &dto.LoginResponse{
		Usuario: s.mapearUsuarioParaDTO(usuario),
		Token: dto.JwtResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(24 * time.Hour), // 24 horas
			TokenType:    "Bearer",
		},
	}, nil
}

// Cadastro registra um novo usuário
func (s *AutenticacaoServico) Cadastro(req dto.CadastroRequest) (*dto.LoginResponse, error) {
	// Verificar se email já existe
	existe, err := s.usuarioRepo.ExistePorEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existe {
		return nil, errors.New("email já cadastrado")
	}

	// Hash da senha
	senhaHash, err := util.HashSenha(req.Senha)
	if err != nil {
		return nil, err
	}

	// Criar usuário
	agora := time.Now()
	usuario := &entidades.Usuario{
		ID:              uuid.New(),
		NomeCompleto:    req.NomeCompleto,
		Email:           req.Email,
		SenhaHash:       senhaHash,
		TrialAtivo:      true,
		DataTrialInicio: &agora,
		DataCriacao:     agora,
	}

	err = s.usuarioRepo.Criar(usuario)
	if err != nil {
		return nil, err
	}

	// Gerar tokens JWT
	accessToken, err := util.GerarToken(usuario.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := util.GerarRefreshToken(usuario.Email)
	if err != nil {
		return nil, err
	}

	// Montar resposta
	return &dto.LoginResponse{
		Usuario: s.mapearUsuarioParaDTO(usuario),
		Token: dto.JwtResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			TokenType:    "Bearer",
		},
	}, nil
}

// RefreshToken renova o access token
func (s *AutenticacaoServico) RefreshToken(req dto.RefreshTokenRequest) (*dto.JwtResponse, error) {
	// Validar refresh token
	claims, err := util.ValidarToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("refresh token inválido ou expirado")
	}

	// Gerar novo access token
	accessToken, err := util.GerarToken(claims.Email)
	if err != nil {
		return nil, err
	}

	// Gerar novo refresh token
	refreshToken, err := util.GerarRefreshToken(claims.Email)
	if err != nil {
		return nil, err
	}

	return &dto.JwtResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		TokenType:    "Bearer",
	}, nil
}

// BuscarUsuarioPorEmail retorna os dados do usuário autenticado
func (s *AutenticacaoServico) BuscarUsuarioPorEmail(email string) (*dto.UsuarioResponse, error) {
	usuario, err := s.usuarioRepo.BuscarPorEmail(email)
	if err != nil {
		return nil, err
	}

	usuarioDTO := s.mapearUsuarioParaDTO(usuario)
	return &usuarioDTO, nil
}

// Gerar2FA gera o QR code para configurar 2FA
func (s *AutenticacaoServico) Gerar2FA(email string) (*dto.GerarQRCode2FAResponse, error) {
	usuario, err := s.usuarioRepo.BuscarPorEmail(email)
	if err != nil {
		return nil, err
	}

	// Gerar secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "IFINU",
		AccountName: usuario.Email,
	})
	if err != nil {
		return nil, err
	}

	// Salvar secret no usuário
	usuario.DuasEtapasSecret = key.Secret()
	err = s.usuarioRepo.Atualizar(usuario)
	if err != nil {
		return nil, err
	}

	return &dto.GerarQRCode2FAResponse{
		Secret:    key.Secret(),
		QRCodeURL: key.URL(),
	}, nil
}

// Ativar2FA ativa o 2FA após validar o código
func (s *AutenticacaoServico) Ativar2FA(email string, codigo string) error {
	usuario, err := s.usuarioRepo.BuscarPorEmail(email)
	if err != nil {
		return err
	}

	// Validar código
	valido := totp.Validate(codigo, usuario.DuasEtapasSecret)
	if !valido {
		return errors.New("código inválido")
	}

	// Ativar 2FA
	usuario.DuasEtapasAtivo = true
	return s.usuarioRepo.Atualizar(usuario)
}

// Verificar2FA valida o código 2FA no login
func (s *AutenticacaoServico) Verificar2FA(req dto.Verificar2FARequest) (*dto.LoginResponse, error) {
	usuario, err := s.usuarioRepo.BuscarPorEmail(req.Email)
	if err != nil {
		return nil, err
	}

	// Validar código
	valido := totp.Validate(req.Codigo, usuario.DuasEtapasSecret)
	if !valido {
		return nil, errors.New("código inválido")
	}

	// Gerar tokens JWT
	accessToken, err := util.GerarToken(usuario.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := util.GerarRefreshToken(usuario.Email)
	if err != nil {
		return nil, err
	}

	// Atualizar último acesso
	agora := time.Now()
	usuario.DataUltimoAcesso = &agora
	s.usuarioRepo.Atualizar(usuario)

	return &dto.LoginResponse{
		Usuario: s.mapearUsuarioParaDTO(usuario),
		Token: dto.JwtResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
			TokenType:    "Bearer",
		},
	}, nil
}

// ObterStatusTrial retorna o status do trial do usuário
func (s *AutenticacaoServico) ObterStatusTrial(email string) (map[string]interface{}, error) {
	usuario, err := s.usuarioRepo.BuscarPorEmail(email)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"trialAtivo":      usuario.TrialAtivo,
		"trialExpirado":   usuario.IsTrialExpirado(),
		"diasRestantes":   usuario.DiasRestantesTrial(),
		"dataTrialInicio": usuario.DataTrialInicio,
		"vitalicio":       usuario.Vitalicio,
	}, nil
}

// mapearUsuarioParaDTO converte Usuario para UsuarioResponse
func (s *AutenticacaoServico) mapearUsuarioParaDTO(usuario *entidades.Usuario) dto.UsuarioResponse {
	return dto.UsuarioResponse{
		ID:                  usuario.ID.String(),
		NomeCompleto:        usuario.NomeCompleto,
		Email:               usuario.Email,
		TrialAtivo:          usuario.TrialAtivo,
		DataTrialInicio:     usuario.DataTrialInicio,
		TrialExpirado:       usuario.IsTrialExpirado(),
		AssinaturaAtiva:     false, // TODO: verificar assinatura ativa
		DataCriacao:         usuario.DataCriacao,
		TwoFactorHabilitado: usuario.DuasEtapasAtivo,
	}
}
