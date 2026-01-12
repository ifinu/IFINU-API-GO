package util

import (
	"time"
)

// HorarioComercial representa a configuração de horário comercial
type HorarioComercial struct {
	HoraInicio    int // Hora de início (0-23)
	HoraFim       int // Hora de fim (0-23)
	DiasUteis     []time.Weekday
	FeriadosBR    map[string]bool // Mapa de feriados brasileiros (formato: "2006-01-02")
}

// HorarioComercialPadrao retorna configuração padrão de horário comercial brasileiro
// Segunda a Sexta: 8h às 18h
func HorarioComercialPadrao() *HorarioComercial {
	return &HorarioComercial{
		HoraInicio: 8,
		HoraFim:    18,
		DiasUteis: []time.Weekday{
			time.Monday,
			time.Tuesday,
			time.Wednesday,
			time.Thursday,
			time.Friday,
		},
		FeriadosBR: ObterFeriadosBrasileiros(),
	}
}

// EstaDentroHorarioComercial verifica se o horário atual está dentro do horário comercial
func (hc *HorarioComercial) EstaDentroHorarioComercial(agora time.Time) bool {
	// Ajustar para timezone de Brasília
	location, _ := time.LoadLocation("America/Sao_Paulo")
	agora = agora.In(location)

	// Verificar se é dia útil
	if !hc.isDiaUtil(agora.Weekday()) {
		return false
	}

	// Verificar se é feriado
	dataStr := agora.Format("2006-01-02")
	if hc.FeriadosBR[dataStr] {
		return false
	}

	// Verificar horário
	hora := agora.Hour()
	return hora >= hc.HoraInicio && hora < hc.HoraFim
}

// ProximoHorarioComercial retorna o próximo horário comercial disponível
func (hc *HorarioComercial) ProximoHorarioComercial(agora time.Time) time.Time {
	location, _ := time.LoadLocation("America/Sao_Paulo")
	agora = agora.In(location)

	// Se já estiver em horário comercial, retornar agora
	if hc.EstaDentroHorarioComercial(agora) {
		return agora
	}

	// Se for depois do horário comercial hoje, ir para início do próximo dia útil
	if agora.Hour() >= hc.HoraFim {
		agora = time.Date(agora.Year(), agora.Month(), agora.Day()+1, hc.HoraInicio, 0, 0, 0, location)
	} else {
		// Se for antes do horário comercial hoje, ir para início do horário hoje
		agora = time.Date(agora.Year(), agora.Month(), agora.Day(), hc.HoraInicio, 0, 0, 0, location)
	}

	// Avançar até encontrar um dia útil não feriado
	maxIteracoes := 30 // Evitar loop infinito
	for i := 0; i < maxIteracoes; i++ {
		if hc.isDiaUtil(agora.Weekday()) && !hc.FeriadosBR[agora.Format("2006-01-02")] {
			return agora
		}
		agora = agora.AddDate(0, 0, 1)
	}

	return agora
}

// isDiaUtil verifica se o dia da semana é dia útil
func (hc *HorarioComercial) isDiaUtil(dia time.Weekday) bool {
	for _, diaUtil := range hc.DiasUteis {
		if dia == diaUtil {
			return true
		}
	}
	return false
}

// ObterFeriadosBrasileiros retorna mapa com feriados nacionais brasileiros de 2026
func ObterFeriadosBrasileiros() map[string]bool {
	feriados := make(map[string]bool)

	// Feriados fixos nacionais 2026
	feriados["2026-01-01"] = true // Ano Novo
	feriados["2026-02-16"] = true // Carnaval
	feriados["2026-02-17"] = true // Carnaval
	feriados["2026-04-03"] = true // Sexta-feira Santa
	feriados["2026-04-21"] = true // Tiradentes
	feriados["2026-05-01"] = true // Dia do Trabalho
	feriados["2026-06-04"] = true // Corpus Christi
	feriados["2026-09-07"] = true // Independência do Brasil
	feriados["2026-10-12"] = true // Nossa Senhora Aparecida
	feriados["2026-11-02"] = true // Finados
	feriados["2026-11-15"] = true // Proclamação da República
	feriados["2026-12-25"] = true // Natal

	return feriados
}

// FormatarProximoHorario formata o próximo horário comercial para exibição
func (hc *HorarioComercial) FormatarProximoHorario(agora time.Time) string {
	proximo := hc.ProximoHorarioComercial(agora)
	location, _ := time.LoadLocation("America/Sao_Paulo")
	proximo = proximo.In(location)

	return proximo.Format("02/01/2006 às 15:04")
}
