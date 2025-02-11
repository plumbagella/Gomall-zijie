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

package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/biz-demo/gomall/app/frontend/infra/rpc"
	"github.com/cloudwego/biz-demo/gomall/app/frontend/utils"
	"github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/auth"
	"github.com/cloudwego/hertz/pkg/app"
)

// GlobalAuth 全局认证中间件，不强制要求认证
func GlobalAuth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		token := extractToken(c)
		if token == "" {
			c.Next(ctx)
			return
		}

		// 验证 token
		resp, err := rpc.AuthClient.VerifyTokenByRPC(ctx, &auth.VerifyTokenReq{
			Token: token,
		})
		if err != nil || !resp.Res {
			c.Next(ctx)
			return
		}

		// token 有效，将用户 ID 存入上下文
		// 注意：由于 VerifyResp 中没有返回 user_id，这里可能需要从 token 中解析
		// 或者需要修改 auth 服务接口以返回用户 ID
		c.Next(ctx)
	}
}

// Auth 强制认证中间件
func Auth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		token := extractToken(c)
		if token == "" {
			redirectToLogin(c)
			return
		}

		// 验证 token
		resp, err := rpc.AuthClient.VerifyTokenByRPC(ctx, &auth.VerifyTokenReq{
			Token: token,
		})
		if err != nil || !resp.Res {
			redirectToLogin(c)
			return
		}

		// token 有效，将用户 ID 存入上下文
		// 同上，需要获取用户 ID
		c.Next(ctx)
	}
}

// extractToken 从请求头或 cookie 中提取 token
func extractToken(c *app.RequestContext) string {
	// 先从 Authorization 头中获取
	auth := string(c.GetHeader("Authorization"))
	if auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// 如果请求头中没有，尝试从 cookie 中获取
	token := c.Cookie("jwt_token")
	return string(token)
}

// redirectToLogin 重定向到登录页面
func redirectToLogin(c *app.RequestContext) {
	byteRef := c.GetHeader("Referer")
	ref := string(byteRef)
	next := "/sign-in"
	if ref != "" && utils.ValidateNext(ref) {
		next = fmt.Sprintf("%s?next=%s", next, ref)
	}
	c.Redirect(302, []byte(next))
	c.Abort()
}
