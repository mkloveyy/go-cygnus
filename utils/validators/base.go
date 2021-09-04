/*
Gin use gopkg.in/go-playground/validator.v8 to validate requests as default.

As definition of a custom validator is absolutely interminable in
validator.v8, we upgrade to validator.v9 and add our custom tags.

The upgrade approach is given by validator's authors.
Ref https://github.com/go-playground/validator/tree/v9/_examples/gin-upgrading-overriding
*/

package validators

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/iancoleman/strcase"
	"gopkg.in/go-playground/validator.v9"
)

// DefaultValidator validator v8 -> v9
type DefaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

func kindOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()

	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}

	return valueType
}

func (v *DefaultValidator) ValidateStruct(obj interface{}) error {
	if kindOfData(obj) == reflect.Struct {
		v.lazyinit()

		if err := v.validate.Struct(obj); err != nil {
			return err
		}
	}

	return nil
}

func (v *DefaultValidator) Engine() interface{} {
	v.lazyinit()
	return v.validate
}

func (v *DefaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("binding")

		// add any custom validations etc. here
		_ = v.validate.RegisterValidation("lowercase", ValidateLowercase)
		_ = v.validate.RegisterValidation("uppercase", ValidateUppercase)
		_ = v.validate.RegisterValidation("capitalized", ValidateCapitalized)
		_ = v.validate.RegisterValidation("whitespace", ValidateWhitespace)
		_ = v.validate.RegisterValidation("commaItemIn", ValidateCommaListItemIn)
	})
}

// ValidatorErrorFormatter error formatter
func ValidatorErrorFormatter(err error) (str string) {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return
	}

	for i, err := range validationErrors {
		errField := strcase.ToSnake(err.Field())

		switch err.Tag() {
		case "required":
			str += fmt.Sprintf("'%s' is required", errField)
		case "lowercase":
			str += fmt.Sprintf("'%s' must be lowercase", errField)
		case "uppercase":
			str += fmt.Sprintf("'%s' must be uppercase", errField)
		case "capitalize":
			str += fmt.Sprintf("'%s' must be capitalized", errField)
		case "whitespace":
			str += fmt.Sprintf("'%s' cannot have whitespace", errField)
		case "min":
			str += fmt.Sprintf("'%s' must be greater than or equal to %s", errField, err.Param())
		case "max":
			str += fmt.Sprintf("'%s' must be less than or equal to %s", errField, err.Param())
		case "endswith":
			str += fmt.Sprintf("'%s' must end with %s", errField, err.Param())
		case "oneof":
			sli := strings.Fields(err.Param())
			str += fmt.Sprintf("'%s' must be one of %s", errField, strings.Join(sli, " / "))
		case "commaItemIn":
			str += fmt.Sprintf("'%s' csv format item must be one of %s", errField, err.Param())
		default:
			str += fmt.Sprintf("'%s' not satisfy tag: %s", errField, err.ActualTag())
		}

		if i < len(validationErrors)-1 {
			str += ", "
		}
	}

	return str
}

// ValidateUppercase validators functions
func ValidateUppercase(fl validator.FieldLevel) bool {
	field := fl.Field().String()

	return field == strings.ToUpper(field)
}

func ValidateLowercase(fl validator.FieldLevel) bool {
	field := fl.Field().String()

	return field == strings.ToLower(field)
}

func ValidateCapitalized(fl validator.FieldLevel) bool {
	field := fl.Field().String()

	return field == strings.Title(field)
}

func ValidateWhitespace(fl validator.FieldLevel) bool {
	field := fl.Field().String()

	return !strings.Contains(field, " ")
}

func ValidateCommaListItemIn(fl validator.FieldLevel) bool {
	fieldVal := fl.Field().String()

	if fieldVal == "" {
		return true
	}

	oneOf := make(map[string]interface{})

	for _, available := range strings.Fields(fl.Param()) {
		oneOf[available] = nil
	}

	for _, v := range strings.Split(fieldVal, ",") {
		if _, ok := oneOf[v]; !ok {
			return false
		}
	}

	return true
}
