package mq

import (
	"context"
	"encoding/json"
	"log"

	pkgMq "poptoy-flashsale/pkg/mq"
)

// StartWorkers 启动后台消费者协程
// 提示：为了解决循环依赖，我们将在外面传入具体的执行函数
func StartWorkers(ctx context.Context, taskHandler func(*FlashTask) error, dlxHandler func(*FlashTask) error) {
	go consumeTaskQueue(ctx, taskHandler)
	go consumeDLXQueue(ctx, dlxHandler)
}

func consumeTaskQueue(ctx context.Context, handler func(*FlashTask) error) {
	msgs, err := pkgMq.Channel.Consume(pkgMq.OrderTaskQueue, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("[MQ] 监听正常任务队列失败: %v", err)
	}

	log.Println("[Worker] 正在监听秒杀任务队列...")
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-msgs:
			var task FlashTask
			json.Unmarshal(msg.Body, &task)
			if err := handler(&task); err == nil {
				msg.Ack(false) // 消费成功，手动 Ack确认
			} else {
				msg.Nack(false, true) // 消费失败，重回队列
			}
		}
	}
}

func consumeDLXQueue(ctx context.Context, handler func(*FlashTask) error) {
	msgs, err := pkgMq.Channel.Consume(pkgMq.OrderDlxQueue, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("[MQ] 监听死信队列失败: %v", err)
	}

	log.Println("[Worker] 正在监听死信队列 (处理 15 分钟超时订单)...")
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-msgs:
			var task FlashTask
			json.Unmarshal(msg.Body, &task)
			// DLX 的任务无论成功还是被 FSM 拦截拒绝，都视作处理完毕，直接 Ack
			handler(&task) 
			msg.Ack(false)
		}
	}
}