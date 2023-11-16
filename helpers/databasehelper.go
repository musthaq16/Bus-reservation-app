package helpers

import (
	models "busapp/models"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

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

// UpdateUserDetailsByUid updates user details in the database
func UpdateUserDetailsByUid(
	ctx context.Context,
	email, username, phoneNumber, password string,
	user_id interface{},
) error {
	// Fetch the existing user details
	existingUser, err := GetUserByUid(ctx, user_id)
	if err != nil {
		return err
	}

	// Set parameters to existing values if they are missing in the request
	if email == "" {
		email = *existingUser.Email
	}
	if username == "" {
		username = *existingUser.Username
	}
	if phoneNumber == "" {
		phoneNumber = *existingUser.Phone
	}
	if password == "" {
		password = *existingUser.Password
	}

	// Check if the new username already exists
	existingUserByUsername, err := GetUserByUsername(ctx, username)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if existingUserByUsername != nil && existingUserByUsername.User_id != user_id {
		return fmt.Errorf("Username %s already exists", username)
	}

	// Check if the new email already exists
	existingUserByEmail, err := GetUserByEmail(ctx, email)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if existingUserByEmail != nil && existingUserByEmail.User_id != user_id {
		return fmt.Errorf("Email %s already exists", email)
	}

	// Check if the new phone number already exists
	existingUserByPhone, err := GetUserByPhoneNumber(ctx, phoneNumber)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if existingUserByPhone != nil && existingUserByPhone.User_id != user_id {
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

func GetUserByUid(ctx context.Context, user_id interface{}) (*models.User, error) {
	var user models.User

	err := userCollection.FindOne(ctx, bson.M{"user_id": user_id}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		// User with the provided user_id not found
		return nil, nil
	} else if err != nil {
		// Other unexpected error
		return nil, err
	}

	return &user, nil
}
