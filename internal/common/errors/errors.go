package errors

import (
	"fmt"
	"net/http"
)

// 错误码设计：5位数
// 4xxxx: 参数/请求错误
// 5xxxx: 服务器内部错误

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code int, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

func Wrap(code int, msg string, err error) *AppError {
	return &AppError{Code: code, Message: msg, Err: err}
}

// 预定义错误
var (
	ErrBadRequest       = New(40000, "请求参数错误")
	ErrUnauthorized     = New(40100, "未登录或登录已过期")
	ErrForbidden        = New(40300, "无权限访问该资源")
	ErrNotFound         = New(40400, "资源不存在")
	ErrConflict         = New(40900, "资源冲突")
	ErrInternalServer   = New(50000, "服务器内部错误")
	ErrDatabase         = New(50010, "数据库错误")

	// 用户相关
	ErrUserNotFound      = New(40401, "用户不存在")
	ErrUserAlreadyExists = New(40901, "用户名已存在")
	ErrInvalidPassword   = New(40101, "用户名或密码错误")
)
