-- Migration: Ajustar constraint de unidade_tempo para permitir NULL
-- Data: 2026-01-12
-- Descrição: Permite que unidade_tempo seja NULL para cobranças não personalizadas

-- Remover constraint antiga que não permitia NULL
ALTER TABLE cobrancas DROP CONSTRAINT IF EXISTS cobrancas_unidade_tempo_check;

-- Adicionar nova constraint permitindo NULL ou valores válidos
ALTER TABLE cobrancas ADD CONSTRAINT cobrancas_unidade_tempo_check
CHECK (unidade_tempo IS NULL OR unidade_tempo IN ('DIAS', 'SEMANAS', 'MESES', 'ANOS'));

-- Verificar constraint
SELECT conname, pg_get_constraintdef(oid)
FROM pg_constraint
WHERE conrelid = 'cobrancas'::regclass
AND conname = 'cobrancas_unidade_tempo_check';
