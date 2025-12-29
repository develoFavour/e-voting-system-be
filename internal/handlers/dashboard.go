package handlers

import (
	"net/http"
	"time"

	"github.com/develoFavour/e-voting-system-be/internal/models"
	"github.com/develoFavour/e-voting-system-be/internal/repository"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetDashboardStats returns comprehensive dashboard statistics
func GetDashboardStats(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRepo := repository.NewUserRepository(db)
		voteRepo := repository.NewVoteRepository(db)
		candidateRepo := repository.NewCandidateRepository(db)
		electionRepo := repository.NewElectionRepository(db)

		// Get total registered users
		totalUsers, err := userRepo.Count()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get total users"})
			return
		}

		// Get approved voters
		approvedUsers, err := userRepo.CountByStatus(models.StatusApproved)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get approved users"})
			return
		}

		// Get pending requests
		pendingUsers, err := userRepo.CountByStatus(models.StatusPending)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pending users"})
			return
		}

		// Get total votes cast
		totalVotes, err := voteRepo.Count()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get vote count"})
			return
		}

		// Get total candidates
		totalCandidates, err := candidateRepo.Count()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get candidate count"})
			return
		}

		// Get current election info
		var electionInfo interface{}
		currentElection, err := electionRepo.GetCurrentElection()
		if err == nil && currentElection != nil {
			// Check if election has expired and update status
			electionRepo.CheckAndEndExpiredElections()

			// Refresh election data after potential status update
			currentElection, _ = electionRepo.GetCurrentElection()
			electionInfo = currentElection
		}

		// Get real recent activities
		activityRepo := repository.NewActivityRepository(db)
		activities, _ := activityRepo.FindRecent(10)

		// Get distribution stats
		deptStats, _ := userRepo.GetDepartmentStats()
		facultyStats, _ := userRepo.GetFacultyStats()

		c.JSON(http.StatusOK, gin.H{
			"totalRegistered":  totalUsers,
			"approvedVoters":   approvedUsers,
			"pendingRequests":  pendingUsers,
			"votesCast":        totalVotes,
			"totalCandidates":  totalCandidates,
			"election":         electionInfo,
			"recentActivities": activities,
			"demographics": gin.H{
				"departments": deptStats,
				"faculties":   facultyStats,
			},
			"lastUpdated": time.Now(),
		})
	}
}

// GetCurrentElection returns the current election status
func GetCurrentElection(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		electionRepo := repository.NewElectionRepository(db)

		// Check and end expired elections first
		electionRepo.CheckAndEndExpiredElections()

		currentElection, err := electionRepo.GetCurrentElection()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusOK, gin.H{"election": nil})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current election"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"election": currentElection})
	}
}
