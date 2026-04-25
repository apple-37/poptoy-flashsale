package repository

import (
	"errors"
	"time"

	"poptoy-flashsale/app/order/internal/model"
	"poptoy-flashsale/pkg/mysql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreateOutboxMessage(msg *model.OutboxMessage) error {
	return mysql.DB.Create(msg).Error
}

func CreateOutboxMessageIfNotExists(msg *model.OutboxMessage) error {
	return mysql.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "biz_key"}},
		DoNothing: true,
	}).Create(msg).Error
}

func MarkOutboxMessageSent(id uint64) error {
	now := time.Now()
	return mysql.DB.Model(&model.OutboxMessage{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        model.OutboxStatusSent,
			"sent_at":       &now,
			"last_error":    "",
			"next_retry_at": now,
		}).Error
}

func MarkOutboxMessageFailed(id uint64, errMsg string) error {
	return mysql.DB.Model(&model.OutboxMessage{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     model.OutboxStatusFailed,
			"last_error": errMsg,
		}).Error
}

func ScheduleOutboxRetry(id uint64, errMsg string, nextRetryAt time.Time) error {
	return mysql.DB.Model(&model.OutboxMessage{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        model.OutboxStatusPending,
			"last_error":    errMsg,
			"next_retry_at": nextRetryAt,
			"retry_count":   gorm.Expr("retry_count + 1"),
		}).Error
}

func ListDuePendingOutbox(limit int) ([]model.OutboxMessage, error) {
	var msgs []model.OutboxMessage
	err := mysql.DB.Where("status = ? AND next_retry_at <= ?", model.OutboxStatusPending, time.Now()).
		Order("id ASC").
		Limit(limit).
		Find(&msgs).Error
	return msgs, err
}

func GetOutboxMessageByBizKey(bizKey string) (*model.OutboxMessage, error) {
	var msg model.OutboxMessage
	err := mysql.DB.Where("biz_key = ?", bizKey).First(&msg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &msg, nil
}
