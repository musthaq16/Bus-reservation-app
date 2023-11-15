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
	r.Use(gin.Logger())

	routes.UserRouter(r)
	r.Use(middleware.Authentication())
	routes.UserRoutes(r)

	// r.POST("/login", controllers.Login())

	// r.POST("/signup", controller.SignUp())
	// r.POST("/login", controller.Login())

	r.Run(":5000")
}
