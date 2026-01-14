package dto

import (
	"time"

	"github.com/ifinu/ifinu-api-go/dominio/entidades"
	"github.com/ifinu/ifinu-api-go/dominio/enums"
)

// CheckoutAssinaturaRequest representa a requisição de checkout
type CheckoutAssinaturaRequest struct {
	PlanoAssinatura enums.PlanoAssinatura `json:"planoAssinatura" binding:"required"`
	SuccessURL      string                `json:"successUrl" binding:"required"`
	CancelURL       string                `json:"cancelUrl" binding:"required"`
}

// CheckoutAssinaturaResponse representa a resposta do checkout
type CheckoutAssinaturaResponse struct {
	CheckoutURL string  `json:"checkoutUrl"`
	SessionID   string  `json:"sessionId"`
	Valor       float64 `json:"valor"`
	Plano       string  `json:"plano"`
}

// AssinaturaResponse representa a assinatura na resposta
type AssinaturaResponse struct {
	ID                        string                       `json:"id"`
	UsuarioID                 string                       `json:"usuarioId"`
	Status                    entidades.StatusAssinatura   `json:"status"`
	PlanoAssinatura           enums.PlanoAssinatura        `json:"planoAssinatura"`
	PlanoDescricao            string                       `json:"planoDescricao"`
	ValorPlano                float64                      `json:"valorPlano"`
	IntervaloMeses            int                          `json:"intervaloMeses"`
	DataInicioPeriodoGratuito *time.Time                   `json:"dataInicioPeriodoGratuito,omitempty"`
	DataFimPeriodoGratuito    *time.Time                   `json:"dataFimPeriodoGratuito,omitempty"`
	DataProximaCobranca       *time.Time                   `json:"dataProximaCobranca,omitempty"`
	DataUltimaCobranca        *time.Time                   `json:"dataUltimaCobranca,omitempty"`
	DiasRestantesTrial        int                          `json:"diasRestantesTrial"`
	TrialAtivo                bool                         `json:"trialAtivo"`
	DataCriacao               time.Time                    `json:"dataCriacao"`
}

// ListarPlanosResponse representa a lista de planos disponíveis
type ListarPlanosResponse struct {
	Planos []PlanoInfo `json:"planos"`
}

// PlanoInfo representa informações de um plano
type PlanoInfo struct {
	Tipo              enums.PlanoAssinatura `json:"tipo"`
	Nome              string                `json:"nome"`
	Descricao         string                `json:"descricao"`
	Valor             float64               `json:"valor"`
	ValorMensal       float64               `json:"valorMensal"`
	IntervaloMeses    int                   `json:"intervaloMeses"`
	PercentualDesconto int                  `json:"percentualDesconto"`
	Recomendado       bool                  `json:"recomendado"`
}
