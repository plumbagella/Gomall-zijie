package service

import (
	"context"

	"github.com/cloudwego/biz-demo/gomall/app/product/biz/dal/mysql"
	"github.com/cloudwego/biz-demo/gomall/app/product/biz/model"
	product "github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/product"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"gorm.io/gorm"
)

type UpdateProductService struct {
	ctx context.Context
} // NewUpdateProductService new UpdateProductService
func NewUpdateProductService(ctx context.Context) *UpdateProductService {
	return &UpdateProductService{ctx: ctx}
}

// Run update product info
func (s *UpdateProductService) Run(req *product.UpdateProductReq) (resp *product.UpdateProductResp, err error) {
	err = mysql.DB.Transaction(func(tx *gorm.DB) error {
		// 查询数据库中的 Categories
		var existingCategories []model.Category
		if len(req.Categories) > 0 {
			if err := tx.Where("name IN (?)", req.Categories).Find(&existingCategories).Error; err != nil {
				return err
			}

			// 确保所有请求的 category 都已存在
			if len(existingCategories) != len(req.Categories) {
				return kerrors.NewBizStatusError(40001, "Invalid category name(s) provided")
			}
		}

		// 组装 Product 数据
		updateProduct := &model.Product{
			Name:        req.Name,
			Description: req.Description,
			Picture:     req.Picture,
			Price:       req.Price,
			Categories:  existingCategories, // 只关联已存在的 categories
		}

		// 更新 Product
		if err := model.UpdateProduct(tx, s.ctx, int(req.Id), updateProduct); err != nil {
			return err
		}

		// 组装响应
		var categoryNames []string
		for _, category := range existingCategories {
			categoryNames = append(categoryNames, category.Name)
		}

		resp = &product.UpdateProductResp{
			Product: &product.Product{
				Id:          req.Id,
				Name:        updateProduct.Name,
				Description: updateProduct.Description,
				Picture:     updateProduct.Picture,
				Price:       updateProduct.Price,
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
