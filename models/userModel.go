package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is the model that governs all notes objects retrived or inserted into the DB
type User struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username   string             `json:"username" bson:"username" validate:"min=4"`
	Password   string             `json:"password,omitempty" bson:"password,omitempty" validate:"min=8"`
	Email      string             `json:"email" bson:"email" validate:"required,email"`
	Phone      string             `json:"phone,omitempty" bson:"phone,omitempty"`
	Role       string             `json:"role,omitempty" bson:"role,omitempty"`
	CreatedAt  time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt  time.Time          `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	UserID     string             `json:"user_id,omitempty" bson:"user_id,omitempty"`
	OTP        string             `json:"otp,omitempty" bson:"otp,omitempty"`
	OTPExpires time.Time          `json:"otp_expires,omitempty" bson:"otp_expires,omitempty"`
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

// // ResetPasswordWithOTPRequest represents the request structure for resetting password with OTP
// type ResetPasswordWithOTPRequest struct {
// 	Email       string `json:"email" binding:"required,email"`
// 	OTP         string `json:"otp" binding:"required"`
// 	NewPassword string `json:"newPassword" binding:"required,min=8"`
// }
