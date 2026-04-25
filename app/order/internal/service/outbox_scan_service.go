package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"poptoy-flashsale/app/order/internal/model"
	"poptoy-flashsale/app/order/internal/mq"
	"poptoy-flashsale/app/order/internal/repository"
)

const (
	outboxScanBatchSize = 100
	outboxMaxRetry      = 20
)

// StartOutboxCompensationWorker 启动本地事务消息补偿扫描协程。
func StartOutboxCompensationWorker(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			scanAndDispatchOutbox(ctx)
		}
	}
}

func scanAndDispatchOutbox(ctx context.Context) {
	msgs, err := repository.ListDuePendingOutbox(outboxScanBatchSize)
	if err != nil {
		log.Printf("[Outbox] 扫描待补偿消息失败: %v", err)
		return
	}

	for _, msg := range msgs {
		if msg.RetryCount >= outboxMaxRetry {
			_ = repository.MarkOutboxMessageFailed(msg.ID, "exceed max retry")
			log.Printf("[Outbox] 消息超过最大重试次数, 已标记失败: id=%d biz_key=%s", msg.ID, msg.BizKey)
			continue
		}

		if err := dispatchOutboxMessage(ctx, &msg); err != nil {
			nextRetry := time.Now().Add(calcRetryDelay(msg.RetryCount))
			_ = repository.ScheduleOutboxRetry(msg.ID, err.Error(), nextRetry)
			log.Printf("[Outbox] 消息补偿发送失败: id=%d biz_key=%s err=%v", msg.ID, msg.BizKey, err)
			continue
		}

		if err := repository.MarkOutboxMessageSent(msg.ID); err != nil {
			log.Printf("[Outbox] 标记消息已发送失败: id=%d err=%v", msg.ID, err)
		}
	}
}

func dispatchOutboxMessage(ctx context.Context, msg *model.OutboxMessage) error {
	var task mq.FlashTask
	if err := json.Unmarshal([]byte(msg.Payload), &task); err != nil {
		return fmt.Errorf("反序列化消息体失败: %w", err)
	}

	sendCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	switch msg.MessageType {
	case model.OutboxTypeCreateOrderTask:
		return mq.PublishFlashTask(sendCtx, &task)
	case model.OutboxTypeDelayCancelTask:
		return mq.PublishDelayTask(sendCtx, &task)
	default:
		return fmt.Errorf("未知消息类型: %s", msg.MessageType)
	}
}

func calcRetryDelay(retryCount int) time.Duration {
	if retryCount < 3 {
		return 5 * time.Second
	}
	if retryCount < 8 {
		return 30 * time.Second
	}
	return 2 * time.Minute
}
