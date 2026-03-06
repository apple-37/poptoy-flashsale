package controller

import (
	"context"
	"fmt"
	"io"
	"log"

	pkgCache "poptoy-flashsale/pkg/cache"

	"github.com/gin-gonic/gin"
)

// HandleSSEResult 建立 Server-Sent Events 连接，实时推送秒杀结果
func HandleSSEResult(c *gin.Context) {
	// 从 JWT 获取用户 ID
	userIDAny, _ := c.Get("userID")
	userID := userIDAny.(uint64)

	// 1. 设置 SSE 必备的 HTTP 头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// 2. 订阅 Redis 中针对该用户的专属频道
	channelName := fmt.Sprintf("order_result_%d", userID)
	pubsub := pkgCache.Rdb.Subscribe(context.Background(), channelName)
	defer pubsub.Close()

	ch := pubsub.Channel()
	
	// 向客户端发送一条握手消息，证明连接成功
	c.Stream(func(w io.Writer) bool {
		c.SSEvent("ping", "connected to SSE server, waiting for result...")
		return false // 仅发一次
	})

	// 3. 阻塞监听：直到收到 Redis 消息，或者客户端断开连接
	clientGone := c.Writer.CloseNotify()

	for {
		select {
		case <-clientGone:
			// 客户端（浏览器）断开了长连接
			log.Printf("[SSE] 用户 %d 断开连接", userID)
			return
		case msg := <-ch:
			// 收到 Redis Pub 发布的订单落盘成功事件
			c.Stream(func(w io.Writer) bool {
				// 发送自定义事件 "order_success"，数据为 JSON 字符串
				c.SSEvent("order_success", msg.Payload)
				return false
			})
			return // 推送完毕，主动断开当前连接
		}
	}
}