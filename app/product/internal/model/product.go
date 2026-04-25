package model

import "time"

// ProductHot 商品热数据 (高频查询、参与锁竞争)
type ProductHot struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Title     string    `gorm:"type:varchar(128);not null;default:''" json:"title"`
	Price     float64   `gorm:"type:decimal(10,2);not null;default:0.00" json:"price"`
	Stock     int       `gorm:"not null;default:0" json:"stock"`
	Status    int8      `gorm:"not null;default:0" json:"status"` // 0:待审核 1:上架 2:下架 3:售罄 4:预售
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ProductHot) TableName() string {
	return "product_hot"
}

// ProductCold 商品冷数据 (长文本，低频查询)
type ProductCold struct {
	ProductID   uint64 `gorm:"primaryKey;column:product_id" json:"product_id"` // 对应热表ID，不自增
	Description string `gorm:"type:text;not null" json:"description"`
	ImagesJson  string `gorm:"type:json;not null" json:"images_json"`
}

func (ProductCold) TableName() string {
	return "product_cold"
}

// ProductDetail 业务层组装对象，用于返回完整数据
type ProductDetail struct {
	*ProductHot
	Description string `json:"description"`
	ImagesJson  string `json:"images_json"`
}