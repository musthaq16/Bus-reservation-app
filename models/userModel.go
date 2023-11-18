package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is the model that governs all notes objects retrived or inserted into the DB
type User struct {
	ID         primitive.ObjectID `bson:"_id"`
	Username   string             `json:"username" validate:min=4"`
	Password   string             `json:"password" validate:min=8"`
	Email      string             `json:"email" validate:"required,email"`
	Phone      string             `json:"phone,omitempty"`
	Role       string             `json:"role,omitempty"`
	Created_at time.Time          `json:"created_at"`
	Updated_at time.Time          `json:"updated_at"`
	User_id    string             `json:"user_id"`
	OTP        string             `json:"otp,omitempty"`
	OTPExpires time.Time          `json:"otpexpires,omitempty"`
}
type LimitedUserDetails struct {
	Username   string    `json:"username" bson:"username"`
	Email      string    `json:"email" bson:"email"`
	Phone      string    `json:"phone" bson:"phone"`
	Created_at time.Time `json:"created_at" bson:"created_at"`
}

// ResetPasswordRequest represents the request payload for resetting a password
// type ResetPasswordRequest struct {
// 	Email       string `json:"email" binding:"omitempty,email"`
// 	ResetToken  string `json:"resetToken" binding:"omitempty"`
// 	NewPassword string `json:"newPassword" binding:"required,min=8"`
// }

// ForgetPasswordRequest is the structure for forget password API request
// type ForgetPasswordRequest struct {
// 	Email string `json:"email" binding:"required,email"`
// }

// ResetPasswordWithOTPRequest represents the request structure for resetting password with OTP
type ResetPasswordWithOTPRequest struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}
