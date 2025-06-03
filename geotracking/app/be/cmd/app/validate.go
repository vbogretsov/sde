package main

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ValidationErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Errors []ValidationErrorResponse `json:"errors"`
}

func (e ErrorResponse) Error() string {
	return "-"
}

type Validator struct {
	validator *validator.Validate
}

func NewValidator() *Validator {
	v := validator.New()
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return &Validator{v}
}

func (v *Validator) Validate(value any) error {
	if err := v.validator.Struct(value); err != nil {
		return formatValidationError(err)
	}
	return nil
}

func formatValidationError(err error) ErrorResponse {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return ErrorResponse{Errors: []ValidationErrorResponse{
			{
				Field:   "error",
				Message: err.Error(),
			},
		},
		}
	}

	out := make([]ValidationErrorResponse, len(ve))
	for i, fe := range ve {
		out[i] = ValidationErrorResponse{
			Field:   fe.Field(),
			Message: fe.Error(),
		}
	}

	return ErrorResponse{Errors: out}
}
