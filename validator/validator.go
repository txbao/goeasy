package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	defaultValidate *validator.Validate
	zhTrans         ut.Translator
)

func init() {
	defaultValidate = validator.New()
	zh := zh.New()
	uni := ut.New(zh, zh)
	var ok bool
	zhTrans, ok = uni.GetTranslator("zh")
	if !ok {
		return
	}
	_ = zhTranslations.RegisterDefaultTranslations(defaultValidate, zhTrans)
}

// Validate 校验结构体。
func Validate(v any) error {
	return defaultValidate.Struct(v)
}

// Engine 返回底层 validator 实例。
func Engine() *validator.Validate {
	return defaultValidate
}

// Translator 返回中文翻译器。
func Translator() ut.Translator {
	return zhTrans
}

// Format 将校验错误格式化为中文说明（供 HTTP 400 返回）。
func Format(err error) string {
	if err == nil {
		return ""
	}
	if ves, ok := err.(validator.ValidationErrors); ok && zhTrans != nil {
		var lines []string
		for _, fe := range ves {
			lines = append(lines, fe.Translate(zhTrans))
		}
		return strings.Join(lines, "; ")
	}
	return err.Error()
}

// ValidationErrors 返回逐条中文错误。
func ValidationErrors(err error) []string {
	if err == nil {
		return nil
	}
	if ves, ok := err.(validator.ValidationErrors); ok && zhTrans != nil {
		out := make([]string, 0, len(ves))
		for _, fe := range ves {
			out = append(out, fe.Translate(zhTrans))
		}
		return out
	}
	return []string{err.Error()}
}

// MustValidate 校验失败时 panic（测试/启动期用）。
func MustValidate(v any) {
	if err := Validate(v); err != nil {
		panic(fmt.Sprintf("validator: %s", Format(err)))
	}
}
