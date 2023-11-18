package controllers

import (
	configs "busapp/database"
	helper "busapp/helpers"
	"context"
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
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
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

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email is already exists"})
			return
		}

		password := helper.HashPassword(*user.Password)
		user.Password = &password

		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the phone number"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this phone number is already exists"})
			return
		}

		count, err = userCollection.CountDocuments(ctx, bson.M{"username": user.Username})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the username"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this username is already exists"})
			return
		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		InsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, gin.H{"insertionID": InsertionNumber})

	}
}

// Login is the api used to tget a single user
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Email or password is incorrect"})
			return
		}

		passwordIsValid, msg := helper.VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		token, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.Username, foundUser.User_id, *foundUser.Role)
		fmt.Println(token)
		c.JSON(http.StatusOK, gin.H{"token": token, "foundUser": foundUser})
	}
}

// HandleForgetPassword is the API endpoint for initiating the forgot password flow
func ForgetPassword(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	// Extract email from the request body
	var request models.ForgetPasswordRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		cancel()
		return
	}

	// Check if the email exists in the database
	user, err := helper.GetUserByEmail(ctx, request.Email)
	defer cancel()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user existence"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User with the provided email not found"})
		return
	}

	// Generate an OTP
	otp := helper.GenerateOTP()

	// Store the OTP and its expiration time in the database
	err = helper.StoreOTPByEmail(ctx, request.Email, otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error storing OTP"})
		return
	}

	// TODO: Send an email to the user with the OTP
	// Send the OTP to the user's email
	err = helper.SendOTPEmail(request.Email, otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP email"})
		return
	}
	// You can use a third-party library or service for sending emails here

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})
}

// HandleResetPasswordWithOTP is the API endpoint for resetting the password using the OTP
func ResetPasswordWithOTP(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	// Extract email and OTP from the request body
	var request models.ResetPasswordWithOTPRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		cancel()
		return
	}

	// Validate the OTP against the database
	isValid, err := helper.ValidateOTPByEmail(ctx, request.Email, request.OTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error validating OTP"})
		cancel()
		return
	}

	if !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
		cancel()
		return
	}

	// // Extract new password from the request body
	// var resetRequest models.ResetPasswordRequest
	// if err := c.BindJSON(&resetRequest); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
	// 	cancel()
	// 	return
	// }

	// Update the user's password in the database
	err = helper.UpdateUserPassword(ctx, request.Email, helper.HashPassword(request.NewPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		cancel()
		return
	}

	// Invalidate the OTP (optional)

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
	cancel()
}

// HandleForgetPassword is the API endpoint for initiating the forgot password flow
func HandleForgetPassword(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	// Extract email from the request body
	var request models.ForgetPasswordRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Check if the email exists in the database
	user, err := helper.GetUserByEmail(ctx, request.Email)
	defer cancel()
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
	err = helper.StoreResetTokenByEmail(ctx, request.Email, resetToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error storing reset token"})
		return
	}

	// TODO: Send an email to the user with a link containing the reset token
	// Send the password reset link to the user's email
	err = helper.SendResetLinkEmail(request.Email, resetToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset link email"})
		cancel()
		return
	}
	// You can use a third-party library or service for sending emails here

	c.JSON(http.StatusOK, gin.H{"message": "Reset token generated and email sent successfully"})
}

// HandleResetPassword is the API endpoint for resetting the user's password
func HandleResetPassword(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	// Extract reset token from the query parameters
	resetToken := c.Query("token")
	if resetToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reset token is required"})
		cancel()
		return
	}

	// Validate the reset token and retrieve the user
	user, err := helper.GetUserByResetToken(ctx, resetToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reset token"})
		cancel()
		return
	}
	fmt.Println(user)

	// Extract new password from the request body
	var request models.ResetPasswordRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		cancel()
		return
	}

	// Update the user's password in the database
	err = helper.UpdateUserPassword(ctx, *user.Email, helper.HashPassword(request.NewPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		cancel()
		return
	}

	// Mark the reset token as used in the database
	err = helper.MarkResetTokenAsUsed(ctx, user.User_id)
	if err != nil {
		// Handle the error, e.g., log it
		fmt.Println("eeror while resetting token", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
	cancel()
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

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call the UpdateUserDetailsByUid function
	err := helper.UpdateUserDetailsByUid(
		ctx,
		updateUserDetailsRequest,
		updateUserDetailsRequest.User_id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update user details: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User details updated successfully"})
}

func Adduser(c *gin.Context) {
	// Extract admin information from the token or any other identifier
	roleFromToken, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Role information not found"})
		return
	}

	// Check if the user making the request is an admin
	isAdmin, _ := roleFromToken.(string)
	if isAdmin != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unauthorized access"})
		return
	}

	// Extract user information from the request
	var addUserRequest models.User
	if err := c.BindJSON(&addUserRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Check if the required fields are present
	if addUserRequest.Email == nil || addUserRequest.Password == nil || addUserRequest.Phone == nil || addUserRequest.Username == nil || addUserRequest.Role == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email,Password,phoneNumber,username and role are required"})
		return
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if the email already exists
	existingUserMail, err := helper.GetUserByEmail(ctx, *addUserRequest.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error checking user Mail existence: %v", err)})
		return
	}
	if existingUserMail != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with the provided email already exists"})
		return
	}
	existingUserName, err := helper.GetUserByUsername(ctx, *addUserRequest.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error checking user Name existence: %v", err)})
		return
	}
	if existingUserName != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with the provided email already exists"})
		return
	}
	existingUserPhone, err := helper.GetUserByPhoneNumber(ctx, *addUserRequest.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error checking user Phonenumber existence: %v", err)})
		return
	}
	if existingUserPhone != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with the provided Phonenumber already exists"})
		return
	}
	createdAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	// Hash the password before storing it in the database
	hashedPassword := helper.HashPassword(*addUserRequest.Password)

	// Create a new user object
	newUser := models.User{
		ID:         primitive.NewObjectID(),
		Email:      addUserRequest.Email,
		Password:   &hashedPassword,
		Username:   addUserRequest.Username,
		User_id:    primitive.NewObjectID().Hex(), // Generate a new UserID
		Phone:      addUserRequest.Phone,
		Role:       addUserRequest.Role,
		Created_at: createdAt,
		Updated_at: updatedAt,
		// Add other fields as needed
	}

	// Insert the new user into the database
	_, err = userCollection.InsertOne(ctx, newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to add user: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User added successfully"})

}

// DeleteUserHandler is the API endpoint to delete a user by user_id (admin only)
func DeleteUserHandler(c *gin.Context) {
	// Extract admin information from the token or any other identifier
	roleFromToken, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Role information not found"})
		return
	}

	// Check if the user making the request is an admin
	isAdmin, _ := roleFromToken.(string)
	if isAdmin != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unauthorized access"})
		return
	}

	// Get the user_id to be deleted from the request
	user_id := c.Query("user_id")
	if user_id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id parameter is required"})
		return
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call the DeleteUserByUid function
	err := helper.DeleteUserByUid(ctx, user_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete user: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func Hello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
