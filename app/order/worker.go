package order

import (
	"context"

	"poptoy-flashsale/app/order/internal/mq"
	"poptoy-flashsale/app/order/internal/service"
)

// InitWorkers 对外暴露的启动后台消费者协程的方法
// 将底层的 MQ 监听与具体的 Service 消费逻辑串联并启动
func InitWorkers(ctx context.Context) {
	mq.StartWorkers(ctx, service.HandleCreateOrderTask, service.HandleDLXCancelTask)
	go service.StartOutboxCompensationWorker(ctx)
}
