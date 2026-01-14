# Configuração de Chaves Stripe por Usuário

## Implementação Concluída

Sistema completo para permitir que cada usuário configure suas próprias chaves de API do Stripe.

## O que foi implementado

### Backend (Go)
- **Entidade**: `stripe_config.go` - Modelo com chave secreta criptografada
- **Repositório**: `stripe_config_repositorio.go` - CRUD para configurações
- **Serviço**: `stripe_config_servico.go` - Lógica de negócio e validações
- **Controlador**: `stripe_config_controlador.go` - Endpoints REST
- **Utilitário**: `crypto_util.go` - Criptografia AES-256-GCM
- **DTOs**: `stripe_config_dto.go` - Estruturas de request/response

### Frontend (já implementado na sessão anterior)
- **Componente**: `StripeConfigManager.tsx` - Interface de configuração
- **Tela**: Integrado em Configurações

### Banco de Dados
- **Migration 007**: Tabela `stripe_config` criada com sucesso
- Relacionamento 1:1 com usuário via `usuario_id`
- Chave secreta armazenada criptografada

## Endpoints Disponíveis

### GET /api/stripe/config
Busca configuração do usuário autenticado.

**Response**:
```json
{
  "configured": true,
  "publishableKey": "pk_test_...",
  "secretKeyMasked": "sk_test_...",
  "testMode": true
}
```

### POST /api/stripe/config
Salva ou atualiza configuração do usuário.

**Request**:
```json
{
  "publishableKey": "pk_test_51...",
  "secretKey": "sk_test_51...",
  "testMode": true
}
```

### DELETE /api/stripe/config
Remove configuração do usuário.

### POST /api/stripe/test-connection
Testa se as chaves configuradas estão válidas.

**Response**:
```json
{
  "success": true,
  "message": "Conexão com Stripe estabelecida com sucesso!"
}
```

## Segurança

1. **Criptografia**: Chaves secretas criptografadas com AES-256-GCM
2. **Mascaramento**: API retorna apenas primeiros 7 + últimos 4 caracteres
3. **Validação**: Formato de chaves validado (pk_test_/pk_live_, sk_test_/sk_live_)
4. **Isolamento**: Cada usuário só acessa suas próprias configurações
5. **Modo**: Validação automática de modo test/live baseado no prefixo

## ⚠️ AÇÃO NECESSÁRIA NO SERVIDOR

Adicionar variável de ambiente no servidor de produção:

### SSH no servidor:
```bash
ssh mpx@192.168.0.100
```

### Editar arquivo .env:
```bash
cd /home/mpx/ifinu-stack
nano .env
```

### Adicionar ao final do arquivo:
```bash
# Criptografia (para chaves Stripe dos usuários)
ENCRYPTION_KEY=ifinu-prod-encryption-key-2026-secure-min-32-chars-random
```

**IMPORTANTE**: Gerar uma chave segura e aleatória com no mínimo 32 caracteres.

### Reiniciar container da API:
```bash
docker-compose restart api
```

### Verificar logs:
```bash
docker logs -f ifinu-api
```

## Como Usar (Frontend)

1. Usuário acessa **Configurações** no painel
2. Vai na aba **Stripe**
3. Insere:
   - Chave Pública (pk_test_ ou pk_live_)
   - Chave Secreta (sk_test_ ou sk_live_)
   - Seleciona modo Test ou Live
4. Clica em **Testar Conexão** (opcional)
5. Clica em **Salvar Configuração**

## Fluxo de Pagamento

Quando usuário cria uma cobrança com Stripe:
1. Sistema busca configuração do usuário via `ObterChaveSecreta()`
2. Usa chave do usuário para criar checkout session
3. Pagamento vai direto para conta Stripe do usuário
4. Webhook notifica o sistema (futuro)

## Próximos Passos

1. ✅ Backend implementado
2. ✅ Migration executada
3. ✅ Commit e push realizados
4. ⏳ Aguardar GitHub Actions deploy
5. ⏳ Configurar ENCRYPTION_KEY no servidor
6. ⏳ Testar fluxo completo no frontend

## Estrutura de Arquivos

```
ifinu-api-go/
├── controlador/
│   └── stripe_config_controlador.go
├── dominio/
│   └── entidades/
│       └── stripe_config.go
├── dto/
│   └── stripe_config_dto.go
├── migrations/
│   └── 007_create_stripe_config.sql
├── repositorio/
│   └── stripe_config_repositorio.go
├── servico/
│   └── stripe_config_servico.go
└── util/
    └── crypto_util.go
```

## Commit

```
commit c7f0853
feat: Implementar configuração de chaves Stripe por usuário
```

Deploy automático em andamento via GitHub Actions.
