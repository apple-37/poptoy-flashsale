package repository

import (
	"errors"

	"poptoy-flashsale/app/user/internal/model"
	"poptoy-flashsale/pkg/mysql"

	"gorm.io/gorm"
)

// CreateUser 创建新用户
func CreateUser(user *model.User) error {
	err := mysql.DB.Create(user).Error
	return err
}

// GetUserByUsername 根据用户名查询用户
func GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := mysql.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 用户不存在不应视为系统错误
		}
		return nil, err // 其他数据库异常
	}
	return &user, nil
}

// UpdateUserStatus 更新用户状态
func UpdateUserStatus(userID uint64, status int8) error {
	return mysql.DB.Model(&model.User{}).Where("id = ?", userID).Update("status", status).Error
}

// UpdateUserStatusIfCurrent 仅当当前状态等于 expectedStatus 时更新状态。
func UpdateUserStatusIfCurrent(userID uint64, expectedStatus int8, status int8) (bool, error) {
	res := mysql.DB.Model(&model.User{}).
		Where("id = ? AND status = ?", userID, expectedStatus).
		Update("status", status)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// UpdateUserStatusIfCurrentIn 仅当当前状态在 expectedStatuses 中时更新状态。
func UpdateUserStatusIfCurrentIn(userID uint64, expectedStatuses []int8, status int8) (bool, error) {
	res := mysql.DB.Model(&model.User{}).
		Where("id = ? AND status IN ?", userID, expectedStatuses).
		Update("status", status)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}