.PHONY: help setup-stripe run build test

help: ## Mostra este menu de ajuda
	@echo "╔════════════════════════════════════════════════════════╗"
	@echo "║           IFINU API GO - Comandos Make                 ║"
	@echo "╚════════════════════════════════════════════════════════╝"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

setup-stripe: ## Configura produtos e prices no Stripe automaticamente
	@echo "🚀 Executando setup do Stripe..."
	@go run scripts/setup_stripe.go

run: ## Inicia a API em modo desenvolvimento
	@echo "🚀 Iniciando API..."
	@go run cmd/api/main.go

build: ## Compila a API
	@echo "🔨 Compilando..."
	@go build -o bin/api cmd/api/main.go
	@echo "✅ API compilada em bin/api"

test: ## Executa testes
	@echo "🧪 Executando testes..."
	@go test ./... -v

clean: ## Remove arquivos de build
	@echo "🧹 Limpando..."
	@rm -rf bin/
	@echo "✅ Limpeza concluída"
