package application

import (
	"authz/domain"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/golang/glog"
)

var validatorInstance *validator.Validate

// ValidateEvent performs validation on the provided event struct and returns true if the struct is valid, else it returns false and an error message object
func ValidateEvent(evt interface{}) (bool, error) {
	if validatorInstance == nil {
		initializeValidator() //TODO: move to startup?
	}

	err := validatorInstance.Struct(evt)

	if err != nil {
		errors, ok := err.(validator.ValidationErrors)
		if !ok {
			glog.Errorf("Failed to validate message %+v. Error: %+v", evt, err)
			return false, nil
		}

		return false, domain.NewErrInvalidRequest(errors.Error())
	}

	return true, nil
}

func initializeValidator() {
	validatorInstance = validator.New()
	err := validatorInstance.RegisterValidation("spicedb", validateSpiceDbID)
	if err != nil {
		glog.Errorf("Failed to register SpiceDB ID validator! Err: %+v", err) //TODO: panic?
	}
}

var spiceDbIDPattern = regexp.MustCompile(`^(([a-zA-Z0-9/_|\-=+]{1,})|\*)$`)

func validateSpiceDbID(fl validator.FieldLevel) bool {
	return spiceDbIDPattern.MatchString(fl.Field().String())
}
