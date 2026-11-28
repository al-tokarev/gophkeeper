package validation

import (
	"regexp"

	"github.com/go-playground/validator"
)

func New() *validator.Validate {
	v := validator.New()

	v.RegisterValidation("expiry", func(fl validator.FieldLevel) bool {
		matched, _ := regexp.MatchString(`^(0[1-9]|1[0-2])/\d{2}$`, fl.Field().String())
		return matched
	})

	return v
}
