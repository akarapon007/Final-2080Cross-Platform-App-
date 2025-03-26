package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ตัวแปรระดับ package
var DB *gorm.DB

// โมเดล
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"-"`
	Name     string `json:"name"`
	Address  string `json:"address"`
}

type Product struct {
	ID          uint    `json:"id" gorm:"primaryKey"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type Cart struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	Name       string     `json:"name"`
	CustomerID uint       `json:"customer_id"`
	Items      []CartItem `json:"items"`
}

type CartItem struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	CartID    uint    `json:"cart_id"`
	ProductID uint    `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Product   Product `json:"product" gorm:"foreignKey:ProductID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// เชื่อมต่อฐานข้อมูล
func ConnectDatabase() {
	dsn := "cp_65011212080:65011212080@csmsu@tcp(202.28.34.197:3306)/cp_65011212080?charset=utf8mb4&parseTime=True&loc=Local"
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	DB = database
	DB.AutoMigrate(&User{}, &Product{}, &Cart{}, &CartItem{})
}

// จำลองข้อมูล
func SeedData() {
	for i := 1; i <= 10; i++ {
		user := User{
			Email:    fmt.Sprintf("user%d@example.com", i),
			Password: "password",
			Name:     fmt.Sprintf("User %d", i),
			Address:  fmt.Sprintf("Address %d", i),
		}
		DB.Create(&user)

		product := Product{
			Name:        fmt.Sprintf("Product %d", i),
			Description: fmt.Sprintf("Description %d", i),
			Price:       float64(i * 100),
		}
		DB.Create(&product)

		cart := Cart{
			Name:       fmt.Sprintf("Cart %d", i),
			CustomerID: uint(i),
		}
		DB.Create(&cart)

		for j := 1; j <= 3; j++ {
			cartItem := CartItem{
				CartID:    cart.ID,
				ProductID: uint(j),
				Quantity:  j,
			}
			DB.Create(&cartItem)
		}
	}
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

	var user User
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
	var user User
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

	if err := DB.Model(&User{}).Where("id = ?", userID).Update("address", updateData.Address).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update address"})
		return
	}

	c.JSON(200, gin.H{"address": updateData.Address})
}

func ChangePassword(c *gin.Context) {
	userID := uint(1) // สมมติว่าได้จาก middleware
	var passwordData struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&passwordData); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	var user User
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

func SearchProducts(c *gin.Context) {
	description := c.Query("description")
	minPrice := c.Query("min_price")
	maxPrice := c.Query("max_price")

	query := DB.Model(&Product{})
	if description != "" {
		query = query.Where("description LIKE ?", "%"+description+"%")
	}
	if minPrice != "" {
		query = query.Where("price >= ?", minPrice)
	}
	if maxPrice != "" {
		query = query.Where("price <= ?", maxPrice)
	}

	var products []Product
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

	var cart Cart
	if err := DB.Where("name = ? AND customer_id = ?", cartName, userID).First(&cart).Error; err != nil {
		cart = Cart{Name: cartName, CustomerID: userID}
		if err := DB.Create(&cart).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to create cart"})
			return
		}
	}

	var cartItem CartItem
	if err := DB.Where("cart_id = ? AND product_id = ?", cart.ID, itemData.ProductID).First(&cartItem).Error; err == nil {
		cartItem.Quantity += itemData.Quantity
		if err := DB.Save(&cartItem).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to update cart item"})
			return
		}
	} else {
		cartItem = CartItem{CartID: cart.ID, ProductID: itemData.ProductID, Quantity: itemData.Quantity}
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

	var carts []Cart
	if err := DB.Preload("Items.Product").Where("customer_id = ?", customerID).Find(&carts).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to get carts"})
		return
	}

	c.JSON(200, carts)
}

func main() {
	ConnectDatabase()
	SeedData()

	r := gin.Default()
	r.POST("/auth/login", Login)
	r.GET("/users/me", GetUserProfile)
	r.PUT("/users/me/address", UpdateAddress)
	r.PUT("/users/me/password", ChangePassword)
	r.GET("/products", SearchProducts)
	r.POST("/carts/:cart_name/items", AddToCart)
	r.GET("/carts", GetCarts)

	r.Run(":8080")
}
