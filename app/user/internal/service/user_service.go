package service

import (
	"context"
	"errors"
	"log"
	"strconv"

	"poptoy-flashsale/app/user/internal/model"
	"poptoy-flashsale/app/user/internal/repository"
	pkgCache "poptoy-flashsale/pkg/cache"
	"poptoy-flashsale/pkg/e"
	"poptoy-flashsale/pkg/jwt"

	"github.com/redis/go-redis/v9"
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

type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
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

	tokenVersion, err := ensureTokenVersion(context.Background(), user.ID)
	if err != nil {
		log.Printf("[User Service] 获取 token version 失败: %v\n", err)
		return e.Error, "", "", err
	}

	// 签发双 Token
	accessToken, refreshToken, err := jwt.GenerateTokens(user.ID, user.Username, tokenVersion)
	if err != nil {
		log.Printf("[User Service] JWT生成失败: %v\n", err)
		return e.Error, "", "", err
	}

	return e.Success, accessToken, refreshToken, nil
}

// Refresh 刷新 Access/Refresh 双 Token。
func Refresh(req *RefreshReq) (int, string, string, error) {
	claims, err := jwt.ParseToken(req.RefreshToken)
	if err != nil || !claims.IsRefresh {
		return e.Unauthorized, "", "", errors.New(e.GetMsg(e.Unauthorized))
	}

	tokenVersion, err := ensureTokenVersion(context.Background(), claims.UserID)
	if err != nil {
		log.Printf("[User Service] 刷新时读取 token version 失败: %v\n", err)
		return e.Error, "", "", err
	}
	if tokenVersion != claims.TokenVersion {
		return e.Unauthorized, "", "", errors.New(e.GetMsg(e.Unauthorized))
	}

	accessToken, refreshToken, err := jwt.GenerateTokens(claims.UserID, claims.Username, tokenVersion)
	if err != nil {
		log.Printf("[User Service] 刷新生成 JWT 失败: %v\n", err)
		return e.Error, "", "", err
	}

	return e.Success, accessToken, refreshToken, nil
}

// Logout 登出并提升 Token 版本号，使旧 Token 立即失效。
func Logout(userID uint64) (int, error) {
	if _, err := pkgCache.Rdb.Incr(context.Background(), jwt.TokenVersionKey(userID)).Result(); err != nil {
		log.Printf("[User Service] 退出登录提升 token version 失败: %v\n", err)
		return e.Error, err
	}
	return e.Success, nil
}

func ensureTokenVersion(ctx context.Context, userID uint64) (int64, error) {
	key := jwt.TokenVersionKey(userID)
	val, err := pkgCache.Rdb.Get(ctx, key).Result()
	if err == nil {
		parsed, parseErr := strconv.ParseInt(val, 10, 64)
		if parseErr != nil {
			return 0, parseErr
		}
		return parsed, nil
	}

	if errors.Is(err, redis.Nil) {
		if setErr := pkgCache.Rdb.Set(ctx, key, 1, 0).Err(); setErr != nil {
			return 0, setErr
		}
		return 1, nil
	}

	return 0, err
}
