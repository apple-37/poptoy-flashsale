package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"poptoy-flashsale/app/order/internal/cache"
	"poptoy-flashsale/app/order/internal/mq"
)

// FlashBuyReq 秒杀请求参数
type FlashBuyReq struct {
	ProductID uint64 `json:"product_id" binding:"required"`
}

var (
	ErrSoldOut   = errors.New("商品已售罄")
	ErrDuplicate = errors.New("您已参与过该活动，请勿重复抢购")
	ErrSystemBusy = errors.New("系统繁忙，请稍后重试")
)

// FlashBuy 处理高并发秒杀请求 (核心链路)
func FlashBuy(ctx context.Context, userID uint64, req *FlashBuyReq) (string, error) {
	// 1. Redis Lua 脚本预扣库存并拦截重复购买
	code, err := cache.ExecFlashBuy(ctx, req.ProductID, userID)
	if err != nil {
		return "", ErrSystemBusy
	}

	if code == 0 {
		return "", ErrSoldOut
	}
	if code == -1 {
		return "", ErrDuplicate
	}

	// 2. 生成全局唯一订单号 (此处为了不引入新依赖，使用时间戳+用户ID简易生成，生产环境推荐 Snowflake)
	orderNo := fmt.Sprintf("ORD%d%06d", time.Now().UnixNano()/1e6, userID%1000000)

	// 3. 构建异步任务发往 RabbitMQ
	task := &mq.FlashTask{
		UserID:    userID,
		ProductID: req.ProductID,
		OrderNo:   orderNo,
	}

	if err := mq.PublishFlashTask(ctx, task); err != nil {
		// 极端情况：Redis 扣减了，但 MQ 发送失败。
		// 真实的微服务中这里会有一张本地消息表(Local Message Table)做补偿。
		// 这里简单处理为返回系统繁忙。
		return "", ErrSystemBusy
	}

	// 4. 返回受理成功和生成的订单号
	return orderNo, nil
}