package enums

type PlanoAssinatura string

const (
	PlanoMensal     PlanoAssinatura = "MENSAL"
	PlanoTrimestral PlanoAssinatura = "TRIMESTRAL"
	PlanoAnual      PlanoAssinatura = "ANUAL"
)

// ObterValorPlano retorna o valor em reais de cada plano
func ObterValorPlano(plano PlanoAssinatura) float64 {
	switch plano {
	case PlanoMensal:
		return 39.00
	case PlanoTrimestral:
		return 99.00 // ~15% desconto (R$ 33/mês)
	case PlanoAnual:
		return 348.00 // ~25% desconto (R$ 29/mês)
	default:
		return 39.00
	}
}

// ObterIntervaloCobranca retorna quantos meses o plano cobre
func ObterIntervaloCobranca(plano PlanoAssinatura) int {
	switch plano {
	case PlanoMensal:
		return 1
	case PlanoTrimestral:
		return 3
	case PlanoAnual:
		return 12
	default:
		return 1
	}
}

// ObterDescricaoPlano retorna descrição humanizada do plano
func ObterDescricaoPlano(plano PlanoAssinatura) string {
	switch plano {
	case PlanoMensal:
		return "Plano Mensal - R$ 39/mês"
	case PlanoTrimestral:
		return "Plano Trimestral - R$ 99 (R$ 33/mês) - Economize 15%"
	case PlanoAnual:
		return "Plano Anual - R$ 348 (R$ 29/mês) - Economize 25%"
	default:
		return "Plano Mensal"
	}
}
