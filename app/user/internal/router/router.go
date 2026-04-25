package router

import (
	"poptoy-flashsale/pkg/middleware"
	"poptoy-flashsale/app/user/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册用户模块路由
func RegisterRoutes(r *gin.RouterGroup) {
	userGroup := r.Group("/users")
	{
		userGroup.POST("/register", controller.HandleRegister)
		userGroup.POST("/login", controller.HandleLogin)
		userGroup.POST("/refresh", controller.HandleRefresh)

		authGroup := userGroup.Group("").Use(middleware.JWTAuth())
		authGroup.POST("/logout", controller.HandleLogout)
	}
}