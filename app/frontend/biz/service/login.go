// Copyright 2024 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"context"

	auth "github.com/cloudwego/biz-demo/gomall/app/frontend/hertz_gen/frontend/auth"
	"github.com/cloudwego/biz-demo/gomall/app/frontend/infra/rpc"
	frontendutils "github.com/cloudwego/biz-demo/gomall/app/frontend/utils"
	authrpc "github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/auth"
	rpcuser "github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/user"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
)

type LoginService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewLoginService(Context context.Context, RequestContext *app.RequestContext) *LoginService {
	return &LoginService{RequestContext: RequestContext, Context: Context}
}

func (h *LoginService) Run(req *auth.LoginReq) (resp string, err error) {
	// 验证用户凭证
	loginResp, err := rpc.UserClient.Login(h.Context, &rpcuser.LoginReq{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return
	}

	// 获取 JWT token
	tokenResp, err := rpc.AuthClient.DeliverTokenByRPC(h.Context, &authrpc.DeliverTokenReq{
		UserId: loginResp.UserId,
	})
	if err != nil {
		return "", err
	}

	// 设置 token 到 cookie
	h.RequestContext.SetCookie(
		"jwt_token",
		tokenResp.Token,
		3600*24, // 24小时过期
		"/",
		"",
		protocol.CookieSameSiteLaxMode, // 设置 Cookie 同站策略
		false,
		true,
	)

	redirect := "/"
	if frontendutils.ValidateNext(req.Next) {
		redirect = req.Next
	}

	return redirect, nil
}
