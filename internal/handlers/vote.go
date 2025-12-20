package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/develoFavour/e-voting-system-be/internal/models"
	"github.com/develoFavour/e-voting-system-be/internal/repository"
	"github.com/develoFavour/e-voting-system-be/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetCandidates returns all candidates
func GetCandidates(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateRepo := repository.NewCandidateRepository(db)
		candidates, err := candidateRepo.FindAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch candidates"})
			return
		}

		c.JSON(http.StatusOK, candidates)
	}
}

// CastVote handles vote submission with atomic double-vote prevention
func CastVote(db *mongo.Database, encryptionKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var req models.CastVoteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userRepo := repository.NewUserRepository(db)
		voteRepo := repository.NewVoteRepository(db)
		candidateRepo := repository.NewCandidateRepository(db)

		// ATOMIC OPERATION: Mark user as voted
		// This prevents double voting even with concurrent requests
		if err := userRepo.MarkAsVoted(userID.(string)); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You have already voted or are not eligible"})
			return
		}

		// Encrypt vote data
		voteData := &models.VoteData{
			Selections: req.Selections,
		}

		encryptedData, err := services.EncryptVote(voteData, encryptionKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process vote"})
			return
		}

		// Create vote hash for integrity
		hash := sha256.Sum256([]byte(encryptedData))
		hashString := hex.EncodeToString(hash[:])

		// Save vote
		userObjectID, _ := primitive.ObjectIDFromHex(userID.(string))
		vote := &models.Vote{
			UserID:            userObjectID,
			EncryptedVoteData: encryptedData,
			Hash:              hashString,
		}

		if err := voteRepo.Create(vote); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save vote"})
			return
		}

		// Increment vote counts for candidates
		for _, candidateID := range req.Selections {
			candidateRepo.IncrementVoteCount(candidateID)
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Vote cast successfully",
			"hash":    hashString,
		})
	}
}

// GetLiveResults returns real-time election results
func GetLiveResults(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateRepo := repository.NewCandidateRepository(db)
		voteRepo := repository.NewVoteRepository(db)

		candidates, err := candidateRepo.FindAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch results"})
			return
		}

		totalVotes, _ := voteRepo.Count()

		c.JSON(http.StatusOK, gin.H{
			"totalVotes": totalVotes,
			"candidates": candidates,
		})
	}
}
