package handlers

import (
	"net/http"

	"github.com/develoFavour/e-voting-system-be/internal/models"
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

// GetManagedUsers returns approved and rejected student users for admin management.
func GetManagedUsers(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRepo := repository.NewUserRepository(db)
		users, err := userRepo.FindManagedUsers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
			return
		}

		for _, user := range users {
			user.PasswordHash = ""
		}

		c.JSON(http.StatusOK, users)
	}
}

// DeleteManagedUser removes an approved or rejected student user.
func DeleteManagedUser(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		userRepo := repository.NewUserRepository(db)
		user, err := userRepo.FindByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		if user.Role != models.RoleStudent {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only student voter accounts can be removed"})
			return
		}

		if user.Status != models.StatusApproved && user.Status != models.StatusRejected {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Only approved or rejected users can be removed from this page"})
			return
		}

		if err := userRepo.DeleteByID(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user"})
			return
		}

		adminID, _ := c.Get("user_id")
		admin, _ := userRepo.FindByID(adminID.(string))
		activityRepo := repository.NewActivityRepository(db)
		_ = activityRepo.Create(&models.Activity{
			Type:      models.ActivityTypeUserRemoved,
			Message:   "Removed user: " + user.FullName + " (" + user.MatricNumber + ")",
			AdminID:   admin.ID,
			AdminName: admin.FullName,
		})

		c.JSON(http.StatusOK, gin.H{"message": "User removed successfully"})
	}
}
