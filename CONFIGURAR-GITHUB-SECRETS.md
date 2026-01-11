# Configurar GitHub Secrets para Deploy Automático

## O que são GitHub Secrets?

São variáveis de ambiente criptografadas que o GitHub Actions usa para armazenar informações sensíveis (senhas, chaves SSH, etc.) de forma segura.

## Passo a Passo

### 1. Acesse as Configurações do Repositório

1. Vá para: https://github.com/ifinu/IFINU-API-GO
2. Clique na aba **Settings** (Configurações)
3. No menu lateral esquerdo, clique em **Secrets and variables**
4. Clique em **Actions**

### 2. Adicione os 3 Secrets Necessários

Clique no botão **New repository secret** para cada um:

#### Secret 1: SSH_HOST
- **Name:** `SSH_HOST`
- **Value:** `192.168.0.100`
- Clique em **Add secret**

#### Secret 2: SSH_USER
- **Name:** `SSH_USER`
- **Value:** `mpx`
- Clique em **Add secret**

#### Secret 3: SSH_PASSWORD
- **Name:** `SSH_PASSWORD`
- **Value:** `Theo231023@`
- Clique em **Add secret**

### 3. Verificar Configuração

Após adicionar os 3 secrets, você deve ver:

```
SSH_HOST          Updated now
SSH_USER          Updated now
SSH_PASSWORD      Updated now
```

### 4. Testar o Deploy Automático

Agora qualquer push para a branch `main` vai acionar o deploy automático!

Para testar manualmente:
1. Vá para a aba **Actions** no GitHub
2. Clique em **Deploy IFINU API GO** (workflow)
3. Clique em **Run workflow**
4. Selecione `main` branch
5. Clique em **Run workflow**

### 5. Acompanhar o Deploy

1. Na aba **Actions**, clique no workflow em execução
2. Clique em **Deploy para Produção**
3. Acompanhe os logs em tempo real

## O que o Deploy Faz Automaticamente

1. ✅ Conecta no servidor via SSH
2. ✅ Atualiza o código (git pull)
3. ✅ Para containers antigos (Java e Go)
4. ✅ Constrói nova imagem Docker
5. ✅ Inicia novo container na network correta
6. ✅ Verifica health check
7. ✅ Mostra logs do container

## Resultado Esperado

Se tudo funcionar corretamente, você verá:

```
✅ Deploy concluído com sucesso!
🔗 API disponível em: http://api.ifinu.io
```

## Troubleshooting

### Erro: "missing server host"
- Verifique se os 3 secrets foram configurados corretamente
- Certifique-se que os nomes estão EXATAMENTE como mostrado acima (case-sensitive)

### Erro: "Permission denied"
- Verifique se a senha está correta no secret SSH_PASSWORD
- Teste o acesso SSH manualmente: `ssh mpx@192.168.0.100`

### Erro: "Docker build failed"
- Verifique os logs do workflow
- Pode ser erro de compilação no código Go

### Container não inicia
- Verifique se o arquivo .env existe no servidor em `/home/mpx/ifinu-api-go/.env`
- Verifique se as credenciais do banco estão corretas

## Links Úteis

- **Repositório:** https://github.com/ifinu/IFINU-API-GO
- **Settings:** https://github.com/ifinu/IFINU-API-GO/settings/secrets/actions
- **Actions:** https://github.com/ifinu/IFINU-API-GO/actions
- **API Produção:** http://api.ifinu.io
- **Health Check:** http://192.168.0.100:8080/health

## Comandos Úteis no Servidor

```bash
# Ver logs do container
docker logs -f ifinu-api-go

# Ver status do container
docker ps | grep ifinu-api-go

# Reiniciar container
docker restart ifinu-api-go

# Parar container
docker stop ifinu-api-go

# Ver logs do health check
curl http://localhost:8080/health
```
