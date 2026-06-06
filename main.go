package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"pizza-tracker/internals/models"

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
	db.AutoMigrate(&models.Order{}, &models.OrderItem{})
}

type OrderRequest struct {
	Name  string `form:"name" binding:"required"`
	Pizza string `form:"pizza" binding:"required"`
	Size  string `form:"size" binding:"required"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	initDB()

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

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

		order := models.Order{
			CustomerName: form.Name,
			Status:       "Order placed",
			Items: []models.OrderItem{
				{
					Pizza: form.Pizza,
					Size:  form.Size,
				},
			},
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

		var pizzaType, pizzaSize string
		if len(order.Items) > 0 {
			pizzaType = order.Items[0].Pizza
			pizzaSize = order.Items[0].Size
		}

		c.HTML(http.StatusOK, "customer.html", gin.H{
			"OrderID":      order.ID,
			"CustomerName": order.CustomerName,
			"PizzaType":    pizzaType,
			"PizzaSize":    pizzaSize,
			"Status":       order.Status,
		})
	})

	router.GET("/admin", func(c *gin.Context) {
		var orders []models.Order
		if err := db.Preload("Items").Order("created_at desc").Find(&orders).Error; err != nil {
			c.String(http.StatusInternalServerError, "Failed to load orders: "+err.Error())
			return
		}

		c.HTML(http.StatusOK, "admin.html", gin.H{
			"Orders": orders,
		})
	})

	router.POST("/admin/order/:id/update", func(c *gin.Context) {
		orderID := c.Param("id")
		newStatus := c.PostForm("status")

		err := db.Model(&models.Order{}).Where("id = ?", orderID).Update("status", newStatus).Error
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to update status: "+err.Error())
			return
		}

		slog.Info("Order status updated in SQLite", "order_id", orderID, "new_status", newStatus)
		notifier.Notify("order:"+orderID, "order_updated")
		c.Redirect(http.StatusSeeOther, "/admin")
	})

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
