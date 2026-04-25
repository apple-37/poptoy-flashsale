package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"poptoy-flashsale/app/order/internal/cache"
	"poptoy-flashsale/app/order/internal/model"
	"poptoy-flashsale/app/order/internal/mq"
	"poptoy-flashsale/app/order/internal/repository"
	pkgCache "poptoy-flashsale/pkg/cache"
	"poptoy-flashsale/pkg/fsm"
	"poptoy-flashsale/pkg/mysql"

	"gorm.io/gorm"
)

// HandleCreateOrderTask 消费正常下单任务
func HandleCreateOrderTask(task *mq.FlashTask) error {
	// 1. 初始化 FSM 状态机 (初始状态为 Init)
	orderFsm := fsm.NewOrderFSM(fsm.StateInit)

	// 注册 Init -> Pending 的数据库操作
	orderFsm.AddTransition(fsm.StateInit, fsm.EventCreate, fsm.StatePending, func(orderNo string) error {
		var existed model.Order
		err := mysql.DB.Where("order_no = ?", task.OrderNo).First(&existed).Error
		if err == nil {
			return nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		newOrder := &model.Order{
			OrderNo:   task.OrderNo,
			UserID:    task.UserID,
			ProductID: task.ProductID,
			Status:    int8(fsm.StatePending),
		}
		// 写入 MySQL (此处可结合扣减真实 MySQL 库存的事务)
		return mysql.DB.Create(newOrder).Error
	})

	// 2. 触发状态机执行落盘
	if err := orderFsm.Trigger(fsm.EventCreate, task.OrderNo); err != nil {
		return fmt.Errorf("订单创建失败: %w", err)
	}

	// 3. 落盘成功后，写入本地事务表并发送 15 分钟 TTL 延迟消息
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("序列化延迟任务失败: %w", err)
	}

	delayOutbox := &model.OutboxMessage{
		MessageType: model.OutboxTypeDelayCancelTask,
		BizKey:      "delay:" + task.OrderNo,
		Payload:     string(payload),
		Status:      model.OutboxStatusPending,
		NextRetryAt: time.Now(),
	}
	if err := repository.CreateOutboxMessageIfNotExists(delayOutbox); err != nil {
		return fmt.Errorf("写入延迟任务本地事务消息失败: %w", err)
	}

	if delayOutbox.ID > 0 {
		if err := publishDelayTaskWithRetry(task, 3); err != nil {
			_ = repository.ScheduleOutboxRetry(delayOutbox.ID, err.Error(), time.Now().Add(30*time.Second))
			log.Printf("[Worker] 延迟任务发送失败，已进入补偿重试: order_no=%s err=%v", task.OrderNo, err)
		} else if err := repository.MarkOutboxMessageSent(delayOutbox.ID); err != nil {
			log.Printf("[Worker] 标记延迟任务消息已发送失败: outbox_id=%d err=%v", delayOutbox.ID, err)
		}
	}

	// 4. Redis 发布 SSE 事件，通知前端
	channel := fmt.Sprintf("order_result_%d", task.UserID)
	eventPayload := fmt.Sprintf(`{"order_no": "%s", "status": "success"}`, task.OrderNo)
	pkgCache.Rdb.Publish(context.Background(), channel, eventPayload)

	log.Printf("[Worker] 订单 %s 创建成功，已通知用户 %d\n", task.OrderNo, task.UserID)
	return nil
}

// HandleDLXCancelTask 消费死信队列，处理超时取消
func HandleDLXCancelTask(task *mq.FlashTask) error {
	// 1. 查询数据库最新状态
	var order model.Order
	if err := mysql.DB.Where("order_no = ?", task.OrderNo).First(&order).Error; err != nil {
		return err // 订单不存在，忽略
	}

	// 2. 只有在 Pending 状态才允许取消 (初始化当前状态)
	orderFsm := fsm.NewOrderFSM(fsm.OrderState(order.Status))

	// 注册 Pending -> Cancelled 的动作
	orderFsm.AddTransition(fsm.StatePending, fsm.EventTimeout, fsm.StateCancelled, func(orderNo string) error {
		// A. 开启事务更新 MySQL 状态
		tx := mysql.DB.Begin()
		if err := tx.Model(&order).Update("status", int8(fsm.StateCancelled)).Error; err != nil {
			tx.Rollback()
			return err
		}

		// B. Redis 库存回滚 (恢复抢购资格)
		stockKey := fmt.Sprintf("%s%d", cache.FlashStockKeyPrefix, task.ProductID)
		userSetKey := fmt.Sprintf("%s%d", cache.FlashPurchasedKeyPrefix, task.ProductID)

		pipe := pkgCache.Rdb.Pipeline()
		pipe.Incr(context.Background(), stockKey)                // 库存 +1
		pipe.SRem(context.Background(), userSetKey, task.UserID) // 移除已购记录
		if _, err := pipe.Exec(context.Background()); err != nil {
			tx.Rollback()
			return err
		}

		return tx.Commit().Error
	})

	// 3. 触发超时事件 (如果订单已被支付，状态不是 Pending，FSM 会报错并阻止执行，完美防并发死锁！)
	err := orderFsm.Trigger(fsm.EventTimeout, task.OrderNo)
	if err != nil {
		log.Printf("[DLX Worker] 订单 %s 取消跳过: %v\n", task.OrderNo, err)
		return nil // 非法流转（如已支付），直接丢弃该死信消息，不算消费失败
	}

	log.Printf("[DLX Worker] 订单 %s 超时未支付，已成功取消并回滚库存！\n", task.OrderNo)
	return nil
}

func publishDelayTaskWithRetry(task *mq.FlashTask, attempts int) error {
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := mq.PublishDelayTask(ctx, task)
		cancel()
		if err == nil {
			return nil
		}

		lastErr = err
		time.Sleep(time.Duration(i+1) * 200 * time.Millisecond)
	}

	return lastErr
}
