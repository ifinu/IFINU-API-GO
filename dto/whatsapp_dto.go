package dto

import "time"

// ConectarWhatsAppResponse representa a resposta de conexão do WhatsApp
type ConectarWhatsAppResponse struct {
	QRCode        string `json:"qrcode"`
	NomeInstancia string `json:"nomeInstancia"`
}

// StatusWhatsAppResponse representa o status da conexão WhatsApp
type StatusWhatsAppResponse struct {
	Conectado     bool      `json:"conectado"`
	NomeInstancia string    `json:"nomeInstancia,omitempty"`
	QRCode        string    `json:"qrcode,omitempty"`
	DataConexao   *time.Time `json:"dataConexao,omitempty"`
}

// EnviarMensagemRequest representa a requisição de envio de mensagem
type EnviarMensagemRequest struct {
	Telefone string `json:"telefone" binding:"required"`
	Mensagem string `json:"mensagem" binding:"required"`
}

// EnviarMensagemResponse representa a resposta de envio de mensagem
type EnviarMensagemResponse struct {
	Sucesso   bool   `json:"sucesso"`
	Mensagem  string `json:"mensagem"`
	MessageID string `json:"messageId,omitempty"`
}

// TestarConexaoResponse representa a resposta de teste de conexão
type TestarConexaoResponse struct {
	Sucesso  bool   `json:"sucesso"`
	Mensagem string `json:"mensagem"`
	Status   string `json:"status"`
}
