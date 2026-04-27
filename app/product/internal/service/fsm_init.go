package service

import (
	"fmt"

	"poptoy-flashsale/app/product/internal/cache"
	"poptoy-flashsale/app/product/internal/repository"
	"poptoy-flashsale/pkg/fsm"
)

// InitFSM 在服务启动阶段注册商品状态机动作。
func InitFSM() {
	fsm.InitProductFSMActions(
		func(productID uint64) error {
			return updateStatusAndInvalidateIfCurrent(productID, fsm.ProductStatePending, fsm.ProductStateOnSale)
		},
		func(productID uint64) error {
			return updateStatusAndInvalidateIfCurrent(productID, fsm.ProductStateOnSale, fsm.ProductStateOffSale)
		},
		func(productID uint64) error {
			return updateStatusAndInvalidateIfCurrent(productID, fsm.ProductStateOnSale, fsm.ProductStateSoldOut)
		},
		func(productID uint64) error {
			return updateStatusAndInvalidateIfCurrent(productID, fsm.ProductStateOffSale, fsm.ProductStateOnSale)
		},
		func(productID uint64) error {
			return updateStatusAndInvalidateIfCurrent(productID, fsm.ProductStateSoldOut, fsm.ProductStateOnSale)
		},
	)
}

func updateStatusAndInvalidateIfCurrent(productID uint64, expectedState fsm.ProductState, state fsm.ProductState) error {
	updated, err := repository.UpdateProductStatusIfCurrent(productID, int8(expectedState), int8(state))
	if err != nil {
		return err
	}
	if !updated {
		return fmt.Errorf("商品状态已变更，流转跳过")
	}

	_ = cache.InvalidateProductDetail(productID)
	_ = cache.InvalidateProductList()
	return nil
}
