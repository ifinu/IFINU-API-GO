package util

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost define o custo do BCrypt (10 = bom equilíbrio segurança/performance)
	BcryptCost = 10
)

// HashSenha faz hash da senha usando BCrypt
func HashSenha(senha string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(senha), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerificarSenha verifica se a senha corresponde ao hash
func VerificarSenha(senha, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha))
	return err == nil
}
