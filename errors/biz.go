package errors

import "fmt"

// BizCode CSDS 6 位业务码（全局保留段，非领域码）。
type BizCode int

// 全局保留业务码（CSDS 10-api-error-code-spec §9）。
const (
	BizCodeParamInvalid   BizCode = 100001 // 参数错误
	BizCodeParamFormat    BizCode = 100002 // 参数格式/校验失败
	BizCodeUnauthorized   BizCode = 200001 // 未授权（通用）
	BizCodeAuthMissing    BizCode = 200002 // 缺少/无效凭证
	BizCodeForbidden      BizCode = 210001 // 无权限
	BizCodeInternal       BizCode = 500001 // 服务内部错误（未映射）
	BizCodeServiceUnavail BizCode = 503001 // 依赖未启用/不可用
)

// BizError 带 6 位业务码的错误。
type BizError struct {
	Code BizCode
	Msg  string
}

func (e *BizError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Msg)
}

func NewBiz(code BizCode, msg string) *BizError {
	return &BizError{Code: code, Msg: msg}
}
