package helpers

import (
	models "busapp/models"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/smtp"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

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

// SendResetLinkEmail sends a password reset link to the user's email
func SendResetLinkEmail(email, resetToken string) error {
	// Implement email sending logic here
	// Example: using a hypothetical email server (replace with your actual email sending code)
	body := fmt.Sprintf("Click the following link to reset your password: http://yourapp.com/reset-password?token=%s", resetToken)
	message := []byte("Subject: Password Reset\n\n" + body)

	err := smtp.SendMail("smtp.gmail.com:587", smtp.PlainAuth("", "musthaqmohamed123456@gmail.com", "fsjflbpzbeuiwskl", "smtp.gmail.com"), "musthaqmohamed123456@gmail.com", []string{email}, message)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUserPassword updates the user's password in the database
func UpdateUserPassword(ctx context.Context, email string, hashedPassword string) error {
	// Assuming you have a database connection or ORM
	// Update the user's password based on the user ID

	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{"Password": hashedPassword}}
	fmt.Println("New Hashed Password:", hashedPassword)

	result, err := userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		fmt.Println("Error updating password:", err)
		return err
	}

	fmt.Printf("Matched %v document(s) and modified %v document(s)\n", result.MatchedCount, result.ModifiedCount)

	return nil
}

// GenerateOTP generates a random six-digit OTP
func GenerateOTP() string {
	randomNumber := rand.Intn(1000000)
	return strconv.Itoa(randomNumber)
}

// StoreOTPByEmail stores the OTP and its expiration time in the database
func StoreOTPByEmail(ctx context.Context, email string, otp string) error {
	// Assuming you have a database connection or ORM
	// Store the OTP and its expiration time based on the user's email

	expirationTime := time.Now().Add(5 * time.Minute) // OTP validity time

	filter := bson.M{"email": email}
	update := bson.M{
		"$set": bson.M{
			"otp":        otp,
			"otpexpires": expirationTime,
		},
	}

	_, err := userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

// Send email using SendGrid
func SendOTPEmail(email string, otp string) error {
	// Implement email sending logic here
	// Example: using a hypothetical email server (replace with your actual email sending code)
	body := fmt.Sprintf("Copy the following otp to reset your password: %s", otp)
	message := []byte("Subject: Password Reset\n\n" + body)

	err := smtp.SendMail("smtp.gmail.com:587", smtp.PlainAuth("", "musthaqmohamed123456@gmail.com", "fsjflbpzbeuiwskl", "smtp.gmail.com"), "musthaqmohamed123456@gmail.com", []string{email}, message)
	if err != nil {
		return err
	}

	return nil
}

// ValidateOTPByEmail validates the provided OTP against the stored OTP in the database
func ValidateOTPByEmail(ctx context.Context, email string, userOTP string) (bool, error) {
	// Assuming you have a database connection or ORM
	// Retrieve the stored OTP and its expiration time based on the user's email

	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return false, err
	}

	// Check if the OTP matches and is still valid
	if user.OTP == userOTP && time.Now().Before(user.OTPExpires) {
		return true, nil
	}
	fmt.Println("Stored OTP:", user.OTP)
	fmt.Println("Provided OTP:", userOTP)
	fmt.Println("OTP Expiration:", user.OTPExpires)

	return false, nil
}

// Invalidate OTP after a successful password reset
func InvalidateOTP(ctx context.Context, email string) error {
	// Update user record to clear OTP and expiration time
	return ClearOTPByEmail(ctx, email)
}

// ClearOTPByEmail clears the OTP for the user with the given email
func ClearOTPByEmail(ctx context.Context, email string) error {
	// Assuming you have a database connection or ORM
	// Update the user's record to clear OTP and expiration time

	// Define the update operation
	update := bson.M{
		"$set": bson.M{
			"otp":        "",
			"otpExpires": time.Time{}, // Set to zero time to clear the expiration
		},
	}

	// Find the user by email and perform the update
	_, err := userCollection.UpdateOne(ctx, bson.M{"email": email}, update)
	if err != nil {
		return err
	}

	return nil
}
