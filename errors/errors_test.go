package errors

import (
	"errors"
	"testing"
)

func TestAppErrorMessage(t *testing.T) {
	tests := []struct {
		name   string
		err    *AppError
		expect string
	}{
		{
			name:   "simple error",
			err:    New(ErrNotFound, "resource not found"),
			expect: "[NOT_FOUND] resource not found",
		},
		{
			name:   "wrapped error",
			err:    Wrap(ErrDatabaseError, "query failed", errors.New("connection refused")),
			expect: "[DATABASE_ERROR] query failed: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expect {
				t.Errorf("Error() = %s, want %s", got, tt.expect)
			}
		})
	}
}

func TestAppErrorWithDetails(t *testing.T) {
	err := New(ErrNotFound, "not found").
		WithDetails(map[string]interface{}{"id": "123", "type": "document"})

	if err.Details["id"] != "123" {
		t.Error("details should contain id")
	}
	if err.Details["type"] != "document" {
		t.Error("details should contain type")
	}
}

func TestAppErrorWithDetail(t *testing.T) {
	err := New(ErrNotFound, "not found").
		WithDetail("id", "456")

	if err.Details["id"] != "456" {
		t.Error("details should contain id")
	}
}

func TestIsErrorCode(t *testing.T) {
	err := New(ErrNotFound, "not found")

	if !Is(err, ErrNotFound) {
		t.Error("Is should return true for matching error code")
	}
	if Is(err, ErrDatabaseError) {
		t.Error("Is should return false for different error code")
	}
}

func TestGetCode(t *testing.T) {
	err := New(ErrNotFound, "not found")
	if GetCode(err) != ErrNotFound {
		t.Error("GetCode should return correct error code")
	}

	stdErr := errors.New("standard error")
	if GetCode(stdErr) != ErrUnknown {
		t.Error("GetCode should return ErrUnknown for non-AppError")
	}
}

func TestGetErrorMessage(t *testing.T) {
	err := New(ErrNotFound, "resource not found")
	if GetMessage(err) != "resource not found" {
		t.Error("GetMessage should return error message")
	}
}

func TestUserMessage(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		expect string
	}{
		{
			name:   "app error",
			err:    New(ErrNotFound, "自定义消息"),
			expect: "自定义消息",
		},
		{
			name:   "network error",
			err:    errors.New("connection refused"),
			expect: "网络连接失败，请检查网络设置",
		},
		{
			name:   "timeout error",
			err:    errors.New("operation timeout"),
			expect: "请求超时，请稍后重试",
		},
		{
			name:   "permission error",
			err:    errors.New("permission denied"),
			expect: "权限不足，无法执行此操作",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserMessage(tt.err)
			if got != tt.expect {
				t.Errorf("UserMessage() = %s, want %s", got, tt.expect)
			}
		})
	}
}

func TestWrapIfNotNil(t *testing.T) {
	// Should wrap non-nil error
	err := errors.New("base error")
	wrapped := WrapIfNotNil(ErrDatabaseError, "wrapped", err)
	if wrapped == nil {
		t.Error("WrapIfNotNil should wrap non-nil error")
	}

	// Should return nil for nil error
	noError := WrapIfNotNil(ErrDatabaseError, "wrapped", nil)
	if noError != nil {
		t.Error("WrapIfNotNil should return nil for nil error")
	}
}

func TestUnwrap(t *testing.T) {
	baseErr := errors.New("base error")
	wrapped := Wrap(ErrDatabaseError, "wrapped", baseErr)

	unwrapped := errors.Unwrap(wrapped)
	if unwrapped != baseErr {
		t.Error("Unwrap should return the cause error")
	}
}
