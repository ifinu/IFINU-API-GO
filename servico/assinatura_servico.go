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
		status["plano"] = assinatura.Plano
		status["status"] = assinatura.Status
		status["dataProximoPagamento"] = assinatura.DataProximoPagamento
		status["valorMensal"] = assinatura.ValorMensal
	}

	return status, nil
}
