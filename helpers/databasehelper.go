package helpers

import (
	configs "busapp/database"
	models "busapp/models"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
func UpdateUserDetailsByUid(ctx context.Context, updateUserDetailsRequest models.User, user_id interface{}) error {
	// Fetch the existing user details
	existingUser, err := GetUserByUid(ctx, user_id)
	if err != nil {
		return err
	}

	// Set parameters to existing values if they are missing in the request
	if updateUserDetailsRequest.Email == "" {
		updateUserDetailsRequest.Email = existingUser.Email
	}
	if updateUserDetailsRequest.Username == "" {
		updateUserDetailsRequest.Username = existingUser.Username
	}
	if updateUserDetailsRequest.Phone == "" {
		updateUserDetailsRequest.Phone = existingUser.Phone
	}
	if updateUserDetailsRequest.Password == "" {
		updateUserDetailsRequest.Password = existingUser.Password
	}

	// Check if the new username already exists
	existingUserByUsername, err := GetUserByUsername(ctx, updateUserDetailsRequest.Username)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if existingUserByUsername != nil && existingUserByUsername.User_id != user_id {
		return fmt.Errorf("Username %s already exists", updateUserDetailsRequest.Username)
	}

	// Check if the new email already exists
	existingUserByEmail, err := GetUserByEmail(ctx, updateUserDetailsRequest.Email)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if existingUserByEmail != nil && existingUserByEmail.User_id != user_id {
		return fmt.Errorf("Email %s already exists", updateUserDetailsRequest.Email)
	}

	// Check if the new phone number already exists
	existingUserByPhone, err := GetUserByPhoneNumber(ctx, updateUserDetailsRequest.Phone)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if existingUserByPhone != nil && existingUserByPhone.User_id != user_id {
		return fmt.Errorf("Phone number %s already exists", updateUserDetailsRequest.Phone)
	}

	// Hash the new password before updating it in the database
	hashedPassword := HashPassword(updateUserDetailsRequest.Password)

	updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	// Define the update query to set the new hashed password, new email, new username, and new phone number
	update := bson.M{
		"$set": bson.M{
			"username":   updateUserDetailsRequest.Username,
			"email":      updateUserDetailsRequest.Email,
			"phone":      updateUserDetailsRequest.Phone,
			"password":   hashedPassword,
			"updated_at": updated_at,
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

// DeleteUserByUid deletes a user by user_id
func DeleteUserByUid(ctx context.Context, user_id interface{}) error {
	_, err := userCollection.DeleteOne(ctx, bson.M{"user_id": user_id})
	return err
}

// GetUserByUid retrieves a user based on the user_id
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

// UpdateUserPasswordByEmail updates the user's password in the database
func UpdateUserPasswordByEmail(ctx context.Context, email string, newPassword string) error {
	// Your MongoDB collection for users
	userCollection := configs.GetCollection(configs.DB, "user")

	// Define the filter to find the user by email
	filter := bson.M{"email": email}

	// Define the update to set the new password and update the timestamp
	update := bson.M{
		"$set": bson.M{
			"password":    HashPassword(newPassword),
			"updated_at":  time.Now(),
			"reset_token": nil, // Optionally, clear the reset token after successful password reset
		},
	}

	// Set up options for the update
	options := options.Update().SetUpsert(false)

	// Perform the update operation
	result, err := userCollection.UpdateOne(ctx, filter, update, options)
	if err != nil {
		return fmt.Errorf("failed to update user password: %v", err)
	}

	// Check if any documents were modified
	if result.ModifiedCount == 0 {
		// If no documents were modified, it means there is no user with the provided email
		return fmt.Errorf("no user found with the email: %s", email)
	}

	return nil
}

// GetUserByResetToken retrieves a user by their reset token from the database
func GetUserByResetToken(ctx context.Context, resetToken string) (models.User, error) {
	var user models.User
	filter := bson.D{{Key: "reset_token", Value: resetToken}}
	err := userCollection.FindOne(ctx, filter).Decode(&user)

	// Replace the above code with the appropriate logic for your database

	return user, err
}

// getAllCustomersFromDatabase retrieves all user details from the database
func GetAllCustomersFromDatabase(ctx context.Context) ([]models.LimitedUserDetails, error) {
	// Assuming you have a MongoDB collection named "users" and a model for the User
	var users []models.LimitedUserDetails
	// Specify the fields you want to retrieve
	projection := bson.M{"username": 1, "email": 1, "phone": 1, "created_at": 1}
	cursor, err := userCollection.Find(ctx, bson.M{"role": "customer"}, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// getAllUsersFromDatabase retrieves all user details from the database
func GetAllUsersFromDatabase(ctx context.Context) ([]models.LimitedUserDetails, error) {
	// Assuming you have a MongoDB collection named "users" and a model for the User
	var users []models.LimitedUserDetails
	// Specify the fields you want to retrieve
	projection := bson.M{"username": 1, "email": 1, "phone": 1, "created_at": 1}
	cursor, err := userCollection.Find(ctx, bson.M{}, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}
