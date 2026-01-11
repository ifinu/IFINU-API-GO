package util

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

var (
	ErrTokenInvalido  = errors.New("token JWT inválido")
	ErrTokenExpirado  = errors.New("token JWT expirado")
	ErrTokenMalformado = errors.New("token JWT malformado")
)

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// GerarToken gera um access token JWT
func GerarToken(email string) (string, error) {
	secret := viper.GetString("JWT_SECRET")
	expiracaoHoras := viper.GetInt("JWT_EXPIRATION_HOURS")

	if expiracaoHoras == 0 {
		expiracaoHoras = 24 // padrão 24 horas
	}

	expiracao := time.Now().Add(time.Hour * time.Duration(expiracaoHoras))

	claims := &Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiracao),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "IFINU",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(secret))
}

// GerarRefreshToken gera um refresh token JWT (válido por mais tempo)
func GerarRefreshToken(email string) (string, error) {
	secret := viper.GetString("JWT_SECRET")
	expiracaoDias := viper.GetInt("JWT_REFRESH_EXPIRATION_DAYS")

	if expiracaoDias == 0 {
		expiracaoDias = 7 // padrão 7 dias
	}

	expiracao := time.Now().Add(time.Hour * 24 * time.Duration(expiracaoDias))

	claims := &Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiracao),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "IFINU",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(secret))
}

// ValidarToken valida e retorna as claims do token
func ValidarToken(tokenString string) (*Claims, error) {
	secret := viper.GetString("JWT_SECRET")

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verificar método de assinatura
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenMalformado
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpirado
		}
		return nil, ErrTokenInvalido
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalido
}

// ExtrairEmail extrai o email do token
func ExtrairEmail(tokenString string) (string, error) {
	claims, err := ValidarToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.Email, nil
}

// IsTokenExpirado verifica se o token está expirado
func IsTokenExpirado(tokenString string) bool {
	_, err := ValidarToken(tokenString)
	return errors.Is(err, ErrTokenExpirado)
}
