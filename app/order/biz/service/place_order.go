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
	"encoding/json"
	"fmt"

	"github.com/cloudwego/biz-demo/gomall/app/order/biz/dal/mysql"
	"github.com/cloudwego/biz-demo/gomall/app/order/biz/model"
	order "github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/order"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type PlaceOrderService struct {
	ctx context.Context
} // NewPlaceOrderService new PlaceOrderService
func NewPlaceOrderService(ctx context.Context) *PlaceOrderService {
	return &PlaceOrderService{ctx: ctx}
}

// Run create note info
func (s *PlaceOrderService) Run(req *order.PlaceOrderReq) (resp *order.PlaceOrderResp, err error) {
	// Finish your business logic.
	if len(req.OrderItems) == 0 {
		err = fmt.Errorf("OrderItems empty")
		return
	}

	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"order_queue", // name
		false,         // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to declare a queue: %v", err)
	}

	body, err := json.Marshal(req) // Serialize req to JSON
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal request: %v", err)
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        body,
        })
	if err != nil {
		return nil, fmt.Errorf("Failed to publish a message: %v", err)
	}

	resp = &order.PlaceOrderResp{
		Order: &order.OrderResult{
			OrderId: "Order is being processed",
		},
	}
	return
}

// CreateOrder processes the order creation logic
func CreateOrder(req *order.PlaceOrderReq) error {
	return mysql.DB.Transaction(func(tx *gorm.DB) error {
		orderId, _ := uuid.NewUUID()

		o := &model.Order{
			OrderId:      orderId.String(),
			OrderState:   model.OrderStatePlaced,
			UserId:       req.UserId,
			UserCurrency: req.UserCurrency,
			Consignee: model.Consignee{
				Email: req.Email,
			},
		}
		if req.Address != nil {
			a := req.Address
			o.Consignee.Country = a.Country
			o.Consignee.State = a.State
			o.Consignee.City = a.City
			o.Consignee.StreetAddress = a.StreetAddress
		}
		if err := tx.Create(o).Error; err != nil {
			return err
		}

		var itemList []*model.OrderItem
		for _, v := range req.OrderItems {
			itemList = append(itemList, &model.OrderItem{
				OrderIdRefer: o.OrderId,
				ProductId:    v.Item.ProductId,
				Quantity:     v.Item.Quantity,
				Cost:         v.Cost,
			})
		}
		if err := tx.Create(&itemList).Error; err != nil {
			return err
		}

		// Connect to RabbitMQ
		conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err != nil {
			return fmt.Errorf("Failed to connect to RabbitMQ: %v", err)
		}
		defer conn.Close()

		ch, err := conn.Channel()
		if err != nil {
			return fmt.Errorf("Failed to open a channel: %v", err)
		}
		defer ch.Close()

		q, err := ch.QueueDeclare(
			"order_response_queue", // name
			false,                  // durable
			false,                  // delete when unused
			false,                  // exclusive
			false,                  // no-wait
			nil,                    // arguments
		)
		if err != nil {
			return fmt.Errorf("Failed to declare a queue: %v", err)
		}

		response := struct {
			FinalOrderId string `json:"final_order_id"`
		}{
			FinalOrderId: o.OrderId,
		}

		body, err := json.Marshal(response)
		if err != nil {
			return fmt.Errorf("Failed to marshal response: %v", err)
		}

		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			})
		if err != nil {
			return fmt.Errorf("Failed to publish a message: %v", err)
		}


		return nil
	})
	// err = mysql.DB.Transaction(func(tx *gorm.DB) error {
	// 	orderId, _ := uuid.NewUUID()

	// 	o := &model.Order{
	// 		OrderId:      orderId.String(),
	// 		OrderState:   model.OrderStatePlaced,
	// 		UserId:       req.UserId,
	// 		UserCurrency: req.UserCurrency,
	// 		Consignee: model.Consignee{
	// 			Email: req.Email,
	// 		},
	// 	}
	// 	if req.Address != nil {
	// 		a := req.Address
	// 		o.Consignee.Country = a.Country
	// 		o.Consignee.State = a.State
	// 		o.Consignee.City = a.City
	// 		o.Consignee.StreetAddress = a.StreetAddress
	// 	}
	// 	if err := tx.Create(o).Error; err != nil {
	// 		return err
	// 	}

	// 	var itemList []*model.OrderItem
	// 	for _, v := range req.OrderItems {
	// 		itemList = append(itemList, &model.OrderItem{
	// 			OrderIdRefer: o.OrderId,
	// 			ProductId:    v.Item.ProductId,
	// 			Quantity:     v.Item.Quantity,
	// 			Cost:         v.Cost,
	// 		})
	// 	}
	// 	if err := tx.Create(&itemList).Error; err != nil {
	// 		return err
	// 	}
	// 	resp = &order.PlaceOrderResp{
	// 		Order: &order.OrderResult{
	// 			OrderId: orderId.String(),
	// 		},
	// 	}

	// 	return nil
	// })

	// return
}
