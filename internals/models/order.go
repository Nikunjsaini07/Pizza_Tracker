package models

var (
	OrderStatus = []string{"Order placed" , "Preparing" , "Baking", "Quality Check" , "Ready"}

    PizzaTypes = []string{
		"Margherita",
		 "Pepperoni",
		  "Cheese", 
		  "Veggie", 
		  "BBQ Chicken", 
		  "Hawaiian", 
		  "Meat Lovers", 
		  "Supreme",
		  "Buffalo Chicken",
		  "Mushroom",
	}
	PizzaSizes = []string{
		"Small" , "Medium" , "Large", "X-large" , 
	}
)

type OrderModel struct {
	DB *gorm.DB    
	// pointer to gorn db
}

type Order struct {
	ID           string      `gorm:"primaryKey;size:14" json:"id"`
	Status       string      `gorm:"not null" json:"status"`
	CustomerName string      `gorm:"not null" json:"customerName"`
	Phone        string      `gorm:"not null" json:"phone"`
	Address      string      `gorm:"not null" json:"address"`
	Items        []OrderItem `gorm:"foreignKey:OrderID" json:"pizzas"`
	CreatedAt    time.Time   `json:"createdAt"`
}
