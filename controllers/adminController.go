package controllers

import (
	helper "busapp/helpers"
	"busapp/models"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddBus(c *gin.Context) {
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

	// Extract bus information from the request
	var addBusRequest models.Bus
	if err := c.BindJSON(&addBusRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Check if the required fields are present
	if addBusRequest.SeatsTotal < 0 || addBusRequest.SeatsTotal > 45 || addBusRequest.Date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Total seats and date are required"})
		return
	}

	createdAt := time.Now()
	updatedAt := time.Now()

	// Create a new user object
	newBus := models.Bus{
		ID:         primitive.NewObjectID(),
		Bus_id:     primitive.NewObjectID().Hex(),
		Date:       addBusRequest.Date,
		SeatsTotal: addBusRequest.SeatsTotal,
		Created_at: createdAt,
		Updated_at: updatedAt,
		// Add other fields as needed
	}

	// Insert the new user into the database
	_, err := busCollection.InsertOne(c, newBus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to add Bus: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bus added successfully"})

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
	if addUserRequest.Email == "" || addUserRequest.Password == "" || addUserRequest.Phone == "" || addUserRequest.Username == "" || addUserRequest.Role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email,Password,phoneNumber,username and role are required"})
		return
	}

	// Check if the email already exists
	existingUserMail, err := helper.GetUserByEmail(c, addUserRequest.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error checking user Mail existence: %v", err)})
		return
	}
	if existingUserMail != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with the provided email already exists"})
		return
	}
	existingUserName, err := helper.GetUserByUsername(c, addUserRequest.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error checking user Name existence: %v", err)})
		return
	}
	if existingUserName != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with the provided email already exists"})
		return
	}
	existingUserPhone, err := helper.GetUserByPhoneNumber(c, addUserRequest.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error checking user Phonenumber existence: %v", err)})
		return
	}
	if existingUserPhone != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with the provided Phonenumber already exists"})
		return
	}
	createdAt := time.Now()
	updatedAt := time.Now()

	// Hash the password before storing it in the database
	hashedPassword := helper.HashPassword(addUserRequest.Password)

	// Create a new user object
	newUser := models.User{
		ID:        primitive.NewObjectID(),
		Email:     addUserRequest.Email,
		Password:  hashedPassword,
		Username:  addUserRequest.Username,
		UserID:    primitive.NewObjectID().Hex(), // Generate a new UserID
		Phone:     addUserRequest.Phone,
		Role:      addUserRequest.Role,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		// Add other fields as needed
	}

	// Insert the new user into the database
	_, err = userCollection.InsertOne(c, newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to add user: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User added successfully"})

}

// DeleteUserHandler is the API endpoint to delete a user by user_id (admin only)
func AdminDeleteUser(c *gin.Context) {
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

	// Call the DeleteUserByUid function
	err := helper.DeleteUserByUid(c, user_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete user: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func AdminGetAllCustomers(c *gin.Context) {
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
	users, err := helper.GetAllCustomersFromDatabase(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving customers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})

}

func AdminGetAllUsers(c *gin.Context) {
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
	users, err := helper.GetAllUsersFromDatabase(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})

}
