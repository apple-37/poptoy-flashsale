package service

import (
	"poptoy-flashsale/app/product/internal/model"
	"poptoy-flashsale/app/product/internal/repository"
	"poptoy-flashsale/pkg/fsm"
)

// UpdateProductStatus 更新商品状态
func UpdateProductStatus(productID uint64, event fsm.ProductEvent) error {
	// 获取商品当前状态
	product, err := repository.GetProductByID(productID)
	if err != nil {
		return err
	}
	
	// 创建商品状态机
	productFSM := fsm.NewProductFSM(fsm.ProductState(product.Status))
	
	// 注册状态转换动作
	productFSM.AddTransition(fsm.ProductStatePending, fsm.ProductEventApprove, fsm.ProductStateOnSale, func(id uint64) error {
		return repository.UpdateProductStatus(id, int8(fsm.ProductStateOnSale))
	})
	
	productFSM.AddTransition(fsm.ProductStateOnSale, fsm.ProductEventTakeOff, fsm.ProductStateOffSale, func(id uint64) error {
		return repository.UpdateProductStatus(id, int8(fsm.ProductStateOffSale))
	})
	
	productFSM.AddTransition(fsm.ProductStateOnSale, fsm.ProductEventSellOut, fsm.ProductStateSoldOut, func(id uint64) error {
		return repository.UpdateProductStatus(id, int8(fsm.ProductStateSoldOut))
	})
	
	productFSM.AddTransition(fsm.ProductStateOffSale, fsm.ProductEventPutOn, fsm.ProductStateOnSale, func(id uint64) error {
		return repository.UpdateProductStatus(id, int8(fsm.ProductStateOnSale))
	})
	
	productFSM.AddTransition(fsm.ProductStateSoldOut, fsm.ProductEventRestock, fsm.ProductStateOnSale, func(id uint64) error {
		return repository.UpdateProductStatus(id, int8(fsm.ProductStateOnSale))
	})
	
	// 触发事件
	return productFSM.Trigger(event, productID)
}

func GetProductList(cursor uint64, size int) ([]*model.ProductHot, error) {
	if size < 1 || size > 100 {
		size = 10
	}
	return repository.GetProductList(cursor, size)
}
// GetProductDetail 获取商品完整详情
func GetProductDetail(id uint64) (*model.ProductDetail, error) {
	return repository.GetProductDetail(id)
}

// ApproveProduct 审核商品
func ApproveProduct(productID uint64) error {
	return UpdateProductStatus(productID, fsm.ProductEventApprove)
}

// PutOnProduct 上架商品
func PutOnProduct(productID uint64) error {
	return UpdateProductStatus(productID, fsm.ProductEventPutOn)
}

// TakeOffProduct 下架商品
func TakeOffProduct(productID uint64) error {
	return UpdateProductStatus(productID, fsm.ProductEventTakeOff)
}

// MarkProductSoldOut 标记商品售罄
func MarkProductSoldOut(productID uint64) error {
	return UpdateProductStatus(productID, fsm.ProductEventSellOut)
}

// RestockProduct 商品补货
func RestockProduct(productID uint64) error {
	return UpdateProductStatus(productID, fsm.ProductEventRestock)
}