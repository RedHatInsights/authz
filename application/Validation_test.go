package application

import (
	"authz/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

type validationTestCase struct {
	ValueIn    string `validate:"in=valid+alsovalid"`
	Identifier string `validate:"identifier"`
	Service    string `validate:"service"`
	expected   bool
}

func TestValidateValue(t *testing.T) {
	tests := []validationTestCase{
		{ValueIn: "valid", expected: true},
		{ValueIn: "alsovalid", expected: true},
		{ValueIn: "", expected: true},
		{ValueIn: "invalid", expected: false},

		{Identifier: "!!!", expected: false},
		{Identifier: "12345678", expected: true},
		{Identifier: "2c4111de-2a6a-11ee-b0d5-e7859b698bcf", expected: true},
		{Identifier: "2c4111de-2a6a-11ee-b0d5-e7859b698bcf+", expected: false},

		{Service: "smarts", expected: true},
		{Service: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", expected: true},
		{Service: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa+", expected: false},
		{Service: "!!!", expected: false},
	}

	for _, test := range tests {
		err := ValidateStruct(test)
		if test.expected {
			assert.NoError(t, err, "Unexpected error. Case: %+v, err: %s", test, err)
		} else {
			var validationErr domain.ErrInvalidRequest
			if assert.Error(t, err, "Expected validation error but got pass. Case: %+v", test) {
				if assert.ErrorAs(t, err, &validationErr, "Expected validation error, got something else. Case: %+v, err: %s", test, err) {
					assert.NotEqual(t, "", validationErr.Reason, "No reason given for validation error. Case: %+v, err: %s", test, err)
				}
			}
		}
	}
}
