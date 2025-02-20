package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/cloudwego/biz-demo/gomall/app/product/biz/dal/mysql" // import redis package
	"github.com/cloudwego/biz-demo/gomall/app/product/biz/dal/redis"
	product "github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/product"
	redisv9 "github.com/redis/go-redis/v9"
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

	// Initialize Redis
	redis.RedisClient = redisv9.NewClient(&redisv9.Options{
		Addr:     "127.0.0.1:6379",
		Username: "",
		Password: "",
		DB:       0,
	})
	if err := redis.RedisClient.Ping(context.Background()).Err(); err != nil {
		panic(err)
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
		Categories:  []string{"Sticker"},
	}
	resp, err := s.Run(req)
	t.Logf("err: %v", err)
	t.Logf("resp: %v", resp)

	productKey := fmt.Sprintf("%s_%s_%s", "cloudwego_shop", "category_by_name", "Sticker") // Example key format for Redis

	// Retrieve the product from Redis
	productData, err := redis.RedisClient.Get(ctx, productKey).Result()
	if err != nil {
		t.Fatalf("Failed to fetch product from Redis: %v", err)
	}

	// Verify the product data
	t.Logf("Product data from Redis: %s", productData)

	// todo: edit your unit test
}
