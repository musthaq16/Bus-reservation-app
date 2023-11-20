package main

import (
	configs "busapp/database"
	middleware "busapp/middleware"
	"busapp/routes"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("hello worldd")
	configs.ConnectDB()

	r := gin.Default()
	r.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "hello from get method",
		})
	})
	// r.Use(gin.Logger())

	routes.Router(r)
	r.Use(middleware.Authentication())
	routes.UserRoutes(r)
	routes.AdminRoutes(r)

	r.Run(":5000")
}
