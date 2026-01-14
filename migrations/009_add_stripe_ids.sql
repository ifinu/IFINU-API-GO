-- Migration 009: Adicionar campos do Stripe
-- Adiciona IDs do Stripe Customer e Subscription

ALTER TABLE assinaturas_usuario
ADD COLUMN IF NOT EXISTS stripe_customer_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS stripe_subscription_id VARCHAR(255);

-- Criar índice para busca rápida por subscription ID
CREATE INDEX IF NOT EXISTS idx_assinaturas_stripe_subscription
ON assinaturas_usuario(stripe_subscription_id);

CREATE INDEX IF NOT EXISTS idx_assinaturas_stripe_customer
ON assinaturas_usuario(stripe_customer_id);
