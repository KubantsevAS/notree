package validator

import (
	"encoding/json"

	playground "github.com/go-playground/validator/v10"
)

var Validate *playground.Validate

func init() {
	Validate = playground.New()

	Validate.RegisterValidation("json_object", JSONObject)
}

func JSONObject(fl playground.FieldLevel) bool {
	data := fl.Field().Bytes()
	if len(data) == 0 {
		return true
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return false
	}

	_, ok := v.(map[string]interface{})
	return ok
}

func Check[T any](payload T) error {
	if err := Validate.Struct(payload); err != nil {
		return err
	}
	return nil
}
