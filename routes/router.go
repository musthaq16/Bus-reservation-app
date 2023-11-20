package routes

import (
	"github.com/gin-gonic/gin"

	controller "busapp/controllers"
)

// UserRoutes function
func Router(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/signup", controller.SignUp())
	incomingRoutes.POST("/login", controller.Login())
	incomingRoutes.POST("/Forgetpassword", controller.HandleForgetPassword) // by using token
	incomingRoutes.POST("ResetPassword", controller.HandleResetPassword)    // by using token
	incomingRoutes.POST("/forgetpassword", controller.ForgetPassword)       //by using otp
	incomingRoutes.POST("/resetpassword", controller.ResetPasswordWithOTP)  //by using otp
}

// UserRoutes function
func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.PATCH("/edituser", controller.UpdateUserDetailsHandler)
	incomingRoutes.GET("/me", controller.GetMyDetails)
	incomingRoutes.GET("helloall", controller.Hello)
}

// UserRoutes function
func AdminRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/admin/adduser", controller.Adduser)
	incomingRoutes.DELETE("/admin/deleteuser", controller.AdminDeleteUser)
	incomingRoutes.GET("/admin/getcustomers", controller.AdminGetAllCustomers)
	incomingRoutes.GET("/admin/getallusers", controller.AdminGetAllUsers)
	incomingRoutes.POST("/admin/addBus", controller.AddBus)
	// incomingRoutes.GET("helloall", controller.Hello)
}
