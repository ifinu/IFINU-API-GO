#!/bin/bash

echo "🔧 Configurando GitHub Secrets para Deploy Automático..."
echo ""

# Verificar se gh CLI está instalado
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) não está instalado!"
    echo ""
    echo "Instale com:"
    echo "  brew install gh"
    echo ""
    exit 1
fi

# Verificar se está autenticado
if ! gh auth status &> /dev/null; then
    echo "🔐 Fazendo login no GitHub..."
    gh auth login
fi

echo "✅ Autenticado no GitHub"
echo ""

# Adicionar os secrets
echo "📝 Adicionando SSH_HOST..."
echo "192.168.0.100" | gh secret set SSH_HOST -R ifinu/IFINU-API-GO

echo "📝 Adicionando SSH_USER..."
echo "mpx" | gh secret set SSH_USER -R ifinu/IFINU-API-GO

echo "📝 Adicionando SSH_PASSWORD..."
echo "Theo231023@" | gh secret set SSH_PASSWORD -R ifinu/IFINU-API-GO

echo ""
echo "✅ Todos os secrets configurados com sucesso!"
echo ""
echo "🎯 Testando deploy automático..."
echo ""

# Disparar workflow
gh workflow run deploy.yml -R ifinu/IFINU-API-GO

echo "✅ Deploy disparado!"
echo ""
echo "📊 Acompanhe o progresso em:"
echo "   https://github.com/ifinu/IFINU-API-GO/actions"
echo ""
