# Arquitetura de Alta Escala - IFINU API

Sistema projetado para enviar **milhares de mensagens WhatsApp simultâneas** com auto-scaling e alta disponibilidade.

## 🎯 Problema Resolvido

### Antes
❌ Sem controle de concorrência (goroutines ilimitadas)
❌ Sem rate limiting (Evolution API bloqueava)
❌ Sem retry automático
❌ Sem escalabilidade horizontal
❌ Timeout de 120s por mensagem
❌ Capacidade: ~100 mensagens/minuto

### Depois
✅ Worker Pool com limite de concorrência
✅ Rate Limiter: 50 msg/s por pod
✅ Retry automático com backoff exponencial
✅ Auto-scaling de 2 a 20 pods (Kubernetes)
✅ Fila distribuída (Redis)
✅ **Capacidade: 60.000+ mensagens/minuto** (1.000 msg/s)

---

## 📊 Capacidade do Sistema

### Com 2 Pods (Mínimo)
- **Workers**: 20 (10 por pod)
- **Throughput**: 100 msg/s
- **Capacidade/hora**: 360.000 mensagens
- **Custo AWS**: ~$50/mês

### Com 10 Pods (Médio)
- **Workers**: 100 (10 por pod)
- **Throughput**: 500 msg/s
- **Capacidade/hora**: 1.800.000 mensagens
- **Custo AWS**: ~$250/mês

### Com 20 Pods (Máximo)
- **Workers**: 200 (10 por pod)
- **Throughput**: 1.000 msg/s
- **Capacidade/hora**: 3.600.000 mensagens
- **Custo AWS**: ~$500/mês

---

## 🏗️ Componentes da Arquitetura

### 1. **Worker Pool (Go)**
```go
// 10 workers por pod processando em paralelo
filaMensagem.IniciarWorkerPool(10)

// Rate Limiter: 50 mensagens/segundo
rate.NewLimiter(rate.Limit(50), 100)
```

**Funcionalidades:**
- Processa mensagens da fila Redis
- Controla concorrência (máx 10 goroutines/pod)
- Respeita rate limit da Evolution API
- Retry automático em falhas

### 2. **Fila Distribuída (Redis)**
```
ifinu:fila:whatsapp → [msg1, msg2, msg3, ...]
```

**Funcionalidades:**
- FIFO (First In, First Out)
- Persistente (sobrevive a reinicializações)
- Suporta múltiplos producers/consumers
- Retry automático até 3 tentativas
- Backoff exponencial: 5min, 25min, 2h

### 3. **Evolution API Pods (Kubernetes)**
```yaml
replicas: 2-20  # Auto-scaling
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi
```

**Funcionalidades:**
- Load Balancer distribui requisições
- Session Affinity mantém sessões
- Health checks automáticos
- Restart automático em falhas

### 4. **HPA (Horizontal Pod Autoscaler)**
```yaml
minReplicas: 2
maxReplicas: 20
targetCPU: 70%
targetMemory: 80%
```

**Comportamento:**
- **Scale Up**: +50% ou +2 pods quando CPU > 70%
- **Scale Down**: -10% ou -1 pod após 5 min estável
- **Tempo de reação**: 60 segundos
- **Estabilização**: 5 minutos

---

## 🔄 Fluxo de Processamento

### 1. Agendador Detecta Cobranças (CRON)
```
09:00 → Buscar cobranças vencendo hoje
      → Buscar cobranças para lembrete (3 dias antes)
```

### 2. Enfileiramento (Agendador → Redis)
```go
for _, cobranca := range cobrancas {
    msg := &MensagemFila{
        ID:              "lembrete_123_1234567890",
        TipoNotificacao: "lembrete",
        Cobranca:        cobranca,
        Tentativas:      0,
    }
    filaMensagem.EnfileirarMensagem(msg)
}
```

### 3. Processamento (Workers → Evolution API)
```
Worker 1 → Redis BRPOP → Mensagem 1 → Evolution API Pod 1 → WhatsApp
Worker 2 → Redis BRPOP → Mensagem 2 → Evolution API Pod 2 → WhatsApp
Worker 3 → Redis BRPOP → Mensagem 3 → Evolution API Pod 1 → WhatsApp
...
Worker 200 → Redis BRPOP → Mensagem N → Evolution API Pod 20 → WhatsApp
```

### 4. Retry em Falha
```
Tentativa 1 (imediato) → FALHA
Tentativa 2 (+5 min)   → FALHA
Tentativa 3 (+25 min)  → SUCESSO ✅
```

### 5. Auto-Scaling Automático
```
Fila: 10.000 mensagens → CPU: 85% → HPA: +4 pods (2→6)
Fila: 50.000 mensagens → CPU: 90% → HPA: +8 pods (6→14)
Fila: vazia           → CPU: 30% → HPA: -2 pods (14→12)
```

---

## 📈 Monitoramento

### Métricas Importantes

1. **Taxa de Processamento**
   ```bash
   kubectl top pods -n ifinu-production
   ```

2. **Tamanho da Fila**
   ```bash
   kubectl exec -it deployment/redis-queue -n ifinu-production -- \
     redis-cli LLEN ifinu:fila:whatsapp
   ```

3. **Número de Pods Ativos**
   ```bash
   kubectl get hpa evolution-api-hpa -n ifinu-production
   ```

4. **Taxa de Sucesso/Falha**
   - Ver logs: `kubectl logs -f deployment/evolution-api -n ifinu-production`
   - Buscar: `✅` (sucesso) e `❌` (falha)

### Alertas Recomendados

```yaml
# Alerta: Fila muito grande
- alert: FilaWhatsAppGrande
  expr: redis_list_length{key="ifinu:fila:whatsapp"} > 10000
  for: 10m
  annotations:
    summary: "Fila WhatsApp com {{ $value }} mensagens pendentes"

# Alerta: Taxa de falha alta
- alert: TaxaFalhaAlta
  expr: rate(whatsapp_falhas_total[5m]) > 0.1
  for: 5m
  annotations:
    summary: "Taxa de falha > 10% nos últimos 5 minutos"

# Alerta: Pods no máximo
- alert: PodsNoMaximo
  expr: kube_deployment_status_replicas{deployment="evolution-api"} >= 20
  for: 30m
  annotations:
    summary: "Evolution API no máximo de pods há 30 minutos"
```

---

## 🧪 Testes de Carga

### Teste 1: 1.000 Mensagens
```bash
# Gerar 1.000 mensagens
for i in {1..1000}; do
  kubectl exec -it deployment/redis-queue -n ifinu-production -- \
    redis-cli LPUSH ifinu:fila:whatsapp "{\"id\":\"test_$i\"}"
done

# Resultado esperado:
# - Tempo: ~20 segundos (50 msg/s × 2 pods = 100 msg/s)
# - Pods: 2 (não escala)
# - CPU: 40-50%
```

### Teste 2: 10.000 Mensagens
```bash
# Gerar 10.000 mensagens
for i in {1..10000}; do
  kubectl exec -it deployment/redis-queue -n ifinu-production -- \
    redis-cli LPUSH ifinu:fila:whatsapp "{\"id\":\"test_$i\"}"
done

# Resultado esperado:
# - Tempo: ~200 segundos iniciais, depois acelera com scaling
# - Pods: 2 → 4 → 6 (escala gradualmente)
# - CPU: 70%+ dispara scaling
# - Tempo total: ~100 segundos
```

### Teste 3: 100.000 Mensagens
```bash
# Gerar 100.000 mensagens
for i in {1..100000}; do
  kubectl exec -it deployment/redis-queue -n ifinu-production -- \
    redis-cli LPUSH ifinu:fila:whatsapp "{\"id\":\"test_$i\"}"
done

# Resultado esperado:
# - Tempo: ~15 minutos
# - Pods: 2 → 10 → 20 (escala até o máximo)
# - CPU: 80%+ sustentado
# - Throughput: 1.000 msg/s (50 msg/s × 20 pods)
```

---

## 🔧 Ajustes de Performance

### Aumentar Throughput

**Opção 1: Mais Workers por Pod**
```go
// cmd/api/main.go
filaMensagem.IniciarWorkerPool(20)  // 20 ao invés de 10
```
- **Impacto**: +100% throughput por pod
- **Custo**: +50% CPU por pod
- **Recomendado para**: Pods com mais CPU (2+ cores)

**Opção 2: Rate Limit Maior**
```go
// servico/fila_mensagem_servico.go
rate.NewLimiter(rate.Limit(100), 200)  // 100 msg/s ao invés de 50
```
- **Impacto**: +100% throughput por pod
- **Risco**: Evolution API pode bloquear
- **Recomendado**: Testar primeiro com 75 msg/s

**Opção 3: Mais Pods Mínimos**
```yaml
# k8s/evolution-api-deployment.yaml
minReplicas: 5  # 5 ao invés de 2
```
- **Impacto**: Sempre pronto para picos
- **Custo**: +150% custo mínimo
- **Recomendado para**: Workload constante

### Reduzir Latência

**Opção 1: Desabilitar Retry**
```go
// Para mensagens não críticas
MaxRetentativas = 1  // 1 ao invés de 3
```

**Opção 2: Retry Mais Rápido**
```go
// Para retry mais agressivo
TempoRetry = 1 * time.Minute  // 1 min ao invés de 5
```

**Opção 3: Priorização de Fila**
```go
// Usar múltiplas filas
FilaMensagensUrgente  = "ifinu:fila:urgente"   // Processada primeiro
FilaMensagensNormal   = "ifinu:fila:whatsapp"  // Processada depois
FilaMensagensBaixa    = "ifinu:fila:baixa"     // Processada quando ociosa
```

---

## 💰 Análise de Custos

### AWS EKS

| Cenário | Pods | CPU | Memória | Custo/Mês | Throughput |
|---------|------|-----|---------|-----------|------------|
| Mínimo | 2 | 1 vCPU | 1 GB | $50 | 100 msg/s |
| Médio | 10 | 10 vCPU | 10 GB | $250 | 500 msg/s |
| Máximo | 20 | 20 vCPU | 20 GB | $500 | 1000 msg/s |

**Economia vs Sempre Máximo:** 60-80% (escala sob demanda)

### Google GKE

| Cenário | Pods | CPU | Memória | Custo/Mês | Throughput |
|---------|------|-----|---------|-----------|------------|
| Mínimo | 2 | 1 vCPU | 1 GB | $40 | 100 msg/s |
| Médio | 10 | 10 vCPU | 10 GB | $200 | 500 msg/s |
| Máximo | 20 | 20 vCPU | 20 GB | $400 | 1000 msg/s |

**Economia vs Sempre Máximo:** 70-85% (escala sob demanda)

---

## 🚀 Próximas Melhorias

### Curto Prazo (1-2 semanas)
1. ✅ Worker Pool implementado
2. ✅ Rate Limiter configurado
3. ✅ Kubernetes HPA funcionando
4. ⏳ Métricas customizadas (Prometheus)
5. ⏳ Dashboard Grafana
6. ⏳ Alertas PagerDuty/Slack

### Médio Prazo (1-2 meses)
1. ⏳ Circuit Breaker para Evolution API
2. ⏳ Cache de instâncias WhatsApp (Redis)
3. ⏳ Fila de prioridade (urgente/normal/baixa)
4. ⏳ Webhook para status de entrega
5. ⏳ Dead Letter Queue (DLQ) para análise
6. ⏳ Backup automático PostgreSQL

### Longo Prazo (3+ meses)
1. ⏳ Multi-região (AWS/GCP)
2. ⏳ Disaster Recovery automático
3. ⏳ Machine Learning para predição de picos
4. ⏳ Otimização de custos automática
5. ⏳ API Gateway com rate limiting
6. ⏳ Observability completa (OpenTelemetry)

---

## 📚 Referências

- [Kubernetes HPA](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
- [Redis Queue Patterns](https://redis.io/docs/manual/patterns/distributed-locks/)
- [Go Rate Limiting](https://pkg.go.dev/golang.org/x/time/rate)
- [Evolution API Docs](https://doc.evolution-api.com/)
- [GORM Documentation](https://gorm.io/docs/)

---

## 🆘 Suporte

**Problemas Comuns:**

1. **Fila não processa**
   - Verificar: Redis conectado? Workers rodando?
   - Comando: `kubectl logs deployment/ifinu-api-go -n ifinu-production | grep Worker`

2. **Scaling não funciona**
   - Verificar: Metrics Server instalado?
   - Comando: `kubectl get apiservice v1beta1.metrics.k8s.io`

3. **Alta latência**
   - Verificar: Rate limiter muito baixo?
   - Solução: Aumentar de 50 para 75 msg/s

4. **Custo muito alto**
   - Verificar: Pods não fazem scale down?
   - Solução: Ajustar `stabilizationWindowSeconds`

---

**Sistema preparado para escala massiva! 🚀**
