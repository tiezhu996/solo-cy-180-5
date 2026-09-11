package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/constants"
)

// Response 统一响应结构。
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// OK 返回成功响应。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: constants.CodeOK, Message: constants.MsgOK, Data: data})
}

// OKMessage 返回带自定义文案的成功响应。
func OKMessage(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Response{Code: constants.CodeOK, Message: message, Data: data})
}

// Fail 返回统一错误响应。
func Fail(c *gin.Context, httpStatus, code int, message string) {
	c.JSON(httpStatus, Response{Code: code, Message: message})
}
