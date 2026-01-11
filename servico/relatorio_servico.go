package servico

import (
	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/repositorio"
)

type RelatorioServico struct {
	clienteRepo  *repositorio.ClienteRepositorio
	cobrancaRepo *repositorio.CobrancaRepositorio
}

func NovoRelatorioServico(clienteRepo *repositorio.ClienteRepositorio, cobrancaRepo *repositorio.CobrancaRepositorio) *RelatorioServico {
	return &RelatorioServico{
		clienteRepo:  clienteRepo,
		cobrancaRepo: cobrancaRepo,
	}
}

// ObterDashboard retorna estatísticas do dashboard
func (s *RelatorioServico) ObterDashboard(usuarioID uuid.UUID) (map[string]interface{}, error) {
	// Buscar todas as cobranças do usuário
	cobrancas, err := s.cobrancaRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		return nil, err
	}

	// Buscar todos os clientes
	clientes, err := s.clienteRepo.BuscarPorUsuario(usuarioID)
	if err != nil {
		return nil, err
	}

	// Calcular estatísticas
	totalCobrancas := len(cobrancas)
	totalClientes := len(clientes)

	var totalRecebido, totalPendente, totalVencido float64
	var pagas, pendentes, vencidas int

	for _, cobranca := range cobrancas {
		switch cobranca.Status {
		case "PAGA":
			totalRecebido += cobranca.Valor
			pagas++
		case "PENDENTE":
			totalPendente += cobranca.Valor
			pendentes++
		case "VENCIDA":
			totalVencido += cobranca.Valor
			vencidas++
		}
	}

	return map[string]interface{}{
		"totalClientes":   totalClientes,
		"totalCobrancas":  totalCobrancas,
		"totalRecebido":   totalRecebido,
		"totalPendente":   totalPendente,
		"totalVencido":    totalVencido,
		"cobrancasPagas":  pagas,
		"cobrancasPendentes": pendentes,
		"cobrancasVencidas": vencidas,
	}, nil
}
