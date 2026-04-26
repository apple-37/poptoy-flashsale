package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"poptoy-flashsale/app/product/internal/model"
	"poptoy-flashsale/pkg/cache"
)

// 商品缓存键定义
const (
	ProductDetailKeyPrefix = "product:detail:"
	ProductListKeyPrefix  = "product:list:"
)

// 缓存过期时间配置
const (
	ProductDetailExpiration = 300 // 5分钟
	ProductListExpiration   = 60  // 1分钟
	RandomExpirationRange   = 60  // 随机范围 60秒
)

// GetProductDetail 从缓存获取商品详情
// 防缓存穿透: 使用布隆过滤器
// 防缓存雪崩: 使用随机过期时间
func GetProductDetail(productID uint64) (*model.ProductDetail, error) {
	// 1. 布隆过滤器检查 - 防缓存穿透
	exists, err := cache.ProductBloomFilter.ExistsString(strconv.FormatUint(productID, 10))
	if err != nil {
		// 布隆过滤器错误，继续执行但记录日志
		// 不应该因为布隆过滤器错误而影响正常业务
		// 这里可以添加日志记录
		log.Printf("Error checking product bloom filter: %v", err)
	} else if !exists {
		// 布隆过滤器判断商品一定不存在，直接返回
		return nil, fmt.Errorf("product not found")
	}

	// 2. 尝试从缓存获取
	key := fmt.Sprintf("%s%d", ProductDetailKeyPrefix, productID)
	val, err := cache.Rdb.Get(context.Background(), key).Result()
	if err == nil {
		// 缓存命中
		var detail model.ProductDetail
		if err := json.Unmarshal([]byte(val), &detail); err == nil {
			return &detail, nil
		}
	}

	// 缓存未命中，返回 nil，让业务层去数据库查询
	return nil, nil
}

// SetProductDetail 将商品详情存入缓存
func SetProductDetail(productID uint64, detail *model.ProductDetail) error {
	// 1. 序列化商品详情
	data, err := json.Marshal(detail)
	if err != nil {
		return err
	}

	// 2. 存入缓存，使用随机过期时间 - 防缓存雪崩
	key := fmt.Sprintf("%s%d", ProductDetailKeyPrefix, productID)
	expiration := cache.RandomExpiration(ProductDetailExpiration, RandomExpirationRange)
	return cache.Rdb.Set(context.Background(), key, data, expiration).Err()
}

// GetProductList 从缓存获取商品列表
func GetProductList(cursor uint64, size int) ([]*model.ProductHot, error) {
	// 1. 尝试从缓存获取
	key := fmt.Sprintf("%s%d:%d", ProductListKeyPrefix, cursor, size)
	val, err := cache.Rdb.Get(context.Background(), key).Result()
	if err == nil {
		// 缓存命中
		var list []*model.ProductHot
		if err := json.Unmarshal([]byte(val), &list); err == nil {
			return list, nil
		}
	}

	// 缓存未命中，返回 nil，让业务层去数据库查询
	return nil, nil
}

// SetProductList 将商品列表存入缓存
func SetProductList(cursor uint64, size int, list []*model.ProductHot) error {
	// 1. 序列化商品列表
	data, err := json.Marshal(list)
	if err != nil {
		return err
	}

	// 2. 存入缓存，使用随机过期时间 - 防缓存雪崩
	key := fmt.Sprintf("%s%d:%d", ProductListKeyPrefix, cursor, size)
	expiration := cache.RandomExpiration(ProductListExpiration, RandomExpirationRange/2) // 列表缓存过期时间更短
	return cache.Rdb.Set(context.Background(), key, data, expiration).Err()
}

// InvalidateProductDetail 使商品详情缓存失效
func InvalidateProductDetail(productID uint64) error {
	key := fmt.Sprintf("%s%d", ProductDetailKeyPrefix, productID)
	return cache.Rdb.Del(context.Background(), key).Err()
}

// InvalidateProductList 使商品列表缓存失效
func InvalidateProductList() error {
	// 使用通配符删除所有列表缓存
	keys, err := cache.Rdb.Keys(context.Background(), ProductListKeyPrefix+"*").Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return cache.Rdb.Del(context.Background(), keys...).Err()
	}
	return nil
}

// AddProductToBloomFilter 将商品ID添加到布隆过滤器
func AddProductToBloomFilter(productID uint64) error {
	return cache.ProductBloomFilter.AddString(strconv.FormatUint(productID, 10))
}

// BatchAddProductsToBloomFilter 批量添加商品ID到布隆过滤器
func BatchAddProductsToBloomFilter(products []*model.ProductHot) error {
	for _, product := range products {
		if err := AddProductToBloomFilter(product.ID); err != nil {
			// 记录错误但继续处理
			// 这里可以添加日志记录
		}
	}
	return nil
}