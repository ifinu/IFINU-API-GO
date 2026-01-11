# 🔄 Configuração do GitHub Actions

## Passo 1: Configurar Secrets no GitHub

1. Acesse: https://github.com/ifinu/IFINU-API-GO/settings/secrets/actions

2. Adicione os seguintes secrets:

   - **SSH_HOST**: `192.168.0.100`
   - **SSH_USER**: `mpx`
   - **SSH_PASSWORD**: `Theo231023@`

### Como adicionar cada secret:

1. Clique em **"New repository secret"**
2. Digite o nome do secret (ex: `SSH_HOST`)
3. Cole o valor correspondente
4. Clique em **"Add secret"**
5. Repita para os outros secrets

## Passo 2: Executar Setup Inicial no Servidor

Antes do primeiro deploy via GitHub Actions, você precisa fazer o setup inicial manualmente:

### Via SSH direto:

```bash
# Conectar ao servidor
ssh mpx@192.168.0.100
# Senha: Theo231023@

# Baixar e executar script de setup
curl -o setup.sh https://raw.githubusercontent.com/ifinu/IFINU-API-GO/main/setup-server.sh
chmod +x setup.sh
bash setup.sh
```

### Ou passo a passo manual:

```bash
# 1. Conectar ao servidor
ssh mpx@192.168.0.100

# 2. Parar Java
docker stop ifinu-api-java 2>/dev/null || true
docker rm ifinu-api-java 2>/dev/null || true
docker stop ifinu-api 2>/dev/null || true
docker rm ifinu-api 2>/dev/null || true

# 3. Liberar porta 8080
lsof -t -i:8080 | xargs kill -9 2>/dev/null || true

# 4. Clonar repositório Go
cd /home/mpx
git clone https://github.com/ifinu/IFINU-API-GO.git ifinu-api-go
cd ifinu-api-go

# 5. Configurar .env
cp .env.example .env
nano .env
# Edite com as credenciais corretas

# 6. Build e executar
docker build -t ifinu-api-go:latest .
docker run -d -p 8080:8080 --name ifinu-api --restart unless-stopped --env-file .env ifinu-api-go:latest

# 7. Verificar
curl http://localhost:8080/health
docker logs ifinu-api
```

## Passo 3: Deploy Automático

Após configurar os secrets e fazer o setup inicial, todo push na branch `main` irá disparar o deploy automático!

### Como funciona:

1. Você faz commit e push para a branch `main`
2. GitHub Actions detecta o push
3. Conecta no servidor via SSH
4. Atualiza o código (git pull)
5. Para o container antigo
6. Builda nova imagem Docker
7. Inicia novo container
8. Testa health check
9. Mostra os logs

### Forçar deploy manual:

No GitHub, vá em:
- **Actions** → **Deploy IFINU API GO** → **Run workflow** → **Run workflow**

## Passo 4: Monitorar Deploys

### Ver status dos deploys:
https://github.com/ifinu/IFINU-API-GO/actions

### Ver logs de um deploy:
1. Clique no workflow executado
2. Clique em "Deploy para Produção"
3. Veja os logs de cada step

### Se o deploy falhar:
1. Verifique os logs no GitHub Actions
2. Conecte no servidor e verifique: `docker logs ifinu-api`
3. Verifique se o .env está configurado corretamente

## Resumo do Fluxo

```
┌─────────────────────┐
│  Código Local       │
│  (git push main)    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  GitHub Actions     │
│  (workflow)         │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  SSH no Servidor    │
│  (192.168.0.100)    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  git pull           │
│  docker build       │
│  docker run         │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  API Rodando ✅     │
│  (porta 8080)       │
└─────────────────────┘
```

## Troubleshooting

### Erro: "Permission denied (publickey)"
- Verifique se os secrets estão configurados corretamente
- Certifique-se que SSH_PASSWORD está correto: `Theo231023@`

### Erro: "Container ifinu-api already exists"
- O script deve remover automaticamente
- Se persistir, conecte no servidor e execute: `docker rm -f ifinu-api`

### Erro: "Port 8080 already in use"
- Conecte no servidor
- Execute: `lsof -t -i:8080 | xargs kill -9`
- Execute o workflow novamente

### Erro: "Health check failed"
- Verifique os logs: `docker logs ifinu-api`
- Verifique o .env no servidor
- Certifique-se que o banco está acessível

## Próximos Passos

- [ ] Configurar secrets no GitHub (SSH_HOST, SSH_USER, SSH_PASSWORD)
- [ ] Executar setup inicial no servidor
- [ ] Testar primeiro deploy manual via GitHub Actions
- [ ] Fazer um commit teste para validar deploy automático
- [ ] Configurar notificações de deploy (Slack/Discord/Email)

---

**Deploy automático configurado!** 🚀
