package servico

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dominio/entidades"
	"github.com/ifinu/ifinu-api-go/dto"
	"github.com/ifinu/ifinu-api-go/repositorio"
	"github.com/ifinu/ifinu-api-go/util"
	"gorm.io/gorm"
)

type StripeConfigServico struct {
	stripeConfigRepo *repositorio.StripeConfigRepositorio
}

func NovoStripeConfigServico(stripeConfigRepo *repositorio.StripeConfigRepositorio) *StripeConfigServico {
	return &StripeConfigServico{
		stripeConfigRepo: stripeConfigRepo,
	}
}

func (s *StripeConfigServico) BuscarConfiguracao(usuarioID uuid.UUID) (*dto.StripeConfigResponse, error) {
	config, err := s.stripeConfigRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &dto.StripeConfigResponse{
				Configured: false,
				TestMode:   true,
			}, nil
		}
		return nil, err
	}

	return &dto.StripeConfigResponse{
		Configured:      true,
		PublishableKey:  config.PublishableKey,
		SecretKeyMasked: config.MaskSecretKey(),
		TestMode:        config.TestMode,
	}, nil
}

func (s *StripeConfigServico) SalvarConfiguracao(usuarioID uuid.UUID, req dto.SaveStripeConfigRequest) error {
	prefixoPublica := "pk_test_"
	prefixoSecreta := "sk_test_"

	if !req.TestMode {
		prefixoPublica = "pk_live_"
		prefixoSecreta = "sk_live_"
	}

	if !strings.HasPrefix(req.PublishableKey, prefixoPublica) {
		return errors.New("chave pública inválida para o modo selecionado")
	}

	existe, err := s.stripeConfigRepo.Existe(usuarioID)
	if err != nil {
		return err
	}

	if existe {
		configExistente, err := s.stripeConfigRepo.BuscarPorUsuario(usuarioID)
		if err != nil {
			return err
		}

		configExistente.PublishableKey = req.PublishableKey
		configExistente.TestMode = req.TestMode
		configExistente.DataAtualizacao = time.Now()

		if req.SecretKey != "" {
			if !strings.HasPrefix(req.SecretKey, prefixoSecreta) {
				return errors.New("chave secreta inválida para o modo selecionado")
			}

			secretKeyCriptografada, err := util.EncryptString(req.SecretKey)
			if err != nil {
				return errors.New("erro ao criptografar chave secreta")
			}
			configExistente.SecretKeyEncrypted = secretKeyCriptografada
		}

		return s.stripeConfigRepo.Atualizar(configExistente)
	}

	if req.SecretKey == "" {
		return errors.New("chave secreta é obrigatória")
	}

	if !strings.HasPrefix(req.SecretKey, prefixoSecreta) {
		return errors.New("chave secreta inválida para o modo selecionado")
	}

	secretKeyCriptografada, err := util.EncryptString(req.SecretKey)
	if err != nil {
		return errors.New("erro ao criptografar chave secreta")
	}

	novaConfig := &entidades.StripeConfig{
		UsuarioID:          usuarioID,
		PublishableKey:     req.PublishableKey,
		SecretKeyEncrypted: secretKeyCriptografada,
		TestMode:           req.TestMode,
		DataCriacao:        time.Now(),
		DataAtualizacao:    time.Now(),
	}

	return s.stripeConfigRepo.Criar(novaConfig)
}

func (s *StripeConfigServico) DeletarConfiguracao(usuarioID uuid.UUID) error {
	existe, err := s.stripeConfigRepo.Existe(usuarioID)
	if err != nil {
		return err
	}

	if !existe {
		return errors.New("configuração não encontrada")
	}

	return s.stripeConfigRepo.Deletar(usuarioID)
}

func (s *StripeConfigServico) ObterChaveSecreta(usuarioID uuid.UUID) (string, error) {
	config, err := s.stripeConfigRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("configuração Stripe não encontrada")
		}
		return "", err
	}

	chaveDescriptografada, err := util.DecryptString(config.SecretKeyEncrypted)
	if err != nil {
		return "", errors.New("erro ao descriptografar chave secreta")
	}

	return chaveDescriptografada, nil
}
