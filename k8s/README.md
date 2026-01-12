# Deploy Kubernetes - IFINU API

Sistema de automação de cobrança com auto-scaling baseado em demanda.

## Arquitetura

```
┌─────────────────────────────────────────────────────────┐
│                    KUBERNETES CLUSTER                    │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │        Evolution API Pods (2-20 replicas)        │  │
│  │  ┌────┐  ┌────┐  ┌────┐       ┌────┐  ┌────┐  │  │
│  │  │Pod1│  │Pod2│  │Pod3│  ...  │PodN│  │PodN│  │  │
│  │  └────┘  └────┘  └────┘       └────┘  └────┘  │  │
│  │           ↑                                      │  │
│  │           │ Load Balancer (Service)              │  │
│  └───────────┼──────────────────────────────────────┘  │
│              │                                          │
│  ┌───────────┼──────────────────────────────────────┐  │
│  │           │      IFINU API GO                     │  │
│  │           ↓                                       │  │
│  │  ┌─────────────────────────────────┐            │  │
│  │  │   Worker Pool (10 workers)      │            │  │
│  │  │   Rate Limiter: 50 msg/s        │            │  │
│  │  └─────────────┬───────────────────┘            │  │
│  │                │                                 │  │
│  │                ↓                                 │  │
│  │  ┌─────────────────────────────────┐            │  │
│  │  │      Redis Queue                │            │  │
│  │  │  (Fila de Mensagens)            │            │  │
│  │  └─────────────────────────────────┘            │  │
│  └──────────────────────────────────────────────────┘  │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │         PostgreSQL (StatefulSet)                  │  │
│  │  - Evolution Data                                 │  │
│  │  - PersistentVolume: 20Gi                        │  │
│  └──────────────────────────────────────────────────┘  │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │    HPA (Horizontal Pod Autoscaler)                │  │
│  │  - Min: 2 pods                                    │  │
│  │  - Max: 20 pods                                   │  │
│  │  - CPU: 70%                                       │  │
│  │  - Memory: 80%                                    │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## Funcionalidades

### 1. Auto-Scaling Inteligente
- **Mínimo**: 2 pods sempre rodando (alta disponibilidade)
- **Máximo**: 20 pods em picos de demanda
- **Gatilhos**: CPU > 70% ou Memória > 80%
- **Scale Up**: +50% ou +2 pods (o que for maior) a cada 60s
- **Scale Down**: -10% ou -1 pod (o que for menor) após 5 minutos

### 2. Worker Pool com Rate Limiting
- **Workers**: 10 workers paralelos por pod
- **Rate Limit**: 50 mensagens/segundo por pod
- **Capacidade Total**: 1.000 msg/s com 20 pods (50 msg/s × 20)

### 3. Fila Distribuída (Redis)
- **Retry Automático**: Até 3 tentativas com backoff exponencial
- **Persistência**: Mensagens não processadas sobrevivem a reinicializações
- **DLQ**: Dead Letter Queue para mensagens falhas

### 4. Alta Disponibilidade
- **Load Balancer**: Distribui requisições entre pods
- **Session Affinity**: Mantém sessões no mesmo pod por 1 hora
- **Health Checks**: Liveness e Readiness probes

## Deploy

### Pré-requisitos

```bash
# Cluster Kubernetes rodando
kubectl cluster-info

# Metrics Server instalado (para HPA)
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Namespace
kubectl create namespace ifinu-production
```

### 1. Deploy PostgreSQL

```bash
kubectl apply -f evolution-storage.yaml
kubectl apply -f evolution-secrets.yaml
kubectl apply -f evolution-postgres.yaml

# Aguardar PostgreSQL estar pronto
kubectl wait --for=condition=ready pod -l app=evolution-postgres -n ifinu-production --timeout=300s
```

### 2. Deploy Redis Queue

```bash
kubectl apply -f redis-queue.yaml

# Aguardar Redis estar pronto
kubectl wait --for=condition=ready pod -l app=redis-queue -n ifinu-production --timeout=120s
```

### 3. Deploy Evolution API com Auto-Scaling

```bash
kubectl apply -f evolution-api-deployment.yaml

# Verificar status
kubectl get pods -n ifinu-production -l app=evolution-api
kubectl get hpa -n ifinu-production
```

### 4. Verificar Deployment

```bash
# Status dos pods
kubectl get pods -n ifinu-production -w

# Logs da Evolution API
kubectl logs -f deployment/evolution-api -n ifinu-production

# Métricas do HPA
kubectl get hpa evolution-api-hpa -n ifinu-production --watch

# Eventos de scaling
kubectl get events -n ifinu-production --sort-by='.lastTimestamp'
```

## Monitoramento

### Métricas em Tempo Real

```bash
# CPU e Memória dos pods
kubectl top pods -n ifinu-production

# Detalhes do HPA
kubectl describe hpa evolution-api-hpa -n ifinu-production

# Tamanho da fila Redis
kubectl exec -it deployment/redis-queue -n ifinu-production -- redis-cli LLEN ifinu:fila:whatsapp
```

### Dashboards Recomendados

1. **Grafana + Prometheus**
   - Taxa de mensagens/segundo
   - Tempo de processamento
   - Taxa de sucesso/falha
   - Número de pods ativos

2. **Kubernetes Dashboard**
   ```bash
   kubectl proxy
   # Acessar: http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/
   ```

## Testes de Carga

### Simular Alta Demanda

```bash
# Gerar 1000 mensagens na fila
for i in {1..1000}; do
  kubectl exec -it deployment/redis-queue -n ifinu-production -- \
    redis-cli LPUSH ifinu:fila:whatsapp "{\"id\":\"test_$i\",\"tipo_notificacao\":\"lembrete\"}"
done

# Observar scaling automático
kubectl get hpa evolution-api-hpa -n ifinu-production --watch
```

### Simular Falha de Pod

```bash
# Deletar um pod (Kubernetes recria automaticamente)
kubectl delete pod -n ifinu-production -l app=evolution-api --field-selector=status.phase=Running | head -1

# Verificar recriação
kubectl get pods -n ifinu-production -w
```

## Configuração de Produção

### Ajustar Limites de Recursos

Editar `evolution-api-deployment.yaml`:

```yaml
resources:
  requests:
    memory: "1Gi"    # Aumentar para workload pesado
    cpu: "1000m"
  limits:
    memory: "4Gi"    # Limite máximo
    cpu: "4000m"
```

### Ajustar Auto-Scaling

```yaml
spec:
  minReplicas: 5     # Aumentar mínimo para sempre ter capacidade
  maxReplicas: 50    # Aumentar máximo para picos extremos
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        averageUtilization: 60  # Escalar mais cedo (60% ao invés de 70%)
```

### Ajustar Worker Pool

No código Go (`cmd/api/main.go`):

```go
// Aumentar número de workers
filaMensagem.IniciarWorkerPool(20)  // 20 workers ao invés de 10

// Ajustar rate limiter
rate.NewLimiter(rate.Limit(100), 200)  // 100 msg/s ao invés de 50
```

## Troubleshooting

### Pods não estão escalando

```bash
# Verificar metrics-server
kubectl get apiservice v1beta1.metrics.k8s.io -o yaml

# Instalar se necessário
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

### Fila Redis cheia

```bash
# Ver tamanho da fila
kubectl exec -it deployment/redis-queue -n ifinu-production -- redis-cli LLEN ifinu:fila:whatsapp

# Limpar fila (CUIDADO!)
kubectl exec -it deployment/redis-queue -n ifinu-production -- redis-cli DEL ifinu:fila:whatsapp
```

### Evolution API não conecta ao PostgreSQL

```bash
# Testar conectividade
kubectl exec -it deployment/evolution-api -n ifinu-production -- sh
apk add postgresql-client
psql "postgresql://evolution:evolution123@evolution-postgres-service:5432/evolution"
```

### Alta latência nas mensagens

```bash
# Aumentar workers e rate limiter no código
# OU
# Aumentar número mínimo de pods
kubectl patch hpa evolution-api-hpa -n ifinu-production --type='json' -p='[{"op": "replace", "path": "/spec/minReplicas", "value":5}]'
```

## Custos Estimados

### AWS EKS (exemplo)

- **2 pods mínimos**: ~$50/mês
- **20 pods no pico**: ~$500/mês
- **PostgreSQL RDS**: ~$100/mês
- **LoadBalancer**: ~$20/mês
- **Total médio**: ~$150-200/mês

### Google GKE (exemplo)

- **2 pods mínimos**: ~$40/mês
- **20 pods no pico**: ~$400/mês
- **Cloud SQL**: ~$80/mês
- **Load Balancer**: ~$15/mês
- **Total médio**: ~$120-180/mês

## Próximos Passos

1. ✅ Deploy básico funcionando
2. ✅ Auto-scaling configurado
3. ✅ Worker pool implementado
4. ⏳ Métricas customizadas (Prometheus)
5. ⏳ Alertas (PagerDuty/Slack)
6. ⏳ Backup automatizado PostgreSQL
7. ⏳ CI/CD pipeline (GitHub Actions)
8. ⏳ Disaster Recovery plan

## Suporte

Para problemas ou dúvidas:
- Logs: `kubectl logs -f deployment/evolution-api -n ifinu-production`
- Events: `kubectl get events -n ifinu-production --sort-by='.lastTimestamp'`
- Describe: `kubectl describe pod <pod-name> -n ifinu-production`
