package model

import "time"

// User 映射数据库 users 表
type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"uniqueIndex:uk_username;type:varchar(64);not null;default:''" json:"username"`
	PasswordHash string    `gorm:"type:varchar(128);not null;default:''" json:"-"` // JSON 序列化时忽略密码
	Status       int8      `gorm:"not null;default:0" json:"status"` // 0:已注册 1:活跃 2:禁用 3:已删除
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}