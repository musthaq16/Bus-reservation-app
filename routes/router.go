package routes

import (
	"github.com/gin-gonic/gin"

	controller "busapp/controllers"
)

// UserRoutes function
func UserRouter(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/signup", controller.SignUp())
	incomingRoutes.POST("/login", controller.Login())
	incomingRoutes.POST("/Forgetpassword", controller.HandleForgetPassword)
	incomingRoutes.POST("ResetPassword", controller.HandleResetPassword)
	incomingRoutes.POST("/forgetpassword", controller.ForgetPassword)
	incomingRoutes.POST("/resetpassword", controller.ResetPasswordWithOTP)

}

// UserRoutes function
func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.PATCH("/edituser", controller.UpdateUserDetailsHandler)
	incomingRoutes.GET("/me", controller.Me)
	incomingRoutes.GET("helloall", controller.Hello)
}

// UserRoutes function
func AdminRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/admin/adduser", controller.Adduser)
	incomingRoutes.DELETE("/admin/deleteuser", controller.AdminDeleteUser)
	incomingRoutes.GET("/admin/getcustomers", controller.AdminGetAllCustomers)
	incomingRoutes.GET("/admin/getallusers", controller.AdminGetAllUsers)
	// incomingRoutes.GET("helloall", controller.Hello)
}
