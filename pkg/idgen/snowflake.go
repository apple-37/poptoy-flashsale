package idgen

import (
	"fmt"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	node     *snowflake.Node
	initOnce sync.Once
	initErr  error
)

// InitSnowflake 初始化雪花算法节点。
func InitSnowflake(nodeID int64) error {
	initOnce.Do(func() {
		node, initErr = snowflake.NewNode(nodeID)
	})
	return initErr
}

// NewOrderNo 生成全局唯一订单号。
func NewOrderNo() (string, error) {
	if node == nil {
		return "", fmt.Errorf("snowflake 节点未初始化")
	}
	return "ORD" + node.Generate().String(), nil
}
