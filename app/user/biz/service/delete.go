package service

import (
	"context"
	"errors"
	"strings"

	"github.com/cloudwego/biz-demo/gomall/app/user/biz/dal/mysql"
	"github.com/cloudwego/biz-demo/gomall/app/user/biz/model"
	user "github.com/cloudwego/biz-demo/gomall/app/user/kitex_gen/user"
)

type DeleteService struct {
	ctx context.Context
} // NewDeleteService new DeleteService
func NewDeleteService(ctx context.Context) *DeleteService {
	return &DeleteService{ctx: ctx}
}

// Run create note info
func (s *DeleteService) Run(req *user.DeleteReq) (resp *user.DeleteResp, err error) {
	// Finish your business logic.
	//删除用户,需要token验证
	if req.Token == "" {
		return &user.DeleteResp{Success: false}, errors.New("token is required")
	}
	//验证token
	token := req.Token
	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		return &user.DeleteResp{Success: false}, errors.New("invalid token")
	}
	if err = model.Delete(mysql.DB, s.ctx, &model.User{Base: model.Base{ID: int(req.UserId)}}); err != nil {
		return
	}
	return &user.DeleteResp{Success: true}, nil
}
