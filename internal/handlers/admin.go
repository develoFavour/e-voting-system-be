package handlers

import (
	"net/http"

	"github.com/develoFavour/e-voting-system-be/internal/models"
	"github.com/develoFavour/e-voting-system-be/internal/repository"
	"github.com/develoFavour/e-voting-system-be/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetPendingAccreditation returns all pending accreditation requests
func GetPendingAccreditation(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRepo := repository.NewUserRepository(db)
		users, err := userRepo.FindPendingAccreditation()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch requests"})
			return
		}

		// Remove sensitive data
		for _, user := range users {
			user.PasswordHash = ""
		}

		c.JSON(http.StatusOK, users)
	}
}

// ApproveVoter approves an accreditation request
func ApproveVoter(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		userRepo := repository.NewUserRepository(db)
		if err := userRepo.UpdateStatus(id, models.StatusApproved); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve voter"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Voter approved successfully"})
	}
}

// RejectVoter rejects an accreditation request
func RejectVoter(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		userRepo := repository.NewUserRepository(db)
		if err := userRepo.UpdateStatus(id, models.StatusRejected); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject voter"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Voter rejected"})
	}
}

// AddCandidate adds a new candidate
func AddCandidate(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.AddCandidateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		candidate := &models.Candidate{
			Name:      req.Name,
			Position:  req.Position,
			Party:     req.Party,
			Manifesto: req.Manifesto,
			ImageURL:  req.ImageURL,
		}

		candidateRepo := repository.NewCandidateRepository(db)
		if err := candidateRepo.Create(candidate); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add candidate"})
			return
		}

		c.JSON(http.StatusCreated, candidate)
	}
}

// GetResults returns decrypted election results (admin only)
func GetResults(db *mongo.Database, encryptionKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		voteRepo := repository.NewVoteRepository(db)
		candidateRepo := repository.NewCandidateRepository(db)

		// Get all votes
		votes, err := voteRepo.FindAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch votes"})
			return
		}

		// Get all candidates
		candidates, err := candidateRepo.FindAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch candidates"})
			return
		}

		// Decrypt and count votes
		voteCounts := make(map[string]int)
		for _, vote := range votes {
			voteData, err := services.DecryptVote(vote.EncryptedVoteData, encryptionKey)
			if err != nil {
				continue // Skip invalid votes
			}

			for _, candidateID := range voteData.Selections {
				voteCounts[candidateID]++
			}
		}

		// Total votes cast
		totalVotes, _ := voteRepo.Count()

		c.JSON(http.StatusOK, gin.H{
			"totalVotes": totalVotes,
			"candidates": candidates,
			"voteCounts": voteCounts,
		})
	}
}

// StartElection starts the election (placeholder for future implementation)
func StartElection(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement election state management
		c.JSON(http.StatusOK, gin.H{"message": "Election started"})
	}
}

// EndElection ends the election (placeholder for future implementation)
func EndElection(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement election state management
		c.JSON(http.StatusOK, gin.H{"message": "Election ended"})
	}
}
