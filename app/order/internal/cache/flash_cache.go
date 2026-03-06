package cache

import (
	"context"
	"fmt"
	"poptoy-flashsale/pkg/cache"
	"github.com/redis/go-redis/v9"
)

// 预热数据的 Key 规则
// 库存 Key: flash:stock:{product_id} (String)
// 已购用户集合 Key: flash:purchased:{product_id} (Set)

const (
	FlashStockKeyPrefix     = "flash:stock:"
	FlashPurchasedKeyPrefix = "flash:purchased:"
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
	stockKey := fmt.Sprintf("%s%d", FlashStockKeyPrefix, productID)
	userSetKey := fmt.Sprintf("%s%d", FlashPurchasedKeyPrefix, productID)

	result, err := flashBuyScript.Run(ctx, cache.Rdb, []string{stockKey, userSetKey}, userID).Int()
	if err != nil && err != redis.Nil {
		return 0, err
	}
	return result, nil
}