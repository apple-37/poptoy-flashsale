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

	code, accessToken, refreshToken, err := service.Login(&req)
	if err != nil {
		response.Error(c, code)
		return
	}

	// 返回 200 OK 及 Token
	response.Success(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// HandleRefresh 刷新 Token 接口。
func HandleRefresh(c *gin.Context) {
	var req service.RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, e.InvalidParams, err.Error())
		return
	}

	code, accessToken, refreshToken, err := service.Refresh(&req)
	if err != nil {
		response.Error(c, code)
		return
	}

	response.Success(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// HandleLogout 登出并使当前用户旧 Token 全部失效。
func HandleLogout(c *gin.Context) {
	userIDAny, exists := c.Get("userID")
	if !exists {
		response.Error(c, e.Unauthorized)
		return
	}

	userID, ok := userIDAny.(uint64)
	if !ok {
		response.Error(c, e.Unauthorized)
		return
	}

	code, err := service.Logout(userID)
	if err != nil {
		response.Error(c, code)
		return
	}

	response.Success(c, gin.H{"logged_out": true})
}
