package service

import (
	"context"

	"github.com/cloudwego/biz-demo/gomall/app/auth/biz/utils"
	auth "github.com/cloudwego/biz-demo/gomall/app/auth/kitex_gen/auth"
)

type DeliverTokenByRPCService struct {
	ctx context.Context
} // NewDeliverTokenByRPCService new DeliverTokenByRPCService
func NewDeliverTokenByRPCService(ctx context.Context) *DeliverTokenByRPCService {
	return &DeliverTokenByRPCService{ctx: ctx}
}

// Run create note info
func (s *DeliverTokenByRPCService) Run(req *auth.DeliverTokenReq) (resp *auth.DeliveryResp, err error) {
	// 生成JWT token
	token, err := utils.GenerateToken(req.UserId)
	if err != nil {
		return nil, err
	}

	return &auth.DeliveryResp{
		Token: token,
	}, nil
}
