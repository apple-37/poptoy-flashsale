package service

import (
	"errors"
	"log"

	"poptoy-flashsale/app/user/internal/model"
	"poptoy-flashsale/app/user/internal/repository"
	"poptoy-flashsale/pkg/e"
	"poptoy-flashsale/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)
var dummyPasswordHash []byte

type RegisterReq struct {
	Username string `json:"username" binding:"required,min=4,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
func init() {
	dummyPasswordHash, _ = bcrypt.GenerateFromPassword([]byte("dummy_password_for_timing_attack"), bcrypt.DefaultCost)
}

// Register 处理用户注册逻辑
func Register(req *RegisterReq) (int, error) {
	// 1. 检查用户是否已存在
	existUser, err := repository.GetUserByUsername(req.Username)
	if err != nil {
		log.Printf("[User Service] 查询用户失败: %v\n", err)
		return e.Error, err
	}
	if existUser != nil {
		return e.UserExists, errors.New(e.GetMsg(e.UserExists))
	}

	// 2. 密码 Bcrypt 加密
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[User Service] 密码加密失败: %v\n", err)
		return e.Error, err
	}

	// 3. 落库
	newUser := &model.User{
		Username:     req.Username,
		PasswordHash: string(hashBytes),
	}
	if err := repository.CreateUser(newUser); err != nil {
		log.Printf("[User Service] 创建用户失败: %v\n", err)
		return e.Error, err
	}

	return e.Created, nil
}

// Login 处理用户登录逻辑 (包含双 Token 和防计时攻击)
func Login(req *LoginReq) (int, string, string, error) {
	user, err := repository.GetUserByUsername(req.Username)
	if err != nil {
		log.Printf("[User Service] 查询用户失败: %v\n", err)
		return e.Error, "", "", err
	}

	var hashToCompare []byte
	var isUserNil bool

	// 核心安全逻辑：无论用户是否存在，都会执行 bcrypt.CompareHashAndPassword
	// 从而保证响应时间一致，防止攻击者通过耗时盲猜库中是否存在该用户名
	if user == nil {
		hashToCompare = dummyPasswordHash
		isUserNil = true
	} else {
		hashToCompare = []byte(user.PasswordHash)
	}

	err = bcrypt.CompareHashAndPassword(hashToCompare, []byte(req.Password))
	
	// 如果用户不存在，或者密码比对失败，均返回统一的登录失败错误
	if isUserNil || err != nil {
		return e.LoginFailed, "", "", errors.New(e.GetMsg(e.LoginFailed))
	}

	// 签发双 Token
	accessToken, refreshToken, err := jwt.GenerateTokens(user.ID, user.Username)
	if err != nil {
		log.Printf("[User Service] JWT生成失败: %v\n", err)
		return e.Error, "", "", err
	}

	return e.Success, accessToken, refreshToken, nil
}