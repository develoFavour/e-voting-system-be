package handlers

import (
	"log"
	"net/http"

	"github.com/develoFavour/e-voting-system-be/internal/config"
	"github.com/develoFavour/e-voting-system-be/internal/models"
	"github.com/develoFavour/e-voting-system-be/internal/repository"
	"github.com/develoFavour/e-voting-system-be/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		user, err := userRepo.FindByID(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}

		if err := userRepo.UpdateStatus(id, models.StatusApproved); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve voter"})
			return
		}

		// Log activity
		adminID, _ := c.Get("user_id")
		admin, _ := userRepo.FindByID(adminID.(string))
		activityRepo := repository.NewActivityRepository(db)
		_ = activityRepo.Create(&models.Activity{
			Type:      models.ActivityTypeVoterApproved,
			Message:   "Approved voter: " + user.FullName + " (" + user.MatricNumber + ")",
			AdminID:   admin.ID,
			AdminName: admin.FullName,
		})

		c.JSON(http.StatusOK, gin.H{"message": "Voter approved successfully"})
	}
}

// RejectVoter rejects an accreditation request
func RejectVoter(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		userRepo := repository.NewUserRepository(db)
		user, err := userRepo.FindByID(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}

		if err := userRepo.UpdateStatus(id, models.StatusRejected); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject voter"})
			return
		}

		// Log activity
		adminID, _ := c.Get("user_id")
		admin, _ := userRepo.FindByID(adminID.(string))
		activityRepo := repository.NewActivityRepository(db)
		_ = activityRepo.Create(&models.Activity{
			Type:      models.ActivityTypeVoterRejected,
			Message:   "Rejected voter: " + user.FullName + " (" + user.MatricNumber + ")",
			AdminID:   admin.ID,
			AdminName: admin.FullName,
		})

		c.JSON(http.StatusOK, gin.H{"message": "Voter rejected"})
	}
}

func GetAdminCandidates(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateRepo := repository.NewCandidateRepository(db)
		electionRepo := repository.NewElectionRepository(db)
		electionRepo.CheckAndEndExpiredElections()

		liveElection, err := electionRepo.GetLiveElection()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				candidates, err := candidateRepo.FindUnscoped()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch candidates"})
					return
				}
				c.JSON(http.StatusOK, candidates)
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

// AddCandidate adds a new candidate
func AddCandidate(db *mongo.Database, cfg *config.Config) gin.HandlerFunc {
	imageService := services.NewImageUploadService(cfg)

	return func(c *gin.Context) {
		name := c.PostForm("name")
		position := c.PostForm("position")
		positionIDStr := c.PostForm("positionId")
		party := c.PostForm("party")
		manifesto := c.PostForm("manifesto")
		department := c.PostForm("department")
		level := c.PostForm("level")

		if name == "" || (position == "" && positionIDStr == "") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
			return
		}

		var positionID primitive.ObjectID
		if positionIDStr != "" {
			oid, err := primitive.ObjectIDFromHex(positionIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid positionId"})
				return
			}
			positionID = oid
		}

		// Handle image upload using Cloudinary
		var imageURL string
		file, err := c.FormFile("image")
		if err == nil {
			// Upload to Cloudinary
			imageURL, err = imageService.UploadImage(file, "candidates")
			if err != nil {
				log.Printf("Failed to upload image to Cloudinary: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
				return
			}
		}

		candidate := &models.Candidate{
			Name:       name,
			Position:   position,
			PositionID: positionID,
			Party:      party,
			Manifesto:  manifesto,
			Department: department,
			Level:      level,
			ImageURL:   imageURL,
		}

		// If positionId is provided, resolve the position name for display/backward compatibility
		if positionIDStr != "" {
			positionRepo := repository.NewPositionRepository(db)
			pos, err := positionRepo.FindByID(positionIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid position"})
				return
			}
			if candidate.Position == "" {
				candidate.Position = pos.Name
			}
		}

		candidateRepo := repository.NewCandidateRepository(db)
		if err := candidateRepo.Create(candidate); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add candidate"})
			return
		}

		// Log activity
		adminID, _ := c.Get("user_id")
		userRepo := repository.NewUserRepository(db)
		admin, _ := userRepo.FindByID(adminID.(string))
		activityRepo := repository.NewActivityRepository(db)
		_ = activityRepo.Create(&models.Activity{
			Type:      models.ActivityTypeCandidateAdded,
			Message:   "Added new candidate: " + candidate.Name + " for " + candidate.Position,
			AdminID:   admin.ID,
			AdminName: admin.FullName,
		})

		c.JSON(http.StatusCreated, candidate)
	}
}

// GetResults returns decrypted election results (admin only)
func GetResults(db *mongo.Database, encryptionKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		voteRepo := repository.NewVoteRepository(db)
		candidateRepo := repository.NewCandidateRepository(db)
		userRepo := repository.NewUserRepository(db)
		electionRepo := repository.NewElectionRepository(db)

		// Get current election
		currentElection, err := electionRepo.GetCurrentElection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current election"})
			return
		}

		if currentElection == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "No active election found"})
			return
		}

		// Get all votes
		votes, err := voteRepo.FindAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch votes"})
			return
		}

		// Get candidates for current election only
		candidates, err := candidateRepo.FindByElectionID(currentElection.ID.Hex())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch candidates"})
			return
		}

		// Get approved voters count
		approvedCount, err := userRepo.CountByStatus(models.StatusApproved)
		if err != nil {
			approvedCount = 0
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

		// Update candidates with vote counts
		for i := range candidates {
			candidates[i].VoteCount = voteCounts[candidates[i].ID.Hex()]
		}

		c.JSON(http.StatusOK, gin.H{
			"totalVotes":     totalVotes,
			"candidates":     candidates,
			"voteCounts":     voteCounts,
			"approvedVoters": approvedCount,
			"election": gin.H{
				"id":        currentElection.ID.Hex(),
				"title":     currentElection.Title,
				"status":    currentElection.Status,
				"startTime": currentElection.StartTime,
				"endTime":   currentElection.EndTime,
			},
		})
	}
}

// StartElection starts the election with duration
func StartElection(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.StartElectionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		electionRepo := repository.NewElectionRepository(db)
		electionRepo.CheckAndEndExpiredElections()

		// Get current election or create new one
		currentElection, err := electionRepo.GetCurrentElection()
		if err != nil && err != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current election"})
			return
		}

		var electionID string
		if currentElection != nil && currentElection.Status == models.ElectionStatusLive {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Election already live"})
			return
		}
		if currentElection == nil || currentElection.Status == models.ElectionStatusClosed {
			// Create new election
			userID, _ := c.Get("user_id")
			userObjID, err := primitive.ObjectIDFromHex(userID.(string))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
				return
			}
			election := &models.Election{
				Title:       req.Title,
				Description: req.Description,
				Status:      models.ElectionStatusPending,
				Duration:    req.Duration,
				CreatedBy:   userObjID,
			}

			if err := electionRepo.Create(election); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create election"})
				return
			}
			electionID = election.ID.Hex()
		} else {
			electionID = currentElection.ID.Hex()
		}

		// Start the election with duration
		if err := electionRepo.StartElection(electionID, req.Duration); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start election"})
			return
		}

		// Attach staged data to this election
		positionRepo := repository.NewPositionRepository(db)
		candidateRepo := repository.NewCandidateRepository(db)
		_ = positionRepo.AttachUnscopedToElection(electionID)
		_ = candidateRepo.AttachUnscopedToElection(electionID)

		// Backfill legacy candidates missing PositionID by matching Position name to attached positions
		positions, err := positionRepo.FindByElectionID(electionID)
		if err == nil {
			posIDByName := make(map[string]primitive.ObjectID)
			for _, p := range positions {
				if p != nil && p.Name != "" {
					posIDByName[p.Name] = p.ID
				}
			}

			cands, err := candidateRepo.FindByElectionID(electionID)
			if err == nil {
				for _, cand := range cands {
					if cand == nil {
						continue
					}
					if cand.PositionID == primitive.NilObjectID && cand.Position != "" {
						if pid, ok := posIDByName[cand.Position]; ok {
							_ = candidateRepo.SetPositionID(cand.ID.Hex(), pid)
						}
					}
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Election started successfully",
			"duration": req.Duration,
		})

		// Log activity
		adminID, _ := c.Get("user_id")
		userRepo := repository.NewUserRepository(db)
		admin, _ := userRepo.FindByID(adminID.(string))
		activityRepo := repository.NewActivityRepository(db)
		_ = activityRepo.Create(&models.Activity{
			Type:      models.ActivityTypeElectionStarted,
			Message:   "Started election: " + req.Title,
			AdminID:   admin.ID,
			AdminName: admin.FullName,
		})
	}
}

// EndElection ends the election
func EndElection(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		electionRepo := repository.NewElectionRepository(db)

		currentElection, err := electionRepo.GetCurrentElection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current election"})
			return
		}

		if currentElection == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No active election found"})
			return
		}

		if err := electionRepo.UpdateStatus(currentElection.ID.Hex(), models.ElectionStatusClosed); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to end election"})
			return
		}

		// Log activity
		adminID, _ := c.Get("user_id")
		userRepo := repository.NewUserRepository(db)
		admin, _ := userRepo.FindByID(adminID.(string))
		activityRepo := repository.NewActivityRepository(db)
		_ = activityRepo.Create(&models.Activity{
			Type:      models.ActivityTypeElectionEnded,
			Message:   "Ended election: " + currentElection.Title,
			AdminID:   admin.ID,
			AdminName: admin.FullName,
		})

		c.JSON(http.StatusOK, gin.H{"message": "Election ended successfully"})
	}
}

// GetAllElections returns all elections (current and previous)
func GetAllElections(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		electionRepo := repository.NewElectionRepository(db)
		candidateRepo := repository.NewCandidateRepository(db)
		userRepo := repository.NewUserRepository(db)

		elections, err := electionRepo.GetAllElections()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch elections"})
			return
		}

		// Get total approved voters count
		approvedCount, err := userRepo.CountByStatus(models.StatusApproved)
		if err != nil {
			approvedCount = 0
		}

		// Enhance elections with statistics
		var enhancedElections []gin.H
		for _, election := range elections {
			// Get candidates for this election
			candidates, err := candidateRepo.FindByElectionID(election.ID.Hex())
			if err != nil {
				candidates = []*models.Candidate{}
			}

			// Calculate total votes
			totalVotes := 0
			for _, candidate := range candidates {
				totalVotes += candidate.VoteCount
			}

			enhancedElections = append(enhancedElections, gin.H{
				"id":             election.ID.Hex(),
				"title":          election.Title,
				"description":    election.Description,
				"status":         election.Status,
				"startTime":      election.StartTime,
				"endTime":        election.EndTime,
				"createdAt":      election.CreatedAt,
				"updatedAt":      election.UpdatedAt,
				"totalVotes":     totalVotes,
				"approvedVoters": approvedCount,
				"turnout": func() float64 {
					if approvedCount > 0 {
						return float64(totalVotes) / float64(approvedCount) * 100
					} else {
						return 0
					}
				}(),
			})
		}

		c.JSON(http.StatusOK, enhancedElections)
	}
}

// GetElectionDetails returns detailed results for a specific election
func GetElectionDetails(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		electionID := c.Param("id")
		if electionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Election ID is required"})
			return
		}

		electionRepo := repository.NewElectionRepository(db)
		candidateRepo := repository.NewCandidateRepository(db)
		userRepo := repository.NewUserRepository(db)

		// Get election details
		election, err := electionRepo.GetElectionById(electionID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Election not found"})
			return
		}

		// Get candidates for this election
		candidates, err := candidateRepo.FindByElectionID(election.ID.Hex())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch candidates"})
			return
		}

		// Get total approved voters for turnout calculation
		approvedCount, err := userRepo.CountByStatus(models.StatusApproved)
		if err != nil {
			approvedCount = 0
		}

		// Calculate total votes
		totalVotes := 0
		for _, candidate := range candidates {
			totalVotes += candidate.VoteCount
		}

		response := gin.H{
			"election": gin.H{
				"id":          election.ID.Hex(),
				"title":       election.Title,
				"description": election.Description,
				"status":      election.Status,
				"startTime":   election.StartTime,
				"endTime":     election.EndTime,
				"createdAt":   election.CreatedAt,
				"updatedAt":   election.UpdatedAt,
			},
			"candidates":     candidates,
			"totalVotes":     totalVotes,
			"approvedVoters": approvedCount,
			"turnout": func() float64 {
				if approvedCount > 0 {
					return float64(totalVotes) / float64(approvedCount) * 100
				} else {
					return 0
				}
			}(),
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetRecentActivities returns the most recent admin activities
func GetRecentActivities(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		activityRepo := repository.NewActivityRepository(db)
		activities, err := activityRepo.FindRecent(10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch activities"})
			return
		}

		c.JSON(http.StatusOK, activities)
	}
}
