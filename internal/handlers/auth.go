package handlers

import (
	"net/http"

	"github.com/develoFavour/e-voting-system-be/internal/models"
	"github.com/develoFavour/e-voting-system-be/internal/repository"
	"github.com/develoFavour/e-voting-system-be/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// Register handles student accreditation registration
func Register(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userRepo := repository.NewUserRepository(db)

		// Check if matric number already exists
		existing, _ := userRepo.FindByMatricNumber(req.MatricNumber)
		if existing != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Matriculation number already registered"})
			return
		}

		// Hash password
		hashedPassword, err := services.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
			return
		}

		// Create user
		user := &models.User{
			MatricNumber: req.MatricNumber,
			FullName:     req.FullName,
			Department:   req.Department,
			Faculty:      req.Faculty,
			PasswordHash: hashedPassword,
			IDCardURL:    req.IDCardURL,
			Status:       models.StatusPending,
			Role:         models.RoleStudent,
			HasVoted:     false,
		}

		if err := userRepo.Create(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Accreditation request submitted successfully",
			"user": gin.H{
				"id":           user.ID.Hex(),
				"matricNumber": user.MatricNumber,
				"fullName":     user.FullName,
				"status":       user.Status,
			},
		})
	}
}

// Login handles student login
func Login(db *mongo.Database, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userRepo := repository.NewUserRepository(db)

		// Find user
		user, err := userRepo.FindByMatricNumber(req.MatricNumber)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Check password
		if !services.CheckPassword(req.Password, user.PasswordHash) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Check if approved
		if user.Status != models.StatusApproved {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  "Your accreditation is still pending approval",
				"status": user.Status,
			})
			return
		}

		// Generate token
		token, err := services.GenerateToken(user.ID.Hex(), string(user.Role), jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// Remove sensitive data
		user.PasswordHash = ""

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user":  user,
		})
	}
}

// AdminLogin handles admin login
func AdminLogin(db *mongo.Database, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userRepo := repository.NewUserRepository(db)

		// Find user
		user, err := userRepo.FindByMatricNumber(req.MatricNumber)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Check if admin
		if user.Role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}

		// Check password
		if !services.CheckPassword(req.Password, user.PasswordHash) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Generate token
		token, err := services.GenerateToken(user.ID.Hex(), string(user.Role), jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// Remove sensitive data
		user.PasswordHash = ""

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user":  user,
		})
	}
}
