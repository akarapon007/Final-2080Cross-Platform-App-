package controller

import (
	"go-gorm/model"

	"github.com/gin-gonic/gin"
)

func SearchProducts(c *gin.Context) {
	description := c.Query("description")
	minPrice := c.Query("min_price")
	maxPrice := c.Query("max_price")

	query := DB.Model(&model.Product{})
	if description != "" {
		query = query.Where("description LIKE ?", "%"+description+"%")
	}
	if minPrice != "" {
		query = query.Where("price >= ?", minPrice)
	}
	if maxPrice != "" {
		query = query.Where("price <= ?", maxPrice)
	}

	var products []model.Product
	if err := query.Find(&products).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to search products"})
		return
	}

	c.JSON(200, products)
}
