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

// UserRoutes function
func AdminRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/admin/adduser", controller.Adduser)
	// incomingRoutes.GET("helloall", controller.Hello)
}
