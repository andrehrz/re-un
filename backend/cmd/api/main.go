package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"backend/internal/database"
	"backend/internal/handlers"
)

func main() {

	fmt.Println("Starting backend")
	if err := godotenv.Load(); err != nil {
		log.Println("Could not load environment variables, no .env file found")
	}

	// Initialize database connection
	db, err := database.NewDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer db.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// this is my router
	router := gin.Default()

	// initialzing handlers
	authHandler := handlers.NewAuthHandler(db)

	// defining some simple GET endpoints
	router.GET("/health", func(c *gin.Context) {
		// returns a JSON response
		if err := db.Ping(); err != nil {
			c.JSON(500, gin.H{
				"status":  "error",
				"message": "could not connect to database",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"message":  "server is running",
			"database": "connected",
		})
	})

	router.POST("/api/auth/register", authHandler.RegisterUser)
	router.POST("/api/auth/login", authHandler.LoginUser)
	router.POST("/api/auth/refresh", authHandler.RefreshToken)

	log.Printf("Server is running on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server")
	}
}
