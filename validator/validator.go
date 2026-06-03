package validator

import (
	"github.com/go-playground/validator/v10"
)

var defaultValidate = validator.New()

// Validate 校验结构体。
func Validate(v any) error {
	return defaultValidate.Struct(v)
}

// Engine 返回底层 validator 实例。
func Engine() *validator.Validate {
	return defaultValidate
}
