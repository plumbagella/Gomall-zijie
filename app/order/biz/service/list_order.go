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

	"github.com/cloudwego/biz-demo/gomall/app/order/biz/dal/mysql"
	"github.com/cloudwego/biz-demo/gomall/app/order/biz/model"
	"github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/cart"
	order "github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/order"
	"github.com/cloudwego/kitex/pkg/klog"
)

type ListOrderService struct {
	ctx context.Context
} // NewListOrderService new ListOrderService
func NewListOrderService(ctx context.Context) *ListOrderService {
	return &ListOrderService{ctx: ctx}
}

// Run 方法用于处理订单列表查询请求，根据传入的用户ID获取对应的订单数据，并构造返回的响应结构体
func (s *ListOrderService) Run(req *order.ListOrderReq) (resp *order.ListOrderResp, err error) {
	// 调用数据层方法 ListOrder 来查询指定用户（req.UserId）的订单记录，
	// mysql.DB 为数据库连接对象，s.ctx 为当前请求上下文
	orders, err := model.ListOrder(mysql.DB, s.ctx, req.UserId)
	if err != nil {
		// 如果查询过程中出现错误，则记录错误日志并返回错误
		klog.Errorf("model.ListOrder.err:%v", err)
		return nil, err
	}

	// 初始化用于存储构造好的订单列表
	var list []*order.Order

	// 遍历数据层返回的每一条订单记录（orders）
	for _, v := range orders {
		// 定义一个空切片，用于保存该订单中所有的订单项
		var items []*order.OrderItem

		// 遍历当前订单记录中的各个订单项（v.OrderItems）
		for _, v := range v.OrderItems {
			// 构造一个 order.OrderItem 对象，并包装内部的商品信息（cart.CartItem）
			// Cost 表示订单项的费用，Item 包含对应商品的ID和购买数量信息
			items = append(items, &order.OrderItem{
				Cost: v.Cost,
				Item: &cart.CartItem{
					ProductId: v.ProductId,
					Quantity:  v.Quantity,
				},
			})
		}

		// 构造order.Order对象，该对象用于描述一个完整的订单记录
		o := &order.Order{
			OrderId:      v.OrderId,                 // 订单唯一标识
			UserId:       v.UserId,                  // 下单用户ID
			UserCurrency: v.UserCurrency,            // 用户使用的货币类型
			Email:        v.Consignee.Email,         // 收件人邮箱（从订单中的收货信息中获取）
			CreatedAt:    int32(v.CreatedAt.Unix()), // 订单创建时间，转换为 Unix 时间戳并转成 int32
			Address: &order.Address{ // 构造订单的收货地址信息
				Country:       v.Consignee.Country,
				City:          v.Consignee.City,
				StreetAddress: v.Consignee.StreetAddress,
				ZipCode:       v.Consignee.ZipCode,
			},
			OrderItems: items, // 订单中所有订单项的集合
		}

		// 将构造好的订单添加到最终返回的订单列表中
		list = append(list, o)
	}

	// 构造返回结果，将订单列表包装进 ListOrderResp 对象中
	resp = &order.ListOrderResp{
		Orders: list,
	}
	// 返回响应结构体和 nil 错误
	return
}
