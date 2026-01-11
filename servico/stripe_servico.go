package servico

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/ifinu/ifinu-api-go/dto"
	"github.com/ifinu/ifinu-api-go/repositorio"
)

type StripeServico struct {
	usuarioRepo *repositorio.UsuarioRepositorio
}

func NovoStripeServico(usuarioRepo *repositorio.UsuarioRepositorio) *StripeServico {
	return &StripeServico{
		usuarioRepo: usuarioRepo,
	}
}

// CriarCheckoutTrial cria uma sessão de checkout Stripe para trial
func (s *StripeServico) CriarCheckoutTrial(usuarioID uuid.UUID) (*dto.CreateCheckoutResponse, error) {
	// Verificar se usuário existe
	_, err := s.usuarioRepo.BuscarPorID(usuarioID)
	if err != nil {
		return nil, err
	}

	// Por enquanto, retornar uma URL mockada
	// TODO: Implementar integração real com Stripe
	sessionID := fmt.Sprintf("cs_test_%s", uuid.New().String()[:8])
	checkoutURL := fmt.Sprintf("https://checkout.stripe.com/pay/%s", sessionID)

	return &dto.CreateCheckoutResponse{
		SessionID:   sessionID,
		CheckoutURL: checkoutURL,
	}, nil
}
