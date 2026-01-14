# Componente Frontend - Stripe Connect

## 📦 Componente: StripeConnectManager.tsx

Crie este arquivo em: `app/components/StripeConnectManager.tsx`

```typescript
'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { ExternalLink, CheckCircle2, AlertCircle, Loader2, CreditCard } from 'lucide-react'
import { apiUrl } from '@/lib/api'
import { toast } from 'sonner'

interface StripeConnectStatus {
  conectado: boolean
  accountId?: string
  onboardingCompleto: boolean
  chargesHabilitado: boolean
  detalhesSubmetidos: boolean
  precisaAcao: boolean
  dashboardUrl?: string
}

export default function StripeConnectManager() {
  const [status, setStatus] = useState<StripeConnectStatus | null>(null)
  const [carregando, setCarregando] = useState(true)
  const [criandoConta, setCriandoConta] = useState(false)
  const [gerandoDashboard, setGerandoDashboard] = useState(false)

  useEffect(() => {
    carregarStatus()
  }, [])

  const carregarStatus = async () => {
    try {
      const token = localStorage.getItem('token')
      const response = await fetch(apiUrl('/api/stripe-connect/status'), {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })

      if (response.ok) {
        const data = await response.json()
        setStatus(data.data)
      }
    } catch (error) {
      console.error('Erro ao carregar status:', error)
    } finally {
      setCarregando(false)
    }
  }

  const conectarStripe = async () => {
    setCriandoConta(true)
    try {
      const token = localStorage.getItem('token')
      const response = await fetch(apiUrl('/api/stripe-connect/criar-conta'), {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      })

      const data = await response.json()

      if (response.ok && data.data.onboardingUrl) {
        // Redirecionar para onboarding do Stripe
        window.location.href = data.data.onboardingUrl
      } else {
        toast.error(data.message || 'Erro ao criar conta Stripe')
      }
    } catch (error) {
      toast.error('Erro ao conectar com Stripe')
    } finally {
      setCriandoConta(false)
    }
  }

  const acessarDashboard = async () => {
    setGerandoDashboard(true)
    try {
      const token = localStorage.getItem('token')
      const response = await fetch(apiUrl('/api/stripe-connect/dashboard-link'), {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })

      const data = await response.json()

      if (response.ok && data.data.url) {
        // Abrir dashboard em nova aba
        window.open(data.data.url, '_blank')
      } else {
        toast.error('Erro ao gerar link do dashboard')
      }
    } catch (error) {
      toast.error('Erro ao acessar dashboard')
    } finally {
      setGerandoDashboard(false)
    }
  }

  const continuarOnboarding = async () => {
    setCriandoConta(true)
    try {
      const token = localStorage.getItem('token')
      const response = await fetch(apiUrl('/api/stripe-connect/refresh-onboarding'), {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      })

      const data = await response.json()

      if (response.ok && data.data.onboardingUrl) {
        window.location.href = data.data.onboardingUrl
      } else {
        toast.error('Erro ao gerar link de onboarding')
      }
    } catch (error) {
      toast.error('Erro ao continuar onboarding')
    } finally {
      setCriandoConta(false)
    }
  }

  if (carregando) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    )
  }

  // Não conectado
  if (!status?.conectado) {
    return (
      <Card>
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-100 rounded-lg">
              <CreditCard className="h-6 w-6 text-blue-600" />
            </div>
            <div>
              <CardTitle>Receber Pagamentos com Stripe</CardTitle>
              <CardDescription>
                Configure sua conta para receber pagamentos dos seus clientes
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <Alert>
            <AlertDescription>
              <strong>Como funciona:</strong>
              <ul className="mt-2 space-y-1 text-sm list-disc list-inside">
                <li>Você receberá os pagamentos diretamente na sua conta Stripe</li>
                <li>O IFINU não toca no seu dinheiro</li>
                <li>Configuração rápida em 5 minutos</li>
                <li>Totalmente seguro e regulamentado</li>
              </ul>
            </AlertDescription>
          </Alert>

          <Button
            onClick={conectarStripe}
            disabled={criandoConta}
            className="w-full"
            size="lg"
          >
            {criandoConta ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Conectando...
              </>
            ) : (
              <>
                <CreditCard className="mr-2 h-4 w-4" />
                Conectar com Stripe
              </>
            )}
          </Button>

          <p className="text-xs text-muted-foreground text-center">
            Você será redirecionado para o Stripe para completar o cadastro
          </p>
        </CardContent>
      </Card>
    )
  }

  // Conectado mas onboarding incompleto
  if (status.conectado && !status.onboardingCompleto) {
    return (
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-yellow-100 rounded-lg">
                <AlertCircle className="h-6 w-6 text-yellow-600" />
              </div>
              <div>
                <CardTitle>Complete seu Cadastro</CardTitle>
                <CardDescription>
                  Finalize o cadastro para começar a receber pagamentos
                </CardDescription>
              </div>
            </div>
            <Badge variant="secondary">Pendente</Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Seu cadastro no Stripe está incompleto. Complete as informações para poder receber pagamentos.
            </AlertDescription>
          </Alert>

          <div className="space-y-2 text-sm">
            <div className="flex items-center gap-2">
              {status.detalhesSubmetidos ? (
                <CheckCircle2 className="h-4 w-4 text-green-600" />
              ) : (
                <AlertCircle className="h-4 w-4 text-yellow-600" />
              )}
              <span>Informações pessoais/empresariais</span>
            </div>
            <div className="flex items-center gap-2">
              {status.chargesHabilitado ? (
                <CheckCircle2 className="h-4 w-4 text-green-600" />
              ) : (
                <AlertCircle className="h-4 w-4 text-yellow-600" />
              )}
              <span>Verificação aprovada pelo Stripe</span>
            </div>
          </div>

          <Button
            onClick={continuarOnboarding}
            disabled={criandoConta}
            className="w-full"
            size="lg"
          >
            {criandoConta ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Carregando...
              </>
            ) : (
              <>
                <ExternalLink className="mr-2 h-4 w-4" />
                Continuar Cadastro no Stripe
              </>
            )}
          </Button>
        </CardContent>
      </Card>
    )
  }

  // Conectado e onboarding completo
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-green-100 rounded-lg">
              <CheckCircle2 className="h-6 w-6 text-green-600" />
            </div>
            <div>
              <CardTitle>Stripe Conectado</CardTitle>
              <CardDescription>
                Sua conta está pronta para receber pagamentos
              </CardDescription>
            </div>
          </div>
          <Badge className="bg-green-600">Ativo</Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <Alert className="border-green-200 bg-green-50">
          <CheckCircle2 className="h-4 w-4 text-green-600" />
          <AlertDescription className="text-green-800">
            Tudo pronto! Seus clientes já podem fazer pagamentos que cairão direto na sua conta Stripe.
          </AlertDescription>
        </Alert>

        <div className="space-y-2 text-sm bg-muted p-4 rounded-lg">
          <p><strong>Account ID:</strong> {status.accountId}</p>
          <div className="flex items-center gap-2">
            <CheckCircle2 className="h-4 w-4 text-green-600" />
            <span>Pagamentos habilitados</span>
          </div>
          <div className="flex items-center gap-2">
            <CheckCircle2 className="h-4 w-4 text-green-600" />
            <span>Verificação completa</span>
          </div>
        </div>

        <Button
          onClick={acessarDashboard}
          disabled={gerandoDashboard}
          variant="outline"
          className="w-full"
        >
          {gerandoDashboard ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Gerando link...
            </>
          ) : (
            <>
              <ExternalLink className="mr-2 h-4 w-4" />
              Acessar Dashboard do Stripe
            </>
          )}
        </Button>

        <p className="text-xs text-muted-foreground text-center">
          Gerencie pagamentos, saques e configurações no dashboard do Stripe
        </p>
      </CardContent>
    </Card>
  )
}
```

## 🎨 Como Integrar

### 1. Na página de Configurações

Adicione na página `app/painel/configuracoes/page.tsx`:

```typescript
import StripeConnectManager from '@/app/components/StripeConnectManager'

// Dentro do componente, adicione uma nova tab:
<TabsContent value="pagamentos">
  <StripeConnectManager />
</TabsContent>
```

### 2. Ou crie uma página específica

Crie `app/painel/pagamentos/page.tsx`:

```typescript
import StripeConnectManager from '@/app/components/StripeConnectManager'

export default function PagamentosPage() {
  return (
    <div className="container max-w-4xl py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Configurar Pagamentos</h1>
        <p className="text-muted-foreground">
          Configure sua conta Stripe para receber pagamentos dos seus clientes
        </p>
      </div>

      <StripeConnectManager />
    </div>
  )
}
```

## 🔄 Fluxo Completo

### 1. Usuário SEM conta conectada

```
[Card de onboarding]
├─ Ícone do Stripe
├─ Título: "Receber Pagamentos com Stripe"
├─ Explicação: Como funciona
└─ Botão: "Conectar com Stripe" → Redireciona para Stripe
```

### 2. Usuário COM conta mas incompleta

```
[Card de ação necessária]
├─ Badge: "Pendente"
├─ Alert: Cadastro incompleto
├─ Checklist:
│  ├─ ❌ Informações pessoais
│  └─ ❌ Verificação aprovada
└─ Botão: "Continuar Cadastro" → Redireciona para Stripe
```

### 3. Usuário COM conta completa

```
[Card de sucesso]
├─ Badge: "Ativo"
├─ Alert: Tudo pronto
├─ Detalhes:
│  ├─ Account ID
│  ├─ ✅ Pagamentos habilitados
│  └─ ✅ Verificação completa
└─ Botão: "Acessar Dashboard" → Abre dashboard em nova aba
```

## 🎯 Recursos do Componente

- ✅ Auto-detecta status da conta
- ✅ Cria conta automaticamente
- ✅ Redireciona para onboarding
- ✅ Mostra progresso do cadastro
- ✅ Link para dashboard Stripe
- ✅ Estados de loading
- ✅ Toasts de feedback
- ✅ Design responsivo
- ✅ Ícones lucide-react
- ✅ Shadcn UI components

## 📱 Screenshots (Estado Visual)

### Não Conectado
- Card azul com ícone de cartão
- Botão azul "Conectar com Stripe"
- Alert informativo

### Incompleto
- Card amarelo com ícone de alerta
- Badge "Pendente"
- Checklist de tarefas
- Botão "Continuar Cadastro"

### Conectado
- Card verde com ícone de check
- Badge "Ativo"
- Detalhes da conta
- Botão "Acessar Dashboard"

## 🔧 Dependências Necessárias

Certifique-se de ter instalado:

```bash
npm install lucide-react sonner
```

E que os componentes Shadcn UI estão configurados:
- Button
- Card
- Badge
- Alert

## 🚀 Pronto!

O componente está completo e pronto para uso. Basta criar o arquivo e integrar na página de configurações ou criar uma página específica de pagamentos.
