# Scripts de Setup

## 🎯 setup_stripe.go

Script automatizado para configurar produtos e prices no Stripe via API.

### O que faz

1. ✅ Cria produto "IFINU - Sistema de Cobrança" no Stripe
2. ✅ Cria 3 prices (Mensal, Trimestral, Anual)
3. ✅ Exibe os Price IDs para copiar no `.env`
4. ✅ Evita duplicação (verifica se produto já existe)

### Como executar

#### Opção 1: Via Make (Recomendado)

```bash
cd /Users/mikael/Documents/OfIIfinu/ifinu-api-go
make setup-stripe
```

#### Opção 2: Via Go Run

```bash
cd /Users/mikael/Documents/OfIIfinu/ifinu-api-go
go run scripts/setup_stripe.go
```

#### Opção 3: No Servidor (SSH)

```bash
# Fazer SSH
ssh mpx@192.168.0.100

# Ir para diretório da API
cd /caminho/para/ifinu-api-go

# Executar script
go run scripts/setup_stripe.go
```

### Pré-requisitos

1. **STRIPE_SECRET_KEY configurada**

Criar arquivo `.env` na raiz do projeto:

```bash
# .env
STRIPE_SECRET_KEY=sk_test_xxxxxxxxxxxxx
```

Ou exportar variável de ambiente:

```bash
export STRIPE_SECRET_KEY=sk_test_xxxxxxxxxxxxx
```

2. **Go instalado** (versão 1.21+)

### Output Esperado

```
╔════════════════════════════════════════════════════════╗
║      IFINU - Setup Automático do Stripe               ║
║      Criando Produtos e Prices                         ║
╚════════════════════════════════════════════════════════╝

🔍 Verificando produtos existentes...

📦 Criando produto IFINU...
✅ Produto criado: IFINU - Sistema de Cobrança (ID: prod_xxxxxxxxxxxxx)

💰 Criando prices para os 3 planos...

1️⃣  Criando Plano Mensal (R$ 39/mês)...
   ✅ Price Mensal criado: price_xxxxxxxxxxxxx

2️⃣  Criando Plano Trimestral (R$ 99 a cada 3 meses)...
   ✅ Price Trimestral criado: price_xxxxxxxxxxxxx

3️⃣  Criando Plano Anual (R$ 348/ano)...
   ✅ Price Anual criado: price_xxxxxxxxxxxxx

════════════════════════════════════════════════════════

📋 PRICES CRIADOS - Copie para o .env:

STRIPE_PRICE_ID_MENSAL=price_xxxxxxxxxxxxx
   → R$ 39.00/mês

STRIPE_PRICE_ID_TRIMESTRAL=price_xxxxxxxxxxxxx
   → R$ 99.00 a cada 3 meses (R$ 33.00/mês)

STRIPE_PRICE_ID_ANUAL=price_xxxxxxxxxxxxx
   → R$ 348.00/ano (R$ 29.00/mês)

════════════════════════════════════════════════════════

✅ Setup concluído com sucesso!

📝 Próximos passos:
   1. Copie as variáveis acima para o arquivo .env do servidor
   2. Reinicie a API: docker-compose restart api
   3. Configure o webhook em: https://dashboard.stripe.com/webhooks
      URL: https://api.ifinu.io/api/stripe/webhook
      Eventos: checkout.session.completed, customer.subscription.*
```

### Próximos Passos

#### 1. Copiar Price IDs para o `.env`

Adicione no arquivo `.env` do servidor:

```bash
STRIPE_SECRET_KEY=sk_live_xxxxxxxxxxxxx
STRIPE_PRICE_ID_MENSAL=price_xxxxxxxxxxxxx
STRIPE_PRICE_ID_TRIMESTRAL=price_xxxxxxxxxxxxx
STRIPE_PRICE_ID_ANUAL=price_xxxxxxxxxxxxx
```

#### 2. Reiniciar API

```bash
docker-compose restart api
```

#### 3. Configurar Webhook

Acesse https://dashboard.stripe.com/webhooks e adicione:

- **URL**: `https://api.ifinu.io/api/stripe/webhook`
- **Eventos**:
  - `checkout.session.completed`
  - `customer.subscription.created`
  - `customer.subscription.updated`
  - `customer.subscription.deleted`

### Troubleshooting

#### Erro: STRIPE_SECRET_KEY não configurada

```
❌ STRIPE_SECRET_KEY não configurada. Configure no arquivo .env
```

**Solução**: Criar arquivo `.env` com a chave do Stripe.

#### Erro: Price já existe

```
⚠️  Erro ao criar price mensal (pode já existir)
```

**Não é problema!** O script detecta prices existentes e exibe eles no final.

#### Verificar no Stripe Dashboard

Se tiver dúvidas, acesse:
- **Produtos**: https://dashboard.stripe.com/products
- **Prices**: Clique no produto criado para ver os prices

### Executar Novamente

O script é **idempotente**:
- Se produto já existe, não cria duplicado
- Se prices já existem, apenas lista os IDs
- Seguro executar múltiplas vezes

### Modo de Teste vs Produção

#### Teste (Desenvolvimento)

```bash
STRIPE_SECRET_KEY=sk_test_xxxxxxxxxxxxx
```

Cria produtos no modo **test** do Stripe.

#### Produção

```bash
STRIPE_SECRET_KEY=sk_live_xxxxxxxxxxxxx
```

Cria produtos no modo **live** do Stripe (cobranças reais).

⚠️ **ATENÇÃO**: Execute em PRODUÇÃO apenas quando estiver pronto para aceitar pagamentos reais!

### Ver Código

O script está em: `scripts/setup_stripe.go`

Usa a biblioteca oficial do Stripe:
- `github.com/stripe/stripe-go/v81`
