package servico

import (
	"github.com/ifinu/ifinu-api-go/repositorio"
)

type AssinaturaServico struct {
	assinaturaRepo *repositorio.AssinaturaRepositorio
	usuarioRepo    *repositorio.UsuarioRepositorio
}

func NovoAssinaturaServico(assinaturaRepo *repositorio.AssinaturaRepositorio, usuarioRepo *repositorio.UsuarioRepositorio) *AssinaturaServico {
	return &AssinaturaServico{
		assinaturaRepo: assinaturaRepo,
		usuarioRepo:    usuarioRepo,
	}
}

// ObterStatus retorna o status da assinatura do usuário
func (s *AssinaturaServico) ObterStatus(email string) (map[string]interface{}, error) {
	usuario, err := s.usuarioRepo.BuscarPorEmail(email)
	if err != nil {
		return nil, err
	}

	// Buscar assinatura do usuário
	assinatura, err := s.assinaturaRepo.BuscarPorUsuario(usuario.ID)

	status := map[string]interface{}{
		"vitalicio":       usuario.Vitalicio,
		"trialAtivo":      usuario.TrialAtivo,
		"trialExpirado":   usuario.IsTrialExpirado(),
		"diasRestantes":   usuario.DiasRestantesTrial(),
		"assinaturaAtiva": false,
	}

	// Se encontrou assinatura, adicionar informações
	if err == nil && assinatura != nil {
		status["assinaturaAtiva"] = assinatura.IsAtiva()
		status["status"] = assinatura.Status
		status["dataProximaCobranca"] = assinatura.DataProximaCobranca
		status["valorMensal"] = assinatura.ValorMensal
		status["planoAssinatura"] = assinatura.PlanoAssinatura
		status["dataCancelamento"] = assinatura.DataCancelamento
	}

	return status, nil
}

// CancelarAssinatura cancela a assinatura do usuário
func (s *AssinaturaServico) CancelarAssinatura(email string) error {
	usuario, err := s.usuarioRepo.BuscarPorEmail(email)
	if err != nil {
		return err
	}

	// Buscar assinatura do usuário
	assinatura, err := s.assinaturaRepo.BuscarPorUsuario(usuario.ID)
	if err != nil {
		return err
	}

	// Cancelar assinatura
	assinatura.Cancelar()

	// Salvar alterações
	return s.assinaturaRepo.Atualizar(assinatura)
}
