package controller

import (
	"go-gorm/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var DB *gorm.DB

func UpdateAddress(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	var updateData struct {
		Address string `json:"address"`
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	if err := DB.Model(&model.User{}).Where("id = ?", userID).Update("address", updateData.Address).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update address"})
		return
	}

	c.JSON(200, gin.H{"address": updateData.Address})
}
