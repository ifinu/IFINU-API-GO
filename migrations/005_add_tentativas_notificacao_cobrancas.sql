-- Migration: Adicionar coluna tentativas_notificacao na tabela cobrancas
-- Data: 2026-01-12
-- Descrição: Adiciona campo para contar tentativas de envio de notificação

-- Adicionar coluna com valor padrão 0
ALTER TABLE cobrancas ADD COLUMN IF NOT EXISTS tentativas_notificacao INTEGER NOT NULL DEFAULT 0;

-- Atualizar registros existentes com NULL para 0
UPDATE cobrancas SET tentativas_notificacao = 0 WHERE tentativas_notificacao IS NULL;

-- Verificar estrutura final
SELECT column_name, data_type, column_default, is_nullable
FROM information_schema.columns
WHERE table_name = 'cobrancas' AND column_name = 'tentativas_notificacao';
