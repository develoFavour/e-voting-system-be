package handlers

import (
	"net/http"

	"github.com/develoFavour/e-voting-system-be/internal/models"
	"github.com/develoFavour/e-voting-system-be/internal/repository"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// AddPosition creates a staging position (unscoped) that will be attached to the next started election
func AddPosition(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.AddPositionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		positionRepo := repository.NewPositionRepository(db)
		if existing, err := positionRepo.FindUnscopedByName(req.Name); err == nil && existing != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Position already exists"})
			return
		}

		position := &models.Position{
			Name:          req.Name,
			Description:   req.Description,
			Order:         req.Order,
			MaxSelections: req.MaxSelections,
		}
		if err := positionRepo.Create(position); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create position"})
			return
		}

		c.JSON(http.StatusCreated, position)
	}
}

// GetStagedPositions lists positions not yet attached to an election
func GetStagedPositions(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		positionRepo := repository.NewPositionRepository(db)
		positions, err := positionRepo.FindUnscoped()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch positions"})
			return
		}

		c.JSON(http.StatusOK, positions)
	}
}

// GetLivePositions returns positions for the currently live election (voter-facing)
func GetLivePositions(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		electionRepo := repository.NewElectionRepository(db)
		positionRepo := repository.NewPositionRepository(db)

		electionRepo.CheckAndEndExpiredElections()
		liveElection, err := electionRepo.GetLiveElection()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusOK, []models.Position{})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get election"})
			return
		}

		positions, err := positionRepo.FindByElectionID(liveElection.ID.Hex())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch positions"})
			return
		}

		c.JSON(http.StatusOK, positions)
	}
}
