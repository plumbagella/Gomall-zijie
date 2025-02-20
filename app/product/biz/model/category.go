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

package model

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Category struct {
	Base
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Products    []Product `json:"product" gorm:"many2many:product_category;constraint:OnDelete:CASCADE"`
}

func (c Category) TableName() string {
	return "category"
}

func GetProductsByCategoryName(db *gorm.DB, ctx context.Context, name string) (category []Category, err error) {
	err = db.WithContext(ctx).Model(&Category{}).Where(&Category{Name: name}).Preload("Products").Find(&category).Error
	return category, err
}

type CategoryQuery struct {
	ctx context.Context
	db  *gorm.DB
}

func (c CategoryQuery) GetByName(name string) (category Category, err error) {
	err = c.db.WithContext(c.ctx).Model(&Category{}).Where(&Category{Name: name}).Preload("Products").First(&category).Error
	return
}

func NewCategoryQuery(ctx context.Context, db *gorm.DB) CategoryQuery {
	return CategoryQuery{ctx: ctx, db: db}
}

type CachedCategoryQuery struct {
	categoryQuery CategoryQuery
	cacheClient   *redis.Client
	prefix        string
}

func (c CachedCategoryQuery) GetByName(name string) (category Category, err error) {
	cacheKey := fmt.Sprintf("%s_%s_%s", c.prefix, "category_by_name", name)
	cachedResult := c.cacheClient.Get(c.categoryQuery.ctx, cacheKey)

	err = func() error {
		err1 := cachedResult.Err()
		if err1 != nil {
			return err1
		}
		cachedResultByte, err2 := cachedResult.Bytes()
		if err2 != nil {
			return err2
		}
		// Optionally check if category is valid, or if it's not in the cache yet
		err3 := json.Unmarshal(cachedResultByte, &category)
		if err3 != nil {
			return err3
		}
		return nil
	}()

	// If category not found, load from DB and cache
	if err != nil {
		err = c.categoryQuery.db.WithContext(c.categoryQuery.ctx).Where("name = ?", name).First(&category).Error
		if err != nil {
			return Category{}, err
		}
		// 加入缓存
		encodedCategory, err := json.Marshal(category)
		if err != nil {
			return Category{}, err
		}
		// 设置随机的过期时间，避免缓存同时过期
		randomExpiry := time.Duration(rand.Intn(10)) * time.Minute
		_ = c.cacheClient.Set(c.categoryQuery.ctx, cacheKey, encodedCategory, randomExpiry)
	}
	return
}

func NewCachedCategoryQuery(categoryQuery CategoryQuery, cacheClient *redis.Client) CachedCategoryQuery {
	return CachedCategoryQuery{categoryQuery: categoryQuery, cacheClient: cacheClient, prefix: "cloudwego_shop"}
}