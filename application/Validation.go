package application

import (
	"authz/domain"
	"regexp"
	"strings"

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

	err = vl.RegisterValidation("in", validateIn)
	if err != nil {
		panic(err)
	}

	vl.RegisterAlias("service", `spicedb-id,max=36`)
	vl.RegisterAlias("identifier", `spicedb-id,max=36`) //identifiers must be valid SpiceDB identifiers /and/ either UUIDs or positive integers or a test value

	return vl
}

var spiceDbIDPattern = regexp.MustCompile(`^(([a-zA-Z0-9/_|\-=+]{1,})|\*)$`)
var spiceDbPermissionPattern = regexp.MustCompile(`^([a-z][a-z0-9_]{1,62}[a-z0-9])?$`)
var spiceDbTypePattern = regexp.MustCompile(`^([a-z][a-z0-9_]{1,61}[a-z0-9]/)?[a-z][a-z0-9_]{1,62}[a-z0-9]$`)

func validateSpiceDbID(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}

	return spiceDbIDPattern.MatchString(val)
}

func validateSpiceDbPermission(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}

	return spiceDbPermissionPattern.MatchString(fl.Field().String())
}

func validateSpiceDbType(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}

	return spiceDbTypePattern.MatchString(fl.Field().String())
}

var inConfigCache = make(map[string]map[string]bool) //Map string -> idiomatic Set<string>
func validateIn(fl validator.FieldLevel) bool {
	param := fl.Param()

	config, ok := inConfigCache[param]
	if !ok {
		config = make(map[string]bool)

		for _, val := range strings.Split(param, "+") {
			config[val] = true
		}

		inConfigCache[param] = config
	}

	val := fl.Field().String()
	if val == "" {
		return true
	}

	return config[fl.Field().String()]
}
