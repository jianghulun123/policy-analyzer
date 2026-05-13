// Package errors 提供统一的错误处理机制
//
// 定义应用级错误类型和错误处理工具
package errors

import (
	"fmt"
	"strings"

	"policy-analyzer/logger"
)

// ErrorCode 错误码
type ErrorCode string

const (
	// 通用错误
	ErrUnknown       ErrorCode = "UNKNOWN"
	ErrInvalidInput  ErrorCode = "INVALID_INPUT"
	ErrNotFound      ErrorCode = "NOT_FOUND"
	ErrAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrPermission    ErrorCode = "PERMISSION_DENIED"
	ErrTimeout       ErrorCode = "TIMEOUT"
	ErrCanceled      ErrorCode = "CANCELED"

	// 文档相关错误
	ErrDocumentNotFound   ErrorCode = "DOCUMENT_NOT_FOUND"
	ErrDocumentParseError ErrorCode = "DOCUMENT_PARSE_ERROR"
	ErrDocumentImport     ErrorCode = "DOCUMENT_IMPORT_ERROR"
	ErrDocumentRead       ErrorCode = "DOCUMENT_READ_ERROR"

	// AI 相关错误
	ErrAIProviderNotConfigured ErrorCode = "AI_PROVIDER_NOT_CONFIGURED"
	ErrAIModellNotAvailable    ErrorCode = "AI_MODEL_NOT_AVAILABLE"
	ErrAIRequestFailed         ErrorCode = "AI_REQUEST_FAILED"
	ErrAIRateLimitExceeded     ErrorCode = "AI_RATE_LIMIT_EXCEEDED"
	ErrAITokenExhausted        ErrorCode = "AI_TOKEN_EXHAUSTED"

	// 工作流相关错误
	ErrWorkflowNotFound   ErrorCode = "WORKFLOW_NOT_FOUND"
	ErrWorkflowParseError ErrorCode = "WORKFLOW_PARSE_ERROR"
	ErrWorkflowExecution  ErrorCode = "WORKFLOW_EXECUTION_ERROR"
	ErrStepFailed         ErrorCode = "STEP_FAILED"
	ErrSkillNotFound      ErrorCode = "SKILL_NOT_FOUND"

	// 配置相关错误
	ErrConfigNotFound ErrorCode = "CONFIG_NOT_FOUND"
	ErrConfigInvalid  ErrorCode = "CONFIG_INVALID"

	// 存储相关错误
	ErrStorageInit    ErrorCode = "STORAGE_INIT_ERROR"
	ErrStorageRead    ErrorCode = "STORAGE_READ_ERROR"
	ErrStorageWrite   ErrorCode = "STORAGE_WRITE_ERROR"
	ErrStorageDelete  ErrorCode = "STORAGE_DELETE_ERROR"
	ErrDatabaseError  ErrorCode = "DATABASE_ERROR"
)

// AppError 应用错误
type AppError struct {
	Code    ErrorCode
	Message string
	Cause   error
	Details map[string]interface{}
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现错误解包
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithDetails 添加错误详情
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// WithDetail 添加单个错误详情
func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// New 创建新的应用错误
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装已有错误
func Wrap(code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// WrapIfNotNil 如果错误不为 nil 则包装
func WrapIfNotNil(code ErrorCode, message string, cause error) *AppError {
	if cause == nil {
		return nil
	}
	return Wrap(code, message, cause)
}

// Is 判断错误类型
func Is(err error, code ErrorCode) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}

// GetCode 获取错误码
func GetCode(err error) ErrorCode {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return ErrUnknown
}

// GetMessage 获取错误消息（用户友好）
func GetMessage(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Message
	}
	return err.Error()
}

// UserMessage 获取用户友好的错误消息
func UserMessage(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Message
	}

	// 处理常见错误类型
	errStr := err.Error()
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "network is unreachable") {
		return "网络连接失败，请检查网络设置"
	}
	if strings.Contains(errStr, "timeout") {
		return "请求超时，请稍后重试"
	}
	if strings.Contains(errStr, "permission denied") {
		return "权限不足，无法执行此操作"
	}

	return "操作失败，请稍后重试"
}

// Log 记录错误日志
func Log(err error, context string) {
	if err == nil {
		return
	}

	if appErr, ok := err.(*AppError); ok {
		fields := []logger.Field{
			logger.F("code", appErr.Code),
		}
		if appErr.Cause != nil {
			fields = append(fields, logger.F("cause", appErr.Cause.Error()))
		}
		if len(appErr.Details) > 0 {
			fields = append(fields, logger.F("details", appErr.Details))
		}
		logger.ErrorErr(context, err, fields...)
	} else {
		logger.ErrorErr(context, err)
	}
}

// LogAndReturn 记录错误日志并返回
func LogAndReturn(err error, context string) error {
	Log(err, context)
	return err
}

// ========== 预定义错误 ==========

var (
	// 通用错误
	ErrNotFoundGeneric = New(ErrNotFound, "资源不存在")
	ErrInvalidInputMsg = New(ErrInvalidInput, "输入参数无效")
	ErrPermissionMsg   = New(ErrPermission, "权限不足")
	ErrTimeoutMsg      = New(ErrTimeout, "操作超时")
	ErrCanceledMsg     = New(ErrCanceled, "操作已取消")

	// 文档错误
	ErrDocumentNotFoundMsg = New(ErrDocumentNotFound, "文档不存在")
	ErrDocumentParseMsg    = New(ErrDocumentParseError, "文档解析失败")
	ErrDocumentImportMsg   = New(ErrDocumentImport, "文档导入失败")

	// AI 错误
	ErrAIProviderNotConfiguredMsg = New(ErrAIProviderNotConfigured, "AI 提供商未配置")
	ErrAIModelNotAvailableMsg     = New(ErrAIModellNotAvailable, "AI 模型不可用")
	ErrAIRequestFailedMsg         = New(ErrAIRequestFailed, "AI 请求失败")
	ErrAIRateLimitExceededMsg     = New(ErrAIRateLimitExceeded, "API 调用频率超限")
	ErrAITokenExhaustedMsg        = New(ErrAITokenExhausted, "API 余额不足")

	// 工作流错误
	ErrWorkflowNotFoundMsg   = New(ErrWorkflowNotFound, "工作流不存在")
	ErrWorkflowParseMsg      = New(ErrWorkflowParseError, "工作流解析失败")
	ErrWorkflowExecutionMsg  = New(ErrWorkflowExecution, "工作流执行失败")
	ErrSkillNotFoundMsg      = New(ErrSkillNotFound, "技能不存在")
)

// ========== 辅助函数 ==========

// PanicIfErr 如果有错误则 panic
func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Must 不允许错误，如果有错误则 panic
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// IgnoreError 忽略错误（仅用于明确不需要处理的错误）
func IgnoreError(err error) {
	// 不做任何事情
}

// OrDefault 如果错误则返回默认值
func OrDefault[T any](v T, err error, defaultVal T) T {
	if err != nil {
		return defaultVal
	}
	return v
}
