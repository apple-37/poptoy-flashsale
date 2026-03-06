package order

import (
	"poptoy-flashsale/app/gateway/middleware"
	"poptoy-flashsale/app/order/internal/controller"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	orderGroup := r.Group("/orders").Use(middleware.JWTAuth())
	{
		orderGroup.POST("/flash-buy", controller.HandleFlashBuy)
		orderGroup.GET("/result/stream", controller.HandleSSEResult) 
	}
}