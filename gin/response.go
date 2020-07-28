package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIHandler gin router回调函数 handler函数定义格式
type APIHandler func(ctx *gin.Context) *JSONResponse

// ResponseWrapper
func ResponseWrapper(handle APIHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, handle(ctx))
	}
}

// JSONResponse 返回结构
type JSONResponse struct {
	Code int         `json:"errcode"`
	Msg  string      `json:"errmsg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

// ErrorResponse 错误返回
func ErrorResponse(code int, msg string) *JSONResponse {
	return &JSONResponse{Code: code, Msg: msg}
}

// SuccessResponse 正确返回
func SuccessResponse(data interface{}) *JSONResponse {
	return &JSONResponse{Code: 0, Msg: "", Data: data}
}
