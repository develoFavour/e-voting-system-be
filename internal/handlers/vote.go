package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/develoFavour/e-voting-system-be/internal/models"
	"github.com/develoFavour/e-voting-system-be/internal/repository"
	"github.com/develoFavour/e-voting-system-be/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetPublicCurrentElection returns the current election (public endpoint for voters)
func GetPublicCurrentElection(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		electionRepo := repository.NewElectionRepository(db)
		// Check and end expired elections first
		electionRepo.CheckAndEndExpiredElections()

		// Get latest election (can be pending, live, or closed)
		election, err := electionRepo.GetCurrentElection()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusOK, gin.H{"election": nil})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get election"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"election": election,
		})
	}
}

// GetCandidates returns all candidates
func GetCandidates(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateRepo := repository.NewCandidateRepository(db)
		electionRepo := repository.NewElectionRepository(db)
		electionRepo.CheckAndEndExpiredElections()

		liveElection, err := electionRepo.GetLiveElection()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusOK, []models.Candidate{})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get election"})
			return
		}

		candidates, err := candidateRepo.FindByElectionID(liveElection.ID.Hex())
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
		electionRepo := repository.NewElectionRepository(db)
		positionRepo := repository.NewPositionRepository(db)

		electionRepo.CheckAndEndExpiredElections()
		liveElection, err := electionRepo.GetLiveElection()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusBadRequest, gin.H{"error": "No active election"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get election"})
			return
		}

		user, err := userRepo.FindByID(userID.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		if user.Status != models.StatusApproved {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not eligible to vote"})
			return
		}

		if _, err := voteRepo.FindByUserAndElectionID(userID.(string), liveElection.ID.Hex()); err == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You have already voted"})
			return
		}

		// Must vote all positions that HAVE candidates
		positions, err := positionRepo.FindByElectionID(liveElection.ID.Hex())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch positions"})
			return
		}

		candidates, err := candidateRepo.FindByElectionID(liveElection.ID.Hex())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch candidates"})
			return
		}

		// Map to check if a position has candidates
		posHasCandidates := make(map[string]bool)
		for _, cand := range candidates {
			posHasCandidates[cand.PositionID.Hex()] = true
		}

		for _, pos := range positions {
			// Only require a vote if the position has candidates
			if posHasCandidates[pos.ID.Hex()] {
				if _, ok := req.Selections[pos.ID.Hex()]; !ok {
					c.JSON(http.StatusBadRequest, gin.H{"error": "You must vote all positions"})
					return
				}
			}
		}

		posByID := make(map[string]*models.Position)
		for _, pos := range positions {
			posByID[pos.ID.Hex()] = pos
		}

		// Validate each candidate belongs to this election
		for positionID, candidateID := range req.Selections {
			pos := posByID[positionID]
			if pos == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid position"})
				return
			}

			cand, err := candidateRepo.FindByID(candidateID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid candidate"})
				return
			}
			if cand.ElectionID != liveElection.ID {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid candidate for this election"})
				return
			}

			// Prefer PositionID matching when available; fallback to legacy Position name matching
			if cand.PositionID != primitive.NilObjectID {
				if cand.PositionID != pos.ID {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid candidate for selected position"})
					return
				}
			} else if cand.Position != "" {
				if cand.Position != pos.Name {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid candidate for selected position"})
					return
				}
			}
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
			ElectionID:        liveElection.ID,
			EncryptedVoteData: encryptedData,
			Hash:              hashString,
		}

		fmt.Printf("About to create vote for user %s, election %s\n", userID.(string), liveElection.ID.Hex())
		if err := voteRepo.Create(vote); err != nil {
			fmt.Printf("Failed to create vote: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save vote"})
			return
		}
		fmt.Printf("Vote created successfully\n")

		// Increment vote counts for candidates
		for _, candidateID := range req.Selections {
			_ = candidateRepo.IncrementVoteCount(candidateID)
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
		electionRepo := repository.NewElectionRepository(db)

		electionRepo.CheckAndEndExpiredElections()

		// If an election is live, only return turnout (C3)
		liveElection, liveErr := electionRepo.GetLiveElection()
		if liveErr == nil && liveElection != nil {
			totalVotes, _ := voteRepo.CountByElectionID(liveElection.ID.Hex())
			c.JSON(http.StatusOK, gin.H{
				"mode":       "turnout_only",
				"totalVotes": totalVotes,
				"candidates": []models.Candidate{},
			})
			return
		}

		// Otherwise return the most recent election's full results if closed
		currentElection, err := electionRepo.GetCurrentElection()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusOK, gin.H{"mode": "turnout_only", "totalVotes": 0, "candidates": []models.Candidate{}})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch results"})
			return
		}

		candidates, err := candidateRepo.FindByElectionID(currentElection.ID.Hex())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch results"})
			return
		}

		totalVotes, _ := voteRepo.CountByElectionID(currentElection.ID.Hex())

		c.JSON(http.StatusOK, gin.H{
			"mode":       "full",
			"totalVotes": totalVotes,
			"candidates": candidates,
		})
	}
}

// GetApprovedVoters returns the number of approved users (for turnout stats)
func GetApprovedVoters(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRepo := repository.NewUserRepository(db)
		count, err := userRepo.CountByStatus(models.StatusApproved)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch approved voters"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"approvedVoters": count})
	}
}
