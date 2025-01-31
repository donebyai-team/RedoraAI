package csv

import (
	"github.com/go-playground/validator/v10"
)

func NewStructValidator() *validator.Validate {
	validator := validator.New()
	return validator
}
