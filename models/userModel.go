package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is the model that governs all notes objects retrived or inserted into the DB
type User struct {
	ID         primitive.ObjectID `bson:"_id"`
	Username   *string            `json:"username" validate:"required,min=4"`
	Password   *string            `json:"Password" validate:"required,min=8"`
	Email      *string            `json:"email" validate:"email,required"`
	Phone      *string            `json:"phone" validate:"required"`
	Role       *string            `json:"role" validate:"required"`
	Created_at time.Time          `json:"created_at"`
	Updated_at time.Time          `json:"updated_at"`
	User_id    string             `json:"user_id"`
}
