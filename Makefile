.PHONY: help build run test docker-build docker-run clean kill-8080 deploy

help:
	@echo "════════════════════════════════════════════════════"
	@echo "   IFINU API GO - Comandos Disponíveis"
	@echo "════════════════════════════════════════════════════"
	@echo "  make build         - Compilar binário Go"
	@echo "  make run           - Executar aplicação"
	@echo "  make test          - Rodar testes"
	@echo "  make docker-build  - Build imagem Docker"
	@echo "  make docker-run    - Executar container Docker"
	@echo "  make clean         - Limpar binários e cache"
	@echo "  make kill-8080     - Matar processo na porta 8080"
	@echo "  make deploy        - Deploy completo (kill + build + run)"
	@echo "  make mod           - Baixar dependências"
	@echo "════════════════════════════════════════════════════"

mod:
	@echo "📦 Baixando dependências..."
	go mod download
	go mod tidy

build:
	@echo "🔨 Compilando binário..."
	go build -o bin/ifinu-api ./cmd/api
	@echo "✅ Binário compilado em: bin/ifinu-api"

run:
	@echo "🚀 Executando aplicação..."
	go run ./cmd/api

test:
	@echo "🧪 Rodando testes..."
	go test -v ./...

docker-build:
	@echo "🐳 Buildando imagem Docker..."
	docker build -t ifinu-api-go:latest .
	@echo "✅ Imagem criada: ifinu-api-go:latest"

docker-run:
	@echo "🐳 Executando container..."
	docker run -d -p 8080:8080 --name ifinu-api ifinu-api-go:latest
	@echo "✅ Container rodando na porta 8080"

docker-stop:
	@echo "🛑 Parando container..."
	docker stop ifinu-api || true
	docker rm ifinu-api || true

clean:
	@echo "🧹 Limpando..."
	rm -rf bin/
	go clean -cache
	@echo "✅ Limpeza concluída"

kill-8080:
	@echo "⚔️  Matando processos na porta 8080..."
	lsof -t -i:8080 | xargs kill -9 || true
	@echo "✅ Porta 8080 liberada"

deploy: kill-8080 build
	@echo "🚀 Fazendo deploy..."
	./bin/ifinu-api

dev:
	@echo "🔧 Modo desenvolvimento (hot reload)..."
	@echo "⚠️  Instale o air: go install github.com/cosmtrek/air@latest"
	air
