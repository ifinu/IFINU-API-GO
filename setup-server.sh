#!/bin/bash

# Script de setup do servidor para IFINU API GO
# Execute este script no servidor: bash setup-server.sh

set -e

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║         SETUP IFINU API GO - Migração Java → Go         ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Função para print colorido
print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 1. Parar Java Spring Boot
print_step "Parando backend Java Spring Boot..."
docker stop ifinu-api-java 2>/dev/null || print_warning "Container ifinu-api-java não encontrado"
docker rm ifinu-api-java 2>/dev/null || true
docker stop ifinu-api 2>/dev/null || print_warning "Container ifinu-api não encontrado"
docker rm ifinu-api 2>/dev/null || true

# Parar via docker-compose se existir
if [ -d "/home/mpx/ifinu-stack" ]; then
    print_step "Parando stack completa..."
    cd /home/mpx/ifinu-stack
    docker-compose stop api 2>/dev/null || true
fi

print_success "Backend Java parado"

# 2. Liberar porta 8080
print_step "Liberando porta 8080..."
lsof -t -i:8080 | xargs kill -9 2>/dev/null || print_warning "Nenhum processo na porta 8080"
print_success "Porta 8080 liberada"

# 3. Clonar/Atualizar repositório Go
print_step "Configurando repositório Go..."
if [ ! -d "/home/mpx/ifinu-api-go" ]; then
    print_step "Clonando repositório..."
    cd /home/mpx
    git clone https://github.com/ifinu/IFINU-API-GO.git ifinu-api-go
    print_success "Repositório clonado"
else
    print_step "Atualizando repositório existente..."
    cd /home/mpx/ifinu-api-go
    git fetch origin
    git reset --hard origin/main
    print_success "Repositório atualizado"
fi

cd /home/mpx/ifinu-api-go

# 4. Configurar .env
print_step "Verificando arquivo .env..."
if [ ! -f ".env" ]; then
    print_warning ".env não encontrado, criando do exemplo..."
    cp .env.example .env
    print_warning "⚠️  IMPORTANTE: Edite o arquivo .env com as credenciais corretas!"
    print_warning "Execute: nano /home/mpx/ifinu-api-go/.env"
    read -p "Pressione ENTER após configurar o .env ou CTRL+C para cancelar..."
else
    print_success ".env já existe"
fi

# 5. Build da imagem Docker
print_step "Buildando imagem Docker Go..."
docker build -t ifinu-api-go:latest .
print_success "Imagem buildada"

# 6. Limpar imagens antigas
print_step "Limpando imagens Docker antigas..."
docker image prune -f
print_success "Imagens antigas removidas"

# 7. Executar container Go
print_step "Iniciando container IFINU API GO..."
docker run -d \
  -p 8080:8080 \
  --name ifinu-api \
  --restart unless-stopped \
  --env-file .env \
  ifinu-api-go:latest

print_success "Container iniciado"

# 8. Aguardar inicialização
print_step "Aguardando inicialização..."
sleep 5

# 9. Verificar status
print_step "Verificando status do container..."
docker ps | grep ifinu-api

# 10. Health check
print_step "Testando health check..."
HEALTH_RESPONSE=$(curl -s http://localhost:8080/health)
if echo "$HEALTH_RESPONSE" | grep -q "ok"; then
    print_success "Health check passou! ✅"
    echo "$HEALTH_RESPONSE" | jq '.' 2>/dev/null || echo "$HEALTH_RESPONSE"
else
    print_error "Health check falhou! ❌"
    echo "Resposta: $HEALTH_RESPONSE"
    exit 1
fi

# 11. Mostrar logs
echo ""
print_step "Últimos logs do container:"
docker logs --tail 30 ifinu-api

# 12. Remover imagens Java antigas (opcional)
echo ""
read -p "Deseja remover imagens Docker do Java? (s/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Ss]$ ]]; then
    print_step "Removendo imagens Java..."
    docker images | grep ifinu | grep java | awk '{print $3}' | xargs docker rmi -f 2>/dev/null || true
    print_success "Imagens Java removidas"
fi

# Resumo final
echo ""
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║                  ✅ SETUP CONCLUÍDO                       ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""
print_success "API IFINU GO está rodando!"
echo ""
echo "📊 Informações:"
echo "  - URL: http://192.168.0.100:8080"
echo "  - Health: http://192.168.0.100:8080/health"
echo "  - Container: ifinu-api"
echo ""
echo "📋 Comandos úteis:"
echo "  - Ver logs: docker logs -f ifinu-api"
echo "  - Parar: docker stop ifinu-api"
echo "  - Reiniciar: docker restart ifinu-api"
echo "  - Status: docker ps | grep ifinu-api"
echo ""
echo "🔄 Próximo deploy:"
echo "  - Será automático via GitHub Actions"
echo "  - Configure os secrets no GitHub:"
echo "    - SSH_HOST: 192.168.0.100"
echo "    - SSH_USER: mpx"
echo "    - SSH_PASSWORD: Theo231023@"
echo ""
print_success "Migração Java → Go concluída! 🎉"
