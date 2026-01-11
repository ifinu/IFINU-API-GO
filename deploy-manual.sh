#!/bin/bash

echo "🚀 Deploy Manual IFINU API GO"
echo "=============================="
echo ""

# Fazer commit e push das mudanças locais
echo "📦 Fazendo commit das mudanças locais..."
git add .
read -p "Mensagem do commit: " MSG
git commit -m "$MSG"
git push origin main

echo ""
echo "🔄 Executando deploy no servidor..."
echo ""

sshpass -p 'Theo231023@' ssh -o StrictHostKeyChecking=no -o PreferredAuthentications=password -o PubkeyAuthentication=no -o IdentitiesOnly=yes mpx@192.168.0.100 << 'ENDSSH'
cd /home/mpx/ifinu-api-go

echo "📥 Atualizando código..."
git pull origin main

echo "🛑 Parando container..."
docker stop ifinu-api-go || true
docker rm ifinu-api-go || true

echo "🔨 Buildando imagem..."
docker build -t ifinu-api-go:latest .

echo "🚀 Iniciando container..."
docker run -d \
  -p 8080:8080 \
  --name ifinu-api-go \
  --network ifinu-network \
  --restart unless-stopped \
  --env-file .env \
  ifinu-api-go:latest

echo "⏳ Aguardando 5s..."
sleep 5

echo "✅ Deploy concluído!"
docker ps | grep ifinu-api-go
curl -s http://localhost:8080/health | python3 -m json.tool

ENDSSH

echo ""
echo "=============================="
echo "✅ DEPLOY CONCLUÍDO COM SUCESSO!"
echo "=============================="
echo ""
echo "🌐 API: http://192.168.0.100:8080"
echo "💚 Health: http://192.168.0.100:8080/health"
echo ""
