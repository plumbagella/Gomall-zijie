package service

import (
	"context"

	"github.com/cloudwego/biz-demo/gomall/app/product/biz/dal/mysql"
	"github.com/cloudwego/biz-demo/gomall/app/product/biz/dal/redis"
	"github.com/cloudwego/biz-demo/gomall/app/product/biz/model"
	product "github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/product"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"gorm.io/gorm"
)

type CreateProductService struct {
	ctx context.Context
}

// NewCreateProductService new CreateProductService
func NewCreateProductService(ctx context.Context) *CreateProductService {
	return &CreateProductService{ctx: ctx}
}

// Run create note info
func (s *CreateProductService) Run(req *product.CreateProductReq) (resp *product.CreateProductResp, err error) {
	err = mysql.DB.Transaction(func(tx *gorm.DB) error {
		// 查询数据库中的 Categories
		var existingCategories []model.Category
		if len(req.Categories) > 0 {
			// 使用缓存查询 Categories
			categoryQuery := model.NewCategoryQuery(s.ctx, tx)
			cachedCategoryQuery := model.NewCachedCategoryQuery(categoryQuery, redis.RedisClient)

			for _, categoryName := range req.Categories {
				category, err := cachedCategoryQuery.GetByName(categoryName)
				if err != nil {
					return err
				}
				if category.Name == "" {
					return kerrors.NewBizStatusError(40001, "Invalid category name(s) provided")
				}
				existingCategories = append(existingCategories, category)
			}
		}

		// 组装 Product 数据
		newProduct := &model.Product{
			Name:        req.Name,
			Description: req.Description,
			Picture:     req.Picture,
			Price:       req.Price,
			Categories:  existingCategories, // 只关联已存在的 categories
		}

		// 创建 Product
		if err := model.CreateProduct(tx, s.ctx, newProduct); err != nil {
			return err
		}

		// 组装响应
		var categoryNames []string
		for _, category := range existingCategories {
			categoryNames = append(categoryNames, category.Name)
		}

		resp = &product.CreateProductResp{
			Product: &product.Product{
				Id:          uint32(newProduct.ID),
				Name:        newProduct.Name,
				Description: newProduct.Description,
				Picture:     newProduct.Picture,
				Price:       newProduct.Price,
				Categories:  categoryNames,
			},
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
