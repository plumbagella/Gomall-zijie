package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/cloudwego/biz-demo/gomall/app/auth/conf"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/hertz-contrib/jwt"
	kitexlogrus "github.com/kitex-contrib/obs-opentelemetry/logging/logrus"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var identityKey = "user_id"

// LoginRequest 登录请求结构
type LoginRequest struct {
	UserID   int32  `form:"user_id,required" json:"user_id,required"`
	Password string `form:"password,required" json:"password,required"`
}

// User 用户信息
type User struct {
	UserID int32 `json:"user_id"`
}

// PingHandler 测试处理器
func PingHandler(ctx context.Context, c *app.RequestContext) {
	user, _ := c.Get(identityKey)
	c.JSON(200, utils.H{
		"message": fmt.Sprintf("user_id:%v", user.(*User).UserID),
	})
}

func main() {
	h := server.Default()

	// JWT中间件配置
	authMiddleware, err := jwt.New(&jwt.HertzJWTMiddleware{
		Realm:       "gomall auth",
		Key:         []byte("gomall_secret_key"),
		Timeout:     24 * time.Hour,
		MaxRefresh:  24 * time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*User); ok {
				return jwt.MapClaims{
					identityKey: v.UserID,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(ctx context.Context, c *app.RequestContext) interface{} {
			claims := jwt.ExtractClaims(ctx, c)
			return &User{
				UserID: int32(claims[identityKey].(float64)),
			}
		},
		Authenticator: func(ctx context.Context, c *app.RequestContext) (interface{}, error) {
			var loginVals LoginRequest
			if err := c.BindAndValidate(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			// 这里应该添加实际的用户验证逻辑
			// 目前简单演示，只要user_id > 0就认为验证通过
			if loginVals.UserID > 0 {
				return &User{
					UserID: loginVals.UserID,
				}, nil
			}

			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, ctx context.Context, c *app.RequestContext) bool {
			if v, ok := data.(*User); ok && v.UserID > 0 {
				return true
			}
			return false
		},
		Unauthorized: func(ctx context.Context, c *app.RequestContext, code int, message string) {
			c.JSON(code, map[string]interface{}{
				"code":    code,
				"message": message,
			})
		},
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
	})

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	if err := authMiddleware.MiddlewareInit(); err != nil {
		log.Fatal("authMiddleware.MiddlewareInit() Error:" + err.Error())
	}

	// 公开接口
	h.POST("/login", authMiddleware.LoginHandler)

	// 需要认证的接口组
	auth := h.Group("/auth")
	auth.GET("/refresh_token", authMiddleware.RefreshHandler)
	auth.Use(authMiddleware.MiddlewareFunc())
	{
		auth.GET("/ping", PingHandler)
	}

	// RPC服务处理
	go func() {
		// 这里启动RPC服务器
	}()

	h.Spin()
}

func kitexInit() (opts []server.Option) {
	// address
	addr, err := net.ResolveTCPAddr("tcp", conf.GetConf().Kitex.Address)
	if err != nil {
		panic(err)
	}
	opts = append(opts, server.WithServiceAddr(addr))

	// service info
	opts = append(opts, server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
		ServiceName: conf.GetConf().Kitex.Service,
	}))

	// klog
	logger := kitexlogrus.NewLogger()
	klog.SetLogger(logger)
	klog.SetLevel(conf.LogLevel())
	asyncWriter := &zapcore.BufferedWriteSyncer{
		WS: zapcore.AddSync(&lumberjack.Logger{
			Filename:   conf.GetConf().Kitex.LogFileName,
			MaxSize:    conf.GetConf().Kitex.LogMaxSize,
			MaxBackups: conf.GetConf().Kitex.LogMaxBackups,
			MaxAge:     conf.GetConf().Kitex.LogMaxAge,
		}),
		FlushInterval: time.Minute,
	}
	klog.SetOutput(asyncWriter)
	server.RegisterShutdownHook(func() {
		asyncWriter.Sync()
	})
	return
}
