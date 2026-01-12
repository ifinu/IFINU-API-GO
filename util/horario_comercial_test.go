package util

import (
	"testing"
	"time"
)

func TestEstaDentroHorarioComercial(t *testing.T) {
	hc := HorarioComercialPadrao()
	location, _ := time.LoadLocation("America/Sao_Paulo")

	tests := []struct {
		nome     string
		horario  time.Time
		esperado bool
	}{
		{
			nome:     "Dentro do horário - 10h segunda-feira",
			horario:  time.Date(2026, 1, 12, 10, 0, 0, 0, location), // Segunda
			esperado: true,
		},
		{
			nome:     "Fora do horário - 7h segunda-feira",
			horario:  time.Date(2026, 1, 12, 7, 0, 0, 0, location),
			esperado: false,
		},
		{
			nome:     "Fora do horário - 19h segunda-feira",
			horario:  time.Date(2026, 1, 12, 19, 0, 0, 0, location),
			esperado: false,
		},
		{
			nome:     "Fora do horário - Sábado 10h",
			horario:  time.Date(2026, 1, 17, 10, 0, 0, 0, location), // Sábado
			esperado: false,
		},
		{
			nome:     "Fora do horário - Domingo 10h",
			horario:  time.Date(2026, 1, 18, 10, 0, 0, 0, location), // Domingo
			esperado: false,
		},
		{
			nome:     "Limite inferior - 8h segunda-feira",
			horario:  time.Date(2026, 1, 12, 8, 0, 0, 0, location),
			esperado: true,
		},
		{
			nome:     "Limite superior - 18h segunda-feira",
			horario:  time.Date(2026, 1, 12, 18, 0, 0, 0, location),
			esperado: false,
		},
		{
			nome:     "Limite superior - 17h59 segunda-feira",
			horario:  time.Date(2026, 1, 12, 17, 59, 0, 0, location),
			esperado: true,
		},
		{
			nome:     "Feriado - Ano Novo 10h",
			horario:  time.Date(2026, 1, 1, 10, 0, 0, 0, location),
			esperado: false,
		},
		{
			nome:     "Feriado - Natal 10h",
			horario:  time.Date(2026, 12, 25, 10, 0, 0, 0, location),
			esperado: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			resultado := hc.EstaDentroHorarioComercial(tt.horario)
			if resultado != tt.esperado {
				t.Errorf("EstaDentroHorarioComercial() = %v, esperado %v para %s",
					resultado, tt.esperado, tt.horario.Format("02/01/2006 15:04 Monday"))
			}
		})
	}
}

func TestProximoHorarioComercial(t *testing.T) {
	hc := HorarioComercialPadrao()
	location, _ := time.LoadLocation("America/Sao_Paulo")

	tests := []struct {
		nome            string
		horario         time.Time
		esperadoHora    int
		esperadoDiaSem  time.Weekday
	}{
		{
			nome:           "Já em horário comercial - deve retornar mesmo horário",
			horario:        time.Date(2026, 1, 12, 10, 0, 0, 0, location), // Segunda 10h
			esperadoHora:   10,
			esperadoDiaSem: time.Monday,
		},
		{
			nome:           "Antes do horário - deve retornar início hoje",
			horario:        time.Date(2026, 1, 12, 7, 0, 0, 0, location), // Segunda 7h
			esperadoHora:   8,
			esperadoDiaSem: time.Monday,
		},
		{
			nome:           "Depois do horário - deve retornar início amanhã",
			horario:        time.Date(2026, 1, 12, 19, 0, 0, 0, location), // Segunda 19h
			esperadoHora:   8,
			esperadoDiaSem: time.Tuesday,
		},
		{
			nome:           "Sexta depois horário - deve pular fim de semana",
			horario:        time.Date(2026, 1, 16, 19, 0, 0, 0, location), // Sexta 19h
			esperadoHora:   8,
			esperadoDiaSem: time.Monday,
		},
		{
			nome:           "Sábado - deve retornar segunda-feira",
			horario:        time.Date(2026, 1, 17, 10, 0, 0, 0, location), // Sábado
			esperadoHora:   8,
			esperadoDiaSem: time.Monday,
		},
	}

	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			proximo := hc.ProximoHorarioComercial(tt.horario)
			if proximo.Hour() != tt.esperadoHora {
				t.Errorf("ProximoHorarioComercial().Hour() = %d, esperado %d",
					proximo.Hour(), tt.esperadoHora)
			}
			if proximo.Weekday() != tt.esperadoDiaSem {
				t.Errorf("ProximoHorarioComercial().Weekday() = %v, esperado %v",
					proximo.Weekday(), tt.esperadoDiaSem)
			}
		})
	}
}
