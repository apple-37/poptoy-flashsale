package controller

import (
	"net/http"

	"poptoy-flashsale/app/order/internal/service"
	"poptoy-flashsale/pkg/e"
	"poptoy-flashsale/pkg/response"

	"github.com/gin-gonic/gin"
)

// HandleFlashBuy 发起秒杀请求
func HandleFlashBuy(c *gin.Context) {
	var req service.FlashBuyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMsg(c, e.InvalidParams, err.Error())
		return
	}

	// 从 JWT 中间件获取 userID (必须是 uint64 格式)
	userIDAny, exists := c.Get("userID")
	if !exists {
		response.Error(c, e.Unauthorized)
		return
	}
	userID := userIDAny.(uint64)

	// 调用核心业务
	orderNo, err := service.FlashBuy(c.Request.Context(), userID, &req)
	if err != nil {
		if err == service.ErrSoldOut || err == service.ErrDuplicate {
			response.ErrorWithMsg(c, 40010, err.Error()) // 业务限制拒绝
		} else {
			response.Error(c, e.Error) // 内部错误
		}
		return
	}

	// HTTP 202 Accepted：任务已接收，处理尚未完成
	c.JSON(http.StatusAccepted, response.Response{
		Code: 20200,
		Msg:  "抢购任务已受理，请监听推送或稍后查看订单",
		Data: gin.H{"order_no": orderNo},
	})
}