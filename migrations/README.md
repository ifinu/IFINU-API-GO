# Migrations - IFINU API

## Como executar migrations no servidor

### Opção 1: Usando o script automatizado (Recomendado)

```bash
# SSH no servidor
ssh mpx@192.168.0.100

# Ir para o diretório da API
cd /opt/ifinu/api

# Dar permissão de execução ao script
chmod +x run_migration.sh

# Executar a migration (exemplo: 011)
./run_migration.sh 011
```

### Opção 2: Manualmente com psql

```bash
# SSH no servidor
ssh mpx@192.168.0.100

# Executar migration diretamente
docker exec -i ifinu-postgres psql -U ifinu -d ifinu < migrations/011_add_recorrencia_cobrancas.sql
```

### Opção 3: Dentro do container do banco

```bash
# Entrar no container do PostgreSQL
docker exec -it ifinu-postgres bash

# Executar psql
psql -U ifinu -d ifinu

# Copiar e colar o conteúdo do arquivo SQL
\i /path/to/migration/011_add_recorrencia_cobrancas.sql
```

## Migration 011: Adicionar Recorrência de Cobranças

**Descrição:** Adiciona as colunas necessárias para configurar recorrência de cobranças.

**Colunas adicionadas:**
- `tipo_recorrencia` (VARCHAR) - Tipo: UNICA, MENSAL, TRIMESTRAL, SEMESTRAL, ANUAL, PERSONALIZADO
- `intervalo_periodo` (INTEGER) - Intervalo para recorrência personalizada
- `unidade_tempo` (VARCHAR) - Unidade: DIAS, MESES, ANOS
- `proxima_cobranca` (TIMESTAMP) - Data da próxima cobrança recorrente
- `recorrencia_ativa` (BOOLEAN) - Se a recorrência está ativa

**Segurança:**
- Usa `IF NOT EXISTS` - não quebra se as colunas já existirem
- Define valores padrão seguros
- Atualiza registros existentes

## Checklist de Deploy

1. ✅ Criar arquivo de migration
2. ✅ Testar localmente
3. ⏳ Fazer backup do banco de produção
4. ⏳ Executar migration no servidor
5. ⏳ Verificar que as colunas foram criadas
6. ⏳ Deploy do código Go atualizado
7. ⏳ Testar criação de cobrança com recorrência

## Rollback

Se precisar reverter a migration:

```sql
ALTER TABLE cobrancas DROP COLUMN IF EXISTS tipo_recorrencia;
ALTER TABLE cobrancas DROP COLUMN IF EXISTS intervalo_periodo;
ALTER TABLE cobrancas DROP COLUMN IF EXISTS unidade_tempo;
ALTER TABLE cobrancas DROP COLUMN IF EXISTS proxima_cobranca;
ALTER TABLE cobrancas DROP COLUMN IF EXISTS recorrencia_ativa;
```
