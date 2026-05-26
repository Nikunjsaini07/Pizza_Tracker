package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
)
// OrderRequest represents the form data submitted by the browser.
type OrderRequest struct {
	Name  string `form:"name"  binding:"required,min=2"`
	Pizza string `form:"pizza" binding:"required"`
	Size  string `form:"size" binding:"required"`
}
// We define our options as global slices (arrays) in Go
var PizzaTypes = []string{"Margherita", "Pepperoni", "Vegetarian", "Hawaiian"}
var PizzaSizes = []string{"Small", "Medium", "Large"}

func main() {
	router := gin.Default()

	// 1. Tell Gin to load all template files inside the templates folder
	router.LoadHTMLGlob("templates/*")

	// 2. Serve the HTML template on "/"
	router.GET("/", func(c *gin.Context) {
		// c.HTML renders the "order.tmpl" template.
		// We pass our Go data using gin.H (a key-value map).
		c.HTML(http.StatusOK, "order.html", gin.H{
			"PizzaTypes": PizzaTypes,
			"PizzaSizes": PizzaSizes,
		})
	})
		// 3. POST /order - Receives the form submission
	router.POST("/order", func(c *gin.Context) {
		var form OrderRequest

		// c.ShouldBind parses the form data and maps it into our 'form' struct variable
		if err := c.ShouldBind(&form); err != nil {
			// If validation fails (e.g., a field is missing), render the form again with an error
			c.HTML(http.StatusOK, "order.html", gin.H{
				"PizzaTypes": PizzaTypes,
				"PizzaSizes": PizzaSizes,
				"Error":      "Invalid submission! All fields are required.",
			})
			return
		}

		// If successful, render the form and show a success message!
		c.HTML(http.StatusOK, "order.html", gin.H{
			"PizzaTypes": PizzaTypes,
			"PizzaSizes": PizzaSizes,
			"Success":    "Thank you, " + form.Name + "! Your " + form.Size + " " + form.Pizza + " pizza has been ordered.",
		})
	})

	router.Run(":8080")
}