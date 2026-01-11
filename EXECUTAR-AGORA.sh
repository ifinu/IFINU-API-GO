#!/bin/bash

# Script para executar no servidor via SSH
# Execute: bash EXECUTAR-AGORA.sh
# Senha quando solicitado: Theo231023@

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║    DEPLOY IFINU API GO - Parando Java e Iniciando Go    ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""
echo "🔐 Senha SSH: Theo231023@"
echo ""
echo "🚀 Conectando no servidor..."
echo ""

ssh mpx@192.168.0.100 << 'ENDSSH'

set -e

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║         DEPLOY IFINU API GO - Parando Java               ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# 1. Parar Java Spring Boot
echo "🛑 [1/10] Parando backend Java..."
docker stop ifinu-api-java 2>/dev/null || echo "   Container ifinu-api-java não encontrado"
docker rm ifinu-api-java 2>/dev/null || true
docker stop ifinu-api 2>/dev/null || echo "   Container ifinu-api não encontrado"
docker rm ifinu-api 2>/dev/null || true
echo "✅ Java parado"

# 2. Liberar porta 8080
echo "⚔️  [2/10] Liberando porta 8080..."
lsof -t -i:8080 | xargs kill -9 2>/dev/null || echo "   Porta já livre"
echo "✅ Porta 8080 liberada"

# 3. Clonar/Atualizar repositório Go
echo "📦 [3/10] Configurando repositório Go..."
if [ ! -d "/home/mpx/ifinu-api-go" ]; then
    echo "   Clonando repositório..."
    cd /home/mpx
    git clone https://github.com/ifinu/IFINU-API-GO.git ifinu-api-go
else
    echo "   Atualizando repositório..."
    cd /home/mpx/ifinu-api-go
    git fetch origin
    git reset --hard origin/main
fi
cd /home/mpx/ifinu-api-go
echo "✅ Repositório atualizado"

# 4. Verificar .env
echo "🔧 [4/10] Verificando .env..."
if [ ! -f ".env" ]; then
    echo "⚠️  .env não encontrado, criando do exemplo..."
    cp .env.example .env
    echo "⚠️  AVISO: Você precisa configurar o .env manualmente!"
    echo "Execute: ssh mpx@192.168.0.100"
    echo "         nano /home/mpx/ifinu-api-go/.env"
else
    echo "✅ .env já existe"
fi

# 5. Build da imagem Docker
echo "🔨 [5/10] Buildando imagem Docker Go..."
docker build -t ifinu-api-go:latest . > /dev/null 2>&1
echo "✅ Imagem buildada"

# 6. Limpar imagens antigas
echo "🧹 [6/10] Limpando imagens antigas..."
docker image prune -f > /dev/null 2>&1
echo "✅ Limpeza concluída"

# 7. Executar container Go
echo "🚀 [7/10] Iniciando container IFINU API GO..."
docker run -d \
  -p 8080:8080 \
  --name ifinu-api \
  --restart unless-stopped \
  --env-file .env \
  ifinu-api-go:latest
echo "✅ Container iniciado"

# 8. Aguardar inicialização
echo "⏳ [8/10] Aguardando inicialização..."
sleep 5
echo "✅ Pronto"

# 9. Verificar status
echo "🔍 [9/10] Verificando status..."
docker ps | grep ifinu-api
echo "✅ Container rodando"

# 10. Health check
echo "💚 [10/10] Testando health check..."
HEALTH=$(curl -s http://localhost:8080/health)
if echo "$HEALTH" | grep -q "ok"; then
    echo "✅ Health check passou!"
    echo "   $HEALTH"
else
    echo "❌ Health check falhou!"
    echo "   Resposta: $HEALTH"
    exit 1
fi

# Logs
echo ""
echo "📋 Últimos logs:"
docker logs --tail 20 ifinu-api

# Resumo
echo ""
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║              ✅ DEPLOY CONCLUÍDO COM SUCESSO              ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""
echo "📊 Status:"
echo "   - API: http://192.168.0.100:8080"
echo "   - Health: http://192.168.0.100:8080/health"
echo "   - Container: ifinu-api"
echo ""
echo "🎉 API IFINU GO está rodando!"
echo "🗑️  Backend Java foi parado e removido"
echo ""

ENDSSH

echo ""
echo "✅ Deploy no servidor concluído!"
echo ""
echo "🔗 Acesse: http://192.168.0.100:8080/health"
echo ""
