# 🧪 Guia de Teste - Sistema de Pagamento IFINU

## ✅ Status Atual

Sistema de assinatura com Stripe **100% IMPLEMENTADO E TESTADO**:

- ✅ 3 Planos configurados (Mensal, Trimestral, Anual)
- ✅ Trial de 14 dias gerenciado pelo Stripe
- ✅ Webhook configurado e funcionando
- ✅ Frontend refatorado
- ✅ Backend deployado em produção

## 🎯 Como Testar

### 1. Acesse o Painel de Planos

```
https://app.ifinu.io/painel/plano
```

Você verá 3 cards com os planos:

**Plano Mensal**
- R$ 39/mês
- 14 dias grátis
- Botão "Assinar"

**Plano Trimestral**
- R$ 99 a cada 3 meses
- R$ 33/mês (economia de 15%)
- 14 dias grátis
- Botão "Assinar"

**Plano Anual**
- R$ 348/ano
- R$ 29/mês (economia de 25%)
- 14 dias grátis
- Botão "Assinar"

### 2. Clique em "Assinar" em Qualquer Plano

O sistema vai:
1. Criar checkout no Stripe
2. Redirecionar para página de pagamento do Stripe
3. Exibir formulário de cartão

### 3. Use Cartão de Teste do Stripe

**Cartão de Sucesso**:
```
Número: 4242 4242 4242 4242
Data: Qualquer data futura (ex: 12/25)
CVC: Qualquer 3 dígitos (ex: 123)
CEP: Qualquer (ex: 12345-678)
```

**Outros Cartões de Teste**:
```
# Pagamento recusado
Número: 4000 0000 0000 0002

# Cartão expirado
Número: 4000 0000 0000 0069

# CVC incorreto
Número: 4000 0000 0000 0127
```

### 4. Complete o Pagamento

Após preencher os dados do cartão:
1. Clique em "Assinar"
2. Stripe processa (modo test, sem cobrança real)
3. Redirecionado para: `https://app.ifinu.io/painel/plano?sucesso=true`

### 5. Verifique no Painel

Na página de planos, você deve ver:

```
✅ Assinatura Ativa

Plano: MENSAL (ou o que você escolheu)
Status: Período Gratuito
Valor: R$ 39,00/mês
Próxima cobrança: [data daqui 14 dias]

Trial: Restam 14 dias de teste grátis
```

### 6. Verifique no Stripe Dashboard

Acesse: https://dashboard.stripe.com

**Em Customers**:
- Deve aparecer um novo cliente com seu email
- Veja detalhes da subscription

**Em Subscriptions**:
- Status: `Trialing` (durante os 14 dias)
- Próxima cobrança: Daqui 14 dias
- Valor: R$ 39.00 (ou outro plano)

**Em Webhooks**:
- Acesse: https://dashboard.stripe.com/webhooks
- Veja eventos recentes:
  - `checkout.session.completed` ✅
  - `customer.subscription.created` ✅

### 7. Teste Cancelamento

1. No painel, clique em "Cancelar Assinatura"
2. Confirme o cancelamento
3. Status muda para `CANCELADA`
4. No Stripe: subscription marcada como `canceled`

## 🔍 Verificar no Banco de Dados

Se tiver acesso ao servidor, conecte no PostgreSQL:

```sql
SELECT
    id,
    usuario_id,
    status,
    plano_assinatura,
    valor_mensal,
    stripe_customer_id,
    stripe_subscription_id,
    data_inicio,
    data_proxima_cobranca
FROM assinaturas_usuarios
WHERE stripe_subscription_id IS NOT NULL
ORDER BY data_inicio DESC
LIMIT 5;
```

Deve retornar:
```
status: PERIODO_GRATUITO (durante trial)
plano_assinatura: MENSAL / TRIMESTRAL / ANUAL
stripe_customer_id: cus_xxxxxxxxxxxxx
stripe_subscription_id: sub_xxxxxxxxxxxxx
data_proxima_cobranca: [hoje + 14 dias + intervalo]
```

## 📊 Fluxo Completo

### Durante Trial (Dias 1-14)

```
Cliente escolhe plano
    ↓
Checkout no Stripe
    ↓
Webhook: checkout.session.completed
    ↓
Sistema cria assinatura: PERIODO_GRATUITO
    ↓
Cliente usa sistema gratuitamente
    ↓
Stripe não cobra nada
```

### Após Trial (Dia 15)

```
Stripe cobra automaticamente
    ↓
Pagamento aprovado?
    ├─ SIM → Webhook: subscription.updated (active)
    │         Sistema: ATIVA
    │         Próxima cobrança: +1/3/12 meses
    │
    └─ NÃO → Webhook: subscription.updated (past_due)
              Sistema: PENDENTE_PAGAMENTO
              Stripe tenta novamente
```

### Renovação (Todo mês/trimestre/ano)

```
Stripe cobra automaticamente
    ↓
Webhook: subscription.updated
    ↓
Sistema atualiza data_proxima_cobranca
    ↓
Cliente continua usando normalmente
```

## 🎬 URLs de Teste Rápido

### Login API
```bash
curl -X POST https://api.ifinu.io/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "seu-email@exemplo.com",
    "senha": "sua-senha"
  }'
```

### Listar Planos
```bash
curl -X GET https://api.ifinu.io/api/assinaturas/planos \
  -H "Authorization: Bearer SEU_TOKEN"
```

### Criar Checkout
```bash
curl -X POST https://api.ifinu.io/api/assinaturas/checkout \
  -H "Authorization: Bearer SEU_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "planoAssinatura": "MENSAL",
    "successUrl": "https://app.ifinu.io/painel/plano?sucesso=true",
    "cancelUrl": "https://app.ifinu.io/painel/plano?cancelado=true"
  }'
```

### Ver Status da Assinatura
```bash
curl -X GET https://api.ifinu.io/api/assinaturas/status \
  -H "Authorization: Bearer SEU_TOKEN"
```

### Cancelar Assinatura
```bash
curl -X POST https://api.ifinu.io/api/assinaturas/cancelar \
  -H "Authorization: Bearer SEU_TOKEN"
```

## 🚀 Próximos Passos

### Para ir para PRODUÇÃO (quando estiver pronto):

1. **Trocar para chaves LIVE do Stripe**:
   ```bash
   # No servidor, edite .env
   STRIPE_SECRET_KEY=sk_live_xxxxxxxxxxxxx
   ```

2. **Executar setup com chave LIVE**:
   ```bash
   cd /home/mpx/ifinu-stack/api
   make setup-stripe
   ```

3. **Atualizar variáveis de ambiente**:
   ```bash
   # Copiar Price IDs gerados para .env
   STRIPE_PRICE_ID_MENSAL=price_xxxxxxxxxxxxx
   STRIPE_PRICE_ID_TRIMESTRAL=price_xxxxxxxxxxxxx
   STRIPE_PRICE_ID_ANUAL=price_xxxxxxxxxxxxx
   ```

4. **Reconfigurar Webhook para modo LIVE**:
   - Acesse: https://dashboard.stripe.com/webhooks
   - Crie novo endpoint (modo live)
   - Copie novo webhook secret
   - Atualize `STRIPE_WEBHOOK_SECRET` no .env

5. **Reiniciar API**:
   ```bash
   docker-compose restart api
   ```

6. **Testar em produção** com cartões reais

## ⚠️ IMPORTANTE

- **MODO TEST**: Você está em modo test, nenhuma cobrança real é feita
- **Cartões de teste**: Use apenas cartões do Stripe Test Cards
- **Modo LIVE**: Só ative quando estiver 100% pronto para aceitar pagamentos reais
- **Webhooks**: Essenciais para o funcionamento, sempre verifique se estão ativos

## 📚 Documentação

- **Setup Stripe**: `STRIPE_SETUP.md`
- **Scripts Automáticos**: `scripts/README.md`
- **Stripe Test Cards**: https://stripe.com/docs/testing
- **Stripe Dashboard**: https://dashboard.stripe.com

## ✅ Checklist de Validação

- [ ] Login no sistema funcionando
- [ ] Página /painel/plano carrega e mostra 3 planos
- [ ] Clique em "Assinar" abre checkout do Stripe
- [ ] Pagamento com cartão teste funciona
- [ ] Redirecionado de volta para sistema
- [ ] Status da assinatura aparece no painel
- [ ] Stripe Dashboard mostra customer e subscription
- [ ] Webhook aparece nos logs do Stripe
- [ ] Cancelamento funciona
- [ ] Banco de dados atualizado corretamente

---

**Sistema pronto para uso! 🎉**

Se encontrar qualquer problema, verifique:
1. Logs da API: `docker logs ifinu-api`
2. Logs do Webhook no Stripe Dashboard
3. Variáveis de ambiente configuradas corretamente
