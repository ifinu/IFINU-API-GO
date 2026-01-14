package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/price"
	"github.com/stripe/stripe-go/v81/product"
)

func main() {
	// Carregar variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis do sistema")
	}

	// Configurar chave do Stripe
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		log.Fatal("❌ STRIPE_SECRET_KEY não configurada. Configure no arquivo .env")
	}
	stripe.Key = stripeKey

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║      IFINU - Setup Automático do Stripe               ║")
	fmt.Println("║      Criando Produtos e Prices                         ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Verificar se produtos já existem
	fmt.Println("🔍 Verificando produtos existentes...")
	existingProducts, err := product.List(&stripe.ProductListParams{
		Active: stripe.Bool(true),
	})
	if err != nil {
		log.Fatalf("❌ Erro ao listar produtos: %v", err)
	}

	var ifinuProduct *stripe.Product
	for existingProducts.Next() {
		p := existingProducts.Product()
		if p.Name == "IFINU - Sistema de Cobrança" {
			ifinuProduct = p
			fmt.Printf("✅ Produto já existe: %s (ID: %s)\n", p.Name, p.ID)
			break
		}
	}

	// Criar produto se não existir
	if ifinuProduct == nil {
		fmt.Println("\n📦 Criando produto IFINU...")
		productParams := &stripe.ProductParams{
			Name:        stripe.String("IFINU - Sistema de Cobrança"),
			Description: stripe.String("Plataforma de automação de cobranças com WhatsApp e Email. Gerencie cobranças recorrentes, envie lembretes automáticos e acompanhe pagamentos."),
		}

		ifinuProduct, err = product.New(productParams)
		if err != nil {
			log.Fatalf("❌ Erro ao criar produto: %v", err)
		}
		fmt.Printf("✅ Produto criado: %s (ID: %s)\n", ifinuProduct.Name, ifinuProduct.ID)
	}

	fmt.Println("\n💰 Criando prices para os 3 planos...\n")

	// Plano Mensal
	fmt.Println("1️⃣  Criando Plano Mensal (R$ 39/mês)...")
	priceMensal, err := price.New(&stripe.PriceParams{
		Product:    stripe.String(ifinuProduct.ID),
		Currency:   stripe.String("brl"),
		UnitAmount: stripe.Int64(3900), // R$ 39.00 em centavos
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String("month"),
		},
		Nickname: stripe.String("Plano Mensal"),
	})
	if err != nil {
		log.Printf("⚠️  Erro ao criar price mensal (pode já existir): %v\n", err)
	} else {
		fmt.Printf("   ✅ Price Mensal criado: %s\n", priceMensal.ID)
	}

	// Plano Trimestral
	fmt.Println("\n2️⃣  Criando Plano Trimestral (R$ 99 a cada 3 meses)...")
	priceTrimestral, err := price.New(&stripe.PriceParams{
		Product:    stripe.String(ifinuProduct.ID),
		Currency:   stripe.String("brl"),
		UnitAmount: stripe.Int64(9900), // R$ 99.00 em centavos
		Recurring: &stripe.PriceRecurringParams{
			Interval:      stripe.String("month"),
			IntervalCount: stripe.Int64(3),
		},
		Nickname: stripe.String("Plano Trimestral"),
	})
	if err != nil {
		log.Printf("⚠️  Erro ao criar price trimestral (pode já existir): %v\n", err)
	} else {
		fmt.Printf("   ✅ Price Trimestral criado: %s\n", priceTrimestral.ID)
	}

	// Plano Anual
	fmt.Println("\n3️⃣  Criando Plano Anual (R$ 348/ano)...")
	priceAnual, err := price.New(&stripe.PriceParams{
		Product:    stripe.String(ifinuProduct.ID),
		Currency:   stripe.String("brl"),
		UnitAmount: stripe.Int64(34800), // R$ 348.00 em centavos
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String("year"),
		},
		Nickname: stripe.String("Plano Anual"),
	})
	if err != nil {
		log.Printf("⚠️  Erro ao criar price anual (pode já existir): %v\n", err)
	} else {
		fmt.Printf("   ✅ Price Anual criado: %s\n", priceAnual.ID)
	}

	// Listar todos os prices do produto para exibir
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("\n📋 PRICES CRIADOS - Copie para o .env:\n")

	priceList, err := price.List(&stripe.PriceListParams{
		Product: stripe.String(ifinuProduct.ID),
		Active:  stripe.Bool(true),
	})
	if err != nil {
		log.Fatalf("❌ Erro ao listar prices: %v", err)
	}

	var mensalID, trimestralID, anualID string

	for priceList.Next() {
		p := priceList.Price()
		interval := "month"
		intervalCount := int64(1)
		if p.Recurring != nil {
			interval = string(p.Recurring.Interval)
			intervalCount = p.Recurring.IntervalCount
		}

		valor := float64(p.UnitAmount) / 100

		if interval == "month" && intervalCount == 1 {
			mensalID = p.ID
			fmt.Printf("STRIPE_PRICE_ID_MENSAL=%s\n", p.ID)
			fmt.Printf("   → R$ %.2f/mês\n\n", valor)
		} else if interval == "month" && intervalCount == 3 {
			trimestralID = p.ID
			fmt.Printf("STRIPE_PRICE_ID_TRIMESTRAL=%s\n", p.ID)
			fmt.Printf("   → R$ %.2f a cada 3 meses (R$ %.2f/mês)\n\n", valor, valor/3)
		} else if interval == "year" {
			anualID = p.ID
			fmt.Printf("STRIPE_PRICE_ID_ANUAL=%s\n", p.ID)
			fmt.Printf("   → R$ %.2f/ano (R$ %.2f/mês)\n\n", valor, valor/12)
		}
	}

	fmt.Println(strings.Repeat("═", 60))
	fmt.Println("\n✅ Setup concluído com sucesso!\n")

	// Verificar se algum price não foi criado
	if mensalID == "" || trimestralID == "" || anualID == "" {
		fmt.Println("⚠️  ATENÇÃO: Alguns prices podem não ter sido criados.")
		fmt.Println("   Verifique se já existem prices para este produto.")
		fmt.Println("   Acesse: https://dashboard.stripe.com/products/" + ifinuProduct.ID)
		fmt.Println()
	}

	fmt.Println("📝 Próximos passos:")
	fmt.Println("   1. Copie as variáveis acima para o arquivo .env do servidor")
	fmt.Println("   2. Reinicie a API: docker-compose restart api")
	fmt.Println("   3. Configure o webhook em: https://dashboard.stripe.com/webhooks")
	fmt.Println("      URL: https://api.ifinu.io/api/stripe/webhook")
	fmt.Println("      Eventos: checkout.session.completed, customer.subscription.*")
	fmt.Println()
}
