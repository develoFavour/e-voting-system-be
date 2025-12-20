package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/develoFavour/e-voting-system-be/internal/config"
	"github.com/develoFavour/e-voting-system-be/internal/database"
	"github.com/develoFavour/e-voting-system-be/internal/handlers"
	"github.com/develoFavour/e-voting-system-be/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to MongoDB
	db, err := database.Connect(cfg.MongoDBURI, cfg.DBName)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Disconnect()

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORS(cfg.FrontendURL))
	router.Use(middleware.RateLimit())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "eVote API is running"})
	})

	// API routes
	api := router.Group("/api")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register(db))
			auth.POST("/login", handlers.Login(db, cfg.JWTSecret))
			auth.POST("/admin/login", handlers.AdminLogin(db, cfg.JWTSecret))
		}

		// User routes (protected)
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			users.GET("/me", handlers.GetCurrentUser(db))
			users.GET("/status", handlers.GetAccreditationStatus(db))
		}

		// Admin routes (protected + admin only)
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		admin.Use(middleware.AdminOnly())
		{
			admin.GET("/accreditation/pending", handlers.GetPendingAccreditation(db))
			admin.PUT("/accreditation/:id/approve", handlers.ApproveVoter(db))
			admin.PUT("/accreditation/:id/reject", handlers.RejectVoter(db))
			admin.POST("/candidates", handlers.AddCandidate(db))
			admin.GET("/results", handlers.GetResults(db, cfg.EncryptionKey))
			admin.POST("/election/start", handlers.StartElection(db))
			admin.POST("/election/end", handlers.EndElection(db))
		}

		// Voting routes (protected)
		vote := api.Group("/vote")
		vote.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			vote.GET("/candidates", handlers.GetCandidates(db))
			vote.POST("/cast", handlers.CastVote(db, cfg.EncryptionKey))
			vote.GET("/results", handlers.GetLiveResults(db))
		}
	}

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Printf("🚀 eVote API server started on port %s", cfg.Port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
