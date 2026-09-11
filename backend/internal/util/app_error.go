package util

import "fmt"

// AppError 统一业务错误，code 对应 constants 错误码。
type AppError struct {
	Code    int
	Message string
	Err     error
}

// Error 实现 error 接口。
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap 支持 errors.Is/As 错误链。
func (e *AppError) Unwrap() error { return e.Err }

// NewAppError 构造业务错误。
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}
