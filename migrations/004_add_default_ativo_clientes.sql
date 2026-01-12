-- Migration: Adicionar valor padrão TRUE para coluna ativo na tabela clientes
-- Data: 2026-01-12
-- Descrição: Define valor padrão TRUE para novos registros e atualiza registros existentes com NULL

-- Atualizar registros existentes com ativo NULL para TRUE
UPDATE clientes SET ativo = true WHERE ativo IS NULL;

-- Adicionar valor padrão TRUE para a coluna ativo
ALTER TABLE clientes ALTER COLUMN ativo SET DEFAULT true;

-- Verificar estrutura final
SELECT column_name, data_type, column_default, is_nullable
FROM information_schema.columns
WHERE table_name = 'clientes' AND column_name = 'ativo';
