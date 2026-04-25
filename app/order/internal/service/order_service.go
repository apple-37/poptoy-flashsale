package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"poptoy-flashsale/app/order/internal/cache"
	"poptoy-flashsale/app/order/internal/model"
	"poptoy-flashsale/app/order/internal/mq"
	"poptoy-flashsale/app/order/internal/repository"
	"poptoy-flashsale/pkg/idgen"
)

// FlashBuyReq 秒杀请求参数
type FlashBuyReq struct {
	ProductID uint64 `json:"product_id" binding:"required"`
}

var (
	ErrSoldOut    = errors.New("商品已售罄")
	ErrDuplicate  = errors.New("您已参与过该活动，请勿重复抢购")
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

	// 2. 生成全局唯一订单号 (Snowflake)
	orderNo, err := idgen.NewOrderNo()
	if err != nil {
		_ = cache.RollbackFlashBuy(ctx, req.ProductID, userID)
		return "", ErrSystemBusy
	}

	// 3. 构建异步任务发往 RabbitMQ
	task := &mq.FlashTask{
		UserID:    userID,
		ProductID: req.ProductID,
		OrderNo:   orderNo,
	}
	payload, err := json.Marshal(task)
	if err != nil {
		_ = cache.RollbackFlashBuy(ctx, req.ProductID, userID)
		return "", ErrSystemBusy
	}

	outboxMsg := &model.OutboxMessage{
		MessageType: model.OutboxTypeCreateOrderTask,
		BizKey:      "flash:" + orderNo,
		Payload:     string(payload),
		Status:      model.OutboxStatusPending,
		NextRetryAt: time.Now(),
	}
	if err := repository.CreateOutboxMessage(outboxMsg); err != nil {
		_ = cache.RollbackFlashBuy(ctx, req.ProductID, userID)
		return "", ErrSystemBusy
	}

	if err := mq.PublishFlashTask(ctx, task); err != nil {
		log.Printf("[Order Service] 即时发送下单消息失败，已写入本地事务表等待补偿: order_no=%s err=%v", orderNo, err)
		return orderNo, nil
	}

	if err := repository.MarkOutboxMessageSent(outboxMsg.ID); err != nil {
		log.Printf("[Order Service] 标记本地下单消息已发送失败: outbox_id=%d err=%v", outboxMsg.ID, err)
	}

	// 4. 返回受理成功和生成的订单号
	return orderNo, nil
}
