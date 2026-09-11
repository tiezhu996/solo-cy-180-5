// Package handler 负责 HTTP 请求解析、校验与响应。
package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/util"
)

// bindJSON 解析并校验请求体，失败时写入统一错误响应。
func bindJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, fmt.Sprintf("%s: %v", constants.MsgValidationFailed, err))
		return false
	}
	return true
}

// bindQuery 解析查询参数。
func bindQuery(c *gin.Context, obj any) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidation, fmt.Sprintf("%s: %v", constants.MsgValidationFailed, err))
		return false
	}
	return true
}

// parseID 解析路径中的 uint 主键。
func parseID(c *gin.Context, name string) (uint, bool) {
	raw := c.Param(name)
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, fmt.Sprintf("路径参数 %s=%s 非法", name, raw))
		return 0, false
	}
	return uint(id), true
}
