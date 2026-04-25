package jwt

import (
	"errors"
	"fmt"
	"time"

	"poptoy-flashsale/pkg/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID       uint64 `json:"user_id"`
	Username     string `json:"username"`
	TokenVersion int64  `json:"token_version"`
	IsRefresh    bool   `json:"is_refresh"` // 标识是否为刷新 Token
	jwt.RegisteredClaims
}

// GenerateTokens 签发双 Token
func GenerateTokens(userID uint64, username string, tokenVersion int64) (accessToken string, refreshToken string, err error) {
	secretKey := []byte(config.GlobalConfig.JWT.Secret)
	nowTime := time.Now()

	accessExpireTime := nowTime.Add(time.Duration(config.GlobalConfig.JWT.AccessExpire) * time.Hour)
	refreshExpireTime := nowTime.Add(time.Duration(config.GlobalConfig.JWT.RefreshExpire) * time.Hour)

	// 1. 签发 Access Token
	accessClaims := Claims{
		UserID:       userID,
		Username:     username,
		TokenVersion: tokenVersion,
		IsRefresh:    false,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			Issuer:    "poptoy-system",
		},
	}
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(secretKey)
	if err != nil {
		return "", "", err
	}

	// 2. 签发 Refresh Token
	refreshClaims := Claims{
		UserID:       userID,
		Username:     username,
		TokenVersion: tokenVersion,
		IsRefresh:    true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			Issuer:    "poptoy-system",
		},
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(secretKey)
	return accessToken, refreshToken, err
}

// ParseToken 解析与校验 Token
func ParseToken(token string) (*Claims, error) {
	secretKey := []byte(config.GlobalConfig.JWT.Secret)
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// TokenVersionKey 返回用户 Token 版本号在 Redis 中的 key。
func TokenVersionKey(userID uint64) string {
	return fmt.Sprintf("auth:token_version:%d", userID)
}
