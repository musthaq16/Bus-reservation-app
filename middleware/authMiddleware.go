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
