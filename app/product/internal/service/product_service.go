package service

import (
	"poptoy-flashsale/app/product/internal/model"
	"poptoy-flashsale/app/product/internal/repository"
)

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