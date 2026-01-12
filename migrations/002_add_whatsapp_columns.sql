-- Migration: Adicionar colunas faltantes na tabela whatsapp_conexoes
-- Data: 2026-01-11
-- Descrição: Adiciona colunas para estatísticas, número conectado e timestamps

-- Adicionar todas as colunas necessárias
ALTER TABLE whatsapp_conexoes ADD COLUMN IF NOT EXISTS numero_conectado VARCHAR(20);
ALTER TABLE whatsapp_conexoes ADD COLUMN IF NOT EXISTS mensagens_enviadas INTEGER DEFAULT 0;
ALTER TABLE whatsapp_conexoes ADD COLUMN IF NOT EXISTS mensagens_sucesso INTEGER DEFAULT 0;
ALTER TABLE whatsapp_conexoes ADD COLUMN IF NOT EXISTS mensagens_falha INTEGER DEFAULT 0;
ALTER TABLE whatsapp_conexoes ADD COLUMN IF NOT EXISTS data_conexao TIMESTAMP;
ALTER TABLE whatsapp_conexoes ADD COLUMN IF NOT EXISTS data_atualizacao TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Atualizar data_atualizacao para registros existentes
UPDATE whatsapp_conexoes SET data_atualizacao = CURRENT_TIMESTAMP WHERE data_atualizacao IS NULL;

-- Verificar estrutura final
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_name = 'whatsapp_conexoes'
ORDER BY ordinal_position;
