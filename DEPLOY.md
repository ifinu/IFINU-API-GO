# 🚀 Guia de Deploy - IFINU API GO

## Pré-requisitos

- Go 1.22+ instalado
- PostgreSQL 14+ rodando
- Acesso ao servidor: `mpx@192.168.0.100`
- Variáveis de ambiente configuradas

## 1️⃣ Configurar Variáveis de Ambiente

Crie o arquivo `.env` na raiz do projeto:

```bash
cp .env.example .env
```

Edite o `.env` com suas configurações:

```env
# Aplicação
APP_ENV=production
APP_PORT=8080

# Banco de Dados (usar o servidor)
DB_HOST=192.168.0.100
DB_PORT=5432
DB_NAME=ifinu
DB_USER=seu_usuario
DB_PASSWORD=sua_senha
DB_SSL_MODE=disable

# JWT
JWT_SECRET=sua_chave_secreta_super_segura_de_64_caracteres_minimo
JWT_EXPIRATION_HOURS=24
JWT_REFRESH_EXPIRATION_DAYS=7

# Evolution API (WhatsApp)
EVOLUTION_API_URL=https://wp.ifinu.io
EVOLUTION_API_KEY=sua_chave_evolution_api

# Resend (Email)
RESEND_API_KEY=sua_chave_resend

# Stripe (Pagamentos)
STRIPE_SECRET_KEY=sua_chave_stripe
```

## 2️⃣ Deploy Local (Desenvolvimento)

```bash
# Baixar dependências
make mod

# Executar aplicação
make run

# Ou com hot reload (requer air)
make dev
```

A API estará disponível em: `http://localhost:8080`

## 3️⃣ Deploy em Produção (Servidor)

### Opção A: Deploy via Docker (Recomendado)

```bash
# 1. Build da imagem
make docker-build

# 2. Parar container antigo (se existir)
make docker-stop

# 3. Executar novo container
make docker-run
```

### Opção B: Deploy com Binário

```bash
# 1. Matar processo Java na porta 8080
make kill-8080

# 2. Build e executar
make deploy
```

### Opção C: Deploy no Servidor via SSH

```bash
# 1. Conectar ao servidor
ssh mpx@192.168.0.100
# Senha: Theo231023@

# 2. Clonar repositório (primeira vez)
cd /home/mpx
git clone https://github.com/ifinu/IFINU-API-GO.git ifinu-api-go
cd ifinu-api-go

# 3. Ou atualizar repositório existente
cd /home/mpx/ifinu-api-go
git pull origin main

# 4. Configurar .env (copiar do exemplo e editar)
cp .env.example .env
nano .env

# 5. Matar processo antigo na porta 8080
lsof -t -i:8080 | xargs kill -9 || true

# 6. Executar com Docker
docker build -t ifinu-api-go:latest .
docker stop ifinu-api || true
docker rm ifinu-api || true
docker run -d -p 8080:8080 --name ifinu-api --restart unless-stopped ifinu-api-go:latest

# 7. Verificar logs
docker logs -f ifinu-api
```

## 4️⃣ Verificar Deploy

```bash
# Health check
curl http://192.168.0.100:8080/health

# Deve retornar:
# {"status":"ok","message":"IFINU API GO está rodando","version":"1.0.0"}

# Testar endpoint de login
curl -X POST http://192.168.0.100:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"teste@example.com","senha":"senha123"}'
```

## 5️⃣ Comandos Úteis

```bash
# Ver status dos containers
docker ps | grep ifinu

# Ver logs da aplicação
docker logs --tail 100 ifinu-api

# Acompanhar logs em tempo real
docker logs -f ifinu-api

# Parar aplicação
docker stop ifinu-api

# Reiniciar aplicação
docker restart ifinu-api

# Remover container
docker stop ifinu-api && docker rm ifinu-api

# Limpar imagens antigas
docker image prune -a
```

## 6️⃣ Endpoints Disponíveis

### Públicos
- `GET /` - Welcome
- `GET /health` - Health check
- `POST /api/auth/login` - Login
- `POST /api/auth/cadastro` - Cadastro
- `POST /api/auth/refresh` - Refresh token

### Protegidos (requer JWT)
- `GET /api/auth/me` - Dados do usuário
- `GET /api/clientes` - Listar clientes
- `POST /api/clientes` - Criar cliente
- `GET /api/cobrancas` - Listar cobranças
- `POST /api/cobrancas` - Criar cobrança
- `POST /api/whatsapp/conectar` - Conectar WhatsApp
- `GET /api/whatsapp/status` - Status WhatsApp
- `POST /api/whatsapp/enviar` - Enviar mensagem

## 7️⃣ Monitoramento

### Logs do Scheduler
O agendador roda automaticamente e loga suas execuções:

```bash
# Filtrar logs do scheduler
docker logs ifinu-api 2>&1 | grep "Executando job"

# Ver notificações enviadas
docker logs ifinu-api 2>&1 | grep "enviado"
```

### Horários dos Jobs
- **9h diariamente**: Notificações de lembrete (3 dias antes)
- **9h diariamente**: Notificações de vencimento (hoje)
- **23h diariamente**: Atualização de cobranças vencidas

## 8️⃣ Troubleshooting

### Erro: Porta 8080 já em uso
```bash
# Matar processo na porta 8080
lsof -t -i:8080 | xargs kill -9

# Ou usar o comando do Makefile
make kill-8080
```

### Erro: Conexão com banco de dados
```bash
# Verificar se PostgreSQL está rodando
systemctl status postgresql

# Testar conexão
psql -h 192.168.0.100 -U seu_usuario -d ifinu
```

### Erro: Container não inicia
```bash
# Ver logs de erro
docker logs ifinu-api

# Verificar variáveis de ambiente
docker exec ifinu-api env | grep DB_
```

### Erro: JWT_SECRET não configurado
```bash
# Gerar uma chave segura
openssl rand -base64 64

# Adicionar no .env
JWT_SECRET=chave_gerada_aqui
```

## 9️⃣ Rollback (Voltar para Java)

Se precisar voltar para a versão Java:

```bash
# Parar container Go
docker stop ifinu-api

# Voltar para porta 8080 com Java
cd /home/mpx/ifinu-stack
docker-compose up -d api
```

## 🔟 Próximos Passos

- [ ] Configurar nginx reverse proxy
- [ ] Configurar SSL/TLS (HTTPS)
- [ ] Configurar backup automático do banco
- [ ] Configurar monitoramento (Prometheus/Grafana)
- [ ] Configurar CI/CD com GitHub Actions
- [ ] Adicionar testes automatizados

## 📞 Suporte

Em caso de problemas:
1. Verificar logs: `docker logs ifinu-api`
2. Verificar health check: `curl http://localhost:8080/health`
3. Verificar conexão com banco: testar credenciais
4. Verificar variáveis de ambiente: `.env` configurado corretamente

---

**Migrado de Java Spring Boot para Go** - Janeiro 2026
