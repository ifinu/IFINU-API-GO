# Configuração de Assinaturas no Stripe

Este documento explica como configurar produtos e assinaturas recorrentes no Stripe Dashboard para o sistema IFINU.

## 🎯 Objetivo

Criar 3 planos de assinatura recorrente no Stripe com trial de 14 dias:
- **Mensal**: R$ 39/mês
- **Trimestral**: R$ 99 a cada 3 meses (R$ 33/mês)
- **Anual**: R$ 348 por ano (R$ 29/mês)

## 📋 Passo a Passo

### 1. Acessar o Stripe Dashboard

https://dashboard.stripe.com/products

### 2. Criar Produto Principal

1. Clique em **"+ Add product"**
2. Preencha:
   - **Name**: `IFINU - Sistema de Cobrança`
   - **Description**: `Plataforma de automação de cobranças com WhatsApp e Email`
   - **Image**: (opcional) Upload logo do IFINU

### 3. Criar Price para Plano Mensal

1. Na mesma tela do produto, em **"Pricing"**:
   - **Pricing model**: `Standard pricing`
   - **Price**: `39.00`
   - **Currency**: `BRL`
   - **Billing period**: `Monthly`
   - **Usage type**: `Licensed` (não metered)

2. Clique em **"Add price"**

3. **COPIE O PRICE ID** gerado (formato: `price_xxxxxxxxxxxxxx`)
   - Exemplo: `price_1ABCDEFGH123456`
   - Salve como `STRIPE_PRICE_ID_MENSAL`

### 4. Criar Price para Plano Trimestral

1. No mesmo produto, clique em **"Add another price"**:
   - **Pricing model**: `Standard pricing`
   - **Price**: `99.00`
   - **Currency**: `BRL`
   - **Billing period**: `Every 3 months`
   - **Usage type**: `Licensed`

2. Clique em **"Add price"**

3. **COPIE O PRICE ID** gerado
   - Salve como `STRIPE_PRICE_ID_TRIMESTRAL`

### 5. Criar Price para Plano Anual

1. No mesmo produto, clique em **"Add another price"**:
   - **Pricing model**: `Standard pricing`
   - **Price**: `348.00`
   - **Currency**: `BRL`
   - **Billing period**: `Yearly`
   - **Usage type**: `Licensed`

2. Clique em **"Add price"**

3. **COPIE O PRICE ID** gerado
   - Salve como `STRIPE_PRICE_ID_ANUAL`

### 6. Configurar Variáveis de Ambiente

No servidor, adicione as variáveis de ambiente no arquivo `.env`:

```bash
# Stripe Configuration
STRIPE_SECRET_KEY=sk_live_xxxxxxxxxxxxx
STRIPE_PRICE_ID_MENSAL=price_xxxxxxxxxxxxx
STRIPE_PRICE_ID_TRIMESTRAL=price_xxxxxxxxxxxxx
STRIPE_PRICE_ID_ANUAL=price_xxxxxxxxxxxxx
```

### 7. Configurar Webhook no Stripe

1. Acesse: https://dashboard.stripe.com/webhooks

2. Clique em **"Add endpoint"**

3. Configure:
   - **Endpoint URL**: `https://api.ifinu.io/api/stripe/webhook`
   - **Description**: `IFINU - Webhook de Assinaturas`
   - **Version**: `Latest API version`

4. Em **"Select events to listen to"**, adicione:
   - ✅ `checkout.session.completed`
   - ✅ `customer.subscription.created`
   - ✅ `customer.subscription.updated`
   - ✅ `customer.subscription.deleted`

5. Clique em **"Add endpoint"**

6. **COPIE O SIGNING SECRET** (formato: `whsec_xxxxxxxxxxxxx`)
   - Guarde para implementar validação de webhook (opcional mas recomendado)

### 8. Reiniciar API

Após configurar as variáveis de ambiente, reinicie a API:

```bash
docker-compose restart api
```

Ou via SSH no servidor:

```bash
cd /home/mpx/ifinu-stack
docker-compose restart api
```

## 🧪 Testar Configuração

### 1. Testar Listagem de Planos

```bash
curl -X GET https://api.ifinu.io/api/assinaturas/planos \
  -H "Authorization: Bearer {TOKEN}"
```

Deve retornar 3 planos.

### 2. Testar Criação de Checkout

```bash
curl -X POST https://api.ifinu.io/api/assinaturas/checkout \
  -H "Authorization: Bearer {TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "planoAssinatura": "MENSAL",
    "successUrl": "https://app.ifinu.io/painel/plano?sucesso=true",
    "cancelUrl": "https://app.ifinu.io/painel/plano?cancelado=true"
  }'
```

Deve retornar:
```json
{
  "success": true,
  "data": {
    "checkoutUrl": "https://checkout.stripe.com/...",
    "sessionId": "cs_test_...",
    "valor": 39,
    "plano": "MENSAL"
  }
}
```

### 3. Testar Pagamento Completo

1. Acesse a `checkoutUrl` retornada
2. Use cartão de teste do Stripe:
   - Número: `4242 4242 4242 4242`
   - Data: Qualquer data futura
   - CVC: Qualquer 3 dígitos
3. Complete o pagamento
4. Verifique no Stripe Dashboard:
   - Em **Customers** deve aparecer o cliente
   - Em **Subscriptions** deve aparecer a assinatura com status `trialing`

### 4. Verificar no Banco de Dados

```sql
SELECT
    id,
    usuario_id,
    status,
    plano_assinatura,
    valor_mensal,
    stripe_customer_id,
    stripe_subscription_id,
    data_proxima_cobranca
FROM assinaturas_usuario
WHERE stripe_subscription_id IS NOT NULL;
```

Deve mostrar:
- `status = 'PERIODO_GRATUITO'` (durante trial)
- `plano_assinatura = 'MENSAL'` (ou outro escolhido)
- `stripe_customer_id` preenchido
- `stripe_subscription_id` preenchido
- `data_proxima_cobranca` = hoje + 14 dias + intervalo do plano

## 🔄 Fluxo de Assinatura

### Durante o Trial (14 dias)

1. Cliente escolhe plano e completa checkout
2. Webhook `checkout.session.completed` é chamado
3. Sistema cria assinatura com status `PERIODO_GRATUITO`
4. Stripe marca subscription como `trialing`
5. Cliente tem acesso total ao sistema
6. Nenhuma cobrança é feita

### Após o Trial (Dia 15)

1. Stripe cobra automaticamente o cartão
2. Se pagamento for bem-sucedido:
   - Webhook `customer.subscription.updated` com status `active`
   - Sistema atualiza para status `ATIVA`
   - Próxima cobrança agendada para daqui 1/3/12 meses
3. Se pagamento falhar:
   - Webhook `customer.subscription.updated` com status `past_due`
   - Sistema atualiza para status `PENDENTE_PAGAMENTO`
   - Stripe tenta novamente automaticamente

### Renovação Automática

- Stripe cobra automaticamente todo mês/trimestre/ano
- Webhooks atualizam o sistema a cada cobrança
- Cliente não precisa fazer nada

### Cancelamento

1. Cliente cancela via interface do IFINU
2. Sistema chama API do Stripe para cancelar
3. Webhook `customer.subscription.deleted` é enviado
4. Status atualizado para `CANCELADA`
5. Cobranças automáticas são interrompidas

## 📊 Monitoramento

### Verificar Assinaturas Ativas

No Stripe Dashboard:
- **Subscriptions**: Veja todas as assinaturas
- **Customers**: Veja todos os clientes
- **Revenue**: Veja receita recorrente

### Logs de Webhook

https://dashboard.stripe.com/webhooks

- Veja todos os eventos enviados
- Status code de cada chamada
- Payload e resposta
- Reenviar eventos manualmente se necessário

## ⚠️ Troubleshooting

### Erro: Price ID não configurado

```
Price ID não configurado para o plano MENSAL
```

**Solução**: Adicionar variável de ambiente `STRIPE_PRICE_ID_MENSAL` no servidor.

### Assinatura não aparece no Stripe

**Causas possíveis**:
1. Webhook não está configurado
2. URL do webhook está incorreta
3. API não está processando webhooks

**Verificar**:
```bash
# Ver logs do webhook no Stripe Dashboard
https://dashboard.stripe.com/webhooks

# Ver logs da API
docker logs ifinu-api
```

### Assinatura criada mas status não atualiza

**Causa**: Webhook não está sendo processado

**Solução**:
1. Verificar se webhook está ativo no Stripe
2. Reenviar evento manualmente no Stripe Dashboard
3. Verificar logs da API

## 🔒 Segurança (Opcional mas Recomendado)

### Validar Signature do Webhook

Adicionar validação para garantir que webhooks vêm realmente do Stripe:

1. Copiar Signing Secret do webhook
2. Adicionar variável de ambiente:
   ```bash
   STRIPE_WEBHOOK_SECRET=whsec_xxxxxxxxxxxxx
   ```
3. Implementar validação no código (futuro)

## 📚 Documentação Stripe

- **Subscriptions**: https://stripe.com/docs/billing/subscriptions/overview
- **Trials**: https://stripe.com/docs/billing/subscriptions/trials
- **Webhooks**: https://stripe.com/docs/webhooks
- **Testing**: https://stripe.com/docs/testing
