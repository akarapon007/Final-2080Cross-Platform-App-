package controller

import (
	"go-gorm/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func ChangePassword(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	var passwordData struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&passwordData); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	var user model.User
	if err := DB.First(&user, userID).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	if !user.CheckPassword(passwordData.OldPassword) {
		c.JSON(401, gin.H{"error": "Incorrect old password"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordData.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := DB.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(200, gin.H{"message": "Password updated successfully"})
}
