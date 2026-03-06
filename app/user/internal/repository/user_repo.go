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