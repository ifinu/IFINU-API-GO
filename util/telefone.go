package util

import (
	"regexp"
	"strings"
)

// FormatarTelefoneBrasileiro formata um número de telefone brasileiro para o formato internacional
// Aceita formatos: (77) 99861-6740, 77998616740, 5577998616740
// Retorna: 5577998616740
func FormatarTelefoneBrasileiro(telefone string) string {
	// Remover todos os caracteres não numéricos
	re := regexp.MustCompile(`[^0-9]`)
	apenasNumeros := re.ReplaceAllString(telefone, "")

	// Se estiver vazio, retornar vazio
	if apenasNumeros == "" {
		return ""
	}

	// Se já tem código do país (55) no início e tamanho correto (13 dígitos)
	if strings.HasPrefix(apenasNumeros, "55") && len(apenasNumeros) == 13 {
		return apenasNumeros
	}

	// Se tem 11 dígitos (DDD + número com 9 dígitos)
	if len(apenasNumeros) == 11 {
		return "55" + apenasNumeros
	}

	// Se tem 10 dígitos (DDD + número com 8 dígitos) - telefone fixo
	if len(apenasNumeros) == 10 {
		return "55" + apenasNumeros
	}

	// Casos especiais: remover 0 ou 15 na frente do DDD
	if strings.HasPrefix(apenasNumeros, "0") && len(apenasNumeros) == 12 {
		return "55" + apenasNumeros[1:]
	}

	// Se começar com 15 (operadora) e tiver 13 dígitos
	if strings.HasPrefix(apenasNumeros, "15") && len(apenasNumeros) == 13 {
		// Remover o 15 da operadora
		return "55" + apenasNumeros[2:]
	}

	// Retornar com 55 na frente de qualquer forma
	return "55" + apenasNumeros
}
