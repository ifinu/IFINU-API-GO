#!/bin/bash

# Script para deploy do Stripe Connect no servidor
# Uso: ./deploy_stripe_connect.sh

set -e

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "════════════════════════════════════════════════════════"
echo "  IFINU - Deploy do Stripe Connect"
echo "════════════════════════════════════════════════════════"
echo ""

echo -e "${BLUE}📦 Passo 1: Pull do código${NC}"
cd /home/mpx/ifinu-stack/api
git pull origin main
echo -e "${GREEN}✅ Código atualizado${NC}"
echo ""

echo -e "${BLUE}🗄️  Passo 2: Executar migration${NC}"
sudo -u postgres psql ifinu_db -f migrations/010_add_stripe_connect_fields.sql
echo -e "${GREEN}✅ Migration executada${NC}"
echo ""

echo -e "${BLUE}🔨 Passo 3: Rebuild da API${NC}"
cd /home/mpx/ifinu-stack
docker-compose build api
echo -e "${GREEN}✅ API reconstruída${NC}"
echo ""

echo -e "${BLUE}🔄 Passo 4: Restart do container${NC}"
docker-compose restart api
sleep 5
echo -e "${GREEN}✅ Container reiniciado${NC}"
echo ""

echo -e "${BLUE}📊 Passo 5: Verificar logs${NC}"
docker logs --tail 50 ifinu-api
echo ""

echo -e "${BLUE}🧪 Passo 6: Testar endpoints${NC}"
echo -e "${YELLOW}Testando health...${NC}"
curl -s http://localhost:8080/health | jq '.'
echo ""

echo "════════════════════════════════════════════════════════"
echo -e "${GREEN}✅ Deploy concluído com sucesso!${NC}"
echo "════════════════════════════════════════════════════════"
echo ""
echo -e "${BLUE}📝 Próximos passos:${NC}"
echo "1. Configurar webhook no Stripe:"
echo "   URL: https://api.ifinu.io/api/stripe-connect/webhook"
echo "   Evento: account.updated"
echo ""
echo "2. Testar onboarding:"
echo "   a) Login em: https://app.ifinu.io"
echo "   b) Ir para: Configurações > Pagamentos"
echo "   c) Clicar: Conectar com Stripe"
echo "   d) Completar onboarding"
echo ""
echo "3. Testar pagamento:"
echo "   a) Criar cliente"
echo "   b) Criar cobrança"
echo "   c) Pagar com cartão teste: 4242 4242 4242 4242"
echo "   d) Verificar no dashboard do usuário"
echo ""
