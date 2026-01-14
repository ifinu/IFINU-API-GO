-- Migration 010: Adicionar campos para Stripe Connect
-- Permite controlar onboarding e status da conta conectada

ALTER TABLE usuarios
ADD COLUMN IF NOT EXISTS stripe_onboarding_completo BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS stripe_charges_habilitado BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS stripe_detalhes_submetidos BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS stripe_data_onboarding TIMESTAMP;

-- Índice para buscar usuários com Stripe conectado
CREATE INDEX IF NOT EXISTS idx_usuarios_stripe_account
ON usuarios(stripe_account_id)
WHERE stripe_account_id IS NOT NULL;

-- Índice para buscar usuários com onboarding completo
CREATE INDEX IF NOT EXISTS idx_usuarios_stripe_onboarding
ON usuarios(stripe_onboarding_completo)
WHERE stripe_onboarding_completo = TRUE;

COMMENT ON COLUMN usuarios.stripe_onboarding_completo IS 'Se usuário completou onboarding do Stripe Connect';
COMMENT ON COLUMN usuarios.stripe_charges_habilitado IS 'Se conta Stripe pode receber pagamentos';
COMMENT ON COLUMN usuarios.stripe_detalhes_submetidos IS 'Se usuário submeteu todos detalhes necessários';
COMMENT ON COLUMN usuarios.stripe_data_onboarding IS 'Data que completou onboarding';
