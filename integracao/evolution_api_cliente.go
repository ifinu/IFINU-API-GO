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

type EvolutionAPICliente struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NovoEvolutionAPICliente() *EvolutionAPICliente {
	return &EvolutionAPICliente{
		baseURL: viper.GetString("EVOLUTION_API_URL"),
		apiKey:  viper.GetString("EVOLUTION_API_KEY"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CriarInstanciaRequest representa a requisição para criar uma instância
type CriarInstanciaRequest struct {
	InstanceName string `json:"instanceName"`
	Integration  string `json:"integration"`
	Token        string `json:"token,omitempty"`
	Qrcode       bool   `json:"qrcode"`
}

// CriarInstanciaResponse representa a resposta ao criar uma instância
type CriarInstanciaResponse struct {
	Instance struct {
		InstanceName string `json:"instanceName"`
		Status       string `json:"status"`
	} `json:"instance"`
	Hash struct {
		ApiKey string `json:"apikey"`
	} `json:"hash"`
	Qrcode struct {
		Code   string `json:"code"`
		Base64 string `json:"base64"`
	} `json:"qrcode"`
}

// StatusInstanciaResponse representa o status de uma instância
type StatusInstanciaResponse struct {
	Instance struct {
		InstanceName string `json:"instanceName"`
		State        string `json:"state"`
	} `json:"instance"`
	Base64 string `json:"base64,omitempty"`
}

// EnviarMensagemRequest representa a requisição de envio de mensagem
type EnviarMensagemRequest struct {
	Number      string            `json:"number"`
	Text        string            `json:"text,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
}

// EnviarMensagemResponse representa a resposta de envio de mensagem
type EnviarMensagemResponse struct {
	Key struct {
		RemoteJid string `json:"remoteJid"`
		FromMe    bool   `json:"fromMe"`
		ID        string `json:"id"`
	} `json:"key"`
	Message struct {
		Conversation string `json:"conversation"`
	} `json:"message"`
	MessageTimestamp string `json:"messageTimestamp"`
	Status           string `json:"status"`
}

// CriarInstancia cria uma nova instância no Evolution API
func (c *EvolutionAPICliente) CriarInstancia(nomeInstancia string) (*CriarInstanciaResponse, error) {
	url := fmt.Sprintf("%s/instance/create", c.baseURL)

	payload := CriarInstanciaRequest{
		InstanceName: nomeInstancia,
		Integration:  "WHATSAPP-BAILEYS",
		Qrcode:       true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro ao criar instância: %s - %s", resp.Status, string(bodyBytes))
	}

	var result CriarInstanciaResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ObterQRCode obtém o QR code de uma instância
func (c *EvolutionAPICliente) ObterQRCode(nomeInstancia string) (string, error) {
	url := fmt.Sprintf("%s/instance/connect/%s", c.baseURL, nomeInstancia)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("apikey", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("erro ao obter QR code")
	}

	var result StatusInstanciaResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	return result.Base64, nil
}

// ObterStatus obtém o status de uma instância
func (c *EvolutionAPICliente) ObterStatus(nomeInstancia string) (*StatusInstanciaResponse, error) {
	url := fmt.Sprintf("%s/instance/connectionState/%s", c.baseURL, nomeInstancia)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("erro ao obter status")
	}

	var result StatusInstanciaResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// EnviarMensagemTexto envia uma mensagem de texto
func (c *EvolutionAPICliente) EnviarMensagemTexto(nomeInstancia, telefone, mensagem string) (*EnviarMensagemResponse, error) {
	url := fmt.Sprintf("%s/message/sendText/%s", c.baseURL, nomeInstancia)

	payload := EnviarMensagemRequest{
		Number: telefone,
		Text:   mensagem,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro ao enviar mensagem: %s - %s", resp.Status, string(bodyBytes))
	}

	var result EnviarMensagemResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DeletarInstancia remove uma instância
func (c *EvolutionAPICliente) DeletarInstancia(nomeInstancia string) error {
	url := fmt.Sprintf("%s/instance/delete/%s", c.baseURL, nomeInstancia)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("apikey", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("erro ao deletar instância")
	}

	return nil
}

// Desconectar desconecta uma instância
func (c *EvolutionAPICliente) Desconectar(nomeInstancia string) error {
	url := fmt.Sprintf("%s/instance/logout/%s", c.baseURL, nomeInstancia)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("apikey", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("erro ao desconectar instância")
	}

	return nil
}
