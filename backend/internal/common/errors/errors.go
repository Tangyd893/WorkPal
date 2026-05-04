package apperrors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code       int            `json:"code"`
	Message    string         `json:"message"`
	httpStatus int            `json:"-"`
	Err        error          `json:"-"`
	Details    map[string]any `json:"details,omitempty"`
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

func (e *AppError) WithDetails(k string, v any) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[k] = v
	return e
}

func (e *AppError) HTTPStatus() int {
	return e.httpStatus
}

func New(code int, msg string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: msg, httpStatus: httpStatus}
}

func Wrap(code int, msg string, httpStatus int, err error) *AppError {
	return &AppError{Code: code, Message: msg, httpStatus: httpStatus, Err: err}
}

func Is(err, target error) bool {
	if e, ok := err.(*AppError); ok {
		if t, ok := target.(*AppError); ok {
			return e.Code == t.Code
		}
	}
	return false
}

var (
	ErrUnauthorized     = New(10100, "未登录或 Token 已过期", http.StatusUnauthorized)
	ErrTokenInvalid     = New(10101, "Token 格式无效", http.StatusUnauthorized)
	ErrTokenExpired     = New(10102, "Token 已过期", http.StatusUnauthorized)
	ErrSignatureInvalid = New(10103, "Token 签名校验失败", http.StatusUnauthorized)
	ErrPermissionDenied = New(10300, "权限不足", http.StatusForbidden)
)

var (
	ErrUserNotFound      = New(20401, "用户不存在", http.StatusNotFound)
	ErrUserAlreadyExists = New(20409, "用户名已存在", http.StatusConflict)
	ErrInvalidPassword   = New(10401, "用户名或密码错误", http.StatusUnauthorized)
	ErrInvalidEmail      = New(10402, "邮箱格式无效", http.StatusBadRequest)
)

var (
	ErrConversationNotFound = New(21401, "会话不存在", http.StatusNotFound)
	ErrNotInConversation    = New(21403, "不在该会话中", http.StatusForbidden)
	ErrCannotChatWithSelf   = New(10403, "不能和自己聊天", http.StatusBadRequest)
	ErrPrivateChatImmutable = New(21404, "私聊无法修改成员", http.StatusBadRequest)
	ErrNotGroupOwner        = New(21405, "只有群主可以执行此操作", http.StatusForbidden)
	ErrMemberAlreadyInConv  = New(21409, "用户已在会话中", http.StatusConflict)

	ErrMessageNotFound     = New(22401, "消息不存在", http.StatusNotFound)
	ErrCannotEditOthersMsg = New(22403, "只能编辑自己发送的消息", http.StatusForbidden)
	ErrCannotRecallOthers  = New(22403, "只能撤回自己发送的消息", http.StatusForbidden)
)

var (
	ErrFileNotFound        = New(23401, "文件不存在", http.StatusNotFound)
	ErrFileTooLarge        = New(23402, "文件大小超出限制", http.StatusBadRequest)
	ErrUnsupportedFileType = New(23403, "不支持的文件类型", http.StatusBadRequest)
	ErrUploadFailed        = New(25401, "文件上传失败", http.StatusInternalServerError)
	ErrStorageUnavailable  = New(26501, "存储服务不可用", http.StatusServiceUnavailable)
)

var (
	ErrSearchServiceUnavailable = New(26502, "搜索服务不可用", http.StatusServiceUnavailable)
	ErrSearchQueryEmpty         = New(14401, "搜索关键词不能为空", http.StatusBadRequest)
)

var (
	ErrNotFound             = New(14040, "资源不存在", http.StatusNotFound)
	ErrBadRequest           = New(14000, "请求参数错误", http.StatusBadRequest)
	ErrConflict             = New(14090, "资源冲突", http.StatusConflict)
	ErrMissingRequiredParam = New(14000, "缺少必填参数", http.StatusBadRequest)
	ErrInvalidConvID        = New(14001, "无效的会话 ID", http.StatusBadRequest)
	ErrInvalidMsgID         = New(14002, "无效的消息 ID", http.StatusBadRequest)
	ErrContentEmpty         = New(14003, "消息内容不能为空", http.StatusBadRequest)
)

var (
	ErrInternalServer = New(15000, "服务器内部错误", http.StatusInternalServerError)
	ErrDatabase       = New(15010, "数据库错误", http.StatusInternalServerError)
	ErrCache          = New(15020, "缓存服务错误", http.StatusInternalServerError)
	ErrInternal       = New(15030, "内部处理错误", http.StatusInternalServerError)
)

var (
	ErrMinIOUnavailable = New(26501, "对象存储服务不可用", http.StatusServiceUnavailable)
	ErrRedisUnavailable = New(26502, "缓存服务不可用", http.StatusServiceUnavailable)
)

func DatabaseError(err error) *AppError {
	return Wrap(15010, "数据库操作失败", http.StatusInternalServerError, err)
}

func CacheError(err error) *AppError {
	return Wrap(15020, "缓存操作失败", http.StatusInternalServerError, err)
}

func ValidateError(msg string) *AppError {
	return New(14000, msg, http.StatusBadRequest)
}
