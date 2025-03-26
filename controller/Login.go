package controller

import (
	"go-gorm/model"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	var loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	var user model.User
	if err := DB.Where("email = ?", loginData.Email).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.CheckPassword(loginData.Password) {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(200, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
	})
}
