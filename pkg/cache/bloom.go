package cache

import (
	"math"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

// BloomFilter Redis 布隆过滤器
// 用于防止缓存穿透
// 基本原理: 多个哈希函数 + 位图
// 特点: 存在误判，不存在绝对准确
// 适用场景: 判断一个元素是否可能存在

type BloomFilter struct {
	rdb       *redis.Client
	key       string  // 布隆过滤器在 Redis 中的键
	size      uint    // 位图大小 (bit)
	k         uint    // 哈希函数个数
	hashFuncs []func([]byte) uint // 哈希函数列表
}

// NewBloomFilter 创建新的布隆过滤器
// estimatedItems: 预计元素数量
// falsePositiveRate: 期望的误判率 (0.01 = 1%)
func NewBloomFilter(rdb *redis.Client, key string, estimatedItems uint, falsePositiveRate float64) *BloomFilter {
	// 计算位图大小和哈希函数个数
	size := calculateSize(estimatedItems, falsePositiveRate)
	k := calculateHashCount(size, estimatedItems)

	filter := &BloomFilter{
		rdb:   rdb,
		key:   key,
		size:  size,
		k:     k,
	}

	// 初始化哈希函数
	filter.hashFuncs = make([]func([]byte) uint, k)
	for i := uint(0); i < k; i++ {
		filter.hashFuncs[i] = func(data []byte) uint {
			// 使用简单的哈希函数，实际可使用更复杂的
			hash := uint(5381)
			for _, b := range data {
				hash = ((hash << 5) + hash) + uint(b)
			}
			return hash
		}
	}

	return filter
}

// Add 添加元素到布隆过滤器
func (b *BloomFilter) Add(data []byte) error {
	ctx := Ctx
	pipe := b.rdb.Pipeline()

	for i := uint(0); i < b.k; i++ {
		hash := b.hashFuncs[i](data) % b.size
		pipe.SetBit(ctx, b.key, int64(hash), 1)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// AddString 添加字符串到布隆过滤器
func (b *BloomFilter) AddString(data string) error {
	return b.Add([]byte(data))
}

// Exists 检查元素是否可能存在
// 返回 true: 可能存在
// 返回 false: 一定不存在
func (b *BloomFilter) Exists(data []byte) (bool, error) {
	ctx := Ctx
	pipe := b.rdb.Pipeline()

	cmds := make([]*redis.IntCmd, b.k)
	for i := uint(0); i < b.k; i++ {
		hash := b.hashFuncs[i](data) % b.size
		cmds[i] = pipe.GetBit(ctx, b.key, int64(hash))
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	// 只要有一个位为 0，元素一定不存在
	for _, cmd := range cmds {
		val, err := cmd.Result()
		if err != nil {
			return false, err
		}
		if val == 0 {
			return false, nil
		}
	}

	return true, nil
}

// ExistsString 检查字符串是否可能存在
func (b *BloomFilter) ExistsString(data string) (bool, error) {
	return b.Exists([]byte(data))
}

// Clear 清空布隆过滤器
func (b *BloomFilter) Clear() error {
	return b.rdb.Del(Ctx, b.key).Err()
}

// 计算位图大小 (bit)
func calculateSize(n uint, p float64) uint {
	m := -float64(n) * math.Log(p) / (math.Ln2 * math.Ln2)
	return uint(math.Ceil(m))
}

// 计算哈希函数个数
func calculateHashCount(m, n uint) uint {
	k := (float64(m) / float64(n)) * math.Ln2
	return uint(math.Ceil(k))
}

// 全局布隆过滤器实例
var (
	// ProductBloomFilter 商品ID布隆过滤器
	ProductBloomFilter *BloomFilter
)

// InitBloomFilters 初始化布隆过滤器
func InitBloomFilters() {
	// 商品ID布隆过滤器
	// 预估100万商品，误判率1%
	ProductBloomFilter = NewBloomFilter(Rdb, "bloom:product:ids", 1000000, 0.01)
}

// 生成随机过期时间 (防缓存雪崩)
// base: 基础过期时间 (秒)
// random: 随机范围 (秒)
func RandomExpiration(base, random int) time.Duration {
	if random <= 0 {
		return time.Duration(base) * time.Second
	}
	rnd := rand.Intn(random)
	return time.Duration(base+rnd) * time.Second
}