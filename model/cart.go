package model

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
