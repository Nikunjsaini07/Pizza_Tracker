package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"

	"pizza-tracker/internals/models"

	"github.com/gin-contrib/sessions"
	gormsessions "github.com/gin-contrib/sessions/gorm"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var notifier = NewNotificationManager()

var PizzaTypes = []string{"Margherita", "Pepperoni", "Vegetarian", "Hawaiian"}
var PizzaSizes = []string{"Small", "Medium", "Large"}

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("pizza.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	db.AutoMigrate(&models.Order{}, &models.OrderItem{}, &models.User{})

	var count int64
	db.Model(&models.User{}).Count(&count)
	if count == 0 {
		hashedPassword, _ := models.HashPassword("password123")
		adminUser := models.User{
			Username: "admin",
			Password: hashedPassword,
		}
		db.Create(&adminUser)
		slog.Info("Default admin user created!", "username", "admin", "password", "password123")
	}
}

type OrderRequest struct {
	Name   string   `form:"name" binding:"required"`
	Pizzas []string `form:"pizza" binding:"required"`
	Sizes  []string `form:"size" binding:"required"`
}

type LoginRequest struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userId := session.Get("userId")
		if userId == nil {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	initDB()

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	store := gormsessions.NewStore(db, true, []byte("pizza-session-secret-key"))
	router.Use(sessions.Sessions("pizza-tracker-session", store))

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "order.html", gin.H{
			"PizzaTypes": PizzaTypes,
			"PizzaSizes": PizzaSizes,
		})
	})

	router.POST("/order", func(c *gin.Context) {
		var form OrderRequest
		if err := c.ShouldBind(&form); err != nil {
			c.HTML(http.StatusOK, "order.html", gin.H{
				"PizzaTypes": PizzaTypes,
				"PizzaSizes": PizzaSizes,
				"Error":      "All fields are required!",
			})
			return
		}

		var orderItems []models.OrderItem
		for i := 0; i < len(form.Pizzas); i++ {
			orderItems = append(orderItems, models.OrderItem{
				Pizza: form.Pizzas[i],
				Size:  form.Sizes[i],
			})
		}

		order := models.Order{
			CustomerName: form.Name,
			Status:       "Order placed",
			Items:        orderItems,
		}

		if err := db.Create(&order).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "order.html", gin.H{
				"PizzaTypes": PizzaTypes,
				"PizzaSizes": PizzaSizes,
				"Error":      "Failed to save order: " + err.Error(),
			})
			return
		}

		slog.Info("Order saved inside SQLite", "order_id", order.ID)
		c.Redirect(http.StatusSeeOther, "/customer/"+order.ID)
	})

	router.GET("/customer/:id", func(c *gin.Context) {
		orderID := c.Param("id")

		var order models.Order
		if err := db.Preload("Items").First(&order, "id = ?", orderID).Error; err != nil {
			c.String(http.StatusNotFound, "Order not found!")
			return
		}

		c.HTML(http.StatusOK, "customer.html", gin.H{
			"OrderID":      order.ID,
			"CustomerName": order.CustomerName,
			"Items":        order.Items,
			"Status":       order.Status,
		})
	})

	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	router.POST("/login", func(c *gin.Context) {
		var form LoginRequest
		if err := c.ShouldBind(&form); err != nil {
			c.HTML(http.StatusOK, "login.html", gin.H{"Error": "Username and password required"})
			return
		}

		var user models.User
		if err := db.First(&user, "username = ?", form.Username).Error; err != nil {
			c.HTML(http.StatusOK, "login.html", gin.H{"Error": "Invalid username or password"})
			return
		}

		if !user.CheckPassword(form.Password) {
			c.HTML(http.StatusOK, "login.html", gin.H{"Error": "Invalid username or password"})
			return
		}

		session := sessions.Default(c)
		session.Set("userId", user.ID)
		session.Save()

		c.Redirect(http.StatusSeeOther, "/admin")
	})

	router.POST("/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		c.Redirect(http.StatusSeeOther, "/login")
	})

	admin := router.Group("/admin")
	admin.Use(AuthMiddleware())
	{
		admin.GET("", func(c *gin.Context) {
			var orders []models.Order
			if err := db.Preload("Items").Order("created_at desc").Find(&orders).Error; err != nil {
				c.String(http.StatusInternalServerError, "Failed to load orders")
				return
			}
			c.HTML(http.StatusOK, "admin.html", gin.H{"Orders": orders})
		})

		admin.POST("/order/:id/update", func(c *gin.Context) {
			orderID := c.Param("id")
			newStatus := c.PostForm("status")

			err := db.Model(&models.Order{}).Where("id = ?", orderID).Update("status", newStatus).Error
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to update status")
				return
			}

			slog.Info("Order status updated in SQLite", "order_id", orderID, "new_status", newStatus)
			notifier.Notify("order:"+orderID, "order_updated")
			c.Redirect(http.StatusSeeOther, "/admin")
		})

		admin.POST("/order/:id/delete", func(c *gin.Context) {
			orderID := c.Param("id")

			if err := db.Delete(&models.Order{}, "id = ?", orderID).Error; err != nil {
				c.String(http.StatusInternalServerError, "Failed to delete")
				return
			}

			c.Redirect(http.StatusSeeOther, "/admin")
		})
	}

	router.GET("/notifications", func(c *gin.Context) {
		orderID := c.Query("orderId")
		if orderID == "" {
			c.String(http.StatusBadRequest, "Missing orderId parameter")
			return
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		clientChan := make(chan string, 10)
		key := "order:" + orderID

		notifier.AddClient(key, clientChan)

		defer func() {
			notifier.RemoveClient(key, clientChan)
			slog.Info("Customer client disconnected", "order_id", orderID)
		}()

		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-clientChan; ok {
				c.SSEvent("message", msg)
				return true
			}
			return false
		})
	})

	router.Run(":8080")
}