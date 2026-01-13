-- Migration: Criar tabela de configurações Stripe do usuário
-- Data: 2026-01-13
-- Descrição: Permite que cada usuário configure suas próprias chaves Stripe

-- Criar tabela stripe_config
CREATE TABLE IF NOT EXISTS stripe_config (
    id BIGSERIAL PRIMARY KEY,
    usuario_id UUID NOT NULL UNIQUE REFERENCES usuarios(id) ON DELETE CASCADE,
    publishable_key VARCHAR(255) NOT NULL,
    secret_key_encrypted TEXT NOT NULL,
    test_mode BOOLEAN NOT NULL DEFAULT true,
    data_criacao TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_atualizacao TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Índices
    CONSTRAINT stripe_config_usuario_id_key UNIQUE (usuario_id)
);

-- Criar índice
CREATE INDEX IF NOT EXISTS idx_stripe_config_usuario_id ON stripe_config(usuario_id);

-- Comentários
COMMENT ON TABLE stripe_config IS 'Configurações Stripe por usuário';
COMMENT ON COLUMN stripe_config.publishable_key IS 'Chave pública do Stripe (pk_test_ ou pk_live_)';
COMMENT ON COLUMN stripe_config.secret_key_encrypted IS 'Chave secreta do Stripe criptografada';
COMMENT ON COLUMN stripe_config.test_mode IS 'Se está em modo teste ou produção';

-- Verificar tabela criada
SELECT
    table_name,
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns
WHERE table_name = 'stripe_config'
ORDER BY ordinal_position;
