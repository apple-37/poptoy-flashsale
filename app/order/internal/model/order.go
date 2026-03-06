package model

import "time"

type Order struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo   string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"order_no"`
	UserID    uint64    `gorm:"not null;index:idx_user_status_time" json:"user_id"`
	ProductID uint64    `gorm:"not null" json:"product_id"`
	Status    int8      `gorm:"not null;default:0;index:idx_user_status_time" json:"status"` // 0:Init, 1:Pending, 2:Paid, 3:Cancelled
	CreatedAt time.Time `gorm:"autoCreateTime;index:idx_user_status_time" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}