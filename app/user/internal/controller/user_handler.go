package controller

import (
	"poptoy-flashsale/app/user/internal/service"
	"poptoy-flashsale/pkg/e"
	"poptoy-flashsale/pkg/response"

	"github.com/gin-gonic/gin"
)

// HandleRegister 用户注册接口
func HandleRegister(c *gin.Context) {
	var req service.RegisterReq

	// 参数绑定与基本校验
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, e.InvalidParams, err.Error())
		return
	}

	// 调用业务层
	code, err := service.Register(&req)
	if err != nil {
		response.Error(c, code)
		return
	}

	// 返回 201 Created
	response.Created(c, nil)
}

// HandleLogin 用户登录接口
func HandleLogin(c *gin.Context) {
	var req service.LoginReq

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, e.InvalidParams, err.Error())
		return
	}

	code, accesstoken, refreshToken,err := service.Login(&req)
	if err != nil {
		response.Error(c, code)
		return
	}

	// 返回 200 OK 及 Token
	response.Success(c, gin.H{
		"access_token": accesstoken,
		"refresh_token":refreshToken,
	})
}