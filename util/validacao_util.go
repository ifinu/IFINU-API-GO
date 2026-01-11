package util

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidarStruct valida uma struct usando tags validate
func ValidarStruct(s interface{}) error {
	err := validate.Struct(s)
	if err != nil {
		// Formatar erros de validação
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var erros []string
			for _, e := range validationErrors {
				erros = append(erros, formatarErroValidacao(e))
			}
			return fmt.Errorf(strings.Join(erros, "; "))
		}
		return err
	}
	return nil
}

func formatarErroValidacao(e validator.FieldError) string {
	campo := e.Field()
	tag := e.Tag()

	switch tag {
	case "required":
		return fmt.Sprintf("%s é obrigatório", campo)
	case "email":
		return fmt.Sprintf("%s deve ser um email válido", campo)
	case "min":
		return fmt.Sprintf("%s deve ter no mínimo %s caracteres", campo, e.Param())
	case "max":
		return fmt.Sprintf("%s deve ter no máximo %s caracteres", campo, e.Param())
	case "gt":
		return fmt.Sprintf("%s deve ser maior que %s", campo, e.Param())
	case "gte":
		return fmt.Sprintf("%s deve ser maior ou igual a %s", campo, e.Param())
	case "lt":
		return fmt.Sprintf("%s deve ser menor que %s", campo, e.Param())
	case "lte":
		return fmt.Sprintf("%s deve ser menor ou igual a %s", campo, e.Param())
	default:
		return fmt.Sprintf("%s não passou na validação %s", campo, tag)
	}
}
