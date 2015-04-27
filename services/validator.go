package services

import (
	"encoding/hex"
	"fmt"
	"gopkg.in/validator.v2"
)

var (
	BsonIdError = validator.TextErr{Err: fmt.Errorf("should be bson uuid in hex form")}
)

func bsonIdValidation(v interface{}, param string) error {
	s, casted := v.(string)
	if !casted {
		return validator.ErrUnsupported
	}
	if len(s) != 24 {
		return BsonIdError
	}
	_, err := hex.DecodeString(s)
	if err != nil {
		return BsonIdError
	}
	return nil
}

func init() {
	validator.SetValidationFunc("bsonId", bsonIdValidation)
}
