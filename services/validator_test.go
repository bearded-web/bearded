package services

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/validator.v2"
	"testing"
)

func TestBsonIdValidator(t *testing.T) {

	type B struct {
		Id string `validate:"bsonId"`
	}
	type C struct {
		Id int `validate:"bsonId"`
	}

	tValidator := validator.NewValidator()
	tValidator.SetValidationFunc("bsonId", bsonIdValidation)

	testData := []struct {
		id     string
		errMap error
	}{
		{"", validator.ErrorMap{"Id": {BsonIdError}}},
		{"12345678901234567890123j", validator.ErrorMap{"Id": {BsonIdError}}},
		{"123456789012345678901234", nil},
	}

	for _, td := range testData {
		assert.Equal(t, td.errMap, tValidator.Validate(B{Id: td.id}))
	}

	c := C{}
	assert.Equal(t,
		validator.ErrorMap{"Id": {validator.ErrUnsupported}},
		tValidator.Validate(c))
}
