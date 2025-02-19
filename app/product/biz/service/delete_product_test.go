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

func TestDeleteProduct_Run(t *testing.T) {
	
	ctx := context.Background()
	s := NewDeleteProductService(ctx)

	// Test case: successful deletion
	req := &product.DeleteProductReq{Id: 1} // Assuming product with ID 1 exists
	resp, err := s.Run(req)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success to be true, got %v", resp.Success)
	}

	// Test case: unsuccessful deletion (non-existing product)
	req = &product.DeleteProductReq{Id: 9999} // Assuming product with ID 9999 does not exist
	resp, err = s.Run(req)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if resp.Success {
		t.Errorf("expected success to be false, got %v", resp.Success)
	}
}
