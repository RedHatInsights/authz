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

	err := vl.RegisterValidation("spicedb-id", validateSpiceDbID)
	if err != nil {
		panic(err)
	}

	err = vl.RegisterValidation("spicedb-permission", validateSpiceDbPermission)
	if err != nil {
		panic(err)
	}

	err = vl.RegisterValidation("spicedb-type", validateSpiceDbType)
	if err != nil {
		panic(err)
	}

	return vl
}

var spiceDbIDPattern = regexp.MustCompile(`^(([a-zA-Z0-9/_|\-=+]{1,})|\*)$`)
var spiceDbPermissionPattern = regexp.MustCompile(`^([a-z][a-z0-9_]{1,62}[a-z0-9])?$`)
var spiceDbTypePattern = regexp.MustCompile(`^([a-z][a-z0-9_]{1,61}[a-z0-9]/)?[a-z][a-z0-9_]{1,62}[a-z0-9]$`)

func validateSpiceDbID(fl validator.FieldLevel) bool {
	return spiceDbIDPattern.MatchString(fl.Field().String())
}

func validateSpiceDbPermission(fl validator.FieldLevel) bool {
	return spiceDbPermissionPattern.MatchString(fl.Field().String())
}

func validateSpiceDbType(fl validator.FieldLevel) bool {
	return spiceDbTypePattern.MatchString(fl.Field().String())
}
