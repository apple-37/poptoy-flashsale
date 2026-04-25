package router

import (
	"poptoy-flashsale/app/product/internal/controller"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册商品模块路由
func RegisterRoutes(r *gin.RouterGroup) {
	productGroup := r.Group("/products")
	{
		productGroup.GET("", controller.GetList)
		productGroup.GET("/:id", controller.GetDetail)
	}
}