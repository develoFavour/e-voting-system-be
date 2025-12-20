package handlers

import (
	"net/http"

	"github.com/develoFavour/e-voting-system-be/internal/repository"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetCurrentUser returns the logged-in user's profile
func GetCurrentUser(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userRepo := repository.NewUserRepository(db)
		user, err := userRepo.FindByID(userID.(string))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Remove sensitive data
		user.PasswordHash = ""

		c.JSON(http.StatusOK, user)
	}
}

// GetAccreditationStatus returns the user's accreditation status
func GetAccreditationStatus(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userRepo := repository.NewUserRepository(db)
		user, err := userRepo.FindByID(userID.(string))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":       user.Status,
			"matricNumber": user.MatricNumber,
			"fullName":     user.FullName,
			"hasVoted":     user.HasVoted,
		})
	}
}
