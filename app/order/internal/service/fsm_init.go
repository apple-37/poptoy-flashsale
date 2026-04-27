package service

import (
	"context"
	"errors"
	"fmt"

	"poptoy-flashsale/app/order/internal/cache"
	"poptoy-flashsale/app/order/internal/model"
	pkgCache "poptoy-flashsale/pkg/cache"
	"poptoy-flashsale/pkg/fsm"
	"poptoy-flashsale/pkg/mysql"

	"gorm.io/gorm"
)

// InitFSM 在服务启动阶段注册订单状态机动作。
func InitFSM() {
	fsm.InitOrderFSMActions(createOrderAction, timeoutCancelOrderAction)
}

func createOrderAction(ctx fsm.Context) error {
	orderNo, _ := ctx["orderNo"].(string)
	userID, _ := ctx["userID"].(uint64)
	productID, _ := ctx["productID"].(uint64)
	if orderNo == "" || userID == 0 || productID == 0 {
		return fmt.Errorf("创建订单上下文不完整")
	}

	var existed model.Order
	err := mysql.DB.Where("order_no = ?", orderNo).First(&existed).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	newOrder := &model.Order{
		OrderNo:   orderNo,
		UserID:    userID,
		ProductID: productID,
		Status:    int8(fsm.StatePending),
	}
	return mysql.DB.Create(newOrder).Error
}

func timeoutCancelOrderAction(ctx fsm.Context) error {
	orderNo, _ := ctx["orderNo"].(string)
	if orderNo == "" {
		return fmt.Errorf("取消订单上下文不完整")
	}

	tx := mysql.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var order model.Order
	if err := tx.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		tx.Rollback()
		return err
	}

	res := tx.Model(&model.Order{}).
		Where("order_no = ? AND status = ?", orderNo, int8(fsm.StatePending)).
		Update("status", int8(fsm.StateCancelled))
	if res.Error != nil {
		tx.Rollback()
		return res.Error
	}
	if res.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("订单状态已变更，取消跳过")
	}

	stockKey := fmt.Sprintf("%s%d", cache.FlashStockKeyPrefix, order.ProductID)
	userSetKey := fmt.Sprintf("%s%d", cache.FlashPurchasedKeyPrefix, order.ProductID)

	pipe := pkgCache.Rdb.Pipeline()
	pipe.Incr(context.Background(), stockKey)
	pipe.SRem(context.Background(), userSetKey, order.UserID)
	if _, err := pipe.Exec(context.Background()); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
