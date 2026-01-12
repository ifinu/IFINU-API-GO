-- Migration: Adicionar colunas faltantes na tabela clientes
-- Data: 2026-01-12
-- Descrição: Adiciona colunas CPF, CNPJ, Cidade, Estado e CEP

-- Adicionar todas as colunas necessárias
ALTER TABLE clientes ADD COLUMN IF NOT EXISTS cpf VARCHAR(14);
ALTER TABLE clientes ADD COLUMN IF NOT EXISTS cnpj VARCHAR(18);
ALTER TABLE clientes ADD COLUMN IF NOT EXISTS cidade VARCHAR(100);
ALTER TABLE clientes ADD COLUMN IF NOT EXISTS estado VARCHAR(2);
ALTER TABLE clientes ADD COLUMN IF NOT EXISTS cep VARCHAR(10);

-- Verificar estrutura final
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_name = 'clientes'
ORDER BY ordinal_position;
