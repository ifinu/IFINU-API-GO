-- Migration: Adicionar colunas de recorrência na tabela cobrancas
-- Data: 2026-01-16
-- Descrição: Adiciona campos para configurar recorrência de cobranças

-- Adicionar coluna tipo_recorrencia
ALTER TABLE cobrancas ADD COLUMN IF NOT EXISTS tipo_recorrencia VARCHAR(20) DEFAULT 'UNICA';

-- Adicionar coluna intervalo_periodo
ALTER TABLE cobrancas ADD COLUMN IF NOT EXISTS intervalo_periodo INTEGER;

-- Adicionar coluna unidade_tempo
ALTER TABLE cobrancas ADD COLUMN IF NOT EXISTS unidade_tempo VARCHAR(20);

-- Adicionar coluna proxima_cobranca
ALTER TABLE cobrancas ADD COLUMN IF NOT EXISTS proxima_cobranca TIMESTAMP;

-- Adicionar coluna recorrencia_ativa
ALTER TABLE cobrancas ADD COLUMN IF NOT EXISTS recorrencia_ativa BOOLEAN DEFAULT false;

-- Atualizar registros existentes com valores padrão
UPDATE cobrancas SET tipo_recorrencia = 'UNICA' WHERE tipo_recorrencia IS NULL;
UPDATE cobrancas SET recorrencia_ativa = false WHERE recorrencia_ativa IS NULL;

-- Verificar estrutura final
SELECT column_name, data_type, column_default, is_nullable
FROM information_schema.columns
WHERE table_name = 'cobrancas'
  AND column_name IN ('tipo_recorrencia', 'intervalo_periodo', 'unidade_tempo', 'proxima_cobranca', 'recorrencia_ativa')
ORDER BY column_name;
