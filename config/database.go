package config

import (
	"fmt"
	"log"

	"github.com/ifinu/ifinu-api-go/dominio/entidades"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConectarBancoDados estabelece conexão com PostgreSQL
func ConectarBancoDados() error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("DB_HOST"),
		viper.GetString("DB_PORT"),
		viper.GetString("DB_USER"),
		viper.GetString("DB_PASSWORD"),
		viper.GetString("DB_NAME"),
		viper.GetString("DB_SSL_MODE"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("erro ao conectar ao banco de dados: %w", err)
	}

	log.Println("✅ Conectado ao banco de dados PostgreSQL")

	// Auto-migrate das tabelas
	if err := MigrarTabelas(); err != nil {
		return err
	}

	return nil
}

// MigrarTabelas faz auto-migration das entidades
func MigrarTabelas() error {
	log.Println("🔄 Executando migrations...")

	err := DB.AutoMigrate(
		&entidades.Usuario{},
		&entidades.Cliente{},
		&entidades.Cobranca{},
		&entidades.AssinaturaUsuario{},
		&entidades.WhatsAppConexao{},
	)

	if err != nil {
		return fmt.Errorf("erro ao migrar tabelas: %w", err)
	}

	log.Println("✅ Migrations executadas com sucesso")
	return nil
}

// ObterDB retorna a instância do banco de dados
func ObterDB() *gorm.DB {
	return DB
}
