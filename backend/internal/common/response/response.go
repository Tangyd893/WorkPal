package response

import (
	"net/http"

	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageData struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func SuccessPage(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	Success(c, PageData{
		Items:    data,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func Fail(c *gin.Context, err *apperrors.AppError) {
	c.JSON(err.HTTPStatus(), Response{
		Code:    err.Code,
		Message: err.Message,
		Data:    nil,
	})
}

func FailWithMessage(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: msg,
		Data:    nil,
	})
}