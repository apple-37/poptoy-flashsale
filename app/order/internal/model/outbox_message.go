package model

import "time"

const (
	OutboxStatusPending int8 = 0
	OutboxStatusSent    int8 = 1
	OutboxStatusFailed  int8 = 2

	OutboxTypeCreateOrderTask = "create_order_task"
	OutboxTypeDelayCancelTask = "delay_cancel_task"
)

// OutboxMessage 本地事务消息表，用于保障 MQ 发送失败后的补偿重试。
type OutboxMessage struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement"`
	MessageType string     `gorm:"type:varchar(64);not null;index:idx_outbox_type_status_next"`
	BizKey      string     `gorm:"type:varchar(128);not null;uniqueIndex:uk_outbox_biz_key"`
	Payload     string     `gorm:"type:longtext;not null"`
	Status      int8       `gorm:"not null;default:0;index:idx_outbox_type_status_next"`
	RetryCount  int        `gorm:"not null;default:0"`
	NextRetryAt time.Time  `gorm:"not null;index:idx_outbox_type_status_next"`
	LastError   string     `gorm:"type:varchar(512);not null;default:''"`
	SentAt      *time.Time `gorm:"default:null"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
}

func (OutboxMessage) TableName() string {
	return "order_outbox_messages"
}
