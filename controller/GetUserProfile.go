package controller

import (
	"go-gorm/model"

	"github.com/gin-gonic/gin"
)

func GetUserProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uint) // สมมติว่ามี middleware ตั้งค่า user_id
	var user model.User
	if err := DB.First(&user, userID).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}
	c.JSON(200, gin.H{
		"id":      user.ID,
		"email":   user.Email,
		"name":    user.Name,
		"address": user.Address,
	})
}
