package errors

import "fmt"

// Code 业务无关的通用错误码。
type Code int

const (
	CodeOK           Code = 0
	CodeInvalid      Code = 1001
	CodeNotFound     Code = 1002
	CodeForbidden    Code = 1003
	CodeUnauthorized Code = 1004
	CodeInternal     Code = 5000
)

// CodedError 带码错误。
type CodedError struct {
	Code Code
	Msg  string
}

func (e *CodedError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Msg)
}

func New(code Code, msg string) *CodedError {
	return &CodedError{Code: code, Msg: msg}
}

func NotFound(msg string) *CodedError  { return New(CodeNotFound, msg) }
func Invalid(msg string) *CodedError   { return New(CodeInvalid, msg) }
func Forbidden(msg string) *CodedError { return New(CodeForbidden, msg) }
func Unauthorized(msg string) *CodedError {
	return New(CodeUnauthorized, msg)
}
