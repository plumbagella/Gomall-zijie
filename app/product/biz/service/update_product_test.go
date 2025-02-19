package service

import (
	"context"
	"testing"

	"github.com/cloudwego/biz-demo/gomall/app/product/biz/dal/mysql"
	product "github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/product"
	sql "gorm.io/driver/mysql"
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

func TestUpdateProduct_Run(t *testing.T) {
	ctx := context.Background()
	s := NewUpdateProductService(ctx)

	// Test case: successful update
	req := &product.UpdateProductReq{
		Id:          1, // Assuming product with ID 1 exists
		Name:        "Updated Product",
		Description: "Updated Description",
		Picture:     "updated.jpg",
		Price:       150.0,
		Categories:  []string{"Sticker"},
	}
	resp, err := s.Run(req)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if resp.Product.Name != req.Name {
		t.Errorf("expected name to be %v, got %v", req.Name, resp.Product.Name)
	}

	// Test case: unsuccessful update (non-existing product)
	req = &product.UpdateProductReq{
		Id:          9999, // Assuming product with ID 9999 does not exist
		Name:        "Non-existing Product",
		Description: "Non-existing Description",
		Picture:     "non-existing.jpg",
		Price:       200.0,
		Categories:  []string{"Category1", "Category2"},
	}
	resp, err = s.Run(req)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if resp != nil {
		t.Errorf("expected resp to be nil, got %v", resp)
	}
}
