package service

import (
	"context"
	"testing"

	product "github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/product"
	sql "gorm.io/driver/mysql"
	"github.com/cloudwego/biz-demo/gomall/app/product/biz/dal/mysql"
	"gorm.io/gorm"
)

func init() {
	// DSN 配置
	dsn := "root:root@tcp(127.0.0.1:3306)/product?charset=utf8mb4&parseTime=True&loc=Local"

	// 初始化全局 mysql.DB
	var err error
	mysql.DB, err = gorm.Open(sql.Open(dsn), &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic("failed to connect to the database")
	}
}

func TestCreateProduct_Run(t *testing.T) {
	ctx := context.Background()
	s := NewCreateProductService(ctx)
	// init req and assert value

	req := &product.CreateProductReq{
		Name:        "Test Product",
		Description: "Test Description",
		Picture:     "test.jpg",
		Price:       100.0,
		Categories:  []string{"Wrong Category"},
	}
	resp, err := s.Run(req)
	t.Logf("err: %v", err)
	t.Logf("resp: %v", resp)

	// todo: edit your unit test
}
