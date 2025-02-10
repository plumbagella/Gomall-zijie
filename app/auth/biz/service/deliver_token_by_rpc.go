package service

import (
	"context"
	"fmt"
	"time"

	auth "github.com/cloudwego/biz-demo/gomall/app/auth/kitex_gen/auth"
	"github.com/golang-jwt/jwt/v5"
)

// 定义一个密钥，实际应用中应该从配置中读取
var jwtSecret = []byte("gomall_secret_key")

type Claims struct {
	UserID int32
	jwt.RegisteredClaims
}

type DeliverTokenByRPCService struct {
	ctx context.Context
} // NewDeliverTokenByRPCService new DeliverTokenByRPCService
func NewDeliverTokenByRPCService(ctx context.Context) *DeliverTokenByRPCService {
	return &DeliverTokenByRPCService{ctx: ctx}
}

// Run create note info
func (s *DeliverTokenByRPCService) Run(req *auth.DeliverTokenReq) (resp *auth.DeliveryResp, err error) {
	claims := Claims{
		UserID: req.UserId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24小时过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("generate token failed: %w", err)
	}

	return &auth.DeliveryResp{
		Token: signedToken,
	}, nil
}
