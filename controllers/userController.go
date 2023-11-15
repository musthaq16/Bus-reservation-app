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
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "user")
var validate = validator.New()

// HashPassword is used to encrypt the password before it is stored in the DB
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

// VerifyPassword checks the input password while verifying it with the passward in the DB.
func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("login or passowrd is incorrect")
		check = false
	}

	return check, msg
}

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

		password := HashPassword(*user.Password)
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

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
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

	// Call the UpdateUserDetailsByEmail function
	err := UpdateUserDetailsByUid(
		ctx,
		*updateUserDetailsRequest.Email,
		*updateUserDetailsRequest.Username,
		*updateUserDetailsRequest.Phone,
		*updateUserDetailsRequest.Password,
		userIdFromToken,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update user details: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User details updated successfully"})
}

// UpdateUserDetailsByUid updates user details in the database
func UpdateUserDetailsByUid(ctx context.Context, email, username, phoneNumber, password string, user_id interface{}) error {
	// Check if the new username already exists
	existingUser, err := GetUserByUsername(ctx, username)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if existingUser != nil && existingUser.User_id != user_id {
		return fmt.Errorf("Username %s already exists", username)
	}

	// Check if the new email already exists
	existingUser, err = GetUserByEmail(ctx, email)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if existingUser != nil && existingUser.User_id != user_id {
		return fmt.Errorf("Email %s already exists", email)
	}

	// Check if the new phone number already exists
	existingUser, err = GetUserByPhoneNumber(ctx, phoneNumber)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if existingUser != nil && existingUser.User_id != user_id {
		return fmt.Errorf("Phone number %s already exists", phoneNumber)
	}

	// Hash the new password before updating it in the database
	hashedPassword := HashPassword(password)

	// Define the update query to set the new hashed password, new email, new username, and new phone number
	update := bson.M{
		"$set": bson.M{
			"username": username,
			"email":    email,
			"phone":    phoneNumber,
			"password": hashedPassword,
		},
	}

	// Execute the update query using the provided context
	result, err := userCollection.UpdateOne(ctx, bson.M{"user_id": user_id}, update)
	if err != nil {
		return err
	}

	// Check if any documents were modified
	if result.ModifiedCount == 0 {
		// If no documents were modified, it means there is no user with the provided user_id
		return fmt.Errorf("No user found with the user_id: %s", user_id)
	}

	return nil
}

func Hello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil // User not found
	}
	if err != nil {
		return nil, err // Other errors
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil // User not found
	}
	if err != nil {
		return nil, err // Other errors
	}
	return &user, nil
}

// GetUserByPhoneNumber retrieves a user by phone number
func GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error) {
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"phone": phoneNumber}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil // User not found
	}
	if err != nil {
		return nil, err // Other errors
	}
	return &user, nil
}
