package repository

import (
	"poptoy-flashsale/app/product/internal/model"
	"poptoy-flashsale/pkg/mysql"
)

func GetProductList(cursor uint64, limit int) ([]*model.ProductHot, error) {
	var products []*model.ProductHot

	query := mysql.DB.Where("status = ?", 1)

	// 如果游标大于0，说明不是第一页，利用主键索引直接定位
	if cursor > 0 {
		// 假设按 ID 降序排列 (最新上架在前)
		query = query.Where("id < ?", cursor)
	}

	err := query.Order("id DESC").Limit(limit).Find(&products).Error
	return products, err
}

// GetProductDetail 获取商品完整详情 (连表或多次查询组装)
func GetProductDetail(id uint64) (*model.ProductDetail, error) {
	var hot model.ProductHot
	var cold model.ProductCold

	// 1. 查热表 (仅上架商品)
	if err := mysql.DB.Where("id = ? AND status = ?", id, 1).First(&hot).Error; err != nil {
		return nil, err
	}

	// 2. 查冷表 (可能存在无长文本的情况，允许 ErrRecordNotFound)
	mysql.DB.Where("product_id = ?", id).First(&cold)

	return &model.ProductDetail{
		ProductHot:  &hot,
		Description: cold.Description,
		ImagesJson:  cold.ImagesJson,
	}, nil
}

// GetProductByID 根据 ID 查询商品
func GetProductByID(id uint64) (*model.ProductHot, error) {
	var product model.ProductHot
	err := mysql.DB.Where("id = ?", id).First(&product).Error
	return &product, err
}

// UpdateProductStatus 更新商品状态
func UpdateProductStatus(productID uint64, status int8) error {
	return mysql.DB.Model(&model.ProductHot{}).Where("id = ?", productID).Update("status", status).Error
}
