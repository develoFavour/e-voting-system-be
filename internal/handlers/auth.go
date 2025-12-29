package handlers

import (
	"log"
	"net/http"

	"github.com/develoFavour/e-voting-system-be/internal/config"
	"github.com/develoFavour/e-voting-system-be/internal/models"
	"github.com/develoFavour/e-voting-system-be/internal/repository"
	"github.com/develoFavour/e-voting-system-be/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// Register handles student accreditation registration
func Register(db *mongo.Database, cfg *config.Config) gin.HandlerFunc {
	imageService := services.NewImageUploadService(cfg)

	return func(c *gin.Context) {
		matricNumber := c.PostForm("matricNumber")
		fullName := c.PostForm("fullName")
		department := c.PostForm("department")
		faculty := c.PostForm("faculty")
		password := c.PostForm("password")

		if matricNumber == "" || fullName == "" || password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
			return
		}

		// Handle file upload using Cloudinary
		file, err := c.FormFile("idCard")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID Card image is required"})
			return
		}

		// Upload to Cloudinary
		idCardURL, err := imageService.UploadImage(file, "id_cards")
		if err != nil {
			log.Printf("Failed to upload ID card to Cloudinary: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload ID card"})
			return
		}

		userRepo := repository.NewUserRepository(db)

		// Check if matric number already exists
		existing, _ := userRepo.FindByMatricNumber(matricNumber)
		if existing != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Matriculation number already registered"})
			return
		}

		// Hash password
		hashedPassword, err := services.HashPassword(password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
			return
		}

		// Create user
		user := &models.User{
			MatricNumber: matricNumber,
			FullName:     fullName,
			Department:   department,
			Faculty:      faculty,
			PasswordHash: hashedPassword,
			IDCardURL:    idCardURL,
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

		// Generate access token
		accessToken, err := services.GenerateToken(user.ID.Hex(), string(user.Role), jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// Generate refresh token
		refreshToken, err := services.GenerateRefreshToken(user.ID.Hex(), string(user.Role), jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
			return
		}

		// Remove sensitive data
		user.PasswordHash = ""

		c.JSON(http.StatusOK, gin.H{
			"token":        accessToken,
			"refreshToken": refreshToken,
			"user":         user,
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

		// Generate access token
		accessToken, err := services.GenerateToken(user.ID.Hex(), string(user.Role), jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// Generate refresh token
		refreshToken, err := services.GenerateRefreshToken(user.ID.Hex(), string(user.Role), jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
			return
		}

		// Remove sensitive data
		user.PasswordHash = ""

		c.JSON(http.StatusOK, gin.H{
			"token":        accessToken,
			"refreshToken": refreshToken,
			"user":         user,
		})
	}
}

// RefreshToken handles token refresh
func RefreshToken(db *mongo.Database, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refreshToken" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate refresh token
		claims, err := services.ValidateRefreshToken(req.RefreshToken, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}

		userRepo := repository.NewUserRepository(db)

		// Verify user still exists and is approved
		user, err := userRepo.FindByID(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// For students, check if still approved
		if user.Role == models.RoleStudent && user.Status != models.StatusApproved {
			c.JSON(http.StatusForbidden, gin.H{"error": "User access revoked"})
			return
		}

		// Generate new access token
		newAccessToken, err := services.GenerateToken(user.ID.Hex(), string(user.Role), jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new token"})
			return
		}

		// Optionally generate new refresh token (security best practice)
		newRefreshToken, err := services.GenerateRefreshToken(user.ID.Hex(), string(user.Role), jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new refresh token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token":        newAccessToken,
			"refreshToken": newRefreshToken,
		})
	}
}
