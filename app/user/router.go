package user

import (
	"poptoy-flashsale/app/gateway/middleware"
	"poptoy-flashsale/app/user/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 对外暴露的路由注册函数
// 因为 router.go 和 internal 在同一个父目录 app/user 下，所以允许导入
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
