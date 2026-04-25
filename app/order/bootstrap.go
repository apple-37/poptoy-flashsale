package order

import (
	"log"

	"poptoy-flashsale/app/order/internal/model"
	"poptoy-flashsale/pkg/mysql"
)

// InitStorage 初始化订单模块存储结构。
func InitStorage() {
	if err := mysql.DB.AutoMigrate(&model.Order{}, &model.OutboxMessage{}); err != nil {
		log.Fatalf("订单模块数据结构初始化失败: %v", err)
	}
}
