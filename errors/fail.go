package errors

// Internal 未映射的服务端错误（bizCode 500001）。
func Internal(msg string) *BizError {
	return NewBiz(BizCodeInternal, msg)
}

// ParamInvalid 参数错误（bizCode 100001）。
func ParamInvalid(msg string) *BizError {
	return NewBiz(BizCodeParamInvalid, msg)
}

// ParamFormat 参数校验失败（bizCode 100002）。
func ParamFormat(msg string) *BizError {
	return NewBiz(BizCodeParamFormat, msg)
}

// AuthMissing 缺少或无效凭证（bizCode 200002）。
func AuthMissing(msg string) *BizError {
	return NewBiz(BizCodeAuthMissing, msg)
}

// PermissionDenied 无权限（bizCode 210001）。
func PermissionDenied(msg string) *BizError {
	return NewBiz(BizCodeForbidden, msg)
}

// ServiceUnavail 依赖未启用（bizCode 503001）。
func ServiceUnavail(msg string) *BizError {
	return NewBiz(BizCodeServiceUnavail, msg)
}
