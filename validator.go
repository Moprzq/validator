package validator

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var ErrNotStruct = errors.New("wrong argument given, should be a struct")
var ErrInvalidValidatorSyntax = errors.New("invalid validator syntax")
var ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")

type ValidationError struct {
	Err error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	errorsOut := make([]string, 0)
	for _, v := range v {
		errorsOut = append(errorsOut, v.Err.Error())
	}
	return strings.Join(errorsOut, "\n")
}

func NewValidationError(fieldName string, err error) ValidationError {
	var newError ValidationError
	if fieldName == "" {
		newError = ValidationError{Err: err}
	} else {
		newError = ValidationError{Err: errors.Wrap(err, fieldName)}
	}
	return newError
}

func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func Validate(v any) error {
	validateErrors := make(ValidationErrors, 0)
	vType := reflect.TypeOf(v)
	if vType.Kind() != reflect.Struct {
		return ErrNotStruct
	}
	for i := 0; i < vType.NumField(); i++ {
		f := vType.Field(i)
		if f.Tag != "" && !f.IsExported() {
			validateErrors = append(validateErrors, NewValidationError("", ErrValidateForUnexportedFields))
			continue
		}
		tagsParams := strings.Split(f.Tag.Get("validate"), " ")
		vValue := reflect.ValueOf(v)
		for _, param := range tagsParams {
			tagsValues := strings.Split(param, ":")
			if f.Type.String() == "string" {
				fieldValue := vValue.Field(i).String()
				if tagsValues[0] == "min" || tagsValues[0] == "max" || tagsValues[0] == "len" {
					tagIntValue, err := strconv.Atoi(tagsValues[1])
					if err != nil {
						validateErrors = append(validateErrors, NewValidationError("", ErrInvalidValidatorSyntax))
					} else if tagsValues[0] == "min" && len(fieldValue) < tagIntValue {
						validateErrors = append(validateErrors, NewValidationError(f.Name, ErrInvalidValidatorSyntax))
					} else if tagsValues[0] == "max" && len(fieldValue) > tagIntValue {
						validateErrors = append(validateErrors, NewValidationError(f.Name, ErrInvalidValidatorSyntax))
					} else if tagsValues[0] == "len" && len(fieldValue) != tagIntValue {
						validateErrors = append(validateErrors, NewValidationError(f.Name, ErrInvalidValidatorSyntax))
					}
				}
			}
		}
	}
	if len(validateErrors) != 0 {
		return validateErrors
	}
	return nil
}
