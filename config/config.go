package config

import (
	"log"

	"github.com/spf13/viper"
)

// CarregarConfiguracoes carrega as configurações do arquivo .env
func CarregarConfiguracoes() error {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	// Valores padrão
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_ENV", "production")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_SSL_MODE", "disable")
	viper.SetDefault("JWT_EXPIRATION_HOURS", 24)
	viper.SetDefault("JWT_REFRESH_EXPIRATION_DAYS", 7)
	viper.SetDefault("UPLOAD_DIR", "./uploads")
	viper.SetDefault("MAX_UPLOAD_SIZE_MB", 10)

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("⚠️  Arquivo .env não encontrado, usando valores padrão: %v", err)
		return nil // Não retorna erro, apenas usa os padrões
	}

	log.Println("✅ Configurações carregadas com sucesso")
	return nil
}
