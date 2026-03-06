package response

import (
	"net/http"

	"poptoy-flashsale/pkg/e"

	"github.com/gin-gonic/gin"
)

// Response 基础响应结构体
type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

// Success 成功返回 (HTTP 200)
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code: e.Success,
		Msg:  e.GetMsg(e.Success),
		Data: data,
	})
}

// Created 创建成功返回 (HTTP 201)
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Response{
		Code: e.Created,
		Msg:  e.GetMsg(e.Created),
		Data: data,
	})
}

// Error 失败返回 (根据错误码推断 HTTP 状态码)
func Error(c *gin.Context, code int) {
	httpStatus := http.StatusBadRequest
	switch code {
case e.Unauthorized, e.LoginFailed:
		httpStatus = http.StatusUnauthorized
	case e.Error:
		httpStatus = http.StatusInternalServerError
	case e.NotFound:
		httpStatus = http.StatusNotFound
	}

	c.JSON(httpStatus, Response{
		Code: code,
		Msg:  e.GetMsg(code),
		Data: nil,
	})
}

// ErrorWithMsg 带有自定义错误信息的返回
func ErrorWithMsg(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusBadRequest, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}