// Package errorx 业务错误码定义
package errorx

import (
	"fmt"
)

const (
	// ========== 通用错误码 1xxxx ==========
	// 请求参数错误 10001-10100
	ErrCodeInvalidParam = 10001

	// ========== 数据理解相关错误码 2xxxx ==========
	// 数据理解模块基础错误码
	ErrCodeDataUnderstanding = 20000

	// 数据查询错误 20001-20100
	ErrCodeFormViewNotFound       = 20001 // 库表视图不存在
	ErrCodeFormFieldNotFound      = 20002 // 字段信息不存在
	ErrCodeBusinessObjectNotFound = 20003 // 业务对象不存在

	// 状态校验错误 20101-20200
	ErrCodeInvalidUnderstandStatus = 20101 // 无效的理解状态

	// 数据操作错误 20201-20300
	ErrCodeQueryFailed    = 20201 // 查询失败
	ErrCodeInsertFailed   = 20202 // 插入失败
	ErrCodeUpdateFailed   = 20203 // 更新失败
	ErrCodeDeleteFailed   = 20204 // 删除失败
	ErrCodeOperationFailed = 20205 // 操作失败

	// AI 服务相关错误 20301-20400
	ErrCodeAIServiceError = 20301 // AI 服务调用失败
)

// CodeError 带错误码的错误
type CodeError struct {
	Code    int
	Message string
	Err     error
}

// Error 实现 error 接口
func (e *CodeError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *CodeError) Unwrap() error {
	return e.Err
}

// New 创建带错误码的错误
func New(code int, message string) *CodeError {
	return &CodeError{
		Code:    code,
		Message: message,
	}
}

// Newf 创建带错误码和格式化消息的错误
func Newf(code int, format string, args ...interface{}) *CodeError {
	return &CodeError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap 包装错误并添加错误码
func Wrap(code int, err error, message string) *CodeError {
	return &CodeError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Wrapf 包装错误并添加错误码和格式化消息
func Wrapf(code int, err error, format string, args ...interface{}) *CodeError {
	return &CodeError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}

// Predefined errors
var (
	// 通用错误
	ErrInvalidParam = New(ErrCodeInvalidParam, "请求参数错误")

	// 数据查询错误
	ErrFormViewNotFound       = New(ErrCodeFormViewNotFound, "库表视图不存在")
	ErrFormFieldNotFound       = New(ErrCodeFormFieldNotFound, "字段信息不存在")
	ErrBusinessObjectNotFound  = New(ErrCodeBusinessObjectNotFound, "业务对象不存在")

	// 状态校验错误
	ErrInvalidUnderstandStatus = New(ErrCodeInvalidUnderstandStatus, "无效的理解状态")

	// 数据操作错误
	ErrQueryFailed    = New(ErrCodeQueryFailed, "查询失败")
	ErrInsertFailed   = New(ErrCodeInsertFailed, "插入失败")
	ErrUpdateFailed   = New(ErrCodeUpdateFailed, "更新失败")
	ErrDeleteFailed   = New(ErrCodeDeleteFailed, "删除失败")
	ErrOperationFailed = New(ErrCodeOperationFailed, "操作失败")

	// AI 服务相关错误
	ErrAIServiceError = New(ErrCodeAIServiceError, "AI 服务调用失败")
)

// NewFormViewNotFound 创建库表视图不存在错误
func NewFormViewNotFound(format string, args ...interface{}) *CodeError {
	return &CodeError{
		Code:    ErrCodeFormViewNotFound,
		Message: fmt.Sprintf(format, args...),
	}
}

// NewInvalidUnderstandStatus 创建无效理解状态错误
func NewInvalidUnderstandStatus(currentStatus int8) *CodeError {
	return Newf(ErrCodeInvalidUnderstandStatus, "当前状态不允许操作，当前状态: %d", currentStatus)
}

// NewQueryFailed 创建查询失败错误
func NewQueryFailed(operation string, err error) *CodeError {
	return Wrapf(ErrCodeQueryFailed, err, "查询%s失败", operation)
}

// NewInsertFailed 创建插入失败错误
func NewInsertFailed(operation string, err error) *CodeError {
	return Wrapf(ErrCodeInsertFailed, err, "插入%s失败", operation)
}

// NewUpdateFailed 创建更新失败错误
func NewUpdateFailed(operation string, err error) *CodeError {
	return Wrapf(ErrCodeUpdateFailed, err, "更新%s失败", operation)
}

// NewDeleteFailed 创建删除失败错误
func NewDeleteFailed(operation string, err error) *CodeError {
	return Wrapf(ErrCodeDeleteFailed, err, "删除%s失败", operation)
}

// NewAIServiceError 创建 AI 服务错误
func NewAIServiceError(err error) *CodeError {
	return Wrap(ErrCodeAIServiceError, err, "AI 服务调用失败")
}
