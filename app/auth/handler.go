package main

import (
	"context"
	"fmt"

	auth "github.com/cloudwego/biz-demo/gomall/app/auth/kitex_gen/auth"
)

// AuthServiceImpl implements the last service interface defined in the IDL.
type AuthServiceImpl struct{}

// DeliverTokenByRPC implements the AuthServiceImpl interface.
func (s *AuthServiceImpl) DeliverTokenByRPC(ctx context.Context, req *auth.DeliverTokenReq) (resp *auth.DeliveryResp, err error) {
	// 使用Hertz JWT中间件生成token
	token, err := authMiddleware.LoginHandler(ctx, &User{UserID: req.UserId})
	if err != nil {
		return nil, fmt.Errorf("generate token failed: %w", err)
	}

	return &auth.DeliveryResp{
		Token: token,
	}, nil
}

// VerifyTokenByRPC implements the AuthServiceImpl interface.
func (s *AuthServiceImpl) VerifyTokenByRPC(ctx context.Context, req *auth.VerifyTokenReq) (resp *auth.VerifyResp, err error) {
	// 使用Hertz JWT中间件验证token
	claims, err := authMiddleware.GetClaimsFromJWT(ctx, req.Token)
	if err != nil {
		return &auth.VerifyResp{Res: false}, nil
	}

	if claims == nil {
		return &auth.VerifyResp{Res: false}, nil
	}

	return &auth.VerifyResp{Res: true}, nil
}
