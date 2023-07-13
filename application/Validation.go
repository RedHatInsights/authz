package application

import (
	"authz/domain"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var validatorInstance = initializeValidator()

// ValidateStruct performs validation on the provided event struct and returns true if the struct is valid, else it returns false and an error message object
func ValidateStruct(evt interface{}) error {
	err := validatorInstance.Struct(evt)

	if err != nil {
		errors, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}

		return domain.NewErrInvalidRequest(errors.Error())
	}

	return nil
}

func initializeValidator() *validator.Validate {
	vl := validator.New()
	err := vl.RegisterValidation("spicedb", validateSpiceDbID)
	if err != nil {
		panic(err)
	}
	return vl
}

var spiceDbIDPattern = regexp.MustCompile(`^(([a-zA-Z0-9/_|\-=+]{1,})|\*)$`)

func validateSpiceDbID(fl validator.FieldLevel) bool {
	return spiceDbIDPattern.MatchString(fl.Field().String())
}
