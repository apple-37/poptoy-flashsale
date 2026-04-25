package router

import (
	"poptoy-flashsale/pkg/middleware"
	"poptoy-flashsale/app/order/internal/controller"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册订单模块路由
func RegisterRoutes(r *gin.RouterGroup) {
	orderGroup := r.Group("/orders").Use(middleware.JWTAuth())
	{
		orderGroup.POST("/flash-buy", controller.HandleFlashBuy)
		orderGroup.GET("/result/stream", controller.HandleSSEResult) 
	}
}