package controller

import (
	"strconv"

	"poptoy-flashsale/app/product/internal/service"
	"poptoy-flashsale/pkg/e"
	"poptoy-flashsale/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetList(c *gin.Context) {
	// 获取游标参数 cursor，默认为 0
	cursor, _ := strconv.ParseUint(c.DefaultQuery("cursor", "0"), 10, 64)
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	list, err := service.GetProductList(cursor, size)
	if err != nil {
		response.Error(c, e.Error)
		return
	}

	// 计算下一次请求的游标
	var nextCursor uint64 = 0
	if len(list) > 0 {
		nextCursor = list[len(list)-1].ID
	}

	// 返回数据和 next_cursor 供前端下次请求使用
	response.Success(c, gin.H{
		"list":        list,
		"next_cursor": nextCursor,
	})
}

// GetDetail 获取商品详情
func GetDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ErrorWithMsg(c, e.InvalidParams, "商品ID格式不正确")
		return
	}

	detail, err := service.GetProductDetail(id)
	if err != nil {
		// 如果查不到商品，返回 404
		response.Error(c, e.NotFound)
		return
	}

	response.Success(c, detail)
}