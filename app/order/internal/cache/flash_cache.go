package cache

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"poptoy-flashsale/pkg/cache"

	"github.com/redis/go-redis/v9"
)

// 预热数据的 Key 规则
// 库存 Key: flash:stock:{product_id} (String)
// 已购用户集合 Key: flash:purchased:{product_id} (Set)

const (
	FlashStockKeyPrefix     = "flash:stock:"
	FlashPurchasedKeyPrefix = "flash:purchased:"
	// 秒杀商品布隆过滤器
	FlashProductBloomKey    = "bloom:flash:product:ids"
)

// 秒杀活动库存预热过期时间
const (
	FlashStockExpiration = 3600 // 1小时
	RandomExpirationRange = 300 // 5分钟随机范围
)

// Lua 脚本：原子性检查库存与重复购买，并执行扣减
var flashBuyScript = redis.NewScript(`
	local stockKey = KEYS[1]
	local userSetKey = KEYS[2]
	local userId = ARGV[1]

	-- 1. 检查是否已经购买过 (一人一单)
	if redis.call("SISMEMBER", userSetKey, userId) == 1 then
		return -1 -- 表示重复购买
	end

	-- 2. 检查库存
	local stock = tonumber(redis.call("GET", stockKey))
	if stock == nil or stock <= 0 then
		return 0 -- 表示售罄或未预热
	end

	-- 3. 扣减库存并记录购买用户
	redis.call("DECR", stockKey)
	redis.call("SADD", userSetKey, userId)
	return 1 -- 表示抢购成功
`)

// ExecFlashBuy 执行秒杀 Lua 脚本
// 返回值: 1(成功), 0(售罄), -1(重复购买)
func ExecFlashBuy(ctx context.Context, productID uint64, userID uint64) (int, error) {
	// 1. 布隆过滤器检查 - 防缓存穿透
	exists, err := cache.ProductBloomFilter.ExistsString(strconv.FormatUint(productID, 10))
	if err != nil {
		// 布隆过滤器错误，继续执行但记录日志
		log.Printf("Error checking flash product bloom filter: %v", err)
	} else if !exists {
		// 布隆过滤器判断商品一定不存在，直接返回售罄
		return 0, nil
	}

	stockKey := fmt.Sprintf("%s%d", FlashStockKeyPrefix, productID)
	userSetKey := fmt.Sprintf("%s%d", FlashPurchasedKeyPrefix, productID)

	result, err := flashBuyScript.Run(ctx, cache.Rdb, []string{stockKey, userSetKey}, userID).Int()
	if err != nil && err != redis.Nil {
		return 0, err
	}
	return result, nil
}

// RollbackFlashBuy 回滚预扣库存与一人一单标记。
func RollbackFlashBuy(ctx context.Context, productID uint64, userID uint64) error {
	stockKey := fmt.Sprintf("%s%d", FlashStockKeyPrefix, productID)
	userSetKey := fmt.Sprintf("%s%d", FlashPurchasedKeyPrefix, productID)

	pipe := cache.Rdb.Pipeline()
	pipe.Incr(ctx, stockKey)
	pipe.SRem(ctx, userSetKey, userID)
	_, err := pipe.Exec(ctx)
	return err
}

// WarmupFlashStock 预热秒杀商品库存
// 防缓存雪崩: 使用随机过期时间
func WarmupFlashStock(ctx context.Context, productID uint64, stock int) error {
	// 1. 设置库存，使用随机过期时间
	stockKey := fmt.Sprintf("%s%d", FlashStockKeyPrefix, productID)
	expiration := cache.RandomExpiration(FlashStockExpiration, RandomExpirationRange)
	if err := cache.Rdb.Set(ctx, stockKey, stock, expiration).Err(); err != nil {
		return err
	}

	// 2. 清空已购用户集合
	userSetKey := fmt.Sprintf("%s%d", FlashPurchasedKeyPrefix, productID)
	if err := cache.Rdb.Del(ctx, userSetKey).Err(); err != nil {
		return err
	}

	// 3. 添加到布隆过滤器
	if err := cache.ProductBloomFilter.AddString(strconv.FormatUint(productID, 10)); err != nil {
		// 记录错误但继续执行
		log.Printf("Error adding flash product to bloom filter: %v", err)
		return err
	}

	return nil
}

// GetFlashStock 获取秒杀商品库存
func GetFlashStock(ctx context.Context, productID uint64) (int, error) {
	stockKey := fmt.Sprintf("%s%d", FlashStockKeyPrefix, productID)
	val, err := cache.Rdb.Get(ctx, stockKey).Int()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ClearFlashStock 清空秒杀商品库存
func ClearFlashStock(ctx context.Context, productID uint64) error {
	stockKey := fmt.Sprintf("%s%d", FlashStockKeyPrefix, productID)
	userSetKey := fmt.Sprintf("%s%d", FlashPurchasedKeyPrefix, productID)

	pipe := cache.Rdb.Pipeline()
	pipe.Del(ctx, stockKey)
	pipe.Del(ctx, userSetKey)
	_, err := pipe.Exec(ctx)
	return err
}
