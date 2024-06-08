package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.Static("/Static", "./Static")

	// Serve the home page
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})

	// Handle form submission or button click
	router.POST("/submit", func(c *gin.Context) {
		// Process the form data here

		// Redirect to another page
		c.Redirect(http.StatusFound, "/success")
	})

	// Serve the success page
	router.GET("/success", func(c *gin.Context) {
		c.HTML(http.StatusOK, "service.html", nil)
	})

	router.LoadHTMLGlob("Static/*")
	// Run the server
	router.Run(":8080")
}
