-- Migration: Adicionar colunas faltantes na tabela whatsapp_conexoes
-- Data: 2026-01-11
-- Descrição: Adiciona colunas para estatísticas e número conectado

-- Adicionar coluna numero_conectado se não existir
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='whatsapp_conexoes' AND column_name='numero_conectado'
    ) THEN
        ALTER TABLE whatsapp_conexoes ADD COLUMN numero_conectado VARCHAR(20);
        RAISE NOTICE 'Coluna numero_conectado adicionada';
    ELSE
        RAISE NOTICE 'Coluna numero_conectado já existe';
    END IF;
END $$;

-- Adicionar coluna mensagens_enviadas se não existir
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='whatsapp_conexoes' AND column_name='mensagens_enviadas'
    ) THEN
        ALTER TABLE whatsapp_conexoes ADD COLUMN mensagens_enviadas INTEGER DEFAULT 0;
        RAISE NOTICE 'Coluna mensagens_enviadas adicionada';
    ELSE
        RAISE NOTICE 'Coluna mensagens_enviadas já existe';
    END IF;
END $$;

-- Adicionar coluna mensagens_sucesso se não existir
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='whatsapp_conexoes' AND column_name='mensagens_sucesso'
    ) THEN
        ALTER TABLE whatsapp_conexoes ADD COLUMN mensagens_sucesso INTEGER DEFAULT 0;
        RAISE NOTICE 'Coluna mensagens_sucesso adicionada';
    ELSE
        RAISE NOTICE 'Coluna mensagens_sucesso já existe';
    END IF;
END $$;

-- Adicionar coluna mensagens_falha se não existir
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='whatsapp_conexoes' AND column_name='mensagens_falha'
    ) THEN
        ALTER TABLE whatsapp_conexoes ADD COLUMN mensagens_falha INTEGER DEFAULT 0;
        RAISE NOTICE 'Coluna mensagens_falha adicionada';
    ELSE
        RAISE NOTICE 'Coluna mensagens_falha já existe';
    END IF;
END $$;

-- Verificar e exibir estrutura final
SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_name = 'whatsapp_conexoes'
ORDER BY ordinal_position;
