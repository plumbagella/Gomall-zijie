package service

import (
	"context"

	"github.com/cloudwego/biz-demo/gomall/app/auth/biz/utils"
	auth "github.com/cloudwego/biz-demo/gomall/app/auth/kitex_gen/auth"
)

type VerifyTokenByRPCService struct {
	ctx context.Context
} // NewVerifyTokenByRPCService new VerifyTokenByRPCService
func NewVerifyTokenByRPCService(ctx context.Context) *VerifyTokenByRPCService {
	return &VerifyTokenByRPCService{ctx: ctx}
}

// Run create note info
func (s *VerifyTokenByRPCService) Run(req *auth.VerifyTokenReq) (resp *auth.VerifyResp, err error) {
	// 验证JWT token
	_, err = utils.ParseToken(req.Token)
	if err != nil {
		return &auth.VerifyResp{
			Res: false,
		}, nil
	}

	// token有效且未过期
	return &auth.VerifyResp{
		Res: true,
	}, nil
}
