package service

import (
	"context"
	"log"

	"github.com/cloudwego/biz-demo/gomall/app/product/biz/dal/mysql"
	"github.com/cloudwego/biz-demo/gomall/app/product/biz/model"
	product "github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/product"
)

type DeleteProductService struct {
	ctx context.Context
} // NewDeleteProductService new DeleteProductService
func NewDeleteProductService(ctx context.Context) *DeleteProductService {
	return &DeleteProductService{ctx: ctx}
}

// Run delete product info
func (s *DeleteProductService) Run(req *product.DeleteProductReq) (resp *product.DeleteProductResp, err error) {
	if mysql.DB == nil {
		log.Println("mysql.DB is nil before running test")
	} else {
		log.Println("Database connected successfully")
	}
	err = model.DeleteProduct(mysql.DB, s.ctx, int(req.Id))
	if err != nil {
		resp = &product.DeleteProductResp{
			Success: false,
		}
		return resp, err
	}
	resp = &product.DeleteProductResp{
		Success: true,
	}
	return resp, nil
}
