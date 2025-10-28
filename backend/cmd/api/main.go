package main

import (
	"backend/internal/database"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	fmt.Println("Starting backend")
	if err := godotenv.Load(); err != nil {
		log.Println("Could not load environment variables, no .env file found")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// this is my router
	router := gin.Default()

	// Initialize database connection
	db, err := database.NewDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer db.Close()

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

	log.Printf("Server is running on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server")
	}
}
