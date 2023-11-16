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

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "2"})
			return
		}

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
		// token, err := helper.GenerateAllTokens(*user.Email, *user.Username, user.User_id, *user.Role)
		// user.Token = &token
		// if err != nil {
		// 	log.Panic(err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while generating token"})
		// 	return
		// }

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
		fmt.Println("1111111111111111111111111111111111111111111111111", token)
		c.JSON(http.StatusOK, gin.H{"token": token, "foundUser": foundUser})
	}
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot update details for a different user"})
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
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access"})
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
		c.JSON(http.StatusConflict, gin.H{"error": "User with the provided email already exists"})
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
		c.JSON(http.StatusConflict, gin.H{"error": "User with the provided Phonenumber already exists"})
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

func Hello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
