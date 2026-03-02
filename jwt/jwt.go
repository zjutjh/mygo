package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/zjutjh/mygo/kit"
)

type JWT[T any] struct {
	conf Config
}

// New 以指定配置创建实例
func New[T any](conf Config) *JWT[T] {
	return &JWT[T]{
		conf: conf,
	}
}

// GenerateToken 生成 JWT Token
func (j *JWT[T]) GenerateToken(identity T) (string, error) {
	claims := CustomClaims[T]{
		Identity: identity,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.conf.Issuer,
			Audience:  j.conf.Audience,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.conf.Expiration)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.conf.Secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ParseToken 解析 JWT Token
func (j *JWT[T]) ParseToken(tokenString string) (T, error) {
	var zero T
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims[T]{}, func(token *jwt.Token) (any, error) {
		return []byte(j.conf.Secret), nil
	})
	if err != nil {
		return zero, fmt.Errorf("解析JWT Token错误: %w", err)
	}
	claims, ok := token.Claims.(*CustomClaims[T])
	if !ok {
		return zero, fmt.Errorf("%w: 转化JWT Token为指定Claims结构失败", kit.ErrDataFormat)
	}
	return claims.Identity, nil
}
