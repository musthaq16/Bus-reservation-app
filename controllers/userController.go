package controllers

import (
	configs "busapp/database"
	helper "busapp/helpers"
	"fmt"
	"log"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"busapp/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "user")
var validate = validator.New()

// CreateUser is the api used to tget a single user
func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "1"})
			return
		}

		// validationErr := validate.Struct(user)
		// if validationErr != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "2"})
		// 	return
		// }

		count, err := userCollection.CountDocuments(c, bson.M{"email": user.Email})

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email is already exists"})
			return
		}

		password := helper.HashPassword(user.Password)
		user.Password = password

		count, err = userCollection.CountDocuments(c, bson.M{"phone": user.Phone})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the phone number"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this phone number is already exists"})
			return
		}

		count, err = userCollection.CountDocuments(c, bson.M{"username": user.Username})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the username"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this username is already exists"})
			return
		}

		user.Created_at = time.Now()
		user.Updated_at = time.Now()
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		InsertionNumber, insertErr := userCollection.InsertOne(c, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{"insertionID": InsertionNumber})

	}
}

// Login is the api used to tget a single user
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(c, bson.M{"email": user.Email}).Decode(&foundUser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Email or password is incorrect"})
			return
		}

		passwordIsValid, msg := helper.VerifyPassword(user.Password, foundUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		token, _ := helper.GenerateAllTokens(foundUser.Email, foundUser.Username, foundUser.User_id, foundUser.Role)
		fmt.Println(token)
		c.JSON(http.StatusOK, gin.H{"token": token, "foundUser": foundUser})
	}
}

// HandleForgetPassword is the API endpoint for initiating the forgot password flow
func ForgetPassword(c *gin.Context) {

	// Extract email from the request body
	var request models.User
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Check if the email exists in the database
	user, err := helper.GetUserByEmail(c, request.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user existence"})
		return
	}
	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with the provided email not found"})
		return
	}

	// Generate an OTP
	otp := helper.GenerateOTP()

	// Store the OTP and its expiration time in the database
	err = helper.StoreOTPByEmail(c, request.Email, otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error storing OTP"})
		return
	}

	// TODO: Send an email to the user with the OTP
	err = helper.SendOTPEmail(request.Email, otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP email"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})
}

// HandleResetPasswordWithOTP is the API endpoint for resetting the password using the OTP
func ResetPasswordWithOTP(c *gin.Context) {

	// Extract email and OTP from the request body
	var request models.User
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})

		return
	}

	// Validate the OTP against the database
	isValid, err := helper.ValidateOTPByEmail(c, request.Email, request.OTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error validating OTP"})
		return
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP"})
		return
	}

	// Update the user's password in the database
	err = helper.UpdateUserPassword(c, request.Email, helper.HashPassword(request.Password))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Invalidate the OTP (optional)
	err = helper.InvalidateOTP(c, request.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear otp and expites time"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// HandleForgetPassword is the API endpoint for initiating the forgot password flow
func HandleForgetPassword(c *gin.Context) {

	// Extract email from the request body
	var request models.User
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Check if the email exists in the database
	user, err := helper.GetUserByEmail(c, request.Email)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user existence"})
		return
	}
	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with the provided email not found"})
		return
	}

	// Generate a reset token
	resetToken, err := helper.GenerateResetToken(request.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error wile resetting the token"})
	}

	// Store the reset token in the database
	err = helper.StoreResetTokenByEmail(c, request.Email, resetToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error storing reset token"})
		return
	}

	// TODO: Send an email to the user with a link containing the reset token
	// Send the password reset link to the user's email
	err = helper.SendResetLinkEmail(request.Email, resetToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset link email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reset token generated and email sent successfully"})
}

// HandleResetPassword is the API endpoint for resetting the user's password
func HandleResetPassword(c *gin.Context) {

	// Extract reset token from the query parameters
	resetToken := c.Query("token")
	if resetToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reset token is required"})

		return
	}

	// Validate the reset token and retrieve the user
	user, err := helper.GetUserByResetToken(c, resetToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reset token"})
		return
	}
	fmt.Println(user)

	// Extract new password from the request body
	var request models.User
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Update the user's password in the database
	err = helper.UpdateUserPassword(c, user.Email, helper.HashPassword(request.Password))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Mark the reset token as used in the database
	err = helper.MarkResetTokenAsUsed(c, user.User_id)
	if err != nil {
		fmt.Println("eeror while resetting token", err)
	}
	user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// UpdateUserDetailsHandler is the API endpoint to update user details
func UpdateUserDetailsHandler(c *gin.Context) {
	// Get username from the token or any other identifier
	userIdFromToken, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Userid not found in the token"})
		return
	}

	// Extract user information from the request
	var updateUserDetailsRequest models.User
	if err := c.BindJSON(&updateUserDetailsRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Check if the updating username matches the username from the token
	if updateUserDetailsRequest.User_id != userIdFromToken.(string) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot update details for a different user"})
		return
	}

	// Call the UpdateUserDetailsByUid function
	err := helper.UpdateUserDetailsByUid(
		c,
		updateUserDetailsRequest,
		updateUserDetailsRequest.User_id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update user details: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User details updated successfully"})
}

func GetMyDetails(c *gin.Context) {

	// Get username from the token or any other identifier
	userIdFromToken, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Userid not found in the token"})
		return
	}

	// Retrieve user details from the database using the user ID
	user, err := helper.GetUserByUid(c, userIdFromToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user details"})
		return
	}

	// Return the user details in the response
	c.JSON(http.StatusOK, gin.H{
		"username":   user.Username,
		"email":      user.Email,
		"phone":      user.Phone,
		"role":       user.Role,
		"updated_at": user.Updated_at,
	})
}

func Hello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
