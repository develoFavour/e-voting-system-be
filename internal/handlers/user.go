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

		electionRepo := repository.NewElectionRepository(db)
		voteRepo := repository.NewVoteRepository(db)
		electionRepo.CheckAndEndExpiredElections()
		if liveElection, err := electionRepo.GetLiveElection(); err == nil && liveElection != nil {
			if _, err := voteRepo.FindByUserAndElectionID(userID.(string), liveElection.ID.Hex()); err == nil {
				user.HasVoted = true
			} else {
				user.HasVoted = false
			}
		} else {
			user.HasVoted = false
		}

		// Remove sensitive data
		user.PasswordHash = ""

		c.JSON(http.StatusOK, user)
	}
}

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

		hasVoted := false
		electionRepo := repository.NewElectionRepository(db)
		voteRepo := repository.NewVoteRepository(db)
		electionRepo.CheckAndEndExpiredElections()
		liveElection, liveErr := electionRepo.GetLiveElection()
		if liveErr == nil && liveElection != nil {
			if _, err := voteRepo.FindByUserAndElectionID(userID.(string), liveElection.ID.Hex()); err == nil {
				hasVoted = true
			}
		} else if liveErr != nil && liveErr != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to determine voting status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":       user.Status,
			"matricNumber": user.MatricNumber,
			"fullName":     user.FullName,
			"hasVoted":     hasVoted,
		})
	}
}
