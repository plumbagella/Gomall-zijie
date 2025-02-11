package utils

import (
	"time"

	"github.com/cloudwego/biz-demo/gomall/app/auth/conf"
	"github.com/golang-jwt/jwt/v5"
)

var JwtSecret []byte = conf.JwtSecret // 在生产环境中应该使用环境变量配置

type Claims struct {
	UserID int32
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT token
func GenerateToken(userID int32) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // token有效期24小时
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JwtSecret)
}

// ParseToken 解析JWT token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return JwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
