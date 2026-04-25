package apperrors

import (
	"fmt"
	"net/http"
)

// 错误码设计规范
// 格式：5位数 ABCDE
//   AB  → 大类（10-99）
//   CDE → 细分（000-999）
//
// 大类划分：
//   10xx  → 认证与授权
//   20xx  → 用户与个人资源
//   21xx  → 会话与消息（IM）
//   22xx  → 文件与存储
//   23xx  → 搜索服务
//   40xx  → 客户端参数错误
//   50xx  → 服务器内部错误
//   60xx  → 第三方服务错误

type AppError struct {
	Code       int            `json:"code"`
	Message    string         `json:"message"`
	httpStatus int            `json:"-"`          // 对应 HTTP 状态码
	Err        error          `json:"-"`
	Details    map[string]any `json:"details,omitempty"` // 附加上下文
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

// WithDetails 返回带附加上下文的错误（不影响原错误）
func (e *AppError) WithDetails(k string, v any) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[k] = v
	return e
}

// HTTPStatus 返回对应的 HTTP 状态码
func (e *AppError) HTTPStatus() int {
	return e.httpStatus
}

// New 创建新的业务错误
func New(code int, msg string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: msg, httpStatus: httpStatus}
}

// Wrap 将底层错误包装为业务错误（保留原始错误链）
func Wrap(code int, msg string, httpStatus int, err error) *AppError {
	return &AppError{Code: code, Message: msg, httpStatus: httpStatus, Err: err}
}

// Is 判断错误是否为指定 AppError（用于 errors.Is 比较）
func Is(err, target error) bool {
	if e, ok := err.(*AppError); ok {
		if t, ok := target.(*AppError); ok {
			return e.Code == t.Code
		}
	}
	return false
}

// ===== 10xx 认证与授权 =====

var (
	ErrUnauthorized      = New(10100, "未登录或 Token 已过期", http.StatusUnauthorized)
	ErrTokenInvalid      = New(10101, "Token 格式无效", http.StatusUnauthorized)
	ErrTokenExpired      = New(10102, "Token 已过期", http.StatusUnauthorized)
	ErrSignatureInvalid  = New(10103, "Token 签名验证失败", http.StatusUnauthorized)
	ErrPermissionDenied  = New(10300, "权限不足", http.StatusForbidden)
)

// ===== 20xx 用户与个人资源 =====

var (
	ErrUserNotFound      = New(20401, "用户不存在", http.StatusNotFound)
	ErrUserAlreadyExists = New(20409, "用户名已存在", http.StatusConflict)
	ErrInvalidPassword   = New(10401, "用户名或密码错误", http.StatusUnauthorized)
	ErrInvalidEmail      = New(10402, "邮箱格式无效", http.StatusBadRequest)
)

// ===== 21xx 会话与消息 =====

var (
	ErrConversationNotFound = New(21401, "会话不存在", http.StatusNotFound)
	ErrNotInConversation    = New(21403, "不在该会话中", http.StatusForbidden)
	ErrCannotChatWithSelf   = New(10403, "不能和自己聊天", http.StatusBadRequest)
	ErrPrivateChatImmutable = New(21404, "私聊无法操作成员", http.StatusBadRequest)
	ErrNotGroupOwner       = New(21405, "只有群主可以执行此操作", http.StatusForbidden)
	ErrMemberAlreadyInConv  = New(21409, "用户已在会话中", http.StatusConflict)

	ErrMessageNotFound    = New(22401, "消息不存在", http.StatusNotFound)
	ErrCannotEditOthersMsg = New(22403, "只能编辑自己发送的消息", http.StatusForbidden)
	ErrCannotRecallOthers  = New(22403, "只能撤回自己发送的消息", http.StatusForbidden)
)

// ===== 22xx 文件与存储 =====

var (
	ErrFileNotFound       = New(23401, "文件不存在", http.StatusNotFound)
	ErrFileTooLarge       = New(23402, "文件大小超出限制", http.StatusBadRequest)
	ErrUnsupportedFileType = New(23403, "不支持的文件类型", http.StatusBadRequest)
	ErrUploadFailed       = New(25401, "文件上传失败", http.StatusInternalServerError)
	ErrStorageUnavailable = New(26501, "存储服务不可用", http.StatusServiceUnavailable)
)

// ===== 23xx 搜索服务 =====

var (
	ErrSearchServiceUnavailable = New(26502, "搜索服务不可用", http.StatusServiceUnavailable)
	ErrSearchQueryEmpty          = New(14401, "搜索关键词不能为空", http.StatusBadRequest)
)

// ===== 40xx 客户端参数错误 =====

var (
	ErrNotFound           = New(14040, "资源不存在", http.StatusNotFound)
	ErrBadRequest         = New(14000, "请求参数错误", http.StatusBadRequest)
	ErrConflict           = New(14090, "资源冲突", http.StatusConflict)
	ErrMissingRequiredParam = New(14000, "缺少必需参数", http.StatusBadRequest)
	ErrInvalidConvID      = New(14001, "无效的会话 ID", http.StatusBadRequest)
	ErrInvalidMsgID       = New(14002, "无效的消息 ID", http.StatusBadRequest)
	ErrContentEmpty      = New(14003, "消息内容不能为空", http.StatusBadRequest)
)

// ===== 50xx 服务器内部错误 =====

var (
	ErrInternalServer = New(15000, "服务器内部错误", http.StatusInternalServerError)
	ErrDatabase       = New(15010, "数据库错误", http.StatusInternalServerError)
	ErrCache          = New(15020, "缓存服务错误", http.StatusInternalServerError)
	ErrInternal       = New(15030, "内部处理错误", http.StatusInternalServerError)
)

// ===== 60xx 第三方服务错误 =====

var (
	ErrMinIOUnavailable = New(26501, "对象存储服务不可用", http.StatusServiceUnavailable)
	ErrRedisUnavailable = New(26502, "缓存服务不可用", http.StatusServiceUnavailable)
)

// ===== 通用 Wrap 快捷函数 =====

// DatabaseError 包装数据库错误
func DatabaseError(err error) *AppError {
	return Wrap(15010, "数据库操作失败", http.StatusInternalServerError, err)
}

// CacheError 包装缓存错误
func CacheError(err error) *AppError {
	return Wrap(15020, "缓存操作失败", http.StatusInternalServerError, err)
}

// ValidateError 包装参数验证错误
func ValidateError(msg string) *AppError {
	return New(14000, msg, http.StatusBadRequest)
}
