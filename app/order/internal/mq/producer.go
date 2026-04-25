package mq

import (
	"context"
	"encoding/json"
	"log"

	pkgMq "poptoy-flashsale/pkg/mq"

	amqp "github.com/rabbitmq/amqp091-go"
)

type FlashTask struct {
	UserID    uint64 `json:"user_id"`
	ProductID uint64 `json:"product_id"`
	OrderNo   string `json:"order_no"`
}

// PublishFlashTask 发送秒杀下单任务到队列 (已有代码)
func PublishFlashTask(ctx context.Context, task *FlashTask) error {
	body, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return pkgMq.Channel.PublishWithContext(ctx, "", pkgMq.OrderTaskQueue, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent, ContentType: "application/json", Body: body,
	})
}

// PublishDelayTask 发送延迟任务到 delay.queue (新增代码)
// 消息会在这里停留 15 分钟，过期后自动掉入 dlx.queue
func PublishDelayTask(ctx context.Context, task *FlashTask) error {
	body, err := json.Marshal(task)
	if err != nil {
		return err
	}
	err = pkgMq.Channel.PublishWithContext(
		ctx,
		"",                    // 默认交换机
		pkgMq.OrderDelayQueue, // 直接发往延迟队列
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		log.Printf("[MQ] 发送延迟任务失败: %v\n", err)
	}
	return err
}
