#!/bin/bash

# Script para configurar produtos e prices no Stripe via API
# Uso: ./setup_stripe.sh

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "╔════════════════════════════════════════════════════════╗"
echo "║      IFINU - Setup Automático do Stripe               ║"
echo "║      Criando Produtos e Prices via API                ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# Verificar se STRIPE_SECRET_KEY está configurada
if [ -z "$STRIPE_SECRET_KEY" ]; then
    # Tentar carregar do .env
    if [ -f .env ]; then
        export $(grep -v '^#' .env | grep STRIPE_SECRET_KEY | xargs)
    fi

    if [ -z "$STRIPE_SECRET_KEY" ]; then
        echo -e "${RED}❌ STRIPE_SECRET_KEY não configurada${NC}"
        echo "Configure a variável de ambiente ou crie arquivo .env"
        exit 1
    fi
fi

echo -e "${BLUE}🔍 Verificando produtos existentes...${NC}"

# Listar produtos existentes
PRODUCTS=$(curl -s https://api.stripe.com/v1/products \
  -u "${STRIPE_SECRET_KEY}:" \
  -d "active=true" \
  -d "limit=100")

# Verificar se produto IFINU já existe
PRODUCT_ID=$(echo "$PRODUCTS" | grep -o '"id":"prod_[^"]*"' | grep -A 10 "IFINU" | head -1 | cut -d'"' -f4)

if [ -n "$PRODUCT_ID" ]; then
    echo -e "${GREEN}✅ Produto já existe: IFINU - Sistema de Cobrança (ID: $PRODUCT_ID)${NC}"
else
    echo -e "${BLUE}📦 Criando produto IFINU...${NC}"

    # Criar produto
    PRODUCT_RESPONSE=$(curl -s https://api.stripe.com/v1/products \
      -u "${STRIPE_SECRET_KEY}:" \
      -d "name=IFINU - Sistema de Cobrança" \
      -d "description=Plataforma de automação de cobranças com WhatsApp e Email")

    PRODUCT_ID=$(echo "$PRODUCT_RESPONSE" | grep -o '"id":"prod_[^"]*"' | head -1 | cut -d'"' -f4)

    if [ -z "$PRODUCT_ID" ]; then
        echo -e "${RED}❌ Erro ao criar produto${NC}"
        echo "$PRODUCT_RESPONSE"
        exit 1
    fi

    echo -e "${GREEN}✅ Produto criado: $PRODUCT_ID${NC}"
fi

echo ""
echo -e "${BLUE}💰 Criando prices para os 3 planos...${NC}"
echo ""

# Criar Price Mensal
echo -e "${YELLOW}1️⃣  Criando Plano Mensal (R$ 39/mês)...${NC}"
PRICE_MENSAL=$(curl -s https://api.stripe.com/v1/prices \
  -u "${STRIPE_SECRET_KEY}:" \
  -d "product=$PRODUCT_ID" \
  -d "unit_amount=3900" \
  -d "currency=brl" \
  -d "recurring[interval]=month" \
  -d "nickname=Plano Mensal")

PRICE_MENSAL_ID=$(echo "$PRICE_MENSAL" | grep -o '"id":"price_[^"]*"' | head -1 | cut -d'"' -f4)

if [ -n "$PRICE_MENSAL_ID" ]; then
    echo -e "   ${GREEN}✅ Price Mensal criado: $PRICE_MENSAL_ID${NC}"
else
    echo -e "   ${YELLOW}⚠️  Pode já existir ou erro ao criar${NC}"
fi

# Criar Price Trimestral
echo ""
echo -e "${YELLOW}2️⃣  Criando Plano Trimestral (R$ 99 a cada 3 meses)...${NC}"
PRICE_TRIMESTRAL=$(curl -s https://api.stripe.com/v1/prices \
  -u "${STRIPE_SECRET_KEY}:" \
  -d "product=$PRODUCT_ID" \
  -d "unit_amount=9900" \
  -d "currency=brl" \
  -d "recurring[interval]=month" \
  -d "recurring[interval_count]=3" \
  -d "nickname=Plano Trimestral")

PRICE_TRIMESTRAL_ID=$(echo "$PRICE_TRIMESTRAL" | grep -o '"id":"price_[^"]*"' | head -1 | cut -d'"' -f4)

if [ -n "$PRICE_TRIMESTRAL_ID" ]; then
    echo -e "   ${GREEN}✅ Price Trimestral criado: $PRICE_TRIMESTRAL_ID${NC}"
else
    echo -e "   ${YELLOW}⚠️  Pode já existir ou erro ao criar${NC}"
fi

# Criar Price Anual
echo ""
echo -e "${YELLOW}3️⃣  Criando Plano Anual (R$ 348/ano)...${NC}"
PRICE_ANUAL=$(curl -s https://api.stripe.com/v1/prices \
  -u "${STRIPE_SECRET_KEY}:" \
  -d "product=$PRODUCT_ID" \
  -d "unit_amount=34800" \
  -d "currency=brl" \
  -d "recurring[interval]=year" \
  -d "nickname=Plano Anual")

PRICE_ANUAL_ID=$(echo "$PRICE_ANUAL" | grep -o '"id":"price_[^"]*"' | head -1 | cut -d'"' -f4)

if [ -n "$PRICE_ANUAL_ID" ]; then
    echo -e "   ${GREEN}✅ Price Anual criado: $PRICE_ANUAL_ID${NC}"
else
    echo -e "   ${YELLOW}⚠️  Pode já existir ou erro ao criar${NC}"
fi

# Listar todos os prices do produto
echo ""
echo "════════════════════════════════════════════════════════"
echo ""
echo -e "${BLUE}📋 PRICES DO PRODUTO - Copie para o .env:${NC}"
echo ""

PRICES_LIST=$(curl -s https://api.stripe.com/v1/prices \
  -u "${STRIPE_SECRET_KEY}:" \
  -d "product=$PRODUCT_ID" \
  -d "active=true" \
  -d "limit=100")

# Extrair e exibir prices
echo "$PRICES_LIST" | grep -o '"id":"price_[^"]*"' | while read -r line; do
    PRICE_ID=$(echo "$line" | cut -d'"' -f4)

    # Buscar detalhes do price
    PRICE_DETAILS=$(curl -s https://api.stripe.com/v1/prices/$PRICE_ID \
      -u "${STRIPE_SECRET_KEY}:")

    AMOUNT=$(echo "$PRICE_DETAILS" | grep -o '"unit_amount":[0-9]*' | cut -d':' -f2)
    INTERVAL=$(echo "$PRICE_DETAILS" | grep -o '"interval":"[^"]*"' | cut -d'"' -f4)
    INTERVAL_COUNT=$(echo "$PRICE_DETAILS" | grep -o '"interval_count":[0-9]*' | cut -d':' -f2)

    if [ -z "$INTERVAL_COUNT" ]; then
        INTERVAL_COUNT=1
    fi

    VALOR=$(echo "scale=2; $AMOUNT / 100" | bc)

    if [ "$INTERVAL" = "month" ] && [ "$INTERVAL_COUNT" = "1" ]; then
        echo -e "${GREEN}STRIPE_PRICE_ID_MENSAL=$PRICE_ID${NC}"
        echo "   → R$ $VALOR/mês"
        echo ""
    elif [ "$INTERVAL" = "month" ] && [ "$INTERVAL_COUNT" = "3" ]; then
        VALOR_MES=$(echo "scale=2; $VALOR / 3" | bc)
        echo -e "${GREEN}STRIPE_PRICE_ID_TRIMESTRAL=$PRICE_ID${NC}"
        echo "   → R$ $VALOR a cada 3 meses (R$ $VALOR_MES/mês)"
        echo ""
    elif [ "$INTERVAL" = "year" ]; then
        VALOR_MES=$(echo "scale=2; $VALOR / 12" | bc)
        echo -e "${GREEN}STRIPE_PRICE_ID_ANUAL=$PRICE_ID${NC}"
        echo "   → R$ $VALOR/ano (R$ $VALOR_MES/mês)"
        echo ""
    fi
done

echo "════════════════════════════════════════════════════════"
echo ""
echo -e "${GREEN}✅ Setup concluído com sucesso!${NC}"
echo ""
echo -e "${BLUE}📝 Próximos passos:${NC}"
echo "   1. Copie as variáveis acima para o arquivo .env do servidor"
echo "   2. Reinicie a API: docker-compose restart api"
echo "   3. Configure o webhook em: https://dashboard.stripe.com/webhooks"
echo "      URL: https://api.ifinu.io/api/stripe/webhook"
echo "      Eventos: checkout.session.completed, customer.subscription.*"
echo ""
