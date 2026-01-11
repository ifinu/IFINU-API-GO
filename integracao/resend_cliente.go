package integracao

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

type ResendCliente struct {
	apiKey string
	client *http.Client
}

func NovoResendCliente() *ResendCliente {
	return &ResendCliente{
		apiKey: viper.GetString("RESEND_API_KEY"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// EnviarEmailRequest representa a requisição de envio de email
type EnviarEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html,omitempty"`
	Text    string   `json:"text,omitempty"`
}

// EnviarEmailResponse representa a resposta de envio de email
type EnviarEmailResponse struct {
	ID string `json:"id"`
}

// EnviarEmail envia um email usando a API do Resend
func (c *ResendCliente) EnviarEmail(de, para, assunto, html, texto string) (string, error) {
	url := "https://api.resend.com/emails"

	payload := EnviarEmailRequest{
		From:    de,
		To:      []string{para},
		Subject: assunto,
		Html:    html,
		Text:    texto,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("erro ao enviar email: %s - %s", resp.Status, string(bodyBytes))
	}

	var result EnviarEmailResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	return result.ID, nil
}

// EnviarEmailCobranca envia email de notificação de cobrança
func (c *ResendCliente) EnviarEmailCobranca(para, nomeCliente, descricao string, valor float64, dataVencimento string) error {
	assunto := "Nova Cobrança - IFINU"

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Nova Cobrança</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2563eb;">Nova Cobrança Recebida</h2>
        <p>Olá, %s</p>
        <p>Você tem uma nova cobrança:</p>
        <div style="background-color: #f3f4f6; padding: 15px; border-radius: 8px; margin: 20px 0;">
            <p><strong>Descrição:</strong> %s</p>
            <p><strong>Valor:</strong> R$ %.2f</p>
            <p><strong>Vencimento:</strong> %s</p>
        </div>
        <p>Atenciosamente,<br>Equipe IFINU</p>
    </div>
</body>
</html>
	`, nomeCliente, descricao, valor, dataVencimento)

	texto := fmt.Sprintf("Nova Cobrança - Descrição: %s - Valor: R$ %.2f - Vencimento: %s",
		descricao, valor, dataVencimento)

	_, err := c.EnviarEmail("noreply@ifinu.io", para, assunto, html, texto)
	return err
}

// EnviarEmailLembrete envia email de lembrete de vencimento
func (c *ResendCliente) EnviarEmailLembrete(para, nomeCliente, descricao string, valor float64, dataVencimento string) error {
	assunto := "Lembrete: Cobrança vence em 3 dias - IFINU"

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Lembrete de Vencimento</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #f59e0b;">Lembrete de Vencimento</h2>
        <p>Olá, %s</p>
        <p>Sua cobrança vence em 3 dias:</p>
        <div style="background-color: #fef3c7; padding: 15px; border-radius: 8px; margin: 20px 0;">
            <p><strong>Descrição:</strong> %s</p>
            <p><strong>Valor:</strong> R$ %.2f</p>
            <p><strong>Vencimento:</strong> %s</p>
        </div>
        <p>Atenciosamente,<br>Equipe IFINU</p>
    </div>
</body>
</html>
	`, nomeCliente, descricao, valor, dataVencimento)

	texto := fmt.Sprintf("Lembrete: Cobrança vence em 3 dias - Descrição: %s - Valor: R$ %.2f - Vencimento: %s",
		descricao, valor, dataVencimento)

	_, err := c.EnviarEmail("noreply@ifinu.io", para, assunto, html, texto)
	return err
}

// EnviarEmailVencimento envia email de vencimento hoje
func (c *ResendCliente) EnviarEmailVencimento(para, nomeCliente, descricao string, valor float64) error {
	assunto := "Cobrança vence hoje - IFINU"

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Vencimento Hoje</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #dc2626;">Cobrança Vence Hoje</h2>
        <p>Olá, %s</p>
        <p>Sua cobrança vence hoje:</p>
        <div style="background-color: #fee2e2; padding: 15px; border-radius: 8px; margin: 20px 0;">
            <p><strong>Descrição:</strong> %s</p>
            <p><strong>Valor:</strong> R$ %.2f</p>
            <p><strong>Vencimento:</strong> HOJE</p>
        </div>
        <p>Atenciosamente,<br>Equipe IFINU</p>
    </div>
</body>
</html>
	`, nomeCliente, descricao, valor)

	texto := fmt.Sprintf("Cobrança vence HOJE - Descrição: %s - Valor: R$ %.2f", descricao, valor)

	_, err := c.EnviarEmail("noreply@ifinu.io", para, assunto, html, texto)
	return err
}

// ValidarConfiguracao verifica se a API Key está configurada
func (c *ResendCliente) ValidarConfiguracao() error {
	if c.apiKey == "" {
		return errors.New("RESEND_API_KEY não configurada")
	}
	return nil
}
