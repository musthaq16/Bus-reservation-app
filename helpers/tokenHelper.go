package helpers

import (
	configs "busapp/database"
	models "busapp/models"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// SignedDetails
type SignedDetails struct {
	Email    string
	Username string

	Role string
	Uid  string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "user")

var SECRET_KEY string = os.Getenv("SECRET_KEY")

// GenerateAllTokens generates both teh detailed token and refresh token
func GenerateAllTokens(email string, username string, uid string, role string) (signedToken string, err error) {
	claims := &SignedDetails{
		Email:    email,
		Username: username,
		Role:     role,
		Uid:      uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return
	}

	if err != nil {
		log.Panic(err)
		return
	}

	return token, err
}

// ValidateToken validates the jwt token
func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)
	if err != nil {
		msg = err.Error()
		return
	}
	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = fmt.Sprintf("the token is invalid")
		msg = err.Error()
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprintf("token is expired")
		msg = err.Error()
		return
	}

	return claims, msg
}

// GenerateResetToken generates a reset token using UUID
func GenerateResetToken(email string) (string, error) {
	claims := &jwt.StandardClaims{
		Subject:   email,
		ExpiresAt: time.Now().Add(time.Hour * 1).Unix(), // Set the expiration time for the reset token (e.g., 1 hour)
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return "", err
	}

	return token, nil

}

// ValidateResetToken validates the reset token and returns the associated email
func ValidateResetToken(ctx context.Context, resetToken string) (string, error) {
	// Implement the logic to validate the reset token in the database

	// Query the user document with the reset token
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"reset_token": resetToken}).Decode(&user)
	if err != nil {
		return "", err
	}

	return *user.Email, nil
}

// StoreResetTokenByEmail stores the reset token in the database
func StoreResetTokenByEmail(ctx context.Context, email string, resetToken string) error {
	// Implement the logic to store the reset token in the database

	// Update the user document with the reset token
	update := bson.M{"$set": bson.M{"reset_token": resetToken}}
	_, err := userCollection.UpdateOne(ctx, bson.M{"email": email}, update)

	return err

}

// MarkResetTokenAsUsed marks the reset token as used in the database
func MarkResetTokenAsUsed(ctx context.Context, userID string) error {
	// Assuming you have a database connection or ORM
	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": bson.M{"reset_token_used": true}}

	_, err := userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}
