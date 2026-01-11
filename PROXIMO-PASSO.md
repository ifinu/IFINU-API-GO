# ⚡ PRÓXIMO PASSO - EXECUTAR DEPLOY

## 🎯 Opção 1: Executar Script Automático (MAIS FÁCIL)

Execute este comando no seu terminal local:

```bash
cd /Users/mikael/Documents/OfIIfinu/ifinu-api-go
bash EXECUTAR-AGORA.sh
```

Quando pedir a senha SSH, digite: `Theo231023@`

O script irá:
- ✅ Parar o backend Java
- ✅ Clonar/Atualizar repositório Go
- ✅ Buildar imagem Docker
- ✅ Iniciar API Go na porta 8080
- ✅ Verificar health check

---

## 🛠️ Opção 2: Executar Comandos Manualmente

### Passo 1: Conectar no servidor

```bash
ssh mpx@192.168.0.100
# Senha: Theo231023@
```

### Passo 2: Copiar e colar este bloco completo

```bash
# Parar Java
echo "🛑 Parando Java..."
docker stop ifinu-api-java 2>/dev/null || true
docker rm ifinu-api-java 2>/dev/null || true
docker stop ifinu-api 2>/dev/null || true
docker rm ifinu-api 2>/dev/null || true
lsof -t -i:8080 | xargs kill -9 2>/dev/null || true

# Clonar/Atualizar Go
echo "📦 Configurando repositório Go..."
if [ ! -d "/home/mpx/ifinu-api-go" ]; then
    cd /home/mpx
    git clone https://github.com/ifinu/IFINU-API-GO.git ifinu-api-go
else
    cd /home/mpx/ifinu-api-go
    git fetch origin && git reset --hard origin/main
fi
cd /home/mpx/ifinu-api-go

# Criar .env se não existir
if [ ! -f ".env" ]; then
    echo "⚠️  Criando .env..."
    cp .env.example .env
    echo "⚠️  Configure o .env: nano .env"
fi

# Build e executar
echo "🔨 Buildando..."
docker build -t ifinu-api-go:latest .
echo "🚀 Iniciando..."
docker run -d -p 8080:8080 --name ifinu-api --restart unless-stopped --env-file .env ifinu-api-go:latest

# Verificar
sleep 5
echo "✅ Verificando..."
docker ps | grep ifinu-api
curl http://localhost:8080/health
docker logs --tail 20 ifinu-api

echo ""
echo "✅ DEPLOY CONCLUÍDO!"
```

---

## 🔍 Verificar Status

Depois do deploy, teste:

```bash
# Health check
curl http://192.168.0.100:8080/health

# Ver logs
ssh mpx@192.168.0.100 "docker logs -f ifinu-api"

# Status do container
ssh mpx@192.168.0.100 "docker ps | grep ifinu-api"
```

---

## 📋 Configurar GitHub Actions (Depois)

Para deploy automático nos próximos commits:

1. Acesse: https://github.com/ifinu/IFINU-API-GO/settings/secrets/actions

2. Adicione estes secrets:
   - `SSH_HOST`: 192.168.0.100
   - `SSH_USER`: mpx
   - `SSH_PASSWORD`: Theo231023@

3. Pronto! Todo push na `main` fará deploy automático

Ver guia completo: [GITHUB-ACTIONS.md](GITHUB-ACTIONS.md)

---

## ⚠️ IMPORTANTE: Configurar .env

Se for a primeira vez, você precisa configurar o `.env` no servidor com suas credenciais:

```bash
ssh mpx@192.168.0.100
nano /home/mpx/ifinu-api-go/.env
```

Configurar:
- `DB_HOST`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`
- `JWT_SECRET` (gere uma chave segura)
- `EVOLUTION_API_URL`, `EVOLUTION_API_KEY`
- `RESEND_API_KEY`
- `STRIPE_SECRET_KEY` (se usar)

Depois reinicie o container:
```bash
docker restart ifinu-api
```

---

## 🎉 Resultado Esperado

✅ Backend Java parado
✅ Porta 8080 liberada
✅ API Go rodando em http://192.168.0.100:8080
✅ Health check retornando {"status":"ok"}
✅ Container reiniciando automaticamente
✅ Logs mostrando "Servidor iniciando na porta 8080..."

---

**Está pronto para executar!** Execute `bash EXECUTAR-AGORA.sh` e acompanhe. 🚀
