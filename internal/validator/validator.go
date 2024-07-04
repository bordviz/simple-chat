package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func Validate(model interface{}) string {
	var errMsgs []string

	validate := validator.New()
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		return field.Tag.Get("json")
	})

	err := validate.Struct(model)

	if err != nil {
		validErr := err.(validator.ValidationErrors)
		for _, errMsg := range validErr {
			switch errMsg.ActualTag() {
			case "required":
				errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required", errMsg.Field()))
			case "min":
				errMsgs = append(errMsgs, fmt.Sprintf("field %s must be at least %s", errMsg.Field(), errMsg.Param()))
			case "email":
				errMsgs = append(errMsgs, fmt.Sprintf("field %s must be a valid email", errMsg.Field()))
			default:
				errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", errMsg.Field()))
			}
		}
		return strings.Join(errMsgs, ", ")
	}
	return ""
}
