package bootstrap

import (
	"poptoy-flashsale/pkg/mysql"
	"context"
	"log"
	"poptoy-flashsale/app/order/internal/model"
	"poptoy-flashsale/app/order/internal/mq"
	"poptoy-flashsale/app/order/internal/service"
)

// InitStorage 初始化订单模块存储结构
func InitStorage() {
	if err := mysql.DB.AutoMigrate(&model.Order{}, &model.OutboxMessage{}); err != nil {
		log.Fatalf("订单模块数据结构初始化失败: %v", err)
	}
}

// InitWorkers 启动订单模块的后台消费者 Worker
func InitWorkers(ctx context.Context) {
	mq.StartWorkers(ctx, service.HandleCreateOrderTask, service.HandleDLXCancelTask)
	go service.StartOutboxCompensationWorker(ctx)
}