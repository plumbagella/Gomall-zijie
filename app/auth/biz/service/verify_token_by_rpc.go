package service

import (
	"context"

	auth "github.com/cloudwego/biz-demo/gomall/app/auth/kitex_gen/auth"
	"github.com/golang-jwt/jwt/v5"
)

type VerifyTokenByRPCService struct {
	ctx context.Context
} // NewVerifyTokenByRPCService new VerifyTokenByRPCService
func NewVerifyTokenByRPCService(ctx context.Context) *VerifyTokenByRPCService {
	return &VerifyTokenByRPCService{ctx: ctx}
}

// Run create note info
func (s *VerifyTokenByRPCService) Run(req *auth.VerifyTokenReq) (resp *auth.VerifyResp, err error) {
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return &auth.VerifyResp{Res: false}, nil
	}

	if !token.Valid {
		return &auth.VerifyResp{Res: false}, nil
	}

	return &auth.VerifyResp{Res: true}, nil
}
