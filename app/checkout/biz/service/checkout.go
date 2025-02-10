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
	"errors"
	"fmt"
	"strconv"

	"github.com/cloudwego/biz-demo/gomall/app/checkout/infra/mq"
	"github.com/cloudwego/biz-demo/gomall/app/checkout/infra/rpc"
	"github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/cart"
	checkout "github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/checkout"
	"github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/email"
	"github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/order"
	"github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/payment"
	"github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/product"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/protobuf/proto"
)

// CheckoutService 负责处理订单结账流程
type CheckoutService struct {
	ctx context.Context
}

// NewCheckoutService 用于创建一个CheckoutService实例
func NewCheckoutService(ctx context.Context) *CheckoutService {
	return &CheckoutService{ctx: ctx}
}

/*
Run 方法用于执行结账流程，主要包括以下步骤：
1. 获取购物车内容。
2. 根据购物车计算总金额及创建订单项。
3. 创建订单。
4. 清空购物车。
5. 发起支付请求。
6. 发送确认邮件。
7. 修改订单状态为已支付。
*/
func (s *CheckoutService) Run(req *checkout.CheckoutReq) (resp *checkout.CheckoutResp, err error) {
	// -------------------------------
	// STEP 1: 获取购物车内容
	// -------------------------------
	// 使用CartClient的GetCart方法获取用户购物车信息
	cartResult, err := rpc.CartClient.GetCart(s.ctx, &cart.GetCartReq{UserId: req.UserId})
	if err != nil {
		// 记录错误日志并返回格式化后的错误信息
		klog.Error(err)
		err = fmt.Errorf("GetCart.err:%v", err)
		return
	}
	// 检查购物车是否为空
	if cartResult == nil || cartResult.Cart == nil || len(cartResult.Cart.Items) == 0 {
		err = errors.New("cart is empty")
		return
	}

	// -------------------------------
	// STEP 2: 根据购物车计算总金额及创建订单项
	// -------------------------------
	var (
		oi    []*order.OrderItem // 存放订单项的切片
		total float32            // 总计金额
	)
	// 遍历购物车中的每个商品项
	for _, cartItem := range cartResult.Cart.Items {
		// 逐个查询每件商品的详细信息
		productResp, resultErr := rpc.ProductClient.GetProduct(s.ctx, &product.GetProductReq{Id: cartItem.ProductId})
		if resultErr != nil {
			// 如果查询出错，则记录错误并返回
			klog.Error(resultErr)
			err = resultErr
			return
		}
		// 如果商品不存在则忽略这个购物车项
		if productResp.Product == nil {
			continue
		}
		p := productResp.Product
		// 计算当前购物车项的花费：商品单价 * 数量
		cost := p.Price * float32(cartItem.Quantity)
		// 累加到总金额中
		total += cost
		// 添加订单项到订单项列表中
		oi = append(oi, &order.OrderItem{
			Item: &cart.CartItem{
				ProductId: cartItem.ProductId,
				Quantity:  cartItem.Quantity,
			},
			Cost: cost,
		})
	}

	// -------------------------------
	// STEP 3: 创建订单
	// -------------------------------
	// 构造订单请求，其中包含用户ID、货币类型以及订单项信息
	orderReq := &order.PlaceOrderReq{
		UserId:       req.UserId,
		UserCurrency: "USD",
		OrderItems:   oi,
		Email:        req.Email,
	}
	// 如果请求中包含地址信息，则进行地址转换和设置
	if req.Address != nil {
		addr := req.Address
		zipCodeInt, _ := strconv.Atoi(addr.ZipCode)
		orderReq.Address = &order.Address{
			StreetAddress: addr.StreetAddress,
			City:          addr.City,
			Country:       addr.Country,
			State:         addr.State,
			ZipCode:       int32(zipCodeInt),
		}
	}
	// 调用OrderClient的PlaceOrder方法向订单服务发送创建订单请求
	orderResult, err := rpc.OrderClient.PlaceOrder(s.ctx, orderReq)
	if err != nil {
		// 如果创建订单失败，则返回错误信息
		err = fmt.Errorf("PlaceOrder.err:%v", err)
		return
	}
	// 记录订单返回结果的日志
	klog.Info("orderResult", orderResult)

	// -------------------------------
	// STEP 4: 清空购物车
	// -------------------------------
	// 使用CartClient的EmptyCart方法清空用户购物车
	emptyResult, err := rpc.CartClient.EmptyCart(s.ctx, &cart.EmptyCartReq{UserId: req.UserId})
	if err != nil {
		err = fmt.Errorf("EmptyCart.err:%v", err)
		return
	}
	// 记录清空购物车后的返回结果
	klog.Info(emptyResult)

	// -------------------------------
	// STEP 5: 发起支付请求
	// -------------------------------
	// 如果订单创建成功，则提取订单ID信息
	var orderId string
	if orderResult != nil || orderResult.Order != nil {
		orderId = orderResult.Order.OrderId
	}
	// 构造支付请求，其中包含了用户信息、订单ID、支付金额以及信用卡信息
	payReq := &payment.ChargeReq{
		UserId:  req.UserId,
		OrderId: orderId,
		Amount:  total,
		CreditCard: &payment.CreditCardInfo{
			CreditCardNumber:          req.CreditCard.CreditCardNumber,
			CreditCardExpirationYear:  req.CreditCard.CreditCardExpirationYear,
			CreditCardExpirationMonth: req.CreditCard.CreditCardExpirationMonth,
			CreditCardCvv:             req.CreditCard.CreditCardCvv,
		},
	}
	// 调用PaymentClient的Charge方法发起支付
	paymentResult, err := rpc.PaymentClient.Charge(s.ctx, payReq)
	if err != nil {
		err = fmt.Errorf("Charge.err:%v", err)
		return
	}

	// -------------------------------
	// STEP 6: 发送确认邮件
	// -------------------------------
	// 构造Email请求数据，用于通知客户订单已创建成功
	data, _ := proto.Marshal(&email.EmailReq{
		From:        "from@example.com",
		To:          req.Email,
		ContentType: "text/plain",
		Subject:     "You just created an order in CloudWeGo shop",
		Content:     "You just created an order in CloudWeGo shop",
	})
	// 构造NATS消息，将邮件请求数据放入消息体中
	msg := &nats.Msg{
		Subject: "email",
		Data:    data,
		Header:  make(nats.Header),
	}
	// 使用OpenTelemetry的Propagator将上下文注入到消息Header中，便于链路追踪
	otel.GetTextMapPropagator().Inject(s.ctx, propagation.HeaderCarrier(msg.Header))
	// 发布消息到NATS队列
	_ = mq.Nc.PublishMsg(msg)
	// 记录支付结果的日志
	klog.Info(paymentResult)

	// -------------------------------
	// STEP 7: 修改订单状态为已支付
	// -------------------------------
	// 调用OrderClient修改订单状态为已支付
	klog.Info(orderResult)
	_, err = rpc.OrderClient.MarkOrderPaid(s.ctx, &order.MarkOrderPaidReq{
		UserId:  req.UserId,
		OrderId: orderId,
	})
	if err != nil {
		klog.Error(err)
		return
	}

	// -------------------------------
	// STEP 8: 返回响应结果
	// -------------------------------
	// 构造checkout响应，包含订单ID和支付交易ID
	resp = &checkout.CheckoutResp{
		OrderId:       orderId,
		TransactionId: paymentResult.TransactionId,
	}
	return
}
