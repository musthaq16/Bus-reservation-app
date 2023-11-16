package middleware

import (
	helper "busapp/helpers"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Authentication validates token and authorizes users
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		// // Get the authorization header
		// authHeader := c.GetHeader("Authorization")
		// if authHeader == "" {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
		// 	c.Abort()
		// 	return
		// }
		// // Check if the header has the "Bearer" prefix
		// authParts := strings.Fields(authHeader)
		// if len(authParts) != 2 || authParts[0] != "Bearer" {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
		// 	c.Abort()
		// 	return
		// }

		// // Extract the token from the header
		// tokenString := authParts[1]
		// fmt.Println("Received Token:", tokenString)

		// // Parse the JWT token
		// token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 	return []byte(helper.SECRET_KEY), nil
		// })

		// if err != nil {
		// 	fmt.Println("Error parsing token:", err)
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		// 	c.Abort()
		// 	return
		// }

		// // Check if the token is valid
		// if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 	fmt.Println("Claims:", claims) // Print claims to inspect their structure

		// 	// Convert claims to the desired type
		// 	signedDetails := &helper.SignedDetails{
		// 		Email: claims["email"].(string),
		// 	}

		// 	// Additional checks for other fields
		// 	if username, ok := claims["username"].(string); ok {
		// 		signedDetails.Username = username
		// 	}

		// 	if role, ok := claims["role"].(string); ok {
		// 		signedDetails.Role = role
		// 	}

		// 	if uid, ok := claims["uid"].(string); ok {
		// 		signedDetails.Uid = uid
		// 	}

		// 	// Attach the user information to the context for use in the route handlers
		// 	c.Set("user", signedDetails)
		// 	c.Next()
		// } else {
		// 	fmt.Println("Invalid token")
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		// 	c.Abort()
		// }
		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("No Authorization header provided")})
			c.Abort()
			return
		}

		claims, err := helper.ValidateToken(clientToken)
		if err != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}

		c.Set("email", claims.Email)
		c.Set("username", claims.Username)
		c.Set("uid", claims.Uid)
		c.Set("role", claims.Role)

		c.Next()

	}
}

// RequireAdmin middleware checks if the user has the "admin" role.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the user has the "admin" role.
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		// User has the required role, continue to the next middleware or route handler.
		c.Next()
	}
}
