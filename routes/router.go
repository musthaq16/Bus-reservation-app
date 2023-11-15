package routes

import (
	"github.com/gin-gonic/gin"

	controller "busapp/controllers"
)

// UserRoutes function
func UserRouter(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/signup", controller.SignUp())
	incomingRoutes.POST("/login", controller.Login())
}

// UserRoutes function
func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.PATCH("/edituser", controller.UpdateUserDetailsHandler)
	// incomingRoutes.GET("helloall", controller.Hello)
}
