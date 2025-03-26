package main

import (
	"fmt"
	"go-gorm/model"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ตัวแปรระดับ package
var DB *gorm.DB

// โมเดล

type Config struct {
	MySQL struct {
		DSN string `yaml:"dsn"`
	} `yaml:"mysql"`
}

// เชื่อมต่อฐานข้อมูล
func LoadConfig() (string, error) {
	var config Config
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %v", err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("failed to unmarshal config: %v", err)
	}
	return config.MySQL.DSN, nil
}

// เชื่อมต่อฐานข้อมูล
func ConnectDatabase() {
	dsn, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	var database *gorm.DB
	for i := 0; i < 3; i++ { // ลองเชื่อมต่อ 3 ครั้ง
		database, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database (attempt %d): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("Failed to connect to database after retries:", err)
	}

	// ตั้งค่า connection pool
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatal("Failed to get sql.DB:", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = database
	DB.AutoMigrate(&model.User{}, &model.Product{}, &model.Cart{}, &model.CartItem{})
}

// API Handlers
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

func GetUserProfile(c *gin.Context) {
	userID := uint(1) // สมมติว่าได้จาก middleware
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

func UpdateAddress(c *gin.Context) {
	userID := uint(1) // สมมติว่าได้จาก middleware
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

func ChangePassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}
	userID := uint(id)

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

	if err := DB.Model(&user).Update("password", passwordData.NewPassword).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(200, gin.H{"message": "Password updated successfully"})
}

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

func AddToCart(c *gin.Context) {
	userID := uint(1) // สมมติว่าได้จาก middleware
	cartName := c.Param("cart_name")
	var itemData struct {
		ProductID uint `json:"product_id"`
		Quantity  int  `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&itemData); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	var cart model.Cart
	if err := DB.Where("name = ? AND customer_id = ?", cartName, userID).First(&cart).Error; err != nil {
		cart = model.Cart{Name: cartName, CustomerID: userID}
		if err := DB.Create(&cart).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to create cart"})
			return
		}
	}

	var cartItem model.CartItem
	if err := DB.Where("cart_id = ? AND product_id = ?", cart.ID, itemData.ProductID).First(&cartItem).Error; err == nil {
		cartItem.Quantity += itemData.Quantity
		if err := DB.Save(&cartItem).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to update cart item"})
			return
		}
	} else {
		cartItem = model.CartItem{CartID: cart.ID, ProductID: itemData.ProductID, Quantity: itemData.Quantity}
		if err := DB.Create(&cartItem).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to add cart item"})
			return
		}
	}

	c.JSON(200, cartItem)
}

func GetCarts(c *gin.Context) {
	customerID := c.Query("customer_id")
	if customerID == "" {
		c.JSON(400, gin.H{"error": "customer_id is required"})
		return
	}

	var carts []model.Cart
	if err := DB.Preload("Items.Product").Where("customer_id = ?", customerID).Find(&carts).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to get carts"})
		return
	}

	c.JSON(200, carts)
}

func main() {
	ConnectDatabase()

	r := gin.Default()
	r.POST("/login", Login)
	r.GET("/users/me", GetUserProfile)
	r.PUT("/users/me/address", UpdateAddress)
	r.PUT("/users/:id/password", ChangePassword)
	r.GET("/products", SearchProducts)
	r.POST("/carts/:cart_name/items", AddToCart)
	r.GET("/carts", GetCarts)

	r.Run(":8080")
}
