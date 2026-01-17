#!/bin/bash

# Script para executar migrations no banco de dados
# Uso: ./run_migration.sh <numero_da_migration>
# Exemplo: ./run_migration.sh 011

set -e

MIGRATION_NUM=$1

if [ -z "$MIGRATION_NUM" ]; then
    echo "❌ Erro: Número da migration não especificado"
    echo "Uso: ./run_migration.sh <numero>"
    echo "Exemplo: ./run_migration.sh 011"
    exit 1
fi

MIGRATION_FILE="migrations/${MIGRATION_NUM}_*.sql"

if ! ls $MIGRATION_FILE 1> /dev/null 2>&1; then
    echo "❌ Erro: Migration $MIGRATION_NUM não encontrada"
    exit 1
fi

echo "🔄 Executando migration: $MIGRATION_FILE"

# Carrega variáveis de ambiente do .env se existir
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Usa variáveis de ambiente ou valores padrão
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-ifinu}
DB_USER=${DB_USER:-postgres}

# Executar migration
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f $MIGRATION_FILE

if [ $? -eq 0 ]; then
    echo "✅ Migration $MIGRATION_NUM executada com sucesso!"
else
    echo "❌ Erro ao executar migration $MIGRATION_NUM"
    exit 1
fi
