package service

import (
	"poptoy-flashsale/app/product/internal/cache"
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
		err := repository.UpdateProductStatus(id, int8(fsm.ProductStateOnSale))
		if err == nil {
			// 状态变更，使缓存失效
			_ = cache.InvalidateProductDetail(id)
			_ = cache.InvalidateProductList()
		}
		return err
	})
	
	productFSM.AddTransition(fsm.ProductStateOnSale, fsm.ProductEventTakeOff, fsm.ProductStateOffSale, func(id uint64) error {
		err := repository.UpdateProductStatus(id, int8(fsm.ProductStateOffSale))
		if err == nil {
			// 状态变更，使缓存失效
			_ = cache.InvalidateProductDetail(id)
			_ = cache.InvalidateProductList()
		}
		return err
	})
	
	productFSM.AddTransition(fsm.ProductStateOnSale, fsm.ProductEventSellOut, fsm.ProductStateSoldOut, func(id uint64) error {
		err := repository.UpdateProductStatus(id, int8(fsm.ProductStateSoldOut))
		if err == nil {
			// 状态变更，使缓存失效
			_ = cache.InvalidateProductDetail(id)
			_ = cache.InvalidateProductList()
		}
		return err
	})
	
	productFSM.AddTransition(fsm.ProductStateOffSale, fsm.ProductEventPutOn, fsm.ProductStateOnSale, func(id uint64) error {
		err := repository.UpdateProductStatus(id, int8(fsm.ProductStateOnSale))
		if err == nil {
			// 状态变更，使缓存失效
			_ = cache.InvalidateProductDetail(id)
			_ = cache.InvalidateProductList()
		}
		return err
	})
	
	productFSM.AddTransition(fsm.ProductStateSoldOut, fsm.ProductEventRestock, fsm.ProductStateOnSale, func(id uint64) error {
		err := repository.UpdateProductStatus(id, int8(fsm.ProductStateOnSale))
		if err == nil {
			// 状态变更，使缓存失效
			_ = cache.InvalidateProductDetail(id)
			_ = cache.InvalidateProductList()
		}
		return err
	})
	
	// 触发事件
	return productFSM.Trigger(event, productID)
}

func GetProductList(cursor uint64, size int) ([]*model.ProductHot, error) {
	if size < 1 || size > 100 {
		size = 10
	}

	// 1. 尝试从缓存获取
	list, err := cache.GetProductList(cursor, size)
	if err == nil && list != nil {
		return list, nil
	}

	// 2. 缓存未命中，从数据库查询
	list, err = repository.GetProductList(cursor, size)
	if err != nil {
		return nil, err
	}

	// 3. 存入缓存
	_ = cache.SetProductList(cursor, size, list)

	// 4. 批量添加到布隆过滤器
	_ = cache.BatchAddProductsToBloomFilter(list)

	return list, nil
}
// GetProductDetail 获取商品完整详情
func GetProductDetail(id uint64) (*model.ProductDetail, error) {
	// 1. 尝试从缓存获取 (包含布隆过滤器检查)
	detail, err := cache.GetProductDetail(id)
	if err == nil && detail != nil {
		return detail, nil
	}

	// 2. 缓存未命中，从数据库查询
	detail, err = repository.GetProductDetail(id)
	if err != nil {
		return nil, err
	}

	// 3. 存入缓存
	_ = cache.SetProductDetail(id, detail)

	// 4. 添加到布隆过滤器
	_ = cache.AddProductToBloomFilter(id)

	return detail, nil
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