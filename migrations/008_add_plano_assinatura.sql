-- Migration: Adicionar campo plano_assinatura
-- Data: 2026-01-13
-- Descrição: Adiciona suporte aos 3 planos (MENSAL, TRIMESTRAL, ANUAL)

-- Adicionar coluna plano_assinatura
ALTER TABLE assinaturas_usuarios
ADD COLUMN IF NOT EXISTS plano_assinatura VARCHAR(20) DEFAULT 'MENSAL';

-- Adicionar constraint para validar valores
ALTER TABLE assinaturas_usuarios
ADD CONSTRAINT plano_assinatura_check
CHECK (plano_assinatura IN ('MENSAL', 'TRIMESTRAL', 'ANUAL'));

-- Comentário
COMMENT ON COLUMN assinaturas_usuarios.plano_assinatura IS 'Tipo de plano: MENSAL (R$ 39), TRIMESTRAL (R$ 99), ANUAL (R$ 348)';

-- Atualizar assinaturas existentes para MENSAL
UPDATE assinaturas_usuarios
SET plano_assinatura = 'MENSAL'
WHERE plano_assinatura IS NULL;

-- Verificar resultado
SELECT
    COUNT(*) as total,
    plano_assinatura,
    status
FROM assinaturas_usuarios
GROUP BY plano_assinatura, status;
